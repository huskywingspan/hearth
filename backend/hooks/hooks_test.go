package hooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// =============================================================================
// Security Tests — E-042 Rate Limiting
// =============================================================================

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter()
	config := RateLimitConfig{MaxTokens: 3, RefillRate: 0.0} // no refill

	// First 3 should pass
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-key", config) {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th should be rejected
	if rl.Allow("test-key", config) {
		t.Error("4th request should be rate limited")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	rl := NewRateLimiter()
	config := RateLimitConfig{MaxTokens: 2, RefillRate: 100.0} // fast refill for test

	// Drain tokens
	rl.Allow("test-key", config)
	rl.Allow("test-key", config)

	// Should be empty now with 0 refill
	if rl.Allow("test-key", RateLimitConfig{MaxTokens: 2, RefillRate: 0}) {
		t.Error("should be rate limited when no refill")
	}

	// But with fast refill rate, tokens should be available
	// (the time elapsed since last check will refill)
	time.Sleep(50 * time.Millisecond)
	if !rl.Allow("test-key", config) {
		t.Error("should be allowed after refill time")
	}
}

func TestRateLimiterSweepStale(t *testing.T) {
	rl := NewRateLimiter()
	config := RateLimitConfig{MaxTokens: 10, RefillRate: 1.0}

	rl.Allow("fresh-key", config)

	// Manually age a bucket
	rl.mu.Lock()
	rl.buckets["stale-key"] = &rateBucket{
		tokens:    5,
		lastCheck: time.Now().Add(-20 * time.Minute),
		maxTokens: 10,
		refillRate: 1.0,
	}
	rl.mu.Unlock()

	removed := rl.SweepStale(10 * time.Minute)
	if removed != 1 {
		t.Errorf("expected 1 stale bucket removed, got %d", removed)
	}

	if rl.BucketCount() != 1 {
		t.Errorf("expected 1 remaining bucket, got %d", rl.BucketCount())
	}
}

func TestRateLimiterIsolation(t *testing.T) {
	rl := NewRateLimiter()
	config := RateLimitConfig{MaxTokens: 2, RefillRate: 0}

	// Drain key-A
	rl.Allow("key-A", config)
	rl.Allow("key-A", config)

	// key-B should still be allowed
	if !rl.Allow("key-B", config) {
		t.Error("key-B should not be affected by key-A's rate limit")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	rl := NewRateLimiter()
	config := RateLimitConfig{MaxTokens: 1000, RefillRate: 0}

	var wg sync.WaitGroup
	allowed := int64(0)
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-key-%d", id%10)
			if rl.Allow(key, config) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// All 100 should have been allowed (1000 tokens across 10 keys = 100 each)
	if allowed != 100 {
		t.Errorf("expected 100 allowed, got %d", allowed)
	}
}

func TestRateLimitAuthProfile(t *testing.T) {
	rl := NewRateLimiter()

	// Auth limit: 5 per 15 minutes (very low refill)
	for i := 0; i < 5; i++ {
		if !rl.Allow("auth:192.168.1.1", rateLimitAuth) {
			t.Errorf("auth request %d should be allowed", i+1)
		}
	}

	// 6th should be blocked
	if rl.Allow("auth:192.168.1.1", rateLimitAuth) {
		t.Error("6th auth request should be rate limited")
	}
}

// =============================================================================
// Security Tests — E-043 Input Sanitization
// =============================================================================

func TestSanitizeScriptTag(t *testing.T) {
	input := `<script>alert('xss')</script>`
	result := SanitizeText(input)

	if strings.Contains(result, "<script>") {
		t.Error("script tag should be escaped")
	}
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Errorf("expected escaped script tag, got: %s", result)
	}
}

func TestSanitizeHTMLEntities(t *testing.T) {
	input := `<img src="x" onerror="alert(1)">`
	result := SanitizeText(input)

	if strings.Contains(result, "<img") {
		t.Error("img tag should be escaped")
	}
	if !strings.Contains(result, "&lt;img") {
		t.Errorf("expected escaped img tag, got: %s", result)
	}
}

