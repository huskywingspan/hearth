# PROJECT VESTA: MASTER DESIGN & TECHNICAL SPECIFICATION

> **Codename:** Project Vesta | **Product Name:** Hearth
> **Phase:** Research & Exploration
> **Last Updated:** 2026-02-10

---

## 1. THE MISSION

To build **Hearth**: a privacy-first, self-hosted, modular communication platform — "The Digital Living Room."

| Paradigm | Platform | Model | Hearth Distinction |
|----------|----------|-------|--------------------|
| The Convention Center | Discord | Archive-first, gamer-maximalist, attention-extracting | Hearth is intimate, not noisy |
| The Email Server | Matrix/Element | Protocol-first, engineer-brutalist, federation-complex | Hearth hides plumbing, shows warmth |
| The Boardroom | Zoom / Meet | Scheduled, performative, grid-of-pixels | Hearth is ambient, always-on co-presence |

**Core Philosophy:**
- Warmth over efficiency. Intimacy over scale. Presence over archival.
- Privacy by default. No telemetry. E2EE where feasible.
- Constraint-driven engineering: every byte and CPU cycle matters.
- Sound fades, memories soften, doors must be opened from the inside.

---

## 2. THE TECH STACK

| Layer | Technology | Role |
|-------|-----------|------|
| Deployment | **Single Docker Container** | Target: 1 vCPU, 1GB RAM |
| Backend / API | **PocketBase** (Go + SQLite) | Auth, real-time DB, chat history, cron jobs, plugin host |
| Voice / Video | **LiveKit** (Go, WebRTC SFU) | Spatial audio, bandwidth mgmt, ICE Lite |
| Frontend | **React + Vite + TailwindCSS** | TypeScript strict mode |
| Plugins | **Extism** (Wasm) | Sandboxed WASM plugins ("Cartridges") |

### 2.1 Architecture: The Co-located Monolith

Hearth collapses the entire stack into a single machine. No microservices, no Kubernetes. Components communicate over the loopback interface:

- **Data Plane (LiveKit):** High-frequency, low-latency UDP traffic for voice/video.
- **Control Plane (PocketBase):** HTTP/WebSocket for chat, auth, signaling.
- **Storage Layer (SQLite):** Embedded in PocketBase — no TCP database protocol overhead.

This minimizes data travel time and eliminates the memory cost of duplicate caching layers.

### 2.2 Memory Budget (Sacred — 1GB Total)

| Component | Allocation | Control Mechanism |
|-----------|-----------|-------------------|
| OS Kernel & System | 150 MB | Minimal Alpine/Debian install |
| PocketBase (Heap) | 250 MB | `GOMEMLIMIT=250MiB` |
| LiveKit SFU (Heap) | 400 MB | `GOMEMLIMIT=400MiB` |
| Wasm Plugin Pool | 50 MB | Fixed instance pool, per-plugin `max_memory` caps |
| SQLite Page Cache | 50 MB | `PRAGMA cache_size` |
| Safety Headroom | 100 MB | Prevents OOM kill |

### 2.3 SQLite Configuration (WAL Mode)

| Pragma | Value | Rationale |
|--------|-------|-----------|
| `journal_mode` | WAL | Non-blocking concurrent reads and writes |
| `synchronous` | NORMAL | Fewer `fsync()` calls; sufficient for app-crash durability |
| `cache_size` | -2000 | ~2MB page cache; rely on OS filesystem cache |
| `mmap_size` | 268435456 | Map up to 256MB of DB file; reduces `read()` syscalls |
| `busy_timeout` | 5000 | 5s lock timeout prevents immediate failures under load |

### 2.4 LiveKit Configuration (Voice-First)

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| `use_ice_lite` | true | Reduced handshake CPU cost |
| `video.enable_transcoding` | false | **Critical** — prevents CPU-killing FFmpeg processes |
| `port_range` | 50000–60000 | Minimal kernel routing table; supports ~200 users |
| `limit_per_ip` | 10 | DoS prevention |
| Simulcast | Disabled | Saves server CPU; single modest video track if permitted |
| `video.dynacast_pause_delay` | 5s | Pauses unsubscribed video streams |

### 2.5 Opus Audio Configuration

| Parameter | Value | Description |
|-----------|-------|-------------|
| `audio_bitrate` | 24,000 bps | Knee-of-the-curve for voice clarity vs. bandwidth |
| `frame_size` | 60 ms | Reduces packet overhead |
| `use_inband_fec` | true | Forward Error Correction for packet loss resilience |
| `use_dtx` | true | Discontinuous Transmission — silence suppression (~90% CPU reduction in typical group call) |

