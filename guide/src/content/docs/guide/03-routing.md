---
title: Routing
description: Cara mendefinisikan routes di GoIgniter.
sidebar:
  order: 3
---

Di CodeIgniter 3, kamu mendefinisikan routes di file `application/config/routes.php`. Di GoIgniter, routes didefinisikan langsung di `main.go` atau menggunakan fitur AutoRoute (dibahas di bab selanjutnya).

## Basic Routes

Perbandingan CI3 dengan GoIgniter:

```php
// CI3: application/config/routes.php
$route['products'] = 'products/index';
$route['products/(:num)'] = 'products/show/$1';
$route['products/create'] = 'products/create';
```

```go
// GoIgniter: main.go
app.GET("/products", listProducts)
app.GET("/products/:id", showProduct)
app.GET("/products/create", createProductForm)
```

Sintaksnya lebih eksplisit - kamu langsung tahu HTTP method apa yang digunakan.

## HTTP Methods

GoIgniter mendukung semua HTTP methods standar:

```go
app.GET("/users", getUsers)        // Ambil data
app.POST("/users", createUser)     // Buat data baru
app.PUT("/users/:id", updateUser)  // Update seluruh data
app.PATCH("/users/:id", patchUser) // Update sebagian data
app.DELETE("/users/:id", deleteUser) // Hapus data
app.OPTIONS("/users", optionsUser) // CORS preflight
app.HEAD("/users", headUsers)      // Cek header saja
```

## Route Parameters

Untuk menangkap bagian dinamis dari URL, gunakan `:nama`:

```go
app.GET("/users/:id", func(c *core.Context) error {
    // Ambil parameter dari URL
    id := c.Param("id")

    return c.JSON(200, core.Map{
        "user_id": id,
    })
})

// GET /users/123 → {"user_id": "123"}
```

Kamu juga bisa punya multiple parameters:

```go
app.GET("/users/:userId/posts/:postId", func(c *core.Context) error {
    userId := c.Param("userId")
    postId := c.Param("postId")

    return c.JSON(200, core.Map{
        "user_id": userId,
        "post_id": postId,
    })
})

// GET /users/5/posts/42 → {"user_id": "5", "post_id": "42"}
```

## Query Parameters

Untuk mengambil query string (`?key=value`):

```go
app.GET("/search", func(c *core.Context) error {
    // GET /search?q=golang&page=2
    keyword := c.Query("q")           // "golang"
    page := c.QueryDefault("page", "1") // "2" atau default "1"

    return c.JSON(200, core.Map{
        "keyword": keyword,
        "page":    page,
    })
})
```

## Route Groups

Untuk mengelompokkan routes dengan prefix yang sama (misalnya untuk admin panel atau API versioning):

```go
// Tanpa middleware
api := app.Group("/api/v1")
api.GET("/users", listUsers)      // GET /api/v1/users
api.POST("/users", createUser)    // POST /api/v1/users

// Dengan middleware
admin := app.Group("/admin", authMiddleware, adminOnlyMiddleware)
admin.GET("/dashboard", dashboard)  // GET /admin/dashboard
admin.GET("/users", adminUsers)     // GET /admin/users
```

Group juga bisa nested:

```go
api := app.Group("/api")
v1 := api.Group("/v1")
v1.GET("/products", listProductsV1) // GET /api/v1/products

v2 := api.Group("/v2")
v2.GET("/products", listProductsV2) // GET /api/v2/products
```

## Static Files

Untuk serve file statis (CSS, JS, images):

```go
// URL /static/css/style.css → file ./public/css/style.css
app.Static("/static/", "./public")
```

Struktur folder:

```
public/
├── css/
│   └── style.css
├── js/
│   └── app.js
└── images/
    └── logo.png
```

## Handler Function

Setiap route membutuhkan handler function dengan signature:

```go
func(c *core.Context) error
```

Handler menerima `*core.Context` yang berisi semua informasi request dan method untuk mengirim response.

```go
app.GET("/example", func(c *core.Context) error {
    // Akses request
    method := c.Method()     // "GET"
    path := c.Path()         // "/example"
    ip := c.IP()             // IP address client
    header := c.Header("User-Agent")

    // Kirim response
    return c.JSON(200, core.Map{"status": "ok"})
})
```

## Catatan Penting

Mendefinisikan routes secara manual seperti di atas cocok untuk API atau route yang kompleks. Tapi untuk aplikasi web dengan banyak controller, cara ini bisa jadi repetitif.

Kabar baiknya: GoIgniter punya **AutoRoute** - fitur andalan yang membuat routing semudah di CI3. Cukup buat file controller, dan routes otomatis terbuat!

---

Penasaran dengan AutoRoute? Lanjut ke [Controller & AutoRoute](/guide/04-controllers).
