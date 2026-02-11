# Hearth — Research Backlog & Open Questions

> **Last Updated:** 2026-02-11 (R-009 gap analysis completed, ADR-007 accepted)
> **Owner:** Researcher Agent

---

## Completed Research Tasks

> All high-priority research (R-001 through R-006) is now complete. Backend, frontend SDK, deployment, and spatial audio are unblocked.

---

## Remaining Research Tasks

### R-001: PocketBase v0.23+ API Verification ✅ COMPLETE
**Priority:** Critical | **Blocks:** E-004, E-010 through E-015
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-001-pocketbase-api-verification.md`](research/R-001-pocketbase-api-verification.md)
**Outcome:** PocketBase v0.36.2 API fully verified. Current correct patterns: `app.OnServe().BindFunc()` for startup hooks, `app.DB()` for raw SQL via dbx builder, `app.Cron().MustAdd()` for scheduling, `app.FindRecordById()`/`app.FindRecordsByFilter()` for records, `core.NewRecord()` + `app.Save()` for creation, `app.RunInTransaction()` for atomic ops, `app.SubscriptionsBroker()` for custom realtime. SQLite PRAGMAs injected via `app.OnBootstrap()`. Deprecated patterns fully documented with migration table. **Backend is now unblocked.**

### R-002: Caddy Reverse Proxy for PocketBase + LiveKit TLS ✅ COMPLETE
**Priority:** Critical | **Blocks:** E-005, E-006
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-002-caddy-livekit-tls-config.md`](research/R-002-caddy-livekit-tls-config.md)
**Outcome:** Major discovery — LiveKit's official deployment uses a **custom Caddy build** (`livekit/caddyl4`) with Layer 4 TLS SNI routing, NOT a standard Caddyfile. Config is YAML, not Caddyfile syntax. Architecture: Caddy listens :443, routes by SNI to TURN (localhost:5349), LiveKit API (localhost:7880), and PocketBase (localhost:8090). All containers use `network_mode: "host"`. Three subdomains: `hearth.example` / `lk.hearth.example` / `turn.hearth.example`. **Redis is NOT required for single-node** (saves ~25MB RAM). Complete `caddy.yaml`, `livekit.yaml`, and `docker-compose.yaml` templates provided. Q-008 resolved. **Docker deployment is now unblocked.**

### R-003: Container Topology Decision (ADR-001) ✅ COMPLETE
**Priority:** High | **Blocks:** E-005
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-003-container-topology.md`](research/R-003-container-topology.md)
**Outcome:** ADR-001 formally accepted. **Docker Compose with 3 containers, all `network_mode: "host"`.** Pre-resolved by R-002 findings — LiveKit's official deployment uses this exact pattern. Single container approach (PIVOT-001) formally retired. Options evaluated: single container + s6-overlay (rejected: non-standard, harder to debug), Docker Compose (accepted: standard practice, clean isolation), hybrid (rejected: mixed networking modes). Self-hoster UX: `git clone && cp .env.example .env && docker compose up -d`.

### R-004: PocketBase JS SDK — Real-time Subscriptions ✅ COMPLETE
**Priority:** High | **Blocks:** K-002, K-010
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-004-pocketbase-js-sdk.md`](research/R-004-pocketbase-js-sdk.md)
**Outcome:** SSE-based realtime (NOT WebSocket). Auto-reconnect with backoff `[200,300,500,1000,1200,1500,2000]ms`, `maxReconnectAttempts: Infinity`. `PB_CONNECT` event fires on every connect/reconnect — use for state resync. React patterns: `useRealtimeMessages` hook with cleanup, `AuthProvider` context with auto-refresh, optimistic updates with revert. Custom topic subscriptions for presence. `onDisconnect` callback for connection loss detection. Auto-cancellation of duplicate requests (use `requestKey` to disable).

