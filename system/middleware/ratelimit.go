package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/semutdev/goigniter/system/core"
)

// RateLimitConfig holds configuration for the rate limiter.
type RateLimitConfig struct {
	Max     int
	Window  time.Duration
	KeyFunc func(c *core.Context) string
	Message string
}

// DefaultRateLimitConfig returns a default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Max:     100,
		Window:  time.Minute,
		Message: "Too many requests",
		KeyFunc: func(c *core.Context) string {
			return c.IP()
		},
	}
}

// RateLimit returns a rate limiting middleware.
func RateLimit(max int, window time.Duration) core.Middleware {
	config := DefaultRateLimitConfig()
	config.Max = max
	config.Window = window
	return RateLimitWithConfig(config)
}

// RateLimitWithConfig returns a rate limiting middleware with custom config.
func RateLimitWithConfig(config RateLimitConfig) core.Middleware {
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *core.Context) string {
			return c.IP()
		}
	}
	if config.Message == "" {
		config.Message = "Too many requests"
	}

	store := newRateLimitStore(config.Window)

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			key := config.KeyFunc(c)

			if !store.Allow(key, config.Max) {
				return c.String(http.StatusTooManyRequests, config.Message)
			}

			return next(c)
		}
	}
}

type rateLimitStore struct {
	mu      sync.RWMutex
	entries map[string]*rateLimitEntry
	window  time.Duration
}

type rateLimitEntry struct {
	count    int
	expireAt time.Time
}

func newRateLimitStore(window time.Duration) *rateLimitStore {
	store := &rateLimitStore{
		entries: make(map[string]*rateLimitEntry),
		window:  window,
	}
	go store.cleanup()
	return store
}

func (s *rateLimitStore) Allow(key string, max int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.entries[key]

	if !exists || now.After(entry.expireAt) {
		s.entries[key] = &rateLimitEntry{
			count:    1,
			expireAt: now.Add(s.window),
		}
		return true
	}

	entry.count++
	return entry.count <= max
}

func (s *rateLimitStore) cleanup() {
	ticker := time.NewTicker(s.window)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.entries {
			if now.After(entry.expireAt) {
				delete(s.entries, key)
			}
		}
		s.mu.Unlock()
	}
}
