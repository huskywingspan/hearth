# Sprint 3 — v0.2.1 "Settling In" (Integration Fixes + Access Model Simplification)

> **Sprint Goal:** Fix all integration bugs discovered during first real two-device testing. Simplify the access model so authenticated users can see and join rooms without the invite flow (appropriate for a small self-hosted "living room"). Denormalize author display names onto messages to eliminate the "Wanderer" bug. Result: two people on a LAN can create rooms, exchange messages, and see each other's names — reliably.
>
> **Target:** Immediate (patch release, not a milestone)
> **Owner:** Builder Agent
> **Predecessor:** v0.2 Kindling (Sprint 2)
> **Context:** First hands-on LAN testing revealed that the frontend was built against an *imagined* API contract. This sprint resolves every mismatch found during the codebase audit of 2026-02-11.

---

## Why This Sprint Exists

Sprint 1 (backend) and Sprint 2 (frontend) were built by the Builder agent in isolation — the frontend was never integration-tested against the running backend. The Researcher's full codebase audit on 2026-02-11 catalogued these issues:

| Bug ID | Symptom | Root Cause |
|--------|---------|------------|
| BUG-008 | Login fails | Zero users after pb_data regeneration (not a code bug) |
| BUG-009 | Login succeeds but nothing happens | No `navigate()` after auth |
| BUG-010 | Messages fade instantly / appear empty | Frontend used `text` field, backend uses `body` |
| BUG-011 | Every message duplicated | Optimistic send + realtime subscription race |
| BUG-012 | "Failed to create record" on new campfire | Frontend creates `room_members` but backend hook already does → unique constraint violation |
| BUG-013 | Presence auto-join race condition | Heartbeat + poll both try to auto-join simultaneously → duplicate key error |
| BUG-014 | "Wanderer" instead of real display name | Realtime subscription records don't include `expand` data |
| BUG-015 | Room page 404 for non-members | `rooms.ViewRule` requires membership, but membership doesn't exist yet on first visit |
| BUG-016 | Message send fails for non-members | `messages.CreateRule` requires membership; race with auto-join |

Additionally, the **access model has a fundamental tension**: the master plan describes a rich Knock/invite system (v1.0 feature), but the current implementation has a half-baked auto-join hack in the presence endpoints. This sprint resolves that by cleanly separating "who can see rooms" from "who can participate."

---

## Design Decisions (ADR-006: Simplified Access Model for Pre-v1.0)

**Status:** Proposed → to be accepted by user before Builder starts

### Context

The current API rules require `room_members` for everything:
- **List/View rooms:** Must be a member
- **List/View/Create messages:** Must be a member of the room
- **Presence heartbeat/poll:** Must be a member (with auto-join hack)

This creates a chicken-and-egg problem: you can't see a room until you're a member, but you can't become a member without an invite flow that doesn't exist yet.

### Decision

For pre-v1.0, simplify to:

| Action | Rule | Rationale |
|--------|------|-----------|
| **List rooms** | Any authenticated user | Small self-hosted instance = few rooms. Everyone can see what's available. |
| **View room** | Any authenticated user | Need to see the room to decide to join |
| **Join room (create membership)** | Any authenticated user | Walking into a room = joining. The Knock system (v1.0) will gate this later. |
| **Create room** | Any authenticated user | Anyone can start a campfire |
| **Create/read messages** | Must be a room member | Joining the room creates membership; then you can chat |
| **Update/delete room** | Room owner only | Unchanged |
| **Update/delete own messages** | Message author only | Unchanged |

### Implementation

1. Relax `rooms.ListRule` and `rooms.ViewRule` → `@request.auth.id != ""`
2. Relax `room_members.CreateRule` → `@request.auth.id != ""` (any authed user can join any room)
3. Keep `messages` rules requiring membership (unchanged)
4. Remove auto-join hacks from presence endpoints — membership happens explicitly on room entry
5. Backend hook in `auth.go` still auto-creates owner membership on room creation

### Revert Path

