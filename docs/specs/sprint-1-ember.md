# Sprint 1 — v0.1 "Ember" (Backend Skeleton + Chat MVP)

> **Sprint Goal:** A running PocketBase instance with optimized SQLite, basic auth, message CRUD with expiry, HMAC invite tokens, LiveKit JWT generation, and Prometheus metrics. **API-only — no frontend.** Proves the data layer works within the 1GB budget.
>
> **Target:** April 2026
> **Owner:** Builder Agent
> **Research Prerequisites:** ✅ R-001 through R-006 complete. Backend fully unblocked.

---

## Research References

Every code pattern in this spec has been verified against current documentation. Do NOT use patterns from AI training data — use these reports:

| Report | Relevance | Key Patterns |
|--------|-----------|-------------|
| [R-001](../research/R-001-pocketbase-api-verification.md) | Go API patterns for PocketBase v0.36.2 | `app.OnServe().BindFunc()`, `app.DB()`, `app.Cron().MustAdd()`, `core.NewRecord()` |
| [R-002](../research/R-002-caddy-livekit-tls-config.md) | Docker Compose, Caddy YAML, LiveKit config | Complete `caddy.yaml`, `livekit.yaml`, `docker-compose.yaml` templates |
| [R-003](../research/R-003-container-topology.md) | ADR-001: Container topology decision | Docker Compose, 3 containers, all `network_mode: "host"` |

⚠️ **PocketBase API Warning:** The Go API changed dramatically at v0.23. Old patterns (`app.Dao()`, `app.OnBeforeServe()`) will NOT compile. See R-001 migration table.

---

## File Tree (Target State After Sprint 1)

```
hearth/
├── backend/
│   ├── main.go                  # PocketBase bootstrap + hook registration
│   ├── go.mod / go.sum          # Go module (PocketBase + LiveKit server SDK)
│   ├── hooks/
│   │   ├── pragmas.go           # SQLite WAL pragma injection
│   │   ├── collections.go       # Programmatic collection creation/migration
│   │   ├── message_gc.go        # Cron: expired message sweep (every 1 min)
│   │   ├── vacuum.go            # Cron: nightly VACUUM (4 AM)
│   │   ├── presence.go          # In-memory presence map + heartbeat endpoints
│   │   ├── invite.go            # HMAC invite token generation + validation
│   │   ├── pow.go               # Proof-of-Work challenge endpoint
│   │   ├── livekit_token.go     # LiveKit JWT generation for room access
│   │   └── metrics.go           # Prometheus /metrics endpoint
│   └── hooks/
│       └── hooks_test.go        # Unit + integration tests
├── config/
│   ├── caddy.yaml               # Caddy L4 TLS SNI routing (from R-002)
│   ├── livekit.yaml             # LiveKit optimized config (from R-002)
│   └── .env.example             # Template env vars for self-hosters
├── docker/
│   ├── Dockerfile.pocketbase    # Multi-stage Go build
│   └── docker-compose.yaml      # 3 containers, host networking (from R-002/R-003)
├── .gitignore
└── README.md                    # Setup instructions for development
```

---

## Phase 0: Project Scaffolding

**Goal:** Go project compiles. Docker Compose starts all 3 containers. Caddy terminates TLS. LiveKit accepts connections. PocketBase serves the admin UI.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **0.1** | E-004 | Scaffold Go backend | `go build` succeeds. PocketBase starts. Admin UI accessible at `:8090/_/`. |
| **0.2** | E-005 | Create Docker Compose with memory constraints | `docker compose up -d` starts 3 containers. `GOMEMLIMIT` set per service. |
| **0.3** | E-006 | Create Caddy config (L4 TLS SNI) | HTTPS terminates. `hearth.example → :8090`, `lk.hearth.example → :7880`, `turn.hearth.example → :5349`. |
| **0.4** | E-007 | Create LiveKit config | LiveKit starts and accepts WebSocket connections on `:7880`. ICE Lite, DTX, no transcoding. |
| **0.5** | — | `.env.example` + README | Self-hoster can clone, edit `.env`, `docker compose up -d`. |

### 0.1 — Scaffold Go Backend (E-004)

