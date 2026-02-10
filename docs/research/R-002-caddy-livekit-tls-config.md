# R-002: Caddy + LiveKit TLS Configuration

> **Status:** Complete | **Date:** 2026-02-10 | **Priority:** Critical
> **Blocks:** E-001 (Docker Compose), E-002 (Caddy Config), E-003 (LiveKit Config)
> **Source:** LiveKit official docs, livekit/deploy GitHub repo, Caddy docs, PocketBase production guide

---

## Summary

LiveKit's official deployment uses a **custom Caddy build** with Layer 4 TLS SNI routing — NOT a standard Caddyfile. This is a major discovery that changes our deployment architecture. Hearth needs a single Caddy instance that routes encrypted traffic to three backends (PocketBase, LiveKit, TURN) based on the domain name in the TLS handshake.

### Critical Findings

| Discovery | Impact |
|-----------|--------|
| LiveKit's Caddy uses Layer 4 module (`caddy-l4`) | Must build custom Caddy binary — stock Caddy won't work |
| Config is YAML, not Caddyfile | Different config syntax, but more flexible for L4 routing |
| All containers use `network_mode: "host"` | No Docker bridge networking; all services on localhost |
| **Redis is OPTIONAL for single-node** | Eliminates ~25MB RAM overhead — huge win for our 1GB budget |
| TURN server is built into LiveKit | No separate TURN service needed |

---

## Architecture Overview

```
Internet (Port 443)
        │
    ┌───▼───┐
    │ Caddy  │  Custom build: livekit/caddyl4
    │ (L4)   │  TLS SNI routing on :443
    └───┬────┘
        │
        ├─── SNI: turn.hearth.example  ──► LiveKit TURN (localhost:5349)
        ├─── SNI: lk.hearth.example    ──► LiveKit API  (localhost:7880)
        └─── SNI: hearth.example       ──► PocketBase   (localhost:8090)

Also open:
  - Port 80   → Caddy (ACME HTTP-01 challenge for cert issuance)
  - Port 7881 → LiveKit TCP fallback (direct, not through Caddy)
  - Ports 50000-60000/UDP → LiveKit WebRTC media (direct)
  - Port 3478/UDP → LiveKit TURN/UDP (direct)
```

### Why Layer 4?

TURN traffic is NOT HTTP — it's raw TLS/DTLS. A standard HTTP reverse proxy (Caddyfile `reverse_proxy`) cannot route TURN. The Layer 4 module intercepts connections at the TLS level, reads the SNI (Server Name Indication) field from the ClientHello, and routes the raw TCP stream to the correct backend. This allows:

- TURN (non-HTTP TLS) on the same :443 as HTTPS
- A single TLS certificate for all subdomains (or separate certs)
- Caddy manages all TLS termination for HTTP services, while passing through TLS for TURN

---

## Custom Caddy Build

### Required Modules

```
github.com/abiosoft/caddy-yaml   — YAML config adapter
github.com/mholt/caddy-l4        — Layer 4 proxy/routing
```

### Dockerfile

```dockerfile
FROM caddy:2-builder AS builder

RUN xcaddy build \
    --with github.com/abiosoft/caddy-yaml \
    --with github.com/mholt/caddy-l4

FROM caddy:2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
```

**Memory Impact:** Negligible. These are compile-time additions, not runtime services. The custom Caddy binary is ~50MB on disk but still uses ~15-20MB RAM at runtime — same as stock Caddy.

---

## Complete Caddy Configuration (caddy.yaml)

This is the full YAML config for Hearth's Caddy instance. It handles all three services.

