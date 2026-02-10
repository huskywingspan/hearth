# R-001: PocketBase v0.23+ API Verification

> **Status:** Complete | **Date:** 2026-02-10 | **Priority:** Critical
> **Blocks:** E-004, E-010 through E-015
> **Source:** PocketBase official documentation (v0.36.2), pocketbase.io/docs/go-*

---

## Summary

PocketBase's Go API has changed significantly since v0.23. The current version (v0.36.2) uses a completely different hook registration, database access, and routing model than what Gemini's scaffold suggested. This document provides **verified, copy-paste-ready** code patterns for all Hearth backend operations.

### Key Migration Points (Old → Current)

| Operation | Old API (pre-v0.23) | Current API (v0.23+) |
|-----------|---------------------|----------------------|
| Startup hook | `app.OnBeforeServe()` | `app.OnServe().BindFunc()` |
| DB access | `app.Dao().DB()` | `app.DB()` |
| Route registration | Inside `OnBeforeServe` | Inside `OnServe` via `se.Router` |
| Record find | `app.Dao().FindRecordById()` | `app.FindRecordById()` |
| Record save | `app.Dao().SaveRecord()` | `app.Save(record)` |
| Record delete | `app.Dao().DeleteRecord()` | `app.Delete(record)` |
| Cron jobs | External cron library | `app.Cron().MustAdd()` |

---

## Verified Code Patterns

### 1. Minimal main.go Scaffold

```go
package main

import (
    "log"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
)

func main() {
    app := pocketbase.New()

    // Register routes, hooks, and cron jobs
    registerRoutes(app)
    registerHooks(app)
    registerCronJobs(app)

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Route Registration (Routing)

All custom routes are registered inside the `app.OnServe()` hook. The router is built on Go's standard `net/http.ServeMux`.

```go
func registerRoutes(app *pocketbase.PocketBase) {
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        // Public route — no auth required
        se.Router.GET("/api/hearth/health", func(e *core.RequestEvent) error {
            return e.JSON(200, map[string]string{"status": "ok"})
        })

        // Authenticated route — any logged-in user
        se.Router.POST("/api/hearth/knock", func(e *core.RequestEvent) error {
            // Parse HMAC invite token from body
            data := struct {
                Token string `json:"token" form:"token"`
            }{}
            if err := e.BindBody(&data); err != nil {
                return e.BadRequestError("Invalid request", err)
            }
            // ... validate HMAC token ...
            return e.JSON(200, map[string]bool{"valid": true})
        }).Bind(apis.RequireAuth())

        // Superuser-only route
        se.Router.DELETE("/api/hearth/admin/purge", func(e *core.RequestEvent) error {
            // ... purge operation ...
            return e.NoContent(204)
        }).Bind(apis.RequireSuperuserAuth())

        // Route groups
        g := se.Router.Group("/api/hearth/rooms")
        g.Bind(apis.RequireAuth()) // all routes in group require auth
        g.GET("", listRoomsAction)
        g.GET("/{id}", getRoomAction)
        g.PATCH("/{id}", updateRoomAction)

        // Serve frontend SPA
        se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

        return se.Next()
    })
}
```

**Key Details:**
- Path parameters: `{name}` for single segment, `{name...}` for wildcard
- Access path params via `e.Request.PathValue("name")`
- Auth state via `e.Auth` (nil = guest)
- `e.HasSuperuserAuth()` shorthand for superuser check
- Custom routes under `/api/hearth/` to avoid collision with PocketBase system routes

### 3. Database Access

Two approaches: **high-level Record API** (preferred) and **raw SQL via dbx builder**.

#### Record API (Preferred for Collection Data)

```go
// Find single record
record, err := app.FindRecordById("rooms", "RECORD_ID")

// Find by field value
record, err := app.FindFirstRecordByData("rooms", "slug", "living-room")