Opus 1.5+ DRED (Deep Redundancy) is enabled client-side for resilience against VPS "noisy neighbor" packet loss.

---

## 3. CORE FEATURES & UX PATTERNS

### A. The Portal (Ambient Spatial Voice)

**Mental Model:** An abstract topological space — not an RPG map. A dinner table, not a dungeon.

**Key Design Decisions (from UX research):**
- **Reject WASD navigation.** Current spatial platforms (Gather.town) force avatar micro-management that causes "navigation fatigue." Users divert cognitive resources to movement instead of conversation.
- **Click-to-Drift:** Users click a destination or person; their representation "drifts" there automatically with fluid easing. Reduces motor load.
- **Magnetic Zones:** When drifting near a conversation circle, the interface gently snaps/gravitates the user into the group — mimicking the social gravity of joining a real circle.
- **No "God View" Dissonance:** Avoid top-down maps where you see people you can't hear. Visual affordances must match audio affordances.

**Audio Visualization — Gradient Ripples:**

| Visual State | Audio Equivalent | Implementation |
|-------------|----------------|----------------|
| Speaking (Near) | Direct sound | Avatar border pulses with high opacity, warm "ember" glow |
| Speaking (Mid) | Early reflections | Semi-transparent border; ripple expands showing range |
| Speaking (Far) | Reverberant field | "Ghostly" low-opacity; blurred waveform |
| Silence | Ambient presence | Soft static glow (breathing animation) |

- **Opacity = Volume.** Distant users appear semi-transparent; approaching makes them solid.
- **Soft Occlusion:** Audio behind barriers uses low-pass filtering (muffling) instead of hard cutoff. Maintains awareness ("mumbling in the kitchen") without compromising private conversation.
- **"Lean In" / Focus Cursor:** Click-and-hold on a user to beamform — boost their audio, duck surrounding noise. Visual spotlight effect.

### B. Campfires (Ephemeral Chat)

**Core Principle:** Messages are ephemeral thoughts, not permanent records. This combats "archival anxiety" and the "exhibition effect" where users self-censor because they're creating a permanent, searchable record.

**4-Stage Transparency Decay:**

| Stage | Opacity | Trigger | Purpose |
|-------|---------|---------|---------|
| Fresh | 100% | Message sent | Active conversation |
| Fading | 50% | Time elapsed / pushed up by newer messages | "Past context, not current" |
| Echo | 10% | Further decay | Visual texture of "past chatter" — the room feels lived-in |
| Gone | 0% | `expires_at` reached | Client + server deletion |

**Implementation:**
- CSS `animation-delay` driven (negative delay starts mid-fade for page reloads), NOT JavaScript loops. Saves mobile battery.
- Server: Cron-based "Lazy Sweep" GC runs every minute, bulk-deletes via indexed `expires_at`. No per-message timers.
- Nightly `VACUUM` for physical data erasure (logical `DELETE` only marks pages as free).

**Typing Presence — "Mumbling":**
- Instead of "User is typing..." (which creates performance anxiety), show a blurred waveform or abstract "scribbles" indicating rhythm, length, and intensity without revealing content.
- Mimics hearing someone take a breath to speak.

**The "Drunk Test":** Does the user feel safe enough to say something stupid? The visual language of fading text must reinforce impermanence — light, airy, and translucent, not solid like a legal document.

### C. The Knock (Security & Onboarding)

**Problem:** Invite links (Discord/Zoom) are impersonal and abuse-prone. Account creation funnels are high-friction. Zoom's waiting room is "purgatory."

**The Foyer Flow:**

1. **The Doorstep (Guest View):** Guest clicks Hearth link → sees a "Door" → enters display name + optional note → "Knocks."
2. **The Peephole (Host View):** Host hears subtle knock sound → notification: "Sarah is at the door" → can peek without guest knowing.
3. **Opening the Door:** Host clicks "Let In" → guest transitions from waiting to living room.
4. **Account Upgrade:** Guest is a session-bound visitor. If they want to return later, *then* prompted to "claim this key" (create account). Gradual engagement.

**The "Front Porch" (Waiting UI):**
- Guest sees blurred activity hints: "3 people are chatting," "Music is playing." Cannot hear or read specifics but senses life.
- Host can customize with a welcome message/image.
- Reduces "is anyone there?" anxiety.

