---
title: Routing
description: How to define routes in Goigniter.
sidebar:
  order: 3
---

Routes map HTTP requests to handler functions. Goigniter supports both explicit route registration and auto-routing via controllers.

## Basic Routes

```go
app.GET("/products", listProducts)
app.GET("/products/:id", showProduct)
app.GET("/products/create", createProductForm)
```

Each route specifies the HTTP method and path pattern.

## HTTP Methods

All standard HTTP methods are supported:

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

Capture dynamic URL segments with `:name`:

```go
app.GET("/users/:id", func(c *core.Context) error {
    id := c.Param("id")
    return c.JSON(200, core.Map{"user_id": id})
})

// GET /users/123 → {"user_id": "123"}
```

Multiple parameters:

```go
app.GET("/users/:userId/posts/:postId", func(c *core.Context) error {
    userId := c.Param("userId")
    postId := c.Param("postId")
    return c.JSON(200, core.Map{"user_id": userId, "post_id": postId})
})

// GET /users/5/posts/42 → {"user_id": "5", "post_id": "42"}
```

## Query Parameters

Access query string values:

```go
app.GET("/search", func(c *core.Context) error {
    keyword := c.Query("q")                    // "golang"
    page := c.QueryDefault("page", "1")        // "2" or default "1"

    return c.JSON(200, core.Map{
        "keyword": keyword,
        "page":    page,
    })
})

// GET /search?q=golang&page=2
```

## Route Groups

Group routes with a common prefix and middleware:

```go
api := app.Group("/api/v1")
api.GET("/users", listUsers)
api.POST("/users", createUser)

admin := app.Group("/admin", authMiddleware, adminOnlyMiddleware)
admin.GET("/dashboard", dashboard)
admin.GET("/users", adminUsers)
```

Nested groups:

```go
api := app.Group("/api")
v1 := api.Group("/v1")
v1.GET("/products", listProductsV1)

v2 := api.Group("/v2")
v2.GET("/products", listProductsV2)
```

## Static Files

Serve static assets:

```go
app.Static("/static/", "./public")
```

File structure:

```
public/
├── css/
│   └── style.css
├── js/
│   └── app.js
└── images/
    └── logo.png
```

## Handler Signature

Every route handler has the same signature:

```go
func(c *core.Context) error
```

```go
app.GET("/example", func(c *core.Context) error {
    method := c.Method()
    path := c.Path()
    ip := c.IP()
    header := c.Header("User-Agent")

    return c.JSON(200, core.Map{"status": "ok"})
})
```

---

Manual routing is great for APIs. For applications with many controllers, Goigniter's **AutoRoute** creates routes automatically. Continue to [Controllers & AutoRoute](/guide/04-controllers).