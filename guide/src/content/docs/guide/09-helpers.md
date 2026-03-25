---
title: Helpers
description: Fungsi helper bawaan GoIgniter untuk mempermudah development.
sidebar:
  order: 9
---

GoIgniter menyediakan beberapa helper function yang bisa digunakan di controller maupun template.

## URL Helpers

### BaseURL & SiteURL

Menghasilkan URL lengkap berdasarkan `APP_URL` atau `APP_PORT` dari environment.

```go
import "github.com/semutdev/goigniter/system/helpers"

// Di controller
url := helpers.BaseURL()              // "http://localhost:8080"
url := helpers.BaseURL("/products")   // "http://localhost:8080/products"
url := helpers.SiteURL("admin/user")  // "http://localhost:8080/admin/user"
```

Di template:

```html
<a href="{{site_url "product/edit"}}">Edit</a>
<link rel="stylesheet" href="{{base_url "/static/css/style.css"}}">
```

### AssetURL

Menghasilkan URL untuk static assets di folder `public/`.

```go
url := helpers.AssetURL("css/style.css")  // "http://localhost:8080/public/css/style.css"
```

Di template:

```html
<link rel="stylesheet" href="{{asset_url "css/style.css"}}">
<script src="{{asset_url "js/app.js"}}"></script>
```

## Debug Helper

### PrintDebug

Helper untuk debugging data di controller. Mencetak data dengan format yang mudah dibaca.

```go
import "github.com/semutdev/goigniter/system/helpers"

func (p *Product) Edit() {
    id := p.Ctx.Param("id")
    var product models.Product
    database.Table("products").Where("id", id).First(&product)

    data := core.Map{
        "Title":   "Edit Product",
        "Product": product,
    }

    // Debug data sebelum render
    helpers.PrintDebug(data)

    p.Ctx.View("admin/inc/header", data)
    p.Ctx.View("admin/product/edit", data)
    p.Ctx.View("admin/inc/footer", data)
}
```

Output:

```
========== DEBUG START ==========
{
  "Title": "Edit Product",
  "Product": {
    "ID": 1,
    "Name": "Laptop ASUS ROG",
    "Price": 15000000,
    "Stock": 10,
    "CreatedAt": "2026-01-15T10:30:00Z",
    "UpdatedAt": "2026-01-20T14:22:00Z"
  }
}
=========== DEBUG END ===========
```

### PrintDebug dengan Nil Check

```go
helpers.PrintDebug(data["Product"])  // Aman jika nil
```

Output jika data nil:

```
========== DEBUG START ==========
Data is nil
=========== DEBUG END ===========
```

## Template Functions

Semua helper functions juga tersedia di template. Lihat [Template Engine](/guide/06-templates) untuk daftar lengkap template functions.

| Function | Description |
|----------|-------------|
| `site_url` | Generate full URL |
| `base_url` | Generate full URL |
| `asset_url` | Generate URL for public assets |
| `upper` | Uppercase string |
| `lower` | Lowercase string |
| `title` | Title case string |
| `trim` | Trim whitespace |
| `safe` | Render HTML without escape |
| `contains` | Check if string contains substring |
| `replace` | Replace all occurrences |
| `split` | Split string to slice |
| `join` | Join slice to string |
| `default` | Return default if value is empty |
| `eq` | Equal comparison |
| `ne` | Not equal comparison |

## Inisialisasi Helpers

Di `main.go`, helpers perlu diinisialisasi dengan base URL:

```go
func main() {
    // ...

    // Initialize helpers
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = ":8080"
    }
    helpers.Init("http://localhost" + port)

    // ...
}
```

---