**File:** `backend/main.go`

```go
package main

import (
    "log"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"

    "hearth/hooks"
)

func main() {
    app := pocketbase.New()

    // Register all hooks
    hooks.RegisterPragmas(app)
    hooks.RegisterCollections(app)
    hooks.RegisterMessageGC(app)
    hooks.RegisterVacuum(app)
    hooks.RegisterPresence(app)
    hooks.RegisterInvite(app)
    hooks.RegisterPoW(app)
    hooks.RegisterLiveKitToken(app)
    hooks.RegisterMetrics(app)

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

**File:** `backend/go.mod`

```
module hearth

go 1.23

require (
    github.com/pocketbase/pocketbase v0.36.2
    github.com/livekit/protocol v1.x.x  // for JWT generation
)
```

**Notes for Builder:**
- Use `go 1.23` or latest stable. PocketBase v0.36.2 requires Go 1.23+.
- The LiveKit server SDK is only needed for JWT token generation — we import `github.com/livekit/protocol/auth` only.
- Each hook file has a `RegisterXxx(app *pocketbase.PocketBase)` function to keep `main.go` clean.
- PocketBase auto-creates `pb_data/` on first run. Add to `.gitignore`.

### 0.2 — Docker Compose (E-005)

**File:** `docker/docker-compose.yaml`

Copy the template from [R-002 §Docker Compose](../research/R-002-caddy-livekit-tls-config.md#docker-compose-docker-composeyaml) with these key constraints:

| Container | Image | Memory Control | Network |
|-----------|-------|---------------|---------|
| `caddy` | Custom build (see R-002 Dockerfile) | — (~15MB, negligible) | `network_mode: "host"` |
| `livekit` | `livekit/livekit-server:latest` | `GOMEMLIMIT=400MiB` | `network_mode: "host"` |
| `pocketbase` | Custom build (our Go binary) | `GOMEMLIMIT=250MiB` | `network_mode: "host"` |

**File:** `docker/Dockerfile.pocketbase`

```dockerfile
# Multi-stage build for minimal image
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o hearth .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/hearth /usr/local/bin/hearth
EXPOSE 8090
ENTRYPOINT ["hearth", "serve", "--http=0.0.0.0:8090"]
```

⚠️ **Builder Note:** PocketBase embeds SQLite via CGo by default. If using the pure-Go SQLite driver (`modernc.org/sqlite`), `CGO_ENABLED=0` works. If using `mattn/go-sqlite3`, you need `CGO_ENABLED=1` and the `gcc` build dependency. **Check the PocketBase dependency tree** — as of v0.36, PocketBase uses the pure-Go driver, so `CGO_ENABLED=0` should work. Verify this before finalizing the Dockerfile.

### 0.3 — Caddy Config (E-006)

**File:** `config/caddy.yaml`

Copy the complete YAML from [R-002 §Caddy Configuration](../research/R-002-caddy-livekit-tls-config.md#complete-caddy-configuration-caddyyaml). Key points:

- Uses `caddy-l4` module for Layer 4 TLS SNI routing
- TURN route (`turn.hearth.example`) = raw TLS passthrough to `:5349`
- LiveKit route (`lk.hearth.example`) = TLS termination + HTTP reverse proxy to `:7880`
- PocketBase route (`hearth.example`) = TLS termination + HTTP reverse proxy to `:8090`
- Domain names are parameterized via `{env.HEARTH_DOMAIN}` (read from `.env`)

### 0.4 — LiveKit Config (E-007)

**File:** `config/livekit.yaml`

Copy from [R-002 §LiveKit Configuration](../research/R-002-caddy-livekit-tls-config.md#livekit-configuration-livekityaml). Critical settings:

```yaml
port: 7880
rtc:
  port_range_start: 50000
  port_range_end: 60000
  use_ice_lite: true
  tcp_fallback_port: 7881

audio:
  active_sensitivity: 20           # dBFS threshold for "speaking"
  audio_level_interval: 300        # ms between audio level updates

room:
  auto_create: false               # Rooms created via PocketBase API only
  max_participants: 25

keys:
  hearth-api: ${LIVEKIT_API_SECRET}   # From .env
