# R-012: Remote Access Architecture — Cloudflare Tunnel, VPS, and WebRTC Constraints

**Date:** 2026-02-15 | **Status:** Complete
**Blocks:** FF-010 (Cloudflare Tunnel), FF-011 (Tunnel config), FF-012 (Connect-to-server flow)
**Informed by:** R-002 (Caddy/LiveKit TLS), R-005 (LiveKit SDK), LiveKit deployment docs, Cloudflare Tunnel docs

---

## Executive Summary

**Cloudflare Tunnel cannot carry LiveKit voice/video traffic.** It only supports HTTP/HTTPS/WebSocket — not the raw UDP that WebRTC media requires. However, it works perfectly for PocketBase (HTTP + SSE).

This creates a **two-mode deployment architecture:**

1. **Quick Test Mode** (CF Tunnel) — For development and demoing chat from home. One command. Chat works flawlessly. Voice falls back to TCP (degraded quality, acceptable for testing).
2. **Production Mode** (VPS with public IP) — For real use. Caddy handles TLS, LiveKit gets full UDP access, voice is crisp. This is the target deployment from our master plan.

**Key insight:** The roadmap targets a VPS with a public IP. A VPS already has a public IP. The "tunneling problem" only exists during development when running from a home machine behind NAT. For production, there's nothing to tunnel.

---

## The Core Problem

### What Cloudflare Tunnel Does

Cloudflare Tunnel (`cloudflared`) creates an **outbound-only** encrypted connection from your server to Cloudflare's edge network. Traffic from the internet hits Cloudflare first, then is forwarded to your origin through the tunnel. No inbound ports need to be opened.

**Supported protocols (public hostnames):**
- HTTP / HTTPS ✅
- WebSocket ✅
- SSE (Server-Sent Events) ✅
- SSH (via browser-rendered terminal) ✅
- RDP ✅
- gRPC (private subnet routing only) ✅

**NOT supported (public hostnames):**
- Raw UDP ❌
- Raw TCP (Layer 4) ❌ — only HTTP-layer proxying

> CF Tunnel CAN proxy UDP via WARP (private networks), but that requires every connecting client to install the WARP app and join your Zero Trust organization. Not viable for "invite a friend to your House."

### What LiveKit Needs

LiveKit (WebRTC SFU) uses multiple ports:

| Port | Protocol | Purpose | Works through CF Tunnel? |
|------|----------|---------|--------------------------|
| 7880 | HTTP/WS | API + WebSocket signaling | ✅ Yes |
| 50000-60000 | UDP | ICE media transport (voice/video) | ❌ No |
| 7881 | TCP | ICE/TCP fallback (when UDP blocked) | ❌ No (not HTTP) |
| 5349 | TLS | TURN/TLS relay (looks like HTTPS) | ⚠️ Theoretically possible but unsupported |

### What PocketBase Needs

| Port | Protocol | Purpose | Works through CF Tunnel? |
|------|----------|---------|--------------------------|
| 8090 | HTTP/WS | REST API + SSE real-time | ✅ Yes, perfectly |

**Bottom line:** PocketBase is a pure HTTP service — CF Tunnel is ideal for it. LiveKit's media transport is fundamentally UDP — CF Tunnel cannot carry it.

---

## LiveKit TCP Fallback: How Bad Is It?

LiveKit has built-in fallback for restrictive networks:

1. **ICE/TCP (port 7881):** When UDP is blocked, LiveKit falls back to TCP for media transport.
2. **TURN/TLS (port 5349):** When even non-HTTPS TCP is blocked (corporate firewalls), LiveKit's embedded TURN server wraps media in TLS on port 443/5349.

### Quality Impact of TCP Fallback

| Factor | UDP (normal) | TCP fallback |
|--------|-------------|--------------|
| Latency | ~20-50ms | ~50-150ms (+30-100ms) |
| Head-of-line blocking | None (packets independent) | Yes (lost packet blocks all subsequent) |
| Packet loss recovery | Skip and continue (acceptable for voice) | Retransmit + wait (causes audio glitches) |
| Jitter | Low | Higher (TCP congestion control adds variability) |
| Voice quality (5 people) | Excellent | Acceptable with occasional artifacts |
| Video quality (5 people) | Good at 480p | Stuttery, buffering likely |

