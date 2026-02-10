# **Hearth: Technical Specification and Engineering Report for a Privacy-First, Resource-Constrained Communications Platform**

## **1\. Architectural Philosophy and Feasibility Analysis**

### **1.1 The Imperative for Sovereign Communication**

The contemporary landscape of digital communication is dominated by hyperscalers—centralized platforms that aggregate user data, impose opaque moderation policies, and monetize social graphs. While these platforms offer convenience and high availability, they fundamentally compromise user privacy and data sovereignty. "Hearth" emerges as a counter-paradigmatic response: a self-hosted, privacy-first alternative to Discord, architected specifically to run on the most accessible class of hardware available—the "low-end" Virtual Private Server (VPS).

This report details the engineering rigor required to build Hearth. Unlike commercial platforms that rely on distributed microservices spanning Kubernetes clusters, Hearth must deliver a robust voice and chat experience within a strictly bounded envelope: **1 vCPU and 1GB of RAM**. This constraint is not merely budgetary; it is a forcing function for efficient software design. It necessitates a rejection of heavy abstraction layers in favor of a monolithic, binary-first architecture that prioritizes zero-copy data handling and efficient memory management.

The selected stack—**PocketBase** (Go/SQLite) for the control plane and **LiveKit** (WebRTC/Go) for the media plane—represents a convergence of high-performance, compiled technologies. However, the naive integration of these components will fail under the 1GB memory limit. This report provides a comprehensive technical specification for optimizing these components, implementing a WebAssembly (Wasm) extensibility layer via **Extism**, and securing the platform without relying on external dependencies like Redis or PostgreSQL.

### **1.2 The Hardware Constraint: 1 vCPU and 1GB RAM**

To understand the architectural decisions detailed herein, one must first appreciate the severity of the hardware limitations. A standard "slice" of compute offering 1 vCPU and 1GB of RAM typically runs a Linux kernel (consuming 100-150MB) and system services (consuming 50MB). This leaves approximately 800MB for the application payload.

The challenge is threefold:

1. **Memory Pressure:** Both PocketBase and LiveKit run on the Go runtime. While Go is efficient, it uses a garbage-collected heap. If the application allocates memory faster than the garbage collector (GC) can free it, the heap grows until it triggers an Out-Of-Memory (OOM) kill by the OS kernel. Conversely, aggressive GC burns CPU cycles, leading to latency spikes in audio forwarding.  
2. **CPU Saturation:** A single vCPU core means there is no true parallelism, only concurrency. The operating system scheduler must context-switch between processing incoming network packets (interrupts), handling database I/O, executing application logic, and forwarding real-time media packets. Real-time audio requires isochronous processing; if the CPU is blocked by a heavy database query, audio packets will be dropped, causing robotic voice artifacts.  
3. **Concurrency Limits:** Theoretical benchmarks suggest that a 16-core server can handle thousands of concurrent connections. Linear extrapolation implies a single core might handle 100\. However, overhead is not perfectly linear. The base overhead of the runtime and the operating system consumes a significant percentage of a single core's capacity, reducing the effective headroom for active users.

### **1.3 Architectural Topology: The Co-located Monolith**

Hearth adopts a co-located topology. In traditional scalable architectures, the database, application server, and media server run on separate nodes to allow independent scaling. For Hearth, the network latency introduced by separating these components is unacceptable given the CPU cost of serializing data over the network.

Instead, Hearth collapses the stack:

* **The Data Plane (LiveKit):** Handles the high-frequency, low-latency UDP traffic for voice and video.  
* **The Control Plane (PocketBase):** Handles HTTP/WebSocket traffic for chat, authentication, and signaling.  
* **The Storage Layer (SQLite):** Embedded directly within the PocketBase process, eliminating the TCP overhead of a database protocol.

These components communicate via the loopback interface (localhost), leveraging the kernel's optimized path for local socket communication. This architecture minimizes the "travel time" of data and reduces the memory footprint by removing the need for duplicate data caching layers often found in distributed systems.

## ---

**2\. The Data Layer: PocketBase and SQLite Optimization**

