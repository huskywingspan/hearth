package hooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterInvite sets up HMAC invite token generation and validation endpoints.
// Invites are stateless — no DB writes on creation. Validation uses constant-time
// comparison (hmac.Equal) to prevent timing side-channel attacks.
func RegisterInvite(app *pocketbase.PocketBase) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// POST /api/hearth/invite/generate
		// Body: { "room_slug": "the-kitchen", "expires_in": 86400 }
		// Returns: { "url": "https://..." }
		// Requires auth + room membership
		se.Router.POST("/api/hearth/invite/generate", func(e *core.RequestEvent) error {
			info, _ := e.RequestInfo()

			data := struct {
				RoomSlug  string `json:"room_slug"`
				ExpiresIn int64  `json:"expires_in"` // seconds from now
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Invalid request body", err)
			}
			if data.RoomSlug == "" {
				return e.BadRequestError("room_slug is required", nil)
			}

			// Default expires_in to 24 hours
			if data.ExpiresIn <= 0 {
				data.ExpiresIn = 86400
			}
			// Cap at 7 days
			if data.ExpiresIn > 604800 {
				data.ExpiresIn = 604800
			}

			// Verify room exists
			room, err := e.App.FindFirstRecordByFilter(
				"rooms",
				"slug = {:slug}",
				dbxParams("slug", data.RoomSlug),
			)
			if err != nil {
				return e.NotFoundError("Room not found", err)
			}

			// Verify user is a member
			_, err = e.App.FindFirstRecordByFilter(
				"room_members",
				"room = {:room} && user = {:user}",
				dbxParams("room", room.Id, "user", info.Auth.Id),
			)
			if err != nil {
				return e.ForbiddenError("Not a member of this room", nil)
			}

			// Generate invite
			secret := getCurrentSecret()
			if secret == nil {
				return e.InternalServerError("Invite system not configured", nil)
			}

			expiresAt := time.Now().Unix() + data.ExpiresIn
			domain := os.Getenv("HEARTH_DOMAIN")
			if domain == "" {
				domain = "localhost:8090"
			}

			url := generateInviteURL(data.RoomSlug, expiresAt, secret, domain)

			return e.JSON(200, map[string]string{
				"url":        url,
				"room_slug":  data.RoomSlug,
				"expires_at": time.Unix(expiresAt, 0).UTC().Format(time.RFC3339),
			})
		}).Bind(apis.RequireAuth())

		// POST /api/hearth/invite/validate
		// Body: { "r": "the-kitchen", "t": "1735689600", "s": "f8a..." }
		// Public endpoint (PoW may be required separately)
		se.Router.POST("/api/hearth/invite/validate", func(e *core.RequestEvent) error {
			data := struct {
				RoomSlug  string `json:"r"`
				Timestamp string `json:"t"`
				Signature string `json:"s"`
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Invalid request body", err)
			}

			timestamp, err := strconv.ParseInt(data.Timestamp, 10, 64)
			if err != nil {
				return e.BadRequestError("Invalid timestamp", err)
			}

			secrets := getSecrets()
			if len(secrets) == 0 {
				return e.InternalServerError("Invite system not configured", nil)
			}

			if !validateInvite(data.RoomSlug, timestamp, data.Signature, secrets) {
				return e.BadRequestError("Invalid or expired invite", nil)
			}

			// Verify room exists
			room, err := e.App.FindFirstRecordByFilter(
				"rooms",
				"slug = {:slug}",
				dbxParams("slug", data.RoomSlug),
			)
			if err != nil {
				return e.NotFoundError("Room not found", nil)
			}

			return e.JSON(200, map[string]any{
				"valid":     true,
				"room_id":   room.Id,
				"room_slug": data.RoomSlug,
				"room_name": room.GetString("name"),
			})
		})

		return se.Next()
	})
}

// generateInviteURL creates a signed invite URL.
func generateInviteURL(roomSlug string, expiresAt int64, secret []byte, domain string) string {
	payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("https://%s/join?r=%s&t=%d&s=%s", domain, roomSlug, expiresAt, sig)
}

// validateInvite checks a signature against multiple secrets (for key rotation).
// Uses hmac.Equal which internally uses crypto/subtle.ConstantTimeCompare.
func validateInvite(roomSlug string, timestamp int64, signature string, secrets [][]byte) bool {
	// Check expiry first
	if time.Now().Unix() > timestamp {
		return false
	}

	payload := roomSlug + "." + strconv.FormatInt(timestamp, 10)

	for _, secret := range secrets {
		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(payload))
		expected := mac.Sum(nil)

		provided, err := hex.DecodeString(signature)
		if err != nil {
			return false
		}

		// hmac.Equal uses constant-time comparison — safe against timing attacks
		if hmac.Equal(expected, provided) {
			return true
		}
	}

	return false
}

// getCurrentSecret returns the current HMAC secret from env.
func getCurrentSecret() []byte {
	s := os.Getenv("HMAC_SECRET_CURRENT")
	if s == "" {
		return nil
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return []byte(s) // fallback to raw string if not hex
	}
	return decoded
}

// getSecrets returns both current and old secrets for rotation support.
func getSecrets() [][]byte {
	var secrets [][]byte

	if current := getCurrentSecret(); current != nil {
		secrets = append(secrets, current)
	}

	if old := os.Getenv("HMAC_SECRET_OLD"); old != "" {
		decoded, err := hex.DecodeString(old)
		if err != nil {
			secrets = append(secrets, []byte(old))
		} else {
			secrets = append(secrets, decoded)
		}
	}

	return secrets
}