When the Knock system lands in v1.0, re-tighten:
- `rooms.ListRule` → add visibility rules (public/private rooms)
- `room_members.CreateRule` → require invite redemption or owner approval
- Add `room_members.visibility` field for public vs. invite-only rooms

---

## Task Breakdown

### Phase 3.A — Backend Fixes (Go)

| ID | Task | Bug(s) | Effort |
|----|------|--------|--------|
| S3-001 | Relax API rules: rooms visible to all authed users | BUG-015 | S |
| S3-002 | Relax `room_members.CreateRule` to allow self-join | BUG-016, BUG-013 | S |
| S3-003 | Remove auto-join code from presence heartbeat + poll endpoints | BUG-013 | S |
| S3-004 | Denormalize `author_name` onto messages (server-side hook) | BUG-014 | M |
| S3-005 | Remove `html.EscapeString` import from `sanitize.go` if still present | — | S |
| S3-006 | Rebuild `hearth.exe` and verify all tests pass | — | S |

### Phase 3.B — Frontend Fixes (React/TypeScript)

| ID | Task | Bug(s) | Effort |
|----|------|--------|--------|
| S3-010 | Remove `room_members.create()` from `RoomList.tsx` `handleCreateRoom` | BUG-012 | S |
| S3-011 | Add "join room" flow: on room entry, create `room_members` record if not exists | BUG-015, BUG-016 | M |
| S3-012 | Update `RoomList.tsx` to list all rooms (not just rooms user is a member of) | BUG-015 | S |
| S3-013 | Update `useMessages.ts` to use `author_name` from message record directly | BUG-014 | S |
| S3-014 | Update `MessageBubble.tsx` to read `message.author_name` instead of expand | BUG-014 | S |
| S3-015 | Verify realtime dedup still works correctly after all changes | BUG-011 | S |
| S3-016 | Build frontend, deploy to `pb_public/` | — | S |

### Phase 3.C — Verification & Cleanup

| ID | Task | Effort |
|----|------|--------|
| S3-020 | Two-device LAN test: create room, exchange messages, verify presence + names | M |
| S3-021 | Git commit all Sprint 3 changes with descriptive message | S |
| S3-022 | Test: create room → room appears in sidebar for both users | S |
| S3-023 | Test: send message from each user → both see correct display names | S |
| S3-024 | Test: presence bar shows correct count (1 person, then 2 people) | S |

---

## Detailed Implementation Guide

### S3-001: Relax Room API Rules

**File:** `backend/hooks/collections.go` → `applyAPIRules()` function

```go
// BEFORE (requires membership):
rooms.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room_members_via_room.user`)
rooms.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room_members_via_room.user`)

// AFTER (any authenticated user):
rooms.ListRule = stringPtr(`@request.auth.id != ""`)
rooms.ViewRule = stringPtr(`@request.auth.id != ""`)
```

Leave `CreateRule`, `UpdateRule`, `DeleteRule` unchanged.

---

### S3-002: Relax room_members CreateRule

**File:** `backend/hooks/collections.go` → `applyAPIRules()` function

```go
// BEFORE (only room owner can add members):
members.CreateRule = stringPtr(`@request.auth.id != "" && @request.auth.id = room.owner`)

// AFTER (any authenticated user can join any room):
members.CreateRule = stringPtr(`@request.auth.id != ""`)
```

This allows the frontend to create a `room_members` record when the user enters a room.

---

### S3-003: Remove Auto-Join from Presence Endpoints

**File:** `backend/hooks/presence.go`

Revert the heartbeat and GET endpoints to their original "check membership, 403 if not found" behavior. The frontend will handle joining before starting heartbeat/poll.

**Heartbeat handler:**
```go
// Verify user is a member of the room
_, err := e.App.FindFirstRecordByFilter(
    "room_members",
    "room = {:room} && user = {:user}",
    dbxParams("room", data.RoomID, "user", info.Auth.Id),
)
if err != nil {
    return e.ForbiddenError("Not a member of this room", nil)
}
```

Same pattern for the GET handler.

