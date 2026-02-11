package hooks

import (
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// PresenceEntry tracks a single user's online status in a specific room.
type PresenceEntry struct {
	UserID      string    `json:"user_id"`
	RoomID      string    `json:"room_id"`
	DisplayName string    `json:"display_name"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PresenceMap is a thread-safe in-memory store for user presence.
// Not persisted to SQLite — presence is ephemeral by design.
type PresenceMap struct {
	mu      sync.RWMutex
	entries map[string]*PresenceEntry // key: userID
}

// Global presence map — singleton for the lifetime of the process.
var presence = &PresenceMap{
	entries: make(map[string]*PresenceEntry),
}

// Heartbeat updates a user's presence. Called every 30s by clients.
func (pm *PresenceMap) Heartbeat(userID, roomID, displayName string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.entries[userID] = &PresenceEntry{
		UserID:      userID,
		RoomID:      roomID,
		DisplayName: displayName,
		UpdatedAt:   time.Now(),
	}
}

// GetRoomPresence returns all online users in a specific room.
func (pm *PresenceMap) GetRoomPresence(roomID string) []*PresenceEntry {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*PresenceEntry
	for _, entry := range pm.entries {
		if entry.RoomID == roomID {
			result = append(result, entry)
		}
	}
	return result
}

// Sweep removes entries older than the threshold.
func (pm *PresenceMap) Sweep(threshold time.Duration) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	cutoff := time.Now().Add(-threshold)
	removed := 0
	for k, v := range pm.entries {
		if v.UpdatedAt.Before(cutoff) {
			delete(pm.entries, k)
			removed++
		}
	}
	return removed
}

// OnlineCount returns the total number of online users.
func (pm *PresenceMap) OnlineCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.entries)
}

// Remove removes a specific user from the presence map.
func (pm *PresenceMap) Remove(userID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.entries, userID)
}

// RegisterPresence sets up heartbeat/presence endpoints and the stale-entry sweep cron.
func RegisterPresence(app *pocketbase.PocketBase) {
	// Sweep stale entries every 2 minutes (users with no heartbeat for >60s)
	app.Cron().MustAdd("hearth_presence_sweep", "*/2 * * * *", func() {
		removed := presence.Sweep(60 * time.Second)
		if removed > 0 {
			app.Logger().Info("presence sweep", "removed", removed)
		}
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// POST /api/hearth/presence/heartbeat
		// Body: { "room_id": "..." }
		// Requires auth.
		se.Router.POST("/api/hearth/presence/heartbeat", func(e *core.RequestEvent) error {
			info, _ := e.RequestInfo()
			data := struct {
				RoomID string `json:"room_id"`
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Invalid request body", err)
			}
			if data.RoomID == "" {
				return e.BadRequestError("room_id is required", nil)
			}

			// Verify user is a member of the room (no auto-join — ADR-006)
			_, err := e.App.FindFirstRecordByFilter(
				"room_members",
				"room = {:room} && user = {:user}",
				dbxParams("room", data.RoomID, "user", info.Auth.Id),
			)
			if err != nil {
				return e.ForbiddenError("Not a member of this room", nil)
			}

			displayName := info.Auth.GetString("display_name")
			if displayName == "" {
				displayName = info.Auth.GetString("email")
			}

			presence.Heartbeat(info.Auth.Id, data.RoomID, displayName)

			return e.JSON(200, map[string]bool{"ok": true})
		}).Bind(apis.RequireAuth())

		// GET /api/hearth/presence/{roomId}
		// Returns online users in a room. Requires auth + room membership.
		se.Router.GET("/api/hearth/presence/{roomId}", func(e *core.RequestEvent) error {
			info, _ := e.RequestInfo()
			roomID := e.Request.PathValue("roomId")

			// Verify membership (no auto-join — ADR-006)
			_, err := e.App.FindFirstRecordByFilter(
				"room_members",
				"room = {:room} && user = {:user}",
				dbxParams("room", roomID, "user", info.Auth.Id),
			)
			if err != nil {
				return e.ForbiddenError("Not a member of this room", nil)
			}

			entries := presence.GetRoomPresence(roomID)
			return e.JSON(200, map[string]any{
				"online": entries,
				"count":  len(entries),
			})
		}).Bind(apis.RequireAuth())

		return se.Next()
	})
}
