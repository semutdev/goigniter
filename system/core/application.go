package core

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Application is the main framework instance.
type Application struct {
	router      *Router
	middlewares []Middleware
	groups      []*Group
	renderer    *TemplateEngine
}

// Group represents a route group with shared prefix and middleware.
type Group struct {
	prefix      string
	middlewares []Middleware
	app         *Application
}

// Use adds global middleware to the application.
func (app *Application) Use(middlewares ...Middleware) {
	app.middlewares = append(app.middlewares, middlewares...)
}

// GET registers a GET route.
func (app *Application) GET(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodGet, pattern, handler)
}

// POST registers a POST route.
func (app *Application) POST(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodPost, pattern, handler)
}

// PUT registers a PUT route.
func (app *Application) PUT(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodPut, pattern, handler)
}

// DELETE registers a DELETE route.
func (app *Application) DELETE(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodDelete, pattern, handler)
}

// PATCH registers a PATCH route.
func (app *Application) PATCH(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodPatch, pattern, handler)
}

// OPTIONS registers an OPTIONS route.
func (app *Application) OPTIONS(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodOptions, pattern, handler)
}

// HEAD registers a HEAD route.
func (app *Application) HEAD(pattern string, handler HandlerFunc) {
	app.router.Add(http.MethodHead, pattern, handler)
}

// Group creates a new route group with the given prefix and middleware.
func (app *Application) Group(prefix string, middlewares ...Middleware) *Group {
	g := &Group{
		prefix:      prefix,
		middlewares: middlewares,
		app:         app,
	}
	app.groups = append(app.groups, g)
	return g
}

// GET registers a GET route in the group.
func (g *Group) GET(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodGet, fullPath, wrapped)
}

// POST registers a POST route in the group.
func (g *Group) POST(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodPost, fullPath, wrapped)
}

// PUT registers a PUT route in the group.
func (g *Group) PUT(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodPut, fullPath, wrapped)
}

// DELETE registers a DELETE route in the group.
func (g *Group) DELETE(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodDelete, fullPath, wrapped)
}

// PATCH registers a PATCH route in the group.
func (g *Group) PATCH(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodPatch, fullPath, wrapped)
}

// OPTIONS registers an OPTIONS route in the group.
func (g *Group) OPTIONS(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodOptions, fullPath, wrapped)
}

// HEAD registers a HEAD route in the group.
func (g *Group) HEAD(pattern string, handler HandlerFunc) {
	fullPath := path.Join(g.prefix, pattern)
	wrapped := g.wrapHandler(handler)
	g.app.router.Add(http.MethodHead, fullPath, wrapped)
}

// Group creates a nested group.
func (g *Group) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		prefix:      path.Join(g.prefix, prefix),
		middlewares: append(g.middlewares, middlewares...),
		app:         g.app,
	}
}

// wrapHandler wraps handler with group middleware.
func (g *Group) wrapHandler(handler HandlerFunc) HandlerFunc {
	return applyMiddleware(handler, g.middlewares...)
}

// Static serves static files from the given directory.
func (app *Application) Static(prefix, root string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	pattern := prefix + "*filepath"

	fs := http.FileServer(http.Dir(root))
	handler := func(c *Context) error {
		c.Request.URL.Path = c.Param("filepath")
		fs.ServeHTTP(c.Response, c.Request)
		return nil
	}

	app.router.Add(http.MethodGet, pattern, handler)
}

// Register registers a controller with the global registry.
func (app *Application) Register(controller ControllerInterface, prefix ...string) {
	globalRegistry.Register(controller, prefix...)
}

// AutoRoute registers routes for all controllers in the registry.
func (app *Application) AutoRoute() {
	globalRegistry.AutoRoute(app)
}

