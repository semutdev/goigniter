---
title: Query Builder
description: Query builder dengan method chaining ala CodeIgniter di GoIgniter.
sidebar:
  order: 8
---

Query Builder adalah cara yang lebih mudah dan aman untuk menulis query database. Dengan method chaining, kamu bisa membangun query kompleks tanpa menulis SQL mentah.

## Konsep Dasar

```go
// Raw SQL (cara lama)
db.Query("SELECT * FROM users WHERE status = ? ORDER BY name ASC LIMIT 10", "active")

// Query Builder (cara baru)
db.Table("users").
    Where("status", "active").
    OrderBy("name", "ASC").
    Limit(10).
    Get(&users)
```

Query Builder:
- Lebih mudah dibaca
- Otomatis escape parameter (aman dari SQL injection)
- Method chaining yang familiar bagi pengguna CI3

## SELECT Query

### Mengambil Semua Data

```go
var users []User
err := db.Table("users").Get(&users)
```

SQL: `SELECT * FROM users`

### Select Kolom Tertentu

```go
var users []User
db.Table("users").
    Select("id", "name", "email").
    Get(&users)
```

SQL: `SELECT id, name, email FROM users`

### First (Satu Record)

```go
var users []User
db.Table("users").
    Where("id", 1).
    First(&users)

if len(users) > 0 {
    user := users[0]
    fmt.Println(user.Name)
}
```

SQL: `SELECT * FROM users WHERE id = ? LIMIT 1`

### Get sebagai Map

Jika tidak ingin membuat struct, gunakan `GetMap()`:

```go
results, err := db.Table("users").GetMap()
for _, row := range results {
    fmt.Println(row["name"], row["email"])
}
```

## WHERE Conditions

### Where Sederhana

```go
// Sama dengan (=)
db.Table("users").Where("status", "active")
// WHERE status = 'active'

// Dengan operator
db.Table("users").Where("age", ">=", 18)
// WHERE age >= 18

db.Table("users").Where("price", ">", 100000)
// WHERE price > 100000
```

### Multiple Where (AND)

```go
db.Table("users").
    Where("status", "active").
    Where("role", "admin").
    Get(&users)
// WHERE status = 'active' AND role = 'admin'
```

### Or Where

```go
db.Table("users").
    Where("status", "active").
    OrWhere("role", "admin").
    Get(&users)
// WHERE status = 'active' OR role = 'admin'
```

### Where In

```go
db.Table("users").
    WhereIn("id", []int{1, 2, 3, 4, 5}).
    Get(&users)
// WHERE id IN (1, 2, 3, 4, 5)

db.Table("products").
    WhereIn("category", []string{"electronics", "gadgets"}).
    Get(&products)
// WHERE category IN ('electronics', 'gadgets')
```

### Where Raw

Untuk kondisi kompleks yang tidak bisa di-cover method standar:

```go
db.Table("products").
    WhereRaw("price BETWEEN ? AND ?", 100000, 500000).
    Get(&products)
// WHERE price BETWEEN 100000 AND 500000

db.Table("users").
    WhereRaw("name LIKE ?", "%john%").
    Get(&users)
// WHERE name LIKE '%john%'
```

## ORDER BY dan LIMIT

```go
// Order By
db.Table("users").
    OrderBy("created_at", "DESC").
    Get(&users)
// ORDER BY created_at DESC

// Limit
db.Table("users").
    Limit(10).
    Get(&users)
// LIMIT 10

// Offset (untuk pagination)
db.Table("users").
    Limit(10).
    Offset(20).
    Get(&users)
// LIMIT 10 OFFSET 20

// Kombinasi
db.Table("products").
    Where("stock", ">", 0).
    OrderBy("price", "ASC").
    Limit(5).
    Get(&products)
```

## JOIN

```go
// Inner Join
db.Table("orders").
    Select("orders.*", "users.name as user_name").
    Join("users", "orders.user_id", "=", "users.id").
    Get(&orders)
// SELECT orders.*, users.name as user_name
// FROM orders
// INNER JOIN users ON orders.user_id = users.id

// Left Join
db.Table("users").
    LeftJoin("orders", "users.id", "=", "orders.user_id").
    Get(&results)

// Right Join
db.Table("orders").
    RightJoin("users", "orders.user_id", "=", "users.id").
    Get(&results)
```

## INSERT

### Insert dengan Map

```go
err := db.Table("users").Insert(map[string]any{
    "name":   "John Doe",
    "email":  "john@example.com",
    "status": "active",
})
```

### Insert dan Dapatkan ID

```go
id, err := db.Table("users").InsertGetId(map[string]any{
    "name":  "Jane Smith",
    "email": "jane@example.com",
})
fmt.Println("New user ID:", id)
```

### Insert dari Struct

```go
type User struct {
    ID     int64  `db:"id"`
    Name   string `db:"name"`
    Email  string `db:"email"`
    Status string `db:"status"`
}

user := User{
    Name:   "Bob Wilson",
    Email:  "bob@example.com",
    Status: "active",
}
db.Table("users").InsertStruct(&user)
```

:::tip[Tag db]
Gunakan tag `db:"column_name"` pada struct field untuk mapping ke kolom database. Field tanpa tag akan diabaikan.
:::

## UPDATE

### Update dengan Map

```go
err := db.Table("users").
    Where("id", 5).
    Update(map[string]any{
        "name":   "Updated Name",
        "status": "inactive",
    })
```

### Update dari Struct

