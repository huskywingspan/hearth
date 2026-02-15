# Sprint 4: First Friend (v0.3)

**Date:** 2026-02-15 | **Target:** April 2026
**Implements:** ADR-007 (Dens/Campfires/DMs/Roles), R-012 (Remote Access), R-009 Gap #4 (Ghost Text)
**Depends on:** v0.2.1 Settling In (complete), R-012 Remote Access Architecture (complete)

---

## Goal

A friend outside the LAN can connect to your House and chat. The schema supports the full ADR-007 channel architecture (Dens, Campfires, DMs, Roles). The server URL is configurable in the frontend. A quick-test script lets you demo to a friend in one command.

**The "First Friend" moment:** You run `./scripts/quick-test.sh`, text the URL to your friend, they open it in a browser, create an account, and see The Den with your message waiting. Under 5 minutes.

---

## Non-Goals (This Sprint)

- Voice/video (v0.4 Hearth Fire)
- E2EE (v1.0 First Light)
- The Knock / guest entry system (v1.0)
- Landing page at hearthapp.chat (deferred — focus on core product)
- QR code generation/scanning (deferred — URL sharing is sufficient for first friend)
- PWA manifest + Service Worker (deferred to v1.0)
- VPS deployment guide (deferred — quick-test mode is the v0.3 path)

---

## Architecture

### What Changes

```
BACKEND (Go / PocketBase)
├── hooks/collections.go     MODIFY  — Add type, voice, video, history_visible to rooms
│                                       Add role, public_key to users
│                                       Create direct_messages + dm_messages collections
├── hooks/auth.go            MODIFY  — Seed "The Den" on first startup
│                                       First user becomes Homeowner
│                                       Denormalize author_name on dm_messages too
├── hooks/roles.go           CREATE  — Role enforcement middleware
├── hooks/dms.go             CREATE  — DM creation + lookup endpoints
├── hooks/hooks_test.go      MODIFY  — Tests for new functionality

FRONTEND (React / Vite / TypeScript)
├── src/lib/pocketbase.ts    MODIFY  — Dynamic server URL from localStorage
├── src/lib/constants.ts     MODIFY  — Add DM and role constants
├── src/pages/ConnectPage.tsx CREATE  — Enter server URL flow
├── src/pages/RoomPage.tsx   MODIFY  — Route to Den or Campfire view
├── src/pages/DmPage.tsx     CREATE  — DM conversation view
├── src/components/
│   ├── layout/
│   │   ├── Shell.tsx        MODIFY  — Sidebar: Dens section + Campfires section + DMs section
│   │   └── RoomList.tsx     MODIFY  — Split into DenList + CampfireList + DmList
│   ├── campfire/            (unchanged — already handles ephemeral chat)
│   ├── den/
│   │   └── DenRoom.tsx      CREATE  — Permanent chat view (no fade)
│   └── dm/
│       └── DmRoom.tsx       CREATE  — DM conversation component
├── src/hooks/
│   └── useDmMessages.ts     CREATE  — DM message subscription
├── src/styles/
│   └── campfire.css         MODIFY  — Ghost Text Echo stage: blur + gray shift
├── src/App.tsx              MODIFY  — Add ConnectPage + DM routes

SCRIPTS
├── scripts/quick-test.sh    CREATE  — One-liner CF Tunnel for demo
├── scripts/quick-test.ps1   CREATE  — Windows equivalent
```

---

## Phase 4.A — Schema Evolution (ADR-007)

**Priority:** Do this first. Everything else depends on the schema.

### FF-001: Add `type` field to rooms collection

Add `type` SelectField to the existing rooms collection using the incremental migration pattern (same as `author_name` in Sprint 3).

```go
// In ensureRoomsCollection — after the "already exists" early return, 
// add incremental field check:
if existing.Fields.GetByName("type") == nil {
    existing.Fields.Add(&core.SelectField{
        Name:      "type",
        Required:  true,
        Values:    []string{"den", "campfire"},
        MaxSelect: 1,
    })
    changed = true
}
```

**Migration note:** Existing rooms (created pre-v0.3) have no `type` field. When we add it with `Required: true`, PocketBase needs a default for existing records. Two options:
1. Set `Required: false` and backfill via a migration query after field creation
2. Run a raw SQL UPDATE to set existing rooms to `"campfire"` (they were campfires)