```

⚠️ **Important:** Set `auto_create: false`. Rooms are created through PocketBase, which then calls LiveKit's API. This prevents unauthorized room creation.

### 0.5 — Environment Template

**File:** `config/.env.example`

```bash
# Hearth — Environment Configuration
# Copy to .env and fill in your values

# Domain (no protocol prefix)
HEARTH_DOMAIN=hearth.example

# PocketBase
PB_ENCRYPTION_KEY=       # 32-byte hex string for data encryption at rest

# LiveKit
LIVEKIT_API_KEY=hearth-api
LIVEKIT_API_SECRET=       # Generate: openssl rand -hex 32

# HMAC Invite Tokens
HMAC_SECRET_CURRENT=      # Generate: openssl rand -hex 32
HMAC_SECRET_OLD=          # Previous key (for rotation grace period)

# Proof of Work
POW_DIFFICULTY=20         # Leading zero bits (20 ≈ 1-2s on modern hardware)
```

---

## Phase 1: Data Layer

**Goal:** SQLite is optimized. Collections exist. Messages have TTL. GC sweeps expired messages. Presence is tracked in-memory.

**Depends on:** Phase 0 complete (PocketBase starts).

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **1.1** | E-010 | SQLite WAL pragma injection | PRAGMAs applied on every boot. Verified via `PRAGMA journal_mode` returning `wal`. |
| **1.2** | E-011 | Define PocketBase collections | Users, Rooms, Messages, RoomMembers collections exist. Schema matches spec below. |
| **1.3** | E-012 | `expires_at` index on Messages | Index exists. `EXPLAIN QUERY PLAN` shows index usage for GC query. |
| **1.4** | E-013 | Cron: message GC (every 1 min) | Expired messages are deleted within 60s of `expires_at`. |
| **1.5** | E-014 | Cron: nightly VACUUM (4 AM) | VACUUM executes. DB file size decreases after message deletion. |
| **1.6** | E-015 | In-memory presence map | Heartbeat endpoint works. Presence list returns online users. Stale users (>60s) are swept. |

### 1.1 — SQLite Pragma Injection (E-010)

**File:** `backend/hooks/pragmas.go`

Use the `app.OnBootstrap()` hook — this fires before any database operations, including migrations.

```go
package hooks

import (
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func RegisterPragmas(app *pocketbase.PocketBase) {
    app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
        pragmas := []string{
            "PRAGMA journal_mode=WAL",
            "PRAGMA synchronous=NORMAL",
            "PRAGMA cache_size=-2000",
            "PRAGMA mmap_size=268435456",
            "PRAGMA busy_timeout=5000",
        }
        for _, pragma := range pragmas {
            if _, err := e.App.DB().NewQuery(pragma).Execute(); err != nil {
                return fmt.Errorf("failed to set %s: %w", pragma, err)
            }
        }
        return e.Next()
    })
}
```

**⚠️ Critical:** Must call `e.Next()` to continue the hook chain. Forgetting this silently breaks PocketBase startup.

**Verification:** After boot, send a request to PocketBase and check logs, or add a log line confirming PRAGMAs applied. For automated testing, query `PRAGMA journal_mode` via a custom endpoint and assert it returns `wal`.

### 1.2 — PocketBase Collections (E-011)

**File:** `backend/hooks/collections.go`

Create collections programmatically in `app.OnServe().BindFunc()` using `app.FindCollectionByNameOrId()` + `app.Save()`. This approach is idempotent (skips if collection already exists) and version-controllable (schema is in code, not in the admin UI).

**Schema Definition:**

#### `users` (Auth Collection — PocketBase built-in)
PocketBase creates this automatically. Extend with:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `display_name` | `text` | Yes | User-facing name (not email) |
| `avatar_url` | `text` | No | URL to avatar image |
| `status` | `text` | No | "cozy", "away", "dnd" |

#### `rooms` (Base Collection)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name` | `text` | Yes | Room display name |
| `slug` | `text` | Yes, unique | URL-safe identifier (`the-kitchen`) |
| `owner` | `relation` → users | Yes | Room creator |
| `description` | `text` | No | Short description |
| `default_ttl` | `number` | Yes | Default message lifetime in seconds (e.g., 3600 = 1hr) |
| `max_participants` | `number` | Yes | Default: 20 |
| `allow_video` | `bool` | Yes | Default: `false` (voice-first) |
| `livekit_room_name` | `text` | Yes, unique | LiveKit room identifier |

