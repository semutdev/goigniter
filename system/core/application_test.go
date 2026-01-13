package core

import (
	"net/http"
	"net/http/httptest"
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