**Recommended:** Option 2. After adding the field:
```go
_, err := app.DB().NewQuery(`UPDATE rooms SET type = 'campfire' WHERE type = '' OR type IS NULL`).Execute()
```

### FF-002: Add `voice`, `video`, `history_visible` fields to rooms

Same incremental pattern. These are all optional/boolean/have defaults:

```go
// voice: boolean, default false (only meaningful for dens)
if existing.Fields.GetByName("voice") == nil {
    existing.Fields.Add(&core.BoolField{Name: "voice"})
    changed = true
}

// video: boolean, default false
if existing.Fields.GetByName("video") == nil {
    existing.Fields.Add(&core.BoolField{Name: "video"})
    changed = true
}

// history_visible: boolean, default true (new members can see old messages)
if existing.Fields.GetByName("history_visible") == nil {
    existing.Fields.Add(&core.BoolField{Name: "history_visible"})
    changed = true
}
```

**API rule update:** Campfire creation should be open; Den creation restricted:

```go
// In applyAPIRules — rooms.CreateRule update:
// Homeowner or Keyholder can create dens; anyone auth'd can create campfires
// PocketBase filter expression:
rooms.CreateRule = stringPtr(`@request.auth.id != "" && (
    @request.body.type = "campfire" ||
    @request.auth.role = "homeowner" ||
    @request.auth.role = "keyholder"
)`)
```

> **Builder note:** Test whether PocketBase supports `@request.body.type` in create rules. If not, enforce via `OnRecordCreate` hook instead.

### FF-003: Add `role` and `public_key` fields to users

Extend `ensureUsersFields`:

```go
// role: homeowner | keyholder | member (default: member)
if collection.Fields.GetByName("role") == nil {
    collection.Fields.Add(&core.SelectField{
        Name:      "role",
        Required:  true,
        Values:    []string{"homeowner", "keyholder", "member"},
        MaxSelect: 1,
    })
}

// public_key: empty for now, E2EE readiness (v1.0)
if collection.Fields.GetByName("public_key") == nil {
    collection.Fields.Add(&core.TextField{
        Name: "public_key",
        Max:  500,
    })
}
```

**Migration:** Existing users need `role = 'member'` backfill. Then find the first-created user and set them as `homeowner`:

```go
// Backfill existing users
_, _ = app.DB().NewQuery(`UPDATE users SET role = 'member' WHERE role = '' OR role IS NULL`).Execute()

// First user = homeowner (by created timestamp)
_, _ = app.DB().NewQuery(`UPDATE users SET role = 'homeowner' WHERE id = (SELECT id FROM users ORDER BY created ASC LIMIT 1) AND role != 'homeowner'`).Execute()
```

### FF-004: Create `direct_messages` + `dm_messages` collections

New function `ensureDirectMessagesCollection`:

```go
func ensureDirectMessagesCollection(app core.App) error {
    _, err := app.FindCollectionByNameOrId("direct_messages")
    if err == nil {
        return nil // already exists
    }

    usersCol, err := app.FindCollectionByNameOrId("users")
    if err != nil {
        return fmt.Errorf("users collection not found: %w", err)
    }

    collection := core.NewBaseCollection("direct_messages")

    // Participant A (always the lower-sorted user ID for uniqueness)
    collection.Fields.Add(&core.RelationField{
        Name:         "participant_a",
        Required:     true,
        CollectionId: usersCol.Id,
        MaxSelect:    1,
    })

    // Participant B
    collection.Fields.Add(&core.RelationField{
        Name:         "participant_b",
        Required:     true,
        CollectionId: usersCol.Id,
        MaxSelect:    1,
    })

    collection.Indexes = []string{
        "CREATE UNIQUE INDEX idx_dm_participants ON direct_messages (participant_a, participant_b)",
    }

    return app.Save(collection)
}
```

New function `ensureDmMessagesCollection`:

```go
func ensureDmMessagesCollection(app core.App) error {
    _, err := app.FindCollectionByNameOrId("dm_messages")
    if err == nil {
        return nil
    }

    dmCol, err := app.FindCollectionByNameOrId("direct_messages")
    if err != nil {
        return fmt.Errorf("direct_messages collection not found: %w", err)
    }
    usersCol, err := app.FindCollectionByNameOrId("users")
    if err != nil {
        return fmt.Errorf("users collection not found: %w", err)
    }

    collection := core.NewBaseCollection("dm_messages")

    collection.Fields.Add(&core.RelationField{
        Name:          "dm",
        Required:      true,
        CollectionId:  dmCol.Id,
        MaxSelect:     1,
        CascadeDelete: true,
    })

    collection.Fields.Add(&core.RelationField{
        Name:         "author",
        Required:     true,
        CollectionId: usersCol.Id,
        MaxSelect:    1,
    })

    collection.Fields.Add(&core.TextField{
        Name: "author_name",
        Max:  50,
    })

    collection.Fields.Add(&core.TextField{
        Name:     "body",
        Required: true,
        Min:      1,
        Max:      4000,
    })

    return app.Save(collection)
}
```

**DM API rules** (in `applyAPIRules`):

```go
// direct_messages: only participants can see/create
dms.ListRule = stringPtr(`participant_a = @request.auth.id || participant_b = @request.auth.id`)
dms.ViewRule = stringPtr(`participant_a = @request.auth.id || participant_b = @request.auth.id`)
dms.CreateRule = stringPtr(`@request.auth.id != ""`)
dms.DeleteRule = nil // DMs cannot be deleted (permanent)

// dm_messages: only DM participants can read/write
dmMsgs.ListRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
dmMsgs.ViewRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
dmMsgs.CreateRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
dmMsgs.UpdateRule = stringPtr(`author = @request.auth.id`)
dmMsgs.DeleteRule = stringPtr(`author = @request.auth.id`)
```

### FF-005: `public_key` on users — covered by FF-003 above (empty field, schema readiness only)

### FF-006: Seed default Den on first startup

In `auth.go`, add a startup hook that creates "The Den" if no dens exist:

```go
// After all collections are created (in RegisterCollections or a new RegisterSeeds)
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Check if any dens exist
    dens, _ := se.App.FindRecordsByFilter("rooms", "type = 'den'", "", 1, 0)
    if len(dens) > 0 {
        return se.Next() // already have dens
    }

    // Find the homeowner (first user)
    owners, _ := se.App.FindRecordsByFilter("users", "role = 'homeowner'", "", 1, 0)
    if len(owners) == 0 {
        return se.Next() // no users yet — seed on first user creation instead
    }

    roomsCol, err := se.App.FindCollectionByNameOrId("rooms")
    if err != nil {
        return se.Next()
    }

    den := core.NewRecord(roomsCol)
    den.Set("name", "The Den")
    den.Set("slug", "the-den")
    den.Set("type", "den")
    den.Set("owner", owners[0].Id)
    den.Set("description", "The main room. Pull up a chair.")
    den.Set("default_ttl", 0)  // dens don't expire messages
    den.Set("max_participants", 25)
    den.Set("livekit_room_name", fmt.Sprintf("hearth-the-den-%d", time.Now().UnixMilli()))
    den.Set("voice", false) // voice enabled in v0.4
    den.Set("video", false)
    den.Set("history_visible", true)

    if err := se.App.Save(den); err != nil {
        se.App.Logger().Error("failed to seed The Den", "error", err)
    } else {
        se.App.Logger().Info("seeded default Den: 'The Den'")
    }

    return se.Next()
})
```

**Important:** Den messages should NOT expire. The `default_ttl = 0` needs to be handled in the message creation hook — if `ttl <= 0`, don't set `expires_at` (or set it to a far-future date). This is a key behavioral difference between Dens and Campfires.

Update in `auth.go` message creation hook:
```go
ttlSeconds := room.GetInt("default_ttl")
if ttlSeconds > 0 {
    // Campfire: messages expire
    expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
    e.Record.Set("expires_at", expiresAt.UTC().Format(time.RFC3339))
} else {
    // Den: messages are permanent — set far-future expiry
    // (keeping the field non-null simplifies GC query)
    e.Record.Set("expires_at", "2099-12-31T23:59:59Z")
}
```

### FF-007: Role-based API rules

Create `hooks/roles.go`:

```go
package hooks

import (
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

// RegisterRoles adds role enforcement hooks.
func RegisterRoles(app *pocketbase.PocketBase) {
    // First user to register becomes Homeowner
    app.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
        // Count total users
        count, err := countRecords(e.App, "users")
        if err != nil || count > 1 {
            return e.Next() // not the first user
        }

        // This is the first user — crown them Homeowner
        e.Record.Set("role", "homeowner")
        if err := e.App.Save(e.Record); err != nil {
            e.App.Logger().Error("failed to set first user as homeowner", "error", err)
        } else {
            e.App.Logger().Info("first user crowned as Homeowner", "user", e.Record.Id)
        }

        return e.Next()
    })

    // Prevent non-admin room creation for dens
    app.OnRecordCreate("rooms").BindFunc(func(e *core.RecordEvent) error {
        roomType := e.Record.GetString("type")
        if roomType == "" {
            e.Record.Set("type", "campfire") // default to campfire
        }

        if roomType == "den" {
            // Only homeowners and keyholders can create dens
            authRecord := e.RequestInfo().Auth
            if authRecord == nil {
                return fmt.Errorf("authentication required")
            }
            role := authRecord.GetString("role")
            if role != "homeowner" && role != "keyholder" {
                return fmt.Errorf("only Homeowners and Keyholders can create Dens")
            }
        }

        return e.Next()
    })
}
```

> **Builder note:** Import `fmt` in roles.go. Test whether PocketBase's `e.RequestInfo().Auth` reliably returns the full user record with custom fields. If it only returns base auth fields, fetch the user record: `user, _ := e.App.FindRecordById("users", authRecord.Id)`.

---

## Phase 4.B — Ghost Text Enhancement (FF-008)

From R-009 gap analysis finding #4: the Echo stage should add blur and gray shift, not just transparency.

Update `campfire.css`:

```css
@keyframes campfire-fade {
  0%   { opacity: 1;   filter: blur(0px); color: inherit; }       /* Fresh */
  40%  { opacity: 0.5; filter: blur(0px); color: inherit; }       /* Fading */
  80%  { opacity: 0.1; filter: blur(1px); color: var(--color-text-muted); }  /* Echo — blur + gray */
  100% { opacity: 0;   filter: blur(2px); color: var(--color-text-muted); }  /* Gone */
}
```

**Performance note:** `filter: blur()` runs on the compositor thread (GPU-accelerated) same as `opacity`. Confirmed safe in R-008. The blur value is subtle (1-2px) — it shouldn't trigger excessive GPU memory. `color` is NOT a compositor property (causes repaint), but at 80-100% of the animation, almost no messages are at that stage simultaneously.

---

## Phase 4.C — Dynamic Server URL (FF-012)

### Frontend: ConnectPage

The PocketBase URL is currently hardcoded in `pocketbase.ts`. Change it to read from `localStorage`:

**`src/lib/pocketbase.ts`** (updated):
```typescript
import PocketBase from 'pocketbase';

function getServerUrl(): string {
  // Priority: 1) localStorage (user-selected), 2) env var, 3) same-origin
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('hearth_server_url');
    if (saved) return saved;
  }
  return import.meta.env.VITE_PB_URL || '/';
}

const pb = new PocketBase(getServerUrl());

/** Update the PocketBase server URL and reload the page. */
export function setServerUrl(url: string) {
  // Normalize: strip trailing slash, ensure https:// or http://
  let normalized = url.trim().replace(/\/+$/, '');
  if (!/^https?:\/\//i.test(normalized)) {
    normalized = 'https://' + normalized;
  }
  localStorage.setItem('hearth_server_url', normalized);
  // Clear auth (different server = different identity)
  pb.authStore.clear();
  window.location.reload();
}

/** Clear the saved server URL (return to default). */
export function clearServerUrl() {
  localStorage.removeItem('hearth_server_url');
  pb.authStore.clear();
  window.location.reload();
}

/** Get the current server URL for display. */
export function getCurrentServerUrl(): string {
  return localStorage.getItem('hearth_server_url') || '(local)';
}

export default pb;
```

