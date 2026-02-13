# Flexible Auto-Routing Design

## Overview

Mengubah auto-routing system agar setiap controller method secara default menerima GET dan POST requests (seperti CodeIgniter 3), dengan opsi untuk restrict/expand HTTP methods per action melalui method `AllowedMethods()` di controller.

## Design Decisions

1. **Default HTTP Methods**: GET + POST untuk semua method (kecuali RESTful special cases)
2. **Override via Controller**: Method `AllowedMethods()` return `map[string][]string`
3. **Format sederhana**: String biasa ("GET", "POST") tanpa konstanta
4. **Full controller name**: Route tetap menggunakan nama lengkap controller (e.g., `/authcontroller/login`)
5. **Backward compatible**: Controller existing tetap berfungsi, middleware system tidak berubah

## Interface Changes

### ControllerInterface

```go
type ControllerInterface interface {
    SetContext(ctx *Context)
    Middleware() []Middleware
    MiddlewareFor() map[string][]Middleware
    AllowedMethods() map[string][]string  // NEW
}
```

### Base Controller

```go
func (c *Controller) AllowedMethods() map[string][]string {
    return nil  // nil = gunakan default
}
```

## Default HTTP Methods per Action

| Method Name | Default HTTP Methods | Route Path |
|-------------|---------------------|------------|
| Index | GET, POST | /controller |
| Show | GET, POST | /controller/:id |
| Create | GET, POST | /controller/create |
| Store | POST | /controller |
| Edit | GET, POST | /controller/:id/edit |
| Update | PUT, POST | /controller/:id |
| Delete | DELETE, POST | /controller/:id |
| Lainnya | GET, POST | /controller/methodname |

## Usage Example

```go
type AuthController struct {
    core.Controller
}

// Optional - hanya define jika mau override default
func (a *AuthController) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Login":   {"GET"},   // GET only
        "Dologin": {"POST"},  // POST only
    }
}

func (a *AuthController) Login() {
    // GET /authcontroller/login
    a.Ctx.View("auth/login", core.Map{"Title": "Login"})
}

func (a *AuthController) Dologin() {
    // POST /authcontroller/dologin
}

func (a *AuthController) Logout() {
    // GET+POST /authcontroller/logout (default)
}
```

## Files to Modify

1. `system/core/interfaces.go` - Add `AllowedMethods()` to interface
2. `system/core/controller.go` - Add default implementation
3. `system/core/registry.go` - Update routing logic

## Notes

- Manual routing (`app.GET()`, `app.POST()`, etc.) tetap berfungsi normal
- Middleware system tidak terpengaruh
- Tidak ada breaking changes untuk controller existing
