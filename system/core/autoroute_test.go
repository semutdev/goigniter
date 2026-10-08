package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type DashboardController struct {
	Controller
}

func (ctrl *DashboardController) Index(c *Context) error {
	return c.String(http.StatusOK, "dashboard index")
}

func (ctrl *DashboardController) FilterSummary(c *Context) error {
	return c.String(http.StatusOK, "filtered: "+c.Query("range"))
}

func (ctrl *DashboardController) AllowedMethods() map[string][]string {
	return map[string][]string{
		"Index":         {"GET"},
		"FilterSummary": {"GET", "POST"},
	}
}

type UserController struct {
	Controller
}

func (ctrl *UserController) Index() error {
	return ctrl.Ctx.String(http.StatusOK, "user index")
}

func (ctrl *UserController) Create() {
	ctrl.Ctx.String(http.StatusCreated, "user created")
}

func (ctrl *UserController) Routes() map[string]string {
	return map[string]string{
		"Create": "add",
	}
}

func TestAutoRoute_Comprehensive(t *testing.T) {
	ResetRegistry()

	app := New()

	// Register with prefix "admin"
	app.Register(&DashboardController{}, "admin")
	// Register without prefix
	app.Register(&UserController{})

	app.AutoRoute()

	tests := []struct {
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		// DashboardController: Suffix "Controller" stripped -> "dashboard", prefix "admin"
		{"GET", "/admin/dashboard", http.StatusOK, "dashboard index"},
		{"GET", "/admin/dashboard/index", http.StatusOK, "dashboard index"},
		// Snake case route automatically supported
		{"GET", "/admin/dashboard/filter_summary?range=today", http.StatusOK, "filtered: today"},
		// Lowercase route also supported
		{"GET", "/admin/dashboard/filtersummary?range=today", http.StatusOK, "filtered: today"},
		// Disallowed method (Index only allows GET)
		{"POST", "/admin/dashboard", http.StatusNotFound, ""},

		// UserController: Suffix "Controller" stripped -> "user", no prefix
		{"GET", "/user", http.StatusOK, "user index"},
		{"GET", "/user/index", http.StatusOK, "user index"},
		// Custom route mapping: "Create" mapped to "add"
		{"GET", "/user/add", http.StatusCreated, "user created"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d for %s %s", tt.expectedCode, rec.Code, tt.method, tt.path)
			}
			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q for %s %s", tt.expectedBody, rec.Body.String(), tt.method, tt.path)
			}
		})
	}
}