**Vouched Entry:** Host who approves a guest "vouches" for them. Guest appears as "Guest of Sarah" in the user list. Social accountability without KYC.

**Cryptographic Invite (Stateless HMAC):**
- `https://hearth.example/join?r=room1&t=1735689600&s=f8a...`
- Server validates: check expiry (`t < now`), compute `HMAC_SHA256(secret, r + "." + t)`, constant-time compare with `s`.
- Zero database hits. Secret rotation instantly revokes all outstanding invites.

### D. Cartridges (Extensibility — Future Phase)

**Problem:** Running external bot processes (Node.js, Python) consumes 50–100MB each on a 1GB server.

**Solution:** Embedded WebAssembly plugins via **Extism**. Plugins compile to `.wasm` binaries and run inside the PocketBase process.

**Host-Guest Interface:**
1. Event triggers (e.g., `OnBeforeMessageCreate`) → load plugin → serialize message data into Wasm memory → call `process()` → return allow/deny or modified content → destroy instance.

**Capability-Based Security (plugins.json):**
- `allow_network`: Whitelist specific domains (e.g., `["api.giphy.com"]`)
- `allow_store`: KV store access (plugin-scoped SQLite)
- `max_memory`: Hard cap (e.g., 4MB) — exceeded = terminated

**Host Function Manifest:**

| Function | Permission | Description |
|----------|-----------|-------------|
| `log_info` | None | Write to server stdout |
| `kv_get` / `kv_set` | Storage | Plugin-scoped KV store |
| `room_kick` | Moderator | Disconnect a participant |
| `fetch_url` | Network | HTTP GET to whitelisted domains only |

---

## 4. DESIGN SYSTEM ("SUBTLE WARMTH")

### 4.1 Color Palette

| Role | Dark Mode | Light Mode | Notes |
|------|----------|------------|-------|
| Background Primary | `#2B211E` (Espresso) | `#FAF9D1` (Warm Linen) | Never pure `#000` or `#FFF` |
| Background Secondary | `#3E2C29` (Warm Charcoal) | `#F2E2D9` (Cream) | |
| Active / Focus | Amber / Gold | Amber / Gold | Ember glow for speakers |
| Alerts | Burnt Clay | Burnt Clay | Warm, not "Notification Red" |
| Accents | Sage Green, Terracotta, Slate Blue | Sage Green, Terracotta | Desaturated nature tones |

### 4.2 Typography

| Role | Font | Fallback | Psychology |
|------|------|----------|------------|
| UI Body | Inter | system-ui | Clean legibility |
| Headers / Story | Merriweather | Georgia | "Editorial, bookish" — novel by a fire, not spreadsheet at a desk |
| Alt candidates | Recoleta, Nunito, Quicksand | — | Per UX research; test in prototyping |

### 4.3 Shape Language
- Maximize `border-radius`. Buttons are "lozenges" / "pillows," not rectangles.
- Soft, diffused drop shadows (simulating candlelight), not sharp directional shadows.
- "Frosted glass" layering for UI depth. No cartoon graphics.
- Generous whitespace. High-end minimalism, not cluttered.

### 4.4 Motion (Disney Principles in UI)
- **Ease-in / ease-out** on all transitions. No linear animations.
- **Squash & stretch** on button interactions — buttons "depress" and bounce back.
- Messages **"float" in** with ease-out curve (leaf falling), don't snap into grid.
- Page transitions use **sliding panes** for spatial continuity, not hard cuts.

### 4.5 Sound Design ("Adult ASMR")

| Interaction | Sound | Avoid |
|------------|-------|-------|
| Button press | Wooden click / "thock" | Synthetic beeps |
| Friend joins | Soft cork pop | Alert chime |
| Message fades | Soft rustle | Silence |
| Knock | Wooden door knock | Doorbell |
| Ambient (optional) | Generative: fire crackle, rain, distant coffee shop | Silence or music |

Dynamic mixing: ambient volume rises during conversation lulls, ducks when someone speaks.

---

## 5. SECURITY ARCHITECTURE

| Mechanism | Description |
|-----------|-------------|
| **Stateless HMAC Invites** | Self-validating tokens (no DB storage). Secret rotation = instant revocation. |
| **Proof-of-Work** | Client Puzzle Protocol on public endpoints (login, join). 1–3s for humans, prohibitive for bots. No CAPTCHAs. |
| **WebRTC E2EE** | Insertable Streams — browser encrypts frames before WebRTC stack. LiveKit sees only encrypted blobs. (Future phase) |
| **Wasm Sandboxing** | Capability-based: whitelisted domains, memory caps, scoped KV. Plugins cannot read files or open arbitrary sockets. |
| **Crypto Hygiene** | Constant-time comparison for all hash checks. No timing side-channels. |
| **Key Rotation** | Two-key system (Current + Old with grace period). Drop Old → all old invites instantly invalid. |

