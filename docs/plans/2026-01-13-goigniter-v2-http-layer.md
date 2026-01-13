# GoIgniter v2 - HTTP Layer Design

**Tanggal:** 2026-01-13
**Status:** Draft
**Author:** Brainstorming session

---

## Overview

GoIgniter v2 adalah rewrite dari GoIgniter dengan HTTP layer custom berbasis `net/http` stdlib Go, menghilangkan ketergantungan pada Echo framework. Tujuannya adalah **kontrol penuh** atas routing, middleware, dan context handling.

### Prinsip Utama

- Zero external dependency untuk HTTP (murni `net/http`)
- Controller registry dengan pattern "extends" ala CI3
- Middleware cascade: Global → Group → Controller → Method
- Context pooling untuk performa optimal

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Application                         │
├─────────────────────────────────────────────────────────┤
│  Controllers (user code)                                 │
│  ┌─────────┐ ┌─────────┐ ┌──────────────┐              │
│  │ Welcome │ │  Auth   │ │ admin/       │              │
│  │         │ │         │ │  Dashboard   │              │
│  └────┬────┘ └────┬────┘ └──────┬───────┘              │
│       │           │             │                       │
│       └───────────┴─────────────┘                       │
│                   │ extends                             │
├───────────────────▼─────────────────────────────────────┤
│              goigniter.Controller                        │
│  ┌─────────────────────────────────────────────────┐    │
│  │ • Response: JSON, HTML, View, Redirect          │    │
│  │ • Input: Param, Query, Form, Bind               │    │
│  │ • Session & Auth: User, IsLoggedIn              │    │
│  │ • Loader: Load.Model, Load.Library              │    │
│  └─────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────┤
│                    Core Engine                           │
│  ┌──────────┐ ┌────────────┐ ┌───────────┐             │
│  │  Router  │ │ Middleware │ │  Context  │             │
│  │  (Radix) │ │   Chain    │ │  (Pool)   │             │
│  └──────────┘ └────────────┘ └───────────┘             │
├─────────────────────────────────────────────────────────┤
│                   net/http stdlib                        │
└─────────────────────────────────────────────────────────┘
```

---

## 1. Router - Radix Tree Implementation

### Route Registration

```go
app := goigniter.New()

// Basic routes
app.GET("/", handler)
app.POST("/users", handler)

// Path parameters
app.GET("/users/:id", handler)           // :id = named param
app.GET("/files/*filepath", handler)     // * = wildcard/catch-all

// Route groups dengan middleware
api := app.Group("/api", middleware.CORS)
{
    v1 := api.Group("/v1", middleware.RateLimit)
    {
        v1.GET("/users", handler)        // /api/v1/users
        v1.POST("/users", handler)
    }
}

