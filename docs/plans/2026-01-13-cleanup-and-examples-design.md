# Cleanup Echo & Move to Examples Design

**Date:** 2026-01-13
**Status:** Approved

## Overview

Menghapus framework Echo dari root project dan memindahkan kode aplikasi (controllers, models, views) ke `examples/full-crud/` dengan struktur CI3-style. Controllers dimigrasikan ke `system/core`, models tetap pakai GORM.

## Final Structure

```
goigniter/
├── system/                    # Framework core
│   ├── core/
│   ├── middleware/
│   ├── libraries/session/
│   └── helpers/
├── examples/
│   ├── simple/
│   ├── autoroute/
│   ├── views/
│   └── full-crud/            # Full CRUD dengan database
│       ├── application/
│       │   ├── controllers/
│       │   │   ├── welcome.go
│       │   │   ├── auth.go
│       │   │   └── admin/
│       │   │       ├── dashboard.go
│       │   │       └── product.go
│       │   ├── models/
│       │   │   ├── user.go
│       │   │   ├── group.go
│       │   │   ├── product.go
│       │   │   └── login_attempt.go
│       │   └── views/
│       │       ├── auth/
│       │       └── admin/
│       ├── config/
│       ├── database/
│       ├── public/
│       ├── main.go
│       ├── go.mod
│       └── .env.example
├── README.md
├── go.mod                     # Framework only (no Echo)
└── CLAUDE.md
```

## Controller Migration

### Before (Echo)

```go
package controllers

import (
    "github.com/labstack/echo/v4"
    "goigniter/libs"
)

type WelcomeController struct{}

func init() {
    libs.Register(&WelcomeController{})
}

func (w *WelcomeController) Index(c echo.Context) error {
    return c.Render(200, "welcome", nil)
}
```

### After (system/core)

```go
package controllers

import (
    "goigniter/system/core"
)

type WelcomeController struct {
    core.Controller
}

func init() {
    core.Register(&WelcomeController{})
}

func (w *WelcomeController) Index() {
    w.Ctx.View("welcome", core.Map{
        "title": "Welcome",
    })
}
```

### Key Differences

- Embed `core.Controller`
- Methods have no parameters (access context via `w.Ctx`)
- Response via `w.Ctx.View()`, `w.Ctx.JSON()`, etc.
- Register to `core.Register()`

## Models

GORM models unchanged - independent from HTTP layer.

## README Structure

Minimal documentation:
- Warning badge "under development"
- Feature list
- Quick start (10 lines)
- Examples table with links
- ~50-60 lines total

## go.mod Changes

### Root go.mod (framework only)

```go
module goigniter

go 1.24.8

// No external HTTP dependencies
```

### examples/full-crud/go.mod

```go
module full-crud

go 1.24.8

require (
    goigniter v0.0.0
    github.com/go-playground/validator/v10 v10.30.1
    github.com/joho/godotenv v1.5.1
    golang.org/x/crypto v0.46.0
    gopkg.in/gomail.v2 v2.0.0
    gorm.io/driver/mysql v1.6.0
    gorm.io/gorm v1.31.1
)

replace goigniter => ../..
```

## Files to Delete from Root

- `main.go`
- `controllers/`
- `models/`
- `views/`
- `config/`
- `database/`
- `libs/`
- `public/`
- `.env`, `.env.example`, `.air.toml`

## Implementation Steps

1. Create `examples/full-crud/` structure
2. Copy and migrate controllers to system/core
3. Copy models (unchanged)
4. Copy views, config, database, public
5. Create example go.mod with replace directive
6. Create main.go using system/core
7. Delete files from root
8. Update root go.mod (remove Echo)
9. Update README.md
10. Run `go mod tidy`
11. Test build and verify
