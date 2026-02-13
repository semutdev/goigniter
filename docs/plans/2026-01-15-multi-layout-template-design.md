# Multi-Layout Template Design (CI3 Style)

## Overview

Menambahkan method `Render()` untuk render template ke string, memungkinkan developer compose multiple views seperti CodeIgniter 3.

## Design Decisions

1. **CI3 Style** - Multiple view calls, explicit composition
2. **Dua method**: `View()` untuk output langsung, `Render()` untuk return string
3. **Simple templates** - Tidak perlu `{{define}}` blocks, setiap file standalone
4. **No breaking changes** - `View()` tetap sama

## Method Signatures

```go
// View - render template langsung ke response (existing)
func (c *Context) View(name string, data Map) error

// Render - render template ke string (NEW)
func (c *Context) Render(name string, data Map) (string, error)
```

## Usage Examples

### Simple View
```go
func (c *HomeController) Index() {
    c.Ctx.View("home", core.Map{"Title": "Home"})
}
```

### Compose with Partials
```go
func (d *DashboardController) Index() {
    sidebar, _ := d.Ctx.Render("admin/partials/sidebar", core.Map{})
    content, _ := d.Ctx.Render("admin/pages/dashboard", core.Map{
        "Title": "Dashboard",
    })

    d.Ctx.View("admin/layouts/main", core.Map{
        "Title":   "Dashboard",
        "Sidebar": sidebar,
        "Content": content,
    })
}
```

### CI3 Style - Multiple Views
```go
func (c *PageController) About() {
    c.Ctx.View("header", core.Map{"Title": "About"})
    c.Ctx.View("pages/about", core.Map{})
    c.Ctx.View("footer", core.Map{})
}
```

## Template Examples

### Layout Template
```html
<!-- views/layouts/admin.html -->
<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
    <nav>{{.Sidebar | safe}}</nav>
    <main>{{.Content | safe}}</main>
</body>
</html>
```

### Partial Template
```html
<!-- views/partials/sidebar.html -->
<ul>
    <li><a href="/admin/dashboard">Dashboard</a></li>
    <li><a href="/admin/product">Products</a></li>
</ul>
```

## Recommended Folder Structure

```
views/
├── layouts/
│   ├── main.html
│   └── admin.html
├── partials/
│   ├── header.html
│   ├── footer.html
│   └── sidebar.html
└── pages/
    ├── home.html
    └── about.html
```

## Files to Modify

1. `system/core/context.go` - Add `Render()` method

## Notes

- Use `| safe` in templates to output HTML strings
- Each template file is standalone (no `{{define}}` blocks needed)
- Render() returns error if template not found
