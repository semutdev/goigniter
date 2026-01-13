package helpers

import (
	"html/template"
	"strings"
)

// AllTemplateFuncs returns all helper functions for templates.
func AllTemplateFuncs() template.FuncMap {
	funcs := template.FuncMap{
		// URL helpers
		"base_url":  BaseURL,
		"site_url":  SiteURL,
		"asset_url": AssetURL,

		// String helpers
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"trim":     strings.TrimSpace,
		"contains": strings.Contains,
		"replace":  strings.ReplaceAll,
		"split":    strings.Split,
		"join":     strings.Join,

		// Conditional helpers
		"default": func(def, val any) any {
			if val == nil || val == "" || val == 0 {
				return def
			}
			return val
		},
		"eq": func(a, b any) bool {
			return a == b
		},
		"ne": func(a, b any) bool {
			return a != b
		},
	}

	return funcs
}
