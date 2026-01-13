---
title: Controller & AutoRoute
description: Fitur andalan GoIgniter - buat controller dan routes otomatis terbuat.
sidebar:
  order: 4
---

Ini dia fitur yang membuat GoIgniter terasa seperti CodeIgniter. Kalau kamu rindu *magic routing* CI3 yang tinggal buat file controller langsung jalan - GoIgniter punya itu.

## Basic Controller

Perbandingan controller di CI3 dan GoIgniter:

```php
// CI3: application/controllers/Welcome.php
<?php
class Welcome extends CI_Controller {
    public function index() {
        $data['title'] = 'Welcome';
        $this->load->view('welcome', $data);
    }

    public function about() {
        $this->load->view('about');
    }
}
```

```go
// GoIgniter: main.go atau controller file
package main

import "goigniter/system/core"

type WelcomeController struct {
    core.Controller
}

func (w *WelcomeController) Index() {
    w.Ctx.View("welcome", core.Map{
        "Title": "Welcome",
    })
}

func (w *WelcomeController) About() {
    w.Ctx.View("about", core.Map{})
}
```

Pola yang sama:
- Controller adalah struct yang embed `core.Controller`
- Method `Index()` adalah method default
- Akses view via `w.Ctx.View()`

## Register & AutoRoute

Untuk mengaktifkan auto-routing, cukup register controller lalu panggil `AutoRoute()`:

```go
func main() {
    app := core.New()

    // Register controllers
    core.Register(&WelcomeController{})
    core.Register(&ProductController{})
    core.Register(&DashboardController{}, "admin") // dengan prefix

    // Aktifkan auto-routing
    app.AutoRoute()

    app.Run(":8080")
}
```

Itu saja! Routes akan otomatis dibuat berdasarkan nama controller dan method.

## CRUD Mapping Otomatis

Berikut method names yang di-mapping secara otomatis ke HTTP routes:

| Method | HTTP | URL |
|--------|------|-----|
| `Index()` | GET | `/welcomecontroller` |
| `Show()` | GET | `/welcomecontroller/:id` |
| `Create()` | GET | `/welcomecontroller/create` |
| `Store()` | POST | `/welcomecontroller` |
| `Edit()` | GET | `/welcomecontroller/:id/edit` |
| `Update()` | PUT | `/welcomecontroller/:id` |
| `Delete()` | DELETE | `/welcomecontroller/:id` |

Ini mengikuti konvensi RESTful yang umum digunakan.

## Custom Method

Method selain yang di atas akan otomatis di-map sebagai GET:

```go
func (w *WelcomeController) About() {
    // Otomatis: GET /welcomecontroller/about
    w.Ctx.View("about", core.Map{})
}

func (w *WelcomeController) Contact() {
    // Otomatis: GET /welcomecontroller/contact
    w.Ctx.View("contact", core.Map{})
}
```

## Nested Controller (Admin Panel)

Untuk controller dengan prefix (misalnya admin), tambahkan parameter kedua saat register:

```go
// main.go
func main() {
    app := core.New()

    // Controller biasa
    core.Register(&ProductController{})

    // Controller dengan prefix "admin"
    core.Register(&DashboardController{}, "admin")
    core.Register(&UserController{}, "admin")

    app.AutoRoute()
    app.Run(":8080")
}

// DashboardController
type DashboardController struct {
    core.Controller
}

func (d *DashboardController) Index() {
    // Route: GET /admin/dashboardcontroller
    d.Ctx.View("admin/dashboard", core.Map{})
}
```

Hasil routing:

```
ProductController     → /productcontroller
DashboardController   → /admin/dashboardcontroller
UserController        → /admin/usercontroller
```

## Akses Request Data

Di dalam method controller, akses data request via `w.Ctx`:

```go
func (p *ProductController) Store() {
    // Form data
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")

    // Query parameter
    page := p.Ctx.Query("page")

    // Route parameter (untuk :id)
    id := p.Ctx.Param("id")

    // JSON body
    var data map[string]any
    p.Ctx.Bind(&data)

    // ... proses dan simpan
}
```

## Response Types

Berbagai cara mengirim response:

```go
// JSON response
p.Ctx.JSON(200, core.Map{"status": "success"})

// HTML template
p.Ctx.View("products/index", core.Map{"products": products})

// Plain text
p.Ctx.String(200, "Hello World")

// Redirect
p.Ctx.Redirect(302, "/login")

// File download
p.Ctx.File("/path/to/file.pdf")

// No content (untuk DELETE sukses)
p.Ctx.NoContent(204)
```

## Contoh Lengkap: Product CRUD

```go
package main

import "goigniter/system/core"

func main() {
    app := core.New()

    core.Register(&ProductController{})

    app.AutoRoute()
    app.Run(":8080")
}

type ProductController struct {
    core.Controller
}

// GET /productcontroller
func (p *ProductController) Index() {
    products := getProductsFromDB()
    p.Ctx.View("products/index", core.Map{
        "Products": products,
    })
}

// GET /productcontroller/:id
func (p *ProductController) Show() {
    id := p.Ctx.Param("id")
    product := getProductByID(id)
    p.Ctx.View("products/show", core.Map{
        "Product": product,
    })
}

// GET /productcontroller/create
func (p *ProductController) Create() {
    p.Ctx.View("products/create", core.Map{})
}

// POST /productcontroller
func (p *ProductController) Store() {
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")

    saveProduct(name, price)

    p.Ctx.Redirect(302, "/productcontroller")
}

// GET /productcontroller/:id/edit
func (p *ProductController) Edit() {
    id := p.Ctx.Param("id")
    product := getProductByID(id)
    p.Ctx.View("products/edit", core.Map{
        "Product": product,
    })
}

// PUT /productcontroller/:id
func (p *ProductController) Update() {
    id := p.Ctx.Param("id")
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")

    updateProduct(id, name, price)

    p.Ctx.Redirect(302, "/productcontroller")
}

// DELETE /productcontroller/:id
func (p *ProductController) Delete() {
    id := p.Ctx.Param("id")
    deleteProduct(id)
    p.Ctx.NoContent(204)
}
```

---

Controller sudah jadi, tapi bagaimana dengan autentikasi dan logging? Lanjut ke [Middleware](/guide/05-middleware).
