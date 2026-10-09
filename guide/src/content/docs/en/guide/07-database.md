---
title: Database
description: Database configuration and connection in Goigniter.
sidebar:
  order: 7
---

Goigniter uses GORM for database access. SQLite and MySQL are supported out of the box.

## Configuration

Set the `DB_DSN` environment variable in `.env`:

```bash
# .env
DB_DSN="root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
APP_PORT=:8080
```

For SQLite:

```bash
DB_DSN="test.db"
```

The database connection is automatically initialized when the application starts.

## Models

Models are Go structs that represent database tables:

```go
// models/product.go
package models

import "time"

type Product struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:255;not null"`
    Price       float64   `gorm:"not null"`
    Description string    `gorm:"type:text"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Tag rules:
- `gorm:"primaryKey"` — primary key
- `gorm:"size:255"` — maximum length
- `gorm:"not null"` — required field
- `gorm:"type:text"` — custom SQL type
- `gorm:"uniqueIndex"` — unique constraint

## Auto Migration

Models are auto-migrated on startup. Add models to the migration list in `main.go`:

```go
config.DB.AutoMigrate(&models.Product{}, &models.User{})
```

This creates or updates tables to match the struct definitions.

## Basic Queries

```go
import "goigniter/config"

// Get all products
var products []models.Product
config.DB.Find(&products)

// Get by ID
var product models.Product
config.DB.First(&product, id)

// Get with condition
config.DB.Where("price > ?", 100).Find(&products)

// Create
product := models.Product{Name: "Coffee", Price: 2.50}
config.DB.Create(&product)

// Update
config.DB.Model(&product).Update("Price", 3.00)

// Delete
config.DB.Delete(&product)
```

## Common Patterns

### Pagination

```go
func getProducts(page int, pageSize int) []models.Product {
    var products []models.Product
    offset := (page - 1) * pageSize
    config.DB.Offset(offset).Limit(pageSize).Find(&products)
    return products
}
```

### Ordering

```go
config.DB.Order("created_at DESC").Find(&products)
config.DB.Order("price ASC").Find(&products)
```

### Count

```go
var count int64
config.DB.Model(&models.Product{}).Where("price > ?", 100).Count(&count)
```

### First or Create

```go
config.DB.Where(models.Product{Name: "Coffee"}).FirstOrCreate(&product)
```

---

Next: [Query Builder](/guide/08-query-builder) for building SQL queries with method chaining.