package hooks

import (
	"os"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RateLimiter implements a sliding-window token bucket rate limiter.
// In-memory only — no Redis needed. A sync.Mutex map with periodic sweep.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
}

type rateBucket struct {
	tokens     float64
	lastCheck  time.Time
	maxTokens  float64
	refillRate float64 // tokens per second
}

// RateLimitConfig defines limits for different endpoint categories.
type RateLimitConfig struct {
	MaxTokens  float64
	RefillRate float64 // tokens per second
}

// Pre-defined rate limit profiles
var (
	// Auth endpoints (login/register): 5 requests per 15 minutes
	rateLimitAuth = RateLimitConfig{MaxTokens: 5, RefillRate: 5.0 / 900.0}
	// Auth refresh (token keepalive): 10 requests per minute — generous, it's not a login attempt
	rateLimitAuthRefresh = RateLimitConfig{MaxTokens: 10, RefillRate: 10.0 / 60.0}
	// Invite validation: 10 requests per minute
	rateLimitInvite = RateLimitConfig{MaxTokens: 10, RefillRate: 10.0 / 60.0}
	// General API: 120 requests per minute (covers room navigation bursts)
	rateLimitGeneral = RateLimitConfig{MaxTokens: 120, RefillRate: 2.0}
	// Message creation: 30 messages per minute (per user)
	rateLimitMessage = RateLimitConfig{MaxTokens: 30, RefillRate: 0.5}
	// Heartbeat: 6 requests per minute (normal is 2/min @ 30s interval)
	rateLimitHeartbeat = RateLimitConfig{MaxTokens: 6, RefillRate: 0.1}
)

// NewRateLimiter creates a new in-memory rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*rateBucket),
	}
}

// Allow checks if the given key is within rate limits. Returns true if allowed.
func (rl *RateLimiter) Allow(key string, config RateLimitConfig) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]
	if !exists {
		rl.buckets[key] = &rateBucket{
			tokens:     config.MaxTokens - 1, // consume one token
			lastCheck:  now,
			maxTokens:  config.MaxTokens,
			refillRate: config.RefillRate,
		}
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * config.RefillRate
	if b.tokens > config.MaxTokens {
		b.tokens = config.MaxTokens
	}
	b.lastCheck = now

	if b.tokens < 1 {
		return false // rate limited
	}

	b.tokens--
	return true
}

// SweepStale removes buckets that haven't been accessed for the given duration.
// Prevents unbounded memory growth.
func (rl *RateLimiter) SweepStale(maxAge time.Duration) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for key, b := range rl.buckets {
		if b.lastCheck.Before(cutoff) {
			delete(rl.buckets, key)
			removed++
		}
	}
	return removed
}

// BucketCount returns the number of active rate limit buckets (for metrics/testing).
func (rl *RateLimiter) BucketCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.buckets)
}

// Global rate limiter instance — singleton for the lifetime of the process.
var limiter = NewRateLimiter()

// RegisterRateLimit sets up per-IP and per-user rate limiting as middleware.
func RegisterRateLimit(app *pocketbase.PocketBase) {
	// Sweep stale buckets every 5 minutes (memory hygiene)
	app.Cron().MustAdd("hearth_ratelimit_sweep", "*/5 * * * *", func() {
		removed := limiter.SweepStale(10 * time.Minute)
		if removed > 0 {
			app.Logger().Info("rate limiter sweep", "removed_buckets", removed)
		}
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Middleware: rate limit all API requests
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			path := e.Request.URL.Path
			ip := e.RealIP()

			// Determine which rate limit profile to use
			var config RateLimitConfig
			var key string

			switch {
			case isAuthRefreshPath(path):
				config = rateLimitAuthRefresh
				key = "auth-refresh:" + ip
			case isAuthPath(path):
				config = rateLimitAuth
				key = "auth:" + ip
			case isInvitePath(path):
				config = rateLimitInvite
				key = "invite:" + ip
			case isMessageCreatePath(path, e.Request.Method):
				// Per-user rate limit for message creation
				config = rateLimitMessage
				info, _ := e.RequestInfo()
				if info != nil && info.Auth != nil {
					key = "msg:" + info.Auth.Id
				} else {
					key = "msg:" + ip
				}
			case isHeartbeatPath(path):
				config = rateLimitHeartbeat
				info, _ := e.RequestInfo()
				if info != nil && info.Auth != nil {
					key = "hb:" + info.Auth.Id
				} else {
					key = "hb:" + ip
				}
			default:
				config = rateLimitGeneral
				key = "api:" + ip
			}

			if !limiter.Allow(key, config) {
				app.Logger().Warn("rate limit exceeded",
					"key", key,
					"ip", ip,
					"path", path,
				)
				e.Response.Header().Set("Retry-After", "60")
				return e.JSON(429, map[string]string{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded. Please slow down.",
				})
			}

			return e.Next()
		})

		return se.Next()
	})
}

// Path classification helpers

func isAuthPath(path string) bool {
	return matchPrefix(path, "/api/collections/users/auth-with-password")
}

func isAuthRefreshPath(path string) bool {
	return matchPrefix(path, "/api/collections/users/auth-refresh")
}

func isInvitePath(path string) bool {
	return matchPrefix(path, "/api/hearth/invite/validate")
}

func isMessageCreatePath(path string, method string) bool {
	return method == "POST" && matchPrefix(path, "/api/collections/messages/records")
}

func isHeartbeatPath(path string) bool {
	return matchPrefix(path, "/api/hearth/presence/heartbeat")
}

func matchPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// GetCORSOrigin returns the allowed CORS origin based on HEARTH_DOMAIN env.
// Used by RegisterCORS to lock origins.
func GetCORSOrigin() string {
	domain := os.Getenv("HEARTH_DOMAIN")
	if domain == "" || domain == "localhost" || domain == "localhost:8090" {
		return "http://localhost:5173" // Vite dev server
	}
	return "https://" + domain
}
