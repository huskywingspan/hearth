# Hearth — Release Roadmap & Task Breakdown

> **Last Updated:** 2026-02-15
> **Phase:** v0.2.1 Settling In complete — ready for v0.3 First Friend

---

## Release Strategy

| Release | Codename | Goal | Target |
|---------|----------|------|--------|
| **v0.1** | Ember | Backend skeleton + chat MVP (no voice) | Feb 2026 ✅ |
| **v0.2** | Kindling | Frontend shell + Campfire (fading chat) | Feb 2026 ✅ |
| **v0.2.1** | Settling In | Integration fixes + access model simplification | Feb 2026 ✅ |
| **v0.3** | First Friend | Remote access, Den/Campfire schema, landing page, QR connect flow | Apr 2026 |
| **v0.4** | Hearth Fire | Voice — Dens with Table + 4 Corners spatial audio | Jun 2026 |
| **v1.0** | First Light | Full MVP: Knock + Chat features (images, reactions, replies, mentions, search, edit/delete) + Admin roles + Chat E2EE + PWA + avatars + link previews + pinned messages + deployment | Oct 2026 |
| **v1.1** | Warm Glow | Polish, accessibility, House navigation model, screen share | Dec 2026 |
| **v2.0** | Open Flame | Cartridges (plugin system) + Voice E2EE + Hearth Persona (DID) + native mobile (Capacitor) | Q1 2027 |

> **Product Principle:** "100% of what 90% of people will use." See `vesta_master_plan.md` §1.1 for the full feature completeness audit and 5-minute-to-voice onboarding trace.

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

> **Sprint Spec:** [`docs/specs/sprint-2-kindling.md`](specs/sprint-2-kindling.md) — Detailed phases, subtasks, code patterns, and acceptance criteria for Builder.

### Phase 0.2.A — Frontend Scaffolding
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-001 | Research: Vite + React + TailwindCSS project config (TS strict) | Research | ✅ Done | Standard tooling — covered in Sprint 2 spec |
| K-002 | Research: PocketBase JS SDK — real-time subscriptions, auth | Research | ✅ Done | R-004 complete. See [`R-004`](research/R-004-pocketbase-js-sdk.md) |
| K-003 | Scaffold `frontend/` (Vite, React, TailwindCSS, TypeScript strict) | Build | ✅ Done | 27 files, `tsc --noEmit` clean, 91KB gzipped |
| K-004 | Implement design tokens: Subtle Warmth palette, typography, spacing | Build | ✅ Done | `globals.css` @theme tokens, light mode, self-hosted fonts |
| K-005 | Component library: Button (pillow), Card, Input, Avatar | Build | ✅ Done | + Spinner. Rounded, warm shadows, squash & stretch |
| K-006 | Motion primitives: ease-in-out transitions, float-in animation | Build | ✅ Done | float-in, squash, warm-pulse, slide-in-right |

### Phase 0.2.B — Campfire (Fading Chat)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-010 | PocketBase SDK client setup + auth integration | Build | ✅ Done | Singleton client, AuthProvider, token auto-refresh |
| K-011 | Chat message list with real-time subscription | Build | ✅ Done | SSE subscription with create/update/delete handlers |
| K-012 | CSS transparency decay engine (4-stage: Fresh→Fading→Echo→Gone) | Build | ✅ Done | `campfire.css` — `animation-delay` with negative offset. R-008 compositor-thread. See [`R-008`](research/R-008-css-animation-performance.md) |
| K-013 | Time sync with server via `Date` header | Build | ✅ Done | RTT/2 offset estimation, 5-min resync |
| K-014 | Optimistic message sending with revert-on-error | Build | ✅ Done | Temp ID → server ID swap, revert on catch |
| K-015 | "Mumbling" typing indicator | Build | ✅ Done | Blurred undulating bars via CSS. Broadcast not yet wired (needs backend topic) |
| K-016 | Exponential backoff WebSocket reconnection | Build | ✅ Done | PB SDK built-in backoff + `useReconnect` resync |
| K-017 | Heartbeat-based presence display (30s interval) | Build | ✅ Done | `usePresence` — POST heartbeat + GET poll |

### Phase 0.2.S — Security Hardening (Frontend)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| SEC-005 | 🛡️ Migrate auth tokens from `localStorage` to `httpOnly` cookies | Security | ⏸️ Deferred | PB SDK doesn't natively support; needs backend auth proxy. localStorage used, mitigated by CSP. |
| SEC-006 | 🛡️ SSE reconnect auth validation | Security | ✅ Done | `useReconnect` — `authRefresh()` on every `PB_CONNECT` after first connect |

