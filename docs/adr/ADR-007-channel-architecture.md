# ADR-007: Channel Architecture — Dens, Campfires, DMs & Roles

**Date:** 2026-02-11 | **Status:** Accepted
**Supersedes:** "Portal" concept (retired)
**Informed by:** R-009 Gap Analysis, UX Research §1-5, Technical Research §2.1.2

---

## Context

Through v0.2.1, Hearth had a flat model: "rooms" (campfires) with fading text. The original master plan described a "Portal" — an abstract continuous-topology spatial voice space. After extensive discussion, we identified that:

1. A rich continuous spatial model is overengineered for v1.0 — it requires full 2D positioning, coordinate sync, and complex audio math.
2. Users need **permanent** conversation spaces (not just ephemeral campfires) for ongoing topics.
3. Direct messages are expected and missing entirely from the spec.
4. The original Technical Research (§2.1.2) already proposed a `type ENUM('voice', 'text', 'hybrid')` on rooms — this was dropped during master plan distillation.
5. Admin delegation is needed so the Homeowner isn't a bottleneck.

## Decision

### Channel Types

| Type | Persistence | Voice | Location | Analogy |
|------|------------|-------|----------|---------|
| **Den** | Permanent text history | Optional (Table + 4 Corners) | Inside the House | A room in your home — the living room, study, game room |
| **Campfire** | Ephemeral (fading text, configurable TTL) | None | The Backyard | Conversations around a fire — intimate, impermanent |
| **DM** | Permanent text history | Optional (1:1 call) | Private | A phone call or letter between two people |

### Voice Model — Table + 4 Corners (Dens only)

The continuous-topology "Portal" is retired. Voice in Dens uses a discrete spatial model:

- **The Table** — Central area where everyone hears everyone equally. Default position when joining voice. Like sitting at a dinner table.
- **4 Corners** — Discrete semi-private positions within the Den. Moving to a Corner with another person creates a semi-private conversation (reduced volume from Table, increased volume from Corner partner). Like stepping aside at a party.
- **No WASD, no coordinate system.** Click to move between Table ↔ Corner positions. Simple state machine, not continuous interpolation.

Spatial audio processing (PannerNode, linear distance model) from R-006 still applies but with discrete positions rather than continuous coordinates.

### Admin Roles

| Role | Capabilities | Assignment |
|------|-------------|------------|
| **Homeowner** | Full server control. Create/delete Dens and Campfires. Manage all settings. Assign Keyholders. | First account created (server setup) |
| **Keyholder** | Create/configure Dens and Campfires (if delegated). Moderate within delegated spaces. Cannot change server-level settings. | Assigned by Homeowner |
| **Member** | Join Dens and Campfires. Send messages. Join voice. | Any authenticated user |
| **Guest** | Session-bound visitor via The Knock. Limited to approved space. | Arrives via Knock, vouched by host |

### New Member History Visibility

When someone joins an existing Den, can they see messages from before they joined?

**Decision:** Configurable per-Den. The Homeowner can:
1. Set a server-wide default (show history / hide history)
2. Enable room creators to override the default per-Den
3. Override any Den's setting directly

This preserves privacy while allowing flexibility. **E2EE interaction:** For E2EE Dens (future), new members cannot decrypt pre-join messages regardless of this setting — crypto enforces the boundary.

### E2EE Scope for v1.0

| Channel Type | E2EE at v1.0 | Rationale |
|-------------|-------------|-----------|
| **Campfires** | ✅ Yes | Ephemeral + encrypted = maximum privacy |
| **DMs** | ✅ Yes | Private conversations demand encryption |
| **Dens** | ❌ No (v2.0) | Permanent history + new-member access + server-side features (search, moderation) conflict with E2EE. Complexity deferred. |
| **Voice** | ❌ No (v2.0) | Insertable Streams E2EE is complex; voice-first priority is getting spatial audio working |

### Video Policy (Dens with voice)

