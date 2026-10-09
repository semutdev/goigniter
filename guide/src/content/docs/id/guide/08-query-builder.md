---
title: Query Builder
description: Membangun SQL query dengan method chaining di Goigniter.
sidebar:
  order: 8
---

GORM menyediakan query builder yang fluent. Halaman ini mencakup pattern query umum.

## Query Dasar

```go
import "goigniter/config"

var product models.Product

config.DB.First(&product)              // Record pertama
config.DB.Last(&product)               // Record terakhir
config.DB.Take(&product)               // Record acak
```

## Conditions

```go
// WHERE name = 'Coffee'
config.DB.Where("name = ?", "Coffee").First(&product)

// WHERE price > 100
config.DB.Where("price > ?", 100).Find(&products)

// WHERE name LIKE '%coffee%'
config.DB.Where("name LIKE ?", "%coffee%").Find(&products)

// WHERE category_id IN (1,2,3)
config.DB.Where("category_id IN ?", []int{1, 2, 3}).Find(&products)
```

## Select

```go
config.DB.Select("name, price").Find(&products)
config.DB.Select([]string{"name", "price"}).Find(&products)
```

## Joins

```go
type Result struct {
    Name  string
    Price float64
    CategoryName string
}

var results []Result
config.DB.Model(&models.Product{}).
    Select("products.name, products.price, categories.name as category_name").
    Joins("left join categories on categories.id = products.category_id").
    Scan(&results)
```

## Agregasi

```go
type Stats struct {
    TotalProducts int64
    AveragePrice  float64
}

var stats Stats
config.DB.Model(&models.Product{}).
    Select("COUNT(*) as total_products, AVG(price) as average_price").
    Scan(&stats)
```

## Group & Having

```go
config.DB.Model(&models.Product{}).
    Select("category_id, COUNT(*) as count").
    Group("category_id").
    Having("COUNT(*) > ?", 5).
    Find(&results)
```

## Transaksi

```go
err := config.DB.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    if err := tx.Create(&payment).Error; err != nil {
        return err
    }
    return nil
})
```

## Raw SQL

```go
var products []models.Product
config.DB.Raw("SELECT * FROM products WHERE price > ?", 100).Scan(&products)
```

## Preloading (Eager Loading)

```go
type Product struct {
    ID       uint
    Name     string
    Category Category `gorm:"foreignKey:CategoryID"`
}

config.DB.Preload("Category").Find(&products)
config.DB.Preload("Category", "active = ?", true).Find(&products)
config.DB.Preload("Category.Translations").Find(&products)
```

## Scopes (Query Reusable)

```go
func ActiveProducts(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "active")
}

func PriceGreaterThan(min float64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("price > ?", min)
    }
}

config.DB.Scopes(ActiveProducts, PriceGreaterThan(100)).Find(&products)
```

---

Lanjut ke [Helpers](/id/guide/09-helpers) untuk fungsi utilitas.