package middleware

import (
	"fmt"

	"github.com/semutdev/goigniter/system/core"
)

// SecurityConfig holds configuration for security headers.
type SecurityConfig struct {
	// XSSProtection enables X-XSS-Protection header
	XSSProtection bool
	// ContentTypeNosniff enables X-Content-Type-Options header
	ContentTypeNosniff bool
	// XFrameOptions sets X-Frame-Options header (DENY, SAMEORIGIN, ALLOW-FROM)
	XFrameOptions string
	// HSTS enables Strict-Transport-Security header
	HSTS bool
	// HSTSMaxAge sets max-age for HSTS (default: 31536000 = 1 year)
	HSTSMaxAge int
	// HSTSIncludeSubdomains includes subdomains in HSTS
	HSTSIncludeSubdomains bool
	// HSTSPreload enables HSTS preload
	HSTSPreload bool
	// ContentSecurityPolicy sets CSP header
	ContentSecurityPolicy string
	// ReferrerPolicy sets Referrer-Policy header
	ReferrerPolicy string
	// PermissionsPolicy sets Permissions-Policy header
	PermissionsPolicy string
}

// DefaultSecurityConfig returns a default security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		XSSProtection:         true,
		ContentTypeNosniff:     true,
		XFrameOptions:         "DENY",
		HSTS:                  false, // Only enable if using HTTPS
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		ContentSecurityPolicy: "",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "",
	}
}

// SecurityHeaders returns a security headers middleware with default config.
func SecurityHeaders() core.Middleware {
	return SecurityHeadersWithConfig(DefaultSecurityConfig())
}

// SecurityHeadersWithConfig returns a security headers middleware with custom config.
func SecurityHeadersWithConfig(config SecurityConfig) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			// X-XSS-Protection
			if config.XSSProtection {
				c.SetHeader("X-XSS-Protection", "1; mode=block")
			}

			// X-Content-Type-Options
			if config.ContentTypeNosniff {
				c.SetHeader("X-Content-Type-Options", "nosniff")
			}

			// X-Frame-Options
			if config.XFrameOptions != "" {
				c.SetHeader("X-Frame-Options", config.XFrameOptions)
			}

			// Strict-Transport-Security (HSTS)
			if config.HSTS {
				hstsValue := fmt.Sprintf("max-age=%d", config.HSTSMaxAge)
				if config.HSTSMaxAge == 0 {
					hstsValue = "max-age=31536000"
				}
				if config.HSTSIncludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if config.HSTSPreload {
					hstsValue += "; preload"
				}
				c.SetHeader("Strict-Transport-Security", hstsValue)
			}

			// Content-Security-Policy
			if config.ContentSecurityPolicy != "" {
				c.SetHeader("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				c.SetHeader("Referrer-Policy", config.ReferrerPolicy)
			}

			// Permissions-Policy
			if config.PermissionsPolicy != "" {
				c.SetHeader("Permissions-Policy", config.PermissionsPolicy)
			}

			return next(c)
		}
	}
}

// XSSProtection returns a middleware that sets X-XSS-Protection header.
func XSSProtection() core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			c.SetHeader("X-XSS-Protection", "1; mode=block")
			return next(c)
		}
	}
}

// NoSniff returns a middleware that sets X-Content-Type-Options header.
func NoSniff() core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			c.SetHeader("X-Content-Type-Options", "nosniff")
			return next(c)
		}
	}
}

// FrameGuard returns a middleware that sets X-Frame-Options header.
func FrameGuard(option string) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			c.SetHeader("X-Frame-Options", option)
			return next(c)
		}
	}
}

// HSTS returns a middleware that sets Strict-Transport-Security header.
// Only use this if your application is served over HTTPS.
func HSTS(maxAge int, includeSubdomains, preload bool) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			value := fmt.Sprintf("max-age=%d", maxAge)
			if maxAge == 0 {
				value = "max-age=31536000"
			}
			if includeSubdomains {
				value += "; includeSubDomains"
			}
			if preload {
				value += "; preload"
			}
			c.SetHeader("Strict-Transport-Security", value)
			return next(c)
		}
	}
}

// CSP returns a middleware that sets Content-Security-Policy header.
func CSP(policy string) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			c.SetHeader("Content-Security-Policy", policy)
			return next(c)
		}
	}
}

// ReferrerPolicy returns a middleware that sets Referrer-Policy header.
func ReferrerPolicyHeader(policy string) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			c.SetHeader("Referrer-Policy", policy)
			return next(c)
		}
	}
}