---

## 6. FRONTEND ENGINEERING

### 6.1 Optimistic UI
- Render user actions immediately; revert on server rejection. Zero-latency illusion.

### 6.2 Visual Decay Engine
- Time sync via `Date` header in API responses.
- CSS `animation-delay` (negative = start mid-fade) with `animation-fill-mode: forwards`.
- No JS loops for fading — all CSS animation driven.

### 6.3 Connectivity
- Exponential backoff for WebSocket reconnection.
- Heartbeat every 30s; offline after 2 missed beats.
- In-memory presence tracking on server (Go `sync.RWMutex` map), not persisted to SQLite.

### 6.4 Performance
- Mobile-first responsive design.
- Aggressive code-splitting via `React.lazy`.
- Only show tools when needed — UI "breathes" with conversation ("Radical Quiet").

---

## 7. COMPETITIVE POSITIONING

| Competitor | Core Flaw | Hearth Remedy |
|-----------|-----------|---------------|
| **Discord** | Noisy, gamer-maximalist, Nitro upsells, "Times Square" UI | Radical Quiet — hide controls during conversation; no upsells |
| **Matrix / Element** | Engineer brutalism — exposes raw protocol plumbing, cold UI, no sense of "place" | Protocol abstraction — hide hashes/keys; warmth-first design |
| **Guilded / Revolt** | Feature-bloat Discord clones — copy the UI, add more buttons | Contextual minimalism — "Home" metaphor, not server/channel folders |
| **Gather.town** | Gamification trap — pixel art, WASD navigation fatigue, "toy not tool" | Abstract topology, click-to-drift, magnetic zones |
| **Zoom** | Scheduled, performative, waiting room purgatory | Ambient always-on presence; "Front Porch" with hospitality |

---

## 8. DEVELOPMENT GUIDELINES

- **TypeScript:** Strict mode. Functional components only. No class components.
- **Go:** Standard conventions. Minimize allocations in hot paths. `GOMEMLIMIT` always set.
- **Naming:** Use Hearth vocabulary — Portal, Campfire, Knock, Cartridge, Front Porch, Peephole, Ember.
- **CSS:** TailwindCSS utilities. Custom design tokens for Subtle Warmth palette.
- **Testing:** Unit tests for all business logic. Integration tests for PocketBase hooks and LiveKit signaling.
- **Privacy:** No telemetry, analytics, or tracking of any kind. Ever.
- **Ops:** Prometheus-format `/metrics` endpoint. Track heap, goroutines, SQLite connections, room count.

---

## 9. PROJECT PLAN — PHASED DEVELOPMENT

### Phase 0: Research & Foundation (Current — Feb 2026)
- [x] Define architecture and philosophy
- [x] Collect UX research (spatial audio, ephemeral messaging, cozy UI, onboarding)
- [x] Collect technical research (PocketBase/SQLite tuning, LiveKit optimization, Wasm plugins, security)
- [x] Create master design specification
- [x] Initialize repository and copilot instructions
- [ ] Finalize design system tokens (colors, typography, spacing, motion curves)
- [ ] Create architecture decision records (ADRs) for key choices
- [ ] Wireframe core flows: Portal, Campfire, Knock

### Phase 1: Backend Skeleton (Target: Mar–Apr 2026)
- [ ] Scaffold Go project with PocketBase
- [ ] Implement SQLite WAL pragmas via startup hooks
- [ ] Define PocketBase collections: Users, Rooms, Messages
- [ ] Implement HMAC invite token generation and validation
- [ ] Implement message GC cron job (lazy sweep + nightly VACUUM)
- [ ] In-memory presence tracking (Go map + `sync.RWMutex`)
- [ ] Basic auth flow (email/password + magic link)
- [ ] LiveKit config + JWT token generation for room access
- [ ] Proof-of-Work challenge endpoint
- [ ] `/metrics` endpoint (Prometheus format)
- [ ] Unit + integration tests for all hooks