### R-005: LiveKit React SDK — Connection Lifecycle ✅ COMPLETE
**Priority:** High | **Blocks:** H-001, H-003
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-005-livekit-react-sdk.md`](research/R-005-livekit-react-sdk.md)
**Outcome:** Two coexisting API surfaces: `LiveKitRoom` (stable — use this) and `SessionProvider`/`useSession` (beta, agent-focused). Key hooks: `useTracks`, `useParticipants`, `useRemoteParticipants`, `useIsSpeaking`, `useConnectionState`. **Critical spatial audio discovery:** `RemoteAudioTrack.setWebAudioPlugins(nodes: AudioNode[])` (experimental) injects custom Web Audio nodes into LiveKit's internal pipeline (`MediaStreamSource → [plugins] → GainNode → destination`). Custom `PortalAudioRenderer` pattern replaces `RoomAudioRenderer` (which MUST NOT be used for Portal — it renders all audio as default `<audio>` elements). `createAudioAnalyser()` utility for Ember glow visualization.

### R-006: Web Audio API — Spatial Audio for 2D Canvas ✅ COMPLETE
**Priority:** High | **Blocks:** H-002, H-005
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-006-web-audio-spatial.md`](research/R-006-web-audio-spatial.md)
**Outcome:** **PannerNode with `distanceModel: 'linear'`** — the only distance model that reaches actual silence at `maxDistance` (critical for "out of hearing range"). Formula: `gain = 1 - rolloffFactor × (distance - refDistance) / (maxDistance - refDistance)`. `panningModel: 'equalpower'` for stereo (cheaper than HRTF). Z=0 for 2D canvas. Canvas pixels as audio coordinate units directly (`refDistance: 50px`, `maxDistance: 500px`, `rolloffFactor: 1`). Complete `useSpatialAudio` hook with per-participant audio chains. Integration via `RemoteAudioTrack.setWebAudioPlugins([pannerNode])` — plugs directly into R-005 findings. Ember glow via `AnalyserNode` with `cloneTrack: true`. Performance: ~2% CPU at 20 participants (Safari needs `webkitAudioContext` polyfill).

### R-007: Organic Sound Library — Foley Sources
**Priority:** Medium | **Blocks:** K-020, K-021
**Question:** Where to source royalty-free organic sounds (wooden clicks, cork pops, fire crackle, rain, bell chimes) that fit the "Subtle Warmth" aesthetic?
**Context:** Need both one-shot interaction sounds and loopable ambient textures. Must be royalty-free / CC0. Candidates: Freesound.org, Sonniss GDC bundles, Zapsplat.
**Deliverable:** Curated asset list with license verification.

### R-008: CSS Animation Performance — Fading at Scale ✅ COMPLETE
**Priority:** Medium | **Blocks:** K-012
**Status:** Complete — 2026-02-11 | **Report:** [`docs/research/R-008-css-animation-performance.md`](research/R-008-css-animation-performance.md)
**Outcome:** CSS `opacity` animations run on the compositor thread (GPU-accelerated), safely handling 200+ concurrent fades at 60 FPS on desktop. Real constraint is GPU memory from layer promotion (~128KB per 400×80 message), not CPU. `content-visibility: auto` (Baseline 2024, all major browsers) provides browser-native virtualization with zero JS overhead. `animationend` event for DOM cleanup keeps active message count bounded. **No JS virtualization library needed.** TanStack Virtual deferred as documented fallback. `will-change` explicitly rejected (browser auto-promotes). **Campfire CSS decay engine is unblocked.**

### R-009: Research Gap Analysis — Ideas We Left Behind ✅ COMPLETE
**Priority:** High | **Blocks:** ADR-007, Sprint 4 planning
**Status:** Complete — 2026-02-11 | **Report:** [`docs/research/R-009-research-gap-analysis.md`](research/R-009-research-gap-analysis.md)
**Outcome:** Compared original research reports against master plan. 80% of ideas carried forward. Key findings: (1) We built Discord's sidebar layout despite research warning against it — House navigation model needed for v1.1. (2) Room type enum was in original research — carried into ADR-007. (3) Docker vs systemd tradeoff was undocumented pivot. (4) Ghost Text Echo stage lost blur+gray shift. (5) Video cap (480p) undefined for voice rooms. Full ranked list of 10 lost ideas with sprint targets.

