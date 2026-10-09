---
title: Upload Library
description: Upload file dan image processing di Goigniter.
sidebar:
  order: 10
---

Library upload menangani upload file dengan validasi dan image processing.

## Upload Dasar

```go
import "goigniter/system/libraries"

func (p *Product) Upload() {
    c := p.Ctx

    upload, err := libraries.NewUpload(c.Request(), "file", "./public/uploads", "images")
    if err != nil {
        c.JSON(400, core.Map{"error": err.Error()})
        return
    }

    c.JSON(200, core.Map{
        "filename": upload.Filename,
        "path":     upload.Path,
        "size":     upload.Size,
    })
}
```

Parameter:
- `c.Request()` — HTTP request
- `"file"` — nama form field
- `"./public/uploads"` — direktori tujuan
- `"images"` — group tipe file yang diizinkan

## Tipe File yang Diizinkan

```go
// Upload gambar
libraries.NewUpload(c.Request(), "avatar", "./uploads", "images")
// images: jpg, jpeg, png, gif, webp

// Upload dokumen
libraries.NewUpload(c.Request(), "file", "./uploads", "docs")
// docs: pdf, doc, docx, xls, xlsx, txt

// Upload arsip
libraries.NewUpload(c.Request(), "archive", "./uploads", "archives")
// archives: zip, rar, tar, gz
```

## Validasi

Uploader memvalidasi:
- Tipe file (berdasarkan group)
- Ukuran file (default: 2MB)

```go
upload, err := libraries.NewUpload(c.Request(), "file", "./uploads", "docs")
if err != nil {
    // Tipe file tidak valid atau ukuran melebihi batas
    c.JSON(400, core.Map{"error": err.Error()})
    return
}
```

## Image Processing

Proses gambar setelah upload:

```go
// Resize ke dimensi tertentu (mempertahankan aspek rasio)
err := upload.Resize(800, 600)

// Buat thumbnail
path, err := upload.Thumbnail(150, 150)

// Crop dari tengah
path, err := upload.Crop(300, 300)
```

File baru disimpan dengan suffix:
- `Resize` → `photo_800x600.jpg`
- `Thumbnail` → `photo_thumb.jpg`
- `Crop` → `photo_cropped.jpg`

## Contoh Lengkap

```go
func (p *Product) UploadImage() {
    c := p.Ctx

    upload, err := libraries.NewUpload(c.Request(), "file", "./public/uploads", "images")
    if err != nil {
        c.JSON(400, core.Map{"error": err.Error()})
        return
    }

    thumbPath, _ := upload.Thumbnail(150, 150)

    c.JSON(200, core.Map{
        "filename":    upload.Filename,
        "path":        upload.Path,
        "thumbnail":   thumbPath,
        "size":        upload.Size,
    })
}
```

---

Lanjut ke [Agentic Migration](/id/guide/11-agentic) untuk migrasi dari CodeIgniter 3.
