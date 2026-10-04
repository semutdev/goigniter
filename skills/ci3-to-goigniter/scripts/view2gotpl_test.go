package scripts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestView2GoTpl_VariableEchoes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short echo variable",
			input:    "<?= $title ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "php echo with semicolon",
			input:    "<?php echo $title; ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "short echo with htmlspecialchars",
			input:    "<?= htmlspecialchars($title) ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "htmlspecialchars with ENT_QUOTES and charset",
			input:    "<?= htmlspecialchars($title, ENT_QUOTES, 'UTF-8') ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "php echo htmlspecialchars",
			input:    "<?php echo htmlspecialchars($title); ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "htmlentities function",
			input:    "<?= htmlentities($title) ?>",
			expected: "{{ .Title }}",
		},
		{
			name:     "embedded inside html tag",
			input:    "<h1><?= $title ?></h1>",
			expected: "<h1>{{ .Title }}</h1>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_ObjectPropertiesAndArrayAccess(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "object property",
			input:    "<?= $user->email ?>",
			expected: "{{ .User.Email }}",
		},
		{
			name:     "array access single quotes",
			input:    "<?= $user['name'] ?>",
			expected: "{{ .User.Name }}",
		},
		{
			name:     "array access double quotes",
			input:    "<?= $user[\"name\"] ?>",
			expected: "{{ .User.Name }}",
		},
		{
			name:     "snake_case property conversion to CamelCase with acronym",
			input:    "<?= $order_item->unit_price ?>",
			expected: "{{ .OrderItem.UnitPrice }}",
		},
		{
			name:     "nested object properties",
			input:    "<?= $user->profile->avatar ?>",
			expected: "{{ .User.Profile.Avatar }}",
		},
		{
			name:     "nested mixed array and object",
			input:    "<?= $data['user']->email ?>",
			expected: "{{ .Data.User.Email }}",
		},
		{
			name:     "id initialism uppercase",
			input:    "<?= $user->id ?>",
			expected: "{{ .User.ID }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_Conditionals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple if endif",
			input:    "<?php if ($logged_in): ?> ... <?php endif; ?>",
			expected: "{{ if .LoggedIn }} ... {{ end }}",
		},
		{
			name:     "if object property with else and endif",
			input:    "<?php if ($user->is_admin): ?> ... <?php else: ?> ... <?php endif; ?>",
			expected: "{{ if .User.IsAdmin }} ... {{ else }} ... {{ end }}",
		},
		{
			name:     "elseif",
			input:    "<?php elseif ($x): ?>",
			expected: "{{ else if .X }}",
		},
		{
			name:     "else if alternate syntax",
			input:    "<?php else if ($x): ?>",
			expected: "{{ else if .X }}",
		},
		{
			name:     "negated condition",
			input:    "<?php if (!$logged_in): ?> Login <?php endif; ?>",
			expected: "{{ if not .LoggedIn }} Login {{ end }}",
		},
		{
			name:     "empty check",
			input:    "<?php if (!empty($users)): ?> Show <?php endif; ?>",
			expected: "{{ if .Users }} Show {{ end }}",
		},
		{
			name:     "brace syntax if else",
			input:    "<?php if ($logged_in) { ?> Welcome <?php } else { ?> Guest <?php } ?>",
			expected: "{{ if .LoggedIn }} Welcome {{ else }} Guest {{ end }}",
		},
		{
			name:     "comparison condition",
			input:    "<?php if ($count > 0): ?> Has items <?php endif; ?>",
			expected: "{{ if gt .Count 0 }} Has items {{ end }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_Loops(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "foreach with object property inside",
			input:    "<?php foreach ($users as $user): ?> <span><?= $user->name ?></span> <?php endforeach; ?>",
			expected: "{{ range .Users }} <span>{{ .Name }}</span> {{ end }}",
		},
		{
			name:     "foreach with array property inside",
			input:    "<?php foreach ($users as $user): ?> <span><?= $user['email'] ?></span> <?php endforeach; ?>",
			expected: "{{ range .Users }} <span>{{ .Email }}</span> {{ end }}",
		},
		{
			name:     "foreach with direct item reference",
			input:    "<?php foreach ($tags as $tag): ?> <li><?= $tag ?></li> <?php endforeach; ?>",
			expected: "{{ range .Tags }} <li>{{ . }}</li> {{ end }}",
		},
		{
			name:     "foreach brace syntax",
			input:    "<?php foreach ($items as $item) { ?> <div><?= $item->title ?></div> <?php } ?>",
			expected: "{{ range .Items }} <div>{{ .Title }}</div> {{ end }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_CommonCI3Helpers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "base_url with asset path",
			input:    "<?= base_url('css/main.css') ?>",
			expected: "/css/main.css",
		},
		{
			name:     "base_url empty root",
			input:    "<?= base_url() ?>",
			expected: "/",
		},
		{
			name:     "base_url empty string",
			input:    "<?= base_url('') ?>",
			expected: "/",
		},
		{
			name:     "base_url double quotes",
			input:    "<?= base_url(\"js/app.js\") ?>",
			expected: "/js/app.js",
		},
		{
			name:     "php echo base_url",
			input:    "<?php echo base_url('images/logo.png'); ?>",
			expected: "/images/logo.png",
		},
		{
			name:     "site_url with concatenation and variable",
			input:    "<?= site_url('users/detail/' . $id) ?>",
			expected: "/users/detail/{{ .ID }}",
		},
		{
			name:     "site_url static path",
			input:    "<?= site_url('users') ?>",
			expected: "/users",
		},
		{
			name:     "site_url empty root",
			input:    "<?= site_url() ?>",
			expected: "/",
		},
		{
			name:     "form_open single action",
			input:    "<?php echo form_open('login'); ?>",
			expected: "<form action=\"/login\" method=\"POST\">",
		},
		{
			name:     "form_open short echo",
			input:    "<?= form_open('login') ?>",
			expected: "<form action=\"/login\" method=\"POST\">",
		},
		{
			name:     "form_open_multipart",
			input:    "<?php echo form_open_multipart('users/upload'); ?>",
			expected: "<form action=\"/users/upload\" method=\"POST\" enctype=\"multipart/form-data\">",
		},
		{
			name:     "form_close",
			input:    "<?php echo form_close(); ?>",
			expected: "</form>",
		},
		{
			name:     "form_close short tag",
			input:    "<?= form_close() ?>",
			expected: "</form>",
		},
		{
			name:     "csrf token name helper",
			input:    "<?php echo $this->security->get_csrf_token_name(); ?>",
			expected: "{{ .csrf_token }}",
		},
		{
			name:     "csrf hash helper",
			input:    "<?= $this->security->get_csrf_hash() ?>",
			expected: "{{ .csrf_token }}",
		},
		{
			name:     "ci3 direct script access check stripped",
			input:    "<?php defined('BASEPATH') OR exit('No direct script access allowed'); ?>",
			expected: "",
		},
		{
			name:     "view partial include",
			input:    "<?php $this->load->view('partials/header'); ?>",
			expected: "{{ template \"partials/header.html\" . }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_UnconvertiblePHP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "arbitrary complex php statement",
			input:    "<?php $calc = compute_hash($a, $b); ?>",
			expected: "{{/* TODO_MIGRATE: $calc = compute_hash($a, $b); */}}",
		},
		{
			name:     "arbitrary multi statement block",
			input:    "<?php $x = 10; $y = 20; ?>",
			expected: "{{/* TODO_MIGRATE: $x = 10; $y = 20; */}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranspilePHPViewToTemplate(tt.input)
			if got != tt.expected {
				t.Errorf("TranspilePHPViewToTemplate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestView2GoTpl_DirectoryWalkAndBatch(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "ci3_views")
	outDir := filepath.Join(tmpDir, "go_views")

	// Create folder hierarchy
	if err := os.MkdirAll(filepath.Join(srcDir, "users"), 0755); err != nil {
		t.Fatalf("failed creating test dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "partials"), 0755); err != nil {
		t.Fatalf("failed creating test dir: %v", err)
	}

	// Security stub files that should be ignored
	stubContent := `<!DOCTYPE html><html><head><title>403 Forbidden</title></head><body><p>Directory access is forbidden.</p></body></html>`
	if err := os.WriteFile(filepath.Join(srcDir, "index.html"), []byte(stubContent), 0644); err != nil {
		t.Fatalf("failed writing stub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "users", "index.html"), []byte(stubContent), 0644); err != nil {
		t.Fatalf("failed writing stub: %v", err)
	}

	// PHP view files
	views := map[string]string{
		"welcome.php": `<h1><?= $title ?></h1><p><?= base_url('css/app.css') ?></p>`,
		filepath.Join("users", "list.php"): `<ul>
<?php foreach ($users as $user): ?>
  <li><?= $user->name ?> (<?= $user->email ?>)</li>
<?php endforeach; ?>
</ul>`,
		filepath.Join("users", "detail.php"): `<div class="user-detail">
  <h2><?= $user->name ?></h2>
  <a href="<?= site_url('users/detail/' . $id) ?>">Link</a>
</div>`,
		filepath.Join("partials", "header.php"): `<header>
  <nav><a href="<?= site_url() ?>">Home</a></nav>
</header>`,
	}

	for relPath, content := range views {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed writing view %s: %v", relPath, err)
		}
	}

	count, err := TranspileViewsDir(srcDir, outDir)
	if err != nil {
		t.Fatalf("TranspileViewsDir failed: %v", err)
	}

	if count != 4 {
		t.Errorf("expected 4 files converted, got %d", count)
	}

	// Verify that index.html stub files were NOT copied or converted
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); !os.IsNotExist(err) {
		t.Errorf("security index.html should not exist in output root")
	}
	if _, err := os.Stat(filepath.Join(outDir, "users", "index.html")); !os.IsNotExist(err) {
		t.Errorf("security index.html should not exist in output users dir")
	}

	// Verify converted .html files exist
	expectedFiles := []string{
		filepath.Join(outDir, "welcome.html"),
		filepath.Join(outDir, "users", "list.html"),
		filepath.Join(outDir, "users", "detail.html"),
		filepath.Join(outDir, "partials", "header.html"),
	}

	for _, ef := range expectedFiles {
		contentBytes, err := os.ReadFile(ef)
		if err != nil {
			t.Errorf("expected file %s to exist: %v", ef, err)
			continue
		}
		content := string(contentBytes)
		if strings.Contains(content, "<?php") || strings.Contains(content, "<?=") {
			t.Errorf("file %s still contains PHP tags:\n%s", ef, content)
		}
	}

	// Verify specific conversion in users/list.html
	listContent, _ := os.ReadFile(filepath.Join(outDir, "users", "list.html"))
	if !strings.Contains(string(listContent), "{{ range .Users }}") {
		t.Errorf("users/list.html missing {{ range .Users }}: %s", string(listContent))
	}
	if !strings.Contains(string(listContent), "{{ .Name }}") {
		t.Errorf("users/list.html missing {{ .Name }}: %s", string(listContent))
	}
}

func TestView2GoTpl_CLI(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "ci3_views")
	outDir := filepath.Join(tmpDir, "go_views")

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed creating src dir: %v", err)
	}

	viewContent := `<h1><?= $heading ?></h1>`
	if err := os.WriteFile(filepath.Join(srcDir, "home.php"), []byte(viewContent), 0644); err != nil {
		t.Fatalf("failed writing home.php: %v", err)
	}

	// Test CLI with -src and -out
	err := RunView2GoTplCLI([]string{"-src", srcDir, "-out", outDir})
	if err != nil {
		t.Fatalf("RunView2GoTplCLI failed: %v", err)
	}

	outHome := filepath.Join(outDir, "home.html")
	if _, err := os.Stat(outHome); os.IsNotExist(err) {
		t.Fatalf("expected output file %s does not exist", outHome)
	}

	content, err := os.ReadFile(outHome)
	if err != nil {
		t.Fatalf("failed reading converted file: %v", err)
	}
	if string(content) != "<h1>{{ .Heading }}</h1>" {
		t.Errorf("unexpected content in home.html: %q", string(content))
	}

	// Test CLI with missing -src
	errMissing := RunView2GoTplCLI([]string{"-out", outDir})
	if errMissing == nil {
		t.Errorf("expected error when -src is missing, got nil")
	}
}
