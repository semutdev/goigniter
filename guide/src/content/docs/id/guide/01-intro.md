---
title: Mengapa Goigniter?
description: Framework MVC ringan untuk Go dengan kesederhanaan ala CodeIgniter.
sidebar:
  order: 1
---

Go adalah bahasa yang luar biasa untuk membangun aplikasi web. Binary-nya kecil, ribuan request concurrent bisa ditangani dengan mudah, dan standard library-nya (`net/http`) sudah mencakup sebagian besar kebutuhan HTTP.

Tapi membangun aplikasi web dari nol di Go sering berarti menulis boilerplate yang sama berulang kali — routing request, parsing parameter, render template, query database.

Goigniter menangani tugas-tugas umum ini sehingga kamu bisa fokus membangun aplikasi.

## Yang Dibutuhkan Setiap Aplikasi Web

Setiap aplikasi web butuh hal yang sama:

- Routing request ke handler
- Parsing form data, query parameter, JSON body
- Rendering HTML template
- Koneksi database
- Manajemen session

Goigniter menyediakan semua ini sebagai toolkit yang terpadu — tanpa perlu assembling.

## Yang Ditawarkan Goigniter

- **Auto-routing** — daftarkan controller, method-nya langsung jadi route
- **Template engine** — `html/template` dengan helper tambahan (`base_url`, `site_url`, formatting)
- **Query builder** — method-chaining SQL builder
- **Session management** — session terenkripsi dan signed
- **Database drivers** — SQLite dan MySQL siap pakai
- **Middleware pipeline** — global, group, dan controller-level
- **Upload file** — dengan validasi dan image processing (resize, crop, thumbnail)
- **Keamanan** — CSRF protection, security headers
- **CLI setup wizard** — satu perintah curl untuk project baru
- **Skill migrasi agentic** — migrasi CodeIgniter 3 dengan AI agent

## Struktur MVC

Goigniter menggunakan struktur MVC yang sudah dikenal:

- `controllers/` — handle HTTP request
- `models/` — interaksi database
- `views/` — render HTML template

## Cocok Untuk Siapa?

Goigniter cocok jika:
- Kamu ingin membangun aplikasi web di Go tanpa menulis boilerplate
- Kamu suka struktur MVC
- Kamu familiar dengan CodeIgniter, Laravel, atau framework sejenis
- Kamu butuh framework ringan — bukan enterprise platform

---

Siap memulai? Lanjut ke [Instalasi](/id/guide/02-installation).