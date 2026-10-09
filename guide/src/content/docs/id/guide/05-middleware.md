---
title: Middleware
description: Cara menggunakan dan membuat middleware di Goigniter.
sidebar:
  order: 5
---

Middleware membungkus handler untuk menjalankan kode sebelum atau sesudah request diproses. Berguna untuk autentikasi, logging, rate limiting, dan recovery.

## Pipeline Middleware

```
Request
   ↓
┌─────────────────┐
│  Logger         │ ← catat waktu mulai
│  ┌───────────┐  │
│  │ Recovery  │  │ ← tangkap panic
│  │ ┌───────┐ │  │
│  │ │Handler│ │  │ ← proses request
│  │ └───────┘ │  │
│  └───────────┘  │
└─────────────────┘
   ↓
Response
```

## Middleware Bawaan

```go
import "goigniter/system/middleware"

func main() {
    app := core.New()

    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())
    app.Use(middleware.CORS())
    app.Use(middleware.RateLimit())

    app.Run(":8080")
}
```

### Logger

Mencatat setiap request ke console:

```go
app.Use(middleware.Logger())
// [GET] /products 1.234ms <nil>
// [POST] /products 5.678ms <nil>
```

### Recovery

Menangkap panic dan mengembalikan error 500:

```go
app.Use(middleware.Recovery())
```

### CORS

Menangani Cross-Origin Resource Sharing:

```go
app.Use(middleware.CORS())
```

### RateLimit

Membatasi request per IP:

```go
app.Use(middleware.RateLimit())
```

## Global vs Group Middleware

### Global

Berlaku untuk **semua** route:

```go
app.Use(middleware.Logger())
app.Use(middleware.Recovery())
```

### Group

Berlaku hanya untuk route dalam group tertentu:

```go
// Public routes
app.GET("/", homeHandler)
app.GET("/products", listProducts)

// Admin routes dengan auth
admin := app.Group("/admin", AuthMiddleware())
admin.GET("/dashboard", dashboard)
admin.GET("/users", adminUsers)
```

## Middleware Kustom

```go
package middleware

import "goigniter/system/core"

func AuthMiddleware() core.Middleware {
    return func(next core.HandlerFunc) core.HandlerFunc {
        return func(c *core.Context) error {
            userID := getSessionUserID(c)

            if userID == 0 {
                return c.Redirect(302, "/login")
            }

            c.Set("user_id", userID)
            return next(c)
        }
    }
}
```

## Controller-Level Middleware

Middleware untuk semua method di controller:

```go
type Admin struct {
    core.Controller
}

func (a *Admin) Middleware() []core.Middleware {
    return []core.Middleware{
        AuthMiddleware(),
        AdminOnlyMiddleware(),
    }
}

func (a *Admin) Index() {
    a.Ctx.View("admin/dashboard", core.Map{})
}
```

## Per-Method Middleware

Middleware untuk method tertentu:

```go
type Product struct {
    core.Controller
}

func (p *Product) MiddlewareFor() map[string][]core.Middleware {
    return map[string][]core.Middleware{
        "Store":  {AuthMiddleware()},
        "Update": {AuthMiddleware()},
        "Delete": {AuthMiddleware(), AdminOnly()},
    }
}
```

---

Lanjut ke [Template Engine](/id/guide/06-templates) untuk rendering HTML views.
