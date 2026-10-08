package core

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplication_BasicRouting(t *testing.T) {
	app := New()

	app.GET("/", func(c *Context) error {
		return c.String(200, "Hello World")
	})

	app.GET("/users/:id", func(c *Context) error {
		return c.String(200, "User: "+c.Param("id"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
		status   int
	}{
		{"GET", "/", "Hello World", 200},
		{"GET", "/users/123", "User: 123", 200},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		if rec.Code != tt.status {
			t.Errorf("%s %s: expected status %d, got %d", tt.method, tt.path, tt.status, rec.Code)
		}

		if rec.Body.String() != tt.expected {
			t.Errorf("%s %s: expected body %q, got %q", tt.method, tt.path, tt.expected, rec.Body.String())
		}
	}
}

func TestApplication_HTTPMethods(t *testing.T) {
	app := New()

	app.GET("/resource", func(c *Context) error { return c.String(200, "GET") })
	app.POST("/resource", func(c *Context) error { return c.String(200, "POST") })
	app.PUT("/resource", func(c *Context) error { return c.String(200, "PUT") })
	app.DELETE("/resource", func(c *Context) error { return c.String(200, "DELETE") })
	app.PATCH("/resource", func(c *Context) error { return c.String(200, "PATCH") })

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/resource", nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)

		if rec.Body.String() != method {
			t.Errorf("Expected %s, got %s", method, rec.Body.String())
		}
	}
}

func TestApplication_Group(t *testing.T) {
	app := New()

	api := app.Group("/api")
	api.GET("/users", func(c *Context) error {
		return c.String(200, "API Users")
	})

	v1 := api.Group("/v1")
	v1.GET("/posts", func(c *Context) error {
		return c.String(200, "V1 Posts")
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/users", "API Users"},
		{"/api/v1/posts", "V1 Posts"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, req)

		if rec.Body.String() != tt.expected {
			t.Errorf("Path %s: expected %q, got %q", tt.path, tt.expected, rec.Body.String())
		}
	}
}

func TestApplication_NotFound(t *testing.T) {
	app := New()
	app.GET("/", func(c *Context) error { return c.String(200, "OK") })

	req := httptest.NewRequest("GET", "/notfound", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestApplication_Middleware(t *testing.T) {
	app := New()

	// Global middleware
	app.Use(func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			c.Set("middleware", "executed")
			return next(c)
		}
	})

	app.GET("/", func(c *Context) error {
		val := c.GetString("middleware")
		return c.String(200, val)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Body.String() != "executed" {
		t.Errorf("Middleware not executed, got: %s", rec.Body.String())
	}
}

func TestTemplateEngine_GlobalParse(t *testing.T) {
	dir := t.TempDir()

	// 1. Template with explicit {{define}} block (e.g. auth/forgot.html)
	if err := os.MkdirAll(filepath.Join(dir, "auth"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "auth", "forgot.html"), []byte(`{{define "auth/forgot"}}<!DOCTYPE html><html><body>Forgot: {{.Title}}</body></html>{{end}}`), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Bare template without {{define}} (e.g. welcome.html)
	if err := os.WriteFile(filepath.Join(dir, "welcome.html"), []byte(`<!DOCTYPE html><html><body>Welcome: {{.Title}}</body></html>`), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Layout and partials with cross-file references (e.g. admin/dashboard and admin/dashboard/partials/_summary)
	if err := os.MkdirAll(filepath.Join(dir, "admin", "dashboard", "partials"), 0755); err != nil {
		t.Fatal(err)
	}
	summaryPartial := `<div class="summary-cards">Cards: {{.Total}}</div>`
	if err := os.WriteFile(filepath.Join(dir, "admin", "dashboard", "partials", "_summary.html"), []byte(summaryPartial), 0644); err != nil {
		t.Fatal(err)
	}

	dashboardHTML := `<main>Dashboard: {{.Title}} - {{ template "admin/dashboard/partials/_summary.html" . }}</main>`
	if err := os.WriteFile(filepath.Join(dir, "admin", "dashboard", "index.html"), []byte(dashboardHTML), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. POS page with leading-slash partial inclusion (e.g. /pos/partials/_grid.html)
	if err := os.MkdirAll(filepath.Join(dir, "pos", "partials"), 0755); err != nil {
		t.Fatal(err)
	}
	gridPartial := `<div class="grid">Grid Items: {{.Count}}</div>`
	if err := os.WriteFile(filepath.Join(dir, "pos", "partials", "_grid.html"), []byte(gridPartial), 0644); err != nil {
		t.Fatal(err)
	}
	posHTML := `<div class="pos">{{ template "/pos/partials/_grid.html" . }}</div>`
	if err := os.WriteFile(filepath.Join(dir, "pos", "index.html"), []byte(posHTML), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := NewTemplateEngine(TemplateConfig{
		Dir:    dir,
		Ext:    ".html",
		Reload: false,
	})
	if err != nil {
		t.Fatalf("NewTemplateEngine failed: %v", err)
	}

	tests := []struct {
		name     string
		data     Map
		contains []string
	}{
		{"auth/forgot", Map{"Title": "Pass"}, []string{"Forgot: Pass"}},
		{"auth/forgot.html", Map{"Title": "Pass"}, []string{"Forgot: Pass"}},
		{"welcome", Map{"Title": "GoIgniter"}, []string{"Welcome: GoIgniter"}},
		{"welcome.html", Map{"Title": "GoIgniter"}, []string{"Welcome: GoIgniter"}},
		{"admin/dashboard/index", Map{"Title": "Admin", "Total": 42}, []string{"Dashboard: Admin", "Cards: 42"}},
		{"admin/dashboard/index.html", Map{"Title": "Admin", "Total": 42}, []string{"Dashboard: Admin", "Cards: 42"}},
		{"pos/index", Map{"Count": 99}, []string{"Grid Items: 99"}},
		{"/pos/index", Map{"Count": 99}, []string{"Grid Items: 99"}},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		err := engine.Render(&buf, tt.name, tt.data)
		if err != nil {
			t.Errorf("Render(%q) error: %v", tt.name, err)
			continue
		}
		result := buf.String()
		for _, exp := range tt.contains {
			if !strings.Contains(result, exp) {
				t.Errorf("Render(%q) = %q, expected to contain %q", tt.name, result, exp)
			}
		}
	}
}
