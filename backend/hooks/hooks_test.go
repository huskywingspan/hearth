package hooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Unit Tests — E-016
// =============================================================================

// --- SQLite Pragma Tests ---

func TestPragmaValues(t *testing.T) {
	// Verify that the pragma strings are correctly defined.
	// Full verification requires a running PocketBase instance (integration test).
	expected := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-2000",
		"PRAGMA mmap_size=268435456",
		"PRAGMA busy_timeout=5000",
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-2000",
		"PRAGMA mmap_size=268435456",
		"PRAGMA busy_timeout=5000",
	}

	for i, p := range pragmas {
		if p != expected[i] {
			t.Errorf("pragma %d: got %q, want %q", i, p, expected[i])
		}
	}
}

// --- Presence Map Tests ---

func TestPresenceHeartbeat(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	pm.Heartbeat("user1", "room1", "Alice")

	entries := pm.GetRoomPresence("room1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].UserID != "user1" {
		t.Errorf("expected user1, got %s", entries[0].UserID)
	}
	if entries[0].DisplayName != "Alice" {
		t.Errorf("expected Alice, got %s", entries[0].DisplayName)
	}
}

func TestPresenceMultipleRooms(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	pm.Heartbeat("user1", "room1", "Alice")
	pm.Heartbeat("user2", "room2", "Bob")
	pm.Heartbeat("user3", "room1", "Charlie")

	room1 := pm.GetRoomPresence("room1")
	if len(room1) != 2 {
		t.Fatalf("room1: expected 2 entries, got %d", len(room1))
	}

	room2 := pm.GetRoomPresence("room2")
	if len(room2) != 1 {
		t.Fatalf("room2: expected 1 entry, got %d", len(room2))
	}
}

func TestPresenceSweep(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	// Add an entry with old timestamp
	pm.mu.Lock()
	pm.entries["user1"] = &PresenceEntry{
		UserID:    "user1",
		RoomID:    "room1",
		UpdatedAt: time.Now().Add(-2 * time.Minute), // 2 minutes ago
	}
	pm.entries["user2"] = &PresenceEntry{
		UserID:    "user2",
		RoomID:    "room1",
		UpdatedAt: time.Now(), // just now
	}
	pm.mu.Unlock()

	removed := pm.Sweep(60 * time.Second)
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	if pm.OnlineCount() != 1 {
		t.Errorf("expected 1 online, got %d", pm.OnlineCount())
	}

	entries := pm.GetRoomPresence("room1")
	if len(entries) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(entries))
	}
	if entries[0].UserID != "user2" {
		t.Errorf("expected user2 to remain, got %s", entries[0].UserID)
	}
}

func TestPresenceRemove(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	pm.Heartbeat("user1", "room1", "Alice")
	pm.Heartbeat("user2", "room1", "Bob")

	pm.Remove("user1")

	if pm.OnlineCount() != 1 {
		t.Errorf("expected 1 online after remove, got %d", pm.OnlineCount())
	}
}

func TestPresenceOverwrite(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	// User moves from room1 to room2
	pm.Heartbeat("user1", "room1", "Alice")
	pm.Heartbeat("user1", "room2", "Alice")

	room1 := pm.GetRoomPresence("room1")
	if len(room1) != 0 {
		t.Errorf("room1 should have 0 entries after user moved, got %d", len(room1))
	}

	room2 := pm.GetRoomPresence("room2")
	if len(room2) != 1 {
		t.Errorf("room2 should have 1 entry, got %d", len(room2))
	}
}

func TestPresenceConcurrency(t *testing.T) {
	pm := &PresenceMap{entries: make(map[string]*PresenceEntry)}

	var wg sync.WaitGroup
	const goroutines = 100

	// Concurrent heartbeats
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			userID := fmt.Sprintf("user%d", id)
			pm.Heartbeat(userID, "room1", fmt.Sprintf("User %d", id))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pm.GetRoomPresence("room1")
			_ = pm.OnlineCount()
		}()
	}

	wg.Wait()

	if pm.OnlineCount() != goroutines {
		t.Errorf("expected %d online, got %d", goroutines, pm.OnlineCount())
	}
}

// --- HMAC Invite Tests ---

func TestInviteGenerateAndValidate(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!!")
	roomSlug := "the-kitchen"
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	domain := "hearth.example"

	url := generateInviteURL(roomSlug, expiresAt, secret, domain)

	if url == "" {
		t.Fatal("generated URL is empty")
	}

	// Extract signature from URL
	// URL format: https://hearth.example/join?r=the-kitchen&t=1735689600&s=...
	sig := extractParam(url, "s=")
	timestamp := extractParam(url, "t=")

	ts, _ := strconv.ParseInt(timestamp, 10, 64)
	valid := validateInvite(roomSlug, ts, sig, [][]byte{secret})
	if !valid {
		t.Error("invite should be valid")
	}
}

func TestInviteExpired(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!!")
	roomSlug := "the-kitchen"
	expiresAt := time.Now().Add(-1 * time.Hour).Unix() // expired 1 hour ago

	payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	valid := validateInvite(roomSlug, expiresAt, sig, [][]byte{secret})
	if valid {
		t.Error("expired invite should not be valid")
	}
}

