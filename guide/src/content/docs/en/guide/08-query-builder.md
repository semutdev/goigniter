---
title: Query Builder
description: Build SQL queries using method chaining in Goigniter.
sidebar:
  order: 8
---

GORM provides a fluent query builder. This page covers common query patterns.

## Basic Queries

```go
import "goigniter/config"

var product models.Product

config.DB.First(&product)              // First record
config.DB.Last(&product)               // Last record
config.DB.Take(&product)               // Random record
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

// WHERE price > ? AND category = ?
config.DB.Where("price > ? AND category = ?", 50, "food").Find(&products)
```

## Struct & Map Conditions

```go
// Struct: uses non-zero fields
config.DB.Where(&models.Product{Name: "Coffee", Price: 2.50}).First(&product)

// Map: includes zero values
config.DB.Where(map[string]interface{}{"name": "Coffee", "price": 0}).Find(&products)
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

## Aggregations

```go
type Stats struct {
    TotalProducts int64
    AveragePrice  float64
    MaxPrice      float64
}

var stats Stats
config.DB.Model(&models.Product{}).
    Select("COUNT(*) as total_products, AVG(price) as average_price, MAX(price) as max_price").
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

## Transactions

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

// Load all products with their categories
config.DB.Preload("Category").Find(&products)

// Load with condition
config.DB.Preload("Category", "active = ?", true).Find(&products)

// Nested preload
config.DB.Preload("Category.Translations").Find(&products)
```

## Scopes (Reusable Query Parts)

```go
func ActiveProducts(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "active")
}

func PriceGreaterThan(min float64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("price > ?", min)
    }
}

// Usage
config.DB.Scopes(ActiveProducts, PriceGreaterThan(100)).Find(&products)
```

---

Next: [Helpers](/en/guide/09-helpers) for utility functions.
