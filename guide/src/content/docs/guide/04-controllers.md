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
    core.Register(&Welcome{})
    core.Register(&Product{})
    core.Register(&Dashboard{}, "admin") // dengan prefix

    // Aktifkan auto-routing
    app.AutoRoute()

    app.Run(":8080")
}
```

Itu saja! Routes akan otomatis dibuat berdasarkan nama controller dan method.

## CRUD Mapping Otomatis

Berikut method names yang di-mapping secara otomatis ke HTTP routes:

| Method | HTTP Methods | URL Default |
|--------|--------------|-------------|
| `Index()` | GET, POST | `/welcome` |
| `Store()` | POST | `/welcome/store` |
| `Create()` | GET, POST | `/welcome/create` |

Untuk routes dengan parameter `:id` (seperti Edit, Update, Delete), gunakan method `Routes()` untuk mendefinisikan custom patterns. Lihat section [Routes - Custom Route Patterns](#routes---custom-route-patterns).

## Custom Method

Method selain yang di atas akan otomatis di-map sebagai GET dan POST (seperti CI3):

```go
func (w *Welcome) About() {
    // Otomatis: GET & POST /welcome/about
    w.Ctx.View("about", core.Map{})
}

func (w *Welcome) Contact() {
    // Otomatis: GET & POST /welcome/contact
    w.Ctx.View("contact", core.Map{})
}
```

## AllowedMethods - Kontrol HTTP Methods

Secara default, semua method controller menerima GET dan POST (seperti CI3). Tapi kamu bisa membatasi HTTP methods yang diizinkan dengan `AllowedMethods()`:

```go
type Auth struct {
    core.Controller
}

// Override untuk membatasi HTTP methods
func (a *Auth) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Login":    {"GET"},           // Hanya GET
        "Dologin":  {"POST"},          // Hanya POST
        "Logout":   {"GET", "POST"},   // Keduanya (explicit)
        "Api":      {"GET", "POST", "PUT"}, // Multiple methods
    }
}

func (a *Auth) Login() {
    // GET /auth/login - tampilkan form
    a.Ctx.View("auth/login", core.Map{"Title": "Login"})
}

func (a *Auth) Dologin() {
    // POST /auth/dologin - proses login
    email := a.Ctx.FormValue("email")
    password := a.Ctx.FormValue("password")
    // ... proses login
}

func (a *Auth) Profile() {
    // GET & POST /auth/profile (default, tidak perlu didefinisikan)
    a.Ctx.View("auth/profile", core.Map{})
}
```

### Default HTTP Methods

Semua method controller menerima **GET dan POST** secara default (seperti CodeIgniter 3):

| Method Name | Default HTTP Methods |
|-------------|---------------------|
| Semua method | GET, POST |

Gunakan `AllowedMethods()` jika ingin membatasi HTTP methods tertentu.

### Perbandingan dengan CI3

Di CI3, semua route menerima GET dan POST secara default. GoIgniter mengikuti behavior yang sama, tapi dengan tambahan kontrol via `AllowedMethods()`:

```php
// CI3: Tidak ada cara built-in untuk restrict HTTP methods
// Biasanya dicek manual di controller:
public function dologin() {
    if ($this->input->method() !== 'post') {
        show_404();
    }
    // ... proses login
}
```

```go
// GoIgniter: Lebih clean dengan AllowedMethods()
func (a *Auth) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Dologin": {"POST"},
    }
}

func (a *Auth) Dologin() {
    // Otomatis hanya menerima POST
    // GET akan return 404
}
```

## Routes - Custom Route Patterns

Secara default, semua method akan di-map ke `/{controller}/{method}`. Untuk custom route patterns dengan parameter (seperti `:id`), override method `Routes()`:

```go
type Product struct {
    core.Controller
}

