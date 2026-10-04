# GoIgniter Security Guide

This document outlines the security features and best practices when using GoIgniter framework.

## Table of Contents

1. [Security Headers Middleware](#security-headers-middleware)
2. [CSRF Protection](#csrf-protection)
3. [Rate Limiting](#rate-limiting)
4. [Session Security](#session-security)
5. [File Upload Security](#file-upload-security)
6. [CORS Configuration](#cors-configuration)
7. [Error Handling in Production](#error-handling-in-production)
8. [Request Body Limits](#request-body-limits)

---

## Security Headers Middleware

The `SecurityHeaders` middleware adds essential security headers to all responses.

### Usage

```go
package main

import (
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/middleware"
)

func main() {
    app := core.New()
    
    // Basic security headers
    app.Use(middleware.SecurityHeaders())
    
    // Or with custom configuration
    app.Use(middleware.SecurityHeadersWithConfig(middleware.SecurityConfig{
        XSSProtection:         true,
        ContentTypeNosniff:    true,
        XFrameOptions:         "DENY",
        HSTS:                  true,  // Only for HTTPS
        HSTSMaxAge:            31536000,
        HSTSIncludeSubdomains: true,
        ContentSecurityPolicy: "default-src 'self'",
        ReferrerPolicy:        "strict-origin-when-cross-origin",
    }))
    
    app.Run(":8080")
}
```

### Headers Added

| Header | Value | Purpose |
|--------|-------|---------|
| X-XSS-Protection | `1; mode=block` | Enables browser XSS filter |
| X-Content-Type-Options | `nosniff` | Prevents MIME type sniffing |
| X-Frame-Options | `DENY` | Prevents clickjacking |
| Strict-Transport-Security | `max-age=31536000` | Forces HTTPS (when enabled) |
| Referrer-Policy | `strict-origin-when-cross-origin` | Controls referrer information |

---

## CSRF Protection

CSRF (Cross-Site Request Forgery) protection is essential for web applications with forms.

### Basic Usage

```go
package main

import (
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/middleware"
)

func main() {
    app := core.New()
    
    // Enable CSRF protection
    app.Use(middleware.CSRF())
    
    // For forms, include CSRF token
    app.GET("/form", func(c *core.Context) error {
        token := middleware.CSRFToken(c)
        html := `<form method="POST">
            <input type="hidden" name="csrf_token" value="` + token + `">
            <input type="text" name="data">
            <button type="submit">Submit</button>
        </form>`
        return c.HTML(200, html)
    })
    
    app.POST("/form", func(c *core.Context) error {
        // CSRF is automatically validated
        return c.String(200, "Success!")
    })
    
    app.Run(":8080")
}
```

### Skip CSRF for API Routes

```go
// Skip CSRF for API routes (using token authentication instead)
app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
    Skipper: middleware.CSRFAPISkipper("/api"),
}))

// Or skip specific paths
app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
    Skipper: middleware.CSRFSkipper("/api", "/webhook"),
}))
```

### AJAX Requests

For AJAX requests, include the token in a header:

```javascript
// Get token from meta tag or cookie
const token = document.querySelector('meta[name="csrf-token"]').content;

fetch('/api/data', {
    method: 'POST',
    headers: {
        'X-CSRF-Token': token,
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({ data: 'example' })
});
```

---

## Rate Limiting

Rate limiting protects against DoS attacks and abuse.

### Basic Usage

```go
// Limit to 100 requests per minute
app.Use(middleware.RateLimit(100, time.Minute))

// Limit to 10 requests per second
app.Use(middleware.RateLimit(10, time.Second))
```

### Custom Configuration

```go
app.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Max:     100,
    Window:  time.Minute,
    Message: "Too many requests, please try again later",
    KeyFunc: func(c *core.Context) string {
        // Use IP + User ID for more granular limiting
        return c.IP() + "_" + c.GetString("user_id")
    },
}))
```

### Different Limits for Different Routes

```go
// Global rate limit
app.Use(middleware.RateLimit(1000, time.Minute))

// API rate limit (stricter)
apiGroup := app.Group("/api")
apiGroup.Use(middleware.RateLimit(100, time.Minute))
```

---

## Session Security

Sessions support both HMAC signing and AES-256 encryption.

### Basic Session (Signed Only)

```go
package main

import (
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/session"
)

func main() {
    // Initialize session with secret key
    session.Init(session.Config{
        Secret:      "your-secret-key-at-least-32-characters-long",
        CookieName:  "goigniter_session",
        MaxAge:      86400, // 24 hours
        Secure:      true,  // Set true in production (HTTPS)
        SameSite:    http.SameSiteStrictMode,
    })
    
    app := core.New()
    
    app.GET("/", func(c *core.Context) error {
        sess := session.Get(c)
        sess.Set("user_id", 123)
        sess.Save(c)
        return c.String(200, "Session saved")
    })
    
    app.Run(":8080")
}
```

### Encrypted Session (Recommended for Production)

```go
import (
    "github.com/semutdev/goigniter/system/libraries/session"
)

func main() {
    // Generate a secure encryption key (32 bytes for AES-256)
    encryptKey := session.MustGenerateKey(32)
    
    session.Init(session.Config{
        Secret:      "your-secret-key-at-least-32-characters-long",
        Encrypt:     true,
        EncryptKey:  encryptKey,
        Secure:      true,  // HTTPS only
        SameSite:    http.SameSiteStrictMode,
    })
    
    // ... rest of application
}
```

### Security Best Practices for Sessions

1. **Always use HTTPS** (`Secure: true`)
2. **Use Strict SameSite** (`SameSiteStrictMode`)
3. **Encrypt sensitive data** (`Encrypt: true`)
4. **Use strong secret keys** (minimum 32 bytes)
5. **Regenerate session after login**

---

## File Upload Security

The upload library has built-in security protections.

### Basic Usage

```go
import (
    "github.com/semutdev/goigniter/system/libraries/upload"
)

func handleUpload(c *core.Context) error {
    uploader := upload.New(upload.Config{
        UploadPath:   "./uploads",
        AllowedTypes: "jpg|jpeg|png|gif|pdf",
        MaxSize:      5000, // 5MB in KB
        FileName:     "random",
        CreateDirs:   true,
    })
    
    result, err := uploader.Do("file", c.Request)
    if err != nil {
        // Handle specific errors
        switch err {
        case upload.ErrFileTooBig:
            return c.String(400, "File too large")
        case upload.ErrInvalidType:
            return c.String(400, "File type not allowed")
        case upload.ErrSuspiciousFile:
            return c.String(400, "Suspicious file detected")
        case upload.ErrMimeMismatch:
            return c.String(400, "File content doesn't match extension")
        }
        return c.String(500, "Upload failed")
    }
    
    return c.JSON(200, result)
}
```

### Security Features

1. **Dangerous extension blocking** - PHP, EXE, BAT, SH, etc. are blocked
2. **Double extension detection** - `file.php.jpg` is blocked
3. **Null byte injection prevention** - `file.php%00.jpg` is blocked
4. **Path traversal prevention** - `../` in filename is blocked
5. **MIME type validation** - Extension must match content type
6. **Magic byte validation** - Actual file content is checked

### Additional Recommendations

1. **Store uploads outside web root**
2. **Use random filenames** to prevent enumeration
3. **Set proper file permissions** (not executable)
4. **Scan for viruses** using ClamAV or similar
5. **Limit file size** appropriately

---

## CORS Configuration

CORS (Cross-Origin Resource Sharing) should be configured carefully.

### Secure Configuration

```go
// Recommended: Explicitly list allowed origins
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins:     []string{"https://yourdomain.com", "https://app.yourdomain.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           86400,
}))

// For development only (not recommended for production)
app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"http://localhost:3000"},
}))
```

### Warning

- **Never use `AllowOrigins: []string{"*"}` with `AllowCredentials: true`**
- Default configuration now has empty `AllowOrigins` for security

---

## Error Handling in Production

Never expose internal error details to users.

### Secure Recovery Configuration

```go
app.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    HideErrorDetails:   true,
    ProductionMessage:  "Something went wrong. Please try again later.",
    DisablePrintStack:  false, // Log stack internally, don't show to user
    LogFunc: func(c *core.Context, err any, stack []byte) {
        // Log to your monitoring system
        log.Printf("[ERROR] %v\nRequest: %s %s\nStack: %s", 
            err, c.Method(), c.Path(), stack)
    },
}))
```

---

## Request Body Limits

Prevent memory exhaustion attacks by limiting request body size.

### Usage

```go
app.POST("/api/data", func(c *core.Context) error {
    var data map[string]any
    
    // Bind with 1MB limit
    err := c.BindWithLimit(&data, 1<<20) // 1MB
    if err == core.ErrBodyTooLarge {
        return c.String(413, "Request body too large")
    }
    if err != nil {
        return c.String(400, "Invalid request")
    }
    
    return c.JSON(200, data)
})
```

---

## Complete Security Setup Example

```go
package main

import (
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/semutdev/goigniter/system/core"
    "github.com/semutdev/goigniter/system/libraries/session"
    "github.com/semutdev/goigniter/system/middleware"
)

func main() {
    // Session setup with encryption
    session.Init(session.Config{
        Secret:      os.Getenv("SESSION_SECRET"),
        Encrypt:     true,
        EncryptKey:  session.MustGenerateKey(32),
        Secure:      true,
        SameSite:    http.SameSiteStrictMode,
    })
    
    app := core.New()
    
    // Security headers
    app.Use(middleware.SecurityHeaders())
    
    // Rate limiting
    app.Use(middleware.RateLimit(100, time.Minute))
    
    // Recovery (hide details in production)
    app.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
        HideErrorDetails:  os.Getenv("APP_ENV") == "production",
        ProductionMessage: "Internal Server Error",
    }))
    
    // CSRF (skip for API routes)
    app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
        Skipper: middleware.CSRFAPISkipper("/api"),
    }))
    
    // CORS for specific origins only
    app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"https://yourdomain.com"},
    }))
    
    // Your routes here...
    
    log.Fatal(app.Run(":8080"))
}
```

---

## Security Checklist

- [ ] Enable Security Headers middleware
- [ ] Enable CSRF protection for web routes
- [ ] Configure Rate Limiting
- [ ] Use encrypted sessions with HTTPS
- [ ] Validate file uploads
- [ ] Configure CORS with explicit origins
- [ ] Hide error details in production
- [ ] Limit request body size
- [ ] Use HTTPS in production
- [ ] Keep dependencies updated