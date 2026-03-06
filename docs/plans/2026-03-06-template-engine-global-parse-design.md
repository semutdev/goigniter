# Template Engine: Global Parse Design

## Problem

The `TemplateEngine` in `system/core/application.go` parses each template file into a separate `*template.Template` instance. This causes two bugs:

1. **Templates using `{{define "name"}}` blocks render blank pages.** `Render()` calls `t.Execute()` which executes the root template (named after the filename), not the named define block. Since all content is inside `{{define}}`, the root template is empty.

2. **Cross-file `{{template "X" .}}` calls fail silently.** When `dashboard.html` calls `{{template "admin/layout" .}}`, the layout template lives in a separate file parsed independently. The dashboard's template tree doesn't know about the layout.

### Affected Pages

- `/auth/forgot` - returns 200 with 1-byte empty body (define block not executed)
- `/admin/dashboard` - layout template not found (cross-file reference)

## Solution: Parse All Templates Into One Global Template

### Changes to `TemplateEngine` struct

Add a `globalTemplate *template.Template` field. Remove `templates map[string]*template.Template`.

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

### Changes to `loadTemplates()`

Instead of parsing each file into a separate template, parse all files into one root template:

```go
func (e *TemplateEngine) loadTemplates() error {
    e.mu.Lock()
    defer e.mu.Unlock()

    root := template.New("").Funcs(e.funcMap)

    err := filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
        if err != nil { return err }
        if info.IsDir() || !strings.HasSuffix(path, e.ext) { return nil }

        content, err := os.ReadFile(path)
        if err != nil { return err }

        _, err = root.Parse(string(content))
        return err
    })

    if err != nil { return err }
    e.globalTemplate = root
    return nil
}
```

### Remove `parseTemplate()`

No longer needed since all parsing happens in `loadTemplates()`.

### Changes to `Render()`

Use `ExecuteTemplate` with the define name instead of `Execute`:

```go
func (e *TemplateEngine) Render(w io.Writer, name string, data any) error {
    if e.reload {
        if err := e.loadTemplates(); err != nil { return err }
    }

    e.mu.RLock()
    defer e.mu.RUnlock()

    return e.globalTemplate.ExecuteTemplate(w, name, data)
}
```

### Naming Convention

Template names match the `{{define "name"}}` block names used in templates. Controllers already call `View("auth/forgot", ...)` which matches `{{define "auth/forgot"}}`. No changes needed in controllers or templates.

### Error Handling

If a define name is not found, Go's `ExecuteTemplate` returns a clear error: `template: no template "X" associated with template ""`. This propagates through `View()` and `ServeHTTP()` returns a 500 error.

### Backward Compatibility

- Controllers calling `View("auth/forgot", data)` continue to work unchanged
- Templates using `{{define "name"}}` continue to work unchanged
- Templates using `{{template "other" .}}` now work correctly (cross-file)
- The only requirement: define names must be globally unique

## Files to Modify

1. `system/core/application.go` - TemplateEngine struct, loadTemplates, parseTemplate, Render