// ServeHTTP implements http.Handler interface.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := acquireContext(w, r, app)
	defer releaseContext(ctx)

	handler, params, found := app.router.Find(r.Method, r.URL.Path)
	if !found {
		http.NotFound(w, r)
		return
	}

	ctx.params = params

	finalHandler := applyMiddleware(handler, app.middlewares...)
	if err := finalHandler(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Run starts the HTTP server on the given address.
func (app *Application) Run(addr string) error {
	// Print banner automatically
	printBanner(addr)
	return http.ListenAndServe(addr, app)
}

// printBanner prints the startup banner
func printBanner(addr string) {
	// Color codes
	cyan := "\033[36m"
	green := "\033[32m"
	yellow := "\033[33m"
	blue := "\033[34m"
	bold := "\033[1m"
	reset := "\033[0m"

	// ASCII Art Logo
	logo := `
                   ++
                   + +
                  +   +
                ++    +
              ++     +++++
             ++     +++* ++
            ++            ++
           ++         +    +
           ++        +++   +
            ++     ++  +  ++
             ++       ++++
                +++ *+*+
                   +
`

	fmt.Printf("%s%s", cyan, logo)
	fmt.Printf("%s%sGoIgniter%s v%s%s\n", bold, green, cyan, Version, reset)
	fmt.Printf("%s⚡ High Performance Web Framework for Go%s\n", yellow, reset)
	fmt.Printf("%s%s%s\n", blue, Website, reset)
	fmt.Println()

	// Print server info
	fmt.Printf("  Go Version: %s\n", runtime.Version())
	fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  Server:     http://localhost%s\n", addr)
	fmt.Println()
}

// SetRenderer sets the template renderer for the application.
func (app *Application) SetRenderer(r *TemplateEngine) {
	app.renderer = r
}

// LoadTemplates loads templates from the given directory.
func (app *Application) LoadTemplates(dir string, reload bool) error {
	r, err := NewTemplateEngine(TemplateConfig{
		Dir:    dir,
		Ext:    ".html",
		Reload: reload,
	})
	if err != nil {
		return err
	}
	app.renderer = r
	return nil
}

// LoadTemplatesWithFuncs loads templates with custom functions.
func (app *Application) LoadTemplatesWithFuncs(dir string, reload bool, funcs template.FuncMap) error {
	r, err := NewTemplateEngine(TemplateConfig{
		Dir:     dir,
		Ext:     ".html",
		Reload:  reload,
		FuncMap: funcs,
	})
	if err != nil {
		return err
	}
	app.renderer = r
	return nil
}

// Renderer returns the template renderer.
func (app *Application) Renderer() *TemplateEngine {
	return app.renderer
}

// applyMiddleware wraps a handler with the given middleware chain.
func applyMiddleware(handler HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// --- Template Engine (embedded in core) ---

// TemplateEngine provides template rendering.
type TemplateEngine struct {
	dir            string
	ext            string
	funcMap        template.FuncMap
	globalTemplate *template.Template
	templates      map[string]*template.Template
	reload         bool
	mu             sync.RWMutex
}

// TemplateConfig holds configuration for TemplateEngine.
type TemplateConfig struct {
	Dir     string
	Ext     string
	FuncMap template.FuncMap
	Reload  bool
}

// NewTemplateEngine creates a new TemplateEngine.
func NewTemplateEngine(config TemplateConfig) (*TemplateEngine, error) {
	if config.Ext == "" {
		config.Ext = ".html"
	}
	if config.FuncMap == nil {
		config.FuncMap = DefaultTemplateFuncs()
	}

	e := &TemplateEngine{
		dir:       config.Dir,
		ext:       config.Ext,
		funcMap:   config.FuncMap,
		templates: make(map[string]*template.Template),
		reload:    config.Reload,
	}

	if err := e.loadTemplates(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *TemplateEngine) loadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	root := template.New("").Funcs(e.funcMap)
	templatesMap := make(map[string]*template.Template)

	type tplFile struct {
		path           string
		relSlash       string
		nameWithoutExt string
		content        string
		hasDefine      bool
	}
	var files []tplFile

	err := filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, e.ext) {
			return nil
		}

		rel, err := filepath.Rel(e.dir, path)
		if err != nil {
			return err
		}

		relSlash := filepath.ToSlash(rel)
		nameWithoutExt := strings.TrimSuffix(relSlash, e.ext)

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(raw)
		hasDefine := strings.Contains(content, "{{define")

		files = append(files, tplFile{
			path:           path,
			relSlash:       relSlash,
			nameWithoutExt: nameWithoutExt,
			content:        content,
			hasDefine:      hasDefine,
		})
		return nil
	})
	if err != nil {
		return err
	}

	// Pass 1: Parse all templates into root
	for _, f := range files {
		parsedText := f.content
		if !f.hasDefine {
			parsedText = `{{define "` + f.nameWithoutExt + `"}}` + f.content + `{{end}}`
		}
		if _, err := root.Parse(parsedText); err != nil {
			return fmt.Errorf("error parsing template %s: %w", f.relSlash, err)
		}
	}

	// Pass 2: Register aliases so templates can be referenced by
	// full path with extension ("admin/dashboard/index.html"),
	// with leading slash ("/admin/dashboard/index", "/admin/dashboard/index.html"),
	// and base name ("_summary_cards.html", "_summary_cards").
	for _, f := range files {
		targetName := f.nameWithoutExt
		if root.Lookup(targetName) == nil {
			if root.Lookup(f.relSlash) != nil {
				targetName = f.relSlash
			}
		}

		if root.Lookup(targetName) != nil {
			baseName := filepath.Base(f.relSlash)
			baseWithoutExt := strings.TrimSuffix(baseName, e.ext)

			aliases := []string{
				f.nameWithoutExt,
				f.relSlash,
				"/" + f.nameWithoutExt,
				"/" + f.relSlash,
				baseName,
				baseWithoutExt,
			}

			for _, alias := range aliases {
				if alias == targetName {
					templatesMap[alias] = root.Lookup(targetName)
					continue
				}
				if root.Lookup(alias) == nil {
					aliasTmpl := fmt.Sprintf(`{{define %q}}{{template %q .}}{{end}}`, alias, targetName)
					if _, err := root.Parse(aliasTmpl); err != nil {
						return fmt.Errorf("error creating alias %s for %s: %w", alias, targetName, err)
					}
				}
				if t := root.Lookup(alias); t != nil {
					templatesMap[alias] = t
				}
			}
		}
	}

	// Also index all defined templates from root
	for _, t := range root.Templates() {
		tName := t.Name()
		if tName != "" {
			templatesMap[tName] = t
		}
	}

	e.globalTemplate = root
	e.templates = templatesMap
	return nil
}