PocketBase serves as the central nervous system of Hearth. It manages user identities, persists chat history, and orchestrates the signaling required to establish WebRTC connections. The choice of SQLite over PostgreSQL is deliberate and critical for the 1GB RAM target. PostgreSQL relies on a process-per-connection model (or a complex thread pool) and significant shared memory buffers, which can easily consume 300MB+ even when idle. SQLite, being a library linked directly into the Go binary, incurs no per-process overhead and shares the memory space of the application.

### **2.1 Schema Design for Relational Real-Time Data**

The database schema must balance the relational integrity required for structured data (users, rooms) with the high-write throughput required for chat messages. PocketBase utilizes "Collections," which map to underlying SQLite tables.

#### **2.1.1 The Users Collection**

This system collection manages authentication. To maintain privacy, Hearth minimizes the metadata stored. The schema includes the standard id and email (optional, for recovery), but enforces a strict separation between the authentication identity and the public profile.

* **username**: A unique string identifier.  
* **avatar**: A file reference stored on the local filesystem.  
* **public\_key**: An optional field for storing the user's public key, enabling the End-to-End Encryption (E2EE) key exchange mechanism described in Section 5\.

#### **2.1.2 The Rooms Collection**

Rooms represent the persistent channels within Hearth. Unlike Discord, which allows infinite nesting, Hearth flattens the hierarchy to reduce query complexity.

* **slug**: A URL-safe unique identifier.  
* **type**: An enum field distinguishing between text, audio, and hybrid rooms.  
* **settings**: A JSON field containing room-specific configurations, such as bitrate limits or retention policies. Storing this as JSON allows for schema flexibility without requiring database migrations for every new feature.

#### **2.1.3 The Messages Collection**

This is the highest-velocity collection. It stores the textual content of the chat.

* **room\_id**: A relation field pointing to the rooms collection.  
* **user\_id**: A relation field pointing to the users collection.  
* **content**: The text payload.  
* **expires\_at**: A DateTime field. This is the cornerstone of Hearth's "fading" message strategy. This field must be indexed to allow the background cleaner to efficiently identify rows for deletion without scanning the entire table.

#### **2.1.4 The Participants Collection (Ephemeral State)**

Tracking who is currently "online" or "in a room" poses a challenge. Writing this state to disk (SQLite) on every connect/disconnect event generates excessive I/O, which degrades performance on a shared VPS SSD. Therefore, Hearth utilizes an **in-memory store** (a Go map protected by a sync.RWMutex) within the PocketBase runtime to track active presence. This state is ephemeral; if the server restarts, the online list is reset, which is acceptable behavior for a real-time presence system.

### **2.2 Optimizing SQLite for Single-Core Throughput**

The default configuration of SQLite is optimized for safety and backward compatibility, utilizing a rollback journal that pauses reads during writes. On a single-core system, this blocking behavior is catastrophic for a real-time application; if a user sends a message (Write) while another user is loading history (Read), the interface will stutter.

To mitigate this, Hearth strictly enforces **Write-Ahead Logging (WAL)** mode. In WAL mode, changes are written to a separate .wal file, allowing readers to proceed concurrently with a writer.

| Pragma Setting | Value | Rationale |
| :---- | :---- | :---- |
| journal\_mode | WAL | Enables non-blocking concurrent reads and writes. 1 |
| synchronous | NORMAL | Reduces the number of fsync() calls. While FULL ensures durability against power loss, NORMAL is sufficient for application crashes and significantly increases transaction speed. 2 |
| cache\_size | \-2000 | Limits the page cache to approx 2MB. We rely on the OS filesystem cache rather than double-caching in SQLite to save application RAM. |
| mmap\_size | 268435456 | Maps up to 256MB of the DB file into memory. This reduces read() syscalls, lowering CPU overhead for read operations. |
| busy\_timeout | 5000 | Sets a 5-second timeout for locking conflicts, preventing immediate failures under load. |

These pragmas are injected at application startup via the PocketBase Go hooks system, ensuring the database is always initialized in the high-performance state.

