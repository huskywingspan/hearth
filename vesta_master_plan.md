# PROJECT VESTA: MASTER DESIGN & TECHNICAL SPECIFICATION

> **Codename:** Project Vesta | **Product Name:** Hearth
> **Phase:** Research & Exploration
> **Last Updated:** 2026-02-11

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

## 1.1 PRODUCT PHILOSOPHY

> "100% of what 90% of people will use" — Hearth must be **feature-complete** before marketing. Users don't adopt platforms that feel unfinished. The gap between Hearth and the group chat they already use must be zero for common features, with Hearth's unique qualities (fading chat, spatial voice, privacy, warmth) as the pull.

### The North Star: 5 Minutes to Voice

**From clicking a link to being in a voice Den: 5 minutes or less.** This is the UX benchmark. Every design decision is measured against it.

### Product Principles

| # | Principle | Meaning |
|---|----------|--------|
| 1 | **5-Minute-to-Voice** | From clicking a link to being in a voice Den: 5 minutes or less. Every friction point is a dropout. |
| 2 | **No Account Required** | The Knock system lets guests join with just a display name. Account creation is an upgrade, not a gate. |
| 3 | **QR Code Connect** | Homeowners share their House via QR code (printed, texted, displayed on a screen). Scan → connect → Knock. Zero typing of URLs. |
| 4 | **Feature Complete Before Marketing** | Don't announce until the basics work flawlessly. Missing image sharing or emoji reactions makes Hearth feel "not ready." |
| 5 | **PWA First** | Mobile is a first-class citizen via Progressive Web App. Install from browser, push notifications, offline support. Native wrapper (Capacitor) comes later. |

### 5-Minute Onboarding Trace

| Step | Time | Action | Hearth Moment |
|------|------|--------|---------------|
| 0:00 | +0s | Friend sends QR code / link via existing group chat | "Check out where we're hanging out" |
| 0:10 | +10s | Click link → hearthapp.chat landing page | See philosophy, one-click "Connect to a House" |
| 0:30 | +30s | Enter House URL or scan QR code | Dynamic server connect — no app install needed |
| 0:45 | +45s | See the Door → enter display name → Knock | "Sarah is at the door" — host hears knock sound |
| 1:30 | +90s | Host approves → Guest enters the House | Warm transition animation, welcome message |
| 2:00 | +120s | Browse Dens, see who's around, read recent chat | Presence indicators, ambient cues |
| 2:30 | +150s | Click into a Den → join Table voice | Ember glow, spatial audio — **they're in** |
| 3:30 | +210s | Talking, laughing, reacting to messages | The magic moment — this is the "Digital Living Room" |

### Feature Completeness Audit ("100% of 90%")

Features that 90% of users in a Discord/group chat use daily. Every row must be green before v1.0 marketing.

| Feature | Discord Has It | Hearth Status | Target Version | Notes |
|---------|---------------|---------------|----------------|-------|
| Text chat (persistent) | ✅ | ✅ Dens | v0.3 | Schema in FF-001 |
| Ephemeral chat | ❌ | ✅ Campfires | v0.2 | Unique differentiator |
| Voice chat | ✅ | 🔲 Planned | v0.4 | Table + 4 Corners |
| Image / file sharing | ✅ | 🔲 Not planned | v1.0 | **Critical gap.** PocketBase has built-in file storage. |
| Emoji reactions | ✅ | 🔲 Not planned | v1.0 | Unicode-only. No custom emoji for v1.0. |
| Reply / thread | ✅ | 🔲 Not planned | v1.0 | Reply-to with scroll-to-parent. No full threads for v1.0. |
| @mentions | ✅ | 🔲 Not planned | v1.0 | `@name` with notification highlight. |
| Message search | ✅ | 🔲 Not planned | v1.0 | SQLite FTS5 — zero dependencies. Dens only (Campfires are ephemeral). |
| Push notifications | ✅ | 🔲 Not planned | v1.0 | PWA + Web Push API. Service Worker. |
| Edit / delete own messages | ✅ | 🔲 Not planned | v1.0 | Essential. "Oops" safety net. |
| User avatars | ✅ | 🔲 Not planned | v1.0 | Upload or generated (initials/Dicebear). PocketBase file field. |
| Link previews | ✅ | 🔲 Not planned | v1.0 | Server-side OpenGraph fetch. Privacy: proxy through PB, don't leak user IPs. |
| Pinned messages | ✅ | 🔲 Not planned | v1.0 | Per-Den. Simple boolean field on messages. |
| DMs | ✅ | 🔲 Planned | v0.3 | Schema in FF-004. E2EE at v1.0. |
| Screen share | ✅ | 🔲 Not planned | v1.1+ | Post-MVP. CPU-heavy. |
| Video | ✅ | 🔲 Planned | v0.4 | 480p/15fps ambient. Not a video call replacement. |