When video is enabled in a Den:
- **Max resolution:** 480p (from Technical Research §3.3)
- **Simulcast:** Disabled (saves server CPU)
- **Dynacast:** Enabled (pause unsubscribed streams after 5s)
- **Default:** `canPublishVideo: false` in LiveKit JWT — host explicitly enables per-Den
- **Frame rate:** 15fps max for "ambient video" / "picture frame" aesthetic

### Vocabulary Update

| Old Term | New Term | Reason |
|----------|---------|--------|
| Portal | **Den** (for rooms) / retired | "Portal" was tied to continuous topology; "Den" is warm, domestic, fits the House metaphor |
| Room | **Den** | "Room" is too generic, too Discord |
| Server | **House** | Fits the domestic metaphor |
| (none) | **Backyard** | Where Campfires live — outdoor annex to the House |
| (none) | **Homeowner** | Server admin |
| (none) | **Keyholder** | Delegated admin |
| (role) | **Table** | Central voice area in a Den |
| (role) | **Corner** | Semi-private voice position in a Den |

## Rationale

- **Den vs Room:** Tested "Nook," "Parlor," "Lounge" — Den is warm, lived-in, and already established as the default room name ("The Den"). Plural is natural ("create a new Den").
- **Table + Corners vs continuous topology:** Dramatically simpler to implement (discrete positions vs coordinate math), easier to understand for users ("sit at the table" vs "drift around an abstract space"), and still provides spatial feel. The continuous model can return in v2.0+ as an advanced mode.
- **Homeowner/Keyholder:** Domestic vocabulary consistent with House/Den. "Keyholder" implies trust delegation without full ownership.
- **E2EE on Campfires + DMs only:** Campfires are ephemeral (key management simpler — no history to re-encrypt), DMs are 1:1 (key exchange simpler). Dens have complex requirements (new member history, search, moderation) that conflict with E2EE.

## Impact on Components

| Component | Change Required |
|-----------|----------------|
| **Schema** | Add `type` field on rooms (`den`, `campfire`). Add `role` field on users (`homeowner`, `keyholder`, `member`). Add DM collection. |
| **Backend hooks** | Room creation respects type. DM creation. Admin role checks. History visibility per-Den config. |
| **Frontend** | Separate Den and Campfire views. DM UI. Role-based admin controls. Voice integration (Table/Corners) in Den view. |
| **Master plan** | §3 rewritten. Vocabulary throughout. |
| **Roadmap** | v0.3 split — "First Friend" (schema + remote access) before "Hearth Fire" (voice). v1.0 adds chat E2EE. |
| **R-005/R-006** | References to "Portal" → "Den" / "Table". Audio rendering still uses same Web Audio approach. |

## Add to Chronicle: Yes

---

## Schema Sketch

### rooms collection (updated)
```
id          TEXT PK
name        TEXT NOT NULL
description TEXT
type        TEXT NOT NULL DEFAULT 'den'   -- 'den' | 'campfire'
default_ttl INTEGER                       -- seconds; NULL for dens (permanent)
voice       BOOLEAN DEFAULT false         -- enable voice (Table + Corners)
video       BOOLEAN DEFAULT false         -- enable video (480p max)
history_visible BOOLEAN DEFAULT true      -- can new members see pre-join messages?
created_by  RELATION(users)
created     DATETIME
updated     DATETIME
```

### direct_messages collection (new)
```
id          TEXT PK
participants RELATION(users, multiple)   -- exactly 2 users
created     DATETIME
updated     DATETIME
```

### dm_messages collection (new)
```
id          TEXT PK
dm          RELATION(direct_messages)
author      RELATION(users)
author_name TEXT
body        TEXT
created     DATETIME
updated     DATETIME
```

### users collection (updated fields)
```
role        TEXT DEFAULT 'member'         -- 'homeowner' | 'keyholder' | 'member'
public_key  TEXT                          -- for E2EE (v1.0), empty until enrolled
```