### Phase 0.2.C — Sound & Polish
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| K-020 | Research: Royalty-free organic foley sounds (wooden clicks, cork pop) | Research | Not Started | |
| K-021 | Sound system: interaction sounds, join/leave, ambient | Build | Not Started | Web Audio API |
| K-022 | Generative ambient engine (fire crackle, rain) — prototype | Build | Not Started | Dynamic ducking on speech |
| K-023 | Mobile-first responsive layout | Build | 🟡 Partial | Shell + responsive breakpoints. Mobile drawer not yet implemented. |
| K-024 | Code-splitting via `React.lazy` | Build | ✅ Done | LoginPage, HomePage, RoomPage lazy-loaded |

### Phase 0.2.M — Marketing Prep (Pre-Alpha Reveal)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| M-001 | Research: Competitive positioning (Revolt, Element, Mumble, Signal) | Research | Not Started | R-009. "Why not X?" FAQ |
| M-002 | Draft Reddit post for r/selfhosted (concept + screenshots) | Docs | In Progress | Post structure drafted. See [`marketing-reddit-draft.md`](specs/marketing-reddit-draft.md) |
| M-003 | Record 2-3 min screen capture (Campfire fading, UI, design system) | Docs | Not Started | GIF for Reddit, video for cross-posting |
| M-004 | README polish — screenshot, philosophy, tech stack, one-liner deploy | Docs | Not Started | GitHub repo is the "storefront" |
| M-005 | License decision (AGPL-3.0 vs. MIT vs. BSL) | Research | Not Started | r/selfhosted will ask immediately |
| M-006 | Second Reddit post plan (r/privacy, r/opensource, HN) | Docs | Not Started | After initial r/selfhosted feedback |

---

## v0.2.1 — "Settling In" (Integration Fixes + Access Model Simplification)

**Goal:** Fix all integration bugs from first LAN testing. Simplify access model: all authenticated users can see and join rooms. Denormalize author names. Result: reliable two-device chat.

> **Sprint Spec:** [`docs/specs/sprint-3-settling-in.md`](specs/sprint-3-settling-in.md) — Bug fixes, API rule changes, denormalization, and acceptance criteria for Builder.

### Phase 3.A — Backend Fixes
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| S3-001 | Relax rooms ListRule/ViewRule to any authenticated user | Build | ✅ Done | ADR-006: pre-v1.0 simplified access |
| S3-002 | Relax room_members CreateRule to allow self-join | Build | ✅ Done | Any authed user can join any room |
| S3-003 | Remove auto-join from presence heartbeat/poll endpoints | Build | ✅ Done | Returns 403 on non-member now |
| S3-004 | Denormalize `author_name` onto messages (collection + hook) | Build | ✅ Done | Server sets display_name on create |
| S3-005 | Verify sanitize.go has no unused imports | Build | ✅ Done | Already clean (PIVOT-004) |
| S3-006 | Rebuild + run all tests | Test | ✅ Done | tsc --noEmit + vite build pass |

### Phase 3.B — Frontend Fixes
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| S3-010 | Remove duplicate room_members.create() from RoomList.tsx | Build | ✅ Done | Backend hook handles owner membership |
| S3-011 | Add join-on-entry: ensureMembership() on room entry | Build | ✅ Done | CampfireRoom creates membership, catches dupes |
| S3-012 | List all rooms in sidebar (not just member rooms) | Build | ✅ Done | rooms.getFullList() — simplified access model |
| S3-013 | Update Message interface + optimistic send for author_name | Build | ✅ Done | |
| S3-014 | Update MessageBubble to read message.author_name | Build | ✅ Done | Falls back to expand then 'Wanderer' |
| S3-015 | Verify realtime dedup still works | Test | ✅ Done | Existing logic unchanged |
| S3-016 | Build frontend + deploy to pb_public | Build | ✅ Done | ~85KB gzipped JS |

### Phase 3.C — Verification
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| S3-020 | Two-device LAN integration test | Test | ✅ Done | Verified during review cycle |
| S3-021 | Git commit all Sprint 3 changes | Build | ✅ Done | v0.2.1 tag |

### Sprint 3.1 — Review Fixes

> **Review Spec:** [`docs/specs/sprint-3.1-review-fixes.md`](specs/sprint-3.1-review-fixes.md) — Reviewer-identified fixes applied by Builder.

| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| S3.1-001 | Restore messages API rules to require membership | Build | ✅ Done | Defense-in-depth per ADR-006 |
| S3.1-002 | Fix 3 stale sanitize tests (PIVOT-004 alignment) | Test | ✅ Done | Passthrough assertions, not escaping |
| S3.1-003 | Add `room_members.UpdateRule` (owner-only) | Build | ✅ Done | Schema completeness |

---

## v0.3 — "First Friend" (Remote Access + Den/Campfire Schema)

**Goal:** A friend outside the LAN can connect, see the House, and chat in a Den. Landing page at hearthapp.chat. Schema supports the new channel architecture (ADR-007).

> **ADR:** [`docs/adr/ADR-007-channel-architecture.md`](adr/ADR-007-channel-architecture.md) — Dens, Campfires, DMs, roles.

### Phase 0.3.A — Schema & Architecture (ADR-007)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| FF-001 | Add `type` field to rooms collection (`den` \| `campfire`) | Build | Not Started | Incremental migration pattern |
| FF-002 | Add `voice`, `video`, `history_visible` fields to rooms | Build | Not Started | Dens get voice option |
| FF-003 | Add `role` field to users (`homeowner` \| `keyholder` \| `member`) | Build | Not Started | First user = homeowner |
| FF-004 | Create `direct_messages` + `dm_messages` collections | Build | Not Started | Permanent DM storage |
| FF-005 | Add `public_key` field to users (empty for now, E2EE readiness) | Build | Not Started | Schema prep for v1.0 |
| FF-006 | Seed default Den ("The Den") on first startup | Build | Not Started | Auto-created if no dens exist |
| FF-007 | Update API rules: Homeowner/Keyholder can create dens; anyone can create campfires (configurable) | Build | Not Started | |
| FF-008 | Ghost Text enhancement: blur + gray shift at Echo stage in campfire.css | Build | Not Started | R-009 gap finding #4 |

### Phase 0.3.B — Remote Access (Cloudflare Tunnel)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| FF-010 | Research: Remote access architecture (CF Tunnel, VPS, WebRTC) | Research | ✅ Done | R-012 complete. CF Tunnel for dev chat, VPS for production. No UDP through CF Tunnel. |
| FF-011 | Quick-test mode: cloudflared script for chat dev/demo | Build | Not Started | `cloudflared tunnel --url http://localhost:8090` |
| FF-012 | Connect-to-server flow (enter/scan server URL in frontend) | Build | Not Started | Dynamic PocketBase endpoint, localStorage |
| FF-013 | First-impression audit: what does a new user see? | Design | Not Started | Onboarding polish |
| FF-014 | QR code generation for House URL | Build | Not Started | Homeowner generates QR → share via text/print. Zero URL typing. |
| FF-015 | QR code scan/connect flow in frontend | Build | Not Started | Camera access → decode → connect |
| FF-016 | PWA manifest + Service Worker (install prompt, offline shell) | Build | Not Started | Mobile as first-class citizen |
| FF-017 | VPS production deployment guide | Docs | Not Started | Caddy + PB + LiveKit on VPS with public IP. Complete firewall + DNS setup. |

### Phase 0.3.C — Landing Page (hearthapp.chat)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| FF-020 | Scaffold Astro landing page | Build | Not Started | Cloudflare Pages deployment |
| FF-021 | Design + content: philosophy, screenshots, "try it" CTA | Design | Not Started | |
| FF-022 | Domain setup (hearthapp.chat on Cloudflare) | Ops | Not Started | ~$6/year |

### Phase 0.3.D — Frontend Updates
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| FF-030 | Separate Den view and Campfire view in frontend | Build | Not Started | Different UI for permanent vs ephemeral |
| FF-031 | DM UI (basic: list conversations, send messages) | Build | Not Started | |
| FF-032 | Role indicators in UI (Homeowner badge, Keyholder badge) | Build | Not Started | |
| FF-033 | Fade time slider for Campfire creators | Build | Not Started | |

---

## v0.4 — "Hearth Fire" (Voice — Table + 4 Corners)

**Goal:** Spatial voice working in Dens. The Table + 4 Corners model with discrete positioning and Ember glow.

