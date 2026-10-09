---
title: Database
description: Konfigurasi dan koneksi database di Goigniter.
sidebar:
  order: 7
---

Goigniter menggunakan GORM untuk akses database. SQLite dan MySQL didukung langsung.

## Konfigurasi

Set environment variable `DB_DSN` di `.env`:

```bash
# .env
DB_DSN="root:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
APP_PORT=:8080
```

Untuk SQLite:

```bash
DB_DSN="test.db"
```

## Models

Model adalah struct Go yang merepresentasikan tabel database:

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

## Auto Migration

Model di-migrasi otomatis saat startup:

```go
config.DB.AutoMigrate(&models.Product{}, &models.User{})
```

## Query Dasar

```go
import "goigniter/config"

// Ambil semua produk
var products []models.Product
config.DB.Find(&products)

// Ambil by ID
var product models.Product
config.DB.First(&product, id)

// Ambil dengan kondisi
config.DB.Where("price > ?", 100).Find(&products)

// Buat
product := models.Product{Name: "Coffee", Price: 2.50}
config.DB.Create(&product)

// Update
config.DB.Model(&product).Update("Price", 3.00)

// Hapus
config.DB.Delete(&product)
```

## Pattern Umum

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
```

### Count

```go
var count int64
config.DB.Model(&models.Product{}).Where("price > ?", 100).Count(&count)
```

---

Lanjut ke [Query Builder](/id/guide/08-query-builder) untuk SQL query dengan method chaining.
