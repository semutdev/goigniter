package helpers

import (
	"html/template"
	"os"
	"strings"
)

var baseURL string

// Init sets the base URL.
func Init(url string) {
	baseURL = strings.TrimRight(url, "/")
}

// BaseURL returns the base URL with optional path.
// Example: BaseURL("/css/style.css") → "http://localhost:8080/css/style.css"
func BaseURL(path ...string) string {
	if baseURL == "" {
		baseURL = os.Getenv("APP_URL")
		if baseURL == "" {
			port := os.Getenv("APP_PORT")
			if port == "" {
				port = ":8080"
			}
			baseURL = "http://localhost" + port
		}
		baseURL = strings.TrimRight(baseURL, "/")
	}

	if len(path) > 0 && path[0] != "" {
		p := path[0]
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return baseURL + p
	}
	return baseURL
}

// SiteURL is an alias for BaseURL (CI3 compatibility).
func SiteURL(path ...string) string {
	return BaseURL(path...)
}

// AssetURL returns URL for static assets.
func AssetURL(path string) string {
	return BaseURL("/public/" + strings.TrimPrefix(path, "/"))
}

// TemplateFuncs returns template.FuncMap with URL helpers.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"base_url":  BaseURL,
		"site_url":  SiteURL,
		"asset_url": AssetURL,
	}
}
