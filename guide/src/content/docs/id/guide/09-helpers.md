---
title: Helpers
description: Fungsi utilitas di Goigniter.
sidebar:
  order: 9
---

Helpers adalah fungsi utilitas yang bisa diakses dari template dan kode Go.

## Template Helpers

Tersedia di dalam template HTML via fungsi `helper`:

### base_url

```html
<link rel="stylesheet" href='{{ helper "base_url" "/static/css/style.css" }}'>
```

### site_url

```html
<a href='{{ helper "site_url" "/products/detail/5" }}'>Detail</a>
```

### is_active

```html
<a href="/" class='{{ helper "is_active" "/" "active" }}'>Home</a>
```

### excerpt

Memotong teks:

```html
<p>{{ helper "excerpt" .Description 100 }}</p>
```

### date

Format waktu:

```html
<p>{{ helper "date" .CreatedAt "2006-01-02" }}</p>
```

### nl2br

Newline ke `<br>`:

```html
<p>{{ helper "nl2br" .Description }}</p>
```

## Code Helpers

Fungsi yang tersedia di kode Go:

### Session

```go
import "goigniter/system/helpers"

// Set session
helpers.SessionSet(c.Ctx, "user_id", "1")

// Get session
userId := helpers.SessionGet(c.Ctx, "user_id")

// Delete session
helpers.SessionDelete(c.Ctx, "user_id")

// Hapus semua session
helpers.SessionDestroy(c.Ctx)
```

### Enkripsi

```go
token, _ := helpers.Encrypt("data-sensitive")
decrypted, _ := helpers.Decrypt(token)
hash, _ := helpers.HashPassword("secret123")
isValid := helpers.CheckPassword("secret123", hash)
```

### Validasi

```go
if helpers.ValidEmail("user@example.com") {
    // email valid
}

if helpers.ValidURL("https://example.com") {
    // URL valid
}
```

### CSRF Protection

```go
token := helpers.CSRFToken(c.Ctx)
field := helpers.CSRFFieldName()
if helpers.CSRFValid(c.Ctx, token) {
    // token valid
}
```

---

Lanjut ke [Upload Library](/id/guide/10-upload) untuk upload file dan image processing.
