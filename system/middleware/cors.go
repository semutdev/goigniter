package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/semutdev/goigniter/system/core"
)

// CORSConfig holds configuration for the CORS middleware.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration.
// WARNING: AllowOrigins defaults to empty for security.
// You should explicitly set allowed origins in production.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{}, // Empty by default - explicitly set allowed origins
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware with default config.
func CORS() core.Middleware {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom config.
// Note: When AllowCredentials is true, AllowOrigins cannot contain "*"
func CORSWithConfig(config CORSConfig) core.Middleware {
	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			origin := c.Header("Origin")

			allowOrigin := ""
			for _, o := range config.AllowOrigins {
				if o == "*" {
					// When AllowCredentials is true, we cannot use "*"
					// Instead, reflect the origin back
					if config.AllowCredentials {
						allowOrigin = origin
					} else {
						allowOrigin = "*"
					}
					break
				}
				if o == origin {
					allowOrigin = o
					break
				}
			}

			if allowOrigin == "" {
				return next(c)
			}

			c.SetHeader("Access-Control-Allow-Origin", allowOrigin)

			if config.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}

			if exposeHeaders != "" {
				c.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}

			if c.Method() == http.MethodOptions {
				c.SetHeader("Access-Control-Allow-Methods", allowMethods)
				c.SetHeader("Access-Control-Allow-Headers", allowHeaders)
				c.SetHeader("Access-Control-Max-Age", maxAge)
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}
