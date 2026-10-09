---
title: Template Engine
description: Rendering template HTML di Goigniter.
sidebar:
  order: 6
---

Goigniter menggunakan Go `html/template` untuk rendering HTML. Template adalah file HTML dengan tag action.

## Rendering Dasar

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

Ini merender `application/views/welcome.html`.

## Lokasi Template

Template dimuat dari direktori `application/views/`. Subdirektori didukung:

```
application/views/
├── welcome.html
├── products/
│   ├── index.html
│   └── detail.html
└── admin/
    └── dashboard.html
```

## Syntax Template

Go template menggunakan `{{ }}` untuk action:

```html
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

### Variable

```html
<h1>{{ .Title }}</h1>
<p>{{ .User.Name }}</p>
```

### Loops

```html
<ul>
{{ range .Products }}
    <li>{{ .Name }} - {{ .Price }}</li>
{{ end }}
</ul>
```

### Conditional

```html
{{ if .IsLoggedIn }}
    <p>Selamat datang, {{ .User.Name }}!</p>
{{ else }}
    <a href="/login">Login</a>
{{ end }}
```

## Layouts (Template Inheritance)

Layout utama:

```html
<!-- application/views/layouts/main.html -->
<!DOCTYPE html>
<html>
<head>
    <title>{{ .Title }} - My App</title>
</head>
<body>
    <header>
        <h1>My App</h1>
    </header>
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
```

Konten yang menggunakan layout:

```html
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

## Helper Bawaan

### base_url

```html
<link rel="stylesheet" href='{{ helper "base_url" "/static/css/style.css" }}'>
```

### site_url

```html
<a href='{{ helper "site_url" "/products/detail/5" }}'>Detail</a>
```

### is_active

```html
<a href="/" class='{{ helper "is_active" "/" "active" }}'>Home</a>
```

### excerpt

Memotong teks:

```html
<p>{{ helper "excerpt" .Description 100 }}</p>
```

### date

Format waktu:

```html
<p>{{ helper "date" .CreatedAt "2006-01-02" }}</p>
```

### nl2br

Newline ke `<br>`:

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

---

Lanjut ke [Database](/id/guide/07-database) untuk koneksi SQLite dan MySQL.