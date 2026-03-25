---
title: Upload Library
description: Library untuk upload file dan gambar dengan mudah seperti CodeIgniter 3.
sidebar:
  order: 10
---

GoIgniter menyediakan library upload yang mirip dengan CodeIgniter 3 untuk mempermudah upload file dan gambar.

## Basic Upload

```go
import "github.com/semutdev/goigniter/system/libraries/upload"

func (c *Controller) Upload() {
    // Konfigurasi upload
    config := upload.Config{
        UploadPath:   "./public/uploads",
        AllowedTypes: "jpg|jpeg|png|gif",
        MaxSize:      2048, // 2MB dalam KB
        FileName:     "timestamp",
        CreateDirs:   true,
    }

    // Lakukan upload
    uploader := upload.New(config)
    result, err := uploader.Do("file", c.Ctx.Request)
    if err != nil {
        // Handle error
        return
    }

    // Gunakan result
    fmt.Println("File uploaded:", result.FileName)
    fmt.Println("File path:", result.FilePath)
}
```

## Konfigurasi

| Parameter | Type | Description |
|-----------|------|-------------|
| `UploadPath` | string | Direktori tujuan upload |
| `AllowedTypes` | string | Ekstensi yang diizinkan, dipisahkan pipa (e.g., `jpg|png|pdf`) |
| `MaxSize` | int64 | Ukuran maksimal dalam KB (0 = unlimited) |
| `FileName` | string | Cara penamaan file: `original`, `random`, `timestamp`, atau custom name |
| `Overwrite` | bool | Timpa file yang sudah ada |
| `CreateDirs` | bool | Buat direktori jika belum ada |
| `FileExt` | string | Paksa ekstensi tertentu (e.g., `.jpg`) |

## Result Properties

| Property | Type | Description |
|----------|------|-------------|
| `FileName` | string | Nama file setelah upload |
| `OriginalName` | string | Nama file asli |
| `FileType` | string | MIME type |
| `FilePath` | string | Path lengkap file |
| `FileSize` | int64 | Ukuran file dalam bytes |
| `FileExt` | string | Ekstensi file |
| `IsImage` | bool | Apakah file adalah gambar |
| `ImageWidth` | int | Lebar gambar (jika image) |
| `ImageHeight` | int | Tinggi gambar (jika image) |
| `ImageType` | string | Tipe gambar: jpeg, png, gif |

## Error Handling

```go
result, err := uploader.Do("file", c.Ctx.Request)
if err != nil {
    switch err {
    case upload.ErrNoFile:
        // Tidak ada file yang diupload
    case upload.ErrFileTooBig:
        // File terlalu besar
    case upload.ErrInvalidType:
        // Tipe file tidak diizinkan
    case upload.ErrFileExists:
        // File sudah ada
    default:
        // Error lain
    }
}
```

## Image Processing

Library juga menyediakan fitur image processing:

### Resize

```go
imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
    Source:              result.FilePath,
    Width:               800,
    Height:              600,
    MaintainAspectRatio: true,
    Quality:             85,
})

err := imgProcessor.Resize()
```

### Create Thumbnail

```go
imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
    Source:          result.FilePath,
    CreateThumbnail: true,
    ThumbnailPrefix: "thumb_",
    ThumbnailWidth:  150,
    ThumbnailHeight: 150,
})

err := imgProcessor.Resize()
```

### Crop

```go
err := imgProcessor.Crop(0, 0, 200, 200)
```

### Fill (Resize + Crop)

Resize dan crop agar pas dengan dimensi yang diinginkan:

```go
imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
    Source: result.FilePath,
    Width:  300,
    Height: 300,
})

err := imgProcessor.Fill()
```

### Rotate

Rotasi gambar (90, 180, 270 derajat):

```go
err := imgProcessor.Rotate(90)
```

## Delete Image

```go
// Hapus gambar
upload.DeleteImage("./public/uploads/products/image.jpg")

// Hapus gambar beserta thumbnail
upload.DeleteWithThumbnail("./public/uploads/products/image.jpg", "thumb_")
```

## Contoh Lengkap

```go
package admin

import (
    "net/http"
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/upload"
)

type Product struct {
    core.Controller
}

func (p *Product) Store() {
    name := p.Ctx.FormValue("name")

    // Upload configuration
    config := upload.Config{
        UploadPath:   "./public/uploads/products",
        AllowedTypes: "jpg|jpeg|png|gif|webp",
        MaxSize:      2048,
        FileName:     "timestamp",
        CreateDirs:   true,
    }

    var imageFilename string
    uploader := upload.New(config)
    result, err := uploader.Do("image", p.Ctx.Request)

    if err == nil {
        imageFilename = result.FileName

        // Create thumbnail
        imgProcessor := upload.NewImageProcessor(upload.ImageConfig{
            Source:          result.FilePath,
            CreateThumbnail: true,
            ThumbnailPrefix: "thumb_",
            ThumbnailWidth:  150,
            ThumbnailHeight: 150,
        })
        imgProcessor.Resize()
    }

    // Save to database
    // ...

    p.Ctx.Redirect(http.StatusSeeOther, "/admin/product/index")
}
```

## Perbandingan dengan CI3

```php
// CI3
$config['upload_path'] = './uploads/';
$config['allowed_types'] = 'gif|jpg|png';
$config['max_size'] = 2048;
$config['file_name'] = 'timestamp';

$this->load->library('upload', $config);
$this->upload->do_upload('userfile');
$data = $this->upload->data();
```

```go
// GoIgniter
config := upload.Config{
    UploadPath:   "./uploads",
    AllowedTypes: "gif|jpg|png",
    MaxSize:      2048,
    FileName:     "timestamp",
}

uploader := upload.New(config)
result, _ := uploader.Do("userfile", c.Ctx.Request)
// result.FileName, result.FilePath, dll.
```

---

Selanjutnya: [Helpers](/guide/09-helpers) - Fungsi helper bawaan GoIgniter.