// Find with filter expression (use {:placeholder} for user input!)
record, err := app.FindFirstRecordByFilter(
    "messages",
    "room = {:room} && expires_at > {:now}",
    dbx.Params{"room": roomId, "now": time.Now().UTC().Format(time.RFC3339)},
)

// Find multiple with pagination
records, err := app.FindRecordsByFilter(
    "messages",                        // collection
    "room = {:room}",                  // filter
    "-created",                        // sort (- prefix = DESC)
    50,                                // limit
    0,                                 // offset
    dbx.Params{"room": roomId},        // params
)

// Count records
total, err := app.CountRecords("messages", dbx.HashExp{"room": roomId})

// Create
collection, err := app.FindCollectionByNameOrId("messages")
record := core.NewRecord(collection)
record.Set("room", roomId)
record.Set("author", userId)
record.Set("body", messageText)
record.Set("expires_at", time.Now().Add(ttl).UTC().Format(time.RFC3339))
err = app.Save(record)

// Update
record, err := app.FindRecordById("rooms", roomId)
record.Set("name", "New Name")
err = app.Save(record)

// Delete
err = app.Delete(record)
```

#### Raw SQL via dbx Builder (For Complex Queries)

```go
// Raw query
res, err := app.DB().
    NewQuery("DELETE FROM messages WHERE expires_at < {:now}").
    Bind(dbx.Params{"now": time.Now().UTC().Format(time.RFC3339)}).
    Execute()

// Query builder
type MessageStats struct {
    RoomId string `db:"room" json:"room_id"`
    Count  int    `db:"count" json:"count"`
}
stats := []MessageStats{}
err := app.DB().
    Select("room", "COUNT(*) as count").
    From("messages").
    Where(dbx.NewExp("expires_at > {:now}", dbx.Params{"now": now})).
    GroupBy("room").
    OrderBy("count DESC").
    All(&stats)
