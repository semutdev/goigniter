package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/semutdev/goigniter/system/core"
)

func TestRateLimitAllowsUpToMax(t *testing.T) {
	app := core.New()
	app.Use(RateLimit(3, 0)) // 3 requests per window (default 1 minute)
	app.GET("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Errorf("Request %d should be allowed, got %d", i+1, w.Code)
		}
	}

	// 4th request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request should be blocked, got %d", w.Code)
	}
}

func TestRateLimitDifferentKeys(t *testing.T) {
	app := core.New()
	app.Use(RateLimit(2, 0))
	app.GET("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	// IP 1 should be allowed 2 times
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Errorf("IP 1 request %d should be allowed", i+1)
		}
	}

	// IP 2 should also be allowed (different key)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code == http.StatusTooManyRequests {
		t.Errorf("IP 2 request should be allowed (different key)")
	}
}

func TestSecurityHeaders(t *testing.T) {
	app := core.New()
	app.Use(SecurityHeaders())
	app.GET("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Check XSS Protection
	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("X-XSS-Protection header not set correctly")
	}

	// Check Content-Type-Options
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options header not set correctly")
	}

	// Check X-Frame-Options
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("X-Frame-Options header not set correctly")
	}

	// Check Referrer-Policy
	if w.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy header not set correctly")
	}
}

func TestCSRFTokenGeneration(t *testing.T) {
	app := core.New()
	app.Use(CSRF())
	app.GET("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// GET should set CSRF cookie
	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" || !containsSubstring(cookie, "csrf_token") {
		t.Errorf("CSRF cookie not set on GET request")
	}
}

func TestCSRFValidation(t *testing.T) {
	app := core.New()
	app.Use(CSRF())
	app.POST("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	// POST without token should fail
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF token should fail, got %d", w.Code)
	}
}

func TestCORSDefaultEmptyOrigins(t *testing.T) {
	config := DefaultCORSConfig()
	if len(config.AllowOrigins) != 0 {
		t.Errorf("DefaultCORSConfig should have empty AllowOrigins, got %v", config.AllowOrigins)
	}
}

func TestCORSWithCredentials(t *testing.T) {
	app := core.New()
	app.Use(CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST"},
	}))

	app.GET("/", func(c *core.Context) error {
		return c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Should reflect origin, not use "*"
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("With credentials, origin should be reflected, got %s", origin)
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Allow-Credentials header not set")
	}
}

func TestRecoveryHideErrorDetails(t *testing.T) {
	app := core.New()
	app.Use(RecoveryWithConfig(RecoveryConfig{
		HideErrorDetails:   true,
		ProductionMessage:  "Something went wrong",
		DisablePrintStack:  true,
	}))
	app.GET("/", func(c *core.Context) error {
		panic("sensitive error with password=secret123")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	body := w.Body.String()
	if containsSubstring(body, "secret123") {
		t.Errorf("Error details should be hidden in production mode")
	}
	if body != "Something went wrong" {
		t.Errorf("Should show production message, got: %s", body)
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}