### Friction Map

Every point where a potential user might abandon the flow.

| Dropout Point | Friction | Fix |
|--------------|---------|-----|
| "What is this?" | Link looks unfamiliar | Landing page: 3 seconds to understand value prop |
| "Do I have to install something?" | App install is a hard stop | PWA — works in browser, optional install |
| "I have to make an account?" | Registration fatigue | The Knock — guest access with display name only |
| "How do I find the server?" | Manual URL entry is error-prone | QR code connect — scan and go |
| "Nobody's here" | Empty room = abandoned | Presence hints on Front Porch ("3 people chatting") |
| "This is confusing" | Unfamiliar UI | Guided first visit: one Den, one action (join voice) |
| "I can't send a picture" | Missing basic feature | Feature completeness before marketing (see audit above) |
| "It doesn't work on my phone" | No mobile app | PWA with responsive design, push notifications |

### Mobile Progression

| Phase | Approach | When | Investment |
|-------|---------|------|------------|
| **Phase 1** | PWA (Service Worker + Web Push API + responsive UI) | v1.0 | Low — it's the same React app |
| **Phase 2** | Capacitor wrapper (native shell around the PWA) | Post-v1.0 | Medium — native push, app store listing |
| **Phase 3** | Evaluate React Native **only if** PWA hits a wall | Never (unless forced) | High — avoid unless absolutely necessary |

The PWA approach means Hearth works on every phone from day one. No app store gatekeeping. "Built by a guy in his room" is a feature, not a limitation — it means no corporate incentives to add tracking or dark patterns.

### Market Position

Hearth is not competing with Discord for millions of users. Hearth competes with **the group chat your friend group already has** — iMessage, WhatsApp, a neglected Discord server. The pitch is: "What if your group chat had a living room you could walk into?"

**Target audience:** Self-hosters, privacy-conscious friend groups, small communities (5–30 people), remote families, tabletop gaming groups, creative collectives.

**Positioning:** A "beloved niche tool" like Immich (photos), Jellyfin (media), or Mealie (recipes) — thousands of passionate users, not millions of indifferent ones.

---

## 2. THE TECH STACK

