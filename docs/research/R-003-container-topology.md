# R-003: Container Topology Decision — ADR-001 Resolution

> **Status:** Complete  
> **Date:** 2026-02-10  
> **Priority:** High | **Blocks:** E-005  
> **Resolves:** ADR-001 (Container Topology), PIVOT-001  
> **Depends on:** R-002 (Caddy + LiveKit TLS Configuration)

---

## Summary

**Decision: Docker Compose with 3 containers, all using `network_mode: "host"`.**

This was pre-resolved by R-002 findings. LiveKit's official deployment architecture (`livekit/deploy` repository) uses exactly this pattern. The "single Docker container" approach from the original master plan (PIVOT-001) is formally retired.

---

## ADR-001: Container Topology — ACCEPTED

**Date:** 2026-02-10 | **Status:** Accepted

### Context

Hearth requires three long-running processes:
1. **Caddy** — TLS termination + Layer 4 SNI routing
2. **LiveKit** — WebRTC SFU for voice/video
3. **PocketBase** — Backend API, auth, real-time DB

The master plan originally said "Single Docker Container." However, LiveKit officially requires `network_mode: host` for WebRTC UDP hole-punching performance. Running three processes in one container would require a process supervisor (s6-overlay or supervisord), adding complexity.

### Options Considered

| # | Option | Pros | Cons |
|---|--------|------|------|
| 1 | **Single container + s6-overlay** | Simplest UX (`docker run` one thing) | Non-standard; harder to debug; process supervision complexity; can't independently restart services; log interleaving; s6 adds ~5MB image size |
| 2 | **Docker Compose (3 containers, host networking)** | Standard Docker practice; clean process isolation; independent restarts; standard logging; matches LiveKit official templates | Requires `docker compose up` (slightly more complex than `docker run`); no container-level network isolation |
| 3 | **Hybrid (PB+Caddy in one, LiveKit on host)** | Reduces to 2 containers | Mixed networking modes; Caddy can't reach LiveKit via bridge if LiveKit is on host; adds complexity without clear benefit |

### Decision

**Option 2: Docker Compose with 3 containers, all `network_mode: "host"`.**

### Rationale

1. **LiveKit mandates it.** The official `livekit/deploy` repository uses `network_mode: "host"` for every container (Caddy, LiveKit, and Redis when present). This is required for WebRTC UDP port binding and TURN server operation.

2. **All services communicate via localhost.** With host networking, inter-container communication is trivial — Caddy routes to `localhost:7880` (LiveKit) and `localhost:8090` (PocketBase). No Docker DNS, no bridge networking complexity.

3. **Industry-standard operations model.** Docker Compose is the expected deployment pattern. Self-hosters understand `docker compose up -d`. It provides: independent container restarts, per-service logs (`docker compose logs caddy`), resource monitoring, and straightforward updates (`docker compose pull && docker compose up -d`).

4. **Process isolation is free.** Each service gets its own PID namespace. If LiveKit crashes, Caddy and PocketBase continue running. With s6-overlay, a crash in any process requires careful supervision configuration to avoid cascading failure.

5. **Memory control.** Each container's `GOMEMLIMIT` environment variable is set independently. PocketBase: `250MiB`, LiveKit: `400MiB`. Docker's `mem_limit` provides a hard ceiling as backup.

6. **Self-hoster UX is still simple.** The entire deployment is:
   ```bash
   git clone https://github.com/hearth-app/hearth && cd hearth
   cp .env.example .env  # Edit domain name and API keys
   docker compose up -d
   ```

### Trade-offs Accepted

- **No container-level network isolation.** All services share the host's network stack. Acceptable because all three services are trusted, co-located, and run by the same operator.
- **Host port conflicts possible.** If the VPS already has something on port 443, 7880, 8090, etc., there will be conflicts. Mitigated by documentation and port-check in a setup script.
- **Docker Compose required.** Users need `docker compose` (v2), not just `docker`. This is ubiquitous on modern systems but technically one more dependency.

### Impact on Components

| Component | Impact |
|-----------|--------|
| Docker deployment | Use `docker-compose.yaml` from R-002 templates |
| Master plan | Update "Single Docker Container" reference to "Docker Compose" |
| Networking | All `localhost` — no Docker DNS or service discovery |
| Self-hosting docs | Provide one-command setup with `.env` configuration |
| CI/CD (future) | Standard `docker compose build` + push/pull pattern |

---

## Deployment Architecture (Confirmed)

```
┌──────────────────────────── Host (1 vCPU / 1GB RAM) ───────────────────────────┐
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐    │
│  │  Docker Compose   (network_mode: host for all containers)              │    │
│  │                                                                        │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────────┐ │    │
│  │  │    Caddy      │  │  PocketBase  │  │        LiveKit SFU          │ │    │
│  │  │   (custom     │  │  Go + SQLite │  │     Go + WebRTC             │ │    │
│  │  │    L4 build)  │  │              │  │                             │ │    │
│  │  │              │  │  :8090/HTTP   │  │  :7880/WS  :7881/TCP       │ │    │
│  │  │  :443/TLS    │  │              │  │  :3478/UDP  :5349/TLS      │ │    │
│  │  │  :80/HTTP    │  │  250MB heap  │  │  :50000-60000/UDP          │ │    │
│  │  │  ~20MB       │  │              │  │  400MB heap                 │ │    │
│  │  └──────────────┘  └──────────────┘  └──────────────────────────────┘ │    │
│  │                                                                        │    │
│  │  Shared volumes: caddy_data (TLS certs), pb_data (SQLite)             │    │
│  └─────────────────────────────────────────────────────────────────────────┘    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Port Summary

| Port | Service | Protocol | Purpose |
|------|---------|----------|---------|
| 80 | Caddy | HTTP | ACME challenge + HTTP→HTTPS redirect |
| 443 | Caddy | TLS | SNI routing to all services |
| 3478 | LiveKit | UDP | TURN relay (UDP) |
| 5349 | LiveKit | TLS | TURN relay (TLS) — routed via Caddy L4 |
| 7880 | LiveKit | HTTP/WS | API + WebSocket signaling |
| 7881 | LiveKit | TCP | RTC over TCP fallback |
| 8090 | PocketBase | HTTP | API + Admin UI |
| 50000-60000 | LiveKit | UDP | WebRTC media (RTP/RTCP) |

---

## Notes for Builder

1. Use the `docker-compose.yaml` template from R-002 as the starting point.
2. Add health checks: Caddy (HTTP 200 on 80), PocketBase (`/api/health`), LiveKit (TCP connect on 7880).
3. Consider a `make setup` or shell script that validates `.env`, checks port availability, and runs `docker compose up -d`.
4. LiveKit and PocketBase containers should use `restart: unless-stopped` policy.
5. Caddy's `caddy_data` volume persists TLS certificates across restarts.

---

## PIVOT-001 Resolution

**Original:** "Single Docker Container" — ship everything as one `docker run` command.  
**Final:** Docker Compose with host networking — matches LiveKit's official deployment, provides better operational characteristics, and is the industry standard for multi-service applications.  
**Status:** PIVOT-001 is now resolved. The master plan should be updated to reflect Docker Compose as the deployment model.
