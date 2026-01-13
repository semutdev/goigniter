package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goigniter/system/core"
)

func TestLogger(t *testing.T) {
	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	middleware := Logger()
	wrapped := middleware(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := createTestContext(rec, req)

	err := wrapped(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != 200 {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestRecovery(t *testing.T) {
	handler := func(c *core.Context) error {
		panic("test panic")
	}

	middleware := Recovery()
	wrapped := middleware(handler)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := createTestContext(rec, req)

	// Should not panic
	wrapped(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestCORS(t *testing.T) {
	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	middleware := CORS()
	wrapped := middleware(handler)

	// Test preflight request
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	ctx := createTestContext(rec, req)

	wrapped(ctx)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header")
	}

	// Test actual request
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Origin", "http://example.com")
	ctx2 := createTestContext(rec2, req2)

	wrapped(ctx2)

	if rec2.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS header on GET request")
	}
}

func TestCORSWithConfig(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"http://allowed.com"},
		AllowMethods:     []string{"GET", "POST"},
		AllowCredentials: true,
	}

	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	middleware := CORSWithConfig(config)
	wrapped := middleware(handler)

	// Test allowed origin
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://allowed.com")
	ctx := createTestContext(rec, req)

	wrapped(ctx)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("Expected allowed origin, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}

	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected credentials header")
	}

	// Test disallowed origin
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Origin", "http://notallowed.com")
	ctx2 := createTestContext(rec2, req2)

	wrapped(ctx2)

	if rec2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set CORS header for disallowed origin")
	}
}

func TestRateLimit(t *testing.T) {
	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	// Allow only 2 requests per second
	middleware := RateLimit(2, time.Second)
	wrapped := middleware(handler)

	// First two requests should pass
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		ctx := createTestContext(rec, req)

		wrapped(ctx)

		if rec.Code != 200 {
			t.Errorf("Request %d should pass, got %d", i+1, rec.Code)
		}
	}

	// Third request should be rate limited
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	ctx := createTestContext(rec, req)

	wrapped(ctx)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Third request should be rate limited, got %d", rec.Code)
	}
}

func TestBasicAuth(t *testing.T) {
	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	middleware := BasicAuth(func(user, pass string) bool {
		return user == "admin" && pass == "secret"
	})
	wrapped := middleware(handler)

	// Test without auth
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := createTestContext(rec, req)

	wrapped(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", rec.Code)
	}

	// Test with correct auth
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.SetBasicAuth("admin", "secret")
	ctx2 := createTestContext(rec2, req2)

	wrapped(ctx2)

	if rec2.Code != 200 {
		t.Errorf("Expected 200 with correct auth, got %d", rec2.Code)
	}

	// Test with wrong auth
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.SetBasicAuth("admin", "wrong")
	ctx3 := createTestContext(rec3, req3)

	wrapped(ctx3)

	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with wrong auth, got %d", rec3.Code)
	}
}

func TestBearerAuth(t *testing.T) {
	handler := func(c *core.Context) error {
		return c.String(200, "OK")
	}

	middleware := BearerAuth(func(token string) bool {
		return token == "valid-token"
	})
	wrapped := middleware(handler)

	// Test without auth
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := createTestContext(rec, req)

	wrapped(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 without auth, got %d", rec.Code)
	}

	// Test with valid token
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "Bearer valid-token")
	ctx2 := createTestContext(rec2, req2)

	wrapped(ctx2)

	if rec2.Code != 200 {
		t.Errorf("Expected 200 with valid token, got %d", rec2.Code)
	}

	// Test with invalid token
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("Authorization", "Bearer invalid-token")
	ctx3 := createTestContext(rec3, req3)

	wrapped(ctx3)

	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 with invalid token, got %d", rec3.Code)
	}
}

// createTestContext creates a context for testing middleware.
func createTestContext(rec *httptest.ResponseRecorder, req *http.Request) *core.Context {
	return &core.Context{
		Request:  req,
		Response: rec,
	}
}
