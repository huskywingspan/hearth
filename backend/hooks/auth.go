package hooks

import (
	"fmt"
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
		if ttlSeconds > 0 {
			// Campfire: messages expire after TTL
			expiresAt := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
			e.Record.Set("expires_at", expiresAt.UTC().Format(time.RFC3339))
		} else {
			// Den: messages are permanent — far-future expiry keeps GC query simple
			e.Record.Set("expires_at", "2099-12-31T23:59:59Z")
		}

		// Default message type to "text"
		if e.Record.GetString("type") == "" {
			e.Record.Set("type", "text")
		}

		// Denormalize author display_name onto the message (S3-004 / BUG-014)
		// This avoids expand queries on read and fixes the "Wanderer" bug.
		authorID := e.Record.GetString("author")
		if authorID != "" {
			author, authErr := e.App.FindRecordById("users", authorID)
			if authErr == nil {
				name := author.GetString("display_name")
				if name == "" {
					name = "Wanderer"
				}
				e.Record.Set("author_name", name)
			}
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

	// Before room creation: default type to campfire if not specified
	app.OnRecordCreate("rooms").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("type") == "" {
			e.Record.Set("type", "campfire")
		}
		return e.Next()
	})

	// First user registration: crown as Homeowner
	app.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		// Default role to member if not set
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", "member")
		}

		// Count total users — if this is the first, make them homeowner
		total, err := e.App.CountRecords("users")
		if err != nil || total > 1 {
			// Not the first user, or count error — save the default role
			if e.Record.GetString("role") == "" {
				e.Record.Set("role", "member")
				_ = e.App.Save(e.Record)
			}
			return e.Next()
		}

		// This is the first user — crown as homeowner
		e.Record.Set("role", "homeowner")
		if err := e.App.Save(e.Record); err != nil {
			e.App.Logger().Error("failed to crown first user as homeowner", "error", err)
		} else {
			e.App.Logger().Info("first user crowned as Homeowner", "user", e.Record.Id)
		}

		// Seed "The Den" now that we have a homeowner
		seedDefaultDen(e.App, e.Record.Id)

		return e.Next()
	})

	// DM message: denormalize author_name (same pattern as room messages)
	app.OnRecordCreate("dm_messages").BindFunc(func(e *core.RecordEvent) error {
		authorID := e.Record.GetString("author")
		if authorID != "" {
			author, err := e.App.FindRecordById("users", authorID)
			if err == nil {
				name := author.GetString("display_name")
				if name == "" {
					name = "Wanderer"
				}
				e.Record.Set("author_name", name)
			}
		}
		return e.Next()
	})
}

// seedDefaultDen creates "The Den" if no dens exist yet.
func seedDefaultDen(app core.App, ownerID string) {
	// Check if any dens already exist
	dens, _ := app.FindRecordsByFilter("rooms", "type = 'den'", "", 1, 0)
	if len(dens) > 0 {
		return // already have dens
	}

	roomsCol, err := app.FindCollectionByNameOrId("rooms")
	if err != nil {
		app.Logger().Error("cannot seed Den: rooms collection not found", "error", err)
		return
	}

	den := core.NewRecord(roomsCol)
	den.Set("name", "The Den")
	den.Set("slug", "the-den")
	den.Set("type", "den")
	den.Set("owner", ownerID)
	den.Set("description", "The main room. Pull up a chair.")
	den.Set("default_ttl", 0) // dens don't expire messages
	den.Set("max_participants", 25)
	den.Set("livekit_room_name", fmt.Sprintf("hearth-the-den-%d", time.Now().UnixMilli()))
	den.Set("voice", false) // voice enabled in v0.4
	den.Set("video", false)
	den.Set("history_visible", true)

	if err := app.Save(den); err != nil {
		app.Logger().Error("failed to seed The Den", "error", err)
	} else {
		app.Logger().Info("seeded default Den: 'The Den'")
	}
}
