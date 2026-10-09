---
title: Instalasi
description: Cara menginstall dan memulai project Goigniter.
sidebar:
  order: 2
---

## Prasyarat

Pastikan Go sudah terinstall:

```bash
go version
```

## Quick Start

### Opsi 1: Setup Wizard (Rekomendasi)

Jalankan satu perintah untuk membuat project baru:

```bash
curl -fsSL https://raw.githubusercontent.com/semutdev/goigniter/main/setup/setup.sh | bash
```

Wizard akan meminta:
- Nama project
- Tipe database (SQLite / MySQL)
- Kredensial MySQL (jika pilih MySQL)

Setelah selesai:

```bash
cd namaproject
go mod tidy
go run main.go
```

### Opsi 2: Download Starter

Clone repository dan gunakan starter template:

```bash
git clone https://github.com/semutdev/goigniter
cd goigniter/starter
```

### Copy Environment

```bash
cp .env.example .env
```

### Install Dependencies

```bash
go mod tidy
```

### Jalankan

```bash
go run main.go
```

Buka http://localhost:8080 — halaman selamat datang akan tampil.

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

| Direktori | Fungsi |
|-----------|--------|
| `application/controllers/` | Handler HTTP request |
| `application/views/` | Template HTML |
| `public/` | Asset statis |
| `main.go` | Entry point aplikasi |

## Hello World

Tambah route baru di `main.go`:

```go
app.GET("/hello", func(c *core.Context) error {
    return c.JSON(200, core.Map{
        "message": "Hello World!",
    })
})
```

Restart server dan buka http://localhost:8080/hello.

## Hot Reload (Opsional)

Gunakan [Air](https://github.com/cosmtrek/air) untuk auto-restart saat file berubah:

```bash
go install github.com/cosmtrek/air@latest
air
```

---

Lanjut ke [Routing](/id/guide/03-routing) untuk belajar cara mendefinisikan routes.

Migrasi dari CodeIgniter 3? Lihat [Agentic Migration](/id/guide/11-agentic) untuk migrasi otomatis dengan AI.