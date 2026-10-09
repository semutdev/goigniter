---
title: Routing
description: Cara mendefinisikan routes di Goigniter.
sidebar:
  order: 3
---

Route memetakan HTTP request ke handler function. Goigniter mendukung registrasi route secara eksplisit dan auto-routing via controller.

## Route Dasar

```go
app.GET("/products", listProducts)
app.GET("/products/:id", showProduct)
app.GET("/products/create", createProductForm)
```

Setiap route menentukan HTTP method dan path pattern.

## HTTP Methods

Semua method HTTP standar didukung:

```go
app.GET("/users", getUsers)
app.POST("/users", createUser)
app.PUT("/users/:id", updateUser)
app.PATCH("/users/:id", patchUser)
app.DELETE("/users/:id", deleteUser)
app.OPTIONS("/users", optionsUser)
app.HEAD("/users", headUsers)
```

## Route Parameters

Ambil segmen URL dinamis dengan `:name`:

```go
app.GET("/users/:id", func(c *core.Context) error {
    id := c.Param("id")
    return c.JSON(200, core.Map{"user_id": id})
})

// GET /users/123 → {"user_id": "123"}
```

Multiple parameter:

```go
app.GET("/users/:userId/posts/:postId", func(c *core.Context) error {
    userId := c.Param("userId")
    postId := c.Param("postId")
    return c.JSON(200, core.Map{"user_id": userId, "post_id": postId})
})
```

## Query Parameters

Akses nilai query string:

```go
app.GET("/search", func(c *core.Context) error {
    keyword := c.Query("q")
    page := c.QueryDefault("page", "1")

    return c.JSON(200, core.Map{
        "keyword": keyword,
        "page":    page,
    })
})

// GET /search?q=golang&page=2
```

## Route Groups

Group route dengan prefix dan middleware yang sama:

```go
api := app.Group("/api/v1")
api.GET("/users", listUsers)
api.POST("/users", createUser)

admin := app.Group("/admin", authMiddleware, adminOnlyMiddleware)
admin.GET("/dashboard", dashboard)
admin.GET("/users", adminUsers)
```

## Static Files

Sajikan file statis:

```go
app.Static("/static/", "./public")
```

Struktur file:

```
public/
├── css/
│   └── style.css
├── js/
│   └── app.js
└── images/
    └── logo.png
```

---

Routing manual bagus untuk API. Untuk aplikasi dengan banyak controller, **AutoRoute** membuat route otomatis. Lanjut ke [Controller & AutoRoute](/id/guide/04-controllers).