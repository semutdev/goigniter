---
title: Template Engine
description: Cara menggunakan template untuk render HTML di GoIgniter.
sidebar:
  order: 6
---

Template di GoIgniter menggunakan package `html/template` bawaan Go. Sintaksnya berbeda dari PHP, tapi konsepnya sama - menampilkan data dinamis dalam HTML.

## Setup Template

Di `main.go`, load templates dari folder views:

```go
func main() {
    app := core.New()

    // Load templates
    // Parameter kedua: true = reload setiap request (development)
    //                  false = cache templates (production)
    err := app.LoadTemplates("./application/views", true)
    if err != nil {
        log.Fatal(err)
    }

    app.AutoRoute()
    app.Run(":8080")
}
```

Struktur folder views:

```
application/views/
├── welcome.html
├── products/
│   ├── index.html
│   ├── show.html
│   └── create.html
└── admin/
    ├── layout.html
    └── dashboard.html
```

## Sintaks Dasar

Perbandingan PHP dengan Go template:

```php
<!-- CI3: views/welcome.php -->
<h1><?= $title ?></h1>
<p>Selamat datang, <?= $user['name'] ?>!</p>

<?php if ($is_admin): ?>
    <a href="/admin">Admin Panel</a>
<?php endif; ?>

<ul>
<?php foreach ($products as $product): ?>
    <li><?= $product['name'] ?> - Rp <?= $product['price'] ?></li>
<?php endforeach; ?>
</ul>
```

```html
<!-- GoIgniter: views/welcome.html -->
<h1>{{.Title}}</h1>
<p>Selamat datang, {{.User.Name}}!</p>

{{if .IsAdmin}}
    <a href="/admin">Admin Panel</a>
{{end}}

<ul>
{{range .Products}}
    <li>{{.Name}} - Rp {{.Price}}</li>
{{end}}
</ul>
```

Perbedaan utama:
- Variabel diawali dengan `.` (dot)
- Tidak ada `$`, `<?php ?>`, atau `;`
- Block diakhiri dengan `{{end}}`, bukan `endif/endforeach`

## Passing Data ke View

Dari controller, kirim data dengan `core.Map`:

```go
func (w *WelcomeController) Index() {
    w.Ctx.View("welcome", core.Map{
        "Title":    "Selamat Datang",
        "User":     user,
        "IsAdmin":  true,
        "Products": products,
    })
}
```

Di template, akses dengan `.NamaKey`:

```html
<h1>{{.Title}}</h1>
```

## Kondisional

### If-Else

```html
{{if .IsLoggedIn}}
    <p>Halo, {{.Username}}!</p>
    <a href="/logout">Logout</a>
{{else}}
    <a href="/login">Login</a>
{{end}}
```

### If-Else If-Else

```html
{{if eq .Role "admin"}}
    <span class="badge">Admin</span>
{{else if eq .Role "editor"}}
    <span class="badge">Editor</span>
{{else}}
    <span class="badge">User</span>
{{end}}
```

### Operator Perbandingan

```html
{{if eq .Status "active"}}   <!-- sama dengan -->
{{if ne .Status "deleted"}}  <!-- tidak sama dengan -->
{{if lt .Count 10}}          <!-- kurang dari -->
{{if le .Count 10}}          <!-- kurang dari atau sama dengan -->
{{if gt .Count 0}}           <!-- lebih dari -->
{{if ge .Count 1}}           <!-- lebih dari atau sama dengan -->
```

### Logical Operators

```html
{{if and .IsLoggedIn .IsAdmin}}
    <!-- Login DAN admin -->
{{end}}

{{if or .IsAdmin .IsEditor}}
    <!-- Admin ATAU editor -->
{{end}}

{{if not .IsDeleted}}
    <!-- Tidak deleted -->
{{end}}
```

## Looping

### Range untuk Slice/Array

```html
<ul>
{{range .Products}}
    <li>{{.Name}} - {{.Price}}</li>
{{end}}
</ul>
```

### Range dengan Index