| Layer | Technology | Role |
|-------|-----------|------|
| Deployment | **Docker Compose** | Target: 1 vCPU, 1GB RAM |
| Backend / API | **PocketBase** (Go + SQLite) | Auth, real-time DB, chat history, cron jobs, plugin host |
| Voice / Video | **LiveKit** (Go, WebRTC SFU) | Spatial audio, bandwidth mgmt, ICE Lite |
| TLS / Proxy | **Caddy** | Auto-TLS (Let's Encrypt), reverse proxy for PocketBase + LiveKit |
| Frontend | **React + Vite + TailwindCSS** | TypeScript strict mode |
| Plugins | **Extism** (Wasm) | Sandboxed WASM plugins ("Cartridges") |

### 2.1 Architecture: The Co-located Monolith

Hearth collapses the entire stack into a single machine. No microservices, no Kubernetes. Components communicate over the loopback interface:

- **Data Plane (LiveKit):** High-frequency, low-latency UDP traffic for voice/video.
- **Control Plane (PocketBase):** HTTP/WebSocket for chat, auth, signaling.
- **Storage Layer (SQLite):** Embedded in PocketBase — no TCP database protocol overhead.

This minimizes data travel time and eliminates the memory cost of duplicate caching layers.

### 2.1.1 TLS Termination (Caddy)

WebRTC **requires** HTTPS/WSS — browsers block microphone access on insecure origins (except `localhost`). PocketBase has built-in Auto-TLS but LiveKit also needs TLS for its WebSocket signaling. **Caddy** serves as a lightweight reverse proxy handling auto-TLS (Let's Encrypt) for both:

- `api.hearth.example` → PocketBase (`:8090`)
- `lk.hearth.example` → LiveKit WebSocket (`:7880`)

Caddy's RAM footprint is negligible (~10MB) and fits within the OS/headroom budget. The exact container topology (single container with process supervisor vs. Docker Compose) is pending ADR-001 (see `docs/RESEARCH_BACKLOG.md`).

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

### The House (Server Architecture)

A Hearth server is a **House** — a single self-hosted instance where a small group lives. The House has Dens (permanent rooms), a Backyard (where Campfires burn), and DMs (private conversations).

**Roles — Who Lives Here:**

| Role | Capabilities | Assignment |
|------|-------------|------------|
| **Homeowner** | Full server control. Create/delete Dens and Campfires. Manage all settings. Assign Keyholders. | First account (server setup) |
| **Keyholder** | Create/configure Dens and Campfires (if delegated by Homeowner). Moderate delegated spaces. | Assigned by Homeowner |
| **Member** | Join Dens and Campfires. Send messages. Join voice. | Any authenticated user |
| **Guest** | Session-bound visitor via The Knock. Limited to approved space. | Arrives via Knock, vouched by host |

### A. Dens (Permanent Rooms + Optional Voice)

> **Replaces:** "Portal" (retired — see ADR-007)

**Mental Model:** A room in the house. The living room, the study, the game room. Permanent text history, with optional voice and video.

**Text:** Persistent messages. Searchable. New member history visibility is **configurable per-Den** — the Homeowner sets a server default, and can delegate the choice to Den creators. Den types: general discussion, topic-focused, announcement.

**Voice — Table + 4 Corners:**

Voice in Dens uses a **discrete spatial model**, not continuous topology:

- **The Table** — Central area where everyone hears everyone equally. Default position on join. Like sitting at a dinner table together.
- **4 Corners** — Semi-private positions. Moving to a Corner with another person creates a quieter side conversation (reduced Table volume, boosted Corner partner). Like stepping aside at a party.
- **Navigation:** Click to move between Table ↔ Corner positions. No WASD, no coordinate dragging. Simple, intentional.

**Audio visualization** still applies but simplified for discrete positions:

| Visual State | Position | Implementation |
|-------------|----------|----------------|
| Speaking at Table | Table | Avatar pulses with warm "Ember" glow; visible to all |
| Speaking in Corner | Corner | Ember glow visible to Corner partners; dimmed to Table |
| Silence | Any | Soft breathing animation (ambient presence) |

- **Ember** glow for active speakers (warm pulse, not green ring).
- **"Lean In"** — Click-and-hold on a user at Table to boost their audio, duck others. Focus cursor.

**Video (when enabled):**
- Max 480p, 15fps — "ambient video" / "picture frame" aesthetic, not video call.
- Simulcast disabled. Dynacast enabled (pause unsubscribed after 5s).
- Default: `canPublishVideo: false` in JWT. Homeowner/Keyholder enables per-Den.

### B. Campfires (Ephemeral Chat — The Backyard)

**Core Principle:** Messages are ephemeral thoughts, not permanent records. This combats "archival anxiety" and the "exhibition effect" where users self-censor because they're creating a permanent, searchable record.

**Location:** Campfires live in the **Backyard** — the outdoor annex to the House. They're casual, impermanent, disposable. A Campfire self-destructs when its last messages fade.

**Fade Time:** Configurable by the Campfire creator (within bounds set by the Homeowner). Slider from minutes to hours.

**4-Stage Transparency Decay:**

| Stage | Opacity | Visual | Purpose |
|-------|---------|--------|---------|
| Fresh | 100% | Full color | Active conversation |
| Fading | 50% | Reduced saturation | "Past context, not current" |
| Echo | 10% | Blurred + gray shift | Visual texture of "past chatter" — the room feels lived-in |
| Gone | 0% | Removed | Client + server deletion |

**Implementation:**
- CSS `animation-delay` driven (negative delay starts mid-fade for page reloads), NOT JavaScript loops. Saves mobile battery.
- **Echo stage enhancement** (from R-009 gap analysis): Add `filter: blur(1px)` and gray color shift at the Echo stage, not just opacity reduction. Creates the "Ghost Text" effect described in the UX research.
- Server: Cron-based "Lazy Sweep" GC runs every minute, bulk-deletes via indexed `expires_at`. No per-message timers.
- Nightly `VACUUM` for physical data erasure (logical `DELETE` only marks pages as free).

**Typing Presence — "Mumbling":**
- Instead of "User is typing..." (which creates performance anxiety), show a blurred waveform or abstract "scribbles" indicating **rhythm, length, and intensity** without revealing content.
- Fast flurry = excited. Slow deliberate strokes = thoughtful. Mimics hearing someone take a breath to speak.

**The "Drunk Test":** Does the user feel safe enough to say something stupid? The visual language of fading text must reinforce impermanence — light, airy, and translucent, not solid like a legal document.

### C. DMs (Direct Messages)

**Persistence:** Permanent (like Dens). DMs are private letters, not campfire whispers.

**Features:**
- 1:1 text conversations between any two Members.
- Optional 1:1 voice/video call (direct WebRTC, no Table/Corners model).
- E2EE at v1.0 (simpler key exchange — only 2 parties).
- No group DMs for v1.0 (Dens serve that purpose).

**Privacy:** DM history lives on the server (PocketBase), encrypted at rest via E2EE. The Homeowner cannot read E2EE DMs even with database access.

### D. The Knock (Security & Onboarding)

**Problem:** Invite links (Discord/Zoom) are impersonal and abuse-prone. Account creation funnels are high-friction. Zoom's waiting room is "purgatory."

**The Foyer Flow:**

1. **The Doorstep (Guest View):** Guest clicks Hearth link → sees a "Door" → enters display name + optional note → "Knocks."
2. **The Peephole (Host View):** Host hears subtle knock sound → notification: "Sarah is at the door" → can peek without guest knowing.
3. **Opening the Door:** Host clicks "Let In" → guest transitions from waiting to the House.
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

### E. Cartridges (Extensibility — Future Phase)

**Problem:** Running external bot processes (Node.js, Python) consumes 50–100MB each on a 1GB server.

**Solution:** Embedded WebAssembly plugins via **Extism**. Plugins compile to `.wasm` binaries and run inside the PocketBase process. Developers can write Cartridges in Rust, Go, or **JavaScript (via QuickJS→Wasm compilation)** — making the ecosystem accessible to web developers.

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
| **Chat E2EE (v1.0)** | Campfire + DM messages encrypted client-side. Server stores ciphertext only. Key exchange via public keys on user records. |
| **Voice E2EE (v2.0)** | Insertable Streams — browser encrypts audio frames before WebRTC stack. LiveKit sees only encrypted blobs. |
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
- **Radical Quiet spec:** Auto-hide all chrome (sidebar, toolbars, mute buttons) after N seconds of pure conversation. Show on hover/tap. The UI should disappear when you're just talking.

### 6.5 Navigation (from R-009 Gap Analysis)
- **Do NOT use Discord's "Left Sidebar Server List" pattern.** (UX Research §5.2 explicitly warns against this.)
- Current `RoomList.tsx` sidebar is functional scaffolding for development — it will be replaced.
- Target: **House navigation model** — visualize the House as a place, not a folder tree. Dens are rooms you walk between. Campfires are visible in the Backyard. DMs are a separate private drawer.
- **Sliding pane transitions** between Dens (View Transitions API or Framer Motion) — maintain spatial continuity.

---

## 7. COMPETITIVE POSITIONING

| Competitor | Core Flaw | Hearth Remedy |
|-----------|-----------|---------------|
| **Discord** | Noisy, gamer-maximalist, Nitro upsells, "Times Square" UI | Radical Quiet — hide controls during conversation; no upsells |
| **Matrix / Element** | Engineer brutalism — exposes raw protocol plumbing, cold UI, no sense of "place" | Protocol abstraction — hide hashes/keys; warmth-first design |
| **Guilded / Revolt** | Feature-bloat Discord clones — copy the UI, add more buttons | Contextual minimalism — "Home" metaphor, not server/channel folders |
| **Gather.town** | Gamification trap — pixel art, WASD navigation fatigue, "toy not tool" | Abstract topology, click-to-drift, magnetic zones |
| **Zoom** | Scheduled, performative, waiting room purgatory | Ambient always-on presence; "Front Porch" with hospitality |
| **iMessage / WhatsApp / Signal** | No voice presence, no "room" concept, group chats are flat text | Always-on voice Dens, spatial audio, visual warmth, fading Campfires |

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

## 9. PROJECT PLAN

> **Detailed release roadmap with task IDs:** See [`docs/ROADMAP.md`](docs/ROADMAP.md)
> **Research backlog & open questions:** See [`docs/RESEARCH_BACKLOG.md`](docs/RESEARCH_BACKLOG.md)

### Release Summary

| Release | Codename | Goal | Target |
|---------|----------|------|--------|
| **v0.1** | Ember | Backend skeleton + chat API (no frontend) | Feb 2026 ✅ |
| **v0.2** | Kindling | Frontend shell + Campfire (fading chat) | Feb 2026 ✅ |
| **v0.2.1** | Settling In | Integration fixes + access model simplification | Feb 2026 ✅ |
| **v0.3** | First Friend | Remote access, Den/Campfire schema, landing page, QR connect flow | Apr 2026 |
| **v0.4** | Hearth Fire | Voice — Dens with Table + 4 Corners spatial audio | Jun 2026 |
| **v1.0** | First Light | Full MVP: Knock + Chat features (images, reactions, replies, search) + Admin roles + Chat E2EE + PWA + deployment | Oct 2026 |
| **v1.1** | Warm Glow | Polish, accessibility, House navigation model, screen share | Dec 2026 |
| **v2.0** | Open Flame | Cartridges (plugin system) + Voice E2EE + Hearth Persona (DID) + native mobile wrapper | Q1 2027 |

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

- **Matrix Protocol Integration?** UX report suggests running on Matrix but hiding it. Evaluate complexity vs. federation benefits. Current lean: No — build native PocketBase-first. Revisit federation in v3+ if demand exists.
- **Generative Ambience Engine:** How to implement lightweight procedural audio (fire, rain) without large asset downloads? Options: pre-recorded loops (simple) vs. Web Audio oscillators (zero download) vs. tiny ML models (too CPU-heavy).
- **Screenshot Prevention:** No reliable cross-browser API exists. Accepted as platform limitation (SEC-RISK-001). Rely on visual affordances.
- **Plugin Marketplace:** How to distribute Cartridges? Curated vs. open? Signing requirements? Deferred to v2.0.
- **Accessibility:** How do spatial audio and fading text work for screen readers and hearing-impaired users? Needs early research to avoid costly retrofitting.
- **House Navigation Model:** How to visualize the House as a place, not a folder tree? Key Ring? Neighborhood Map? (UX Research §5.2) — defines the core UI differentiation from Discord.
- **Docker vs systemd:** Technical Research §7.1 recommended systemd for memory savings (~100MB). We chose Docker for UX simplicity. Tradeoff accepted (see ADR-007 notes, R-009 gap analysis).
- **Hearth Persona (Cross-Server Identity):** DID-based portable identity for users across multiple Houses. Research task R-010.
- **RTT Intimacy Mode:** Optional per-relationship live typing for very close friends. v2.0+ feature.