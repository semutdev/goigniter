---
title: Helpers
description: Utility functions available in Goigniter.
sidebar:
  order: 9
---

Helpers are utility functions accessible from templates and Go code.

## Template Helpers

These helpers are available inside HTML templates via the `helper` function:

### base_url

Returns the base URL of the application:

```html
<link rel="stylesheet" href='{{ helper "base_url" "/static/css/style.css" }}'>
```

### site_url

Returns a full URL for a given route:

```html
<a href='{{ helper "site_url" "/products/detail/5" }}'>Detail</a>
```

### asset_url

Returns a URL for static assets:

```html
<img src='{{ helper "asset_url" "/images/logo.png" }}'>
```

### is_active

Returns a CSS class if the current route matches:

```html
<a href="/" class='{{ helper "is_active" "/" "active" }}'>Home</a>
```

### excerpt

Truncates text to a specified length with ellipsis:

```html
<p>{{ helper "excerpt" .Description 100 }}</p>
```

### date

Formats a time value:

```html
<p>{{ helper "date" .CreatedAt "2006-01-02" }}</p>
```

### nl2br

Converts newlines to `<br>` tags:

```html
<p>{{ helper "nl2br" .Description }}</p>
```

## Code Helpers

These functions are available in Go code:

### Session

```go
import "goigniter/system/helpers"

// Set session data
helpers.SessionSet(c.Ctx, "user_id", "1")

// Get session data
userId := helpers.SessionGet(c.Ctx, "user_id")

// Delete session data
helpers.SessionDelete(c.Ctx, "user_id")

// Destroy entire session
helpers.SessionDestroy(c.Ctx)
```

### Encryption

```go
// Encrypt data
token, _ := helpers.Encrypt("sensitive-data")

// Decrypt data
decrypted, _ := helpers.Decrypt(token)

// Hash a password
hash, _ := helpers.HashPassword("secret123")

// Verify a password
isValid := helpers.CheckPassword("secret123", hash)
```

### Validation

```go
import "goigniter/system/helpers"

// Validate email
if helpers.ValidEmail("user@example.com") {
    // email is valid
}

// Validate URL
if helpers.ValidURL("https://example.com") {
    // URL is valid
}
```

### CSRF Protection

```go
// Generate CSRF token
token := helpers.CSRFToken(c.Ctx)

// Get CSRF field name
field := helpers.CSRFFieldName()

// Validate CSRF token
if helpers.CSRFValid(c.Ctx, token) {
    // valid token
}
```

---

Next: [Upload Library](/guide/10-upload) for file uploads and image processing.