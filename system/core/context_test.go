package core

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	err := ctx.JSON(200, Map{"message": "hello"})
	if err != nil {
		t.Fatal(err)
	}

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Expected JSON content type, got %s", rec.Header().Get("Content-Type"))
	}

	if !strings.Contains(rec.Body.String(), `"message":"hello"`) {
		t.Errorf("Expected JSON body, got %s", rec.Body.String())
	}
}

func TestContext_String(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	ctx.String(200, "Hello World")

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", rec.Body.String())
	}
}

func TestContext_HTML(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	ctx.HTML(200, "<h1>Hello</h1>")

	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected HTML content type, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestContext_Redirect(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	ctx.Redirect(302, "/new-location")

	if rec.Code != 302 {
		t.Errorf("Expected status 302, got %d", rec.Code)
	}

	if rec.Header().Get("Location") != "/new-location" {
		t.Errorf("Expected redirect to /new-location, got %s", rec.Header().Get("Location"))
	}
}

func TestContext_Param(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	ctx.params["id"] = "123"
	ctx.params["name"] = "john"
	defer releaseContext(ctx)

	if ctx.Param("id") != "123" {
		t.Errorf("Expected id=123, got %s", ctx.Param("id"))
	}

	if ctx.Param("name") != "john" {
		t.Errorf("Expected name=john, got %s", ctx.Param("name"))
	}

	id, err := ctx.ParamInt("id")
	if err != nil || id != 123 {
		t.Errorf("Expected id=123 as int, got %d, err=%v", id, err)
	}
}

func TestContext_Query(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?page=5&limit=10", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	if ctx.Query("page") != "5" {
		t.Errorf("Expected page=5, got %s", ctx.Query("page"))
	}

	if ctx.Query("limit") != "10" {
		t.Errorf("Expected limit=10, got %s", ctx.Query("limit"))
	}

	if ctx.QueryDefault("missing", "default") != "default" {
		t.Errorf("Expected default value")
	}

	page, err := ctx.QueryInt("page")
	if err != nil || page != 5 {
		t.Errorf("Expected page=5 as int, got %d", page)
	}

	missing := ctx.QueryIntDefault("missing", 100)
	if missing != 100 {
		t.Errorf("Expected default 100, got %d", missing)
	}
}

func TestContext_Bind(t *testing.T) {
	body := `{"name":"John","age":30}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := ctx.Bind(&data)
	if err != nil {
		t.Fatal(err)
	}

	if data.Name != "John" || data.Age != 30 {
		t.Errorf("Expected {John, 30}, got {%s, %d}", data.Name, data.Age)
	}
}

func TestContext_Body(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader("raw body content"))
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	body, err := ctx.Body()
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "raw body content" {
		t.Errorf("Expected 'raw body content', got %s", string(body))
	}
}

func TestContext_Headers(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	if ctx.Header("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected custom-value, got %s", ctx.Header("X-Custom-Header"))
	}

	ctx.SetHeader("X-Response-Header", "response-value")
	if rec.Header().Get("X-Response-Header") != "response-value" {
		t.Errorf("Expected response-value, got %s", rec.Header().Get("X-Response-Header"))
	}
}

func TestContext_Cookie(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	cookie, err := ctx.Cookie("session")
	if err != nil || cookie.Value != "abc123" {
		t.Errorf("Expected session=abc123, got %v, err=%v", cookie, err)
	}

	ctx.SetCookie(&http.Cookie{Name: "new_cookie", Value: "new_value"})
	// Check if cookie was set in response
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "new_cookie" && c.Value == "new_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected new_cookie to be set")
	}
}

func TestContext_Store(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	ctx.Set("key", "value")
	ctx.Set("number", 42)

	if ctx.Get("key") != "value" {
		t.Errorf("Expected 'value', got %v", ctx.Get("key"))
	}

	if ctx.GetString("key") != "value" {
		t.Errorf("Expected 'value' as string")
	}

	if ctx.GetInt("number") != 42 {
		t.Errorf("Expected 42, got %d", ctx.GetInt("number"))
	}

	if ctx.Get("missing") != nil {
		t.Error("Expected nil for missing key")
	}
}

func TestContext_RequestInfo(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/create", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	if ctx.Method() != "POST" {
		t.Errorf("Expected POST, got %s", ctx.Method())
	}

	if ctx.Path() != "/users/create" {
		t.Errorf("Expected /users/create, got %s", ctx.Path())
	}

	if ctx.IP() != "192.168.1.1:12345" {
		t.Errorf("Expected 192.168.1.1:12345, got %s", ctx.IP())
	}
}

func TestContext_IP_XForwardedFor(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.1:12345"
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	if ctx.IP() != "10.0.0.1" {
		t.Errorf("Expected 10.0.0.1 from X-Forwarded-For, got %s", ctx.IP())
	}
}

func TestContext_Blob(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
	ctx.Blob(200, "image/png", data)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Expected image/png, got %s", rec.Header().Get("Content-Type"))
	}

	body, _ := io.ReadAll(rec.Body)
	if !bytes.Equal(body, data) {
		t.Error("Body mismatch")
	}
}

func TestContext_NoContent(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/", nil)
	ctx := acquireContext(rec, req, nil)
	defer releaseContext(ctx)

	ctx.NoContent(204)

	if rec.Code != 204 {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}

	if rec.Body.Len() != 0 {
		t.Error("Expected empty body")
	}
}
