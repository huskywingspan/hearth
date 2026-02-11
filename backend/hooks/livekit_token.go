package hooks

import (
	"os"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterLiveKitToken sets up the endpoint for generating LiveKit room access tokens.
// Voice-first: tokens grant audio publish/subscribe but NOT video by default.
func RegisterLiveKitToken(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// POST /api/hearth/rooms/{id}/token
		// Requires: authenticated user who is a member of the room
		se.Router.POST("/api/hearth/rooms/{id}/token", func(e *core.RequestEvent) error {
			info, _ := e.RequestInfo()
			roomID := e.Request.PathValue("id")

			// Find the room
			room, err := e.App.FindRecordById("rooms", roomID)
			if err != nil {
				return e.NotFoundError("Room not found", err)
			}

			// Verify user is a member
			_, err = e.App.FindFirstRecordByFilter(
				"room_members",
				"room = {:room} && user = {:user}",
				dbxParams("room", roomID, "user", info.Auth.Id),
			)
			if err != nil {
				return e.ForbiddenError("Not a member of this room", nil)
			}

			// Get LiveKit credentials from env
			apiKey := os.Getenv("LIVEKIT_API_KEY")
			apiSecret := os.Getenv("LIVEKIT_API_SECRET")
			if apiKey == "" || apiSecret == "" {
				return e.InternalServerError("LiveKit not configured", nil)
			}

			displayName := info.Auth.GetString("display_name")
			if displayName == "" {
				displayName = info.Auth.GetString("email")
			}

			livekitRoom := room.GetString("livekit_room_name")
			allowVideo := room.GetBool("allow_video")

			token, err := generateLiveKitToken(
				apiKey,
				apiSecret,
				livekitRoom,
				info.Auth.Id,
				displayName,
				allowVideo,
			)
			if err != nil {
				e.App.Logger().Error("LiveKit token generation failed", "error", err)
				return e.InternalServerError("Failed to generate token", nil)
			}

			return e.JSON(200, map[string]string{
				"token":       token,
				"room":        livekitRoom,
				"identity":    info.Auth.Id,
				"displayName": displayName,
			})
		}).Bind(apis.RequireAuth())

		return se.Next()
	})
}

// generateLiveKitToken creates a signed JWT for LiveKit room access.
// Voice-first: only microphone source is allowed unless allowVideo is true.
func generateLiveKitToken(apiKey, apiSecret, roomName, identity, displayName string, allowVideo bool) (string, error) {
	at := auth.NewAccessToken(apiKey, apiSecret)

	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         roomName,
		CanPublish:   boolPtr(true),
		CanSubscribe: boolPtr(true),
	}

	// Voice-first: restrict publish sources
	if allowVideo {
		grant.CanPublishSources = []string{"microphone", "camera"}
	} else {
		grant.CanPublishSources = []string{"microphone"}
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetName(displayName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}

func boolPtr(b bool) *bool {
	return &b
}
