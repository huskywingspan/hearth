# Hearth — Release Roadmap & Task Breakdown

> **Last Updated:** 2026-02-11
> **Phase:** Research & Exploration → Sprint 1 ready

---

## Release Strategy

| Release | Codename | Goal | Target |
|---------|----------|------|--------|
| **v0.1** | Ember | Backend skeleton + chat MVP (no voice) | Apr 2026 |
| **v0.2** | Kindling | Frontend shell + Campfire (fading chat) | Jun 2026 |
| **v0.3** | Hearth Fire | Voice — The Portal (spatial audio) | Aug 2026 |
| **v1.0** | First Light | Full MVP: chat + voice + Knock + deployment | Oct 2026 |
| **v1.1** | Warm Glow | Polish, accessibility, admin tools | Dec 2026 |
| **v2.0** | Open Flame | Cartridges (plugin system) + E2EE | Q1 2027 |

---

## v0.1 — "Ember" (Backend Skeleton + Chat MVP)

**Goal:** A running PocketBase instance with optimized SQLite, basic auth, message CRUD with expiry, and HMAC invite tokens. No frontend — API-only. Proves the data layer works within the 1GB budget.

> **Sprint Spec:** [`docs/specs/sprint-1-ember.md`](specs/sprint-1-ember.md) — Detailed phases, subtasks, code patterns, and acceptance criteria for Builder.

### Phase 0.1.A — Project Scaffolding
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| E-001 | Research: Verify PocketBase v0.23+ API (hooks, DB access, cron) | Research | ✅ Done | R-001 complete. See [`R-001`](research/R-001-pocketbase-api-verification.md) |
| E-002 | Research: Caddy reverse proxy config for PocketBase + LiveKit TLS | Research | ✅ Done | R-002 complete. See [`R-002`](research/R-002-caddy-livekit-tls-config.md) |
| E-003 | ADR-001: Container topology (single vs. multi-container Docker Compose) | Research | ✅ Done | R-003 complete. ADR-001 accepted. See [`R-003`](research/R-003-container-topology.md) |
| E-004 | Scaffold Go backend project (`backend/main.go`, `go.mod`) | Build | Not Started | |
| E-005 | Create `docker-compose.yml` with memory constraints | Build | Not Started | `GOMEMLIMIT` enforced per service |
| E-006 | Create Caddyfile for reverse proxy + auto-TLS | Build | Not Started | Routes: API (`:8090`) + LiveKit WS (`:7880`) |
| E-007 | Create LiveKit `config.yaml` with optimized settings | Build | Not Started | ICE Lite, DTX, no transcoding, port range |
| E-008 | `.gitignore` updates, CI placeholder | Build | Done | |

### Phase 0.1.B — Data Layer
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| E-010 | Implement SQLite WAL pragma injection on startup | Build | Not Started | Verify correct PocketBase hook lifecycle |
| E-011 | Define PocketBase collections: Users, Rooms, Messages | Build | Not Started | Schema per technical research spec |
| E-012 | Implement `expires_at` index on Messages collection | Build | Not Started | Required for GC performance |
| E-013 | Implement cron-based message GC (lazy sweep, every 1 min) | Build | Not Started | Bulk delete via indexed `expires_at` |
| E-014 | Implement nightly VACUUM cron job | Build | Not Started | Physical data erasure for privacy |
| E-015 | In-memory presence tracking (Go `sync.RWMutex` map) | Build | Not Started | Ephemeral — survives no restarts |
| E-016 | Unit tests: message expiry, GC correctness, presence map | Test | Not Started | |

### Phase 0.1.C — Auth & Security
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| E-020 | Basic auth flow (email/password via PocketBase) | Build | Not Started | |
| E-021 | HMAC invite token generation endpoint | Build | Not Started | Stateless, no DB writes |
| E-022 | HMAC invite validation with constant-time compare | Build | Not Started | Two-key rotation (current + old) |
| E-023 | Proof-of-Work challenge endpoint (Client Puzzle Protocol) | Build | Not Started | SHA256 partial collision |
| E-024 | LiveKit JWT token generation for room access | Build | Not Started | `canPublishVideo: false` by default |
| E-025 | Integration tests: invite flow, PoW, auth | Test | Not Started | |