### Phase 0.4.A — LiveKit Integration
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-001 | Research: LiveKit client SDK (React) — connection, tracks, events | Research | ✅ Done | R-005 complete. See [`R-005`](research/R-005-livekit-react-sdk.md) |
| H-002 | Research: Web Audio API spatial audio (PannerNode, HRTF) | Research | ✅ Done | R-006 complete. PannerNode linear model. See [`R-006`](research/R-006-web-audio-spatial.md) |
| H-003 | LiveKit client connection + room join flow | Build | Not Started | |
| H-004 | Audio track publish/subscribe | Build | Not Started | DTX + Opus DRED config |
| H-005 | Table position: equal volume for all participants | Build | Not Started | |
| H-006 | Corner positions: semi-private spatial audio | Build | Not Started | PannerNode linear distance model |

### Phase 0.4.B — Den Voice UI
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-010 | Table + Corners layout — click to move between positions | Build | Not Started | Discrete positions, not continuous |
| H-011 | "Ember" glow for active speakers | Build | Not Started | Warm pulse via AnalyserNode |
| H-012 | "Lean In" focus cursor (boost one source, duck others) | Build | Not Started | Click-hold to beamform |
| H-013 | Voice activity detection (VAD) integration | Build | Not Started | |
| H-014 | Mute/unmute controls + push-to-talk option | Build | Not Started | |

### Phase 0.4.C — Audio Polish
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| H-020 | Soft occlusion: low-pass filter for Corner → Table audio bleed | Build | Not Started | |
| H-021 | Dynacast pause for unsubscribed video | Build | Not Started | |
| H-022 | Video cap enforcement: 480p/15fps max | Build | Not Started | Per ADR-007 |

---

## v1.0 — "First Light" (Full MVP)

**Goal:** Ship a complete, self-hostable Hearth instance with chat (Dens + Campfires + DMs), voice, onboarding (The Knock), admin roles, Chat E2EE for Campfires and DMs, and Docker deployment.

### Phase 1.0.A — The Knock (Onboarding)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-001 | "Door" landing page (guest enters name + note) | Build | Not Started | |
| F-002 | "Peephole" notification for host | Build | Not Started | Knock sound, peek without guest knowing |
| F-003 | "Front Porch" waiting UI (blurred activity hints) | Build | Not Started | |
| F-004 | "Let In" → instant transition to House | Build | Not Started | |
| F-005 | Vouched entry: "Guest of [Host]" in user list | Build | Not Started | |
| F-006 | Guest-to-account upgrade ("claim this key") | Build | Not Started | Gradual engagement |

### Phase 1.0.B — Chat E2EE (Campfires + DMs)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-020 | Research: Chat E2EE implementation for PocketBase | Research | Not Started | R-011 |
| F-021 | `public_key` enrollment flow for users | Build | Not Started | |
| F-022 | Client-side encryption/decryption for Campfire messages | Build | Not Started | |
| F-023 | Client-side encryption/decryption for DMs | Build | Not Started | |
| F-024 | Key exchange mechanism (Signal Protocol or simpler?) | Build | Not Started | |

### Phase 1.0.C — Admin Roles
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-030 | Homeowner/Keyholder/Member role enforcement in API rules | Build | Not Started | ADR-007 |
| F-031 | Admin UI for role assignment | Build | Not Started | |
| F-032 | Per-Den history visibility configuration | Build | Not Started | |
| F-033 | Server-wide Campfire settings (fade time bounds, creation permissions) | Build | Not Started | |

### Phase 1.0.D — Deployment
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| F-010 | Multi-stage Docker build (Alpine, static Go binary) | Build | Not Started | |
| F-011 | Docker Compose with Caddy + PocketBase + LiveKit | Build | Not Started | |
| F-012 | Self-hosting documentation | Docs | Not Started | |
| F-013 | Performance profiling (1 vCPU, 1GB, ~20 users) | Test | Not Started | |
| F-014 | Smoke test suite for full flow | Test | Not Started | |

### Phase 1.0.E — Feature Completeness ("100% of 90%")

> These features close the gap between Hearth and what users expect from any modern chat platform. See `vesta_master_plan.md` §1.1 for the full audit table and rationale.

| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| FC-001 | Image / file sharing (upload + display + download) | Build | Not Started | PocketBase file field. Max size configurable by Homeowner. |
| FC-002 | Emoji reactions on messages | Build | Not Started | Unicode emoji picker. Reaction counts on messages. |
| FC-003 | Reply-to messages (with scroll-to-parent) | Build | Not Started | `reply_to` relation on messages. Visual thread line. |
| FC-004 | @mentions with notification highlight | Build | Not Started | `@name` autocomplete, highlight in message list, notify target. |
| FC-005 | Message search (Dens only) | Build | Not Started | SQLite FTS5 — zero external deps. Campfires excluded (ephemeral). |
| FC-006 | Push notifications (PWA Web Push API) | Build | Not Started | Service Worker + VAPID keys. Mention/DM/Knock triggers. |
| FC-007 | Edit / delete own messages | Build | Not Started | Edit shows "(edited)" indicator. Delete shows "message removed." |
| FC-008 | User avatars (upload or generated) | Build | Not Started | PocketBase file field on users. Fallback: Dicebear or initials. |
| FC-009 | Link previews (OpenGraph) | Build | Not Started | Server-side fetch via PB hook. Privacy: proxy through PB, don't leak user IPs. |
| FC-010 | Pinned messages (per-Den) | Build | Not Started | Boolean `pinned` field. Pinned message drawer in Den UI. |