```yaml
# caddy.yaml — Hearth reverse proxy configuration
# Loaded with: caddy run --config /etc/caddy/caddy.yaml --adapter yaml

logging:
  logs:
    default:
      level: WARN

# Storage for TLS certificates
storage:
  module: file_system
  root: /data/caddy

apps:
  # === Layer 4: TLS SNI Routing ===
  layer4:
    servers:
      main:
        listen:
          - ":443"
        routes:
          # Route 1: TURN traffic → LiveKit's built-in TURN server
          - match:
              - tls:
                  sni:
                    - "turn.hearth.example"
            handle:
              - handler: proxy
                upstreams:
                  - dial:
                      - "localhost:5349"

          # Route 2: LiveKit signaling (WebSocket API)
          - match:
              - tls:
                  sni:
                    - "lk.hearth.example"
            handle:
              - handler: tls
              - handler: proxy
                upstreams:
                  - dial:
                      - "localhost:7880"

          # Route 3: PocketBase API + SPA
          - match:
              - tls:
                  sni:
                    - "hearth.example"
            handle:
              - handler: tls
              - handler: proxy
                upstreams:
                  - dial:
                      - "localhost:8090"

  # === HTTP: ACME Challenge + HTTPS Redirect ===
  http:
    http_port: 80
    servers:
      http_redirect:
        listen:
          - ":80"
        routes:
          - match:
              - path:
                  - "/.well-known/acme-challenge/*"
            # ACME challenges handled automatically by Caddy
          - handle:
              - handler: static_response
                status_code: 301
                headers:
                  Location:
                    - "https://{http.request.host}{http.request.uri}"

  # === TLS: Certificate Management ===
  tls:
    automation:
      policies:
        - subjects:
            - "hearth.example"
            - "lk.hearth.example"
            - "turn.hearth.example"
          issuers:
            - module: acme
              # email: admin@hearth.example  # Set in production
```

### Configuration Notes

