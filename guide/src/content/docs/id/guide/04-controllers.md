---
title: Controller & AutoRoute
description: Membuat controller dengan registrasi route otomatis.
sidebar:
  order: 4
---

Controller menangani HTTP request. Dengan AutoRoute, cukup daftarkan controller — route dibuat otomatis berdasarkan nama method.

## Controller Dasar

```go
package main

import "github.com/semutdev/goigniter/system/core"

type Welcome struct {
    core.Controller
}

func (w *Welcome) Index() {
    w.Ctx.View("welcome", core.Map{
        "Title": "Welcome",
    })
}

func (w *Welcome) About() {
    w.Ctx.View("about", core.Map{})
}
```

Poin penting:
- Controller meng-embed `core.Controller`
- `Index()` adalah method default, terpetakan ke root path controller
- Akses request dan response via `w.Ctx`

## Register & AutoRoute

```go
func main() {
    app := core.New()

    core.Register(&Welcome{})
    core.Register(&Product{})
    core.Register(&Dashboard{}, "admin") // dengan prefix

    app.AutoRoute()
    app.Run(":8080")
}
```

Route digenerate otomatis dari nama controller dan method.

## Method Routing

| Method | HTTP Methods | URL Pattern |
|--------|--------------|-------------|
| `Index()` | GET, POST | `/{controller}` |
| `Store()` | POST | `/{controller}/store` |
| `Create()` | GET, POST | `/{controller}/create` |
| Method lainnya | GET, POST | `/{controller}/{method}` |

## Mengontrol HTTP Methods

Batasi method HTTP yang diterima:

```go
type Auth struct {
    core.Controller
}

func (a *Auth) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Login":   {"GET"},
        "Dologin": {"POST"},
        "Logout":  {"GET", "POST"},
    }
}
```

## Custom Route Patterns

Untuk route dengan parameter seperti `:id`:

```go
type Product struct {
    core.Controller
}

func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Edit":   "edit/:id",
        "Update": "update/:id",
        "Delete": "delete/:id",
        "Detail": "detail/:id",
    }
}
```

## Controller Bersarang (Prefix)

Tambah prefix saat register:

```go
core.Register(&Dashboard{}, "admin")
core.Register(&User{}, "admin")
```

Hasil:

```
Product     → /product
Dashboard   → /admin/dashboard
User        → /admin/user
```

## Akses Data Request

```go
func (p *Product) Store() {
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")
    page := p.Ctx.Query("page")
    id := p.Ctx.Param("id")

    var data map[string]any
    p.Ctx.Bind(&data)
}
```

## Tipe Response

```go
p.Ctx.JSON(200, core.Map{"status": "success"})
p.Ctx.View("products/index", core.Map{"products": products})
p.Ctx.String(200, "Hello World")
p.Ctx.Redirect(302, "/login")
p.Ctx.File("/path/to/file.pdf")
p.Ctx.NoContent(204)
```

---

Lanjut ke [Middleware](/id/guide/05-middleware) untuk autentikasi, logging, dan keamanan.