---
title: Mengapa GoIgniter?
description: Alasan di balik pembuatan GoIgniter - framework Go dengan rasa CodeIgniter.
sidebar:
  order: 1
---

Kamu sudah bertahun-tahun menggunakan CodeIgniter 3. Nyaman dengan struktur folder `controllers/`, `models/`, `views/`. Hafal luar kepala cara bikin CRUD, pasang middleware, render template. CI3 sudah seperti rumah kedua.

Tapi seiring waktu, ada yang mulai terasa kurang.

## Masalah dengan PHP

PHP adalah bahasa yang luar biasa untuk memulai. Mudah dipelajari, ekosistem besar, hosting murah di mana-mana. Tapi untuk aplikasi yang mulai berkembang, beberapa keterbatasan mulai terasa:

- **Interpreted language** - Setiap request, PHP harus membaca dan mengeksekusi kode dari awal. Meskipun ada OpCache, tetap tidak secepat compiled binary.
- **Memory per-request** - Setiap request PHP memakan memory tersendiri. 100 concurrent users = 100x memory usage.
- **Scaling horizontal** - Untuk handle traffic besar, kamu butuh banyak server PHP-FPM atau worker.

## Solusi: Pindah ke Go

Go (Golang) adalah bahasa yang dibuat Google untuk mengatasi masalah-masalah di atas:

- **Compiled binary** - Kode dikompilasi sekali, jalan ribuan kali tanpa overhead parsing.
- **Memory efisien** - Satu aplikasi Go bisa handle ribuan request dengan memory ~10-50MB.
- **Concurrent by design** - Goroutines memungkinkan handle banyak request secara paralel dengan mudah.

Tapi ada masalah: framework Go umumnya punya filosofi berbeda dari PHP. Strukturnya asing. Kurva belajarnya curam untuk developer PHP.

## GoIgniter: Rasa Rumah di Dunia Go

GoIgniter hadir untuk menjembatani gap ini. Filosofinya sederhana:

> *"Pindah ke Go tanpa kehilangan rasa rumah (CI3)"*

Apa yang familiar:
- Struktur folder `controllers/`, `models/`, `views/` - sama persis
- Pattern controller dengan method `Index()`, `Show()`, `Store()` - mirip CI3
- **Auto-routing** - Buat file controller, method langsung jadi route. Magic!

Apa yang kamu dapat:
- Performa Go yang jauh lebih cepat dari PHP
- Single binary deployment - tidak perlu install PHP, composer, atau web server terpisah
- Type safety - error ketahuan saat compile, bukan saat production
- Built on `net/http` stdlib - zero external dependency untuk HTTP handling

## Untuk Siapa GoIgniter?

GoIgniter cocok untuk:
- Developer CI3/PHP yang ingin eksplorasi Go dengan kurva belajar rendah
- Tim yang ingin migrasi bertahap dari PHP ke Go
- Project baru yang butuh performa Go tapi tim familiar dengan pattern CI

GoIgniter mungkin bukan untuk:
- Developer yang sudah mahir Go dan prefer pattern Go idiomatic
- Project yang butuh fitur enterprise kompleks (gunakan framework seperti Go-Kit)

---

Siap mencoba? Lanjut ke [Instalasi](/guide/02-installation).