### **2.3 The "Fading" Data Strategy**

Hearth distinguishes itself with a privacy-centric "fading" message architecture. Messages are not meant to be permanent records; they are ephemeral thoughts. Implementing this on a low-end server requires a strategy that avoids the "thundering herd" problem where deleting thousands of records simultaneously spikes CPU usage.

#### **2.3.1 Cron-Based Garbage Collection**

PocketBase includes a native Go cron scheduler.3 Hearth utilizes this to implement a "Lazy Sweep" strategy. Instead of a separate timer for every message (which would require millions of goroutines), a single background job runs every minute.

Go

// Go Implementation of Lazy Sweep  
app.Cron().MustAdd("message\_gc", "\*/1 \* \* \* \*", func() {  
    // Current time  
    now := time.Now()  
      
    // Execute bulk delete  
    // Uses dbx builder for raw performance, bypassing Record layer overhead  
    \_, err := app.Dao().DB().NewQuery(  
        "DELETE FROM messages WHERE expires\_at \< {:now}",  
    ).Bind(dbx.Params{"now": now}).Execute()  
      
    if err\!= nil {  
        app.Logger().Error("GC failure", "error", err)  
    }  
})

This approach batches the I/O operations. The expires\_at index ensures the database engine can jump directly to the records to be deleted.

#### **2.3.2 Privacy Implications of Deletion**

It is crucial to note that DELETE in SQLite does not immediately overwrite the data on the disk; it marks the database pages as "free" for future use. For a privacy-first application, this "logical deletion" is often insufficient. However, running a VACUUM command to rewrite the database and obliterate the data is a heavy, blocking operation that copies the entire database file.

Hearth adopts a compromise: **Secure Overwrite**. Before the delete operation, the application can optionally update the content field to a string of zeros. However, on a low-end VPS, the I/O cost of UPDATE \+ DELETE is double. The recommended strategy is to run a VACUUM operation once nightly during a low-traffic window (e.g., 4 AM local time), scheduled via the same cron system. This balances privacy (eventual physical erasure) with system stability.

### **2.4 Go Runtime Tuning: Managing the Heap**

The Go Garbage Collector (GC) is designed to trade memory for CPU. By default, it allows the heap to grow by 100% (GOGC=100) before triggering a collection. On a 1GB server, if the base memory usage is 400MB, the GC might wait until 800MB to run. If a sudden spike in traffic pushes usage to 850MB before the GC triggers, the OS might kill the process.

To prevent this, Hearth must be run with the **Soft Memory Limit** introduced in Go 1.19:

GOMEMLIMIT=450MiB

This environment variable tells the Go runtime to ignore the GOGC percentage and trigger a GC cycle whenever the heap approaches 450MB. This creates a predictable memory profile, ensuring PocketBase stays within its allocated slice of the 1GB pie, leaving room for the operating system and the LiveKit processes.

## ---

**3\. The Media Layer: LiveKit on the Edge**

LiveKit is a Selective Forwarding Unit (SFU). Unlike older Multipoint Control Units (MCUs) that decode, mix, and re-encode audio streams (consuming vast amounts of CPU), an SFU acts as an intelligent router. It inspects incoming WebRTC packets and forwards them to the appropriate subscribers. While this is efficient, forwarding packets at high frequency (50 packets per second per user) still generates significant CPU interrupt load.

### **3.1 Network Protocol Architecture**

The transport layer is the bottleneck for real-time performance.

#### **3.1.1 UDP vs. TCP**

WebRTC prefers UDP (User Datagram Protocol) because it is connectionless and does not perform retransmission, which is ideal for real-time media where a late packet is a useless packet. TCP, with its handshake and retransmission logic (Head-of-Line blocking), introduces unacceptable latency.

On a 1 vCPU server, the kernel overhead of managing TCP connection state tables is non-trivial. Hearth explicitly prioritizes UDP.

* **Recommendation:** The configuration rtc.use\_ice\_lite: true is enabled. ICE Lite is a simplified implementation of the Interactive Connectivity Establishment protocol, designed for servers with a public IP address. It drastically reduces the complexity of the connection handshake, saving CPU cycles during the negotiation phase.

