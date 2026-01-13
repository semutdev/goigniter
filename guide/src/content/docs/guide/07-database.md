---
title: Database
description: Cara koneksi database dan menggunakan raw query di GoIgniter.
sidebar:
  order: 7
---

GoIgniter menyediakan library database yang ringan dan mudah digunakan. Mendukung SQLite dan MySQL, dengan API yang familiar bagi pengguna CI3.

## Driver yang Tersedia

| Driver | Library | CGO Required |
|--------|---------|--------------|
| SQLite | `modernc.org/sqlite` | Tidak (Pure Go) |
| MySQL | `github.com/go-sql-driver/mysql` | Tidak |

## Koneksi Database

### SQLite

SQLite adalah pilihan paling mudah untuk development karena tidak perlu setup server database.

```go
package main

import (
    "log"

    "goigniter/system/libraries/database"
    _ "goigniter/system/libraries/database/drivers" // Register semua drivers
)

func main() {
    // Buka koneksi SQLite
    db, err := database.Open("sqlite", "./app.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Set sebagai default (opsional, untuk akses global)
    database.SetDefault(db)
}
```

File `app.db` akan otomatis dibuat jika belum ada.

### MySQL

Untuk production, MySQL adalah pilihan yang lebih umum.

```go
package main

import (
    "log"

    "goigniter/system/libraries/database"
    _ "goigniter/system/libraries/database/drivers"
)

func main() {
    // Format DSN: user:password@tcp(host:port)/database?parseTime=true
    dsn := "root:password@tcp(localhost:3306)/myapp?parseTime=true"

    db, err := database.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    database.SetDefault(db)
}
```

:::tip[parseTime=true]
Selalu tambahkan `parseTime=true` di DSN MySQL agar kolom `DATETIME` dan `TIMESTAMP` otomatis di-parse ke `time.Time` Go.
:::

### Konfigurasi via Environment

Best practice: jangan hardcode credential database.

```go
import "os"

func main() {
    driver := os.Getenv("DB_DRIVER") // "sqlite" atau "mysql"
    dsn := os.Getenv("DB_DSN")

    if driver == "" {
        driver = "sqlite"
        dsn = "./app.db"
    }

    db, err := database.Open(driver, dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

File `.env`:
```bash
# Development (SQLite)
DB_DRIVER=sqlite
DB_DSN=./app.db

# Production (MySQL)
DB_DRIVER=mysql
DB_DSN=user:password@tcp(localhost:3306)/myapp?parseTime=true
```

## Raw Query

Untuk query SQL langsung, gunakan method `Query()` dan `Exec()`.

### SELECT dengan Query()

```go
// Query dengan parameter (prepared statement)
var users []User
db.Query("SELECT * FROM users WHERE status = ?", "active").Get(&users)

// Query ke map (tanpa struct)
results, err := db.Query("SELECT * FROM users").GetMap()
for _, row := range results {
    fmt.Println(row["name"], row["email"])
}
```

### INSERT, UPDATE, DELETE dengan Exec()

```go
// Insert
result, err := db.Exec(
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "John Doe", "john@example.com",
)
if err != nil {
    log.Fatal(err)
}
lastId, _ := result.LastInsertId()
fmt.Println("New user ID:", lastId)

// Update
result, err = db.Exec(
    "UPDATE users SET status = ? WHERE id = ?",
    "inactive", 5,
)
affected, _ := result.RowsAffected()
fmt.Println("Rows updated:", affected)

// Delete
db.Exec("DELETE FROM users WHERE status = ?", "deleted")
```

### Prepared Statements

Selalu gunakan placeholder `?` untuk parameter, jangan concatenate string langsung. Ini mencegah SQL injection.

```go
// BENAR - aman dari SQL injection
db.Query("SELECT * FROM users WHERE email = ?", userInput)

// SALAH - rentan SQL injection!
db.Query("SELECT * FROM users WHERE email = '" + userInput + "'")
```

## Membuat Tabel

Gunakan `Exec()` untuk DDL (CREATE, ALTER, DROP).

```go
// SQLite
db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        status TEXT DEFAULT 'active',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
`)

// MySQL
db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE,
        status VARCHAR(50) DEFAULT 'active',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
`)
```

:::note[Perbedaan Syntax]
SQLite dan MySQL punya perbedaan syntax untuk beberapa hal:
- Auto increment: `AUTOINCREMENT` (SQLite) vs `AUTO_INCREMENT` (MySQL)
- Tipe data: `TEXT` (SQLite) vs `VARCHAR(255)` (MySQL)
- Number: `REAL` (SQLite) vs `DECIMAL(15,2)` (MySQL)
:::

## Akses Global

Setelah memanggil `database.SetDefault(db)`, kamu bisa akses database dari mana saja:

```go
// Di controller atau file lain
import "goigniter/system/libraries/database"

func ListUsers(c *core.Context) error {
    db := database.Default()

    var users []User
    db.Query("SELECT * FROM users").Get(&users)

    return c.JSON(200, users)
}
```

---

Untuk cara yang lebih mudah dan type-safe, lihat [Query Builder](/guide/08-query-builder) di halaman berikutnya.