```go
user := User{
    Name:   "New Name",
    Status: "active",
}
db.Table("users").
    Where("id", 5).
    UpdateStruct(&user)
```

### Update Multiple Records

```go
// Update semua user dengan status pending jadi active
db.Table("users").
    Where("status", "pending").
    Update(map[string]any{
        "status": "active",
    })
```

## DELETE

```go
// Delete satu record
db.Table("users").
    Where("id", 5).
    Delete()

// Delete dengan kondisi
db.Table("products").
    Where("stock", "<=", 0).
    Delete()

// Delete dengan pattern
db.Table("logs").
    WhereRaw("created_at < ?", "2024-01-01").
    Delete()
```

:::caution[Hati-hati!]
Selalu gunakan `Where()` sebelum `Delete()`. Tanpa kondisi, SEMUA data di tabel akan terhapus!
:::

## Aggregate Functions

### Count

```go
total, err := db.Table("users").Count()
fmt.Println("Total users:", total)

// Dengan kondisi
activeUsers, _ := db.Table("users").
    Where("status", "active").
    Count()
```

### Sum, Avg, Min, Max

```go
// Total revenue
totalRevenue, _ := db.Table("orders").Sum("total")

// Rata-rata harga
avgPrice, _ := db.Table("products").Avg("price")

// Harga terendah dan tertinggi
minPrice, _ := db.Table("products").Min("price")
maxPrice, _ := db.Table("products").Max("price")

// Dengan kondisi
totalStock, _ := db.Table("products").
    Where("category", "electronics").
    Sum("stock")
```

## Transaction

Transaction memastikan serangkaian operasi berhasil semua atau gagal semua (atomic).

### Callback Style (Recommended)

```go
err := db.Transaction(func(tx *database.DB) error {
    // Kurangi stock
    err := tx.Table("products").
        Where("id", productId).
        Update(map[string]any{
            "stock": newStock,
        })
    if err != nil {
        return err // Rollback
    }

    // Buat order
    err = tx.Table("orders").Insert(map[string]any{
        "user_id":    userId,
        "product_id": productId,
        "quantity":   qty,
        "total":      total,
    })
    if err != nil {
        return err // Rollback
    }

    return nil // Commit
})

if err != nil {
    log.Println("Transaction failed:", err)
}
```

### Manual Style

```go
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

// Operasi 1
err = tx.Table("users").Insert(...)
if err != nil {
    tx.Rollback()
    return err
}

// Operasi 2
err = tx.Table("orders").Insert(...)
if err != nil {
    tx.Rollback()
    return err
}

// Semua berhasil, commit
tx.Commit()
```

## Debug: Lihat SQL yang Dihasilkan

Gunakan `ToSQL()` untuk melihat query yang akan dijalankan:

```go
sql := db.Table("users").
    Where("status", "active").
    OrderBy("name", "ASC").
    Limit(10).
    ToSQL()

fmt.Println(sql)
// Output: SELECT * FROM users WHERE status = ? ORDER BY name ASC LIMIT 10
```

## Contoh Lengkap: CRUD Controller

```go
package controllers

import (
    "goigniter/system/core"
    "goigniter/system/libraries/database"
)

type UserModel struct {
    ID     int64  `db:"id" json:"id"`
    Name   string `db:"name" json:"name"`
    Email  string `db:"email" json:"email"`
    Status string `db:"status" json:"status"`
}

type User struct{}

func (u *User) Index(c *core.Context) error {
    db := database.Default()
    status := c.Query("status")

    builder := db.Table("users")
    if status != "" {
        builder = builder.Where("status", status)
    }

    var users []UserModel
    err := builder.OrderBy("name", "ASC").Get(&users)
    if err != nil {
        return c.JSON(500, core.Map{"error": err.Error()})
    }

    return c.JSON(200, core.Map{
        "users": users,
        "count": len(users),
    })
}

func (u *User) Show(c *core.Context) error {
    db := database.Default()
    id := c.Param("id")

    var users []UserModel
    db.Table("users").Where("id", id).First(&users)

    if len(users) == 0 {
        return c.JSON(404, core.Map{"error": "User not found"})
    }

    return c.JSON(200, users[0])
}

func (u *User) Store(c *core.Context) error {
    db := database.Default()

    var input struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := c.BindJSON(&input); err != nil {
        return c.JSON(400, core.Map{"error": "Invalid JSON"})
    }

    id, err := db.Table("users").InsertGetId(map[string]any{
        "name":   input.Name,
        "email":  input.Email,
        "status": "active",
    })
    if err != nil {
        return c.JSON(500, core.Map{"error": err.Error()})
    }

    return c.JSON(201, core.Map{
        "message": "User created",
        "id":      id,
    })
}

func (u *User) Update(c *core.Context) error {
    db := database.Default()
    id := c.Param("id")

    var input struct {
        Name   string `json:"name"`
        Status string `json:"status"`
    }
    c.BindJSON(&input)

    err := db.Table("users").
        Where("id", id).
        Update(map[string]any{
            "name":   input.Name,
            "status": input.Status,
        })
    if err != nil {
        return c.JSON(500, core.Map{"error": err.Error()})
    }

    return c.JSON(200, core.Map{"message": "User updated"})
}

func (u *User) Delete(c *core.Context) error {
    db := database.Default()
    id := c.Param("id")

    db.Table("users").Where("id", id).Delete()

    return c.JSON(200, core.Map{"message": "User deleted"})
}
```

---

Untuk contoh lengkap yang bisa langsung dijalankan, lihat `examples/database/` di repository GoIgniter.
