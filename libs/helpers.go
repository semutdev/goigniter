package libs

import (
	"os"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("APP_KEY")))

// BaseURL returns the base URL from APP_URL env
func BaseURL(path ...string) string {
	base := os.Getenv("APP_URL")
	if base == "" {
		base = "http://localhost" + os.Getenv("APP_PORT")
	}

	// Pastikan tidak ada trailing slash
	base = strings.TrimRight(base, "/")

	if len(path) > 0 && path[0] != "" {
		// Pastikan path dimulai dengan /
		p := path[0]
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return base + p
	}

	return base
}

// SiteURL returns the site URL (base + path)
// Alias untuk BaseURL, di CI3 bedanya site_url include index.php
func SiteURL(path ...string) string {
	return BaseURL(path...)
}

// SetFlash sets a flash message in session
func SetFlash(c echo.Context, key string, value string) {
	session, _ := store.Get(c.Request(), "flash")
	session.AddFlash(value, key)
	session.Save(c.Request(), c.Response())
}

// GetFlash gets and removes a flash message from session
func GetFlash(c echo.Context, key string) string {
	session, _ := store.Get(c.Request(), "flash")
	flashes := session.Flashes(key)
	session.Save(c.Request(), c.Response())

	if len(flashes) > 0 {
		return flashes[0].(string)
	}
	return ""
}

// TemplateFuncs returns FuncMap untuk template
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"base_url": BaseURL,
		"site_url": SiteURL,
	}
}
