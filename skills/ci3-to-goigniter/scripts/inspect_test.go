package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createMockCI3Project(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Controllers
	ctrlDir := filepath.Join(dir, "application", "controllers")
	if err := os.MkdirAll(ctrlDir, 0755); err != nil {
		t.Fatalf("failed to create controllers dir: %v", err)
	}

	usersCtrl := `<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class Users extends CI_Controller {
    public function __construct() {
        parent::__construct();
        $this->load->model('User_model');
    }

    public function index() {
        $this->load->view('users/index');
    }

    public function detail($id) {
        $this->load->view('users/detail');
    }

    public function create() {
        // create user logic
    }

    protected function internal_helper() {
        // should not be detected as public action
    }

    private function secret_auth() {
        // should not be detected as public action
    }
}
`
	if err := os.WriteFile(filepath.Join(ctrlDir, "Users.php"), []byte(usersCtrl), 0644); err != nil {
		t.Fatalf("failed to write Users.php: %v", err)
	}

	// Models
	modelDir := filepath.Join(dir, "application", "models")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create models dir: %v", err)
	}

	userModel := `<?php
defined('BASEPATH') OR exit('No direct script access allowed');

class User_model extends CI_Model {
    protected $table = 'users';

    public function get_all() {
        return $this->db->get($this->table)->result();
    }

    public function get_by_id($id) {
        return $this->db->where('id', $id)->get($this->table)->row();
    }
}
`
	if err := os.WriteFile(filepath.Join(modelDir, "User_model.php"), []byte(userModel), 0644); err != nil {
		t.Fatalf("failed to write User_model.php: %v", err)
	}

	// Views
	viewDir := filepath.Join(dir, "application", "views", "users")
	if err := os.MkdirAll(viewDir, 0755); err != nil {
		t.Fatalf("failed to create views dir: %v", err)
	}

	indexView := `<h1>Users List</h1>`
	detailView := `<h1>User Detail</h1>`

	if err := os.WriteFile(filepath.Join(viewDir, "index.php"), []byte(indexView), 0644); err != nil {
		t.Fatalf("failed to write index.php view: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viewDir, "detail.php"), []byte(detailView), 0644); err != nil {
		t.Fatalf("failed to write detail.php view: %v", err)
	}

	// Config / Routes
	configDir := filepath.Join(dir, "application", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	routesContent := `<?php
defined('BASEPATH') OR exit('No direct script access allowed');

$route['default_controller'] = 'welcome';
$route['404_override'] = '';
$route['translate_uri_dashes'] = FALSE;

$route['users'] = 'users/index';
$route['users/(:num)'] = 'users/detail/$1';
$route['api/users']['post'] = 'users/create';
`
	if err := os.WriteFile(filepath.Join(configDir, "routes.php"), []byte(routesContent), 0644); err != nil {
		t.Fatalf("failed to write routes.php: %v", err)
	}

	// Autoload
	autoloadContent := `<?php
$autoload['libraries'] = array('database', 'session');
$autoload['helper'] = array('url', 'file');
$autoload['model'] = array('User_model');
`
	if err := os.WriteFile(filepath.Join(configDir, "autoload.php"), []byte(autoloadContent), 0644); err != nil {
		t.Fatalf("failed to write autoload.php: %v", err)
	}

	return dir
}

func TestInspect(t *testing.T) {
	mockDir := createMockCI3Project(t)

	inv, err := InspectProject(mockDir)
	if err != nil {
		t.Fatalf("InspectProject failed: %v", err)
	}
	if inv == nil {
		t.Fatal("expected non-nil ProjectInventory")
	}

	// 1. Verify Controllers
	if len(inv.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(inv.Controllers))
	}
	ctrl := inv.Controllers[0]
	if ctrl.Name != "Users" {
		t.Errorf("expected controller name 'Users', got '%s'", ctrl.Name)
	}
	if !strings.HasSuffix(ctrl.File, "Users.php") {
		t.Errorf("expected controller file to end with 'Users.php', got '%s'", ctrl.File)
	}
	expectedMethods := []string{"index", "detail", "create"}
	if len(ctrl.Methods) != len(expectedMethods) {
		t.Fatalf("expected %d methods %v, got %d %v", len(expectedMethods), expectedMethods, len(ctrl.Methods), ctrl.Methods)
	}
	for i, m := range expectedMethods {
		if ctrl.Methods[i] != m {
			t.Errorf("expected method[%d] == %s, got %s", i, m, ctrl.Methods[i])
		}
	}

	// 2. Verify Models
	if len(inv.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(inv.Models))
	}
	model := inv.Models[0]
	if model.Name != "User_model" {
		t.Errorf("expected model name 'User_model', got '%s'", model.Name)
	}
	if !strings.HasSuffix(model.File, "User_model.php") {
		t.Errorf("expected model file to end with 'User_model.php', got '%s'", model.File)
	}

	// 3. Verify Views
	if len(inv.Views) != 2 {
		t.Fatalf("expected 2 views, got %d (%v)", len(inv.Views), inv.Views)
	}
	hasIndexView := false
	hasDetailView := false
	for _, v := range inv.Views {
		if strings.Contains(v, "users/index.php") {
			hasIndexView = true
		}
		if strings.Contains(v, "users/detail.php") {
			hasDetailView = true
		}
	}
	if !hasIndexView {
		t.Errorf("expected views to include 'users/index.php', got %v", inv.Views)
	}
	if !hasDetailView {
		t.Errorf("expected views to include 'users/detail.php', got %v", inv.Views)
	}

	// 4. Verify Routes
	// Expect at least 'users' and 'users/(:num)'
	foundUsersRoute := false
	foundParamRoute := false
	for _, r := range inv.Routes {
		if r.Pattern == "users" && r.Target == "users/index" {
			foundUsersRoute = true
		}
		if r.Pattern == "users/(:num)" && r.Target == "users/detail/$1" {
			foundParamRoute = true
		}
	}
	if !foundUsersRoute {
		t.Errorf("route 'users' -> 'users/index' not found in %v", inv.Routes)
	}
	if !foundParamRoute {
		t.Errorf("route 'users/(:num)' -> 'users/detail/$1' not found in %v", inv.Routes)
	}

	// 5. Verify Autoload
	if len(inv.Autoload.Libraries) == 0 {
		t.Errorf("expected autoload libraries to be detected")
	}

	// 6. Verify Markdown Report Generation
	md := inv.GenerateMarkdown()
	if !strings.Contains(md, "# CodeIgniter 3 Migration Inventory Report") {
		t.Errorf("markdown report missing header")
	}
	if !strings.Contains(md, "Users") || !strings.Contains(md, "User_model") {
		t.Errorf("markdown report missing Users or User_model")
	}
	if !strings.Contains(md, "users/(:num)") {
		t.Errorf("markdown report missing route users/(:num)")
	}

	// 7. Verify JSON Output
	jsonBytes, err := inv.GenerateJSON()
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}
	var parsed ProjectInventory
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal generated JSON: %v", err)
	}
	if len(parsed.Controllers) != 1 || parsed.Controllers[0].Name != "Users" {
		t.Errorf("parsed JSON mismatch: %+v", parsed)
	}
}

