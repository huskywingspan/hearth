package hooks

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterCollections creates the Hearth data model programmatically.
// This is idempotent — existing collections are skipped.
func RegisterCollections(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Pass 1: Create collections WITHOUT API rules.
		// Rules reference back-relations across collections (rooms ↔ room_members),
		// so all collections must exist before rules can be validated by PocketBase.
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
		if err := ensureDirectMessagesCollection(se.App); err != nil {
			se.App.Logger().Error("failed to create direct_messages collection", "error", err)
		}
		if err := ensureDmMessagesCollection(se.App); err != nil {
			se.App.Logger().Error("failed to create dm_messages collection", "error", err)
		}

		// Pass 2: Apply API rules now that all collections exist.
		if err := applyAPIRules(se.App); err != nil {
			se.App.Logger().Error("failed to apply API rules", "error", err)
		}

		if err := createIndexes(se.App); err != nil {
			se.App.Logger().Error("failed to create indexes", "error", err)
		}

		// Pass 3: Backfill fields added to existing records (schema migrations).
		if err := backfillSchemaDefaults(se.App); err != nil {
			se.App.Logger().Error("failed to backfill schema defaults", "error", err)
		}

		// Pass 4: Seed "The Den" if no dens exist and we have a homeowner.
		// This handles existing installs upgrading to v0.3 (first-user hook
		// only seeds on NEW user creation, not retroactively).
		owns, _ := se.App.FindRecordsByFilter("users", "role = 'homeowner'", "", 1, 0)
		if len(owns) > 0 {
			seedDefaultDen(se.App, owns[0].Id)
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

	// Add role if missing (ADR-007: Homeowner / Keyholder / Member)
	if collection.Fields.GetByName("role") == nil {
		collection.Fields.Add(&core.SelectField{
			Name:      "role",
			Values:    []string{"homeowner", "keyholder", "member"},
			MaxSelect: 1,
		})
	}

	// Add public_key if missing (E2EE readiness — v1.0, empty until enrolled)
	if collection.Fields.GetByName("public_key") == nil {
		collection.Fields.Add(&core.TextField{
			Name: "public_key",
			Max:  500,
		})
	}

	return app.Save(collection)
}

// ensureRoomsCollection creates the rooms collection if it doesn't exist.
// If it does exist, adds any missing ADR-007 fields incrementally.
func ensureRoomsCollection(app core.App) error {
	existing, err := app.FindCollectionByNameOrId("rooms")
	if err == nil {
		// Collection exists — ensure ADR-007 fields are present (schema migration)
		changed := false

		// Relax default_ttl constraints to allow 0 for dens (no message expiry).
		// Original schema (v0.2.1) had Required:true, Min:60 — incompatible with den TTL=0.
		if f := existing.Fields.GetByName("default_ttl"); f != nil {
			if nf, ok := f.(*core.NumberField); ok {
				if nf.Required || (nf.Min != nil && *nf.Min > 0) {
					nf.Required = false
					nf.Min = floatPtr(0)
					changed = true
				}
			}
		}

		if existing.Fields.GetByName("type") == nil {
			existing.Fields.Add(&core.SelectField{
				Name:      "type",
				Values:    []string{"den", "campfire"},
				MaxSelect: 1,
			})
			changed = true
		}

		if existing.Fields.GetByName("voice") == nil {
			existing.Fields.Add(&core.BoolField{Name: "voice"})
			changed = true
		}

		if existing.Fields.GetByName("video") == nil {
			existing.Fields.Add(&core.BoolField{Name: "video"})
			changed = true
		}

		if existing.Fields.GetByName("history_visible") == nil {
			existing.Fields.Add(&core.BoolField{Name: "history_visible"})
			changed = true
		}

		if changed {
			return app.Save(existing)
		}
		return nil
	}

	// Look up the actual users collection ID — PocketBase requires
	// RelationField.CollectionId to be the real ID, not the name.
	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return fmt.Errorf("users collection not found: %w", err)
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
		CollectionId:  usersCol.Id,
		MaxSelect:     1,
		CascadeDelete: false,
	})

	collection.Fields.Add(&core.TextField{
		Name: "description",
		Max:  500,
	})

	collection.Fields.Add(&core.NumberField{
		Name: "default_ttl",
		Min:  floatPtr(0),     // 0 = den (no expiry), 60+ = campfire
		Max:  floatPtr(86400), // 24 hours maximum
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

	// ADR-007: Channel architecture fields
	collection.Fields.Add(&core.SelectField{
		Name:      "type",
		Values:    []string{"den", "campfire"},
		MaxSelect: 1,
	})

	collection.Fields.Add(&core.BoolField{
		Name: "voice",
	})

	collection.Fields.Add(&core.BoolField{
		Name: "video",
	})

	collection.Fields.Add(&core.BoolField{
		Name: "history_visible",
	})

	// Add unique indexes (rules applied in pass 2 via applyAPIRules)
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_rooms_slug ON rooms (slug)",
		"CREATE UNIQUE INDEX idx_rooms_livekit ON rooms (livekit_room_name)",
	}

	return app.Save(collection)
}

