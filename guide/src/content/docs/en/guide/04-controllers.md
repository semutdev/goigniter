---
title: Controllers & AutoRoute
description: Create controllers with automatic route registration.
sidebar:
  order: 4
---

Controllers handle HTTP requests. With AutoRoute, registering a controller is enough — routes are created automatically based on method names.

## Basic Controller

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

Key points:
- Controllers embed `core.Controller`
- `Index()` is the default method, mapped to the controller root path
- Access request and response via `w.Ctx`

## Register & AutoRoute

```go
func main() {
    app := core.New()

    core.Register(&Welcome{})
    core.Register(&Product{})
    core.Register(&Dashboard{}, "admin") // with prefix

    app.AutoRoute()
    app.Run(":8080")
}
```

Routes are automatically generated from controller and method names.

## Method Routing

| Method | HTTP Methods | URL Pattern |
|--------|--------------|-------------|
| `Index()` | GET, POST | `/{controller}` |
| `Store()` | POST | `/{controller}/store` |
| `Create()` | GET, POST | `/{controller}/create` |
| Any other method | GET, POST | `/{controller}/{method}` |

## Controlling HTTP Methods

Restrict which HTTP methods a controller action accepts:

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

func (a *Auth) Login() {
    a.Ctx.View("auth/login", core.Map{"Title": "Login"})
}

func (a *Auth) Dologin() {
    email := a.Ctx.FormValue("email")
    password := a.Ctx.FormValue("password")
    // authenticate
}
```

All methods accept GET and POST by default.

## Custom Route Patterns

For routes with parameters like `:id`, define custom patterns:

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

Routes starting with `/` are absolute paths:

```go
func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Detail": "/item/:id",    // absolute: /item/:id
        "Edit":   "edit/:id",     // relative: /product/edit/:id
    }
}
```

Complete example:

```go
type Product struct {
    core.Controller
}

func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Edit":   "edit/:id",
        "Update": "update/:id",
        "Delete": "delete/:id",
    }
}

func (p *Product) Index() {
    products := getProductsFromDB()
    p.Ctx.View("products/index", core.Map{"Products": products})
}

func (p *Product) Edit() {
    id := p.Ctx.Param("id")
    product := getProductByID(id)
    p.Ctx.View("products/edit", core.Map{"Product": product})
}

func (p *Product) Update() {
    id := p.Ctx.Param("id")
    name := p.Ctx.FormValue("name")
    updateProduct(id, name)
    p.Ctx.Redirect(302, "/product")
}

func (p *Product) Delete() {
    id := p.Ctx.Param("id")
    deleteProduct(id)
    p.Ctx.JSON(200, core.Map{"success": true})
}
```

## Nested Controllers (Prefix)

Add a prefix when registering:

```go
core.Register(&Dashboard{}, "admin")
core.Register(&User{}, "admin")
```

Results:

```
Product     → /product
Dashboard   → /admin/dashboard
User        → /admin/user
```

## Accessing Request Data

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

## Response Types

```go
p.Ctx.JSON(200, core.Map{"status": "success"})
p.Ctx.View("products/index", core.Map{"products": products})
p.Ctx.String(200, "Hello World")
p.Ctx.Redirect(302, "/login")
p.Ctx.File("/path/to/file.pdf")
p.Ctx.NoContent(204)
```

---

Next: [Middleware](/guide/05-middleware) for authentication, logging, and security.