#### **3.1.2 Port Management**

LiveKit typically opens a wide range of UDP ports. To keep the kernel's routing table small and efficient, Hearth restricts the UDP port range to 50000-60000. This provides 10,000 ephemeral ports, which is sufficient for the target concurrency (approx. 200 users) while minimizing the attack surface and kernel memory usage for port tracking.

### **3.2 Audio Engineering: The Opus Codec**

Audio quality in Hearth relies on the **Opus** codec. Opus is highly versatile, scaling from low-bitrate speech to high-quality music.

#### **3.2.1 Complexity vs. Quality**

Opus encoding happens on the *client* side. The server does not encode or decode audio; it handles opaque payloads. However, the *size* of these payloads matters.

* **Bitrate:** Hearth defaults to **24 kbps** for voice channels. This is the "knee of the curve" where speech intelligibility is high, but bandwidth usage is minimal.4  
* **Discontinuous Transmission (DTX):** This feature is mandatory for the 1 vCPU environment. DTX detects when a user is silent and stops sending audio packets, sending only occasional "comfort noise" updates. In a typical voice chat with 10 users, usually only one person is speaking. With DTX, the server only processes packets from 1 user instead of 10\. This reduces the CPU load by approximately 90%.6

#### **3.2.2 Opus 1.5/1.6: Deep Redundancy (DRED)**

The latest iterations of Opus (1.5+) introduce machine learning-enhanced redundancy (DRED) and Packet Loss Concealment (PLC).7

* **Mechanism:** The encoder (client) analyzes the speech and embeds a highly compressed, low-bitrate redundancy stream into the packet. If a packet is lost, the decoder (receiver) uses this redundancy to synthesize the missing audio.  
* **Relevance to Hearth:** Low-end VPS instances often suffer from "noisy neighbor" issues, leading to micro-bursts of packet loss. DRED provides resilience against this infrastructure jitter without requiring the server to implement complex NACK (Negative Acknowledgement) buffers or retransmission logic. By enabling DRED support in the client SDKs, Hearth achieves "premium" audio reliability on "budget" infrastructure.

### **3.3 Video Constraints**

Video is the adversary of the 1 vCPU server. A single 720p video stream @ 30fps generates roughly 1.5 Mbps of traffic. If 10 users subscribe to this stream, the server must replicate and forward 15 Mbps. The CPU cost of copying these memory buffers and pushing them to the network interface card (NIC) will saturate a single core quickly.

Hearth implements a **"Voice-First"** policy:

1. **Default Permissions:** The generated JWTs for standard users have canPublishVideo: false.9  
2. **Simulcast Disabled:** Simulcast (sending multiple quality layers of the same video) multiplies the ingress bandwidth. While it helps subscribers with poor connections, it hurts the server's CPU. Hearth disables simulcast, enforcing a single, modest video track (e.g., 360p or 480p) if video is permitted at all.  
3. **Dynacast:** The configuration video.dynacast\_pause\_delay is set to 5s. If no one is subscribing to a video track (e.g., all users are tabbed away or looking at a different screen), the server instructs the publisher to pause the stream. This saves bandwidth and processing power.

### **3.4 LiveKit Configuration Specification**

The config.yaml file for LiveKit must be manually tuned. The default example provided in the documentation is intended for multi-core cloud instances and will lead to instability on a 1GB instance.

| Parameter | Optimized Value | Explanation |
| :---- | :---- | :---- |
| port\_range\_start | 50000 | High range to avoid conflicts. |
| tcp\_port | 7881 | Separate form HTTP/WS to allow traffic shaping. |
| use\_ice\_lite | true | Reduces handshake CPU cost. |
| audio.packet\_loss\_percentage | 0 | Disables artificial loss simulation. |
| video.enable\_transcoding | false | **Critical.** Prevents CPU-killing FFmpeg processes. |
| limit\_per\_ip | 10 | Prevents single-source DoS attacks. |
| logging.level | info | Reduces disk I/O from verbose logs. |