// Render renders a template with the given data.
func (e *TemplateEngine) Render(w io.Writer, name string, data any) error {
	if e.reload {
		if err := e.loadTemplates(); err != nil {
			return err
		}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var t *template.Template
	if e.templates != nil {
		t = e.templates[name]
	}
	if t == nil && e.globalTemplate != nil {
		t = e.globalTemplate.Lookup(name)
		if t == nil {
			normName := strings.TrimPrefix(name, "/")
			t = e.globalTemplate.Lookup(normName)
			if t == nil {
				t = e.globalTemplate.Lookup(strings.TrimSuffix(normName, e.ext))
			}
			if t == nil {
				t = e.globalTemplate.Lookup(normName + e.ext)
			}
		}
	}

	if t == nil {
		return &TemplateNotFoundError{Name: name}
	}

	return t.Execute(w, data)
}

// TemplateNotFoundError is returned when a template is not found.
type TemplateNotFoundError struct {
	Name string
}

func (e *TemplateNotFoundError) Error() string {
	return "template not found: " + e.Name
}

// DefaultTemplateFuncs returns the default template function map.
func DefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"base_url": func(path ...string) string {
			p := ""
			if len(path) > 0 {
				p = path[0]
			}
			if p != "" && !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			return p
		},
		"site_url": func(path ...string) string {
			p := ""
			if len(path) > 0 {
				p = path[0]
			}
			if p != "" && !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			return p
		},
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
	}
}