// ensureMessagesCollection creates the messages collection if it doesn't exist.
func ensureMessagesCollection(app core.App) error {
	existing, err := app.FindCollectionByNameOrId("messages")
	if err == nil {
		// Collection exists — ensure new fields are present (schema migration)
		changed := false
		if existing.Fields.GetByName("author_name") == nil {
			existing.Fields.Add(&core.TextField{
				Name: "author_name",
				Max:  50,
			})
			changed = true
		}
		if existing.Fields.GetByName("created") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			})
			changed = true
		}
		if existing.Fields.GetByName("updated") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			})
			changed = true
		}
		if changed {
			return app.Save(existing)
		}
		return nil
	}

	// Look up actual collection IDs (PocketBase requires real IDs, not names)
	roomsCol, err := app.FindCollectionByNameOrId("rooms")
	if err != nil {
		return fmt.Errorf("rooms collection not found: %w", err)
	}
	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return fmt.Errorf("users collection not found: %w", err)
	}

	collection := core.NewBaseCollection("messages")

	collection.Fields.Add(&core.RelationField{
		Name:          "room",
		Required:      true,
		CollectionId:  roomsCol.Id,
		MaxSelect:     1,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "author",
		Required:      true,
		CollectionId:  usersCol.Id,
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

	collection.Fields.Add(&core.TextField{
		Name: "author_name",
		Max:  50,
	})

	collection.Fields.Add(&core.DateField{
		Name:     "expires_at",
		Required: true,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	// Rules applied in pass 2 via applyAPIRules

	return app.Save(collection)
}

// ensureRoomMembersCollection creates the room_members join collection.
func ensureRoomMembersCollection(app core.App) error {
	_, err := app.FindCollectionByNameOrId("room_members")
	if err == nil {
		return nil
	}

	// Look up actual collection IDs (PocketBase requires real IDs, not names)
	roomsCol, err := app.FindCollectionByNameOrId("rooms")
	if err != nil {
		return fmt.Errorf("rooms collection not found: %w", err)
	}
	usersCol, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return fmt.Errorf("users collection not found: %w", err)
	}

	collection := core.NewBaseCollection("room_members")

	collection.Fields.Add(&core.RelationField{
		Name:          "room",
		Required:      true,
		CollectionId:  roomsCol.Id,
		MaxSelect:     1,
		CascadeDelete: true,
	})

	collection.Fields.Add(&core.RelationField{
		Name:          "user",
		Required:      true,
		CollectionId:  usersCol.Id,
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
		CollectionId: usersCol.Id,
		MaxSelect:    1,
	})

	// Rules applied in pass 2 via applyAPIRules

	// Unique constraint: one membership per user per room
	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_room_members_unique ON room_members (room, \"user\")",
	}

	return app.Save(collection)
}

// ensureDirectMessagesCollection creates the direct_messages collection for 1:1 DMs.
func ensureDirectMessagesCollection(app core.App) error {
	existing, err := app.FindCollectionByNameOrId("direct_messages")
	if err == nil {
		// Collection exists — ensure autodate fields are present
		changed := false
		if existing.Fields.GetByName("created") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			})
			changed = true
		}
		if existing.Fields.GetByName("updated") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			})
			changed = true
		}
		if changed {
			return app.Save(existing)
		}
		return nil
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

	// Participant B (always the higher-sorted user ID)
	collection.Fields.Add(&core.RelationField{
		Name:         "participant_b",
		Required:     true,
		CollectionId: usersCol.Id,
		MaxSelect:    1,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	collection.Indexes = []string{
		"CREATE UNIQUE INDEX idx_dm_participants ON direct_messages (participant_a, participant_b)",
	}

	return app.Save(collection)
}

// ensureDmMessagesCollection creates the dm_messages collection for DM text messages.
func ensureDmMessagesCollection(app core.App) error {
	existing, err := app.FindCollectionByNameOrId("dm_messages")
	if err == nil {
		// Collection exists — ensure autodate fields are present
		changed := false
		if existing.Fields.GetByName("created") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			})
			changed = true
		}
		if existing.Fields.GetByName("updated") == nil {
			existing.Fields.Add(&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			})
			changed = true
		}
		if changed {
			return app.Save(existing)
		}
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

	collection.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	collection.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return app.Save(collection)
}

// backfillSchemaDefaults sets default values on existing records that lack new fields.
// This handles the v0.2.1 → v0.3 migration (ADR-007).
func backfillSchemaDefaults(app core.App) error {
	// Backfill rooms: existing rooms without a type → campfire
	if _, err := app.DB().NewQuery(
		`UPDATE rooms SET type = 'campfire' WHERE type = '' OR type IS NULL`,
	).Execute(); err != nil {
		return fmt.Errorf("backfill rooms.type: %w", err)
	}

	// Backfill rooms: set history_visible default for existing rooms
	if _, err := app.DB().NewQuery(
		`UPDATE rooms SET history_visible = 1 WHERE history_visible IS NULL`,
	).Execute(); err != nil {
		return fmt.Errorf("backfill rooms.history_visible: %w", err)
	}

	// Backfill users: existing users without a role → member
	if _, err := app.DB().NewQuery(
		`UPDATE users SET role = 'member' WHERE role = '' OR role IS NULL`,
	).Execute(); err != nil {
		return fmt.Errorf("backfill users.role: %w", err)
	}

	// Crown the first user as homeowner (by creation timestamp)
	if _, err := app.DB().NewQuery(
		`UPDATE users SET role = 'homeowner' WHERE id = (SELECT id FROM users ORDER BY created ASC LIMIT 1) AND role != 'homeowner'`,
	).Execute(); err != nil {
		// Don't fail startup if there are no users yet
		app.Logger().Warn("homeowner backfill skipped (maybe no users yet)", "error", err)
	}

	return nil
}