// Routes mendefinisikan custom route patterns
func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Edit":   "edit/:id",      // /product/edit/:id
        "Update": "update/:id",    // /product/update/:id
        "Delete": "delete/:id",    // /product/delete/:id
        "Detail": "detail/:id",    // /product/detail/:id
    }
}
```

### Default vs Custom Routes

| Method | Default Route | Custom Route |
|--------|--------------|--------------|
| `Index()` | `/product` | - |
| `Edit()` | `/product/edit` | `/product/edit/:id` |
| `Update()` | `/product/update` | `/product/update/:id` |
| `Delete()` | `/product/delete` | `/product/delete/:id` |
| `Detail()` | `/product/detail` | `/product/detail/:id` |

### Absolute Routes

Jika route dimulai dengan `/`, akan dianggap sebagai absolute path (tidak relative ke controller):

```go
func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Detail": "/item/:id",           // Absolute: /item/:id
        "Edit":   "edit/:id",            // Relative: /product/edit/:id
    }
}
```

### Contoh Lengkap dengan Routes

```go
type Product struct {
    core.Controller
}

// Custom routes dengan parameter
func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Edit":   "edit/:id",
        "Update": "update/:id",
        "Delete": "delete/:id",
    }
}

// GET /product/edit/:id
func (p *Product) Edit() {
    id := p.Ctx.Param("id")
    product := getProductByID(id)
    p.Ctx.View("product/edit", core.Map{
        "Product": product,
    })
}

// POST /product/update/:id
func (p *Product) Update() {
    id := p.Ctx.Param("id")
    name := p.Ctx.FormValue("name")
    updateProduct(id, name)
    p.Ctx.Redirect(302, "/product")
}

// POST /product/delete/:id
func (p *Product) Delete() {
    id := p.Ctx.Param("id")
    deleteProduct(id)
    p.Ctx.JSON(200, core.Map{"success": true})
}
```

## Nested Controller (Admin Panel)

Untuk controller dengan prefix (misalnya admin), tambahkan parameter kedua saat register:

```go
// main.go
func main() {
    app := core.New()

    // Controller biasa
    core.Register(&Product{})

    // Controller dengan prefix "admin"
    core.Register(&Dashboard{}, "admin")
    core.Register(&User{}, "admin")

    app.AutoRoute()
    app.Run(":8080")
}

// Dashboard
type Dashboard struct {
    core.Controller
}

func (d *Dashboard) Index() {
    // Route: GET /admin/dashboard
    d.Ctx.View("admin/dashboard", core.Map{})
}
```

Hasil routing:

```
Product     → /product
Dashboard   → /admin/dashboard
User        → /admin/user
```

## Akses Request Data

Di dalam method controller, akses data request via `p.Ctx`:

```go
func (p *Product) Store() {
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

    core.Register(&Product{})

    app.AutoRoute()
    app.Run(":8080")
}

type Product struct {
    core.Controller
}

// Custom routes untuk method dengan parameter :id
func (p *Product) Routes() map[string]string {
    return map[string]string{
        "Edit":   "edit/:id",
        "Update": "update/:id",
        "Delete": "delete/:id",
    }
}

// GET /product
func (p *Product) Index() {
    products := getProductsFromDB()
    p.Ctx.View("products/index", core.Map{
        "Products": products,
    })
}

// GET /product/create
func (p *Product) Create() {
    p.Ctx.View("products/create", core.Map{})
}

// POST /product/store
func (p *Product) Store() {
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")

    saveProduct(name, price)

    p.Ctx.Redirect(302, "/product/index")
}

// GET /product/edit/:id
func (p *Product) Edit() {
    id := p.Ctx.Param("id")
    product := getProductByID(id)
    p.Ctx.View("products/edit", core.Map{
        "Product": product,
    })
}

// POST /product/update/:id
func (p *Product) Update() {
    id := p.Ctx.Param("id")
    name := p.Ctx.FormValue("name")
    price := p.Ctx.FormValue("price")

    updateProduct(id, name, price)

    p.Ctx.Redirect(302, "/product/index")
}

// POST /product/delete/:id
func (p *Product) Delete() {
    id := p.Ctx.Param("id")
    deleteProduct(id)
    p.Ctx.JSON(200, core.Map{"success": true})
}
```

---

Controller sudah jadi, tapi bagaimana dengan autentikasi dan logging? Lanjut ke [Middleware](/guide/05-middleware).
