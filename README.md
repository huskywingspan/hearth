# 🏠 Hearth — The Digital Living Room

> A privacy-first, self-hosted communication platform. Warm, intimate, high-fidelity voice and chat.

**Codename:** Project Vesta | **Version:** v0.1-ember (Backend Skeleton + Chat MVP)

---

## Quick Start (Development)

### Prerequisites
- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- A domain with 3 DNS A records pointing to your server (or use localhost for dev)

### Local Development (without Docker)

```bash
cd backend
go mod tidy
go build -o hearth .
./hearth serve --http=0.0.0.0:8090
```

PocketBase admin UI: http://localhost:8090/_/

### Docker Deployment

```bash
# 1. Clone and configure
git clone <repo-url> hearth
cd hearth
cp config/.env.example .env
# Edit .env with your domain and secrets

# 2. Generate secrets
openssl rand -hex 32  # → LIVEKIT_API_SECRET
openssl rand -hex 32  # → HMAC_SECRET_CURRENT
openssl rand -hex 32  # → PB_ENCRYPTION_KEY

# 3. Set up DNS
# Point these A records to your server IP:
#   hearth.example     → <server-ip>
#   lk.hearth.example  → <server-ip>
#   turn.hearth.example → <server-ip>

# 4. Launch
docker compose up -d

# 5. Verify
curl http://localhost:8090/api/health
curl http://localhost:8090/metrics
```

### Firewall Rules

```bash
sudo ufw allow 80/tcp          # ACME (Let's Encrypt)
sudo ufw allow 443/tcp         # HTTPS / WSS / TURN-TLS
sudo ufw allow 7881/tcp        # LiveKit TCP fallback
sudo ufw allow 3478/udp        # TURN/UDP
sudo ufw allow 50000:60000/udp # WebRTC media
```

---

## Architecture

```
Internet :443 → Caddy L4 (TLS SNI) → turn.hearth.example → LiveKit TURN :5349
                                    → lk.hearth.example   → LiveKit API  :7880
                                    → hearth.example       → PocketBase   :8090
```

| Component | Memory Budget | Purpose |
|-----------|-------------|---------|
| PocketBase | 250 MB | Auth, chat API, realtime, cron jobs |
| LiveKit | 400 MB | WebRTC SFU, spatial audio |
| Caddy | ~15 MB | TLS termination, SNI routing |
| OS/headroom | ~200 MB | Kernel, safety margin |
| **Total** | **1 GB** | |

---

## API Endpoints

### PocketBase Built-in
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/collections/users/auth-with-password` | Login |
| POST | `/api/collections/users/records` | Register |
| POST | `/api/collections/users/auth-refresh` | Token refresh |

### Hearth Custom
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/hearth/presence/heartbeat` | Yes | Client heartbeat (every 30s) |
| GET | `/api/hearth/presence/{roomId}` | Yes | Online users in room |
| POST | `/api/hearth/invite/generate` | Yes | Create signed invite URL |
| POST | `/api/hearth/invite/validate` | No | Validate invite token |
| GET | `/api/hearth/pow/challenge` | No | Get PoW puzzle |
| POST | `/api/hearth/pow/verify` | No | Submit PoW solution |
| POST | `/api/hearth/rooms/{id}/token` | Yes | Get LiveKit room token |
| GET | `/metrics` | No | Prometheus metrics |

---

## Project Structure

```
hearth/
├── backend/
│   ├── main.go                  # PocketBase bootstrap + hook registration
│   ├── go.mod                   # Go module definition
│   └── hooks/
│       ├── pragmas.go           # SQLite WAL pragma injection
│       ├── collections.go       # Programmatic collection creation
│       ├── auth.go              # Auth hooks (TTL enforcement, auto-membership)
│       ├── message_gc.go        # Cron: expired message sweep (every 1 min)
│       ├── vacuum.go            # Cron: nightly VACUUM (4 AM)
│       ├── presence.go          # In-memory presence map + endpoints
│       ├── invite.go            # HMAC invite token generation + validation
│       ├── pow.go               # Proof-of-Work challenge endpoints
│       ├── livekit_token.go     # LiveKit JWT generation
│       ├── metrics.go           # Prometheus /metrics endpoint
│       ├── helpers.go           # Shared utilities
│       └── hooks_test.go        # Unit + integration tests
├── config/
│   ├── caddy.yaml               # Caddy L4 TLS SNI routing
│   ├── livekit.yaml             # LiveKit optimized config
│   └── .env.example             # Template env vars
├── docker/
│   ├── caddy/Dockerfile         # Custom Caddy build (L4 + YAML)
│   └── Dockerfile.pocketbase    # Multi-stage Go build
├── docker-compose.yaml          # 3 containers, host networking
├── docs/                        # Design docs, research, specs
└── .gitignore
```

---

## Testing

```bash
cd backend
go test ./hooks/ -v -race
```

---

## License

TBD