### Phase 0.1.D — Observability
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| E-030 | `/metrics` endpoint (Prometheus format) | Build | Not Started | Heap, goroutines, SQLite, rooms |
| E-031 | Structured logging setup | Build | Not Started | |

---

## v0.2 — "Kindling" (Frontend Shell + Campfire Chat)

**Goal:** A working React frontend that connects to the PocketBase backend. Campfire (fading chat) is the first user-facing feature. Design system implemented.

### Phase 0.2.A — Frontend Scaffolding
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-001 | Research: Vite + React + TailwindCSS project config (TS strict) | Research | Not Started | |
| K-002 | Research: PocketBase JS SDK — real-time subscriptions, auth | Research | ✅ Done | R-004 complete. See [`R-004`](research/R-004-pocketbase-js-sdk.md) |
| K-003 | Scaffold `frontend/` (Vite, React, TailwindCSS, TypeScript strict) | Build | Not Started | |
| K-004 | Implement design tokens: Subtle Warmth palette, typography, spacing | Build | Not Started | Tailwind config + CSS custom properties |
| K-005 | Component library: Button (pillow), Card, Input, Avatar | Build | Not Started | Rounded, soft shadows, squash & stretch |
| K-006 | Motion primitives: ease-in-out transitions, float-in animation | Build | Not Started | No linear transitions |

### Phase 0.2.B — Campfire (Fading Chat)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-010 | PocketBase SDK client setup + auth integration | Build | Not Started | |
| K-011 | Chat message list with real-time subscription | Build | Not Started | WebSocket via PB real-time API |
| K-012 | CSS transparency decay engine (4-stage: Fresh→Fading→Echo→Gone) | Build | Not Started | `animation-delay` with negative offset |
| K-013 | Time sync with server via `Date` header | Build | Not Started | |
| K-014 | Optimistic message sending with revert-on-error | Build | Not Started | |
| K-015 | "Mumbling" typing indicator | Build | Not Started | Blurred waveform, not "User is typing..." |
| K-016 | Exponential backoff WebSocket reconnection | Build | Not Started | |
| K-017 | Heartbeat-based presence display (30s interval) | Build | Not Started | |

### Phase 0.2.C — Sound & Polish
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-020 | Research: Royalty-free organic foley sounds (wooden clicks, cork pop) | Research | Not Started | |
| K-021 | Sound system: interaction sounds, join/leave, ambient | Build | Not Started | Web Audio API |
| K-022 | Generative ambient engine (fire crackle, rain) — prototype | Build | Not Started | Dynamic ducking on speech |
| K-023 | Mobile-first responsive layout | Build | Not Started | |
| K-024 | Code-splitting via `React.lazy` | Build | Not Started | |

---

## v0.3 — "Hearth Fire" (Voice — The Portal)

**Goal:** Spatial voice working in the browser. The abstract topological space UI with proximity-based volume attenuation.

### Phase 0.3.A — LiveKit Integration
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-001 | Research: LiveKit client SDK (React) — connection, tracks, events | Research | ✅ Done | R-005 complete. See [`R-005`](research/R-005-livekit-react-sdk.md) |
| H-002 | Research: Web Audio API spatial audio (PannerNode, HRTF) | Research | ✅ Done | R-006 complete. PannerNode linear model. See [`R-006`](research/R-006-web-audio-spatial.md) |
| H-003 | LiveKit client connection + room join flow | Build | Not Started | |
| H-004 | Audio track publish/subscribe | Build | Not Started | DTX + Opus DRED config |
| H-005 | Proximity-based volume attenuation (distance → gain) | Build | Not Started | |

### Phase 0.3.B — Portal UI
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-010 | Abstract topological space canvas | Build | Not Started | Not an RPG map |
| H-011 | Click-to-drift navigation with easing | Build | Not Started | No WASD |
| H-012 | Magnetic zones (auto-snap to conversation circles) | Build | Not Started | |
| H-013 | Gradient ripple visualization (opacity = volume) | Build | Not Started | |
| H-014 | "Ember" glow for active speakers | Build | Not Started | Warm pulse, not green ring |
| H-015 | "Lean In" focus cursor (beamforming UX) | Build | Not Started | Click-hold to boost one source |

