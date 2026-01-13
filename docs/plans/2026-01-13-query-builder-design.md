# Query Builder Library Design

## Overview

Query builder library untuk GoIgniter sebagai alternatif GORM. Menggunakan gaya CI3 dengan method chaining, plus dukungan raw query.

## Goals

- Menggantikan kebutuhan GORM dengan library yang lebih ringan
- API mirip CI3 `$this->db->select()->from()->where()`
- Support multiple database drivers
- Zero external dependencies (hanya `database/sql` stdlib)

## Phased Implementation

1. **Fase 1**: SQLite driver
2. **Fase 2**: MySQL driver
3. **Fase 3**: PostgreSQL driver

## File Structure

```
system/libraries/database/
├── database.go          # Interface & factory
├── builder.go           # Query builder (chaining)
├── result.go            # Result handling (struct/map)
├── drivers/
│   ├── driver.go        # Driver interface
│   ├── sqlite.go        # SQLite driver (fase 1)
│   ├── mysql.go         # MySQL driver (fase 2)
│   └── postgres.go      # PostgreSQL driver (fase 3)
└── database_test.go     # Unit tests
```

## API Design

### Connection

```go
// SQLite
db, err := database.Open("sqlite", "./database.db")

// MySQL (fase 2)
db, err := database.Open("mysql", "user:pass@tcp(localhost:3306)/dbname")

// PostgreSQL (fase 3)
db, err := database.Open("postgres", "postgres://user:pass@localhost/dbname")

// Global instance (opsional)
database.SetDefault(db)
```

### Select Query

```go
// Basic select
db.Table("users").Get(&users)

// Select specific columns
db.Table("users").Select("id", "name", "email").Get(&users)

// Where conditions
db.Table("users").Where("status", "active").Get(&users)
db.Table("users").Where("age", ">", 18).Get(&users)

// Multiple where (AND)
db.Table("users").Where("status", "active").Where("role", "admin").Get(&users)

// Or where
db.Table("users").Where("role", "admin").OrWhere("role", "editor").Get(&users)

// Where In
db.Table("users").WhereIn("id", []int{1, 2, 3}).Get(&users)

// Order & Limit
db.Table("users").OrderBy("created_at", "DESC").Limit(10).Offset(20).Get(&users)

// Get single row
db.Table("users").Where("id", 1).First(&user)

// Get as map (tanpa struct)
results := db.Table("users").GetMap()
```

### Insert

```go
// Insert single row
db.Table("users").Insert(map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
})

// Insert dan get ID
id := db.Table("users").InsertGetId(map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
})

// Insert dari struct
db.Table("users").InsertStruct(&user)
```

### Update

```go
// Update dengan where
db.Table("users").Where("id", 1).Update(map[string]any{
    "name": "Jane Doe",
})

// Update dari struct
db.Table("users").Where("id", 1).UpdateStruct(&user)
```

### Delete

```go
db.Table("users").Where("id", 1).Delete()
db.Table("users").WhereIn("id", []int{1, 2, 3}).Delete()
```

### Join

```go
// Inner join
db.Table("users").
    Select("users.name", "orders.total").
    Join("orders", "users.id", "=", "orders.user_id").
    Get(&results)

// Left join
db.Table("users").
    LeftJoin("orders", "users.id", "=", "orders.user_id").
    Get(&results)
```

### Aggregates

```go
count := db.Table("users").Where("status", "active").Count()
total := db.Table("orders").Where("user_id", 1).Sum("amount")
avg := db.Table("products").Avg("price")
min := db.Table("products").Min("price")
max := db.Table("products").Max("price")
```

### Transaction

```go
// Transaction dengan callback
err := db.Transaction(func(tx *database.DB) error {
    tx.Table("users").Insert(map[string]any{"name": "John"})
    tx.Table("wallets").Insert(map[string]any{"user_id": 1, "balance": 0})
    return nil // Commit, atau return error untuk rollback
})

// Manual transaction
tx := db.Begin()
tx.Table("users").Insert(...)
tx.Commit() // atau tx.Rollback()
```

### Raw Query

```go
// Raw select
db.Query("SELECT * FROM users WHERE status = ?", "active").Get(&users)

// Raw execute
db.Exec("UPDATE users SET status = ? WHERE id = ?", "inactive", 1)

// Raw with map result
results := db.Query("SELECT * FROM users").GetMap()
```

## Driver Interface

```go
type Driver interface {
    Open(dsn string) (*sql.DB, error)
    Placeholder(index int) string  // SQLite/MySQL: "?", Postgres: "$1"
}
```

## Result Handling

Support dua jenis output:
1. **Struct** - Scan langsung ke struct Go
2. **Map** - `[]map[string]any` untuk fleksibilitas

## Error Handling

```go
err := db.Table("users").Get(&users)
if err != nil {
    // handle error
}

// Atau cek last error
if db.Error() != nil {
    // handle
}
```

## Implementation Notes

- Gunakan `database/sql` stdlib
- Reflection untuk struct scanning
- Prepared statements untuk keamanan
- Connection pooling dari `sql.DB`