```html
<ul>
{{range $index, $product := .Products}}
    <li>{{$index}}. {{$product.Name}}</li>
{{end}}
</ul>
```

### Range dengan Else (untuk empty)

```html
{{range .Products}}
    <li>{{.Name}}</li>
{{else}}
    <li>Tidak ada produk.</li>
{{end}}
```

## Template Functions Bawaan

GoIgniter menyediakan beberapa fungsi template:

```html
<!-- Huruf besar -->
{{upper .Name}}
<!-- Output: JOHN DOE -->

<!-- Huruf kecil -->
{{lower .Name}}
<!-- Output: john doe -->

<!-- Title case -->
{{title .Name}}
<!-- Output: John Doe -->

<!-- Trim whitespace -->
{{trim .Description}}

<!-- Render HTML tanpa escape -->
{{safe .HtmlContent}}

<!-- Cek apakah string mengandung substring -->
{{if contains .Email "@gmail.com"}}
    Gmail user
{{end}}

<!-- Replace string -->
{{replace .Text "foo" "bar"}}

<!-- Split string -->
{{range split .Tags ","}}
    <span>{{.}}</span>
{{end}}

<!-- Join slice -->
{{join .Tags ", "}}
```

## Custom Template Functions

Tambahkan fungsi custom saat load templates:

```go
import "html/template"

func main() {
    app := core.New()

    funcs := template.FuncMap{
        // Format rupiah
        "rupiah": func(n int) string {
            return fmt.Sprintf("Rp %d", n)
        },

        // Format tanggal
        "formatDate": func(t time.Time) string {
            return t.Format("02 Jan 2006")
        },

        // Cek apakah slice kosong
        "empty": func(s any) bool {
            return reflect.ValueOf(s).Len() == 0
        },
    }

    app.LoadTemplatesWithFuncs("./views", true, funcs)

    app.Run(":8080")
}
```

Penggunaan di template:

```html
<p>Harga: {{rupiah .Price}}</p>
<!-- Output: Harga: Rp 150000 -->

<p>Tanggal: {{formatDate .CreatedAt}}</p>
<!-- Output: Tanggal: 13 Jan 2026 -->

{{if empty .Products}}
    <p>Tidak ada produk.</p>
{{end}}
```

## Tips Migrasi dari PHP

Tabel konversi sintaks PHP ke Go template:

| PHP | Go Template | Keterangan |
|-----|-------------|------------|
| `<?= $var ?>` | `{{.Var}}` | Tampilkan variabel |
| `<?= $arr['key'] ?>` | `{{.Arr.Key}}` | Akses map/struct |
| `<?= $arr[0] ?>` | `{{index .Arr 0}}` | Akses array by index |
| `<?php if($x): ?>` | `{{if .X}}` | Kondisional |
| `<?php else: ?>` | `{{else}}` | Else |
| `<?php endif; ?>` | `{{end}}` | Tutup block |
| `<?php foreach($items as $item): ?>` | `{{range .Items}}` | Loop |
| `<?php endforeach; ?>` | `{{end}}` | Tutup loop |
| `<?= htmlspecialchars($x) ?>` | `{{.X}}` | Auto-escaped |
| `<?= $x ?>` (raw HTML) | `{{safe .X}}` | Tanpa escape |

## Catatan Penting

1. **Auto-escape** - Semua output di-escape secara otomatis. Untuk render HTML mentah, gunakan `{{safe .Html}}`.

2. **Case-sensitive** - `.Title` berbeda dengan `.title`. Field harus diawali huruf kapital agar bisa diakses dari template (exported field).

3. **Nil handling** - Jika variabel nil, template tidak error tapi tidak menampilkan apa-apa. Gunakan `{{if .Var}}` untuk cek.

4. **Tidak ada partial (yet)** - Fitur `$this->load->view('partial')` belum ada. Coming soon di versi mendatang.

---

Selamat! Kamu sudah menguasai dasar-dasar GoIgniter. Untuk contoh aplikasi lengkap, lihat folder `examples/` di repository.