---

### S3-004: Denormalize `author_name` onto Messages

**Concept:** When a message is created, the server-side hook reads the author's `display_name` and writes it directly onto the message record as `author_name`. This means realtime subscription records carry the name without needing `expand`.

**File 1:** `backend/hooks/collections.go` → `ensureMessagesCollection()`

Add a new field to the messages collection:

```go
collection.Fields.Add(&core.TextField{
    Name: "author_name",
    Max:  50,
})
```

Note: NOT `Required: true` — field is set by the server hook, not the client.

**File 2:** `backend/hooks/auth.go` → `OnRecordCreate("messages")` hook

After the existing TTL logic, add:

```go
// Denormalize author display name onto the message
authorID := e.Record.GetString("author")
if authorID != "" {
    author, err := e.App.FindRecordById("users", authorID)
    if err == nil {
        e.Record.Set("author_name", author.GetString("display_name"))
    }
}
```

**Frontend:** `MessageBubble.tsx` reads `message.author_name` directly. Fallback to expand, then to "Wanderer".

---

### S3-010: Remove Duplicate room_members Creation in Frontend

**File:** `frontend/src/components/layout/RoomList.tsx`

In `handleCreateRoom`, remove the `room_members.create()` call. The backend `OnRecordAfterCreateSuccess("rooms")` hook already creates the owner membership.

```tsx
// BEFORE:
const room = await pb.collection('rooms').create({...});
await pb.collection('room_members').create({
  room: room.id,
  user: pb.authStore.record?.id,
  role: 'owner',
});

// AFTER:
const room = await pb.collection('rooms').create({...});
// Backend auto-creates owner membership via OnRecordAfterCreateSuccess hook
```

---

### S3-011: Add Join-on-Entry Flow

**File:** `frontend/src/hooks/useMessages.ts` (or a new `useRoomMembership.ts` hook)

Before subscribing to messages or sending heartbeats, ensure the user is a member of the room:

```typescript
// Ensure membership exists (idempotent — catches unique constraint on duplicate)
async function ensureMembership(roomId: string) {
  const userId = pb.authStore.record?.id;
  if (!userId) return;

  try {
    // Check if already a member
    await pb.collection('room_members').getFirstListItem(
      `room = "${roomId}" && user = "${userId}"`
    );
  } catch {
    // Not a member — join
    try {
      await pb.collection('room_members').create({
        room: roomId,
        user: userId,
        role: 'member',
      });
    } catch {
      // Unique constraint race — another tab/request already joined. That's fine.
    }
  }
}
```

Call this in `CampfireRoom.tsx` (or `RoomPage.tsx`) on mount, before initializing messages and presence.

---

### S3-012: List All Rooms in Sidebar

**File:** `frontend/src/components/layout/RoomList.tsx`

Replace the `room_members` query with a direct rooms query:

```typescript
// BEFORE: Query via room_members
const memberships = await pb.collection('room_members').getFullList({
  filter: `user = "${userId}"`,
  expand: 'room',
});
const roomList = memberships.map(m => m.expand?.room).filter(Boolean);

// AFTER: List all rooms directly
const roomList = await pb.collection('rooms').getFullList<Room>({
  sort: 'name',
});
```

This works because we relaxed `rooms.ListRule` to allow any authenticated user.

---

### S3-013 & S3-014: Use `author_name` from Message Record

**File:** `frontend/src/hooks/useMessages.ts` — Update `Message` interface:

```typescript
export interface Message {
  id: string;
  body: string;
  room: string;
  author: string;
  author_name: string;  // NEW — denormalized from author record
  expires_at: string;
  created: string;
  expand?: {
    author?: { display_name: string; avatar_url: string };
  };
}
```

**File:** `frontend/src/components/campfire/MessageBubble.tsx`:

```tsx
// BEFORE:
const authorName = message.expand?.author?.display_name ?? 'Wanderer';

// AFTER:
const authorName = message.author_name || message.expand?.author?.display_name || 'Wanderer';
```

