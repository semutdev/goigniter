---
title: Instalasi
description: Cara menginstall dan memulai project GoIgniter.
sidebar:
  order: 2
---

## Prasyarat

Pastikan kamu sudah menginstall:

- **Go 1.21** atau yang lebih baru - [Download Go](https://go.dev/dl/)

Cek versi Go:

```bash
go version
```

## Quick Start

### 1. Download

Download starter dari [GitHub Releases](https://github.com/semutdev/goigniter/releases/latest).

Atau clone langsung:

```bash
git clone https://github.com/semutdev/goigniter
cd goigniter/starter
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Jalankan

```bash
go run main.go
```

Buka http://localhost:8080 - Welcome to GoIgniter!

## Struktur Folder

```
myapp/
├── application/
│   ├── controllers/    # Controller kamu
│   └── views/          # Template HTML
├── public/             # Static files (CSS, JS, images)
├── go.mod
└── main.go             # Entry point
```

Jika kamu familiar dengan CI3:

| CodeIgniter 3 | GoIgniter |
|---------------|-----------|
| `application/controllers/` | `application/controllers/` |
| `application/views/` | `application/views/` |
| `index.php` | `main.go` |

## Hello World

Edit `main.go` untuk menambah route baru:

```go
app.GET("/hello", func(c *core.Context) error {
    return c.JSON(200, core.Map{
        "message": "Hello World!",
    })
})
```

Restart server dan buka http://localhost:8080/hello

## Hot Reload (Opsional)

Untuk auto-restart saat file berubah, gunakan [Air](https://github.com/cosmtrek/air):

```bash
go install github.com/cosmtrek/air@latest
air
```

---

Lanjut ke [Routing](/guide/03-routing) untuk belajar cara mendefinisikan routes.