**`src/pages/ConnectPage.tsx`** (new):
```tsx
import { useState } from 'react';
import { setServerUrl } from '@/lib/pocketbase';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card } from '@/components/ui/Card';

export default function ConnectPage() {
  const [url, setUrl] = useState('');
  const [error, setError] = useState('');
  const [checking, setChecking] = useState(false);

  const handleConnect = async () => {
    if (!url.trim()) return;
    setChecking(true);
    setError('');

    try {
      // Ping the server's health endpoint to verify it's a PocketBase instance
      let target = url.trim().replace(/\/+$/, '');
      if (!/^https?:\/\//i.test(target)) target = 'https://' + target;
      
      const res = await fetch(`${target}/api/health`, { 
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      });
      
      if (!res.ok) throw new Error('Not a Hearth server');
      
      setServerUrl(target); // This reloads the page
    } catch {
      setError('Could not reach that server. Check the URL and try again.');
      setChecking(false);
    }
  };

  return (
    <div className="flex items-center justify-center h-screen bg-[var(--color-bg-primary)] p-4">
      <Card className="max-w-md w-full animate-float-in">
        <div className="text-center mb-6">
          <h1 className="font-display text-3xl text-[var(--color-accent-amber)] mb-2">
            Hearth
          </h1>
          <p className="text-[var(--color-text-secondary)]">
            Enter your House address to connect.
          </p>
        </div>

        <div className="space-y-4">
          <Input
            type="url"
            placeholder="https://hearth.example.com"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleConnect()}
            disabled={checking}
            autoFocus
          />

          {error && (
            <p className="text-sm text-[var(--color-alert-clay)]">{error}</p>
          )}

          <Button onClick={handleConnect} disabled={checking} className="w-full">
            {checking ? 'Connecting...' : 'Enter House'}
          </Button>
        </div>
      </Card>
    </div>
  );
}
```

**`src/App.tsx`** (updated routes):
```tsx
const ConnectPage = lazy(() => import('@/pages/ConnectPage'));
const DmPage = lazy(() => import('@/pages/DmPage'));

// New route structure:
<Routes>
  <Route path="/connect" element={<ConnectPage />} />
  <Route path="/login" element={<LoginPage />} />
  <Route element={<AuthGuard />}>
    <Route path="/" element={<HomePage />} />
    <Route path="/room/:roomId" element={<RoomPage />} />
    <Route path="/dm/:dmId" element={<DmPage />} />
  </Route>
</Routes>
```

**Flow:** If no `hearth_server_url` is in localStorage AND no `VITE_PB_URL` env var is set AND the app isn't served from PocketBase (same-origin), redirect to `/connect`. Otherwise, skip directly to login/home.

> **Builder note:** The redirect logic should live in a `useServerUrl()` hook or in `App.tsx`. If `pb.baseURL` is `/` and the fetch to `/api/health` fails, redirect to `/connect`. This handles the case where the frontend is opened standalone (e.g., `npx vite` during dev) without a backend running.

---

## Phase 4.D — Sidebar Refactor (FF-030 through FF-033)

### Updated Sidebar Structure

The `RoomList` component becomes a full navigation sidebar with three sections:

```
┌─────────────────────┐
│ 🔥 Hearth           │  ← House name (future: configurable)
├─────────────────────┤
│ DENS                │  ← Permanent rooms
│   The Den           │
│   Game Night        │
│   + New Den ★       │  ← Only visible to Homeowner/Keyholder
├─────────────────────┤
│ CAMPFIRES           │  ← Ephemeral, fading rooms
│   Late Night Vibes  │
│   + New Campfire    │  ← Anyone can create (configurable)
├─────────────────────┤
│ MESSAGES            │  ← DM conversations
│   Alice             │
│   Bob               │
│   + New Message     │
├─────────────────────┤
│ 🟢 Username         │
│ Sign out │ ⚙️       │  ← Settings gear (future: server URL, profile)
└─────────────────────┘
```

**Implementation approach:** Keep `RoomList.tsx` but refactor to fetch rooms and split by `type`:

```tsx
const dens = rooms.filter(r => r.type === 'den');
const campfires = rooms.filter(r => r.type === 'campfire');
```

Den creation only shows the "+" button if `user.role === 'homeowner' || user.role === 'keyholder'`.

### DM List

Fetch `direct_messages` where current user is `participant_a` or `participant_b`. Display the OTHER participant's name. Use PocketBase `expand` to resolve participant names.

