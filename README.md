# GoIgniter <img src="./public/logo.png" alt="GoIgniter" height="30">

Go web framework dengan ruh CodeIgniter 3.

---

## Daftar Isi

- [Fitur](#fitur)
- [Instalasi](#instalasi)
- [Quick Start](#quick-start)
- [Struktur Folder](#struktur-folder)
- [Controller](#controller)
  - [Membuat Controller](#membuat-controller)
  - [Controller dengan Subfolder](#controller-dengan-subfolder)
  - [Parameter Injection](#parameter-injection)
  - [HTTP Method Restriction](#http-method-restriction)
- [Model](#model)
  - [Membuat Model](#membuat-model)
  - [Relasi](#relasi)
  - [Query Data](#query-data)
- [View](#view)
  - [Membuat View](#membuat-view)
  - [Multi Layout](#multi-layout)
  - [Template Functions](#template-functions)
- [Authentication](#authentication)
  - [Setup](#setup-authentication)
  - [Login & Register](#login--register)
  - [Proteksi Route](#proteksi-route)
  - [Flash Message](#flash-message)
- [Middleware](#middleware)
- [Database](#database)
  - [Konfigurasi](#konfigurasi-database)
  - [Migration](#migration)
  - [Seeder](#seeder)
- [Helper Functions](#helper-functions)
- [Hot Reload](#hot-reload)
- [Environment Variables](#environment-variables)
- [Tech Stack](#tech-stack)

---

## Fitur

- **Auto-routing** - Buat controller, langsung bisa diakses di browser
- **Self-registration** - Controller register otomatis via `init()`
- **Nested folders** - Subfolder jadi URL prefix (`admin/dashboard` → `/admin/dashboard/index`)
- **Default index** - `/users` otomatis panggil `Users.Index()`
- **Auto parameter injection** - Method bisa terima `id` langsung atau via `c.Param("id")`
- **HTTP method restriction** - Optional, untuk API yang butuh strict method
- **Multi-layout template** - Support multiple layout untuk auth, admin, dll
- **Ion Auth style authentication** - Login, register, forgot password, reset password
- **Flash session** - Pesan sukses/error yang hilang setelah ditampilkan

---

## Instalasi

### Prasyarat

- Go 1.21+
- MySQL/MariaDB
- Git

### Clone Repository

```bash
git clone https://github.com/semutdev/goigniter
cd goigniter
```

### Setup Environment

```bash
cp .env.example .env
```

Edit file `.env` sesuai konfigurasi database Anda:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=goigniter

APP_PORT=:6789
APP_URL=http://localhost:6789
APP_KEY=your-secret-key-min-32-chars

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=your@email.com
```

### Install Dependencies

```bash
go mod tidy
```

### Jalankan

```bash
go run main.go
```

Buka http://localhost:6789

---

## Struktur Folder

```
goigniter/
├── config/
│   └── database.go         # Koneksi database (GORM)
├── controllers/
│   ├── admin/              # Controller admin (subfolder)
│   │   ├── dashboard.go
│   │   └── product.go
│   ├── auth.go             # Controller authentication
│   └── welcome.go          # Controller welcome
├── database/
│   └── seeder.go           # Database seeder
├── libs/
│   ├── auth.go             # Library authentication
│   ├── autoroute.go        # Auto-routing engine
│   ├── helpers.go          # Helper functions
│   ├── interfaces.go       # Interface definitions
│   ├── mail.go             # Email helper
│   └── registry.go         # Controller registry
├── models/
│   ├── user.go             # Model User
│   ├── group.go            # Model Group
│   ├── login_attempt.go    # Model Login Attempt
│   └── product.go          # Model Product
├── views/
│   ├── admin/              # Views admin
│   │   ├── layout.html     # Layout admin
│   │   ├── dashboard.html
│   │   └── product/
│   │       ├── index.html
│   │       ├── add.html
│   │       └── edit.html
│   ├── auth/               # Views authentication
│   │   ├── layout.html     # Layout auth
│   │   ├── login.html
│   │   ├── register.html
│   │   ├── forgot.html
│   │   └── reset.html
│   ├── layout.html         # Layout utama
│   └── welcome.html
├── public/                 # Static files (CSS, JS, images)
├── main.go
├── .env
└── .air.toml               # Config hot reload
```

---

## Controller

### Membuat Controller

Buat file di folder `controllers/`:

```go
// controllers/products.go
package controllers

import (
    "goigniter/libs"
    "github.com/labstack/echo/v4"
)

func init() {
    libs.Register("products", &Products{})
}

type Products struct{}

func (p *Products) Index(c echo.Context) error {
    return c.String(200, "Product list")
}

func (p *Products) Detail(c echo.Context) error {
    id := c.Param("id")
    return c.String(200, "Product #" + id)
}

func (p *Products) Add(c echo.Context) error {
    return c.String(200, "Add product form")
}

func (p *Products) Store(c echo.Context) error {
    // Simpan ke database
    return c.Redirect(303, "/products")
}
```

**URL Mapping:**

| URL | Method | Handler |
|-----|--------|---------|
| `/products` | GET | `Products.Index()` |
| `/products/detail/5` | GET | `Products.Detail()` |
| `/products/add` | GET | `Products.Add()` |
| `/products/store` | POST | `Products.Store()` |

### Controller dengan Subfolder

Untuk controller di subfolder (misal admin):

```go
// controllers/admin/dashboard.go
package admin

import (
    "goigniter/libs"
    "github.com/labstack/echo/v4"
)

func init() {
    libs.Register("admin/dashboard", &Dashboard{})
}

type Dashboard struct{}

func (d *Dashboard) Index(c echo.Context) error {
    return c.Render(200, "admin/dashboard", map[string]interface{}{
        "Title": "Dashboard",
    })
}
```

**Penting:** Tambahkan import di `main.go`:

```go
import (
    _ "goigniter/controllers"
    _ "goigniter/controllers/admin"  // tambah ini
)
```

URL: `/admin/dashboard` → `Dashboard.Index()`

### Parameter Injection

GoIgniter mendukung auto-inject parameter dari URL:

```go
// URL: /products/detail/5
func (p *Products) Detail(c echo.Context) error {
    id := c.Param("id")  // "5"
    return c.String(200, "Product #" + id)
}
```

### HTTP Method Restriction

Untuk API yang butuh strict HTTP method:

```go
type Users struct{}

func (u *Users) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Index":  {"GET"},
        "Add":    {"POST"},
        "Update": {"PUT", "PATCH"},
        "Delete": {"DELETE"},
    }
}
```

---

## Model

### Membuat Model

Model menggunakan GORM. Buat file di folder `models/`:

```go
// models/product.go
package models

import "time"

type Product struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"size:255;not null" json:"name"`
    Price     float64   `gorm:"not null" json:"price"`
    Stock     int       `gorm:"default:0" json:"stock"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

Tambahkan ke AutoMigrate di `main.go`:

```go
config.DB.AutoMigrate(
    &models.User{},
    &models.Product{},  // tambah ini
)
```

### Relasi

```go
// models/order.go
type Order struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `json:"user_id"`
    User      User      `gorm:"foreignKey:UserID"`  // belongs to
    Items     []OrderItem `gorm:"foreignKey:OrderID"` // has many
}

type OrderItem struct {
    ID        uint    `gorm:"primaryKey"`
    OrderID   uint    `json:"order_id"`
    ProductID uint    `json:"product_id"`
    Product   Product `gorm:"foreignKey:ProductID"`
    Quantity  int     `json:"quantity"`
}
```

### Query Data

```go
// Di controller
func (p *Products) Index(c echo.Context) error {
    var products []models.Product

    // Get all
    config.DB.Find(&products)

    // With conditions
    config.DB.Where("stock > ?", 0).Find(&products)

    // With order & limit
    config.DB.Order("created_at desc").Limit(10).Find(&products)

    // Find by ID
    var product models.Product
    config.DB.First(&product, 1)

    return c.JSON(200, products)
}
```

---

## View

### Membuat View

View menggunakan Go template. Buat file `.html` di folder `views/`:

```html
<!-- views/welcome.html -->
{{define "welcome"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>Welcome to GoIgniter!</p>
</body>
</html>
{{end}}
```

Render di controller:

```go
func (w *Welcome) Index(c echo.Context) error {
    return c.Render(200, "welcome", map[string]interface{}{
        "Title": "Home",
    })
}
```

### Multi Layout

GoIgniter mendukung multi layout dengan nama block unik.

**1. Buat Layout:**

```html
<!-- views/admin/layout.html -->
{{define "admin/layout"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}} - Admin</title>
    {{template "admin-head" .}}
</head>
<body>
    <nav><!-- Sidebar --></nav>
    <main>
        {{template "admin-content" .}}
    </main>
    {{template "admin-scripts" .}}
</body>
</html>
{{end}}

{{define "admin-head"}}{{end}}
{{define "admin-scripts"}}{{end}}
```

**2. Buat Page yang Menggunakan Layout:**

```html
<!-- views/admin/dashboard.html -->
{{define "admin/dashboard"}}
    {{template "admin/layout" .}}
{{end}}

{{define "admin-content"}}
<h1>{{.Title}}</h1>
<p>Welcome to admin dashboard!</p>
{{end}}
```

**3. Render di Controller:**

```go
return c.Render(200, "admin/dashboard", data)
```

### Template Functions

Tersedia template functions:

```html
<!-- base_url - URL dasar aplikasi -->
<a href="{{base_url}}">Home</a>
<!-- Output: http://localhost:6789 -->

<!-- base_url dengan path -->
<img src="{{base_url "static/logo.png"}}">
<!-- Output: http://localhost:6789/static/logo.png -->

<!-- site_url - Sama dengan base_url -->
<a href="{{site_url "auth/login"}}">Login</a>
<!-- Output: http://localhost:6789/auth/login -->
```

---

## Authentication

GoIgniter menyediakan sistem authentication ala Ion Auth.

### Setup Authentication

Pastikan environment variables sudah diset:

```env
APP_KEY=your-secret-key-min-32-chars
JWT_SECRET=your-jwt-secret

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASSWORD=your-app-password
```

### Login & Register

**Login:**
```go
// Di controller
func (a *Auth) Dologin(c echo.Context) error {
    email := c.FormValue("email")
    password := c.FormValue("password")
    remember := c.FormValue("remember") == "on"

    user, err := libs.Login(email, password, c)
    if err != nil {
        // Handle error
    }

    // Set session
    libs.SetSession(c, user, remember)

    return c.Redirect(303, "/admin/dashboard")
}
```

**Register:**
```go
user, err := libs.RegisterUser(email, password, firstName, lastName)
if err != nil {
    // Handle error
}
```

**Logout:**
```go
func (a *Auth) Logout(c echo.Context) error {
    libs.ClearSession(c)
    return c.Redirect(303, "/auth/login")
}
```

### Proteksi Route

Proteksi halaman yang butuh login:

```go
func (d *Dashboard) Index(c echo.Context) error {
    // Cek login
    if !libs.IsLoggedIn(c) {
        return c.Redirect(303, "/auth/login")
    }

    // Get current user
    user := libs.GetUser(c)

    return c.Render(200, "admin/dashboard", map[string]interface{}{
        "Title": "Dashboard",
        "User":  user,
    })
}
```

**Cek Group/Role:**
```go
if !libs.InGroup(c, "admin") {
    return c.String(403, "Forbidden")
}
```

### Flash Message

Untuk menampilkan pesan sukses/error sekali:

**Set Flash:**
```go
libs.SetFlash(c, "success", "Product berhasil ditambahkan")
return c.Redirect(303, "/admin/product")
```

**Get Flash (di controller):**
```go
data := map[string]interface{}{
    "Success": libs.GetFlash(c, "success"),
    "Error":   libs.GetFlash(c, "error"),
}
```

**Tampilkan di View:**
```html
{{if .Success}}
<div class="alert alert-success">{{.Success}}</div>
{{end}}

{{if .Error}}
<div class="alert alert-danger">{{.Error}}</div>
{{end}}
```

---

## Middleware

Echo middleware bisa ditambahkan di `main.go`:

```go
// Logger
e.Use(middleware.Logger())

// Recover
e.Use(middleware.Recover())

// CORS
e.Use(middleware.CORS())

// Custom middleware
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Before handler
        err := next(c)
        // After handler
        return err
    }
})
```

**Auth Middleware:**

```go
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        if !libs.IsLoggedIn(c) {
            return c.Redirect(303, "/auth/login")
        }
        return next(c)
    }
}

// Gunakan
e.GET("/admin/*", handler, AuthMiddleware)
```

---

## Database

### Konfigurasi Database

Edit file `.env`:

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=goigniter
```

Koneksi otomatis dibuat di `config/database.go`.

### Migration

AutoMigrate dijalankan otomatis saat aplikasi start:

```go
// main.go
config.DB.AutoMigrate(
    &models.User{},
    &models.Group{},
    &models.Product{},
)
```

### Seeder

Buat seeder di `database/seeder.go`:

```go
func Seed(db *gorm.DB) {
    // Cek apakah data sudah ada
    var count int64
    db.Model(&models.User{}).Count(&count)
    if count > 0 {
        return
    }

    // Insert default data
    admin := models.User{
        Email:    "admin@example.com",
        Password: hashPassword("password"),
    }
    db.Create(&admin)
}
```

Jalankan seeder dengan set `DB_SEED=true` di `.env`:

```env
DB_SEED=true
```

---

## Helper Functions

### URL Helpers

```go
// Di Go code
url := libs.BaseURL("admin/product")  // http://localhost:6789/admin/product
url := libs.SiteURL("auth/login")     // http://localhost:6789/auth/login
```

```html
<!-- Di template -->
<a href="{{base_url "admin"}}">Admin</a>
<a href="{{site_url "auth/login"}}">Login</a>
```

### Flash Session

```go
// Set flash
libs.SetFlash(c, "success", "Data saved!")
libs.SetFlash(c, "error", "Something went wrong")

// Get flash (otomatis hilang setelah dibaca)
msg := libs.GetFlash(c, "success")
```

### Auth Helpers

```go
// Cek login
if libs.IsLoggedIn(c) { ... }

// Get current user
user := libs.GetUser(c)

// Cek group
if libs.InGroup(c, "admin") { ... }

// Login
user, err := libs.Login(email, password, c)

// Register
user, err := libs.RegisterUser(email, password, firstName, lastName)

// Logout
libs.ClearSession(c)
```

---

## Hot Reload

Gunakan Air untuk hot reload saat development:

**Install:**
```bash
go install github.com/cosmtrek/air@v1.49.0
```

**Tambahkan ke PATH:**
```bash
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
source ~/.zshrc
```

**Jalankan:**
```bash
air
```

Setiap perubahan file `.go`, `.html`, `.css`, `.js` akan otomatis restart server.

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Server port | `:6789` |
| `APP_URL` | Base URL aplikasi | `http://localhost:6789` |
| `APP_KEY` | Secret key untuk session | - |
| `DB_HOST` | Database host | `127.0.0.1` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database username | `root` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `goigniter` |
| `DB_SEED` | Run seeder on start | `false` |
| `JWT_SECRET` | Secret untuk JWT token | - |
| `SMTP_HOST` | SMTP server host | - |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USER` | SMTP username | - |
| `SMTP_PASSWORD` | SMTP password | - |
| `SMTP_FROM` | Email sender address | - |

---

## Tech Stack

- [Echo](https://echo.labstack.com/) - High performance web framework
- [GORM](https://gorm.io/) - ORM library
- [Bootstrap 5](https://getbootstrap.com/) - CSS framework
- [DataTables](https://datatables.net/) - Table plugin
- [HTMX](https://htmx.org/) - Frontend interactivity (optional)
- [Air](https://github.com/cosmtrek/air) - Hot reload

---

## License

MIT License

---

Made with ❤️ by [SemutDev](https://github.com/semutdev)