**API Rules (Access Control):**
- List/View: Authenticated users who are members (relation check via `room_members`)
- Create: Authenticated users
- Update/Delete: Owner only (`@request.auth.id = owner.id`)

#### `messages` (Base Collection)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `room` | `relation` → rooms | Yes | Which room |
| `author` | `relation` → users | Yes | Who sent it |
| `body` | `text` | Yes | Message content |
| `type` | `select` | Yes | `"text"`, `"system"`, `"emote"` |
| `expires_at` | `date` | Yes | When to GC. Set by server: `now() + room.default_ttl` |
| `created` | `autodate` | Auto | PocketBase sets this automatically |

**API Rules:**
- List/View: Authenticated room members
- Create: Authenticated room members. **Server-side hook** overrides `expires_at` = `now + room.default_ttl` (clients cannot set their own TTL).
- Update: Author only, within 60s of creation
- Delete: Author or room owner

#### `room_members` (Base Collection)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `room` | `relation` → rooms | Yes | |
| `user` | `relation` → users | Yes | |
| `role` | `select` | Yes | `"owner"`, `"member"`, `"guest"` |
| `vouched_by` | `relation` → users | No | Who let them in (for guests) |
| `joined_at` | `autodate` | Auto | |

**Unique constraint:** `room` + `user` pair must be unique.

**API Rules:**
- List/View: Authenticated room members
- Create: Room owner or admin
- Delete: Room owner (kick), or self (leave)

### 1.3 — `expires_at` Index (E-012)

Create the index programmatically after collection creation:

```go
// In collections.go, after creating the messages collection:
_, err := app.DB().NewQuery(`
    CREATE INDEX IF NOT EXISTS idx_messages_expires_at 
    ON messages (expires_at) 
    WHERE expires_at IS NOT NULL
`).Execute()
```

**Verification:** Run `EXPLAIN QUERY PLAN DELETE FROM messages WHERE expires_at <= datetime('now')` and confirm `SEARCH messages USING INDEX idx_messages_expires_at`.

### 1.4 — Message GC Cron (E-013)

**File:** `backend/hooks/message_gc.go`

```go
func RegisterMessageGC(app *pocketbase.PocketBase) {
    app.Cron().MustAdd("hearth_message_gc", "* * * * *", func() {
        res, err := app.DB().
            NewQuery("DELETE FROM messages WHERE expires_at <= {:now}").
            Bind(dbx.Params{"now": time.Now().UTC().Format(time.RFC3339)}).
            Execute()
        if err != nil {
            app.Logger().Error("message GC failed", "error", err)
            return
        }
        if affected, _ := res.RowsAffected(); affected > 0 {
            app.Logger().Info("message GC sweep", "deleted", affected)
        }
    })
}
```

**Key decisions:**
- Runs every 60 seconds (cron `* * * * *`).
- Bulk `DELETE` — no per-message overhead.
- Uses the `idx_messages_expires_at` index → O(log n) scan.
- Logs only when messages were actually deleted (avoids log spam).

### 1.5 — Nightly VACUUM (E-014)

**File:** `backend/hooks/vacuum.go`

```go
func RegisterVacuum(app *pocketbase.PocketBase) {
    app.Cron().MustAdd("hearth_nightly_vacuum", "0 4 * * *", func() {
        if _, err := app.DB().NewQuery("VACUUM").Execute(); err != nil {
            app.Logger().Error("nightly VACUUM failed", "error", err)
        } else {
            app.Logger().Info("nightly VACUUM complete")
        }
    })
}
```

**Why:** `DELETE` only marks SQLite pages as free — data remains on disk until `VACUUM` rewrites the file. This is critical for Hearth's privacy promise (physical data erasure).

⚠️ `VACUUM` is a blocking operation that rewrites the entire DB. At Hearth's scale (~20 users), this takes milliseconds. At larger scale, consider `PRAGMA incremental_vacuum` instead.

