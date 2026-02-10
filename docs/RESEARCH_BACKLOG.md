# Hearth — Research Backlog & Open Questions

> **Last Updated:** 2026-02-10
> **Owner:** Researcher Agent

---

## Active Research Tasks

### R-001: PocketBase v0.23+ API Verification ✅ COMPLETE
**Priority:** Critical | **Blocks:** E-004, E-010 through E-015
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-001-pocketbase-api-verification.md`](research/R-001-pocketbase-api-verification.md)
**Outcome:** PocketBase v0.36.2 API fully verified. Current correct patterns: `app.OnServe().BindFunc()` for startup hooks, `app.DB()` for raw SQL via dbx builder, `app.Cron().MustAdd()` for scheduling, `app.FindRecordById()`/`app.FindRecordsByFilter()` for records, `core.NewRecord()` + `app.Save()` for creation, `app.RunInTransaction()` for atomic ops, `app.SubscriptionsBroker()` for custom realtime. SQLite PRAGMAs injected via `app.OnBootstrap()`. Deprecated patterns fully documented with migration table. **Backend is now unblocked.**

### R-002: Caddy Reverse Proxy for PocketBase + LiveKit TLS ✅ COMPLETE
**Priority:** Critical | **Blocks:** E-005, E-006
**Status:** Complete — 2026-02-10 | **Report:** [`docs/research/R-002-caddy-livekit-tls-config.md`](research/R-002-caddy-livekit-tls-config.md)
**Outcome:** Major discovery — LiveKit's official deployment uses a **custom Caddy build** (`livekit/caddyl4`) with Layer 4 TLS SNI routing, NOT a standard Caddyfile. Config is YAML, not Caddyfile syntax. Architecture: Caddy listens :443, routes by SNI to TURN (localhost:5349), LiveKit API (localhost:7880), and PocketBase (localhost:8090). All containers use `network_mode: "host"`. Three subdomains: `hearth.example` / `lk.hearth.example` / `turn.hearth.example`. **Redis is NOT required for single-node** (saves ~25MB RAM). Complete `caddy.yaml`, `livekit.yaml`, and `docker-compose.yaml` templates provided. Q-008 resolved. **Docker deployment is now unblocked.**

### R-003: Container Topology Decision (ADR-001)
**Priority:** High | **Blocks:** E-005
**Question:** Single container (supervisor process) vs. Docker Compose (3 containers: PocketBase + LiveKit + Caddy)?
**Context:** The master plan says "Single Docker Container" but the actual deployment likely needs 3 processes (PocketBase, LiveKit, Caddy). Options:
1. **Single container** with supervisord/s6-overlay running all 3 processes — simplest UX for self-hosters (`docker run` one thing)
2. **Docker Compose** with 3 containers — cleaner process isolation, standard Docker practice, but more complex setup
3. **Hybrid** — PocketBase + Caddy in one container, LiveKit in host network mode (needed for UDP perf)
**Deliverable:** ADR-001 with recommendation.

### R-004: PocketBase JS SDK — Real-time Subscriptions
**Priority:** High | **Blocks:** K-002, K-010
**Question:** How does the PocketBase JavaScript SDK handle real-time record subscriptions, auth token refresh, and reconnection?
**Deliverable:** Integration guide for React with patterns for hooks, subscription cleanup, and optimistic updates.

### R-005: LiveKit React SDK — Connection Lifecycle
**Priority:** High | **Blocks:** H-001, H-003
**Question:** What's the current LiveKit React SDK (`@livekit/components-react`) API for room connection, track management, and spatial audio configuration?
**Deliverable:** Integration guide covering connection, publish/subscribe, and audio processing hooks.

### R-006: Web Audio API — Spatial Audio for 2D Canvas
**Priority:** High | **Blocks:** H-002, H-005
**Question:** How to implement proximity-based volume attenuation using Web Audio API's PannerNode or GainNode controlled by 2D canvas position?
**Context:** We need distance → gain mapping, not full 3D HRTF. Options:
- Simple GainNode with linear/exponential rolloff based on Euclidean distance
- PannerNode with `distanceModel: 'inverse'` in a flattened 3D space
- Custom curve: `volume = 1 - clamp(distance / maxRange, 0, 1)` with smoothing
**Deliverable:** Prototype + recommended approach.

### R-007: Organic Sound Library — Foley Sources
**Priority:** Medium | **Blocks:** K-020, K-021
**Question:** Where to source royalty-free organic sounds (wooden clicks, cork pops, fire crackle, rain, bell chimes) that fit the "Subtle Warmth" aesthetic?
**Context:** Need both one-shot interaction sounds and loopable ambient textures. Must be royalty-free / CC0. Candidates: Freesound.org, Sonniss GDC bundles, Zapsplat.
**Deliverable:** Curated asset list with license verification.

### R-008: CSS Animation Performance — Fading at Scale
**Priority:** Medium | **Blocks:** K-012
**Question:** What are the browser limits for concurrent CSS animations? If 200 messages are visible and all fading simultaneously, does compositor-thread rendering hold up?
**Deliverable:** Benchmark results + recommendation (batch animations, virtualize old messages, or trust the compositor).

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

### Q-004: Video Policy
**Source:** Technical Research (Section 3.3)
**Question:** When and how to enable video beyond voice-first default?
**Options:**
1. Host-controlled room toggle (`allowVideo: true/false`)
2. Per-user permission escalation (host grants `canPublishVideo` to specific users)
3. "Picture frame" mode — low-res, low-fps ambient video (webcam as portrait, not video call)
**Status:** Deferred to v0.3 research.

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
