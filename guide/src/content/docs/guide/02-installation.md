---
title: Instalasi
description: Cara menginstall dan memulai project GoIgniter.
sidebar:
  order: 2
---

## Prasyarat

Sebelum memulai, pastikan kamu sudah menginstall:

- **Go 1.20** atau yang lebih baru - [Download Go](https://go.dev/dl/)
- **Text editor** - VS Code dengan [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) sangat direkomendasikan
- **Terminal** - Command prompt, PowerShell, atau terminal favorit kamu

Untuk mengecek versi Go:

```bash
go version
# Output: go version go1.21.0 darwin/amd64 (atau sejenisnya)
```

## Cara Install

### Clone Repository

```bash
git clone https://github.com/semutdev/goigniter myapp
cd myapp
```

### Jalankan Contoh

GoIgniter menyediakan beberapa contoh di folder `examples/`:

```bash
# Contoh sederhana
go run examples/simple/main.go

# Contoh dengan auto-routing
go run examples/autoroute/main.go

# Contoh lengkap dengan database
cd examples/full-crud
go mod tidy
go run main.go
```

Buka browser ke `http://localhost:8080` dan kamu akan melihat hasilnya.

## Struktur Folder

Jika kamu familiar dengan CI3, struktur folder GoIgniter akan terasa seperti rumah:

```
CodeIgniter 3:              GoIgniter:
───────────────             ──────────
application/                application/
├── controllers/            ├── controllers/
├── models/                 ├── models/
├── views/                  ├── views/
├── config/                 ├── config/
└── libraries/              └── libs/

system/                     system/
                            ├── core/
                            ├── middleware/
                            ├── libraries/
                            └── helpers/

index.php                   main.go
```

Perbedaan utama:
- `index.php` diganti `main.go` sebagai entry point
- Folder `system/` berisi framework core (mirip CI3, tapi jangan dimodifikasi)
- Tidak ada folder `cache/` atau `logs/` - Go handle ini secara berbeda

## Hello World

Berikut contoh `main.go` paling sederhana:

```go
package main

import (
    "goigniter/system/core"
    "goigniter/system/middleware"
)

func main() {
    // Buat aplikasi baru
    app := core.New()

    // Pasang middleware
    app.Use(middleware.Logger())
    app.Use(middleware.Recovery())

    // Route sederhana
    app.GET("/", func(c *core.Context) error {
        return c.JSON(200, core.Map{
            "message": "Hello GoIgniter!",
        })
    })

    // Jalankan server
    app.Run(":8080")
}
```

Jalankan dengan:

```bash
go run main.go
```

Buka `http://localhost:8080` di browser, kamu akan melihat response JSON:

```json
{"message": "Hello GoIgniter!"}
```

## Hot Reload untuk Development (Opsional)

Secara default, kamu harus restart server setiap kali ada perubahan kode. Untuk development yang lebih nyaman, gunakan [Air](https://github.com/cosmtrek/air):

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Jalankan dengan hot reload
air
```

Sekarang setiap kali kamu save file `.go`, server akan otomatis restart.

---

Sudah berhasil running? Lanjut ke [Routing](/guide/03-routing) untuk belajar cara mendefinisikan routes.