**For Hearth's use case (5-10 friends, voice-first):** TCP fallback is *usable* but noticeably worse than UDP. Audio will have occasional glitches under packet loss. Video (already capped at 480p/15fps) will be rough. The "5-minute-to-voice" UX benchmark demands crisp audio — TCP fallback does not meet this bar for production.

---

## Deployment Options Evaluated

### Option A: VPS with Public IP ⭐ RECOMMENDED (Production)

Deploy all 3 containers (Caddy + PocketBase + LiveKit) to a VPS. Caddy handles TLS with Let's Encrypt. LiveKit gets direct public IP access for UDP media.

| Factor | Assessment |
|--------|-----------|
| Voice quality | Full UDP — optimal |
| Setup complexity | `git clone && docker compose up` |
| Cost | $4-6/mo (Hetzner CX22, Oracle Cloud free, DigitalOcean $4) |
| RAM fit | ✅ All within 1GB budget |
| Privacy | ✅ No intermediary sees traffic. Self-hosted TLS. |
| Port requirements | 80, 443, 7880, 7881, 50000-60000 (standard VPS firewall rules) |
| Domain needed | Yes — for Let's Encrypt TLS |

**This is the target deployment from our master plan.** A VPS has a public IP by definition. There's nothing to tunnel.

### Option B: CF Tunnel for PocketBase (Chat Development/Testing)

Run PocketBase on home machine, expose via CF Tunnel. Use for v0.3 development when voice isn't in scope yet.

| Factor | Assessment |
|--------|-----------|
| Chat quality | ✅ Perfect — HTTP + SSE works flawlessly |
| Voice quality | ❌ Not possible through CF Tunnel (no UDP, no raw TCP) |
| Setup complexity | `cloudflared tunnel --url http://localhost:8090` (one command) |
| Cost | Free (CF Tunnel free tier) |
| RAM overhead | ~30-50MB for cloudflared (Go binary, same as our stack) |
| Privacy | ⚠️ Traffic routes through Cloudflare. They can see metadata (not content if E2EE). |
| Domain needed | No — CF generates a `*.trycloudflare.com` URL (quick tunnels) or use your own domain |

**Perfect for v0.3 "First Friend" sprint** where only chat needs to work remotely. One command, zero configuration, friends get a URL and connect.

### Option C: Tailscale (Private Testing Network)

Both you and your friend install Tailscale. Creates a WireGuard mesh. Friend connects directly to your machine's Tailscale IP.

| Factor | Assessment |
|--------|-----------|
| Chat quality | ✅ Full performance |
| Voice quality | ✅ Full UDP — WireGuard tunnels everything |
| Setup complexity | Both parties install Tailscale client + join tailnet |
| Cost | Free for 3 users (Personal plan) |
| RAM overhead | ~30-50MB for tailscaled |
| Privacy | ✅ WireGuard E2E. Tailscale coordinates, doesn't see traffic. |
| Friction | ❌ Friend must install Tailscale app. Not "scan QR and go." |

**Good for trusted testers** who are willing to install software. Bad for the "First Friend" UX where someone should be able to join via URL/QR code. WireGuard provides full UDP tunneling, so voice works perfectly.

### Option D: Home Server + Port Forwarding (No Tunnel)

Open router ports for PocketBase and LiveKit. Friends connect via your public IP or dynamic DNS.

| Factor | Assessment |
|--------|-----------|
| Quality | ✅ Full performance, no intermediary |
| Setup complexity | ❌ Router configuration, dynamic DNS, firewall rules, port range 50000-60000 |
| Cost | Free |
| Privacy | ❌ Exposes home IP address |
| Security | ❌ Attack surface. Most home ISPs block or restrict port ranges. |

**Not recommended.** Too much friction, security risk for users, and ISP restrictions on UDP port ranges make it unreliable.

### Option E: Hybrid — CF Tunnel (PocketBase) + Direct (LiveKit)

CF Tunnel for PocketBase API. LiveKit exposed directly on public IP with its own TLS.

