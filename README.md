# GoIgniter <img src="./public/logo.png" alt="GoIgniter" height="30">

Go web framework dengan ruh CodeIgniter 3.

## Fitur

- **Auto-routing** - Buat controller, langsung bisa diakses di browser
- **Self-registration** - Controller register otomatis via `init()`
- **Nested folders** - Subfolder jadi URL prefix (`admin/dashboard` → `/admin/dashboard/index`)
- **Default index** - `/users` otomatis panggil `Users.Index()`
- **Auto parameter injection** - Method bisa terima `id` langsung atau via `c.Param("id")`
- **HTTP method restriction** - Optional, untuk API yang butuh strict method

## Quick Start

```bash
# Clone & setup
git clone https://github.com/semutdev/goigniter
cd goigniter
cp .env.example .env

# Run
go run main.go
```

Buka http://localhost:6789

## Membuat Controller

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

func (p *Products) Detail(c echo.Context, id int) error {
    return c.String(200, fmt.Sprintf("Product #%d", id))
}
```

URL mapping:
- `/products` → `Products.Index()`
- `/products/detail/5` → `Products.Detail(5)`

## Controller dengan Subfolder

```go
// controllers/admin/dashboard.go
package admin

import "goigniter/libs"

func init() {
    libs.Register("admin/dashboard", &Dashboard{})
}

type Dashboard struct{}

func (d *Dashboard) Index(c echo.Context) error {
    return c.String(200, "Admin Dashboard")
}
```

Tambahkan import di `main.go`:
```go
import (
    _ "goigniter/controllers"
    _ "goigniter/controllers/admin"  // tambah ini
)
```

URL: `/admin/dashboard` → `Dashboard.Index()`

## HTTP Method Restriction (untuk API)

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

## Struktur Folder

```
goigniter/
├── config/
│   └── database.go      # Koneksi database (GORM)
├── controllers/
│   ├── init.go          # Package init
│   └── users.go         # Controller users
├── libs/
│   ├── autoroute.go     # Auto-routing engine
│   ├── interfaces.go    # Interface definitions
│   └── registry.go      # Controller registry
├── models/
│   └── user.go          # GORM models
├── views/
│   └── *.html           # Go templates
├── public/              # Static files
├── main.go
└── .env
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `127.0.0.1` |
| `DB_PORT` | Database port | `3306` |
| `DB_USER` | Database username | `root` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `goigniter` |
| `APP_PORT` | Server port | `:6789` |

## Tech Stack

- [Echo](https://echo.labstack.com/) - Web framework
- [GORM](https://gorm.io/) - ORM
- [HTMX](https://htmx.org/) - Frontend interactivity