### R-010: Hearth Persona — Cross-Server Identity (DID)
**Priority:** Low | **Blocks:** O-020, O-021
**Question:** How to implement DID-based portable identity so users can carry their persona across multiple Houses (Hearth instances)?
**Context:** Users should be able to prove they're "the same person" across Houses without a central authority. DIDs (Decentralized Identifiers) offer self-sovereign identity. Need to evaluate: did:key, did:web, did:peer methods. Key questions: storage of DID documents, key rotation, revocation, trust establishment between Houses.
**Deliverable:** Research report with recommended DID method, implementation architecture, and library evaluation.
**Timing:** v2.0 — not blocking any current sprint.
**Status:** Not started.

### R-011: Chat E2EE for Campfires + DMs
**Priority:** High | **Blocks:** F-020 through F-024
**Question:** How to implement client-side encryption for Campfire messages and DMs in a PocketBase backend?
**Context:** ADR-007 approved E2EE for Campfires + DMs at v1.0, Dens deferred to v2.0. Need to evaluate: Signal Protocol (double ratchet) vs simpler approaches, key exchange mechanism (leveraging `public_key` on user records), key management for ephemeral Campfire messages (key lifespan = message lifespan), group key distribution for Campfires (multiple participants), backward secrecy implications for new Campfire joiners.
**Special constraint:** Campfires are ephemeral — keys can be short-lived. DMs are permanent — keys must support long-term storage.
**Deliverable:** Research report with protocol choice, implementation architecture, library evaluation (libsignal-protocol-javascript, tweetnacl), and PocketBase integration pattern.
**Timing:** Pre-v1.0 — needed before F-020.
**Status:** Not started.

### R-009-M: Pre-Alpha Marketing Prep — Reddit Post & Community Strategy
**Priority:** Medium | **Blocks:** M-001 through M-006
**Question:** What's the optimal framing, post structure, subreddit targeting, and visual assets needed for Hearth's first public reveal? How do competing projects (Revolt, Element, Mumble) position themselves, and how do we differentiate?
**Deliverable:** Reddit post draft, "Why not X?" FAQ, competitive positioning guide, README polish spec, visual asset checklist.
**Timing:** Begin research when v0.2 (Kindling) frontend is screenshottable. Post targets end of v0.2.
**Status:** Post structure drafted. See [`docs/specs/marketing-reddit-draft.md`](specs/marketing-reddit-draft.md). Full research (competitive analysis, FAQ, README) scheduled for v0.2 completion.

---

## Open Questions (Unresolved Design Decisions)

### Q-001: Matrix Protocol Integration?
**Source:** UX Research Report (Section 5.1)
**Question:** Should Hearth use the Matrix protocol under the hood (for federation and protocol-level encryption), hiding it behind a friendly UI?
**Tradeoff:** Federation lets Hearth instances talk to each other — huge for adoption. But Matrix adds massive complexity (Synapse/Dendrite server, DAG-based event resolution, federation key management). This likely explodes the 1GB RAM target.
**Current Lean:** No. Build native PocketBase-first. Revisit federation in v3+ if demand exists.
**Status:** Parked — needs formal ADR if we revisit.

### Q-002: Generative Ambience Engine
**Source:** UX Research Report (Section 3.4)
**Question:** How to implement lightweight procedural audio (fire crackle, rain, coffee shop murmur) without large asset downloads?
**Options:**
1. Pre-recorded loops (simple, ~2-5MB per texture, boring)
2. Web Audio API oscillators + noise generators (zero download, needs design work)
3. Tiny ML model generating ambient textures (novel, likely too CPU-heavy)
**Status:** Needs research spike (R-007 covers asset sourcing; this is about procedural generation)

### Q-003: Screenshot Prevention
**Source:** UX Research Report (Section 2.4 — "Drunk Test")
**Question:** Can we detect or prevent screenshots in a web browser?
**Reality:** No reliable cross-browser screenshot detection exists for web apps. The `visibilitychange` API can detect tab switches (possible screen recording) but not OS-level screenshots.
**Current Lean:** Accept limitation. Rely on visual affordances (fading text "feels" impermanent) and culture. Document as a known limitation.
**Status:** Parked.

### Q-004: Video Policy ✅ RESOLVED
**Source:** Technical Research (Section 3.3)
**Question:** When and how to enable video beyond voice-first default?
**Answer:** ADR-007 resolves this: 480p max, 15fps, simulcast disabled, dynacast enabled. `canPublishVideo: false` by default in JWT. Homeowner/Keyholder enables per-Den.
**Resolved by:** ADR-007 (2026-02-11).

