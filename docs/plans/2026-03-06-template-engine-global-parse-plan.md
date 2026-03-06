# Template Engine Global Parse Fix - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix blank page rendering for templates that use `{{define}}` blocks and cross-file `{{template}}` calls by parsing all template files into a single global template tree.

**Architecture:** Replace per-file template parsing with a single `*template.Template` root that contains all define blocks. Templates without `{{define}}` wrappers get auto-wrapped using their path-based name. `Render()` switches from `Execute()` to `ExecuteTemplate()`.

**Tech Stack:** Go `html/template`, existing `TemplateEngine` in `system/core/application.go`

---

### Task 1: Write failing test for global template rendering

**Files:**
- Modify: `system/core/application_test.go`

**Step 1: Create test template directory with test fixtures**

Create a temporary directory structure in the test with templates that reproduce the bug:

```go
func TestTemplateEngine_GlobalParse(t *testing.T) {
	// Setup temp dir with test templates
	dir := t.TempDir()

	// Template with {{define}} block (like auth/forgot.html)
	os.MkdirAll(filepath.Join(dir, "auth"), 0755)
	os.WriteFile(filepath.Join(dir, "auth", "forgot.html"), []byte(`{{define "auth/forgot"}}<!DOCTYPE html><html><body>Forgot: {{.Title}}</body></html>{{end}}`), 0644)

	// Template without {{define}} (like welcome.html)
	os.WriteFile(filepath.Join(dir, "welcome.html"), []byte(`<!DOCTYPE html><html><body>Welcome: {{.Title}}</body></html>`), 0644)

	// Layout template with cross-file references
	os.MkdirAll(filepath.Join(dir, "admin"), 0755)
	os.WriteFile(filepath.Join(dir, "admin", "layout.html"), []byte(`{{define "admin/layout"}}<html><body>{{template "admin-content" .}}</body></html>{{end}}{{define "admin-content"}}default{{end}}`), 0644)

	// Template that references layout (like admin/dashboard.html)
	os.WriteFile(filepath.Join(dir, "admin", "dashboard.html"), []byte(`{{define "admin/dashboard"}}{{template "admin/layout" .}}{{end}}{{define "admin-content"}}<h1>Dashboard: {{.Title}}</h1>{{end}}`), 0644)

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
		contains string
	}{
		{"auth/forgot", "Forgot:"},
		{"welcome", "Welcome:"},
		{"admin/dashboard", "Dashboard:"},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		err := engine.Render(&buf, tt.name, Map{"Title": "Test"})
		if err != nil {
			t.Errorf("Render(%q) error: %v", tt.name, err)
			continue
		}
		if !strings.Contains(buf.String(), tt.contains) {
			t.Errorf("Render(%q) = %q, want to contain %q", tt.name, buf.String(), tt.contains)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/jamal/dev/golang/goigniter && go test ./system/core/ -run TestTemplateEngine_GlobalParse -v`
Expected: FAIL — `auth/forgot` renders empty, `admin/dashboard` fails with missing template

**Step 3: Commit test**

```bash
git add system/core/application_test.go
git commit -m "test: add failing test for global template parse"
```

---

### Task 2: Modify TemplateEngine struct

**Files:**
- Modify: `system/core/application.go:218-225`

**Step 1: Replace `templates map` with `globalTemplate`**

Change the struct from:
```go
type TemplateEngine struct {
	dir       string
	ext       string
	funcMap   template.FuncMap
	templates map[string]*template.Template
	reload    bool
	mu        sync.RWMutex
}
```

To:
```go
type TemplateEngine struct {
	dir            string
	ext            string
	funcMap        template.FuncMap
	globalTemplate *template.Template
	reload         bool
	mu             sync.RWMutex
}
```

**Step 2: Update `NewTemplateEngine` to not init the map**

In `NewTemplateEngine` (line 244-250), change from:
```go
e := &TemplateEngine{
    dir:       config.Dir,
    ext:       config.Ext,
    funcMap:   config.FuncMap,
    templates: make(map[string]*template.Template),
    reload:    config.Reload,
}
```