### 1.6 — In-Memory Presence Map (E-015)

**File:** `backend/hooks/presence.go`

```go
type PresenceEntry struct {
    UserID    string
    RoomID    string
    UpdatedAt time.Time
}

type PresenceMap struct {
    mu      sync.RWMutex
    entries map[string]*PresenceEntry // key: userID
}
```

**Endpoints (registered in `app.OnServe()`):**

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/api/hearth/presence/heartbeat` | Client sends every 30s. Body: `{ "room_id": "..." }`. Requires auth. |
| `GET` | `/api/hearth/presence/:roomId` | Returns list of online users in a room. Requires auth + room membership. |

**Sweep cron:** Every 2 minutes, remove entries where `UpdatedAt` is older than 60 seconds.

```go
app.Cron().MustAdd("hearth_presence_sweep", "*/2 * * * *", func() {
    threshold := time.Now().Add(-60 * time.Second)
    pm.mu.Lock()
    defer pm.mu.Unlock()
    for k, v := range pm.entries {
        if v.UpdatedAt.Before(threshold) {
            delete(pm.entries, k)
        }
    }
})
```

**Design rationale:**
- **Not persisted to SQLite.** Presence is ephemeral — generating WAL writes for heartbeats would be wasteful I/O.
- **`sync.RWMutex`** allows concurrent reads (presence queries) with exclusive writes (heartbeats, sweeps).
- Presence data is lost on server restart — this is intentional. Users re-establish presence after reconnecting.

---

## Phase 2: Auth & Security

**Goal:** Users can register/login, generate invite links, validate invites, solve PoW challenges, and get LiveKit room tokens.

**Depends on:** Phase 1 complete (collections exist).

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **2.1** | E-020 | Basic auth (email/password) | Register, login, logout, refresh via PocketBase built-in auth. |
| **2.2** | E-021 | HMAC invite generation | `POST /api/hearth/invite/generate` returns a signed URL. |
| **2.3** | E-022 | HMAC invite validation | `POST /api/hearth/invite/validate` accepts valid tokens, rejects expired/tampered. Constant-time comparison. |
| **2.4** | E-023 | Proof-of-Work challenge | `GET /api/hearth/pow/challenge` returns a puzzle. `POST /api/hearth/pow/verify` validates the solution. |
| **2.5** | E-024 | LiveKit JWT token generation | `POST /api/hearth/rooms/:id/token` returns a LiveKit JWT with correct grants. `canPublishVideo: false` by default. |

### 2.1 — Basic Auth (E-020)

PocketBase handles this out of the box. No custom code needed for the core flow:

- `POST /api/collections/users/auth-with-password` — login
- `POST /api/collections/users/records` — register (if collection allows)
- `POST /api/collections/users/auth-refresh` — token refresh

**Builder tasks:**
1. Configure the `users` auth collection to allow email/password auth
2. Set sensible defaults: minimum password length (10), token duration
3. Add an `app.OnRecordCreate("users")` hook to set `display_name` from the request if provided
4. Ensure auth tokens are returned with `httpOnly` cookie option disabled (SPA needs the token in JS for PocketBase SDK auth)

### 2.2 — HMAC Invite Generation (E-021)

**File:** `backend/hooks/invite.go`

```go
// POST /api/hearth/invite/generate
// Body: { "room_slug": "the-kitchen", "expires_in": 86400 }
// Returns: { "url": "https://hearth.example/join?r=the-kitchen&t=1735689600&s=f8a..." }

func generateInviteURL(roomSlug string, expiresAt int64, secret []byte, domain string) string {
    payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(payload))
    sig := hex.EncodeToString(mac.Sum(nil))
    return fmt.Sprintf("https://%s/join?r=%s&t=%d&s=%s", domain, roomSlug, expiresAt, sig)
}
```

**Access control:** Only room owners/members can generate invites for their rooms.

### 2.3 — HMAC Invite Validation (E-022)

```go
// POST /api/hearth/invite/validate
// Body: { "r": "the-kitchen", "t": "1735689600", "s": "f8a..." }