| Factor | Assessment |
|--------|-----------|
| Chat quality | ✅ Perfect through CF Tunnel |
| Voice quality | ✅ Full UDP on direct IP |
| Complexity | ❌ High — two different access paths, split DNS, mixed networking |
| Requirement | Still needs public IP (VPS or port forwarding) for LiveKit |

**Over-engineered.** If you already have a public IP for LiveKit, just put PocketBase behind Caddy on the same IP. No tunnel needed.

---

## cloudflared Resource Footprint

CF's docs recommend 4GB RAM + 4 CPU for enterprise (8,000 WARP users). For Hearth's scale (~20 users, minimal HTTP traffic):

| Resource | Estimated Usage |
|----------|----------------|
| RAM | ~30-50MB (Go binary, idle tunnel) |
| CPU | Near zero (proxying a few HTTP requests) |
| Disk | ~30MB (cloudflared binary) |
| Network | Adds ~5-10ms latency vs direct (routes through CF edge) |

Within our 1GB budget if used as a 4th container, but only for development — not production.

---

## Privacy Assessment

| Concern | CloudFlare Tunnel | VPS Direct | Tailscale |
|---------|-------------------|------------|-----------|
| Traffic visibility | CF edge can see unencrypted HTTP | No intermediary | Tailscale coordinates only; WireGuard E2E |
| Metadata logging | CF logs connections (can opt out of some) | Server logs only | Tailscale coordination server sees connection metadata |
| IP exposure | User IP → CF edge → tunnel. Origin IP hidden. | User IP → VPS IP directly | Both IPs visible to Tailscale coordination, WireGuard tunnel is E2E |
| TLS termination | At CF edge (then re-encrypted to tunnel) | At Caddy on your server | At endpoints (WireGuard) |
| Hearth philosophy fit | ⚠️ "Privacy by default" — traffic through third party | ✅ Full control | ✅ Encrypted mesh, minimal metadata |

**For a privacy-first project,** routing all user traffic through Cloudflare's infrastructure is a philosophical tension. Acceptable for development/testing; not ideal for production. VPS direct or Tailscale better align with our principles.

---

## Architecture Recommendation

### Two-Mode Deployment

```
┌─────────────────────────────────────────────────────────────────┐
│                    DEVELOPMENT / TESTING                        │
│                                                                 │
│   Home Machine                    Cloudflare Edge               │
│   ┌──────────┐    outbound       ┌──────────────┐              │
│   │PocketBase│───tunnel──────────│CF Tunnel Proxy│──→ Internet  │
│   │ :8090    │    (HTTP/WSS)     │  *.trycloudflare.com         │
│   └──────────┘                   └──────────────┘              │
│                                                                 │
│   LiveKit NOT exposed (voice not in v0.3 scope)                │
│   Friends access chat via CF-generated URL                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    PRODUCTION (VPS)                              │
│                                                                 │
│   VPS (1 vCPU, 1GB, public IP)                                 │
│   ┌───────┐  ┌──────────┐  ┌────────┐                         │
│   │ Caddy │  │PocketBase│  │LiveKit │                          │
│   │ :443  │  │ :8090    │  │ :7880  │                          │
│   │ (TLS) │  │ (API)    │  │ :50000-60000 (UDP)               │
│   └───┬───┘  └──────────┘  └────────┘                         │
│       │       ↑ reverse      ↑ SNI route                       │
│       └───────┴──────────────┘                                 │
│                                                                 │
│   Domain: hearth.example.com + lk.hearth.example.com            │
│   TLS: Let's Encrypt via Caddy (auto-renewal)                  │
│   All UDP ports open in VPS firewall                            │
│   No tunnel, no intermediary, full performance                  │
└─────────────────────────────────────────────────────────────────┘
```

### Roadmap Impact

The original roadmap had FF-010 as "Research: Cloudflare Tunnel setup for PocketBase." This research reveals the scope should be broader:

| Original | Revised |
|----------|---------|
| FF-010: Research CF Tunnel for PocketBase | FF-010: ✅ Complete (this report, R-012) |
| FF-011: Tunnel configuration + docs | FF-011: Quick-test mode (CF Tunnel for chat-only dev) |
| FF-012: Connect-to-server flow | FF-012: Dynamic server URL in frontend (supports both modes) |
| (not planned) | FF-017: VPS production deployment guide |

