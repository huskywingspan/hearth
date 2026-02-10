# Copilot Instructions — Project Vesta (Hearth)

## Project Overview
Hearth is a privacy-first, self-hosted, modular communication platform — a "Digital Living Room" alternative to Discord. Codename: **Project Vesta**. We are building a warm, intimate, high-fidelity voice/chat experience that runs on a single 1 vCPU / 1GB RAM VPS.

## Philosophy
- **"The Digital Living Room"** — not a convention center (Discord), not an email server (Matrix).
- Privacy by default. No telemetry. E2EE where feasible.
- Warmth over efficiency. Intimacy over scale. Presence over archival.
- Constraint-driven engineering: every byte and CPU cycle matters.

## Tech Stack
| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend / API | **PocketBase** (Go + SQLite) | Auth, real-time DB, chat history, cron jobs |
| Voice / Video | **LiveKit** (Go, WebRTC SFU) | Spatial audio, bandwidth mgmt, ICE Lite |
| TLS / Proxy | **Caddy** | Auto-TLS (Let's Encrypt), reverse proxy for PB + LiveKit |
| Frontend | **React + Vite + TailwindCSS** | TypeScript (strict mode) |
| Plugins | **Extism** (Wasm) | Sandboxed WASM plugins for extensibility |
| Deployment | **Docker Compose** | Target: 1 vCPU, 1GB RAM |

## Architecture Constraints
- **Memory budget is sacred.** PocketBase heap: ~250MB, LiveKit heap: ~400MB, Wasm pool: 50MB, OS+headroom: 200MB.
- SQLite in **WAL mode** with tuned pragmas (`synchronous=NORMAL`, `cache_size=-2000`, `mmap_size=268435456`).
- Go runtime must use `GOMEMLIMIT` to prevent OOM kills.
- LiveKit: `use_ice_lite=true`, `video.enable_transcoding=false`, DTX enabled, voice-first (video restricted by default).
- No Redis, no PostgreSQL, no external dependencies beyond the two Go binaries.

## Code Style & Conventions
- **TypeScript**: Strict mode. Functional React components. No class components.
- **Go**: Follow standard Go conventions. Minimize allocations in hot paths.
- **Naming**: Use the Hearth vocabulary — "Portal" (ambient voice), "Campfire" (ephemeral chat), "Knock" (guest entry), "Cartridge" (plugin).
- **Components**: Mobile-first responsive. Aggressive code-splitting via React.lazy.
- **CSS**: TailwindCSS utility classes. Custom design tokens for the "Subtle Warmth" palette.
- **Testing**: Unit tests for all business logic. Integration tests for PocketBase hooks and LiveKit signaling.

## Design System — "Subtle Warmth"
- **Palette**: Deep warm charcoals for backgrounds (`#2B211E`, `#3E2C29`), Amber/Gold for active states, Burnt Clay for alerts, Cream (`#F2E2D9`) for light mode.
- **Typography**: Inter (UI body), Merriweather (headers/story elements). Consider Recoleta or Nunito as alternatives per UX research.
- **Motion**: Ease-in/ease-out curves. Squash & stretch on interactive elements. No linear transitions. Messages "float" in, don't snap.
- **Sound**: Organic foley — wooden clicks, soft bells, fire crackles. No synthetic beeps. Optional generative ambience.
- **Shape language**: Rounded corners (high border-radius). Soft diffused shadows. "Pillow" buttons, not rectangles.

## Key UX Patterns (from research)
1. **The Portal (Spatial Voice)**: Abstract topological space, not an RPG map. "Click-to-drift" navigation. Magnetic zones. Opacity = volume. Gradient ripple visualization for speaker range.
2. **Campfires (Fading Chat)**: 4-stage transparency decay (Fresh → Fading → Echo → Gone). CSS animation-driven, not JS loops. "Mumbling" typing indicator instead of "User is typing...".
3. **The Knock (Onboarding)**: Stateless HMAC invite links. Guest sees a "Door" → Knocks → Host sees "Peephole" notification → Approves → Guest enters. "Front Porch" waiting UI with blurred activity hints. Vouched entry system.
4. **Cartridges (Plugins)**: Wasm-based via Extism. Capability-based security (allowed domains, memory caps, KV store access). Plugins run inside PocketBase process.

## Security Principles
- Stateless HMAC invite tokens — no DB storage for invites.
- Proof-of-Work bot deterrent on public endpoints (no CAPTCHAs).
- WebRTC E2EE via Insertable Streams (future).
- Capability-based Wasm sandboxing for plugins.
- Constant-time comparison for all cryptographic checks.
- Secret key rotation for instant invite revocation.

## Performance Rules
- Optimistic UI: render immediately, revert on server rejection.
- In-memory presence tracking (Go `sync.RWMutex` map), not persisted to SQLite.
- Lazy sweep GC for expired messages (cron every minute, indexed `expires_at`).
- Nightly VACUUM for physical data erasure.
- Exponential backoff for WebSocket reconnection.
- Client-side heartbeat every 30s; offline after 2 missed beats.

## File & Folder Conventions (planned)
```
hearth/
├── backend/           # PocketBase + Go hooks
│   ├── hooks/         # Go hook handlers (message GC, auth, plugins)
│   ├── plugins/       # .wasm plugin binaries
│   └── pb_data/       # PocketBase data directory (gitignored)
├── frontend/          # React + Vite app
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── hooks/        # Custom React hooks
│   │   ├── stores/       # State management
│   │   ├── styles/       # Tailwind config + design tokens
│   │   └── lib/          # Utilities, LiveKit client, API client
│   └── public/           # Static assets, sounds
├── livekit/           # LiveKit config
├── docker/            # Dockerfile + compose
├── docs/              # Design docs, research, ADRs
└── .github/           # CI, copilot instructions
```

## What NOT to Do
- Do NOT add telemetry, analytics, or tracking of any kind.
- Do NOT use heavy frameworks (Next.js, Remix) — Vite + React SPA only.
- Do NOT introduce external database dependencies (Postgres, Redis, Mongo).
- Do NOT use permanent message storage as the default. Messages fade.
- Do NOT copy Discord's "server/channel" vocabulary or left-sidebar layout.
- Do NOT use pixel art, chibi avatars, or gamified aesthetics. Keep it "high-end cozy."

## Current Phase
**Research & Exploration** — Defining architecture, collecting research, stubbing out the design document. No production code yet.

## Key Project Documents
- **`vesta_master_plan.md`** — Master design & technical specification
- **`docs/ROADMAP.md`** — Release roadmap with versioned milestones and task IDs
- **`docs/RESEARCH_BACKLOG.md`** — Active research tasks and open questions
- **`docs/research/`** — Technical and UX research reports
- **`.github/agents/`** — Specialized agent roles (Builder, Researcher, Reviewer)