## ---

**4\. Extensibility: The WebAssembly (Wasm) Plugin System**

A static chat application is limited. Users expect bots, moderation tools, and fun commands (e.g., /roll, /giphy). In a traditional architecture (like Discord), these are external processes (Bots) connecting via WebSocket. Running external Node.js or Python processes on a 1GB server is wasteful; each runtime consumes 50-100MB of baseline RAM.

Hearth introduces an **embedded plugin system** using **WebAssembly (Wasm)**. This allows users to write plugins in Rust, Go, or JavaScript (QuickJS) that compile to compact .wasm binaries. These binaries run *inside* the PocketBase process, sharing its memory space but executing in a secure sandbox.

### **4.1 The Runtime Engine: Extism**

Hearth utilizes **Extism**, a framework that wraps the Wasmtime runtime. Extism provides a high-level "Plug-in Development Kit" (PDK) that simplifies the interface between the Host (Go/PocketBase) and the Guest (Wasm Plugin).10

#### **4.1.1 Why Extism/Wasm?**

* **Memory Efficiency:** A Wasm module is just a block of memory. It doesn't need a full OS process context. We can instantiate a plugin, run a function, and destroy it in milliseconds.  
* **Security:** Wasm is sandboxed by default. A plugin cannot read files, open network sockets, or access environment variables unless the Host explicitly provides those capabilities.  
* **Polyglot:** Developers can write plugins in languages they know, provided they compile to Wasm.

### **4.2 The Host-Guest Interface**

The interaction model relies on **Host Functions**. These are Go functions exposed to the Wasm environment.

**Architecture of a Plugin Hook:**

1. **Event Trigger:** A user sends a message. PocketBase triggers the OnBeforeMessageCreate hook.  
2. **Context Creation:** The Go application creates an Extism context and loads the filter.wasm plugin from the pb\_data/plugins directory.  
3. **Data Passing:** The message content is serialized (JSON or raw bytes) and copied into the Wasm memory.  
4. **Execution:** The plugin's process() function is called.  
5. **Host Function Call:** If the plugin needs to persist data (e.g., counting how many times a user swore), it calls the host function db\_increment().  
6. **Result:** The plugin returns a modified string or a boolean (allow/deny).  
7. **Teardown:** The plugin instance is destroyed to free memory.

### **4.3 Implementing the Host in Go**

To integrate Extism into Hearth, the main.go file must wire up the SDK.

Go

// Go Host Implementation  
import "github.com/extism/go-sdk"

func (app \*HearthApp) runPlugin(name string, inputbyte) (byte, error) {  
    // 1\. Load Manifest  
    manifest := extism.Manifest{  
        Wasm:extism.Wasm{  
            extism.WasmFile("plugins/" \+ name \+ ".wasm"),  
        },  
    }

    // 2\. Define Host Functions (Capabilities)  
    // This allows the plugin to log to the server console  
    logFn := extism.NewHostFunction(  
        "hostLog",  
        func(plugin \*extism.CurrentPlugin, stackuint64) {  
            msg, \_ := plugin.ReadString(stack)  
            log.Printf("\[PLUGIN %s\] %s", name, msg)  
        },  
       extism.ValueType{extism.I64}, // Input: Ptr to string  
       extism.ValueType{},           // Output: None  
    )

    // 3\. Instantiate and Call  
    plugin, err := extism.NewPlugin(ctx, manifest, config,extism.HostFunction{logFn})  
    if err\!= nil { return nil, err }  
      
    return plugin.Call("handle\_message", input)  
}

### **4.4 Capability-Based Security**

The power of this system lies in granular control. Hearth defines permissions in the plugins.json configuration file.

* allow\_network: \["api.giphy.com"\] \- The plugin can only make HTTP requests to specific domains via the host\_http\_request function.  
* allow\_store: true \- The plugin can use the KV store.  
* max\_memory: 4MB \- If the plugin exceeds this, it is terminated.

