package hooks

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// powChallenge stores an active Proof-of-Work challenge.
type powChallenge struct {
	ID         string
	Difficulty int
	ExpiresAt  time.Time
}

// powStore holds active PoW challenges in memory. Short-lived, no DB needed.
type powStore struct {
	mu         sync.RWMutex
	challenges map[string]*powChallenge
}

var powChallenges = &powStore{
	challenges: make(map[string]*powChallenge),
}

// RegisterPoW sets up the Client Puzzle Protocol endpoints.
// SHA256 partial collision: client must find nonce where
// SHA256(challenge_id + nonce) has N leading zero bits.
func RegisterPoW(app *pocketbase.PocketBase) {
	// Sweep expired challenges every 5 minutes
	app.Cron().MustAdd("hearth_pow_sweep", "*/5 * * * *", func() {
		powChallenges.sweep()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// GET /api/hearth/pow/challenge
		// Returns a new challenge for the client to solve.
		se.Router.GET("/api/hearth/pow/challenge", func(e *core.RequestEvent) error {
			difficulty := getPowDifficulty()

			// Generate random challenge ID
			challengeID, err := generateRandomHex(16)
			if err != nil {
				return e.InternalServerError("Failed to generate challenge", err)
			}

			expiresAt := time.Now().Add(5 * time.Minute)

			powChallenges.mu.Lock()
			powChallenges.challenges[challengeID] = &powChallenge{
				ID:         challengeID,
				Difficulty: difficulty,
				ExpiresAt:  expiresAt,
			}
			powChallenges.mu.Unlock()

			return e.JSON(200, map[string]any{
				"challenge_id": challengeID,
				"difficulty":   difficulty,
				"expires":      expiresAt.Unix(),
			})
		})

		// POST /api/hearth/pow/verify
		// Body: { "challenge_id": "...", "nonce": "..." }
		// Returns a one-time-use PoW token on success.
		se.Router.POST("/api/hearth/pow/verify", func(e *core.RequestEvent) error {
			data := struct {
				ChallengeID string `json:"challenge_id"`
				Nonce       string `json:"nonce"`
			}{}
			if err := e.BindBody(&data); err != nil {
				return e.BadRequestError("Invalid request body", err)
			}

			// Look up and consume the challenge (one-time use)
			powChallenges.mu.Lock()
			challenge, exists := powChallenges.challenges[data.ChallengeID]
			if exists {
				delete(powChallenges.challenges, data.ChallengeID)
			}
			powChallenges.mu.Unlock()

			if !exists {
				return e.BadRequestError("Unknown or already-used challenge", nil)
			}

			if time.Now().After(challenge.ExpiresAt) {
				return e.BadRequestError("Challenge expired", nil)
			}

			// Verify the solution
			if !verifyPoW(challenge.ID, data.Nonce, challenge.Difficulty) {
				return e.BadRequestError("Invalid solution", nil)
			}

			// Generate a one-time PoW token (used as proof for protected endpoints)
			token, err := generateRandomHex(32)
			if err != nil {
				return e.InternalServerError("Failed to generate token", err)
			}

			return e.JSON(200, map[string]any{
				"valid": true,
				"token": token,
			})
		})

		return se.Next()
	})
}

// verifyPoW checks if SHA256(challengeID + nonce) has `difficulty` leading zero bits.
func verifyPoW(challengeID, nonce string, difficulty int) bool {
	input := challengeID + nonce
	hash := sha256.Sum256([]byte(input))

	// Check leading zero bits
	bitsChecked := 0
	for _, b := range hash {
		if bitsChecked+8 <= difficulty {
			if b != 0 {
				return false
			}
			bitsChecked += 8
		} else {
			remaining := difficulty - bitsChecked
			if remaining > 0 {
				mask := byte(0xFF) << (8 - remaining)
				if b&mask != 0 {
					return false
				}
			}
			break
		}
	}
	return true
}

// generateRandomHex creates a cryptographically random hex string.
func generateRandomHex(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// getPowDifficulty reads difficulty from env or returns default (20 bits).
func getPowDifficulty() int {
	s := os.Getenv("POW_DIFFICULTY")
	if s == "" {
		return 20
	}
	d, err := strconv.Atoi(s)
	if err != nil || d < 1 || d > 32 {
		return 20
	}
	return d
}

// sweep removes expired challenges from the in-memory store.
func (ps *powStore) sweep() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	now := time.Now()
	for k, v := range ps.challenges {
		if now.After(v.ExpiresAt) {
			delete(ps.challenges, k)
		}
	}
}