// Auto-route dari controller registry
app.AutoRoute()
```

### Auto-Route Naming Convention

| Method Name | HTTP Method + Path |
|-------------|-------------------|
| Index | GET /controller |
| Show | GET /controller/:id |
| Create | GET /controller/create |
| Store | POST /controller |
| Edit | GET /controller/:id/edit |
| Update | PUT /controller/:id |
| Delete | DELETE /controller/:id |
| Custom | GET /controller/custom |

### Router Features

- **Radix tree** untuk lookup O(log n)
- **Path params** dengan `:name` dan wildcard `*`
- **Route groups** dengan prefix dan shared middleware
- **Auto-route** generate routes dari controller methods

---

## 2. Context - Request/Response Handler

```go
type Context struct {
    Request    *http.Request
    Response   http.ResponseWriter
    params     map[string]string      // path params
    query      url.Values             // cached query
    store      map[string]any         // request-scoped data
    controller ControllerInterface    // parent controller
}
```

### Response Helpers

```go
func (c *Context) JSON(code int, data any) error
func (c *Context) HTML(code int, html string) error
func (c *Context) View(name string, data Map) error
func (c *Context) Redirect(code int, url string) error
func (c *Context) String(code int, s string) error
func (c *Context) File(filepath string) error
func (c *Context) NoContent(code int) error
```

### Input Helpers

```go
func (c *Context) Param(name string) string
func (c *Context) ParamInt(name string) (int, error)
func (c *Context) Query(name string) string
func (c *Context) QueryDefault(name, def string) string
func (c *Context) Form(name string) string
func (c *Context) Bind(dest any) error
func (c *Context) Body() []byte
```

### Request Info

```go
func (c *Context) Method() string
func (c *Context) Path() string
func (c *Context) IP() string
func (c *Context) Header(name string) string
func (c *Context) Cookie(name string) (*http.Cookie, error)
func (c *Context) SetCookie(cookie *http.Cookie)
```

### Request-Scoped Storage

```go
func (c *Context) Set(key string, value any)
func (c *Context) Get(key string) any
```

---

## 3. Middleware Chain

### Signature

```go
type Middleware func(next HandlerFunc) HandlerFunc
type HandlerFunc func(c *Context) error
```

### Built-in Middlewares

```go
app.Use(middleware.Recovery())
app.Use(middleware.Logger())
app.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))
app.Use(middleware.RateLimit(100, time.Minute))
```

### Cascade Priority

Global → Group → Controller → Method

```go
// 1. Global
app.Use(middleware.Logger, middleware.Recovery)

// 2. Group
admin := app.Group("/admin", middleware.Auth)

// 3. Controller-level
func (c *Dashboard) Middleware() []Middleware {
    return []Middleware{middleware.AdminOnly}
}

// 4. Method-level
func (c *Dashboard) MiddlewareFor() map[string][]Middleware {
    return map[string][]Middleware{
        "Delete": {middleware.SuperAdminOnly},
    }
}
```

### Execution Flow

```
Request
   │
   ▼
┌─────────────────┐
│ Global: Logger  │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Global: Recovery│
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Group: Auth     │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Controller:     │
│ AdminOnly       │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│ Method:         │
│ SuperAdminOnly  │
└─────────────────┘
   │
   ▼
┌─────────────────┐
│    Handler      │
└─────────────────┘
   │
   ▼
Response (bubble back up)
```

---

## 4. Controller System & Registry

### Base Controller

```go
type Controller struct {
    Ctx    *Context
    Load   *Loader
    Data   Map
}

type Loader struct {
    controller *Controller
    models     map[string]any
    libraries  map[string]any
}

func (l *Loader) Model(name string) any
func (l *Loader) Library(name string) any
func (l *Loader) View(name string, data Map) error
func (l *Loader) Helper(name string)
```

### Controller Example

```go
// controllers/welcome.go
package controllers

import "github.com/semutdev/goigniter"

type Welcome struct {
    goigniter.Controller
}

func init() {
    goigniter.Register(&Welcome{})
}

func (c *Welcome) Index() {
    c.Ctx.View("welcome/index", goigniter.Map{
        "title": "Welcome to GoIgniter",
    })
}

func (c *Welcome) Show() {
    id := c.Ctx.Param("id")
    c.Ctx.JSON(200, goigniter.Map{"id": id})
}
```

### Nested Controller

```go
// controllers/admin/dashboard.go
package admin

import "github.com/semutdev/goigniter"

type Dashboard struct {
    goigniter.Controller
}

func init() {
    goigniter.Register(&Dashboard{}, "admin")
}

func (c *Dashboard) Index() {
    c.Ctx.View("admin/dashboard", goigniter.Map{
        "user": c.Ctx.Get("user"),
    })
}

func (c *Dashboard) Middleware() []goigniter.Middleware {
    return []goigniter.Middleware{
        middleware.Auth,
        middleware.AdminOnly,
    }
}
```

### Registry Internals

```go
var registry = &controllerRegistry{
    controllers: make(map[string]ControllerFactory),
}

