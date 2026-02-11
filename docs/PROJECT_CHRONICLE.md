# Hearth (Project Vesta) — Project Chronicle

> **Purpose:** Institutional knowledge capture for AI-assisted development. This document records bugs, architecture decisions, failed approaches, lessons learned, and backburnered ideas throughout Hearth's development.
>
> **For Copilot Agents:** Reference this document when working on Hearth to avoid repeating past mistakes and understand why things are built the way they are. **Ask the user before modifying:** SQLite pragma config, HMAC crypto, memory budget allocations, LiveKit config, or Docker resource limits.
>
> **Version:** 1.0 (February 10, 2026) — Project kickoff. Research & Exploration phase. No production code yet.

---

## 📊 Project Status Dashboard

| Component | Status | Notes |
|-----------|--------|-------|
| **Master Specification** | ✅ Complete | `vesta_master_plan.md` — 11 sections, fully expanded |
| **Technical Research** | ✅ Complete | PocketBase/SQLite, LiveKit, Extism/Wasm, security |
| **UX Research** | ✅ Complete | Spatial audio, ephemeral messaging, cozy UI, onboarding |
| **Release Roadmap** | ✅ Complete | 6 releases (v0.1 Ember → v2.0 Open Flame) with task IDs |
| **Research Backlog** | ✅ Complete | 8 research tasks: R-001 through R-006 + R-008 ✅ complete. R-007 remaining (medium priority, blocks sound only). 7 open questions. |
| **Agent Roles** | ✅ Complete | Builder, Researcher, Reviewer — specialized for Hearth |
| **Backend (PocketBase)** | ✅ Complete | v0.1 Ember shipped. Auth, CRUD, GC, presence, HMAC, PoW, LiveKit JWT. 36/36 tests. |
| **Frontend (React/Vite)** | ✅ Complete | v0.2 Kindling shipped. 27 files, 91KB gzipped, Subtle Warmth design system, Campfire chat. |
| **Voice (LiveKit)** | 🟡 Unblocked | R-005 (React SDK) + R-006 (Web Audio spatial) complete — ready for v0.3 |
| **Docker Deployment** | ✅ Complete | PocketBase serves SPA via `pb_public/`. Standalone `Dockerfile.frontend` also available. |
| **Plugin System (Extism)** | 🔲 Not Started | Scheduled for v2.0 |

### Current Milestone
- **v0.1 — Ember** (Backend skeleton + chat API) — COMPLETE ✅
- **v0.2 — Kindling** (Frontend + Campfire chat) — COMPLETE ✅
- Next: **v0.3 Hearth Fire** (Voice — The Portal) — Target: Aug 2026
- Marketing prep: Post #1 (r/selfhosted concept pitch) targeted for end of v0.2 polish

---

## 📚 Reference Documents

| Document | Purpose |
|----------|---------|
| [`vesta_master_plan.md`](../vesta_master_plan.md) | Master design & technical specification |
| [`docs/ROADMAP.md`](ROADMAP.md) | Release roadmap with versioned milestones and task IDs |
| [`docs/RESEARCH_BACKLOG.md`](RESEARCH_BACKLOG.md) | Active research tasks and open questions |
| [`docs/research/Hearth Technical Research.md`](research/Hearth%20Technical%20Research.md) | Technical deep-dive (SQLite, LiveKit, Extism, security) |
| [`docs/research/Hearth_ Digital Living Room UX Report.md`](research/Hearth_%20Digital%20Living%20Room%20UX%20Report.md) | UX research (spatial audio, fading text, cozy UI, The Knock) |
| [`.github/copilot-instructions.md`](../.github/copilot-instructions.md) | AI development context, coding standards |
| [`.github/agents/Builder.md`](../.github/agents/Builder.md) | Builder agent role (implementation specialist) |
| [`.github/agents/Researcher.md`](../.github/agents/Researcher.md) | Researcher agent role (tech investigation) |
| [`.github/agents/Reviewer.md`](../.github/agents/Reviewer.md) | Reviewer agent role (QA + sprint coordination) |

---

## Table of Contents