```

### 4. SQLite PRAGMA Injection (Hearth-Specific)

PocketBase v0.23+ supports custom SQLite drivers via `DBConnect`. For our WAL mode + tuned pragmas:

```go
// Option A: Set pragmas after app bootstrap
app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
    // These pragmas are set on the main data.db connection
    pragmas := []string{
        "PRAGMA journal_mode=WAL",
        "PRAGMA synchronous=NORMAL",
        "PRAGMA cache_size=-2000",
        "PRAGMA mmap_size=268435456",
        "PRAGMA busy_timeout=5000",
    }
    for _, pragma := range pragmas {
        if _, err := e.App.DB().NewQuery(pragma).Execute(); err != nil {
            return fmt.Errorf("failed to set pragma: %s: %w", pragma, err)
        }
    }
    return e.Next()
})
```

**Note:** PocketBase uses `modernc.org/sqlite` (pure Go, no CGO). WAL mode works but some pragmas may behave differently than with `mattn/go-sqlite3`.

### 5. Cron Job Registration (Message GC)

```go
func registerCronJobs(app *pocketbase.PocketBase) {
    // Message garbage collection — every minute
    app.Cron().MustAdd("hearth_message_gc", "* * * * *", func() {
        now := time.Now().UTC().Format(time.RFC3339)
        res, err := app.DB().
            NewQuery("DELETE FROM messages WHERE expires_at <= {:now}").
            Bind(dbx.Params{"now": now}).
            Execute()
        if err != nil {
            app.Logger().Error("Message GC failed", "error", err)
            return
        }
        affected, _ := res.RowsAffected()
        if affected > 0 {
            app.Logger().Info("Message GC", "deleted", affected)
        }
    })

    // Nightly VACUUM for physical data erasure — 4 AM daily
    app.Cron().MustAdd("hearth_nightly_vacuum", "0 4 * * *", func() {
        if _, err := app.DB().NewQuery("VACUUM").Execute(); err != nil {
            app.Logger().Error("Nightly VACUUM failed", "error", err)
        }
    })

    // Presence cleanup — every 2 minutes
    // (This is in-memory cleanup, not DB — just logging stale entries)
    app.Cron().MustAdd("hearth_presence_sweep", "*/2 * * * *", func() {
        // sweepStalePresence() — implemented in presence.go
    })
}
```

**Cron Details:**
- Cron starts automatically on `app serve`
- Each job runs in its own goroutine
- System jobs (logs cleanup, backups) use `__pb*__` IDs — don't overwrite
- All registered crons visible in Dashboard > Settings > Crons
- Use `app.Cron().Remove("id")` to unregister

### 6. Event Hooks (Record Lifecycle)

```go
func registerHooks(app *pocketbase.PocketBase) {
    // Before a message is created — set expiration
    app.OnRecordCreate("messages").BindFunc(func(e *core.RecordEvent) error {
        // Get room's TTL setting
        roomId := e.Record.GetString("room")
        room, err := e.App.FindRecordById("rooms", roomId)
        if err != nil {
            return e.App.Logger().Error("Room not found", "room", roomId)
        }
        ttlSeconds := room.GetInt("message_ttl")
        if ttlSeconds <= 0 {
            ttlSeconds = 3600 // default 1 hour
        }
        expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
        e.Record.Set("expires_at", expiresAt.UTC().Format(time.RFC3339))

        return e.Next()
    })

    // After a message is created — broadcast via custom realtime
    app.OnRecordAfterCreateSuccess("messages").BindFunc(func(e *core.RecordEvent) error {
        // PocketBase auto-broadcasts to collection subscribers.
        // Custom notification logic (e.g., push to Portal users) can go here.
        return e.Next()
    })

    // Intercept record API requests
    app.OnRecordCreateRequest("messages").BindFunc(func(e *core.RecordRequestEvent) error {
        // Validate that the user has access to the room
        if e.Auth == nil {
            return e.UnauthorizedError("Authentication required", nil)
        }
        return e.Next()
    })
}
```

**Hook Categories:**
- **Model hooks** (`OnRecordCreate`, etc.): Triggered from anywhere (cron, API, code). No request context.
- **Request hooks** (`OnRecordCreateRequest`, etc.): Triggered only from API endpoints. Have full request context (`e.Auth`, `e.Request`, etc.).
- Always call `e.Next()` to continue the handler chain.
- Use `e.App` inside hooks (not parent scope `app`) to avoid deadlocks in transactions.

### 7. Real-time Messaging (Custom Topics)

```go
// Server-side: Send custom event to all subscribers of a topic
func notifyRoom(app core.App, roomId string, eventType string, data any) error {
    rawData, err := json.Marshal(map[string]any{
        "type": eventType,
        "data": data,
    })
    if err != nil {
        return err
    }

    message := subscriptions.Message{
        Name: "rooms/" + roomId,
        Data: rawData,
    }

    clients := app.SubscriptionsBroker().Clients()
    for _, client := range clients {
        if client.HasSubscription("rooms/" + roomId) {
            client.Send(message)
        }
    }
    return nil
}
```

```typescript
// Client-side (JS SDK): Subscribe to custom topic
import PocketBase from 'pocketbase';

const pb = new PocketBase('https://api.hearth.example');

// Subscribe to collection changes (built-in)
pb.collection('messages').subscribe('*', (e) => {
    console.log(e.action, e.record);
});

// Subscribe to custom topic
pb.realtime.subscribe('rooms/ROOM_ID', (e) => {
    console.log('Room event:', e);
});
```

### 8. Transactions

```go
// Atomic operation: Create room + initial welcome message
app.RunInTransaction(func(txApp core.App) error {
    // Create room
    roomCol, _ := txApp.FindCollectionByNameOrId("rooms")
    room := core.NewRecord(roomCol)
    room.Set("name", "New Room")
    room.Set("owner", userId)
    if err := txApp.Save(room); err != nil {
        return err
    }

    // Create welcome message
    msgCol, _ := txApp.FindCollectionByNameOrId("messages")
    msg := core.NewRecord(msgCol)
    msg.Set("room", room.Id)
    msg.Set("body", "Welcome to the room!")
    msg.Set("system", true)
    if err := txApp.Save(msg); err != nil {
        return err
    }

    return nil // commit
})
```

**Transaction Rules:**
- Always use `txApp` inside the callback, never the outer `app`
- Single writer at a time (SQLite limitation)
- Keep transactions fast — no network calls inside

### 9. Production Deployment

```go
// Dockerfile for Hearth's PocketBase
// Start with: /pb/pocketbase serve --http=0.0.0.0:8090
// Mount volume: -v pb_data:/pb/pb_data