func TestInspectInvalidDir(t *testing.T) {
	_, err := InspectProject("non_existent_directory_xyz_12345")
	if err == nil {
		t.Error("expected error for non-existent directory, got nil")
	}
}

func TestInspectNestedAndDirectAppDir(t *testing.T) {
	dir := t.TempDir()

	// Nested controller under admin/
	adminDir := filepath.Join(dir, "controllers", "admin")
	if err := os.MkdirAll(adminDir, 0755); err != nil {
		t.Fatal(err)
	}
	dashCtrl := `<?php
class Dashboard extends CI_Controller {
    public function stats() {}
}
`
	if err := os.WriteFile(filepath.Join(adminDir, "Dashboard.php"), []byte(dashCtrl), 0644); err != nil {
		t.Fatal(err)
	}

	// Passing application dir directly as srcDir
	inv, err := InspectProject(dir)
	if err != nil {
		t.Fatalf("InspectProject failed: %v", err)
	}
	if len(inv.Controllers) != 1 {
		t.Fatalf("expected 1 controller, got %d", len(inv.Controllers))
	}
	if inv.Controllers[0].Name != "Dashboard" {
		t.Errorf("expected controller name 'Dashboard', got '%s'", inv.Controllers[0].Name)
	}
	if len(inv.Controllers[0].Methods) != 1 || inv.Controllers[0].Methods[0] != "stats" {
		t.Errorf("expected methods [stats], got %v", inv.Controllers[0].Methods)
	}
}

func TestGenerateMarkdownEmpty(t *testing.T) {
	inv := &ProjectInventory{
		SourceDir:   "/empty",
		Controllers: nil,
		Models:      nil,
		Views:       nil,
		Routes:      nil,
	}
	md := inv.GenerateMarkdown()
	if !strings.Contains(md, "_No controllers found._") {
		t.Error("expected empty controllers message")
	}
	if !strings.Contains(md, "_No models found._") {
		t.Error("expected empty models message")
	}
	if !strings.Contains(md, "_No views found._") {
		t.Error("expected empty views message")
	}
	if !strings.Contains(md, "_No custom routes found._") {
		t.Error("expected empty routes message")
	}
}
