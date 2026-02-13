---
title: Middleware
description: Cara menggunakan dan membuat middleware di GoIgniter.
sidebar:
  order: 5
---

Di CodeIgniter 3, untuk menjalankan kode sebelum atau sesudah controller, kamu menggunakan **Hooks**. Di GoIgniter, konsep ini lebih modern dan fleksibel dengan **Middleware**.

## Konsep Middleware

Middleware adalah fungsi yang membungkus handler. Ia bisa menjalankan kode sebelum request diproses, sesudahnya, atau keduanya.

```
Request
   ↓
┌─────────────────┐
│  Logger         │ ← catat waktu mulai
│  ┌───────────┐  │
│  │ Recovery  │  │ ← tangkap panic
│  │ ┌───────┐ │  │
│  │ │Handler│ │  │ ← proses request
│  │ └───────┘ │  │
│  └───────────┘  │
└─────────────────┘
   ↓
Response
```

## Built-in Middleware

GoIgniter menyediakan beberapa middleware yang siap pakai:

```go
import "goigniter/system/middleware"

func main() {
    app := core.New()

    // Log setiap request (method, path, duration)
    app.Use(middleware.Logger())

    // Tangkap panic agar server tidak crash
    app.Use(middleware.Recovery())

    // Handle CORS untuk API
    app.Use(middleware.CORS())

    // Batasi request per IP
    app.Use(middleware.RateLimit())

    app.Run(":8080")
}
```

### Logger

Mencatat setiap request ke console:

```go
app.Use(middleware.Logger())

// Output:
// [GET] /products 1.234ms <nil>
// [POST] /products 5.678ms <nil>
```

Dengan konfigurasi custom:

```go
app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Format:    "[%s] %s %v",
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### Recovery

Menangkap panic dan mengembalikan error 500 (bukan crash):

```go
app.Use(middleware.Recovery())

// Tanpa recovery: panic = server mati
// Dengan recovery: panic = response 500, server tetap jalan
```

### CORS

Menghandle Cross-Origin Resource Sharing untuk API:

```go
app.Use(middleware.CORS())
```

### RateLimit

Membatasi jumlah request per IP:

```go
app.Use(middleware.RateLimit())
```

## Global vs Group Middleware

### Global Middleware

Berlaku untuk **semua** routes:

```go
app.Use(middleware.Logger())   // Semua request di-log
app.Use(middleware.Recovery()) // Semua panic ditangkap
```

### Group Middleware

Berlaku hanya untuk routes dalam group tertentu:

```go
// Public routes - tanpa auth
app.GET("/", homeHandler)
app.GET("/products", listProducts)

// Admin routes - dengan auth
admin := app.Group("/admin", AuthMiddleware())
admin.GET("/dashboard", dashboard)  // Butuh login
admin.GET("/users", adminUsers)     // Butuh login
```

## Membuat Middleware Sendiri

Perbandingan dengan CI3 Hooks:

```php
// CI3: application/hooks/Auth_hook.php
<?php
class Auth_hook {
    public function check_login() {
        $CI =& get_instance();
        if (!$CI->session->userdata('user_id')) {
            redirect('login');
        }
    }
}

// application/config/hooks.php
$hook['post_controller_constructor'] = array(
    'class'    => 'Auth_hook',
    'function' => 'check_login',
    'filename' => 'Auth_hook.php',
    'filepath' => 'hooks'
);
```

```go
// GoIgniter: middleware/auth.go
package middleware

import "goigniter/system/core"

func AuthMiddleware() core.Middleware {
    return func(next core.HandlerFunc) core.HandlerFunc {
        return func(c *core.Context) error {
            // Cek session atau token
            userID := getSessionUserID(c)

            if userID == 0 {
                // Tidak login, redirect ke login page
                return c.Redirect(302, "/login")
            }

            // Simpan user info untuk dipakai di handler
            c.Set("user_id", userID)

            // Lanjut ke handler berikutnya
            return next(c)
        }
    }
}
```

Penggunaan:

```go
// Untuk satu group
admin := app.Group("/admin", AuthMiddleware())

// Atau untuk route tertentu
app.GET("/profile", AuthMiddleware()(profileHandler))
```

## Middleware dengan Konfigurasi

Contoh middleware yang bisa dikonfigurasi:

```go
type AuthConfig struct {
    LoginURL    string
    ExcludePath []string
}

func AuthWithConfig(config AuthConfig) core.Middleware {
    skipPaths := make(map[string]bool)
    for _, path := range config.ExcludePath {
        skipPaths[path] = true
    }

    return func(next core.HandlerFunc) core.HandlerFunc {
        return func(c *core.Context) error {
            // Skip untuk path tertentu
            if skipPaths[c.Path()] {
                return next(c)
            }

            // Cek auth
            if !isLoggedIn(c) {
                return c.Redirect(302, config.LoginURL)
            }

            return next(c)
        }
    }
}

// Penggunaan
app.Use(AuthWithConfig(AuthConfig{
    LoginURL:    "/auth/login",
    ExcludePath: []string{"/", "/about", "/contact"},
}))
```

## Controller-Level Middleware

Middleware juga bisa didefinisikan di level controller:

```go
type Admin struct {
    core.Controller
}

// Middleware untuk semua method di controller ini
func (a *Admin) Middleware() []core.Middleware {
    return []core.Middleware{
        AuthMiddleware(),
        AdminOnlyMiddleware(),
    }
}

func (a *Admin) Index() {
    // Sudah pasti user login dan admin
    a.Ctx.View("admin/dashboard", core.Map{})
}
```

## Per-Method Middleware

Untuk middleware yang hanya berlaku di method tertentu:

```go
type Product struct {
    core.Controller
}

// Middleware untuk method tertentu saja
func (p *Product) MiddlewareFor() map[string][]core.Middleware {
    return map[string][]core.Middleware{
        "Store":  {AuthMiddleware()},           // POST butuh login
        "Update": {AuthMiddleware()},           // PUT butuh login
        "Delete": {AuthMiddleware(), AdminOnly()}, // DELETE butuh admin
    }
}

func (p *Product) Index() {
    // Public - tanpa auth
}

func (p *Product) Store() {
    // Butuh login
}

func (p *Product) Delete() {
    // Butuh login + admin
}
```

## Urutan Middleware

Middleware dijalankan sesuai urutan registrasi:

```go
app.Use(middleware.Logger())   // 1. Pertama masuk, terakhir keluar
app.Use(middleware.Recovery()) // 2. Kedua masuk
app.Use(AuthMiddleware())      // 3. Ketiga masuk, pertama keluar
```

---

Middleware untuk keamanan sudah siap. Sekarang tinggal tampilkan datanya! Lanjut ke [Template Engine](/guide/06-templates).