### "New Message" flow

Click "+" → show a user search/select → find-or-create the `direct_messages` record → navigate to `/dm/:dmId`.

**DM creation logic** (canonical ordering for uniqueness):

```typescript
async function findOrCreateDm(otherUserId: string): Promise<string> {
  const myId = pb.authStore.record?.id;
  // Sort IDs to ensure canonical pair (matches unique index)
  const [a, b] = [myId, otherUserId].sort();
  
  // Try to find existing
  try {
    const existing = await pb.collection('direct_messages').getFirstListItem(
      `participant_a = "${a}" && participant_b = "${b}"`
    );
    return existing.id;
  } catch {
    // Create new
    const dm = await pb.collection('direct_messages').create({
      participant_a: a,
      participant_b: b,
    });
    return dm.id;
  }
}
```

---

## Phase 4.E — Den View (FF-030)

### DenRoom Component

Dens display permanent messages — NO fade animation. The component is similar to `CampfireRoom` but without the decay CSS:

```tsx
// src/components/den/DenRoom.tsx
// Key differences from CampfireRoom:
// 1. No campfire-fade CSS class on messages
// 2. No animationend cleanup (messages persist)
// 3. Load more history (pagination, not just recent)
// 4. Optional: "New messages since you were away" divider
```

**Message display:** Use the same `MessageList` component but pass a `persistent={true}` prop that disables the fade animation. Or create a `DenMessageList` that renders messages without the `campfire-message` class.

### RoomPage Router

`RoomPage.tsx` should fetch the room, check `room.type`, and render either `<CampfireRoom>` or `<DenRoom>`:

```tsx
export default function RoomPage() {
  const { roomId } = useParams();
  const [room, setRoom] = useState<Room | null>(null);

  // ... fetch room ...

  if (room.type === 'den') {
    return <Shell><DenRoom roomId={roomId} roomName={room.name} /></Shell>;
  }
  return <Shell><CampfireRoom roomId={roomId} roomName={room.name} /></Shell>;
}
```

---

## Phase 4.F — DM View (FF-031)

### DmPage + DmRoom

```tsx
// src/pages/DmPage.tsx
export default function DmPage() {
  const { dmId } = useParams();
  // Fetch the DM record, get other participant's name
  // Render DmRoom
}
```