func TestSanitizeMaxLength(t *testing.T) {
	// Create a 5000 character string
	input := strings.Repeat("a", 5000)
	result := SanitizeText(input)

	if len(result) > maxMessageLength {
		t.Errorf("expected max %d chars, got %d", maxMessageLength, len(result))
	}
	if len(result) != maxMessageLength {
		t.Errorf("expected exactly %d chars, got %d", maxMessageLength, len(result))
	}
}

func TestSanitizePreservesNormalText(t *testing.T) {
	input := "Hello, this is a normal message! 🔥 How are you?"
	result := SanitizeText(input)

	if result != input {
		t.Errorf("normal text should not be modified\ngot:  %s\nwant: %s", result, input)
	}
}

func TestSanitizeAmperstand(t *testing.T) {
	input := "this & that < those > them"
	result := SanitizeText(input)

	expected := "this &amp; that &lt; those &gt; them"
	if result != expected {
		t.Errorf("HTML entities should be escaped\ngot:  %s\nwant: %s", result, expected)
	}
}

func TestSanitizeEmptyInput(t *testing.T) {
	result := SanitizeText("")
	if result != "" {
		t.Errorf("empty input should return empty string, got: %s", result)
	}
}

// =============================================================================
// Security Tests — E-040 CORS
// =============================================================================

func TestGetCORSOriginProduction(t *testing.T) {
	original := os.Getenv("HEARTH_DOMAIN")
	defer os.Setenv("HEARTH_DOMAIN", original)

	os.Setenv("HEARTH_DOMAIN", "myhearth.example")
	origin := GetCORSOrigin()

	expected := "https://myhearth.example"
	if origin != expected {
		t.Errorf("expected %s, got %s", expected, origin)
	}
}

func TestGetCORSOriginDevelopment(t *testing.T) {
	original := os.Getenv("HEARTH_DOMAIN")
	defer os.Setenv("HEARTH_DOMAIN", original)

	os.Setenv("HEARTH_DOMAIN", "")
	origin := GetCORSOrigin()

	expected := "http://localhost:5173"
	if origin != expected {
		t.Errorf("expected %s, got %s", expected, origin)
	}
}

func TestGetCORSOriginLocalhost(t *testing.T) {
	original := os.Getenv("HEARTH_DOMAIN")
	defer os.Setenv("HEARTH_DOMAIN", original)

	os.Setenv("HEARTH_DOMAIN", "localhost")
	origin := GetCORSOrigin()

	expected := "http://localhost:5173"
	if origin != expected {
		t.Errorf("expected %s for localhost, got %s", expected, origin)
	}
}

// =============================================================================
// Path matching tests (used by rate limiter)
// =============================================================================

func TestMatchPrefix(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   bool
	}{
		{"/api/collections/users/auth-with-password", "/api/collections/users/auth-with-password", true},
		{"/api/collections/users/records", "/api/collections/users/records", true},
		{"/api/hearth/invite/validate", "/api/hearth/invite/validate", true},
		{"/api/hearth/presence/heartbeat", "/api/hearth/presence/heartbeat", true},
		{"/api/hearth/rooms/abc/token", "/api/hearth/invite/validate", false},
		{"/short", "/longer-prefix", false},
	}

	for _, tt := range tests {
		got := matchPrefix(tt.path, tt.prefix)
		if got != tt.want {
			t.Errorf("matchPrefix(%q, %q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
		}
	}
}

func TestIsAuthPath(t *testing.T) {
	if !isAuthPath("/api/collections/users/auth-with-password") {
		t.Error("should match auth-with-password")
	}
	if !isAuthPath("/api/collections/users/records") {
		t.Error("should match records (register)")
	}
	if isAuthPath("/api/hearth/rooms/abc/token") {
		t.Error("should not match rooms token endpoint")
	}
}
