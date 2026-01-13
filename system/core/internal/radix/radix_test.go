package radix

import (
	"testing"
)

func TestTree_StaticRoutes(t *testing.T) {
	tree := New()
	tree.Insert("/", "root")
	tree.Insert("/users", "users")
	tree.Insert("/users/profile", "profile")

	tests := []struct {
		path     string
		expected string
		found    bool
	}{
		{"/", "root", true},
		{"/users", "users", true},
		{"/users/profile", "profile", true},
		{"/notfound", "", false},
	}

	for _, tt := range tests {
		handler, _, found := tree.Search(tt.path)
		if found != tt.found {
			t.Errorf("path %s: expected found=%v, got %v", tt.path, tt.found, found)
			continue
		}
		if found && handler.(string) != tt.expected {
			t.Errorf("path %s: expected handler=%s, got %s", tt.path, tt.expected, handler)
		}
	}
}

func TestTree_ParamRoutes(t *testing.T) {
	tree := New()
	tree.Insert("/users/:id", "user")
	tree.Insert("/users/:id/posts", "posts")
	tree.Insert("/users/:id/posts/:pid", "post")

	tests := []struct {
		path           string
		expected       string
		found          bool
		expectedParams map[string]string
	}{
		{"/users/123", "user", true, map[string]string{"id": "123"}},
		{"/users/456/posts", "posts", true, map[string]string{"id": "456"}},
		{"/users/789/posts/42", "post", true, map[string]string{"id": "789", "pid": "42"}},
	}

	for _, tt := range tests {
		handler, params, found := tree.Search(tt.path)
		if found != tt.found {
			t.Errorf("path %s: expected found=%v, got %v", tt.path, tt.found, found)
			continue
		}
		if found {
			if handler.(string) != tt.expected {
				t.Errorf("path %s: expected handler=%s, got %s", tt.path, tt.expected, handler)
			}
			for k, v := range tt.expectedParams {
				if params[k] != v {
					t.Errorf("path %s: expected param %s=%s, got %s", tt.path, k, v, params[k])
				}
			}
		}
	}
}

func TestTree_WildcardRoutes(t *testing.T) {
	tree := New()
	tree.Insert("/files/*filepath", "files")

	tests := []struct {
		path           string
		expected       string
		found          bool
		expectedParams map[string]string
	}{
		{"/files/doc.txt", "files", true, map[string]string{"filepath": "doc.txt"}},
		{"/files/images/photo.jpg", "files", true, map[string]string{"filepath": "images/photo.jpg"}},
		{"/files/a/b/c/d.txt", "files", true, map[string]string{"filepath": "a/b/c/d.txt"}},
	}

	for _, tt := range tests {
		handler, params, found := tree.Search(tt.path)
		if found != tt.found {
			t.Errorf("path %s: expected found=%v, got %v", tt.path, tt.found, found)
			continue
		}
		if found {
			if handler.(string) != tt.expected {
				t.Errorf("path %s: expected handler=%s, got %s", tt.path, tt.expected, handler)
			}
			for k, v := range tt.expectedParams {
				if params[k] != v {
					t.Errorf("path %s: expected param %s=%s, got %s", tt.path, k, v, params[k])
				}
			}
		}
	}
}

func TestTree_Priority(t *testing.T) {
	tree := New()
	tree.Insert("/users/new", "new")
	tree.Insert("/users/:id", "user")

	// Static route should have higher priority
	handler, _, found := tree.Search("/users/new")
	if !found || handler.(string) != "new" {
		t.Errorf("static route should take priority, got: %v, %v", handler, found)
	}

	handler, params, found := tree.Search("/users/123")
	if !found || handler.(string) != "user" || params["id"] != "123" {
		t.Errorf("param route should match, got: %v, %v, %v", handler, params, found)
	}
}
