package libs

import (
	"html/template"
	"os"
	"strings"

	"github.com/semutdev/goigniter/system/libraries/session"
	"github.com/semutdev/goigniter/system/core"
)

// BaseURL returns the base URL from APP_URL env
func BaseURL(path ...string) string {
	base := os.Getenv("APP_URL")
	if base == "" {
		base = "http://localhost" + os.Getenv("APP_PORT")
	}

	base = strings.TrimRight(base, "/")

	if len(path) > 0 && path[0] != "" {
		p := path[0]
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return base + p
	}

	return base
}

// SiteURL returns the site URL (alias for BaseURL)
func SiteURL(path ...string) string {
	return BaseURL(path...)
}

// SetFlash sets a flash message
func SetFlash(c *core.Context, key string, value string) {
	session.SetFlash(c, key, value)
}

// GetFlash gets and removes a flash message
func GetFlash(c *core.Context, key string) string {
	return session.GetFlash(c, key)
}

// TemplateFuncs returns FuncMap untuk template
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"base_url": BaseURL,
		"site_url": SiteURL,
	}
}
