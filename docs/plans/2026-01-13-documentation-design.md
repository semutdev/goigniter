# GoIgniter Documentation Design

## Overview

Dokumentasi step-by-step untuk GoIgniter framework, ditujukan untuk developer PHP/CodeIgniter 3 yang ingin migrasi ke Go.

## Target Audience

Developer PHP/CI3 yang:
- Sudah familiar dengan struktur CI3 (controllers, models, views)
- Ingin memanfaatkan performa Go
- Mencari kurva belajar yang rendah

## Pendekatan Penulisan

- **Bahasa:** Indonesia yang baik dan benar
- **Tone:** Santai tapi informatif
- **Contoh kode:** Progressive
  - Side-by-side (CI3 vs GoIgniter) untuk konsep dasar
  - GoIgniter only untuk topik advanced

## Struktur Dokumentasi

```
guide/src/content/docs/guide/
├── 01-intro.md        # Mengapa GoIgniter?
├── 02-installation.md # Instalasi
├── 03-routing.md      # Routing
├── 04-controllers.md  # Controller & AutoRoute
├── 05-middleware.md   # Middleware
└── 06-templates.md    # Template Engine
```

---

## Konten Per Bagian

### 1. Intro (01-intro.md)

**Judul:** Mengapa GoIgniter?

**Alur narasi:**
1. Opening - "Kamu sudah bertahun-tahun pakai CI3, nyaman dengan strukturnya"
2. Masalah - PHP interpreted, memory per-request, scaling butuh banyak server
3. Solusi - GoIgniter: pindah ke Go tanpa kehilangan rasa rumah
4. Keunggulan Go - compiled, memory efisien, concurrent
5. Bonus Modern Stack - single binary, type safety, zero dependency HTTP

**Panjang:** ~300-400 kata

---

### 2. Instalasi (02-installation.md)

**Judul:** Instalasi

**Konten:**
1. Prasyarat (Go 1.20+, editor, terminal)
2. Dua cara install:
   - Clone starter project (recommended)
   - Manual via go get
3. Struktur folder (side-by-side dengan CI3)
4. Hello World
5. Hot reload dengan air (opsional)

**Panjang:** ~250 kata + code blocks

---

### 3. Routing (03-routing.md)

**Judul:** Routing

**Konten:**
1. Pengantar (CI3 routes.php vs GoIgniter main.go)
2. Basic routes (side-by-side)
3. HTTP Methods (GET, POST, PUT, DELETE, PATCH)
4. Route parameters
5. Route groups
6. Static files
7. Teaser ke AutoRoute

**Panjang:** ~300 kata + code blocks

---

### 4. Controller & AutoRoute (04-controllers.md)

**Judul:** Controller & AutoRoute

**Konten:**
1. Pengantar - fitur andalan GoIgniter
2. Basic controller (side-by-side)
3. Register & AutoRoute - dua langkah
4. Tabel CRUD mapping otomatis:
   - Index() → GET /controller
   - Show() → GET /controller/:id
   - Create() → GET /controller/create
   - Store() → POST /controller
   - Edit() → GET /controller/:id/edit
   - Update() → PUT /controller/:id
   - Delete() → DELETE /controller/:id
5. Custom method → otomatis GET
6. Nested controller (admin prefix)
7. Akses request data
8. Response types (JSON, View, String, Redirect)

**Panjang:** ~400 kata + code blocks (terpanjang, fitur utama)

---

### 5. Middleware (05-middleware.md)

**Judul:** Middleware

**Konten:**
1. Pengantar (CI3 hooks vs GoIgniter middleware)
2. Konsep middleware (diagram flow)
3. Built-in middleware:
   - Logger()
   - Recovery()
   - CORS()
   - RateLimit()
4. Global vs Group middleware
5. Buat middleware sendiri (side-by-side dengan CI3 hooks)
6. Controller-level middleware
7. Per-method middleware

**Panjang:** ~350 kata + code blocks

---

### 6. Template Engine (06-templates.md)

**Judul:** Template Engine

**Konten:**
1. Pengantar (Go html/template standar)
2. Setup template:
   ```go
   // true = reload setiap request (development)
   // false = cache (production)
   app.LoadTemplates("./views", true)
   ```
3. Sintaks dasar (side-by-side PHP vs Go template)
4. Passing data ke view
5. Kondisional (if/else)
6. Template functions bawaan (upper, lower, safe, trim)
7. Custom template functions
8. Tips migrasi dari PHP (tabel konversi sintaks)

**Panjang:** ~350 kata + code blocks

---

## Update index.mdx

Landing page sudah ada, perlu minor update untuk link ke guide baru.

## Astro Config

Update `astro.config.mjs` sidebar jika diperlukan untuk menampilkan semua guide dengan urutan benar.

---

## Implementation Notes

- Setiap file menggunakan frontmatter `sidebar.order` untuk urutan navigasi
- Code blocks menggunakan syntax highlighting (go, php, html, bash)
- Tabel menggunakan markdown standard