`DmRoom` is similar to `DenRoom` (permanent messages, no fade) but:
- No presence bar (it's 1:1 — presence is shown in DM list)
- Header shows the other user's display name
- Messages from `dm_messages` collection (not `messages`)

### useDmMessages Hook

Similar to `useMessages` but subscribes to `dm_messages` collection filtered by `dm = :dmId`:

```typescript
// Subscribe to real-time DM messages
pb.collection('dm_messages').subscribe('*', (e) => {
  if (e.record.dm !== dmId) return; // filter client-side
  // handle create/update/delete
});
```

---

## Phase 4.G — Quick-Test Script (FF-011)

### `scripts/quick-test.sh`

```bash
#!/bin/bash
# Hearth Quick Test — expose your local PocketBase to a friend via Cloudflare Tunnel
# Usage: ./scripts/quick-test.sh
# Requires: cloudflared (https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/)

set -e

PORT="${HEARTH_PORT:-8090}"

echo ""
echo "  🔥 Hearth Quick Test Mode"
echo "  ========================="
echo ""
echo "  Starting Cloudflare Tunnel to localhost:${PORT}..."
echo "  Your friend will get a URL to connect."
echo "  Press Ctrl+C to stop."
echo ""

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "  ❌ cloudflared not found."
    echo ""
    echo "  Install it:"
    echo "    macOS:   brew install cloudflared"
    echo "    Linux:   curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared && chmod +x /usr/local/bin/cloudflared"
    echo "    Windows: winget install Cloudflare.cloudflared"
    echo ""
    exit 1
fi

# Check if PocketBase is running
if ! curl -s "http://localhost:${PORT}/api/health" > /dev/null 2>&1; then
    echo "  ⚠️  PocketBase doesn't seem to be running on localhost:${PORT}"
    echo "  Start it first: cd backend && go run . serve --http=0.0.0.0:${PORT}"
    echo ""
    echo "  Continuing anyway (Cloudflare Tunnel will wait for it)..."
    echo ""
fi

cloudflared tunnel --url "http://localhost:${PORT}"
```

### `scripts/quick-test.ps1`

```powershell
# Hearth Quick Test — expose your local PocketBase to a friend via Cloudflare Tunnel
# Usage: .\scripts\quick-test.ps1

$Port = if ($env:HEARTH_PORT) { $env:HEARTH_PORT } else { "8090" }

Write-Host ""
Write-Host "  🔥 Hearth Quick Test Mode" -ForegroundColor Yellow
Write-Host "  ========================="
Write-Host ""
Write-Host "  Starting Cloudflare Tunnel to localhost:$Port..."
Write-Host "  Your friend will get a URL to connect."
Write-Host "  Press Ctrl+C to stop."
Write-Host ""

# Check if cloudflared is installed
if (-not (Get-Command "cloudflared" -ErrorAction SilentlyContinue)) {
    Write-Host "  ❌ cloudflared not found." -ForegroundColor Red
    Write-Host ""
    Write-Host "  Install it: winget install Cloudflare.cloudflared"
    Write-Host ""
    exit 1
}

cloudflared tunnel --url "http://localhost:$Port"
```

---

## Phase 4.H — Role Badges (FF-032)

Simple visual indicators in the sidebar and message bubbles:

```tsx
function RoleBadge({ role }: { role: string }) {
  if (role === 'homeowner') return <span title="Homeowner">🏠</span>;
  if (role === 'keyholder') return <span title="Keyholder">🔑</span>;
  return null; // members get no badge
}
```

Show in:
- Sidebar user section (own role)
- Presence bar (each user's role)
- Message bubbles (author role — requires denormalizing role onto messages or expanding author)

> **Builder decision:** Denormalize `author_role` onto messages (like `author_name`), or use expand? For v0.3, expand is fine (small user counts). Denormalize in v1.0 if needed.

---

## Task Summary

| ID | Task | Type | Phase | Priority | Notes |
|----|------|------|-------|----------|-------|
| FF-001 | Add `type` field to rooms + backfill | Build | 4.A | P0 | |
| FF-002 | Add `voice`, `video`, `history_visible` to rooms | Build | 4.A | P0 | |
| FF-003 | Add `role`, `public_key` to users + backfill | Build | 4.A | P0 | First user = homeowner |
| FF-004 | Create `direct_messages` + `dm_messages` collections | Build | 4.A | P0 | |
| FF-005 | `public_key` on users | Build | 4.A | P0 | Covered by FF-003 |
| FF-006 | Seed "The Den" on first startup | Build | 4.A | P1 | |
| FF-007 | Role enforcement (den creation, room rules) | Build | 4.A | P1 | `hooks/roles.go` |
| FF-008 | Ghost Text: blur + gray at Echo stage | Build | 4.B | P2 | CSS only |
| FF-012 | Dynamic server URL (ConnectPage) | Build | 4.C | P0 | |
| FF-011 | Quick-test scripts (bash + ps1) | Build | 4.G | P1 | |
| FF-030 | Sidebar: split Dens / Campfires / DMs | Build | 4.D | P0 | |
| FF-030b | DenRoom component (permanent chat) | Build | 4.E | P0 | |
| FF-031 | DM UI (list + send) | Build | 4.F | P1 | |
| FF-032 | Role badges in UI | Build | 4.H | P2 | |
| FF-033 | Fade time slider for campfire creators | Build | — | P3 | Deferred to v1.0 |
| T-001 | Tests: schema migration (type backfill) | Test | 4.A | P0 | |
| T-002 | Tests: role enforcement (den creation, homeowner) | Test | 4.A | P1 | |
| T-003 | Tests: DM creation + canonical ordering | Test | 4.F | P1 | |
| T-004 | Tests: Den message permanence (TTL=0) | Test | 4.E | P1 | |

---

## Build Order

```
Phase 4.A  Schema (backend)          ← do first, everything depends on it
  ├── FF-001  rooms.type + backfill
  ├── FF-002  rooms.voice/video/history_visible
  ├── FF-003  users.role/public_key + backfill
  ├── FF-004  direct_messages + dm_messages
  ├── FF-006  Seed "The Den"
  ├── FF-007  Role enforcement hooks
  └── T-001/T-002  Tests

Phase 4.B  Ghost Text (frontend CSS)   ← tiny, independent
  └── FF-008  campfire.css blur+gray

Phase 4.C  Connect flow (frontend)     ← independent of schema
  └── FF-012  Dynamic server URL + ConnectPage

Phase 4.D  Sidebar refactor (frontend) ← needs schema
  └── FF-030  Split Den/Campfire/DM list

Phase 4.E  Den view (frontend)         ← needs schema + sidebar
  └── FF-030b DenRoom component + RoomPage router

Phase 4.F  DM view (frontend)          ← needs DM schema + sidebar
  ├── FF-031  DmRoom + DmPage + useDmMessages
  └── T-003   DM tests

Phase 4.G  Quick-test scripts          ← independent
  └── FF-011  bash + ps1 scripts

Phase 4.H  Polish                      ← last
  └── FF-032  Role badges
```

---

## Acceptance Criteria

1. **Schema:** `rooms` has `type`, `voice`, `video`, `history_visible` fields. `users` has `role`, `public_key` fields. `direct_messages` and `dm_messages` collections exist with correct API rules.
2. **Migration:** Existing rooms are backfilled as `type: "campfire"`. Existing users are backfilled as `role: "member"`. First user is `homeowner`.
3. **Dens:** "The Den" is auto-created on first startup. Den messages don't expire. New dens can only be created by Homeowner/Keyholder.
4. **DMs:** Users can start a 1:1 DM conversation. Messages are permanent. Canonical participant ordering prevents duplicate conversations.
5. **Connect flow:** User can enter a server URL on `/connect`. URL is validated via `/api/health` ping. Stored in localStorage. Page reloads with new PocketBase endpoint.
6. **Sidebar:** Rooms are visually separated into Dens, Campfires, and Messages sections.
7. **Ghost Text:** Echo stage (80% of fade) shows 1px blur and gray text color.
8. **Quick-test:** `./scripts/quick-test.sh` starts a Cloudflare Tunnel. Output shows a `*.trycloudflare.com` URL. Friend can open it and access the Hearth frontend.
9. **All existing tests still pass.** New tests cover schema migration, role enforcement, DM creation, and den message permanence.
10. **Build:** `tsc --noEmit` clean. Vite build under 150KB gzipped.

---

## Gotchas for Builder

1. **PocketBase `Required: true` + backfill race:** If you add a required SelectField (`type`, `role`) and existing records don't have it, PocketBase may reject the save. Add the field as NOT required first, backfill, then change to required. OR do the backfill in the same `OnServe` hook immediately after field addition.

2. **Den TTL = 0 and message GC:** The message GC cron (`message_gc.go`) deletes messages where `expires_at < NOW()`. Den messages have `expires_at = 2099-12-31`. Make sure the GC query doesn't accidentally catch them. Current query uses `expires_at <= :now` — this is safe since 2099 > now.

3. **DM participant ordering:** Always sort the two user IDs alphabetically before creating a `direct_messages` record. This ensures the unique index on `(participant_a, participant_b)` prevents duplicate conversations regardless of who initiates.

4. **Room creation: TypeScript update.** The frontend `handleCreateRoom` in `RoomList.tsx` currently doesn't send a `type` field. Update it to send `type: "campfire"` for the campfire creation button and `type: "den"` for the den creation button.

5. **SSE subscription for DMs:** PocketBase SSE subscriptions are per-collection. The `dm_messages` subscription will fire for ALL dm_messages across all conversations. Filter client-side by `dm` field, OR use PocketBase's topic-based subscription: `pb.collection('dm_messages').subscribe(dmId, ...)` — test if PocketBase supports record-level subscriptions on non-ID fields; if not, filter client-side.

6. **Auth.go message hook:** The `OnRecordCreate("messages")` hook sets `expires_at` and `author_name`. This needs to behave differently for dens (TTL=0 → far-future expiry). The hook already reads `default_ttl` from the room — just update the conditional.

7. **CORS for CF Tunnel:** The existing CORS hook should allow `*.trycloudflare.com` origins. Current implementation may use a whitelist — update to use `*` for dev or add a wildcard pattern.