1. [Project Timeline](#project-timeline)
2. [Architecture Overview](#architecture-overview)
3. [Architecture Decisions](#architecture-decisions)
4. [Bug Registry](#bug-registry)
5. [Production Incidents](#production-incidents)
6. [Failed Approaches & Pivots](#failed-approaches--pivots)
7. [Magic Numbers & Configuration](#magic-numbers--configuration)
8. [External Service Learnings](#external-service-learnings)
9. [Backburner Ideas](#backburner-ideas)
10. [Communication Patterns](#communication-patterns)

---

## Project Timeline

| Date | Milestone | Notes |
|------|-----------|-------|
| 2026-02-10 | **Project Kickoff** | Initial discussion with Gemini 3 Pro. Philosophy, tech stack, and constraints defined. |
| 2026-02-10 | **Master Spec v1** | `vesta_master_plan.md` stub created — mission, tech stack, core features, design system |
| 2026-02-10 | **Research Collection** | Two comprehensive research reports completed (technical + UX) |
| 2026-02-10 | **Repo Initialized** | Git repo, `.gitignore`, `copilot-instructions.md` |
| 2026-02-10 | **Spec Expansion** | Master plan expanded to 11 sections incorporating all research findings |
| 2026-02-10 | **Gemini Review** | External review validated co-located monolith, CSS decay engine, click-to-drift. Identified SSL/TLS gap → Caddy added to stack. Flagged stale PocketBase API. |
| 2026-02-10 | **Planning Artifacts** | Release roadmap (6 versions), research backlog (8 tasks, 8 questions), specialized agent roles |
| 2026-02-10 | **Project Chronicle** | This document — institutional knowledge capture begins |
| 2026-02-10 | **R-001 Complete** | PocketBase v0.36.2 API verified. `app.OnServe()`, `app.DB()`, `app.Cron().MustAdd()` confirmed. Deprecated API migration table documented. Backend unblocked. |
| 2026-02-10 | **R-002 Complete** | Critical discovery: LiveKit uses custom Caddy build with Layer 4 TLS SNI routing (YAML config, not Caddyfile). All Docker containers use `network_mode: host`. Redis optional for single-node. Complete deployment templates produced. Q-008 resolved. |
| 2026-02-10 | **R-003 Complete** | ADR-001 formally accepted: Docker Compose with 3 containers, all `network_mode: "host"`. Pre-resolved by R-002 findings. PIVOT-001 (single container → compose) confirmed. |
| 2026-02-10 | **R-004 Complete** | PocketBase JS SDK integration guide. SSE-based realtime (NOT WebSocket), auto-reconnect with backoff, PB_CONNECT resync event. React hooks for subscriptions, auth provider, optimistic updates. |
| 2026-02-10 | **R-005 Complete** | LiveKit React SDK guide. Two API surfaces: `LiveKitRoom` (stable) and `SessionProvider` (beta). **Key discovery:** `RemoteAudioTrack.setWebAudioPlugins()` — experimental API to inject Web Audio nodes into LiveKit's audio pipeline. Custom `PortalAudioRenderer` pattern (must NOT use `RoomAudioRenderer` for Portal). |
| 2026-02-10 | **R-006 Complete** | Web Audio spatial audio for 2D canvas. `PannerNode` with `linear` distanceModel (only model that reaches true silence). `equalpower` panning, Z=0 for 2D. Complete `useSpatialAudio` hook. ~2% CPU at 20 participants. Ember glow via `AnalyserNode`. |
| 2026-02-11 | **Sprint 1 Spec** | `docs/specs/sprint-1-ember.md` — 4 phases, 20 subtasks for Builder. Covers scaffolding, data layer (collections, GC, presence), auth & security (HMAC, PoW, LiveKit JWT), observability (metrics, logging), and testing. |
| 2026-02-11 | **Builder Implementation Review** | All Sprint 1 backend code reviewed and verified. **Critical fix:** `go.mod` PocketBase version corrected from v0.26.6 → v0.36.2 (was preventing compilation). LiveKit protocol updated v1.24.0 → v1.44.0, dbx v1.11.0 → v1.12.0. Two type mismatch bugs fixed (`gcDeletedTotal.Add` needed `int64` not `float64`; `countRecords` return type `int` → `int64`). Go 1.24.0 minimum enforced by `go mod tidy`. **Result:** `go build` ✅ clean, 36/36 tests passing ✅. |
| 2026-02-11 | **R-008 Complete** | CSS animation performance for Campfire fading at scale. Compositor-thread GPU animation safe at 200+ concurrent fades. `content-visibility: auto` (Baseline 2024) for browser-native virtualization. No JS virtualization needed. TanStack Virtual deferred as fallback. `will-change` rejected. Performance budget defined: <300 DOM messages, <50 animating, <30 GPU layers. |
| 2026-02-11 | **Sprint 2 Spec** | `docs/specs/sprint-2-kindling.md` — 5 phases covering frontend scaffolding (Vite+React+TS+Tailwind), auth & SSE reconnect (SEC-005/SEC-006), Campfire chat (CSS decay engine, real-time messages, mumbling indicator, presence), mobile responsive layout, and Docker frontend build. ~13 days estimated. |
| 2026-02-11 | **R-009 Initiated + Marketing Draft** | Pre-alpha marketing strategy formalized. Reddit post structure drafted (`docs/specs/marketing-reddit-draft.md`). Two-post plan: concept pitch at end of v0.2, "try it" post at v1.0. Target communities: r/selfhosted (primary), r/privacy, r/opensource, HN (deferred). Competitive response playbook for Matrix/Revolt/Mumble. Marketing phase (M-001–M-006) added to roadmap. |
| 2026-02-11 | **Sprint 2 Implementation Complete (v0.2 Kindling)** | Builder delivered 27 files. Frontend shell: Vite + React 19 + TypeScript strict + Tailwind v4. Subtle Warmth design system fully implemented (dark + light mode, @fontsource fonts, pillow buttons, candlelight shadows). Campfire chat: SSE real-time subscription, 4-stage CSS fade (`campfire.css`), negative `animation-delay` for mid-fade page loads, `animationend` DOM cleanup, optimistic send with revert, time sync via Date header RTT/2, mumbling indicator (CSS bars), heartbeat presence (30s). Auth: `AuthProvider` context, token auto-refresh, `useReconnect` with `PB_CONNECT` resync (SEC-006 ✅). Code-splitting via `React.lazy` (K-024). PocketBase serves SPA from `pb_public/`. **Build:** `tsc --noEmit` clean, Vite 1.76s, ~91KB gzipped (under 150KB budget). **Deferred:** SEC-005 httpOnly cookies (PB SDK limitation), typing broadcast (needs backend topic), mobile drawer, error toasts, sound/foley (R-007). |

---

## Architecture Overview

### System Topology

```
┌──────────────────────────────────────────────────────────────┐
│                    1 vCPU / 1GB RAM VPS                       │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    Caddy (~10MB)                         │ │
│  │  Auto-TLS (Let's Encrypt)                               │ │
│  │  api.hearth.example → PocketBase :8090                  │ │
│  │  lk.hearth.example  → LiveKit WS  :7880                │ │
│  └────────────────────────┬────────────────────────────────┘ │
│                           │                                   │
│  ┌────────────────────────┴────────────────────────────────┐ │
│  │                                                          │ │
│  │  ┌──────────────────┐    ┌───────────────────────────┐  │ │
│  │  │   PocketBase     │    │     LiveKit SFU           │  │ │
│  │  │   (250MB heap)   │    │     (400MB heap)          │  │ │
│  │  │                  │    │                           │  │ │
│  │  │  • Auth          │    │  • WebRTC (UDP)           │  │ │
│  │  │  • Chat API      │    │  • Opus audio forwarding  │  │ │
│  │  │  • WebSocket     │    │  • ICE Lite               │  │ │
│  │  │  • Cron (GC)     │    │  • DTX (silence suppress) │  │ │
│  │  │  • Wasm plugins  │    │  • Dynacast               │  │ │
│  │  │                  │    │                           │  │ │
│  │  │  ┌────────────┐  │    └───────────────────────────┘  │ │
│  │  │  │  SQLite    │  │                                    │ │
│  │  │  │  (WAL mode)│  │    ┌───────────────────────────┐  │ │
│  │  │  │  ~50MB     │  │    │  Wasm Plugin Pool (50MB)  │  │ │
│  │  │  └────────────┘  │    │  Extism / Wasmtime        │  │ │
│  │  └──────────────────┘    └───────────────────────────┘  │ │
│  │                                                          │ │
│  │              OS Headroom: ~100MB                          │ │
│  └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### Memory Budget (1GB Total — Sacred)

| Component | Allocation | Control |
|-----------|-----------|---------|
| OS Kernel & System | 150 MB | Minimal Alpine/Debian |
| PocketBase (Heap) | 250 MB | `GOMEMLIMIT=250MiB` |
| LiveKit SFU (Heap) | 400 MB | `GOMEMLIMIT=400MiB` |
| Wasm Plugin Pool | 50 MB | Fixed instance pool |
| SQLite Page Cache | 50 MB | `PRAGMA cache_size` |
| Safety Headroom | 100 MB | Prevent OOM kill |

### Data Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Browser    │    │   Browser    │    │   Browser    │
│   (React)    │    │   (React)    │    │   (React)    │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │ WSS/HTTPS         │ WSS/HTTPS         │ UDP (WebRTC)
       ▼                   ▼                   ▼
┌──────────────────────────────────────────────────────┐
│                      Caddy                            │
│              (TLS termination)                        │
└──────────┬───────────────────────────┬───────────────┘
           │                           │
    ┌──────▼──────┐            ┌───────▼──────┐
    │  PocketBase  │            │   LiveKit    │
    │  (Chat/Auth) │◄──JWT──────│   (Voice)    │
    │  + SQLite    │            │   (SFU)      │
    └──────────────┘            └──────────────┘
```

---

## Architecture Decisions

### ADR-001: Container Topology — Docker Compose with Host Networking

**Date:** February 10, 2026 | **Status:** ✅ Accepted (R-003)

**Context:** The master plan originally said "Single Docker Container." Actual deployment needs 3 processes: PocketBase, LiveKit, Caddy. LiveKit requires `network_mode: host` for WebRTC UDP hole punching and TURN port binding.

**Options Evaluated:**

| Option | Pros | Cons | Verdict |
|--------|------|------|--------|
| Single container + s6-overlay | Simplest UX (`docker run` one thing) | Non-standard, harder to debug, process supervision complexity, opaque failure modes | ❌ Rejected |
| Docker Compose (3 containers) | Standard Docker practice, clean isolation, independent restarts, standard logging | Three-container UX (mitigated by simple `.env` + `docker compose up -d`) | ✅ Accepted |
| Hybrid (PB+Caddy in one, LiveKit on host) | Balances simplicity + LiveKit UDP perf | Two different networking modes, more complex networking, confusing to debug | ❌ Rejected |

**Decision:** Docker Compose with 3 containers (Caddy, PocketBase, LiveKit), all using `network_mode: "host"`. All inter-service communication via localhost. Pre-resolved by R-002 discovery that LiveKit's own official deployment template uses this exact pattern.

**Self-Hoster UX:** `git clone → cp .env.example .env → edit .env → docker compose up -d`

**Full spec:** [`docs/research/R-003-container-topology.md`](research/R-003-container-topology.md)

---

### ADR-002: PocketBase as Backend Framework

**Date:** February 10, 2026 | **Status:** Accepted

**Decision:** Use PocketBase (Go + embedded SQLite) as the sole backend framework.

**Rationale:**
- Single Go binary with embedded DB eliminates TCP database protocol overhead
- Built-in auth, real-time subscriptions, file storage, admin UI
- SQLite in WAL mode provides concurrent reads with minimal RAM
- Go's `GOMEMLIMIT` prevents OOM in constrained environments
- No need for Redis, PostgreSQL, or any external dependency

**Risk:** PocketBase is a single-maintainer project. If abandoned, we'd need to fork or migrate. Mitigated by: it's open-source Go, well-structured, and our usage is straightforward.

---

### ADR-003: CSS-Driven Visual Decay (Not JavaScript)

**Date:** February 10, 2026 | **Status:** Accepted

**Decision:** Message fading (transparency decay) is driven entirely by CSS animations, not JavaScript timers.

**Rationale:**
- CSS animations are compositor-thread (often GPU-accelerated) and "free" in terms of CPU/battery
- `setInterval`/`setTimeout` keep the main thread awake and drain mobile batteries
- Negative `animation-delay` allows mid-fade rendering on page reload (e.g., 30s into a 60s message → renders at 50% opacity)
- This is a core optimization identified in both the technical research and Gemini review

**Implementation Pattern:**
```jsx
const style = {
  animationName: 'fadeOut',
  animationDuration: `${ttl}s`,
  animationDelay: `-${age}s`,
  animationTimingFunction: 'linear',
  animationFillMode: 'forwards'
};
```

**Open Question:** R-008 — performance at scale (200 concurrent CSS animations). Needs benchmarking.

---

### ADR-004: Stateless HMAC Invite Tokens

**Date:** February 10, 2026 | **Status:** Accepted

**Decision:** Invite links are self-validating HMAC tokens with zero database storage.

**Rationale:**
- No DB writes on invite creation (can't flood the database)
- Validation requires only CPU (one hash operation) — extremely fast
- Secret key rotation instantly invalidates all outstanding invites
- Two-key system (current + old) provides grace period during rotation

**Implementation:**
```
URL: https://hearth.example/join?r=room1&t=1735689600&s=f8a...
Validate: HMAC_SHA256(secret, r + "." + t) == s (constant-time compare)
```

**Constraint:** MUST use `crypto/subtle.ConstantTimeCompare` in Go. Never `==` or `bytes.Equal` for hash comparison — timing side-channel risk.

---

### ADR-005: Caddy for TLS Termination (Layer 4 SNI Routing)

**Date:** February 10, 2026 | **Status:** Accepted — Updated with R-002 findings

**Decision:** Use a **custom Caddy build** (`caddy-l4` + `caddy-yaml` modules) as a Layer 4 TLS SNI router for all three services.

**Context:** WebRTC requires HTTPS/WSS — browsers block microphone access on insecure origins (except localhost). PocketBase has built-in Auto-TLS, but LiveKit also needs TLS for its WebSocket signaling endpoint. Identified during Gemini 3 Pro review. **R-002 research revealed that TURN traffic is NOT HTTP — it requires Layer 4 (raw TCP/TLS) proxying, which a standard Caddyfile `reverse_proxy` cannot handle.**

**Architecture:**
```
Internet :443 → Caddy L4 (TLS SNI) → turn.hearth.example → LiveKit TURN :5349
                                      → lk.hearth.example   → LiveKit API  :7880
                                      → hearth.example       → PocketBase   :8090
```

**Rationale:**
- Caddy uses near-zero RAM (~15-20MB) — fits in the OS headroom budget
- Automatic Let's Encrypt certificate management for all three subdomains
- Layer 4 SNI routing enables TURN (non-HTTP TLS) on the same port 443 as HTTPS
- TURN route passes raw TLS stream to LiveKit; HTTP routes terminate TLS at Caddy
- PocketBase's built-in TLS can be disabled to avoid double-termination

**Required Custom Build:**
```
xcaddy build --with github.com/abiosoft/caddy-yaml --with github.com/mholt/caddy-l4
```

**Config Format:** YAML (loaded with `caddy run --config caddy.yaml --adapter yaml`), not Caddyfile.

**Key Constraint:** TURN cert path is fragile — LiveKit must read Caddy-managed cert files. Shared Docker volume required.

**Full spec:** [`docs/research/R-002-caddy-livekit-tls-config.md`](research/R-002-caddy-livekit-tls-config.md)

**Open Question:** LiveKit cert hot-reload after Caddy auto-renewal (may need SIGHUP or container restart on cert change).

---

## Bug Registry

| ID | Severity | Component | Title | Status | Date Found | Date Fixed |
|----|----------|-----------|-------|--------|-----------|-----------|
| BUG-001 | LOW | `message_gc.go` | `gcDeletedTotal.Add()` called with `float64` but `atomic.Int64.Add` requires `int64` | ✅ Fixed | 2026-02-11 | 2026-02-11 |
| BUG-002 | LOW | `metrics.go` | `countRecords()` returned `int` but `app.CountRecords()` returns `int64` — type mismatch | ✅ Fixed | 2026-02-11 | 2026-02-11 |
| BUG-003 | CRITICAL | `go.mod` | PocketBase version declared as v0.26.6 (nonexistent version matching code's API) — should be v0.36.2. Code used v0.36+ patterns (OnServe, OnBootstrap, core.NewBaseCollection). Prevented compilation. | ✅ Fixed | 2026-02-11 | 2026-02-11 |
| BUG-004 | CRITICAL | `pragmas.go` | `OnBootstrap` hook called `e.App.DB()` before `e.Next()`. DB is `nil` at this point — PocketBase opens the database *during* the bootstrap chain. Panic on startup. Fix: call `e.Next()` first, then apply PRAGMAs, then return `nil`. | ✅ Fixed | 2026-02-11 | 2026-02-11 |
| BUG-005 | CRITICAL | `collections.go` | Rooms API rules reference `room_members_via_room` back-relation, but `room_members` collection doesn't exist yet during `rooms` creation. PocketBase validates rules during `Save()` and fails silently — collections never created. Fix: two-pass approach — create all collections without rules, then apply rules after all exist. | ✅ Fixed | 2026-02-11 | 2026-02-11 |

---

## Builder Implementation Decisions (Sprint 1)

> These are design decisions Builder made during Sprint 1 implementation that are not captured in the original spec. Documented here for institutional knowledge.

| Decision | Rationale | Verified |
|----------|-----------|----------|
| `RelationField.CollectionId` uses collection name strings (e.g., `"users"`, `"rooms"`) instead of runtime IDs | PocketBase resolves name → ID at save time, avoiding bootstrap ordering issues | ✅ Confirmed via PB v0.36.2 API docs |
| `expires_at` enforced server-side in `OnRecordCreate("messages")` hook — clients cannot set their own TTL | Prevents abuse: server reads room's `default_ttl` and computes expiry. Defense-in-depth. | ✅ Auth.go line 50-67 |
| Room creation auto-adds creator as `room_members` with role `"owner"` via `OnRecordAfterCreateSuccess("rooms")` | Ensures every room has at least one owner. Race-free: runs after DB commit. | ✅ Auth.go line 73-89 |
| PoW challenges consumed on verify (one-time use) with 5-min expiry | Prevents challenge reuse attacks. Challenge deleted from map before validation. | ✅ Pow.go line 88-93 |
| Metrics endpoint (`/metrics`) is unauthenticated | Prometheus convention — metrics scraper shouldn't need app auth. Endpoint exposed on internal network only (Docker host networking). | ✅ Accepted — SEC convention |
| `gcDeletedTotal` uses `atomic.Int64` for lock-free increments | Counter updated from cron goroutine, read from HTTP handler. `sync/atomic` avoids mutex contention. | ✅ Thread-safe by design |
| CORS middleware sets headers manually instead of using PocketBase's built-in CORS | Needed explicit origin validation and preflighting control — PB's built-in CORS is too permissive for production. | ✅ Cors.go |
| Rate limiter uses sliding-window token bucket (in-memory) with 5-min sweep cron | No Redis needed. `sync.Mutex`-protected map. Memory bounded by sweep. 5 profiles: auth, invite, message, heartbeat, general. | ✅ 6 tests pass |
| Input sanitization uses `html.EscapeString` (Go stdlib) — no external dependency | Lightweight, covers all HTML entity escaping. Applied to: message body, display_name, room name/description. Max 4000 chars. | ✅ 6 tests pass |

## Builder Implementation Decisions (Sprint 2)

> Design decisions Builder made during Sprint 2 (v0.2 Kindling) implementation.

| Decision | Rationale | Verified |
|----------|-----------|----------|
| PocketBase serves SPA via `pb_public/` — frontend dist baked into PB Docker image | Avoids separate static file server. PB natively serves static files. Caddy still handles TLS. Standalone `Dockerfile.frontend` also provided as alternate. | ✅ Vite build output confirmed |
| SEC-005 httpOnly cookies deferred — using `localStorage` auth store | PB JS SDK doesn't natively support httpOnly cookie auth; would need custom backend proxy hooks. `localStorage` mitigated by CSP `script-src 'self'` (SEC-003). Documented as accepted risk. | ✅ Intentional deferral |
| React Router over TanStack Router | Spec said either works. React Router is simpler, more widely documented, smaller API surface for a 3-page app. | ✅ Standard choice |
| No `will-change: opacity` on `.campfire-message` | Per R-008: browser auto-promotes elements with active CSS animations. Explicit `will-change` wastes GPU memory by creating permanent layers. | ✅ R-008 validated |
| `@fontsource/inter` + `@fontsource/merriweather` (npm, self-hosted) | Privacy-first: no Google Fonts CDN requests. Fonts bundled in build output. | ✅ No external requests |
| Tailwind v4 `@theme` directive for design tokens (not `tailwind.config.js`) | Tailwind v4 moved to CSS-native config. `@theme` in `globals.css` replaces `theme.extend` in config file. | ✅ Tailwind v4 pattern |
| Single `campfire-fade` keyframe with negative `animation-delay` | Per R-008: one animation handles all 4 decay stages (Fresh→Fading→Echo→Gone). Negative delay starts mid-fade for messages already in progress on page load. | ✅ R-008 validated |
| `content-visibility: auto` + `contain-intrinsic-size: auto 80px` on messages | Per R-008: browser-native virtualization. Offscreen messages skip rendering entirely. Zero JS overhead. | ✅ R-008 validated |
| Optimistic messages marked with `data-optimistic="true"` and `animation: none` | Prevents optimistic messages from fading before server confirms. Server-assigned `expires_at` triggers real fade after confirmation. | ✅ Spec pattern |
| `useReconnect` fires `authRefresh()` before data resync on `PB_CONNECT` | Per R-004: reconnect events don't replay missed SSE messages. Auth must be re-validated first (SEC-006), then full state re-fetched. | ✅ SEC-006 resolved |

---
## 🛡️ Security Concerns Tracker

> **Purpose:** Dedicated tracking for security items to ensure nothing is forgotten. These are flagged with elevated visibility because security is foundational to Hearth's privacy-first promise.
>
> **Rule:** No security concern leaves this table without being either **resolved** (with a task ID and date) or **explicitly accepted as a known risk** with documented rationale.

### Resolved in Sprint 1 (v0.1)

| ID | Concern | Severity | Resolution | Task ID | Date |
|----|---------|----------|-----------|---------|------|
| SEC-001 | **Rate Limiting** — No protection against request flooding on authenticated endpoints. A single user or script could exhaust the 1 vCPU server. | HIGH | Per-IP + per-user sliding-window rate limiter in Go. Auth: 5/15min, API: 60/min, Messages: 30/min per user. | E-042 | Sprint 1 |
| SEC-002 | **CORS Policy** — Without CORS, any website could hijack a logged-in user's session to make API requests (cross-site request forgery). | HIGH | PocketBase `AllowedOrigins` locked to `https://{HEARTH_DOMAIN}`. No wildcard. | E-040 | Sprint 1 |
| SEC-003 | **CSP Headers** — Missing Content Security Policy allows potential XSS if input sanitization is bypassed. | MEDIUM | Full CSP header via Caddy: `script-src 'self'`, `frame-ancestors 'none'`, plus security headers (X-Frame-Options, nosniff, Referrer-Policy, Permissions-Policy). | E-041 | Sprint 1 |
| SEC-004 | **Input Sanitization** — User-generated chat content rendered in other browsers without server-side sanitization. | HIGH | `html.EscapeString()` on all user text (messages, display names, room names) before DB save. 4000-char limit. | E-043 | Sprint 1 |

### Deferred — Requires Action Before Production

| ID | Concern | Severity | Target Sprint | Rationale for Deferral | Blocked By |
|----|---------|----------|---------------|----------------------|------------|
| SEC-005 | **Auth Token Storage** — PocketBase default stores JWT in `localStorage`, which is readable by any JavaScript on the page (including XSS). `httpOnly` cookies are more secure. | MEDIUM | ~~v0.2~~ v1.0 | PB JS SDK doesn't support httpOnly cookies natively. Requires backend auth proxy hooks. Mitigated by CSP `script-src 'self'` (SEC-003). | Backend proxy hooks |
| SEC-006 | **SSE Reconnect Auth Race** — When PocketBase's SSE connection drops and auto-reconnects, the auth token may have expired, causing silent unauthenticated state. | MEDIUM | ✅ v0.2 Done | Resolved: `useReconnect` calls `authRefresh()` on every `PB_CONNECT` after initial connect. Auth failure clears store. | — |
| SEC-007 | **Secret Management in Docker** — `.env` files with HMAC secrets and API keys are readable by any process. Production should use Docker secrets or mounted files with restricted permissions. | LOW → HIGH | v1.0 (First Light) | Acceptable for development. Must be hardened before any real users. | Deployment hardening (F-010) |
| SEC-008 | **Dependency Supply Chain** — Go modules pulled from the internet could be compromised. Need vulnerability scanning and version pinning. | LOW → MEDIUM | v1.0 (First Light) | Go's `go.sum` provides cryptographic verification. `govulncheck` should be added to CI before release. | CI pipeline |
| SEC-009 | **LiveKit Room Name Guessability** — If room names in LiveKit are predictable (sequential IDs), token generation could be targeted. Room names should be UUIDs or random slugs. | HIGH | Sprint 1 (verify) | Already spec'd in E-024 — token endpoint verifies room membership. Flagged here as a reminder to verify the membership check is airtight during code review. | E-024 implementation |

### Accepted Risks (Documented)

| ID | Risk | Severity | Rationale | Mitigation |
|----|------|----------|-----------|------------|
| SEC-RISK-001 | **Screenshot prevention is impossible** in web browsers. Users can screenshot fading messages. | LOW | No reliable cross-browser API exists. Accepted as platform limitation. | Visual affordances (fading text "feels" impermanent) + community culture. Documented in Q-003. |
| SEC-RISK-002 | **No E2EE until v2.0.** Server admin can theoretically read messages and listen to voice (LiveKit sees unencrypted audio). | MEDIUM | E2EE (Insertable Streams) is complex and deferred to v2.0. For self-hosted instances, the admin IS the user. | Messages auto-delete (TTL + VACUUM). Voice is never stored. Documented in BB-003. |

---
## Production Incidents

> No production incidents — project has not shipped yet.

| Date | Severity | Description | Root Cause | Resolution | Duration |
|------|----------|-------------|-----------|------------|----------|
| — | — | — | — | — | — |

---

## Failed Approaches & Pivots

### PIVOT-001: "Single Docker Container" → Docker Compose

**Date:** February 10, 2026

**Original Plan:** Ship as a single Docker container for maximum simplicity.

**What Changed:** Realized we need 3 processes (PocketBase, LiveKit, Caddy). LiveKit benefits significantly from `network_mode: host` for UDP performance. Running 3 processes in one container requires a process supervisor (s6-overlay or supervisord), which is non-standard and harder to debug.

**Current Direction:** Docker Compose with host networking for all containers. **Confirmed by R-002, formalized by R-003 (ADR-001 accepted):** LiveKit's official deployment template uses `network_mode: "host"` for every container. All services communicate via localhost. This is mandatory for WebRTC UDP performance.

**Resolution:** PIVOT-001 is now formally resolved. ADR-001 accepted.

**Lesson Learned:** "Single container" sounds simple but multi-process containers are actually more complex to operate than Docker Compose. Host networking eliminates bridge/host conflicts but sacrifices container-level network isolation (acceptable for trusted co-located services).

---

### PIVOT-002: PocketBase API Version

**Date:** February 10, 2026

**Context:** Gemini 3 Pro provided a `main.go` scaffold using `app.OnBeforeServe()` and `app.Dao().DB()`. These are the **older** PocketBase API (pre-v0.23).

**What's Actually Current:** PocketBase v0.23+ uses `app.OnServe()` and `app.DB()` directly. The entire hook and data access API has shifted.

**Lesson Learned:** Always verify API versions against official documentation before writing code. AI models (including Gemini) can have stale training data. Flagged as R-001 in the research backlog — this MUST be resolved before any Go code is written.

---

### PIVOT-003: go.mod Version Discrepancy (Builder Review)

**Date:** February 11, 2026

**Context:** Builder agent wrote all Sprint 1 Go code using correct v0.36+ API patterns (from R-001 research) but declared `github.com/pocketbase/pocketbase v0.26.6` in go.mod — a nonexistent version that caused compilation failure. LiveKit protocol was also stale (v1.24.0 → v1.44.0).

**Root Cause:** Builder didn't have Go installed at build time, so `go mod tidy` never ran. The version number was likely hallucinated by the LLM.

**Resolution:** Researcher (Vesta) corrected go.mod to v0.36.2, ran `go mod tidy` (which also upgraded Go directive from 1.23.0 → 1.24.0), and fixed two type mismatch bugs revealed by compilation.

**Lesson Learned:** **Always run `go mod tidy` and `go build` immediately after writing Go code.** An LLM writing module dependencies without compiler verification will produce plausible-but-wrong version numbers. Builder should be instructed to verify compilability or flag it as a known issue for Researcher.

---

### Testing Note: Race Detector

**Date:** February 11, 2026

`go test -race` requires CGo (a C compiler) which is not installed on the development machine. Tests run without `-race` for now. **Action for CI:** Docker build uses Alpine with Go — `go test -race` should work in the containerized build. For local development, install `mingw-w64` or use WSL for race-detected testing.

---

## Magic Numbers & Configuration

> Configuration values that are tuned for the 1GB constraint. **Do not change without understanding the memory implications.**

### SQLite Pragmas

| Pragma | Value | Why This Value |
|--------|-------|---------------|
| `journal_mode` | WAL | Non-blocking concurrent reads/writes |
| `synchronous` | NORMAL | Fewer `fsync()`; sufficient for app crashes (not power loss) |
| `cache_size` | -2000 | ~2MB; rely on OS filesystem cache to save app RAM |
| `mmap_size` | 268435456 | 256MB mmap; reduces `read()` syscalls |
| `busy_timeout` | 5000 | 5s lock timeout; prevents immediate failures under load |

### LiveKit Tuning

| Parameter | Value | Why |
|-----------|-------|-----|
| `audio_bitrate` | 24,000 bps | Knee-of-curve for voice clarity vs. bandwidth |
| `frame_size` | 60 ms | Reduces packet header/payload ratio overhead |
| `use_dtx` | true | ~90% CPU reduction in typical group call (1 speaker, 9 silent) |
| `use_inband_fec` | true | Packet loss resilience without server-side NACK buffers |
| `use_ice_lite` | true | Reduced handshake CPU cost |
| `port_range` | 50000–60000 | 10K ports; supports ~200 users with minimal kernel overhead |

### Go Runtime

| Variable | Value | Why |
|----------|-------|-----|
| `GOMEMLIMIT` (PocketBase) | 250MiB | Triggers GC before heap grows to dangerous levels |
| `GOMEMLIMIT` (LiveKit) | 400MiB | Same — prevents OOM from GC delay |

### Message Decay

| Parameter | Value | Why |
|-----------|-------|-----|
| GC cron interval | 1 minute | Batches deletes; avoids thundering herd on expiry |
| VACUUM schedule | Nightly (4 AM) | Full DB rewrite for physical data erasure; blocking, so off-peak only |
| Default message TTL | TBD | Needs UX research — likely 1–24 hours configurable per room |

### Presence

| Parameter | Value | Why |
|-----------|-------|-----|
| Heartbeat interval | 30 seconds | Balance between freshness and server load |
| Offline threshold | 2 missed beats (60s) | Prevents "ghost" users from cluttering UI |

---

## External Service Learnings

### PocketBase
- **Version Churn:** API changed significantly at v0.23. Always verify against current docs. (PIVOT-002)
- **Single Maintainer:** Bus factor of 1. Mitigated by: open source, Go, well-structured.
- **WAL Mode:** Must be explicitly set via PRAGMAs at startup. Not the default.
- **In-Memory State:** Use Go `sync.RWMutex` maps for ephemeral data (presence). Don't persist transient state to SQLite — generates excessive I/O.

### LiveKit
- **Network Mode:** Official deployment uses `network_mode: host` for ALL containers (Caddy, LiveKit, Redis). This is mandatory for WebRTC UDP hole punching and TURN port binding. Confirmed by `livekit/deploy` templates.
- **Redis is OPTIONAL for single-node:** Omitting the `redis` config section runs LiveKit in single-node mode. Saves ~25MB RAM. Redis only needed for distributed multi-node setups. Can be added later with zero code changes.
- **ICE Lite:** Drastically reduces connection handshake CPU. Must be enabled for 1 vCPU target.
- **No Transcoding:** `video.enable_transcoding: false` is a hard requirement. Transcoding spawns FFmpeg, which will kill a 1 vCPU server.
- **DTX is Free Performance:** Discontinuous Transmission suppresses silence. In a 10-person call, usually only 1 person speaks → 90% CPU savings.
- **TURN Built-In:** LiveKit has a built-in TURN server (no separate coturn needed). Needs TLS cert files from Caddy. Listens on :5349 (TLS) and :3478 (UDP).
- **Port Range:** 50000-60000/UDP for WebRTC media. Could be reduced to 50000-50100 for Hearth's scale (~20 users).
- **Source:** R-002 research, LiveKit official docs, `livekit/livekit` config-sample.yaml.

### Caddy
- **Near-Zero Overhead:** ~15-20MB RAM. Perfect for constrained environments.
- **Auto-TLS:** Handles Let's Encrypt automatically. No manual cert management.
- **Layer 4 Module Required:** Stock Caddy cannot route TURN traffic. Must build custom binary with `caddy-l4` (Layer 4 proxy) and `caddy-yaml` (YAML config adapter) modules. The build uses `xcaddy`.
- **Config is YAML, not Caddyfile:** LiveKit's deployment templates use YAML config format. This maps directly to Caddy's internal JSON config. Standard Caddyfile examples do NOT translate 1:1 to L4 routes.
- **TLS SNI Routing:** Layer 4 reads the SNI field from TLS ClientHello to route traffic. TURN gets raw TLS passthrough; HTTP services get TLS termination + HTTP reverse proxy.
- **Certificate Sharing:** Caddy auto-manages certs, but LiveKit needs to read the cert files for its TURN server. Requires shared Docker volume with specific cert paths.
- **Source:** R-002 research, `livekit/deploy` GitHub repo (`templates/caddy.go`, `caddyl4/Dockerfile`).

### PocketBase JS SDK (Client-Side)
- **Realtime is SSE, NOT WebSocket:** The PocketBase JS SDK uses Server-Sent Events (`EventSource`) under the hood, not raw WebSocket. This is a one-way server-push channel. Client sends commands by POSTing to `/api/realtime`.
- **Auto-Reconnect:** Built-in with escalating intervals `[200, 300, 500, 1000, 1200, 1500, 2000]ms`, then repeats the last (2000ms). `maxReconnectAttempts: Infinity` by default.
- **PB_CONNECT is Critical:** The `PB_CONNECT` event fires on initial connect AND every reconnect. This is your signal to resync state (re-fetch latest records to fill the gap during disconnection).
- **Auto-Cancellation:** Duplicate in-flight requests to the same endpoint get auto-cancelled. Use `requestKey: null` to disable for fire-and-forget mutations.
- **Auth Token Refresh:** `pb.authStore.onChange()` fires on every auth state change. Use `pb.collection('users').authRefresh()` on app startup to validate/refresh tokens.
- **Custom Topics for Presence:** `pb.realtime.subscribe('topic', callback)` enables custom realtime channels — use for presence heartbeats without a DB collection.
- **Source:** R-004 research, PocketBase JS SDK v0.25+ docs, `pocketbase/js-sdk` GitHub.

### LiveKit React SDK (Client-Side)
- **Two API Surfaces:** The `@livekit/components-react` package has two coexisting APIs: (1) `LiveKitRoom` component (stable, production-ready) and (2) `SessionProvider`/`useSession` (beta, agent-focused). **Use `LiveKitRoom` for Hearth.**
- **`setWebAudioPlugins()` — Key Discovery:** `RemoteAudioTrack` has an experimental `setWebAudioPlugins(nodes: AudioNode[])` method that injects custom Web Audio processing nodes into LiveKit's internal audio pipeline: `MediaStreamSource → [plugin nodes] → GainNode → ctx.destination`. This is the cleanest integration point for spatial audio.
- **DO NOT use `RoomAudioRenderer` for Portal:** `RoomAudioRenderer` renders ALL remote audio via standard `<audio>` elements. Using it alongside Web Audio spatial processing causes double audio playback. Build a custom `PortalAudioRenderer` instead.
- **`createAudioAnalyser()` Utility:** Utility function that creates an `AnalyserNode` connected to a track — perfect for Ember glow (speaker visualization). Use `cloneTrack: true` to avoid pipeline conflicts.
- **Selective Subscription:** `autoSubscribe: false` on `LiveKitRoom`, then `publication.setSubscribed(true)` / `publication.setEnabled(false)` per track. Critical for proximity-based audio in Portal.
- **Source:** R-005 research, `livekit/components-js` GitHub, `livekit/client-sdk-js` GitHub.

### Web Audio API (Spatial Audio)
- **Linear Distance Model is Correct:** Only `distanceModel: 'linear'` reaches actual silence (gain=0) at `maxDistance`. `inverse` and `exponential` are asymptotic — they fade but never reach zero. For "out of hearing range" behavior in Portal, linear is mandatory.
- **Formula:** `gain = 1 - rolloffFactor × (distance - refDistance) / (maxDistance - refDistance)`, clamped to [0, 1].
- **2D Mapping:** Set Z=0 for all positions. Use canvas pixel coordinates directly as audio units. `PannerNode.positionX/Y` maps to user position; `AudioListener.positionX/Y` maps to local user (camera center).
- **`equalpower` is Sufficient:** `panningModel: 'equalpower'` provides basic stereo left/right separation. HRTF is more immersive but significantly more CPU-expensive and not needed for 2D.
- **Coordinate System:** AudioListener `forward = (0, 1, 0)` ("up" on canvas) and `up = (0, 0, 1)` gives correct left/right stereo for a top-down 2D view.
- **Performance:** ~0.1% CPU per participant's PannerNode. 20 participants ≈ 2% CPU. Well within budget.
- **Safari Gotcha:** Use `webkitAudioContext` fallback. Safari also requires explicit `audioContext.resume()` after user gesture (same as Chrome).
- **Source:** R-006 research, MDN Web Audio API docs, PannerNode specification.

---

## Backburner Ideas

> Ideas that are interesting but deferred — not currently on the roadmap. May be revisited in future versions.

### BB-001: Matrix Protocol Integration
**Source:** UX Research Report (Section 5.1)
**Idea:** Run Hearth on the Matrix protocol (for federation) but hide the complexity behind a friendly UI.
**Why Deferred:** Matrix adds massive complexity (Synapse/Dendrite server, DAG-based event resolution). Likely explodes the 1GB RAM target. Federation is a v3+ consideration if demand exists.
**Revisit When:** After v1.0 ships and we understand real user demand for cross-instance communication.

### BB-002: Procedural Generative Ambience
**Source:** UX Research Report (Section 3.4)
**Idea:** Use Web Audio API oscillators + noise generators for zero-download procedural ambient sounds (fire, rain, coffee shop).
**Why Deferred:** Needs dedicated audio engineering research. Pre-recorded loops are simpler for v0.2. Procedural generation could be a distinguishing feature but isn't MVP-critical.
**Revisit When:** After v0.2 ships with basic sound design.

### BB-003: Insertable Streams E2EE
**Source:** Technical Research (Section 5.3)
**Idea:** WebRTC E2EE where the browser encrypts audio/video frames before the WebRTC stack. LiveKit sees only encrypted blobs ("trust the math, not the admin").
**Why Deferred:** Complex key management (room key distribution, rotation, revocation). Browser API support still evolving. Scheduled for v2.0 but could be deferred further.
**Revisit When:** v1.0 is stable and security audit is complete.

### BB-004: Plugin Marketplace / Distribution
**Source:** Master Plan (Section 11)
**Idea:** Curated or open marketplace for Cartridges. Signing requirements. Discovery mechanism.
**Why Deferred:** Need the plugin system (v2.0) to exist first. Distribution is a post-launch concern.
**Revisit When:** After v2.0 Cartridge system is stable with real plugins in use.

### BB-005: "Picture Frame" Ambient Video
**Source:** Open Question Q-004
**Idea:** Low-res, low-fps ambient video — webcam as a portrait on the wall, not a traditional video call. Fits the "living room" metaphor.
**Why Deferred:** Video is CPU-expensive on 1 vCPU. Voice-first is the v1.0 priority.
**Revisit When:** After voice (v0.3) ships and we have real performance data.

---

## Communication Patterns

### Hearth Vocabulary (Use Consistently)

| Term | Meaning | Avoid Saying |
|------|---------|-------------|
| **Portal** | Ambient spatial voice space | "voice channel," "voice room" |
| **Campfire** | Ephemeral fading chat | "text channel," "chat room" |
| **Knock** | Guest entry request | "invite," "join request" |
| **Peephole** | Host preview of a Knock | "notification," "alert" |
| **Front Porch** | Waiting UI for guests | "waiting room," "lobby" |
| **Cartridge** | Wasm plugin | "bot," "extension," "add-on" |
| **Ember** | Active speaker glow | "green ring," "speaking indicator" |
| **Subtle Warmth** | Design system name | "theme," "color scheme" |
| **Hearth** | The product | "app," "platform" (in user-facing context) |
| **Project Vesta** | Development codename | (internal only) |

### Agent Handoff Format

When passing work between agents (Researcher → Builder, Builder → Reviewer):

```markdown
## Handoff: [From Agent] → [To Agent]
**Task ID:** [E/K/H/F/W/O-XXX]
**Context:** [What was done, key findings]
**Deliverables:** [Files created/modified]
**Blockers Resolved:** [What was blocking, how it was resolved]
**Open Items:** [Anything deferred or needing attention]
**Next Step:** [Specific action for receiving agent]
```