func Register(controller ControllerInterface, prefix ...string) {
    name := reflect.TypeOf(controller).Elem().Name()
    path := strings.ToLower(name)

    if len(prefix) > 0 {
        path = prefix[0] + "/" + path
    }

    registry.controllers[path] = func() ControllerInterface {
        return reflect.New(reflect.TypeOf(controller).Elem()).Interface().(ControllerInterface)
    }
}

func (app *Application) AutoRoute() {
    for path, factory := range registry.controllers {
        app.registerControllerRoutes(path, factory)
    }
}
```

---

## 5. Package Structure

### Framework Structure

```
goigniter/
├── goigniter.go           # Main entry: New(), Register(), Map{}
├── application.go         # Application struct, Use(), Group(), AutoRoute()
├── router.go              # Radix tree router implementation
├── context.go             # Context struct & methods
├── controller.go          # Base Controller, Loader
├── registry.go            # Controller registry
├── middleware.go          # Middleware types & chain executor
│
├── middleware/            # Built-in middlewares
│   ├── logger.go
│   ├── recovery.go
│   ├── cors.go
│   ├── ratelimit.go
│   └── auth.go
│
├── render/                # Template rendering
│   ├── render.go
│   ├── html.go
│   └── json.go
│
└── internal/
    ├── radix/             # Radix tree implementation
    └── pool/              # sync.Pool utilities
```

### User Project Structure

```
myapp/
├── main.go
├── go.mod
├── .env
│
├── config/
│   └── app.go
│
├── controllers/
│   ├── welcome.go
│   ├── auth.go
│   └── admin/
│       └── dashboard.go
│
├── models/
│   └── user.go
│
├── middleware/
│   └── custom.go
│
├── views/
│   ├── layout.html
│   ├── welcome/
│   │   └── index.html
│   └── admin/
│       └── dashboard.html
│
└── public/
    ├── css/
    ├── js/
    └── images/
```

### Example main.go

```go
package main

import (
    "github.com/semutdev/goigniter"
    "github.com/semutdev/goigniter/middleware"

    _ "myapp/controllers"
    _ "myapp/controllers/admin"
)

func main() {
    app := goigniter.New()

    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())

    app.Static("/public", "./public")

    app.AutoRoute()

    app.Run(":8080")
}
```

---

## 6. Session & Auth Helpers

```go
func (c *Dashboard) Index() {
    // Session
    c.Session().Set("key", "value")
    val := c.Session().Get("key")

    // Auth helpers
    if c.IsLoggedIn() {
        user := c.User()
        c.Ctx.View("dashboard", Map{"user": user})
    } else {
        c.Ctx.Redirect(302, "/login")
    }
}
```

---

## 7. View Helpers

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{ .title }}</title>
    {{ base_url "css/style.css" }}
</head>
<body>
    {{ if .user }}
        Welcome, {{ .user.Name }}!
    {{ end }}

    {{ view "partials/footer" }}
</body>
</html>
```

---

## Summary

| Komponen | Deskripsi |
|----------|-----------|
| **Router** | Radix tree, path params, groups, auto-route |
| **Context** | Request/response helpers, pooled |
| **Middleware** | Cascade: Global → Group → Controller → Method |
| **Controller** | Registry pattern, extends base controller |
| **Loader** | `c.Load.Model()`, `c.Load.Library()` ala CI3 |
| **Session/Auth** | Built-in `c.Session()`, `c.User()`, `c.IsLoggedIn()` |
| **Views** | Go html/template dengan helper functions |

---

## Phase 2 (Future)

Database layer dengan:
- Query builder ala CI3 (`db.Where().Get()`)
- Migration system
- Model base class

---

## Next Steps

1. Setup repository GoIgniter v2
2. Implement core: Router (radix tree)
3. Implement Context dengan pooling
4. Implement Middleware chain
5. Implement Controller registry & base controller
6. Implement built-in middlewares
7. Implement view/template rendering
8. Testing & documentation