func validateInvite(roomSlug string, timestamp int64, signature string, secrets [][]byte) bool {
    if time.Now().Unix() > timestamp {
        return false // expired
    }
    payload := roomSlug + "." + strconv.FormatInt(timestamp, 10)
    for _, secret := range secrets {
        mac := hmac.New(sha256.New, secret)
        mac.Write([]byte(payload))
        expected := mac.Sum(nil)
        provided, err := hex.DecodeString(signature)
        if err != nil {
            return false
        }
        if hmac.Equal(expected, provided) {
            return true // hmac.Equal is constant-time
        }
    }
    return false
}
```

**Critical implementation notes:**
- `hmac.Equal()` uses `crypto/subtle.ConstantTimeCompare` internally — **never use `==` or `bytes.Equal`** for hash comparison (timing side-channel risk).
- `secrets` is an array of `[currentKey, oldKey]` — the two-key rotation system. This provides a grace period when rotating keys.
- On successful validation, create a `room_members` entry with `role: "guest"` and `vouched_by: null` (the Knock flow in v1.0 adds the voucher).

### 2.4 — Proof-of-Work Challenge (E-023)

**File:** `backend/hooks/pow.go`

Client Puzzle Protocol — SHA256 partial collision:

```
Challenge: find nonce such that SHA256(challenge_id + nonce) has N leading zero bits
```

**Endpoints:**

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/hearth/pow/challenge` | Returns `{ "challenge_id": "<random>", "difficulty": 20, "expires": 1735689600 }` |
| `POST` | `/api/hearth/pow/verify` | Body: `{ "challenge_id": "...", "nonce": "..." }`. Returns `{ "token": "<one-time-use PoW token>" }` |

**Difficulty tuning:**
- `difficulty: 20` = ~20 leading zero bits ≈ 1–2 seconds on modern hardware
- Self-hosters can tune via `POW_DIFFICULTY` env var
- Challenge IDs are random UUIDs with short expiry (5 min) to prevent pre-computation

**Where PoW is required:** Applied as a middleware on public endpoints — invite validation and guest Knock requests. Authenticated users skip PoW.

### 2.5 — LiveKit JWT Token Generation (E-024)

**File:** `backend/hooks/livekit_token.go`

```go
import "github.com/livekit/protocol/auth"

// POST /api/hearth/rooms/:id/token
// Requires: authenticated user who is a member of the room

func generateLiveKitToken(apiKey, apiSecret, roomName, identity, displayName string) (string, error) {
    at := auth.NewAccessToken(apiKey, apiSecret)
    grant := &auth.VideoGrant{
        RoomJoin: true,
        Room:     roomName,
    }
    at.SetVideoGrant(grant).
        SetIdentity(identity).
        SetName(displayName).
        SetValidFor(24 * time.Hour)

    return at.ToJWT()
}
```

**Grant policy — voice-first:**
- `RoomJoin: true` — allowed to join
- `CanPublish: true` — can publish audio
- `CanPublishData: true` — can publish data messages (for presence sync)
- `CanSubscribe: true` — can receive others' audio
- **`CanPublishSources: ["microphone"]`** — explicitly NO camera, NO screen share (enforces voice-first)
- Video is a room-level toggle (`room.allow_video`) that the host can enable. When enabled, token includes `"camera"` in `CanPublishSources`.

**Access control:** Only authenticated users who are members of the room (verified via `room_members` collection) can request a token.

---

## Phase 3: Observability

**Goal:** Prometheus-format metrics and structured logging for ops visibility.

**Depends on:** Phase 0 complete.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **3.1** | E-030 | `/metrics` endpoint | Returns Prometheus text format with heap, goroutines, SQLite stats, room count. |
| **3.2** | E-031 | Structured logging | JSON logs with level, timestamp, component, message. |

### 3.1 — Prometheus Metrics (E-030)

**File:** `backend/hooks/metrics.go`

Register a `GET /metrics` endpoint in `app.OnServe()`. Expose:

