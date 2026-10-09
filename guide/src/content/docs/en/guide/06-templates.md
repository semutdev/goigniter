---
title: Template Engine
description: Rendering HTML templates in Goigniter.
sidebar:
  order: 6
---

Goigniter uses Go's `html/template` for rendering HTML. Templates are HTML files with embedded action tags.

## Basic Rendering

```go
type Welcome struct {
    core.Controller
}

func (w *Welcome) Index() {
    w.Ctx.View("welcome", core.Map{
        "Title": "Welcome",
    })
}
```

This renders `application/views/welcome.html`.

## Template Locations

Templates are loaded from the `application/views/` directory. Subdirectories are supported:

```
application/views/
├── welcome.html
├── products/
│   ├── index.html
│   └── detail.html
└── admin/
    └── dashboard.html
```

Render by path:

```go
w.Ctx.View("products/index", core.Map{})
w.Ctx.View("admin/dashboard", core.Map{})
```

## Template Syntax

Go templates use `{{ }}` for actions:

```html
<!-- application/views/welcome.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }}</title>
</head>
<body>
    <h1>{{ .Title }}</h1>
    <p>Welcome to Goigniter!</p>
</body>
</html>
```

### Variables

```html
<h1>{{ .Title }}</h1>
<p>{{ .User.Name }}</p>
<p>{{ .Count }}</p>
```

### Loops

```html
<ul>
{{ range .Products }}
    <li>{{ .Name }} - {{ .Price }}</li>
{{ end }}
</ul>
```

### Conditionals

```html
{{ if .IsLoggedIn }}
    <p>Welcome back, {{ .User.Name }}!</p>
{{ else }}
    <a href="/login">Login</a>
{{ end }}
```

## Layouts (Template Inheritance)

Define a layout with `{{template}}`:

```html
<!-- application/views/layouts/main.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }} - My App</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header>
        <h1>My App</h1>
        <nav>
            <a href="/">Home</a>
            <a href="/products">Products</a>
        </nav>
    </header>

    <main>
        {{template "content" .}}
    </main>

    <footer>
        <p>&copy; 2025 My App</p>
    </footer>
</body>
</html>
```

Content template that uses the layout:

```html
<!-- application/views/products/index.html -->
{{template "layouts/main" .}}

{{define "content"}}
    <h2>Products</h2>
    <ul>
    {{range .Products}}
        <li>{{.Name}} - {{.Price}}</li>
    {{end}}
    </ul>
{{end}}
```

## Built-in Helpers

### base_url

Generates a public URL:

```html
<link rel="stylesheet" href='{{ helper "base_url" "/static/css/style.css" }}'>
<!-- Output: http://localhost:8080/static/css/style.css -->
```

### site_url

Generates a route URL:

```html
<a href='{{ helper "site_url" "/products/detail/5" }}'>Detail</a>
<!-- Output: http://localhost:8080/products/detail/5 -->
```

### asset_url

Generates a URL for static assets:

```html
<img src='{{ helper "asset_url" "/images/logo.png" }}'>
```

### is_active

Returns "active" if the current route matches:

```html
<a href="/" class='{{ helper "is_active" "/" "active" }}'>Home</a>
<a href="/products" class='{{ helper "is_active" "/products" "active" }}'>Products</a>
```

### excerpt

Truncates text to a specified length:

```html
<p>{{ helper "excerpt" .Description 100 }}</p>
```

### date

Formats a time value:

```html
<p>{{ helper "date" .CreatedAt "2006-01-02" }}</p>
```

### nl2br

Converts newlines to `<br>` tags:

```html
<p>{{ helper "nl2br" .Description }}</p>
```

## JSON Response

```go
func (p *Product) Index() {
    products := getProducts()
    p.Ctx.JSON(200, core.Map{
        "success": true,
        "data":    products,
    })
}
```

## String Response

```go
func (w *Welcome) Health() {
    w.Ctx.String(200, "OK")
}
```

---

Next: [Database](/guide/07-database) for connecting to SQLite and MySQL.