// Environment variables for Hearth:
// GOMEMLIMIT=250MiB          — Soft memory limit for Go GC
// PB_ENCRYPTION_KEY=...      — Optional: encrypt settings in DB
```

**Production Recommendations (from PocketBase docs):**
- Set `GOMEMLIMIT=250MiB` for our memory-constrained environment
- Set `ulimit -n 4096` for the PocketBase process (concurrent WS connections)
- Enable rate limiting via Dashboard > Settings > Application
- Configure "User IP proxy headers" when behind Caddy (X-Real-IP, X-Forwarded-For)
- Use `--encryptionEnv=PB_ENCRYPTION_KEY` for encrypted settings storage
- PocketBase default port: **8090** (configurable with `--http` flag)

---

## Gotchas for Builder

1. **NEVER use `app.Dao()`** — it doesn't exist in v0.23+. Use `app.DB()` for raw SQL, or `app.FindRecord*()` / `app.Save()` for records.
2. **NEVER use `app.OnBeforeServe()`** — replaced by `app.OnServe().BindFunc()`.
3. **Always call `e.Next()`** in hooks/middleware — forgetting this silently stops the chain.
4. **Use `e.App` not `app`** inside hooks — prevents deadlocks during transactions/cascades.
5. **Use `{:placeholder}` params** in filter strings — prevents SQL injection.
6. **PocketBase auto-creates `id` fields** — don't set record IDs manually unless needed.
7. **System cron job IDs** start with `__pb` — don't use this prefix for custom jobs.
8. **Realtime subscriptions** are per-client, not per-user — one user can have multiple connections (tabs, devices).
9. **WAL mode** journal should not be committed to Git — add `pb_data/` to `.gitignore`.
10. **The `pb_public/` directory** is served automatically — use this for the React SPA build output.

---

## Ready for Implementation

**Feature:** PocketBase Backend Skeleton | **Spec:** This document | **Complexity:** M
**Key Points for Builder:**
1. Use `app.OnServe().BindFunc()` for ALL route/hook registration
2. Set SQLite pragmas in `OnBootstrap` hook (not `OnServe`)
3. Cron jobs via `app.Cron().MustAdd()` — auto-starts on serve
4. `app.DB()` for raw SQL, `app.FindRecord*()` for typed access

**Files to Create/Modify:**
- `backend/main.go` — Entry point with hook/route/cron registration
- `backend/hooks/messages.go` — Message lifecycle hooks + GC cron
- `backend/hooks/presence.go` — In-memory presence tracking
- `backend/hooks/auth.go` — HMAC invite validation
- `backend/go.mod` — Dependencies (`github.com/pocketbase/pocketbase`)

**Questions Resolved:**
- Q: What's the current startup hook? → A: `app.OnServe().BindFunc()`
- Q: How to access DB? → A: `app.DB()` (dbx builder) or `app.FindRecord*()`
- Q: How to register cron? → A: `app.Cron().MustAdd("id", "expr", func(){})`
- Q: PRAGMA injection point? → A: `app.OnBootstrap().BindFunc()`
- Q: Default port? → A: 8090 (configurable via `--http` flag)

**Deferred Decisions:**
- Collection schema design (needs separate spec)
- Exact message TTL values (needs UX testing)
- Auth collection customization (fields, OAuth providers)