### Phase 1.0.S — Security Hardening (Pre-Production)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| SEC-007 | 🛡️ Docker secret management | Security | Not Started | Replace `.env` with Docker secrets or mounted files with `chmod 600`. No secrets in image layers. |
| SEC-008 | 🛡️ Dependency vulnerability scanning | Security | Not Started | Add `govulncheck` to CI. Pin exact versions in `go.mod`. Audit dependency tree. |
| SEC-009 | 🛡️ LiveKit room isolation audit | Security | Not Started | Verify room membership check in token endpoint is airtight. Confirm room names use UUIDs. |

---

## v1.1 — "Warm Glow" (Polish & Admin)

| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| W-001 | **House navigation model** — replace sidebar list with spatial/visual House metaphor | Build | Not Started | #1 differentiator from R-009 gap analysis |
| W-002 | Sliding pane transitions between Dens (View Transitions API / Framer Motion) | Build | Not Started | Spatial continuity |
| W-003 | Radical Quiet: auto-hide chrome during conversation | Build | Not Started | UX Research §5.3 |
| W-004 | Accessibility audit (screen readers, fading text, spatial audio) | Research | Not Started | Critical open question |
| W-005 | Light mode ("Cream" palette) | Build | Not Started | |
| W-006 | Keyboard navigation + ARIA for all components | Build | Not Started | |
| W-007 | Admin guide documentation | Docs | Not Started | |
| W-008 | Sound design integration (thock, cork pop, foley) | Build | Not Started | R-007 |
| W-009 | Screen share (WebRTC `getDisplayMedia`) | Build | Not Started | CPU-heavy — post-MVP. Low-res, Den-only. |
| W-010 | Group DMs (if demand exists post-v1.0) | Build | Not Started | v1.0 is 1:1 only; Dens serve group use case |

---

## v2.0 — "Open Flame" (Plugins + Voice E2EE + Persona)

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
| O-008 | QuickJS→Wasm support for JavaScript Cartridges | Build | Not Started | R-009 gap finding #7 |

### Voice E2EE
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| O-010 | Research: Insertable Streams API — browser support, perf impact | Research | Not Started | |
| O-011 | WebRTC E2EE via Insertable Streams | Build | Not Started | |
| O-012 | Key exchange mechanism for voice rooms | Build | Not Started | |
| O-013 | Security audit (HMAC, PoW, Wasm sandbox, all E2EE) | Test | Not Started | |

### Hearth Persona (Cross-Server Identity)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| O-020 | Research: DID-based portable identity | Research | Not Started | R-010 |
| O-021 | Persona creation + export/import flow | Build | Not Started | |
| O-022 | Cross-House identity resolution | Build | Not Started | |

### Native Mobile (Capacitor Wrapper)
| ID | Task | Type | Status | Notes |
|----|------|------|--------|-------|
| O-030 | Capacitor project setup (wrapping PWA) | Build | Not Started | Native push, app store listing |
| O-031 | iOS + Android builds and testing | Build | Not Started | |
| O-032 | App store submission (if warranted by demand) | Ops | Not Started | |

---

## Milestone Summary

```
Feb 2026  █████░░░░░░░░░░░░░░░  v0.2.1 — Settling In (Integration Fixes) ✅
Apr 2026  ███████░░░░░░░░░░░░░  v0.3 — First Friend (Remote + Schema + QR Connect) ← NEXT
Jun 2026  ██████████░░░░░░░░░░  v0.4 — Hearth Fire (Voice + Table/Corners)
Oct 2026  ██████████████░░░░░░  v1.0 — First Light (MVP + Feature Complete + PWA)
Dec 2026  ████████████████░░░░  v1.1 — Warm Glow (Polish + House Nav + Screen Share)
Q1  2027  ████████████████████  v2.0 — Open Flame (Plugins + Voice E2EE + Persona + Native Mobile)
```