### For Builder

**v0.3 "First Friend" implementation:**

1. **Frontend: Dynamic server URL** — Currently the PocketBase URL is hardcoded. Add a "connect to server" flow where the user enters/scans a URL (e.g., `https://abc123.trycloudflare.com` for testing, `https://hearth.example.com` for production). Store in localStorage.

2. **Quick-test script** — Add a `scripts/quick-test.sh` that runs `cloudflared tunnel --url http://localhost:8090` and prints the generated URL + QR code. Zero configuration required.

3. **VPS deployment** — The existing Docker Compose from R-002/R-003 already handles this. Document the VPS setup: create droplet → clone repo → set domain → `docker compose up -d`.

4. **LiveKit can wait** — Voice is v0.4 scope. The VPS deployment guide needs to include LiveKit + TURN config, but that's a v0.4 deliverable.

---

## Quick-Test Mode: cloudflared One-Liner

For v0.3 development, a developer can expose their PocketBase to a friend with zero setup:

```bash
# No account needed. Generates a random *.trycloudflare.com URL.
cloudflared tunnel --url http://localhost:8090
```

Output:
```
INF +-------------------------------------------+
INF | Your quick tunnel URL is:                  |
INF | https://abc123-random.trycloudflare.com    |
INF +-------------------------------------------+
```

Share that URL → friend opens it in browser → sees the Hearth frontend → can sign up and chat. The SSE real-time subscriptions work through CF Tunnel's WebSocket support.

**This is the "First Friend" moment** — zero DNS, zero domain, zero VPS, zero cost.

---

## Gotchas for Builder

1. **PocketBase SSE through CF Tunnel:** Works, but CF has a 100-second idle timeout on HTTP connections. PocketBase SSE sends keepalives, so this should be fine. If not, the existing reconnection logic (exponential backoff from K-016) handles it.

2. **CF Tunnel quick mode doesn't persist URLs.** Each restart generates a new URL. For persistent URLs, you need a Cloudflare account + named tunnel + domain. This is fine for dev — the QR code / URL sharing is per-session anyway.

3. **LiveKit JWT tokens contain the server URL.** When the frontend connects to LiveKit, the JWT must reference the correct LiveKit endpoint. In quick-test mode (no LiveKit exposed), the voice features should be gracefully disabled in the UI.

4. **CORS:** CF Tunnel preserves origin headers. PocketBase CORS config should allow the `*.trycloudflare.com` origin or use `*` for dev mode.

5. **cloudflared needs to be installed separately.** It's a single binary download, not bundled in our Docker Compose. Alternatively, add it as an optional Docker service with an `--profile quick-test` flag.

---

## Sources

- [Cloudflare Tunnel overview](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) — Protocol support, outbound-only architecture
- [Cloudflare Tunnel FAQ](https://developers.cloudflare.com/cloudflare-one/faq/cloudflare-tunnels-faq/) — WebSocket support confirmed, no UDP mention
- [Cloudflare Tunnel system requirements](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/configure-tunnels/tunnel-availability/system-requirements/) — 4GB/4CPU enterprise recommendation, "designed for Raspberry Pi"
- [LiveKit deployment guide](https://docs.livekit.io/home/self-hosting/deployment/) — TURN/TLS setup, host networking, SSL termination
- [LiveKit ports & firewall](https://docs.livekit.io/home/self-hosting/ports-firewall/) — Port table: 7880 (API), 50000-60000 (UDP), 7881 (TCP fallback)
- [Tailscale Funnel](https://tailscale.com/kb/1223/funnel) — TCP proxy only (ports 443/8443/10000), TLS-encrypted connections only, bandwidth limits
- [Tailscale pricing](https://tailscale.com/pricing) — Free: 3 users/100 devices. Personal Plus: $5/mo
- R-002: Caddy + LiveKit TLS Configuration — Our existing deployment architecture
- R-005: LiveKit React SDK — Connection lifecycle and audio plugins
