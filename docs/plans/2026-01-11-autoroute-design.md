# GoIgniter Auto-Route Design

Tanggal: 2026-01-11

## Tujuan

Membuat sistem auto-routing ala CodeIgniter 3 di Go. Controller otomatis bisa diakses di browser tanpa registrasi manual di main.go.

## Keputusan Design

| Aspek | Keputusan |
|-------|-----------|
| Registration | Self-registration via `init()` |
| Folder structure | Nested folder = URL prefix |
| Default method | `Index` sebagai default |
| HTTP methods | Default bebas, optional restrict via interface |
| Naming | Nama register = path file relatif dari controllers/ |
| Parameters | Auto-detect: inject ke param atau pakai `c.Param()` |

## Arsitektur

### 1. Registry Global

```go
// libs/registry.go
package libs

var controllerRegistry = make(map[string]interface{})

func Register(name string, controller interface{}) {
    controllerRegistry[name] = controller
}

func GetRegistry() map[string]interface{} {
    return controllerRegistry
}
```

### 2. Self-Registration di Controller

```go
// controllers/products.go
package controllers

import "goigniter/libs"

func init() {
    libs.Register("products", &Products{})
}

type Products struct{}

func (p *Products) Index(c echo.Context) error {
    return c.String(200, "Product list")
}
```

Untuk subfolder:

```go
// controllers/admin/dashboard.go
package admin

import "goigniter/libs"

func init() {
    libs.Register("admin/dashboard", &Dashboard{})
}

type Dashboard struct{}
```

### 3. AutoRoute Enhanced

URL patterns yang didukung:

```
/:controller                    → Controller.Index()
/:controller/:method            → Controller.Method()
/:controller/:method/:id        → Controller.Method(id)
/admin/:controller/:method/:id  → admin/Controller.Method(id)
```

Logic:

```go
func AutoRoute(e *echo.Echo, registry map[string]interface{}) {
    handler := func(c echo.Context) error {
        path := c.Param("path")
        controller, method, id := parsePath(path, registry)

        if method == "" {
            method = "Index"
        }

        ctrl, exists := registry[controller]
        if !exists {
            return echo.ErrNotFound
        }

        return callMethod(ctrl, method, c, id)
    }

    e.Any("/*path", handler)
}
```

### 4. Path Parsing

```go
func parsePath(path string, registry map[string]interface{}) (controller, method, id string) {
    // Coba match dari path terpanjang
    // "admin/users/edit/123" cek:
    //   1. "admin/users/edit/123" ada di registry? tidak
    //   2. "admin/users/edit" ada? tidak
    //   3. "admin/users" ada? YA → method="edit", id="123"
}
```

### 5. Method Calling dengan Auto-Detect Parameter

```go
func callMethod(ctrl interface{}, methodName string, c echo.Context, id string) error {
    method := reflect.ValueOf(ctrl).MethodByName(strings.Title(methodName))

    var args []reflect.Value
    args = append(args, reflect.ValueOf(c))

    // Jika method punya param kedua, inject id
    if method.Type().NumIn() == 2 && id != "" {
        paramType := method.Type().In(1)
        switch paramType.Kind() {
        case reflect.String:
            args = append(args, reflect.ValueOf(id))
        case reflect.Int:
            idInt, _ := strconv.Atoi(id)
            args = append(args, reflect.ValueOf(idInt))
        }
    }

    results := method.Call(args)
    // handle return error...
}
```

Controller bisa ditulis:

```go
// Cara 1: ambil dari context
func (u *Users) Edit(c echo.Context) error {
    id := c.Param("id")
}

// Cara 2: auto-inject string
func (u *Users) Edit(c echo.Context, id string) error {}

// Cara 3: auto-inject int
func (u *Users) Edit(c echo.Context, id int) error {}
```

### 6. HTTP Method Restriction (Opsional)

```go
// libs/interfaces.go
type MethodRestrictor interface {
    AllowedMethods() map[string][]string
}
```

Contoh:

```go
type Users struct{}

func (u *Users) AllowedMethods() map[string][]string {
    return map[string][]string{
        "Index":  {"GET"},
        "Store":  {"POST"},
        "Update": {"PUT", "PATCH"},
        "Delete": {"DELETE"},
    }
}
```

## Struktur Folder

```
goigniter/
├── libs/
│   ├── registry.go      # Register() & GetRegistry()
│   ├── autoroute.go     # AutoRoute(), parsePath(), callMethod()
│   └── interfaces.go    # MethodRestrictor interface
├── controllers/
│   ├── init.go          # trigger import
│   ├── users.go         # → /users/*
│   ├── products.go      # → /products/*
│   └── admin/
│       ├── init.go
│       ├── dashboard.go # → /admin/dashboard/*
│       └── users.go     # → /admin/users/*
├── main.go
```

## URL Mapping

| URL | Controller | Method |
|-----|------------|--------|
| `/users` | users | Index |
| `/users/store` | users | Store |
| `/users/edit/5` | users | Edit(5) |
| `/admin/dashboard` | admin/dashboard | Index |
| `/admin/users/delete/3` | admin/users | Delete(3) |

## Future Enhancements

- Middleware per controller/method
- Group routes dengan shared prefix