This prevents a malicious or poorly written plugin from performing a Denial-of-Service (DoS) attack on the server or exfiltrating private chat logs.11

## ---

**5\. Security Architecture**

Running a self-hosted service on the public internet requires rigorous security. Without a dedicated security team, the architecture itself must be defensive.

### **5.1 Stateless Authentication: The Cryptographic Invite**

Traditional invite systems store a record in the database: InviteCode: ABC, Uses: 0/10. An attacker can flood the database by generating millions of invites. Hearth uses a **stateless** approach.

#### **5.1.1 HMAC Construction**

An invite link is a self-validating token.

Link: https://hearth.chat/join?r=room1\&t=1735689600\&s=f8a...

* r (Room ID): The target room.  
* t (Timestamp): The expiration time (Unix epoch).  
* s (Signature): HMAC\_SHA256(Server\_Secret, r \+ "." \+ t)

When a user clicks the link:

1. Server checks if t \< Now. If expired, reject (0 DB hits).  
2. Server computes Hash \= HMAC(Secret, r \+ "." \+ t).  
3. Server performs a **constant-time comparison** between the computed hash and s. If they match, the invite is valid.

This relies entirely on CPU (hashing), which is extremely fast and requires no storage.

#### **5.1.2 Secret Rotation**

To revoke invites, Hearth implements a key rotation strategy. The server holds a list of \`\`.

* To validate, it tries the CurrentSecret.  
* If that fails, it tries OldSecret (to allow a grace period).  
* To "Ban All Invites," the admin triggers a rotation: OldSecret is dropped, CurrentSecret becomes OldSecret, and a new random CurrentSecret is generated. This instantly invalidates any invite signed with the dropped key.13

### **5.2 Bot Mitigation: Proof of Work (PoW)**

Public instances are targets for scrapers. To protect the login and join endpoints without using invasive CAPTCHAs (which hurt privacy), Hearth uses a **Client Puzzle Protocol**.14

* **Challenge:** When a client attempts to POST to /api/login, the server returns a 401 with a Challenge header containing a salt and a difficulty factor (e.g., "Find a nonce such that SHA256(salt \+ nonce) ends in '00000'").  
* **Work:** The client's browser must burn CPU cycles to brute-force the nonce. This takes 1-3 seconds for a human device but makes it prohibitively expensive for a bot to perform thousands of attempts per second.  
* **Verification:** The server validates the solution with a single hash operation.

### **5.3 WebRTC End-to-End Encryption (E2EE)**

Standard WebRTC encrypts data in transit (DTLS-SRTP). However, the SFU decrypts this to forward it. For maximum privacy, Hearth supports **Insertable Streams**.

* **Mechanism:** The browser's JavaScript encrypts the audio/video frames *before* passing them to the WebRTC stack.  
* **Key Management:** Users share a "Room Key" out-of-band (or encrypted via public keys).  
* **Server Blindness:** LiveKit sees only encrypted binary blobs. It cannot record audio or perform speech-to-text. This shifts the privacy guarantee from "Trust the Server Admin" to "Trust the Mathematics".9

## ---

**6\. Frontend and User Experience Engineering**

The frontend (built in React or Svelte) serves as the presentation layer. Its primary responsibility in this architecture is to mask the latency and constraints of the backend.

### **6.1 Optimistic UI**

When a user sends a message, the client immediately appends it to the DOM with a "sending" state. It does not wait for the WebSocket confirmation. If the Wasm plugin rejects the message (e.g., moderation), the client reverts the DOM change and displays an error. This creates the illusion of zero-latency interaction even on a slow VPS connection.

### **6.2 The Visual Decay Engine**

Hearth's "fading" messages require precise synchronization.

1. **Time Sync:** The client synchronizes its clock with the server using the NTP-like Date header in API responses.  
2. **CSS Variables:** The decay is driven by CSS, not JavaScript. JS loops are inefficient and drain mobile battery.

.message { \--life: 60s; /\* Total TTL \*/ \--age: calc(var(--server-now) \- var(--created-at)); opacity: calc(1 \- (var(--age) / var(--life))); } \*Correction:\* Standard CSS cannot calculate time dynamically. The correct implementation sets the \`animation-delay\` based on the message timestamp when the component mounts.javascript // React Component const style \= { animationName: 'fadeOut', animationDuration: ${ttl}s, animationDelay: \-${age}s, // Negative delay starts animation mid-way animationTimingFunction: 'linear', animationFillMode: 'forwards' }; \`\`\` This ensures that if a user reloads the page 30 seconds into a 60-second message, the message renders at 50% opacity and fades out in exactly 30 seconds.15

### **6.3 Connectivity Management**

The client maintains a WebSocket connection to LiveKit. On mobile devices, this connection is often broken when the app is backgrounded. Hearth implements an **Exponential Backoff** reconnection strategy. Furthermore, the "Participant" state in PocketBase is updated via a "Heartbeat" (ping) every 30 seconds. If a heartbeat is missed, the user is marked as offline after a grace period (2 missed beats), preventing "ghost" users from cluttering the UI.

## ---

**7\. Operational Strategy and Conclusion**

### **7.1 Operational Deployment**

To avoid the memory overhead of Docker Daemon and container runtimes (which can consume 100MB+), Hearth is deployed as a **systemd service**.

* **Single Binary:** The Go application is compiled to a static binary hearth.  
* **LiveKit:** The livekit-server binary runs alongside it.  
* **Supervision:** Systemd handles restart-on-failure and log rotation (journald).

This "bare-metal" approach yields the highest possible density.

### **7.2 Monitoring**

Observability is implemented via a lightweight /metrics endpoint (Prometheus format) exposed by PocketBase. It tracks:

* Go Heap Allocations (go\_memstats\_heap\_alloc\_bytes)  
* Active Goroutines  
* SQLite Open Connections  
* LiveKit Room Count

A separate, external monitoring agent (or a simple script checking the endpoint) alerts the admin if memory usage crosses 900MB.

### **7.3 Conclusion**

Building "Hearth" on a 1 vCPU / 1GB RAM VPS is a rigorous exercise in constraint-based engineering. It requires rejecting the comfort of heavy frameworks and managed cloud services in favor of bare-metal optimization. By leveraging the efficiency of the Go runtime, the flexibility of SQLite in WAL mode, the raw speed of the LiveKit SFU, and the safety of Wasm for extensibility, Hearth demonstrates that digital sovereignty does not require enterprise-grade hardware. It is a viable, robust, and secure platform for small communities, achieved through architectural discipline and deep technical integration.

## ---

**8\. Appendix: Structured Data & Reference Configurations**

### **Table 1: Memory Budget Allocation (1GB Total)**

| Component | Allocation (MB) | Control Mechanism |
| :---- | :---- | :---- |
| **OS Kernel & System** | 150 | Minimal Alpine/Debian install |
| **PocketBase (Heap)** | 250 | GOMEMLIMIT=250MiB |
| **LiveKit SFU (Heap)** | 400 | GOMEMLIMIT=400MiB |
| **Wasm Plugin Pool** | 50 | Fixed Instance Pool |
| **SQLite Page Cache** | 50 | PRAGMA cache\_size |
| **Safety Headroom** | 100 | Prevent OOM Kill |
| **Total** | **1000** |  |

### **Table 2: Opus Codec Configuration for Voice**

| Parameter | Value | Description |
| :---- | :---- | :---- |
| audio\_bitrate | 24,000 bps | Optimal trade-off for voice clarity vs. bandwidth. |
| frame\_size | 60 ms | Reduces packet overhead (header/payload ratio). |
| use\_inband\_fec | true | Forward Error Correction for packet loss resilience. |
| use\_dtx | true | Discontinuous Transmission (Silence Suppression). |

### **Table 3: Wasm Host Function Manifest**

| Function Name | Permission Level | Description |
| :---- | :---- | :---- |
| log\_info | None | Writes string to server stdout. |
| kv\_get | Storage | Reads from plugin-scoped SQLite store. |
| kv\_set | Storage | Writes to plugin-scoped SQLite store. |
| room\_kick | Moderator | Disconnects a participant (requires specific scope). |
| fetch\_url | Network | HTTP GET to whitelisted domains only. |

#### **Works cited**

1. https://github.com/pocketbase/pocketbase/blob/master/tools/store/store.go I woul... | Hacker News, accessed February 10, 2026, [https://news.ycombinator.com/item?id=46079892](https://news.ycombinator.com/item?id=46079892)  
2. SQLite Optimizations For Ultra High-Performance \- PowerSync, accessed February 10, 2026, [https://www.powersync.com/blog/sqlite-optimizations-for-ultra-high-performance](https://www.powersync.com/blog/sqlite-optimizations-for-ultra-high-performance)  
3. Extend with Go \- Jobs scheduling \- Docs \- PocketBase, accessed February 10, 2026, [https://pocketbase.io/docs/go-jobs-scheduling/](https://pocketbase.io/docs/go-jobs-scheduling/)  
4. Best Audio Codec for Online Video Streaming in 2026, accessed February 10, 2026, [https://antmedia.io/best-audio-codec/](https://antmedia.io/best-audio-codec/)  
5. Opus Recommended Settings \- XiphWiki \- Xiph.org, accessed February 10, 2026, [https://wiki.xiph.org/Opus\_Recommended\_Settings](https://wiki.xiph.org/Opus_Recommended_Settings)  
6. LiveKit SFU, accessed February 10, 2026, [https://docs.livekit.io/reference/internals/livekit-sfu/](https://docs.livekit.io/reference/internals/livekit-sfu/)  
7. Neural encoding enables more-efficient recovery of lost audio packets \- Amazon Science, accessed February 10, 2026, [https://www.amazon.science/blog/neural-encoding-enables-more-efficient-recovery-of-lost-audio-packets](https://www.amazon.science/blog/neural-encoding-enables-more-efficient-recovery-of-lost-audio-packets)  
8. News Archive – Opus Codec, accessed February 10, 2026, [https://opus-codec.org/news/](https://opus-codec.org/news/)  
9. Codecs and more \- LiveKit Documentation, accessed February 10, 2026, [https://docs.livekit.io/transport/media/advanced/](https://docs.livekit.io/transport/media/advanced/)  
10. Projects using Extism in the wild? \#684 \- GitHub, accessed February 10, 2026, [https://github.com/extism/extism/discussions/684](https://github.com/extism/extism/discussions/684)  
11. Extism \- make all software programmable. Extend from within. | Extism \- make all software programmable. Extend from within., accessed February 10, 2026, [https://extism.org/](https://extism.org/)  
12. What is the best way to run untrusted hooks/plugins?, accessed February 10, 2026, [https://softwareengineering.stackexchange.com/questions/387558/what-is-the-best-way-to-run-untrusted-hooks-plugins](https://softwareengineering.stackexchange.com/questions/387558/what-is-the-best-way-to-run-untrusted-hooks-plugins)  
13. How to Secure APIs with HMAC Request Signing in Go \- OneUptime, accessed February 10, 2026, [https://oneuptime.com/blog/post/2026-01-25-secure-apis-hmac-request-signing-go/view](https://oneuptime.com/blog/post/2026-01-25-secure-apis-hmac-request-signing-go/view)  
14. sequentialread/pow-bot-deterrent: A proof-of-work based bot deterrent. Lightweight, self-hosted and copyleft licensed. \- GitHub, accessed February 10, 2026, [https://github.com/sequentialread/pow-bot-deterrent](https://github.com/sequentialread/pow-bot-deterrent)  
15. Fading in... and fading out with CSS transitions \- DEV Community, accessed February 10, 2026, [https://dev.to/nicm42/fading-in-and-fading-out-with-css-transitions-3lc1](https://dev.to/nicm42/fading-in-and-fading-out-with-css-transitions-3lc1)  
16. Synchronizing with Effects \- React, accessed February 10, 2026, [https://react.dev/learn/synchronizing-with-effects](https://react.dev/learn/synchronizing-with-effects)