### Phase 2: Frontend Shell (Target: May–Jun 2026)
- [ ] Scaffold React + Vite + TailwindCSS project (TypeScript strict)
- [ ] Implement design system: tokens, components, motion primitives
- [ ] Build Campfire (chat) UI with transparency decay (CSS animation engine)
- [ ] Build "Knock" onboarding flow (Door → Peephole → Front Porch → Entry)
- [ ] "Mumbling" typing indicator
- [ ] Optimistic message sending with revert-on-error
- [ ] WebSocket connection management with exponential backoff
- [ ] Heartbeat-based presence display
- [ ] Sound system: organic foley library + generative ambient engine
- [ ] Mobile-first responsive layout + code-splitting

### Phase 3: Voice — The Portal (Target: Jul–Aug 2026)
- [ ] LiveKit client SDK integration
- [ ] Spatial audio: proximity-based volume attenuation
- [ ] Abstract topological space UI (click-to-drift, magnetic zones)
- [ ] Gradient ripple visualization for speaker range
- [ ] "Lean In" focus cursor (beamforming UX)
- [ ] Soft occlusion (low-pass filtering behind barriers)
- [ ] DTX + Opus DRED client configuration
- [ ] Voice-first permissions (video restricted by default)

### Phase 4: Cartridges — Plugin System (Target: Sep–Oct 2026)
- [ ] Extism/Wasm integration in PocketBase hooks
- [ ] Host function manifest (log, KV, kick, fetch)
- [ ] Plugin capability configuration (plugins.json)
- [ ] Plugin lifecycle: load → execute → teardown (per-event)
- [ ] Memory-capped instance pool (50MB total)
- [ ] Example plugins: moderation filter, /roll command
- [ ] Plugin developer documentation + PDK examples

### Phase 5: Security Hardening & E2EE (Target: Nov–Dec 2026)
- [ ] WebRTC E2EE via Insertable Streams
- [ ] Key exchange mechanism (public key storage, room key distribution)
- [ ] Security audit of HMAC flow, PoW, Wasm sandbox
- [ ] Penetration testing of public endpoints
- [ ] Secret rotation admin tooling

### Phase 6: Deployment & Polish (Target: Q1 2027)
- [ ] Single Docker container build (Alpine-based, multi-stage)
- [ ] Docker Compose for easy self-hosting
- [ ] systemd service alternative (bare-metal deployment)
- [ ] Admin dashboard (room management, user management, plugin management)
- [ ] Documentation: self-hosting guide, admin guide, plugin developer guide
- [ ] Performance profiling under target constraints (1 vCPU, 1GB RAM, ~20 concurrent users)
- [ ] Beta release

---

## 10. FILE & FOLDER STRUCTURE (PLANNED)

```
hearth/
├── backend/              # PocketBase + Go hooks
│   ├── main.go           # Entry point, pragma injection, cron setup
│   ├── hooks/            # Go hook handlers (message GC, auth, plugins)
│   ├── plugins/          # .wasm plugin binaries
│   ├── config/           # LiveKit config, plugins.json
│   └── pb_data/          # PocketBase data directory (gitignored)
├── frontend/             # React + Vite app
│   ├── src/
│   │   ├── components/   # React components (Portal, Campfire, Knock, etc.)
│   │   ├── hooks/        # Custom React hooks (usePresence, useDecay, useSpatialAudio)
│   │   ├── stores/       # State management
│   │   ├── styles/       # Tailwind config + design tokens
│   │   ├── lib/          # Utilities, LiveKit client, PocketBase API client
│   │   └── sounds/       # Organic foley audio files
│   └── public/           # Static assets
├── livekit/              # LiveKit server config
├── docker/               # Dockerfile + docker-compose.yml
├── docs/                 # Design docs, research, ADRs
│   ├── research/         # UX and technical research reports
│   └── adr/              # Architecture Decision Records
└── .github/              # CI workflows, copilot instructions
```

---

## 11. OPEN QUESTIONS & FUTURE EXPLORATION

- **Matrix Protocol Integration?** UX report suggests running on Matrix but hiding it. Evaluate complexity vs. federation benefits.
- **Generative Ambience Engine:** How to implement lightweight procedural audio (fire, rain) without large asset downloads?
- **Screenshot Prevention:** Feasibility of screenshot detection/notification in a web app context.
- **Video Policy:** When and how to enable video beyond voice-first? Per-room toggle? Host-controlled?
- **Plugin Marketplace:** How to distribute Cartridges? Curated vs. open? Signing requirements?
- **Accessibility:** How do spatial audio and fading text work for screen readers and hearing-impaired users?