The optimistic send in `useMessages.ts` should also include `author_name`:

```typescript
const optimistic: Message = {
  id: tempId,
  body: text,
  room: roomId,
  author: pb.authStore.record?.id ?? '',
  author_name: (pb.authStore.record as Record<string, unknown>)?.display_name as string ?? '',
  expires_at: new Date(Date.now() + 300_000).toISOString(),
  created: new Date().toISOString(),
};
```

---

## Fixes Already Applied (Pre-Sprint, During Debugging)

These were hotfixed during the 2026-02-11 debugging session. Builder should verify they're still correct after Sprint 3 changes:

| Fix | File | Description |
|-----|------|-------------|
| BUG-009 | `LoginPage.tsx`, `LoginForm.tsx` | Added `navigate('/')` after successful auth |
| BUG-010 | `useMessages.ts`, `MessageBubble.tsx` | Changed `text` → `body` field name |
| BUG-011 | `useMessages.ts` | Added dedup check in realtime subscription handler |
| `html.EscapeString` removal | `sanitize.go` | Removed server-side HTML escaping (React handles it) |
| Presence key fix | `usePresence.ts` | Changed `data.users` → `data.online` |

---

## Acceptance Criteria

1. **Room creation works:** Click "+ New campfire" → room appears in sidebar → no error
2. **Room visible to all users:** Second user sees the room in their sidebar without any invite
3. **Room entry auto-joins:** Navigating to a room URL creates a membership record silently
4. **Messages show real names:** Display name appears on every message (not "Wanderer")
5. **Presence count is accurate:** Shows correct number of people in the room
6. **Two-device chat works:** User A and User B on LAN can exchange messages in real-time
7. **No console errors:** Clean browser console (no 403s, no failed requests)
8. **Backend tests pass:** `go test ./hooks/...` — all existing tests still green
9. **Frontend builds clean:** `tsc --noEmit` + `npm run build` — no errors

---

## What This Sprint Does NOT Do

- Does not implement the Knock/invite system (v1.0)
- Does not add sound effects or ambient audio (deferred)
- Does not add mobile drawer navigation (K-023 partial)
- Does not add typing indicator broadcast (needs backend topic)
- Does not implement error toast system
- Does not add the `author_name` field to existing messages in the database (they'll show "Wanderer" until they fade and new ones are sent)

---

## Ready for Implementation

**Feature:** Sprint 3 — Settling In | **Spec:** `docs/specs/sprint-3-settling-in.md` | **Complexity:** M

**Key Points for Builder:**
1. The backend `OnRecordAfterCreateSuccess("rooms")` hook in `auth.go` already auto-creates owner membership — do NOT duplicate this in frontend code
2. The `author_name` field must be added to the messages collection definition in `collections.go` AND populated in the `OnRecordCreate("messages")` hook in `auth.go`
3. The `room_members.CreateRule` must allow self-join, but the unique index prevents duplicates — frontend must catch the constraint error silently
4. After relaxing room rules, presence endpoints should revert to strict membership checks (no auto-join) — the frontend handles joining explicitly

**Files to Create/Modify:**
- `backend/hooks/collections.go` — API rules + `author_name` field on messages
- `backend/hooks/auth.go` — Denormalize `author_name` in message create hook
- `backend/hooks/presence.go` — Remove auto-join, revert to 403 on non-member
- `frontend/src/components/layout/RoomList.tsx` — List all rooms, remove redundant membership create
- `frontend/src/hooks/useMessages.ts` — Add `author_name` to interface + optimistic send
- `frontend/src/components/campfire/MessageBubble.tsx` — Read `author_name` from record
- `frontend/src/pages/RoomPage.tsx` or `frontend/src/components/campfire/CampfireRoom.tsx` — Add `ensureMembership()` on room entry

**Deferred Decisions (Builder can decide):**
- Whether `ensureMembership()` lives in a new hook (`useRoomMembership.ts`) or inline in `CampfireRoom.tsx`
- Whether to refetch the room list in the sidebar after joining a new room (via event or callback)