### Q-005: Plugin Marketplace
**Source:** Master Plan (Section 11)
**Question:** How to discover and distribute Cartridges? Curated vs. open? Signing requirements?
**Status:** Deferred to v2.0.

### Q-006: Accessibility — Spatial Audio & Fading Text
**Source:** Master Plan (Section 11)
**Question:** How do spatial audio and transparency-decay text work for:
- Screen reader users? (Fading text is purely visual — need ARIA `live` announcements?)
- Hearing-impaired users? (Spatial audio is meaningless — need captions/visual indicators?)
- Motor-impaired users? (Click-to-drift needs keyboard alternative)
**Status:** Needs research (R-xxx) — scheduled for v1.1 but should be considered early to avoid costly retrofitting.
**Priority:** Should be elevated — accessibility as an afterthought is a debt bomb.

### Q-007: PocketBase Scaling Ceiling
**Question:** At what user count does the single-PocketBase + SQLite model break? Is it 20 concurrent? 50? 100?
**Context:** The spec targets ~20 concurrent users. We need real load testing to know the ceiling and identify the first bottleneck (CPU? WAL contention? WebSocket fan-out?).
**Status:** Scheduled for v1.0 performance profiling (F-013).

### Q-008: LiveKit Host Network Mode + Docker Compose ✅ RESOLVED
**Question:** LiveKit documentation recommends `network_mode: host` for WebRTC UDP performance. This conflicts with Docker Compose networking (Caddy can't reach LiveKit via `vesta-net` if LiveKit is on the host network).
**Answer:** **Option 2 — ALL containers on host network.** LiveKit's official Docker Compose template (from `livekit/deploy` repo) uses `network_mode: "host"` for every container (Caddy, LiveKit, Redis). All services communicate via `localhost`. This eliminates the bridge/host conflict entirely. Trade-off is no Docker network isolation, which is acceptable since all services are trusted.
**Resolved by:** R-002 (2026-02-10). See [`R-002-caddy-livekit-tls-config.md`](research/R-002-caddy-livekit-tls-config.md).

---

## Research Completion Log

| ID | Topic | Date Completed | Outcome |
|----|-------|---------------|---------|
| R-001 | PocketBase v0.23+ API Verification | 2026-02-10 | PocketBase v0.36.2 API fully documented. Deprecated→current migration table. 10 verified code patterns. Backend unblocked. |
| R-002 | Caddy + LiveKit TLS Configuration | 2026-02-10 | Layer 4 TLS SNI routing via custom Caddy build. YAML config (not Caddyfile). Host networking for all containers. Redis-free single-node. Complete deployment templates. |
| R-003 | Container Topology ADR-001 | 2026-02-10 | Docker Compose with 3 containers, all `network_mode: "host"`. ADR-001 formally accepted. PIVOT-001 retired. |
| R-004 | PocketBase JS SDK Real-time | 2026-02-10 | SSE-based realtime, auto-reconnect, PB_CONNECT resync. React hooks for subscriptions, auth, optimistic updates. |
| R-005 | LiveKit React SDK Lifecycle | 2026-02-10 | Two API surfaces (LiveKitRoom stable, SessionProvider beta). `RemoteAudioTrack.setWebAudioPlugins()` for spatial audio. Custom PortalAudioRenderer pattern. |
| R-006 | Web Audio Spatial Audio (2D) | 2026-02-10 | PannerNode linear distance model, equalpower panning, Z=0 for 2D. Complete `useSpatialAudio` hook. ~2% CPU at 20 participants. |
| R-008 | CSS Animation Performance (Fading) | 2026-02-11 | Compositor-thread GPU animation safe at 200+ concurrent. `content-visibility: auto` for browser-native virtualization. No JS libs needed. TanStack Virtual deferred as fallback. |
| R-009 | Research Gap Analysis | 2026-02-11 | Compared research reports vs master plan. 80% carry-through. Key drops: sidebar layout warning, room type enum, Docker pivot, Ghost Text detail. Top 10 ranked lost ideas. ADR-007 informed. |