1. **TURN route does NOT terminate TLS** — it passes the raw TLS stream to LiveKit. LiveKit's TURN server handles its own TLS with the Caddy-managed certificate.
2. **LiveKit and PocketBase routes DO terminate TLS** — Caddy decrypts, then proxies plaintext HTTP to localhost backends.
3. **Certificates are auto-managed** by Caddy (Let's Encrypt) for all three domains.
4. **TURN needs access to the TLS certificate** — LiveKit config must reference the Caddy-managed cert files.

### Certificate Sharing with LiveKit TURN

LiveKit's TURN server needs the TLS certificate. Since Caddy manages certs, LiveKit must read from Caddy's certificate storage:

```yaml
# In livekit.yaml (TURN section):
turn:
  enabled: true
  domain: turn.hearth.example
  tls_port: 5349
  udp_port: 3478
  cert_file: /data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/turn.hearth.example/turn.hearth.example.crt
  key_file: /data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/turn.hearth.example/turn.hearth.example.key
```

**Alternative:** Use Caddy's `on_demand` TLS or a shared cert directory. The exact path depends on the ACME issuer. Check with `caddy trust` or inspect `/data/caddy/certificates/`.

---

## LiveKit Configuration (livekit.yaml)

Single-node, no Redis, optimized for 1 vCPU / 400MB:

```yaml
# livekit.yaml — Hearth voice/video server

port: 7880
# bind_addresses:
#   - "127.0.0.1"  # Caddy handles external access

# === RTC (WebRTC) ===
rtc:
  port_range_start: 50000
  port_range_end: 60000
  use_ice_lite: true
  # tcp_port: 7881  # Optional: TCP fallback for restrictive firewalls
  # use_external_ip: true  # Set if server has multiple interfaces

# === Keys ===
keys:
  hearth-api-key: "CHANGE_ME_IN_PRODUCTION"

# === Room Defaults ===
room:
  auto_create: false     # Rooms created explicitly by PocketBase
  empty_timeout: 300     # 5 min — close room if empty
  max_participants: 20   # Hearth is intimate, not massive

# === Audio (Hearth is voice-first) ===
audio:
  active_sensitivity: 30
  min_percentile: 15

# === Video (Restricted by Default) ===
# No transcoding — saves massive CPU
# Video only enabled when explicitly requested

# === TURN ===
turn:
  enabled: true
  domain: turn.hearth.example
  tls_port: 5349
  udp_port: 3478
  cert_file: /data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/turn.hearth.example/turn.hearth.example.crt
  key_file: /data/caddy/certificates/acme-v02.api.letsencrypt.org-directory/turn.hearth.example/turn.hearth.example.key

# === Logging ===
logging:
  level: info
  pion_level: warn

# === Limits (1GB constraint) ===
limit:
  num_tracks: 50          # Max simultaneous audio tracks
  bytes_per_sec: 500000   # 500 KB/s per participant
```

### Redis: Deliberately Omitted

From LiveKit docs: *"When redis is set, LiveKit will automatically operate in a fully distributed fashion."* For Hearth's single-node deployment, omitting Redis means:

- LiveKit runs in **single-node mode** (all participants on one server)
- Saves **~25MB RAM** (Redis baseline)
- One less container to manage
- No data persistence concern (voice state is ephemeral anyway)

If Hearth ever needs multiple servers, Redis can be added later with zero code changes.

---

## Docker Compose (docker-compose.yaml)

```yaml
# docker-compose.yaml — Hearth production deployment
# Target: 1 vCPU, 1 GB RAM

services:
  caddy:
    build:
      context: ./docker/caddy
      dockerfile: Dockerfile
    network_mode: "host"
    volumes:
      - caddy_data:/data/caddy
      - ./config/caddy.yaml:/etc/caddy/caddy.yaml:ro
    command: caddy run --config /etc/caddy/caddy.yaml --adapter yaml
    restart: unless-stopped

  livekit:
    image: livekit/livekit-server:latest
    network_mode: "host"
    volumes:
      - caddy_data:/data/caddy:ro   # Read TURN certificates
      - ./config/livekit.yaml:/etc/livekit.yaml:ro
    command: --config /etc/livekit.yaml
    environment:
      - GOMEMLIMIT=400MiB
    restart: unless-stopped
    depends_on:
      - caddy

  pocketbase:
    build:
      context: ./backend
      dockerfile: Dockerfile
    network_mode: "host"
    volumes:
      - pb_data:/pb/pb_data
    environment:
      - GOMEMLIMIT=250MiB
    restart: unless-stopped
    depends_on:
      - caddy

volumes:
  caddy_data:
    driver: local
  pb_data:
    driver: local
```

### Why `network_mode: "host"`?

1. **WebRTC requires it** — LiveKit needs direct UDP access on ports 50000-60000. Docker bridge NAT adds latency and breaks ICE.
2. **TURN requires it** — TURN server binds multiple ports that must be directly accessible.
3. **Simplicity** — All services communicate via localhost. No Docker DNS, no port mapping.
4. **Performance** — Zero overhead from Docker's iptables NAT rules.

**Trade-off:** No network isolation between containers. Acceptable for Hearth since all services are trusted and part of the same deployment.

---

## Firewall / Port Requirements

| Port | Protocol | Service | Direction | Notes |
|------|----------|---------|-----------|-------|
| 443 | TCP | Caddy (L4) | Inbound | TLS SNI → TURN, LiveKit, PocketBase |
| 80 | TCP | Caddy (HTTP) | Inbound | ACME challenges + HTTPS redirect |
| 7881 | TCP | LiveKit | Inbound | WebRTC TCP fallback (optional) |
| 3478 | UDP | LiveKit TURN | Inbound | TURN/UDP relay |
| 50000-60000 | UDP | LiveKit RTC | Inbound | WebRTC media streams |

**Outbound:** HTTP/HTTPS for Let's Encrypt ACME (port 80/443), STUN (if configured), package updates.

### UFW Example

```bash
sudo ufw allow 80/tcp    # ACME
sudo ufw allow 443/tcp   # All HTTPS/WSS/TURN-TLS
sudo ufw allow 7881/tcp  # LiveKit TCP fallback
sudo ufw allow 3478/udp  # TURN/UDP
sudo ufw allow 50000:60000/udp  # WebRTC
```

---

## DNS Configuration

| Record | Type | Value |
|--------|------|-------|
| `hearth.example` | A | `<server-ip>` |
| `lk.hearth.example` | A | `<server-ip>` |
| `turn.hearth.example` | A | `<server-ip>` |

All three A records point to the same IP. Caddy's Layer 4 SNI routing separates traffic.

**Alternative:** Use a wildcard record (`*.hearth.example → <server-ip>`) for simplicity, but this issues a wildcard cert which requires DNS-01 ACME challenge (more complex).

---

## Answers to Open Questions

### Q-008: Docker Bridge vs. Host Networking

**Answer: Host networking.** livekiT's official Docker Compose template uses `network_mode: "host"` for ALL containers. This is required for:
- WebRTC UDP hole punching
- TURN UDP port binding
- Low-latency media transport

All services (Caddy, LiveKit, PocketBase) should use host networking and communicate via localhost.

### Q-003: Single Caddy vs. Separate Ingress

**Answer: Single Caddy with Layer 4.** One Caddy instance handles all routing via TLS SNI. This means:
- Single point for TLS certificate management
- One container instead of multiple proxies
- All traffic enters through port 443

### Partial Q-006: Subdomain Strategy

**Recommendation:** Three subdomains:
- `hearth.example` — PocketBase (API + SPA)
- `lk.hearth.example` — LiveKit signaling
- `turn.hearth.example` — TURN relay

This is the cleanest approach for TLS SNI routing and matches LiveKit's official templates.

---

## Gotchas for Builder

1. **Stock Caddy won't work** — Must build custom binary with `caddy-l4` and `caddy-yaml` modules.
2. **TURN cert path is fragile** — Caddy stores certs in a specific directory structure. If the ACME issuer changes (staging vs. production), the path changes. Consider a cert sync script or symlink.
3. **Host networking means port conflicts are possible** — Don't run anything else on ports 443, 80, 7880, 7881, 3478, or 5349.
4. **LiveKit's `use_ice_lite: true`** eliminates the need for STUN — but TURN is still needed for clients behind symmetric NAT.
5. **PocketBase's `--http` flag** should be `0.0.0.0:8090` in the Dockerfile, but with host networking it's accessible from all interfaces. Consider binding to `127.0.0.1:8090` since Caddy handles external access.
6. **Certificate renewal** — Caddy auto-renews, but LiveKit reads cert files on startup. LiveKit may need a signal (SIGHUP or restart) to pick up renewed certs. Investigate LiveKit's cert hot-reload capability.
7. **YAML config quirks** — The Caddy YAML adapter maps directly to Caddy's internal JSON config. Official Caddyfile examples don't translate 1:1.
8. **`network_mode: "host"` disables Docker's built-in DNS** — Services can't reference each other by container name. Use `localhost` only.
9. **UDP port range** — 50000-60000 is 10,000 ports. For Hearth's small scale (20 participants max), this could be reduced to 50000-50100 to minimize firewall rules.
10. **No healthchecks in compose** — Add `healthcheck` for each service. PocketBase: `curl http://localhost:8090/api/health`. LiveKit: TCP check on 7880. Caddy: TCP check on 443.

---

## Ready for Implementation

**Feature:** Docker Deployment Stack | **Spec:** This document | **Complexity:** L
**Key Points for Builder:**
1. Build custom Caddy with `caddy-l4` + `caddy-yaml` (Dockerfile provided)
2. All containers use `network_mode: "host"` — no exceptions
3. NO Redis for single-node deployment — saves 25MB RAM
4. TURN cert path must reference Caddy's cert storage (volume sharing)

**Files to Create/Modify:**
- `docker/caddy/Dockerfile` — Custom Caddy build
- `config/caddy.yaml` — Layer 4 TLS SNI routing config
- `config/livekit.yaml` — Single-node LiveKit config (no Redis)
- `docker-compose.yaml` — Full stack with host networking
- `backend/Dockerfile` — PocketBase container

**Questions Resolved:**
- Q-003: Single Caddy with L4 routing → ✅ Confirmed
- Q-006: Subdomain strategy → `hearth.example` / `lk.` / `turn.`
- Q-008: Host networking → ✅ Required for WebRTC/TURN, all containers

**Deferred Decisions:**
- Exact TURN cert sync mechanism (symlink vs. copy script vs. shared volume)
- LiveKit cert hot-reload investigation (separate research task)
- UDP port range reduction (50000-50100 vs. 50000-60000)
- Health check endpoints and monitoring strategy