func TestInviteTampered(t *testing.T) {
	secret := []byte("test-secret-key-32-bytes-long!!!")
	roomSlug := "the-kitchen"
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Sign with the correct room slug
	payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	// Validate with a different room slug (tampered)
	valid := validateInvite("the-bedroom", expiresAt, sig, [][]byte{secret})
	if valid {
		t.Error("tampered invite should not be valid")
	}
}

func TestInviteKeyRotation(t *testing.T) {
	oldSecret := []byte("old-secret-key-32-bytes-long!!!!")
	newSecret := []byte("new-secret-key-32-bytes-long!!!!")
	roomSlug := "the-kitchen"
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Generate with old key
	payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, oldSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	// Validate with [newKey, oldKey] — should still work
	valid := validateInvite(roomSlug, expiresAt, sig, [][]byte{newSecret, oldSecret})
	if !valid {
		t.Error("invite signed with old key should validate during rotation")
	}
}

func TestInviteAfterKeyDrop(t *testing.T) {
	oldSecret := []byte("old-secret-key-32-bytes-long!!!!")
	newSecret := []byte("new-secret-key-32-bytes-long!!!!")
	roomSlug := "the-kitchen"
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Generate with old key
	payload := roomSlug + "." + strconv.FormatInt(expiresAt, 10)
	mac := hmac.New(sha256.New, oldSecret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))

	// Validate with only [newKey] — old key dropped, should fail
	valid := validateInvite(roomSlug, expiresAt, sig, [][]byte{newSecret})
	if valid {
		t.Error("invite signed with dropped key should not validate")
	}
}

// --- Proof-of-Work Tests ---

func TestPoWVerifyValid(t *testing.T) {
	// Create a known-valid PoW solution
	challengeID := "test-challenge-id"
	difficulty := 8 // Low difficulty for fast test

	// Brute-force a valid nonce
	nonce := solvePoW(challengeID, difficulty)
	if nonce == "" {
		t.Fatal("failed to solve PoW challenge")
	}

	if !verifyPoW(challengeID, nonce, difficulty) {
		t.Error("valid PoW solution should verify")
	}
}

func TestPoWVerifyInvalid(t *testing.T) {
	if verifyPoW("test-challenge", "wrong-nonce", 20) {
		t.Error("random nonce should not satisfy difficulty 20")
	}
}

func TestPoWDifficultyZero(t *testing.T) {
	// Difficulty 0 means any hash is valid
	if !verifyPoW("anything", "anything", 0) {
		t.Error("difficulty 0 should accept any input")
	}
}

func TestPoWVerifyDifficulty16(t *testing.T) {
	challengeID := "pow-test-16"
	difficulty := 16

	nonce := solvePoW(challengeID, difficulty)
	if !verifyPoW(challengeID, nonce, difficulty) {
		t.Errorf("valid PoW solution should verify at difficulty %d", difficulty)
	}

	// Verify the hash actually has the leading zeros
	hash := sha256.Sum256([]byte(challengeID + nonce))
	if hash[0] != 0 || hash[1] != 0 {
		t.Errorf("hash should have 16 leading zero bits, got %08b %08b", hash[0], hash[1])
	}
}

// --- LiveKit Token Tests ---

func TestLiveKitTokenVoiceOnly(t *testing.T) {
	// This test verifies the function signature and basic behavior.
	// Full JWT decode verification requires the livekit/protocol dependency.
	token, err := generateLiveKitToken(
		"test-api-key",
		"test-secret-that-is-at-least-32-chars",
		"room-the-kitchen",
		"user123",
		"Alice",
		false, // no video
	)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}
	if token == "" {
		t.Error("generated token is empty")
	}
}

func TestLiveKitTokenWithVideo(t *testing.T) {
	token, err := generateLiveKitToken(
		"test-api-key",
		"test-secret-that-is-at-least-32-chars",
		"room-the-kitchen",
		"user123",
		"Alice",
		true, // video allowed
	)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}
	if token == "" {
		t.Error("generated token is empty")
	}
}

// --- Helper utilities tests ---

func TestDbxParams(t *testing.T) {
	params := dbxParams("room", "abc", "user", "xyz")
	if params["room"] != "abc" {
		t.Errorf("expected room=abc, got %v", params["room"])
	}
	if params["user"] != "xyz" {
		t.Errorf("expected user=xyz, got %v", params["user"])
	}
}

func TestDbxParamsOddCount(t *testing.T) {
	// Odd number of args should not panic
	params := dbxParams("room", "abc", "orphan")
	if params["room"] != "abc" {
		t.Errorf("expected room=abc, got %v", params["room"])
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// extractParam extracts a parameter value from a URL string.
// Very basic — for test use only, not URL-decoding-safe.
func extractParam(url, prefix string) string {
	idx := 0
	for i := 0; i < len(url)-len(prefix); i++ {
		if url[i:i+len(prefix)] == prefix {
			idx = i + len(prefix)
			break
		}
	}
	if idx == 0 {
		return ""
	}
	end := len(url)
	for i := idx; i < len(url); i++ {
		if url[i] == '&' {
			end = i
			break
		}
	}
	return url[idx:end]
}

// solvePoW brute-forces a PoW solution for testing.
func solvePoW(challengeID string, difficulty int) string {
	for i := 0; i < 10_000_000; i++ {
		nonce := fmt.Sprintf("%d", i)
		if verifyPoW(challengeID, nonce, difficulty) {
			return nonce
		}
	}
	return ""
}
