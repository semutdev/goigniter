package middleware

import (
	"net/http"

	"goigniter/system/core"
)

// Auth returns a basic authentication middleware.
func Auth(authFunc func(c *core.Context) bool) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			if !authFunc(c) {
				return c.String(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
	}
}

// AuthConfig holds configuration for authentication middleware.
type AuthConfig struct {
	AuthFunc    func(c *core.Context) bool
	RedirectURL string
	Message     string
}

// AuthWithConfig returns an Auth middleware with custom config.
func AuthWithConfig(config AuthConfig) core.Middleware {
	if config.Message == "" {
		config.Message = "Unauthorized"
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			if !config.AuthFunc(c) {
				if config.RedirectURL != "" {
					return c.Redirect(http.StatusFound, config.RedirectURL)
				}
				return c.String(http.StatusUnauthorized, config.Message)
			}
			return next(c)
		}
	}
}

// BasicAuth returns a HTTP Basic Auth middleware.
func BasicAuth(validator func(username, password string) bool) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			username, password, ok := c.Request.BasicAuth()
			if !ok || !validator(username, password) {
				c.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
				return c.String(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
	}
}

// BearerAuth returns a Bearer token authentication middleware.
func BearerAuth(validator func(token string) bool) core.Middleware {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c *core.Context) error {
			auth := c.Header("Authorization")
			if len(auth) < 7 || auth[:7] != "Bearer " {
				return c.String(http.StatusUnauthorized, "Unauthorized")
			}

			token := auth[7:]
			if !validator(token) {
				return c.String(http.StatusUnauthorized, "Invalid token")
			}

			return next(c)
		}
	}
}
