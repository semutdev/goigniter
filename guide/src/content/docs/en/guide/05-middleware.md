---
title: Middleware
description: Using and creating middleware in Goigniter.
sidebar:
  order: 5
---

Middleware wraps a handler to execute code before or after the request is processed. It is useful for authentication, logging, rate limiting, and recovery.

## Middleware Pipeline

```
Request
   ↓
┌─────────────────┐
│  Logger         │ ← log start time
│  ┌───────────┐  │
│  │ Recovery  │  │ ← catch panics
│  │ ┌───────┐ │  │
│  │ │Handler│ │  │ ← process request
│  │ └───────┘ │  │
│  └───────────┘  │
└─────────────────┘
   ↓
Response
```

Middleware is executed in registration order. Each middleware can either pass the request to the next handler or short-circuit the chain (e.g., return a 401 for unauthorized requests).

## Built-in Middleware

```go
import "goigniter/system/middleware"

func main() {
    app := core.New()

    // Log every request (method, path, duration)
    app.Use(middleware.Logger())

    // Catch panics to prevent server crash
    app.Use(middleware.Recovery())

    // Handle CORS for API
    app.Use(middleware.CORS())

    // Rate limit per IP
    app.Use(middleware.RateLimit())

    app.Run(":8080")
}
```

### Logger

Logs every request to console:

```go
app.Use(middleware.Logger())

// Output:
// [GET] /products 1.234ms <nil>
// [POST] /products 5.678ms <nil>
```

Custom configuration:

```go
app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format:    "[%s] %s %v",
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### Recovery

Catches panics and returns a 500 error instead of crashing:

```go
app.Use(middleware.Recovery())
```

### CORS

Handles Cross-Origin Resource Sharing for APIs:

```go
app.Use(middleware.CORS())
```

### RateLimit

Limits requests per IP address:

```go
app.Use(middleware.RateLimit())
```

## Global vs Group Middleware

### Global

Applies to **all** routes:

```go
app.Use(middleware.Logger())   // All requests logged
app.Use(middleware.Recovery()) // All panics caught
```

### Group

Applies only to routes in a specific group:

```go
// Public routes - no auth
app.GET("/", homeHandler)
app.GET("/products", listProducts)

// Admin routes - with auth
admin := app.Group("/admin", AuthMiddleware())
admin.GET("/dashboard", dashboard)  // Requires login
admin.GET("/users", adminUsers)     // Requires login
```

## Custom Middleware

```go
package middleware

import "goigniter/system/core"

func AuthMiddleware() core.Middleware {
    return func(next core.HandlerFunc) core.HandlerFunc {
        return func(c *core.Context) error {
            // Check session or token
            userID := getSessionUserID(c)

            if userID == 0 {
                // Not logged in, redirect to login
                return c.Redirect(302, "/login")
            }

            // Store user info for the handler
            c.Set("user_id", userID)

            // Proceed to the next handler
            return next(c)
        }
    }
}
```

Usage:

```go
// Group middleware
admin := app.Group("/admin", AuthMiddleware())

// Per-route middleware
app.GET("/profile", AuthMiddleware()(profileHandler))
```

## Middleware with Configuration

```go
type AuthConfig struct {
    LoginURL    string
    ExcludePath []string
}

func AuthWithConfig(config AuthConfig) core.Middleware {
    skipPaths := make(map[string]bool)
    for _, path := range config.ExcludePath {
        skipPaths[path] = true
    }

    return func(next core.HandlerFunc) core.HandlerFunc {
        return func(c *core.Context) error {
            if skipPaths[c.Path()] {
                return next(c)
            }

            if !isLoggedIn(c) {
                return c.Redirect(302, config.LoginURL)
            }

            return next(c)
        }
    }
}

// Usage
app.Use(AuthWithConfig(AuthConfig{
    LoginURL:    "/auth/login",
    ExcludePath: []string{"/", "/about", "/contact"},
}))
```

## Controller-Level Middleware

Define middleware that applies to all methods in a controller:

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
    // Already authenticated and verified as admin
    a.Ctx.View("admin/dashboard", core.Map{})
}
```

## Per-Method Middleware

Apply middleware to specific methods:

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

func (p *Product) Index() {
    // Public - no auth
}

func (p *Product) Store() {
    // Requires login
}

func (p *Product) Delete() {
    // Requires login + admin
}
```

---

Next: [Template Engine](/en/guide/06-templates) for rendering HTML views.
