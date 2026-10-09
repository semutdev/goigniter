---
title: Upload Library
description: File upload and image processing in Goigniter.
sidebar:
  order: 10
---

The upload library handles file uploads with validation and image processing.

## Basic Upload

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

Parameters:
- `c.Request()` — the HTTP request
- `"file"` — form field name
- `"./public/uploads"` — destination directory
- `"images"` — allowed file type group

## Allowed File Types

File type groups determine which files are accepted:

```go
// Upload images only
libraries.NewUpload(c.Request(), "avatar", "./uploads", "images")
// images: jpg, jpeg, png, gif, webp

// Upload documents
libraries.NewUpload(c.Request(), "file", "./uploads", "docs")
// docs: pdf, doc, docx, xls, xlsx, txt

// Upload archives
libraries.NewUpload(c.Request(), "archive", "./uploads", "archives")
// archives: zip, rar, tar, gz
```

## Validation

The uploader validates:
- File type (based on group)
- File size (default: 2MB)

```go
upload, err := libraries.NewUpload(c.Request(), "file", "./uploads", "docs")
if err != nil {
    // Invalid file type or size exceeded
    c.JSON(400, core.Map{"error": err.Error()})
    return
}
```

## Image Processing

Resize images after upload:

```go
// Resize to specific dimensions (maintains aspect ratio)
err := upload.Resize(800, 600)

// Resize to thumbnail
path, err := upload.Thumbnail(150, 150)

// Crop from center
path, err := upload.Crop(300, 300)
```

The image is saved as a new file in the same directory with a suffix:
- `Resize` → `photo_800x600.jpg`
- `Thumbnail` → `photo_thumb.jpg`
- `Crop` → `photo_cropped.jpg`

## Complete Example

```go
func (p *Product) UploadImage() {
    c := p.Ctx

    upload, err := libraries.NewUpload(c.Request(), "file", "./public/uploads", "images")
    if err != nil {
        c.JSON(400, core.Map{"error": err.Error()})
        return
    }

    // Create thumbnail
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

Next: [Agentic Migration](/guide/11-agentic) for migrating CodeIgniter 3 projects.