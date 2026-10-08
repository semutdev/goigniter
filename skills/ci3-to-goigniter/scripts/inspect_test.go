package scripts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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

/**
 * Users controller documentation block
 * @package Application\Controllers
 */
class Users extends CI_Controller {
    public function __construct() {
        parent::__construct();
        $this->load->model('User_model');
    }

    // Commented out method:
    // public function deleted_action() { return false; }

    # Another commented out method:
    # public function hash_commented_action() {}

    /* Block commented out:
    public function old_method() {}
    */

    public function index() {
        $this->load->view('users/index');
    }

    public function detail($id) {
        $this->load->view('users/detail');
    }

    public static function create() {
        // create user logic
    }

    final public function export() {
        // export logic
    }

    protected function internal_helper() {
        // should not be detected as public action
    }

    private function secret_auth() {
        // should not be detected as public action
    }

    protected static function secret_static() {
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

	// Security stub index.html files (must be filtered out)
	securityStub1 := `<!DOCTYPE html><html><head><title>403 Forbidden</title></head><body><p>Directory access is forbidden.</p></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "application", "views", "index.html"), []byte(securityStub1), 0644); err != nil {
		t.Fatalf("failed to write root views index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(viewDir, "index.html"), []byte(securityStub1), 0644); err != nil {
		t.Fatalf("failed to write users index.html: %v", err)
	}

	// Legitimate html view (must NOT be filtered out)
	legitHtml := `<div>Custom HTML View</div>`
	if err := os.WriteFile(filepath.Join(viewDir, "custom.html"), []byte(legitHtml), 0644); err != nil {
		t.Fatalf("failed to write custom.html view: %v", err)
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

// Commented out route:
// $route['old_users'] = 'old_users/index';

$route[ 'users' ] = "users/index";
$route['users/(:num)'] = 'users/detail/$1';
$route[ "api/users" ][ 'post' ] = "users/create";
$route['catalog/([a-z]+|[\d]+)'] = 'catalog/view/$1';
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
	// Action methods: index, detail, create (public static), export (final public)
	// Commented methods and protected/private must NOT be included!
	expectedMethods := []string{"index", "detail", "create", "export"}
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

	// 3. Verify Views (index.php, detail.php, and custom.html; index.html security stubs MUST be skipped)
	if len(inv.Views) != 3 {
		t.Fatalf("expected 3 views, got %d (%v)", len(inv.Views), inv.Views)
	}
	for _, v := range inv.Views {
		if strings.HasSuffix(v, "index.html") {
			t.Errorf("expected security stub index.html to be excluded, but got: %s", v)
		}
	}

	// 4. Verify Routes
	foundUsersRoute := false
	foundParamRoute := false
	foundApiPostRoute := false
	foundPipeRoute := false
	foundDefaultCtrl := false
	found404 := false
	for _, r := range inv.Routes {
		if r.Pattern == "users" && r.Target == "users/index" {
			foundUsersRoute = true
		}
		if r.Pattern == "users/(:num)" && r.Target == "users/detail/$1" {
			foundParamRoute = true
		}
		if r.Pattern == "api/users" && r.HTTPMethod == "POST" && r.Target == "users/create" {
			foundApiPostRoute = true
		}
		if r.Pattern == "catalog/([a-z]+|[\\d]+)" && r.Target == "catalog/view/$1" {
			foundPipeRoute = true
		}
		if r.Pattern == "default_controller" {
			foundDefaultCtrl = true
			if !r.IsSpecial || !strings.Contains(r.Notes, "Default Controller") {
				t.Errorf("expected default_controller to be marked special with notes, got: %+v", r)
			}
		}
		if r.Pattern == "404_override" {
			found404 = true
			if !r.IsSpecial || !strings.Contains(r.Notes, "404 Override") {
				t.Errorf("expected 404_override to be marked special with notes, got: %+v", r)
			}
		}
		if r.Pattern == "old_users" {
			t.Errorf("commented route old_users was not stripped!")
		}
	}
	if !foundUsersRoute {
		t.Errorf("route 'users' -> 'users/index' not found in %v", inv.Routes)
	}
	if !foundParamRoute {
		t.Errorf("route 'users/(:num)' -> 'users/detail/$1' not found in %v", inv.Routes)
	}
	if !foundApiPostRoute {
		t.Errorf("route POST 'api/users' -> 'users/create' not found in %v", inv.Routes)
	}
	if !foundPipeRoute {
		t.Errorf("route with pipe 'catalog/([a-z]+|[\\d]+)' not found in %v", inv.Routes)
	}
	if !foundDefaultCtrl {
		t.Errorf("default_controller route not found in %v", inv.Routes)
	}
	if !found404 {
		t.Errorf("404_override route not found in %v", inv.Routes)
	}

	// 5. Verify Autoload
	if len(inv.Autoload.Libraries) == 0 {
		t.Errorf("expected autoload libraries to be detected")
	}

	// 6. Verify Markdown Report Generation with escaped pipes
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
	// Verify pipe escaping in markdown table
	if !strings.Contains(md, `catalog/([a-z]+\|[\d]+)`) {
		t.Errorf("expected markdown table to escape pipe character in route regex, got markdown:\n%s", md)
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

func TestStripPHPComments(t *testing.T) {
	input := `<?php
// Single line slash comment
# Single line hash comment
/* Multi line
   block comment */
/** Docblock */
$url = "http://example.com//test#hash";
$msg = 'It\'s /* not a comment */';
function test() {}
`
	cleaned := stripPHPComments(input)
	if strings.Contains(cleaned, "Single line slash comment") {
		t.Error("slash comment not stripped")
	}
	if strings.Contains(cleaned, "Single line hash comment") {
		t.Error("hash comment not stripped")
	}
	if strings.Contains(cleaned, "block comment") {
		t.Error("block comment not stripped")
	}
	if strings.Contains(cleaned, "Docblock") {
		t.Error("docblock not stripped")
	}
	if !strings.Contains(cleaned, `"http://example.com//test#hash"`) {
		t.Error("string literal content was corrupted")
	}
	if !strings.Contains(cleaned, `'It\'s /* not a comment */'`) {
		t.Error("escaped string literal content was corrupted")
	}
	if !strings.Contains(cleaned, "function test() {}") {
		t.Error("valid code was removed")
	}
}

func TestSortTieBreaker(t *testing.T) {
	inv := &ProjectInventory{
		Controllers: []ControllerInfo{
			{Name: "User", File: "controllers/v2/User.php"},
			{Name: "User", File: "controllers/v1/User.php"},
			{Name: "Auth", File: "controllers/Auth.php"},
		},
		Models: []ModelInfo{
			{Name: "User_model", File: "models/v2/User_model.php"},
			{Name: "User_model", File: "models/v1/User_model.php"},
		},
	}

	// Replicate tie-breaker sort
	ctrls := append([]ControllerInfo(nil), inv.Controllers...)
	sort.Slice(ctrls, func(i, j int) bool {
		if ctrls[i].Name == ctrls[j].Name {
			return ctrls[i].File < ctrls[j].File
		}
		return ctrls[i].Name < ctrls[j].Name
	})

	if ctrls[0].Name != "Auth" {
		t.Errorf("expected Auth first, got %s", ctrls[0].Name)
	}
	if ctrls[1].File != "controllers/v1/User.php" {
		t.Errorf("expected v1 User before v2 User, got %s", ctrls[1].File)
	}
	if ctrls[2].File != "controllers/v2/User.php" {
		t.Errorf("expected v2 User third, got %s", ctrls[2].File)
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

func TestRunInspectCLI(t *testing.T) {
	// 1. Test help flag
	err := RunInspectCLI([]string{"-h"})
	if err != nil {
		t.Errorf("expected nil error on -h, got %v", err)
	}

	// 2. Test running on mock project with JSON output to file
	mockDir := createMockCI3Project(t)
	outFile := filepath.Join(t.TempDir(), "inventory.json")

	err = RunInspectCLI([]string{"-src", mockDir, "-out", outFile, "-format", "json"})
	if err != nil {
		t.Fatalf("RunInspectCLI failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read out file: %v", err)
	}
	var inv ProjectInventory
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("failed to unmarshal output json: %v", err)
	}
	if len(inv.Controllers) != 1 || inv.Controllers[0].Name != "Users" {
		t.Errorf("unexpected output inventory: %+v", inv)
	}
}