| Metric | Type | Source |
|--------|------|--------|
| `hearth_go_heap_bytes` | gauge | `runtime.MemStats.HeapAlloc` |
| `hearth_go_goroutines` | gauge | `runtime.NumGoroutine()` |
| `hearth_rooms_total` | gauge | `SELECT COUNT(*) FROM rooms` |
| `hearth_users_online` | gauge | `len(presenceMap.entries)` |
| `hearth_messages_total` | gauge | `SELECT COUNT(*) FROM messages` |
| `hearth_gc_deleted_total` | counter | Incremented by message GC cron |
| `hearth_sqlite_wal_pages` | gauge | `PRAGMA wal_checkpoint(PASSIVE)` returns log/checkpointed pages |

**Format:** Plain Prometheus text exposition (no need for a Prometheus client library — it's a simple text format):

```
# HELP hearth_go_heap_bytes Current Go heap allocation in bytes
# TYPE hearth_go_heap_bytes gauge
hearth_go_heap_bytes 12345678
```

**Access:** Unauthenticated (standard for metrics endpoints) but consider IP-restricting in production Caddy config.

### 3.2 — Structured Logging (E-031)

PocketBase v0.36 uses Go's `slog` package internally. Configure it on startup:

```go
// In main.go or pragmas.go (early init)
app.Logger() // PocketBase provides a structured logger
```

**Builder tasks:**
- Use `app.Logger().Info()`, `.Error()`, `.Warn()` throughout all hooks
- Include structured fields: `"component"`, `"room_id"`, `"user_id"`, `"action"`
- Log format: JSON (for machine parsing) in production, text in development
- Log levels: `INFO` for normal operations, `WARN` for recoverable errors, `ERROR` for failures

---

## Phase 4: Testing

**Goal:** Confidence that the core data layer, auth, and security work correctly.

**Depends on:** Phases 1 and 2 complete.

| Subtask | ID | Description | Acceptance Criteria |
|---------|----|-------------|---------------------|
| **4.1** | E-016 | Unit tests: message expiry, GC, presence | All pass. GC correctly deletes expired messages. Presence sweep removes stale entries. |
| **4.2** | E-025 | Integration tests: invite flow, PoW, auth | Full invite → validate → join flow works. PoW rejects invalid solutions. Auth flow works end-to-end. |

### 4.1 — Unit Tests (E-016)

**File:** `backend/hooks/hooks_test.go`

| Test | What It Verifies |
|------|-----------------|
| `TestPragmasApplied` | `PRAGMA journal_mode` returns `wal` after boot |
| `TestMessageCreatedWithTTL` | Creating a message auto-sets `expires_at = now + room.default_ttl` |
| `TestMessageGCSweep` | Insert messages with past `expires_at`, run GC, verify deleted |
| `TestMessageGCPreservesActive` | Insert messages with future `expires_at`, run GC, verify NOT deleted |
| `TestPresenceHeartbeat` | Send heartbeat, verify user appears in presence map |
| `TestPresenceSweep` | Set old `UpdatedAt`, run sweep, verify user removed |
| `TestPresenceConcurrency` | Concurrent heartbeats + reads don't race (use `-race` flag) |

**Testing approach:** PocketBase supports a test app (`pocketbase.NewTestApp()`) that runs against an in-memory database. Use this for all unit tests — no Docker required.

### 4.2 — Integration Tests (E-025)

| Test | What It Verifies |
|------|-----------------|
| `TestInviteGenerateAndValidate` | Generate invite → validate with correct params → success |
| `TestInviteExpired` | Generate invite with past timestamp → validation fails |
| `TestInviteTampered` | Modify any URL param → validation fails |
| `TestInviteKeyRotation` | Generate with old key → validate with `[newKey, oldKey]` → still works |
| `TestInviteAfterKeyDrop` | Generate with old key → validate with `[newKey]` only → fails |
| `TestPoWChallengeSolve` | Get challenge → find valid nonce → verify succeeds |
| `TestPoWInvalidNonce` | Submit wrong nonce → verify fails |
| `TestPoWExpiredChallenge` | Submit after challenge expiry → fails |
| `TestLiveKitTokenGrants` | Generate token → decode JWT → verify grants match (no video, has audio) |
| `TestAuthRegisterLogin` | Register user → login → get valid token → access protected endpoint |
| `TestRoomMembershipRequired` | Non-member tries to get room token → 403 |

---

## Dependency Graph

```
Phase 0 (Scaffolding)
  ├── 0.1 Go project ─────────────────────────────────────────┐
  ├── 0.2 Docker Compose ──── needs 0.1 (for Dockerfile)      │
  ├── 0.3 Caddy config ────── independent                     │
  ├── 0.4 LiveKit config ──── independent                     │
  └── 0.5 .env + README ──── independent                      │
                                                               │
Phase 1 (Data Layer) ──── needs: Phase 0.1                     │
  ├── 1.1 SQLite PRAGMAs ─── first (runs on bootstrap)        │
  ├── 1.2 Collections ─────── needs 1.1 (DB must be ready)    │
  ├── 1.3 expires_at index ── needs 1.2 (messages collection) │
  ├── 1.4 Message GC cron ── needs 1.3 (index for perf)       │
  ├── 1.5 VACUUM cron ─────── needs 1.1 (just DB access)      │
  └── 1.6 Presence map ────── independent (in-memory only)     │
                                                               │
Phase 2 (Auth & Security) ── needs: Phase 1.2 (collections)   │
  ├── 2.1 Basic auth ──────── needs 1.2 (users collection)    │
  ├── 2.2 HMAC invite gen ── needs 2.1 (auth for access ctrl) │
  ├── 2.3 HMAC invite val ── needs 2.2                        │
  ├── 2.4 PoW challenge ──── independent                      │
  └── 2.5 LiveKit JWT ─────── needs 1.2 + 2.1 (rooms + auth) │
                                                               │
Phase 3 (Observability) ──── needs: Phase 0.1                  │
  ├── 3.1 /metrics ─────────── can start early                 │
  └── 3.2 Structured logging ── can start early                │
                                                               │
Phase 4 (Testing) ──────── needs: Phases 1 + 2 complete        │
  ├── 4.1 Unit tests                                           │
  └── 4.2 Integration tests                                    │
```

---

## Definition of Done (Sprint 1)

Sprint 1 is complete when **ALL** of the following are true:

- [ ] `go build` succeeds with zero warnings
- [ ] `docker compose up -d` starts all 3 containers
- [ ] PocketBase admin UI is accessible
- [ ] SQLite PRAGMAs verified (WAL mode, correct cache size)
- [ ] All 4 collections exist with correct schemas
- [ ] Message creation auto-sets `expires_at`
- [ ] Message GC deletes expired messages within 60s
- [ ] Nightly VACUUM cron is registered
- [ ] Presence heartbeat + query + sweep all work
- [ ] Auth: register → login → refresh → protected endpoint
- [ ] HMAC invite: generate → validate → key rotation
- [ ] PoW: challenge → solve → verify
- [ ] LiveKit JWT: authenticated member gets valid token with voice-only grants
- [ ] `/metrics` returns Prometheus-format data
- [ ] All unit tests pass (including with `-race` flag)
- [ ] All integration tests pass
- [ ] Memory usage stays under 250MB for PocketBase under basic load

---

## Estimated Effort

| Phase | Subtasks | Estimated Effort | Parallelizable |
|-------|----------|-----------------|----------------|
| Phase 0 | 5 | 1 day | 0.3–0.5 can run in parallel |
| Phase 1 | 6 | 2 days | 1.5 + 1.6 can run in parallel |
| Phase 2 | 5 | 2 days | 2.4 is independent of 2.1–2.3 |
| Phase 3 | 2 | 0.5 day | Both parallelizable with Phase 1 |
| Phase 4 | 2 | 1.5 days | After Phases 1 + 2 |
| **Total** | **20** | **~7 days** | |

---

## Handoff: Researcher → Builder

**Task IDs:** E-004 through E-031
**Context:** All research complete (R-001–R-006). Go API patterns verified. Docker templates ready. No open blockers.
**Deliverables:** This spec + 6 research reports in `docs/research/`
**Blockers Resolved:** PocketBase API version (R-001), TLS routing (R-002), container topology (R-003/ADR-001)
**Open Items:** PocketBase CGo vs pure-Go SQLite driver — Builder should verify during E-004
**Next Step:** Begin Phase 0.1 (Go project scaffold)