// applyAPIRules sets API rules on all Hearth collections.
// This runs AFTER all collections are created, so back-relation rules
// (e.g., rooms referencing room_members_via_room) can be validated.
func applyAPIRules(app core.App) error {
	// Users rules — allow authenticated users to search for other users (for DMs)
	users, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return fmt.Errorf("users not found for rules: %w", err)
	}
	users.ListRule = stringPtr(`@request.auth.id != ""`)
	users.ViewRule = stringPtr(`@request.auth.id != ""`)
	if err := app.Save(users); err != nil {
		return fmt.Errorf("users rules: %w", err)
	}

	// Rooms rules
	rooms, err := app.FindCollectionByNameOrId("rooms")
	if err != nil {
		return fmt.Errorf("rooms not found for rules: %w", err)
	}
	// ADR-006: Any authenticated user can list/view rooms (open-lobby model)
	rooms.ListRule = stringPtr(`@request.auth.id != ""`)
	rooms.ViewRule = stringPtr(`@request.auth.id != ""`)
	rooms.CreateRule = stringPtr(`@request.auth.id != ""`)
	rooms.UpdateRule = stringPtr(`@request.auth.id = owner`)
	rooms.DeleteRule = stringPtr(`@request.auth.id = owner`)
	if err := app.Save(rooms); err != nil {
		return fmt.Errorf("rooms rules: %w", err)
	}

	// Messages rules
	messages, err := app.FindCollectionByNameOrId("messages")
	if err != nil {
		return fmt.Errorf("messages not found for rules: %w", err)
	}
	// ADR-006: Messages require room membership (defense-in-depth)
	messages.ListRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	messages.ViewRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	messages.CreateRule = stringPtr(`@request.auth.id != "" && @request.auth.id ?= room.room_members_via_room.user`)
	messages.UpdateRule = stringPtr(`@request.auth.id = author`)
	messages.DeleteRule = stringPtr(`@request.auth.id = author || @request.auth.id = room.owner`)
	if err := app.Save(messages); err != nil {
		return fmt.Errorf("messages rules: %w", err)
	}

	// Room members rules
	members, err := app.FindCollectionByNameOrId("room_members")
	if err != nil {
		return fmt.Errorf("room_members not found for rules: %w", err)
	}
	members.ListRule = stringPtr(`@request.auth.id != ""`)
	members.ViewRule = stringPtr(`@request.auth.id != ""`)
	// ADR-006: Any auth user can create membership (self-join)
	members.CreateRule = stringPtr(`@request.auth.id != ""`)
	members.UpdateRule = stringPtr(`@request.auth.id = room.owner`)
	members.DeleteRule = stringPtr(`@request.auth.id = room.owner || @request.auth.id = user`)
	if err := app.Save(members); err != nil {
		return fmt.Errorf("room_members rules: %w", err)
	}

	// Direct messages rules — only participants can see their DMs
	dms, err := app.FindCollectionByNameOrId("direct_messages")
	if err != nil {
		return fmt.Errorf("direct_messages not found for rules: %w", err)
	}
	dms.ListRule = stringPtr(`participant_a = @request.auth.id || participant_b = @request.auth.id`)
	dms.ViewRule = stringPtr(`participant_a = @request.auth.id || participant_b = @request.auth.id`)
	dms.CreateRule = stringPtr(`@request.auth.id != ""`)
	dms.UpdateRule = nil // DMs cannot be updated
	dms.DeleteRule = nil // DMs cannot be deleted (permanent)
	if err := app.Save(dms); err != nil {
		return fmt.Errorf("direct_messages rules: %w", err)
	}

	// DM messages rules — only participants of the parent DM can read/write
	dmMsgs, err := app.FindCollectionByNameOrId("dm_messages")
	if err != nil {
		return fmt.Errorf("dm_messages not found for rules: %w", err)
	}
	dmMsgs.ListRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
	dmMsgs.ViewRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
	dmMsgs.CreateRule = stringPtr(`dm.participant_a = @request.auth.id || dm.participant_b = @request.auth.id`)
	dmMsgs.UpdateRule = stringPtr(`author = @request.auth.id`)
	dmMsgs.DeleteRule = stringPtr(`author = @request.auth.id`)
	if err := app.Save(dmMsgs); err != nil {
		return fmt.Errorf("dm_messages rules: %w", err)
	}

	return nil
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