To:
```go
e := &TemplateEngine{
    dir:     config.Dir,
    ext:     config.Ext,
    funcMap: config.FuncMap,
    reload:  config.Reload,
}
```

---

### Task 3: Rewrite `loadTemplates()` for global parse

**Files:**
- Modify: `system/core/application.go:259-290`

**Step 1: Rewrite loadTemplates**

Replace the entire `loadTemplates` method with:

```go
func (e *TemplateEngine) loadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	root := template.New("").Funcs(e.funcMap)

	err := filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, e.ext) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Auto-wrap bare templates (no {{define}}) with a path-based name
		text := string(content)
		if !strings.Contains(text, "{{define") {
			rel, err := filepath.Rel(e.dir, path)
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(rel, e.ext)
			name = strings.ReplaceAll(name, string(os.PathSeparator), "/")
			text = `{{define "` + name + `"}}` + text + `{{end}}`
		}

		_, err = root.Parse(text)
		return err
	})

	if err != nil {
		return err
	}
	e.globalTemplate = root
	return nil
}
```

Key detail: templates without `{{define}}` get auto-wrapped with their path-based name (e.g., `welcome.html` becomes `{{define "welcome"}}...{{end}}`). This maintains backward compatibility with bare templates like `auth/login.html` and `welcome.html`.

---

### Task 4: Remove `parseTemplate()` and update `Render()`

**Files:**
- Modify: `system/core/application.go:292-319`

**Step 1: Delete `parseTemplate` method**

Remove lines 292-300:
```go
func (e *TemplateEngine) parseTemplate(path string) (*template.Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := template.New(filepath.Base(path)).Funcs(e.funcMap)
	return t.Parse(string(content))
}
```

**Step 2: Rewrite `Render` method**

Replace the current `Render` (lines 303-319) with:

```go
func (e *TemplateEngine) Render(w io.Writer, name string, data any) error {
	if e.reload {
		if err := e.loadTemplates(); err != nil {
			return err
		}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.globalTemplate == nil {
		return &TemplateNotFoundError{Name: name}
	}

	return e.globalTemplate.ExecuteTemplate(w, name, data)
}
```

---

### Task 5: Run tests and verify fix

**Step 1: Run the new test**

Run: `cd /Users/jamal/dev/golang/goigniter && go test ./system/core/ -run TestTemplateEngine_GlobalParse -v`
Expected: PASS

**Step 2: Run all existing tests**

Run: `cd /Users/jamal/dev/golang/goigniter && go test ./system/core/ -v`
Expected: All PASS

**Step 3: Build the full-crud example to check compilation**

Run: `cd /Users/jamal/dev/golang/goigniter/examples/full-crud && go build -o /dev/null .`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add system/core/application.go system/core/application_test.go
git commit -m "fix: use global template parse to fix blank page rendering

Parse all template files into a single template tree so that:
- Templates with {{define}} blocks are executed correctly by name
- Cross-file {{template}} calls resolve properly
- Bare templates (no {{define}}) are auto-wrapped with path-based names"
```

---

### Task 6: Fix bare templates in full-crud example

The `auth/login.html` and `welcome.html` templates don't use `{{define}}`. The auto-wrap in Task 3 handles this, but we should verify they render correctly with the new engine.

**Step 1: Start the full-crud server and test all affected pages**

Run: `cd /Users/jamal/dev/golang/goigniter/examples/full-crud && go run main.go &` then:
- `curl -s http://localhost:8080/auth/forgot` — should contain "Lupa Password"
- `curl -s http://localhost:8080/auth/login` — should contain "Login"
- `curl -s http://localhost:8080/welcome` — should contain "GoIgniter"

**Step 2: If any page still blank, debug and fix**

Check server logs for template errors. The auto-wrap should handle bare templates.

**Step 3: Commit any template fixes if needed**

```bash
git add examples/full-crud/
git commit -m "fix: update example templates for global template engine"
```
