# GoIgniter - System Restructure Design

**Tanggal:** 2026-01-13
**Status:** Approved

---

## Overview

Restructure GoIgniter dari `pkg/goigniter/` ke `system/` dengan struktur CI3-style. Framework berbasis net/http stdlib dengan zero external HTTP dependencies.

---

## Keputusan Design

| Aspek | Keputusan |
|-------|-----------|
| Lokasi framework | `system/` (seperti CI3) |
| Versioning | Tanpa versi dulu, fokus development |
| Existing libs/ | Hybrid: helpers → system/, auth tetap di libs/ |
| Session | Custom cookie-based (no gorilla dependency) |
| Struktur | CI3-style: core/, libraries/, helpers/, middleware/ |
| Import | Multiple imports per sub-package |

---

## Struktur Direktori

```
goigniter/
├── system/                          ← Framework core
│   ├── core/                        ← Main framework components
│   │   ├── goigniter.go             ← New(), Register(), Map{}
│   │   ├── application.go           ← Application, Use(), Group()
│   │   ├── router.go                ← Router wrapper
│   │   ├── context.go               ← Context + pooling
│   │   ├── controller.go            ← Base Controller, Loader
│   │   ├── registry.go              ← Controller registry
│   │   ├── render.go                ← Template rendering
│   │   └── internal/
│   │       └── radix/               ← Radix tree (internal)
│   │
│   ├── libraries/                   ← Reusable libraries
│   │   └── session/                 ← Custom session (cookie-based)
│   │       └── session.go
│   │
│   ├── helpers/                     ← Helper functions
│   │   ├── url.go                   ← base_url(), site_url()
│   │   └── template.go              ← Template functions
│   │
│   ├── middleware/                  ← Built-in middlewares
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   ├── ratelimit.go
│   │   └── auth.go
│   │
│   └── database/                    ← (Future: query builder)
│
├── libs/                            ← App-specific (user code)
│   └── auth.go                      ← Auth dengan GORM
│
├── controllers/
├── models/
├── views/
└── main.go
```

---

## Import & Usage

```go
package main

import (
    "goigniter/system/core"
    "goigniter/system/middleware"
    "goigniter/system/helpers"
    "goigniter/system/libraries/session"
)

func main() {
    app := core.New()

    app.LoadTemplates("./views", true)

    session.Init(session.Config{
        Secret:   "your-secret-key",
        MaxAge:   86400,
        HttpOnly: true,
    })

    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())

    app.GET("/", func(c *core.Context) error {
        return c.JSON(200, core.Map{
            "message": "Welcome",
            "url":     helpers.BaseURL("/api"),
        })
    })

    app.AutoRoute()
    app.Run(":8080")
}
```

---

## Custom Session

Cookie-based dengan HMAC signing:

```go
type Config struct {
    Secret     string
    CookieName string        // Default: "goigniter_session"
    MaxAge     int           // Seconds
    Path       string        // Default: "/"
    HttpOnly   bool          // Default: true
    Secure     bool          // Default: false
    SameSite   http.SameSite // Default: Lax
}

type Session struct {
    ID    string
    Data  map[string]any
    Flash map[string]string
}

// Cookie format: base64(json) + "." + hmac_signature
```

---

## Migration Plan

```
pkg/goigniter/goigniter.go      →   system/core/goigniter.go
pkg/goigniter/application.go    →   system/core/application.go
pkg/goigniter/context.go        →   system/core/context.go
pkg/goigniter/router.go         →   system/core/router.go
pkg/goigniter/controller.go     →   system/core/controller.go
pkg/goigniter/registry.go       →   system/core/registry.go
pkg/goigniter/internal/radix/   →   system/core/internal/radix/
pkg/goigniter/middleware/       →   system/middleware/
pkg/goigniter/render/           →   system/core/render.go

(NEW)                           →   system/libraries/session/
(NEW)                           →   system/helpers/
```

---

## Next Steps

1. Buat struktur direktori `system/`
2. Migrate kode dari `pkg/goigniter/`
3. Update package names dan imports
4. Implement custom session
5. Implement helpers (base_url, site_url)
6. Update tests
7. Update example
8. Hapus `pkg/goigniter/`
