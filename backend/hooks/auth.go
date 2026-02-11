package hooks

import (
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterAuth configures PocketBase's built-in auth and adds Hearth-specific hooks.
func RegisterAuth(app *pocketbase.PocketBase) {
	// Before user creation: ensure display_name is set
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		displayName := e.Record.GetString("display_name")
		if displayName == "" {
			// Fall back to email username if no display_name provided
			email := e.Record.GetString("email")
			if email != "" {
				for i, c := range email {
					if c == '@' {
						displayName = email[:i]
						break
					}
				}
			}
			if displayName == "" {
				displayName = "Wanderer"
			}
			e.Record.Set("display_name", displayName)
		}

		// Default status to "cozy"
		if e.Record.GetString("status") == "" {
			e.Record.Set("status", "cozy")
		}

		return e.Next()
	})

	// Before message creation: server-side TTL enforcement
	// Clients cannot set their own expires_at — the server overrides it
	app.OnRecordCreate("messages").BindFunc(func(e *core.RecordEvent) error {
		roomID := e.Record.GetString("room")
		if roomID == "" {
			return e.Next()
		}

		room, err := e.App.FindRecordById("rooms", roomID)
		if err != nil {
			return err
		}

		ttlSeconds := room.GetInt("default_ttl")
		if ttlSeconds <= 0 {
			ttlSeconds = 3600 // default 1 hour
		}

		expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
		e.Record.Set("expires_at", expiresAt.UTC().Format(time.RFC3339))

		// Default message type to "text"
		if e.Record.GetString("type") == "" {
			e.Record.Set("type", "text")
		}

		return e.Next()
	})

	// After room creation: auto-add the creator as owner member
	app.OnRecordAfterCreateSuccess("rooms").BindFunc(func(e *core.RecordEvent) error {
		memberCol, err := e.App.FindCollectionByNameOrId("room_members")
		if err != nil {
			e.App.Logger().Error("room_members collection not found", "error", err)
			return e.Next()
		}

		member := core.NewRecord(memberCol)
		member.Set("room", e.Record.Id)
		member.Set("user", e.Record.GetString("owner"))
		member.Set("role", "owner")

		if err := e.App.Save(member); err != nil {
			e.App.Logger().Error("failed to create owner membership", "error", err, "room", e.Record.Id)
		}

		return e.Next()
	})
}
