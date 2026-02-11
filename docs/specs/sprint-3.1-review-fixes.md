# Builder Sprint: Sprint 3.1 — Review Fixes

> **Review:** Sprint 3 Code Review (Reviewer Agent, 2026-02-11)
> **Spec:** `docs/specs/sprint-3-settling-in.md`
> **Baseline:** `tsc --noEmit` clean, `vite build` clean. `go test` **expected to fail** (3 stale tests — this sprint fixes them)

---

## Context

The Reviewer's Sprint 3 code review found 2 critical issues, 2 medium issues, and 2 low-priority items. This sprint fixes the 2 criticals and 1 medium that require code changes. The remaining items are non-blocking and tracked for a future pass.

**Critical insight for Issue #1:** ADR-006 (in `sprint-3-settling-in.md`, §"Design Decisions") explicitly says _"Keep `messages` rules requiring membership (unchanged)"_ — but the Builder relaxed messages rules to `@request.auth.id != ""`. This breaks defense-in-depth: any authenticated user can read/write messages to any room without joining. The frontend's `ensureMembership()` pattern is not a substitute for server-side enforcement.

---

## Phase 1 — Restore Messages API Rules · **Critical** · S

**File:** `backend/hooks/collections.go` → `applyAPIRules()` function

**What:** Lines 305-308 currently set messages rules to any-authenticated-user. Restore membership-checking rules so messages operations require the requesting user to be a member of the room.

**Current (wrong):**
```go
// ADR-006: Membership enforced at app level (ensureMembership on frontend)
messages.ListRule = stringPtr(`@request.auth.id != ""`)
messages.ViewRule = stringPtr(`@request.auth.id != ""`)
messages.CreateRule = stringPtr(`@request.auth.id != ""`)
```

**Required:**
```go
// ADR-006: Messages require room membership (defense-in-depth)
messages.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
messages.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
messages.CreateRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
```

The `?=` operator is PocketBase's "has/any" operator checking the back-relation. `room.room_members_via_room.user` traverses from message → room → room_members → user. This is the same pattern that was used for rooms before Sprint 3 relaxed it.

**Leave unchanged:** `UpdateRule` and `DeleteRule` are already correct (author-only / room-owner).

**Why this is safe:** `ensureMembership()` in `CampfireRoom.tsx` creates the membership *before* messages or presence are loaded. So the frontend will always have a valid membership by the time it hits the messages API. The server-side rule is defense-in-depth against direct API access.

---

## Phase 2 — Fix Stale Sanitize Tests · **Critical** · S

**File:** `backend/hooks/hooks_test.go`

**What:** Three tests still assert HTML entity escaping behavior from pre-PIVOT-004. `SanitizeText()` now only enforces length — no HTML escaping (React handles that). These tests will fail on `go test`.

**Test 1: `TestSanitizeScriptTag`** (lines 558-567)

Replace the existing assertions. The function should pass HTML through unchanged (React escapes on render):

```go
func TestSanitizeScriptTag(t *testing.T) {
	input := `<script>alert('xss')</script>`
	result := SanitizeText(input)

	// PIVOT-004: No server-side HTML escaping — React handles output safety.
	// SanitizeText only enforces length.
	if result != input {
		t.Errorf("SanitizeText should not modify HTML (React handles escaping)\ngot:  %s\nwant: %s", result, input)
	}
}
```

**Test 2: `TestSanitizeHTMLEntities`** (lines 569-579)

```go
func TestSanitizeHTMLEntities(t *testing.T) {
	input := `<img src="x" onerror="alert(1)">`
	result := SanitizeText(input)

	// PIVOT-004: No server-side HTML escaping — React handles output safety.
	if result != input {
		t.Errorf("SanitizeText should not modify HTML (React handles escaping)\ngot:  %s\nwant: %s", result, input)
	}
}
```

**Test 3: `TestSanitizeAmperstand`** (lines 601-609)

```go
func TestSanitizeAmperstand(t *testing.T) {
	input := "this & that < those > them"
	result := SanitizeText(input)

	// PIVOT-004: No server-side HTML escaping — React handles output safety.
	if result != input {
		t.Errorf("SanitizeText should not modify special chars (React handles escaping)\ngot:  %s\nwant: %s", result, input)
	}
}
```

**Leave unchanged:** `TestSanitizeMaxLength`, `TestSanitizePreservesNormalText`, `TestSanitizeEmptyInput` — these are correct.

---

## Phase 3 — Add `UpdateRule` to `room_members` · **Medium** · S

**File:** `backend/hooks/collections.go` → `applyAPIRules()` function

**What:** `room_members` has no `UpdateRule` set. Add one so the room owner can change member roles (e.g., promote to owner). This is defensive — there's no UI for it yet, but the schema should be correct.

After the existing `members.CreateRule` line and before `members.DeleteRule`, add:

```go
members.UpdateRule = stringPtr(`@request.auth.id = room.owner`)
```

---

## Phase 4 — Verify & Build · S

1. Run `go test ./hooks/...` — all tests must pass (including the 3 fixed sanitize tests)
2. Run `npx tsc --noEmit` — must be zero errors
3. Run `npx vite build` — must succeed
4. Copy build to pb_public: `Copy-Item -Recurse -Force .\dist\* ..\backend\pb_public\`

**Note:** Go is not installed on the local Windows machine. If this remains the case, verify `go test` compiles correctly by reading the test code for logical correctness, and flag that CI/Docker verification is still needed. Alternatively, use WSL or Docker to run `go test`.

---

## Ground Rules

- **Do NOT modify any other API rules** — rooms and presence rules are reviewed and correct
- **Do NOT touch `CampfireRoom.tsx` `ensureMembership()`** — the create-first pattern is approved. Issue #3 (better error handling) is tracked for a future sprint.
- **Do NOT add new dependencies**
- Run tests after each phase — must stay green

## Key Code Locations

| What | Where |
|------|-------|
| API rules definition | `backend/hooks/collections.go` → `applyAPIRules()` |
| Sanitize tests | `backend/hooks/hooks_test.go` lines 555-610 |
| Sanitize implementation | `backend/hooks/sanitize.go` → `SanitizeText()` |
| ADR-006 spec reference | `docs/specs/sprint-3-settling-in.md` §"Design Decisions" |
| PIVOT-004 context | `docs/PROJECT_CHRONICLE.md` §"Failed Approaches & Pivots" |

## Success Criteria

- [ ] `messages.ListRule`, `ViewRule`, `CreateRule` require membership via `room_members_via_room` back-relation
- [ ] `room_members.UpdateRule` set to room-owner-only
- [ ] `TestSanitizeScriptTag` passes (asserts passthrough, not escaping)
- [ ] `TestSanitizeHTMLEntities` passes (asserts passthrough)
- [ ] `TestSanitizeAmperstand` passes (asserts passthrough)
- [ ] All other existing tests remain green
- [ ] `tsc --noEmit` zero errors
- [ ] `vite build` clean
- [ ] No new dependencies introduced

---

## Deferred Items (Non-Blocking, Future Sprint)

| # | Severity | Issue | Tracked For |
|---|----------|-------|-------------|
| 3 | Medium | `ensureMembership()` catches all errors silently — should log non-409 errors | Next polish sprint |
| 5 | Low | Optimistic `author_name` uses bracket notation on untyped record | Next polish sprint |
| 6 | Low | `RoomList` doesn't re-fetch on external changes (other tabs, realtime) | Next polish sprint |
