# GoIgniter Starter

Starter project untuk memulai aplikasi dengan GoIgniter.

## Cara Pakai

### 1. Copy folder ini ke lokasi project baru

```bash
cp -r starter/ ~/projects/myapp
cd ~/projects/myapp
```

### 2. Setup go.mod

```bash
mv go.mod.example go.mod
```

Edit `go.mod` dan ganti `myapp` dengan nama module project kamu.

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Jalankan

```bash
go run main.go
```

Buka browser ke `http://localhost:8080/welcomecontroller`

## Struktur Folder

```
myapp/
├── application/
│   ├── controllers/    # Tambah controller di sini
│   └── views/          # Template HTML
│       └── layouts/    # Layout utama
├── public/             # Static files (CSS, JS, images)
│   ├── css/
│   └── js/
├── main.go             # Entry point
└── go.mod
```

## Langkah Selanjutnya

1. Edit `main.go` untuk menambah controller baru
2. Buat views di `application/views/`
3. Tambah static files di `public/`

Lihat [dokumentasi lengkap](https://github.com/semutdev/goigniter) untuk panduan lebih detail.
