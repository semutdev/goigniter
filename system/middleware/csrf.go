package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/semutdev/goigniter/system/core"
)

// CSRFConfig holds configuration for CSRF protection.
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token (default: 32)
	TokenLength int
	// CookieName is the name of the CSRF cookie (default: "csrf_token")
	CookieName string
	// HeaderName is the name of the CSRF header (default: "X-CSRF-Token")
	HeaderName string
	// FormFieldName is the name of the CSRF form field (default: "csrf_token")
	FormFieldName string
	// ContextKey is the key used to store token in context (default: "csrf_token")
	ContextKey string
	// Secure cookie flag (default: false)
	Secure bool
	// HttpOnly cookie flag (default: false - must be accessible to JS)
	HttpOnly bool
	// SameSite policy (default: Lax)
	SameSite http.SameSite
	// Path for cookie (default: "/")
	Path string
	// Domain for cookie
	Domain string
	// Error handler for custom error responses
	ErrorHandler func(c *core.Context) error
	// Skipper allows skipping CSRF for certain requests
	Skipper func(c *core.Context) bool
}

// DefaultCSRFConfig returns a default CSRF configuration.
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:   32,
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "csrf_token",
		ContextKey:    "csrf_token",
		Secure:        false,
		HttpOnly:      false, // Must be accessible to JavaScript
		SameSite:      http.SameSiteLaxMode,
		Path:          "/",
		ErrorHandler:  nil,
		Skipper:       nil,
	}
}

// CSRF returns a CSRF protection middleware with default config.
func CSRF() core.Middleware {
	return CSRFWithConfig(DefaultCSRFConfig())
}

// CSRFWithConfig returns a CSRF protection middleware with custom config.
func CSRFWithConfig(config CSRFConfig) core.Middleware {
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.FormFieldName == "" {
		config.FormFieldName = "csrf_token"
	}
	if config.ContextKey == "" {
		config.ContextKey = "csrf_token"
	}
	if config.Path == "" {
		config.Path = "/"
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			// Skip if configured
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			// Safe methods don't need CSRF validation
			if isSafeMethod(c.Method()) {
				// Generate or get existing token
				token := getOrCreateToken(c, &config)
				// Store token in context for templates
				c.Set(config.ContextKey, token)
				return next(c)
			}

			// Validate CSRF token for unsafe methods
			if err := validateCSRFToken(c, &config); err != nil {
				if config.ErrorHandler != nil {
					return config.ErrorHandler(c)
				}
				return c.String(http.StatusForbidden, "Invalid or missing CSRF token")
			}

			// Token is valid, proceed
			return next(c)
		}
	}
}

// isSafeMethod checks if the HTTP method is safe (doesn't modify state)
func isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions
}

// getOrCreateToken gets existing token from cookie or creates a new one
func getOrCreateToken(c *core.Context, config *CSRFConfig) string {
	// Try to get existing token from cookie
	cookie, err := c.Cookie(config.CookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Generate new token
	token := generateCSRFToken(config.TokenLength)

	// Set cookie
	newCookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    token,
		Path:     config.Path,
		Domain:   config.Domain,
		Secure:   config.Secure,
		HttpOnly: config.HttpOnly,
		SameSite: config.SameSite,
	}
	c.SetCookie(newCookie)

	return token
}

// validateCSRFToken validates the CSRF token from request
func validateCSRFToken(c *core.Context, config *CSRFConfig) error {
	var providedToken string

	// Try header first
	providedToken = c.Header(config.HeaderName)

	// Try form field if header is empty
	if providedToken == "" {
		providedToken = c.Form(config.FormFieldName)
	}

	// Try query parameter as fallback
	if providedToken == "" {
		providedToken = c.Query(config.FormFieldName)
	}

	// No token provided
	if providedToken == "" {
		return errCSRFMissing
	}

	// Get expected token from cookie
	cookie, err := c.Cookie(config.CookieName)
	if err != nil || cookie.Value == "" {
		return errCSRFMissing
	}

	// Compare tokens (constant-time comparison)
	if !constantTimeEqual(providedToken, cookie.Value) {
		return errCSRFInvalid
	}

	return nil
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use timestamp-based token (less secure but better than nothing)
		panic("failed to generate secure random token: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// constantTimeEqual compares two strings in constant time
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// Errors
var (
	errCSRFMissing = &CSRFError{Message: "CSRF token missing"}
	errCSRFInvalid = &CSRFError{Message: "CSRF token invalid"}
)

// CSRFError represents a CSRF validation error
type CSRFError struct {
	Message string
}

func (e *CSRFError) Error() string {
	return e.Message
}

// CSRFToken returns the CSRF token from the context for use in templates
func CSRFToken(c *core.Context) string {
	token := c.GetString("csrf_token")
	return token
}

// CSRFFormField generates a hidden form field with the CSRF token
func CSRFFormField(c *core.Context) string {
	token := CSRFToken(c)
	if token == "" {
		return ""
	}
	return `<input type="hidden" name="csrf_token" value="` + token + `">`
}

// CSRFMetaTag generates a meta tag with the CSRF token for AJAX requests
func CSRFMetaTag(c *core.Context) string {
	token := CSRFToken(c)
	if token == "" {
		return ""
	}
	return `<meta name="csrf-token" content="` + token + `">`
}

// CSRFSkipper creates a skipper function for certain paths
func CSRFSkipper(paths ...string) func(c *core.Context) bool {
	return func(c *core.Context) bool {
		path := c.Path()
		for _, p := range paths {
			if strings.HasPrefix(path, p) {
				return true
			}
		}
		return false
	}
}

// CSRFAPISkipper creates a skipper for API routes (using token auth instead)
func CSRFAPISkipper(apiPrefix string) func(c *core.Context) bool {
	return func(c *core.Context) bool {
		return strings.HasPrefix(c.Path(), apiPrefix)
	}
}