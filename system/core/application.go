package core

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	return http.ListenAndServe(addr, app)
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
	dir       string
	ext       string
	funcMap   template.FuncMap
	templates map[string]*template.Template
	reload    bool
	mu        sync.RWMutex
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

	e.templates = make(map[string]*template.Template)

	return filepath.Walk(e.dir, func(path string, info os.FileInfo, err error) error {
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

		name := strings.TrimSuffix(rel, e.ext)
		name = strings.ReplaceAll(name, string(os.PathSeparator), "/")

		t, err := e.parseTemplate(path)
		if err != nil {
			return err
		}

		e.templates[name] = t
		return nil
	})
}

func (e *TemplateEngine) parseTemplate(path string) (*template.Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t := template.New(filepath.Base(path)).Funcs(e.funcMap)
	return t.Parse(string(content))
}

// Render renders a template with the given data.
func (e *TemplateEngine) Render(w io.Writer, name string, data any) error {
	if e.reload {
		if err := e.loadTemplates(); err != nil {
			return err
		}
	}

	e.mu.RLock()
	t, ok := e.templates[name]
	e.mu.RUnlock()

	if !ok {
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