### Phase 0.3.C — Audio Polish
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-020 | Soft occlusion: low-pass filter behind barriers | Build | Not Started | |
| H-021 | Dynacast pause for unsubscribed video | Build | Not Started | |
| H-022 | Voice activity detection (VAD) integration | Build | Not Started | |

---

## v1.0 — "First Light" (Full MVP)

**Goal:** Ship a complete, self-hostable Hearth instance with chat, voice, onboarding, and Docker deployment.

### Phase 1.0.A — The Knock (Onboarding)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-001 | "Door" landing page (guest enters name + note) | Build | Not Started | |
| F-002 | "Peephole" notification for host | Build | Not Started | Knock sound, peek without guest knowing |
| F-003 | "Front Porch" waiting UI (blurred activity hints) | Build | Not Started | |
| F-004 | "Let In" → instant transition to room | Build | Not Started | |
| F-005 | Vouched entry: "Guest of [Host]" in user list | Build | Not Started | |
| F-006 | Guest-to-account upgrade ("claim this key") | Build | Not Started | Gradual engagement |

### Phase 1.0.B — Deployment
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-010 | Multi-stage Docker build (Alpine, static Go binary) | Build | Not Started | |
| F-011 | Docker Compose with Caddy + PocketBase + LiveKit | Build | Not Started | |
| F-012 | Self-hosting documentation | Docs | Not Started | |
| F-013 | Performance profiling (1 vCPU, 1GB, ~20 users) | Test | Not Started | |
| F-014 | Smoke test suite for full flow | Test | Not Started | |

---

## v1.1 — "Warm Glow" (Polish & Admin)

| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| W-001 | Admin dashboard (room, user, plugin management) | Build | Not Started | |
| W-002 | Accessibility audit (screen readers, fading text, spatial audio) | Research | Not Started | Critical open question |
| W-003 | Light mode ("Cream" palette) | Build | Not Started | |
| W-004 | Keyboard navigation + ARIA for all components | Build | Not Started | |
| W-005 | systemd bare-metal deployment alternative | Build | Not Started | |
| W-006 | Admin guide documentation | Docs | Not Started | |

---

## v2.0 — "Open Flame" (Plugins + E2EE)

### Cartridges (Plugin System)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| O-001 | Research: Extism Go SDK — current API, lifecycle, memory model | Research | Not Started | |
| O-002 | Extism integration in PocketBase hooks | Build | Not Started | |
| O-003 | Host function manifest (log, KV, kick, fetch) | Build | Not Started | |
| O-004 | `plugins.json` capability config | Build | Not Started | |
| O-005 | Memory-capped instance pool (50MB) | Build | Not Started | |
| O-006 | Example plugins: moderation filter, `/roll` | Build | Not Started | |
| O-007 | Plugin developer docs + PDK examples | Docs | Not Started | |

### E2EE
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| O-010 | Research: Insertable Streams API — browser support, perf impact | Research | Not Started | |
| O-011 | WebRTC E2EE via Insertable Streams | Build | Not Started | |
| O-012 | Key exchange mechanism (public key + room key distribution) | Build | Not Started | |
| O-013 | Security audit (HMAC, PoW, Wasm sandbox, E2EE) | Test | Not Started | |

---

## Milestone Summary

```
Feb 2026  ████░░░░░░░░░░░░░░░░  v0.0 — Research & Foundation (NOW)
Apr 2026  ████████░░░░░░░░░░░░  v0.1 — Ember (Backend API)
Jun 2026  ████████████░░░░░░░░  v0.2 — Kindling (Frontend + Chat)
Aug 2026  ████████████████░░░░  v0.3 — Hearth Fire (Voice)
Oct 2026  ██████████████████░░  v1.0 — First Light (MVP Ship)
Dec 2026  ███████████████████░  v1.1 — Warm Glow (Polish)
Q1  2027  ████████████████████  v2.0 — Open Flame (Plugins + E2EE)
```
