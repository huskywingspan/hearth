package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterCollections creates the Hearth data model programmatically.
// This is idempotent — existing collections are skipped.
func RegisterCollections(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if err := ensureUsersFields(se.App); err != nil {
			se.App.Logger().Error("failed to extend users collection", "error", err)
		}
		if err := ensureRoomsCollection(se.App); err != nil {
			se.App.Logger().Error("failed to create rooms collection", "error", err)
		}
		if err := ensureMessagesCollection(se.App); err != nil {
			se.App.Logger().Error("failed to create messages collection", "error", err)
		}
		if err := ensureRoomMembersCollection(se.App); err != nil {
			se.App.Logger().Error("failed to create room_members collection", "error", err)
		}
		if err := createIndexes(se.App); err != nil {
			se.App.Logger().Error("failed to create indexes", "error", err)
		}

		se.App.Logger().Info("Hearth collections verified")
		return se.Next()
	})
}

// ensureUsersFields extends the built-in users auth collection with Hearth-specific fields.
func ensureUsersFields(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}

	// Add display_name if missing
	if collection.Fields.GetByName("display_name") == nil {
		collection.Fields.Add(&core.TextField{
			Name:     "display_name",
			Required: true,
			Min:      1,
			Max:      50,
		})
	}

	// Add avatar_url if missing
	if collection.Fields.GetByName("avatar_url") == nil {
		collection.Fields.Add(&core.URLField{
			Name: "avatar_url",
		})
	}

	// Add status if missing
	if collection.Fields.GetByName("status") == nil {
		collection.Fields.Add(&core.SelectField{
			Name:      "status",
			Values:    []string{"cozy", "away", "dnd"},
			MaxSelect: 1,
		})
	}

	return app.Save(collection)
}

// ensureRoomsCollection creates the rooms collection if it doesn't exist.
func ensureRoomsCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("rooms")
	if err == nil {
		return nil // already exists
	}

	collection := core.NewBaseCollection("rooms")

	collection.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Min:      1,
		Max:      100,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "slug",
		Required: true,
		Min:      1,
		Max:      100,
		Pattern:  `^[a-z0-9]+(?:-[a-z0-9]+)*$`,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "owner",
		Required:      true,
		CollectionId:  "users",
		MaxSelect:     1,
		CascadeDelete: false,
	})

	collection.Fields.Add(&core.TextField{
		Name: "description",
		Max:  500,
	})

	collection.Fields.Add(&core.NumberField{
		Name:     "default_ttl",
		Required: true,
		Min:      floatPtr(60),    // 1 minute minimum
		Max:      floatPtr(86400), // 24 hours maximum
	})

	collection.Fields.Add(&core.NumberField{
		Name:     "max_participants",
		Required: true,
		Min:      floatPtr(2),
		Max:      floatPtr(25),
	})

	collection.Fields.Add(&core.BoolField{
		Name: "allow_video",
	})

	collection.Fields.Add(&core.TextField{
		Name:     "livekit_room_name",
		Required: true,
		Min:      1,
		Max:      200,
	})

	// API Rules — access control
	collection.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room_members_via_room.user`)
	collection.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room_members_via_room.user`)
	collection.CreateRule = stringPtr(`@request.auth.id != ""`)
	collection.UpdateRule = stringPtr(`@request.auth.id = owner`)
	collection.DeleteRule = stringPtr(`@request.auth.id = owner`)

	// Add unique indexes
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_rooms_slug ON rooms (slug)",
		"CREATE UNIQUE INDEX idx_rooms_livekit ON rooms (livekit_room_name)",
	}

	return app.Save(collection)
}

// ensureMessagesCollection creates the messages collection if it doesn't exist.
func ensureMessagesCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("messages")
	if err == nil {
		return nil
	}

	collection := core.NewBaseCollection("messages")

	collection.Fields.Add(&core.RelationField{
		Name:          "room",
		Required:      true,
		CollectionId:  "rooms",
		MaxSelect:     1,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "author",
		Required:      true,
		CollectionId:  "users",
		MaxSelect:     1,
		CascadeDelete: false,
	})

	collection.Fields.Add(&core.TextField{
		Name:     "body",
		Required: true,
		Min:      1,
		Max:      4000,
	})

	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		Values:    []string{"text", "system", "emote"},
		MaxSelect: 1,
	})

	collection.Fields.Add(&core.DateField{
		Name:     "expires_at",
		Required: true,
	})

	// API Rules
	collection.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	collection.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	collection.CreateRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	collection.UpdateRule = stringPtr(`@request.auth.id = author`)
	collection.DeleteRule = stringPtr(`@request.auth.id = author || @request.auth.id = room.owner`)

	return app.Save(collection)
}

// ensureRoomMembersCollection creates the room_members join collection.
func ensureRoomMembersCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("room_members")
	if err == nil {
		return nil
	}

	collection := core.NewBaseCollection("room_members")

	collection.Fields.Add(&core.RelationField{
		Name:          "room",
		Required:      true,
		CollectionId:  "rooms",
		MaxSelect:     1,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "user",
		Required:      true,
		CollectionId:  "users",
		MaxSelect:     1,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.SelectField{
		Name:      "role",
		Required:  true,
		Values:    []string{"owner", "member", "guest"},
		MaxSelect: 1,
	})

	collection.Fields.Add(&core.RelationField{
		Name:         "vouched_by",
		CollectionId: "users",
		MaxSelect:    1,
	})

	// API Rules
	collection.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	collection.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	collection.CreateRule = stringPtr(`@request.auth.id != "" && @request.auth.id = room.owner`)
	collection.DeleteRule = stringPtr(`@request.auth.id = room.owner || @request.auth.id = user`)

	// Unique constraint: one membership per user per room
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_room_members_unique ON room_members (room, \"user\")",
	}

	return app.Save(collection)
}

// createIndexes adds performance-critical indexes for the message GC query.
func createIndexes(app core.App) error {
	_, err := app.DB().NewQuery(`
		CREATE INDEX IF NOT EXISTS idx_messages_expires_at 
		ON messages (expires_at) 
		WHERE expires_at IS NOT NULL
	`).Execute()
	return err
}

// Helper to create a *string from a string literal (for PocketBase API rules).
func stringPtr(s string) *string {
	return &s
}

// Helper to create a *float64 for number field constraints.
func floatPtr(f float64) *float64 {
	return &f
}
