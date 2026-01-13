package core

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// Context represents the context of an HTTP request.
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter

	params     map[string]string
	query      url.Values
	store      map[string]any
	controller ControllerInterface
	written    bool
	app        *Application
}

var contextPool = sync.Pool{
	New: func() any {
		return &Context{
			params: make(map[string]string),
			store:  make(map[string]any),
		}
	},
}

func acquireContext(w http.ResponseWriter, r *http.Request, app *Application) *Context {
	ctx := contextPool.Get().(*Context)
	ctx.Request = r
	ctx.Response = w
	ctx.query = nil
	ctx.written = false
	ctx.app = app
	return ctx
}

func releaseContext(ctx *Context) {
	ctx.Request = nil
	ctx.Response = nil
	ctx.controller = nil
	ctx.app = nil
	for k := range ctx.params {
		delete(ctx.params, k)
	}
	for k := range ctx.store {
		delete(ctx.store, k)
	}
	contextPool.Put(ctx)
}

// --- Response Helpers ---

func (c *Context) JSON(code int, data any) error {
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(code)
	c.written = true
	return json.NewEncoder(c.Response).Encode(data)
}

func (c *Context) HTML(code int, html string) error {
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(code)
	c.written = true
	_, err := c.Response.Write([]byte(html))
	return err
}

func (c *Context) String(code int, s string) error {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(code)
	c.written = true
	_, err := c.Response.Write([]byte(s))
	return err
}

func (c *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		code = http.StatusFound
	}
	http.Redirect(c.Response, c.Request, url, code)
	c.written = true
	return nil
}

func (c *Context) File(filepath string) error {
	http.ServeFile(c.Response, c.Request, filepath)
	c.written = true
	return nil
}

func (c *Context) NoContent(code int) error {
	c.Response.WriteHeader(code)
	c.written = true
	return nil
}

func (c *Context) Blob(code int, contentType string, data []byte) error {
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.WriteHeader(code)
	c.written = true
	_, err := c.Response.Write(data)
	return err
}

// --- Input Helpers ---

func (c *Context) Param(name string) string {
	return c.params[name]
}

func (c *Context) ParamInt(name string) (int, error) {
	return strconv.Atoi(c.params[name])
}

func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *Context) QueryDefault(name, def string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	if v := c.query.Get(name); v != "" {
		return v
	}
	return def
}

func (c *Context) QueryInt(name string) (int, error) {
	return strconv.Atoi(c.Query(name))
}

func (c *Context) QueryIntDefault(name string, def int) int {
	if v, err := c.QueryInt(name); err == nil {
		return v
	}
	return def
}

func (c *Context) Form(name string) string {
	return c.Request.FormValue(name)
}

// FormValue is an alias for Form (for compatibility).
func (c *Context) FormValue(name string) string {
	return c.Request.FormValue(name)
}

func (c *Context) Bind(dest any) error {
	contentType := c.Request.Header.Get("Content-Type")
	switch {
	case contentType == "application/json" || contentType == "":
		return json.NewDecoder(c.Request.Body).Decode(dest)
	default:
		if err := c.Request.ParseForm(); err != nil {
			return err
		}
		return json.NewDecoder(c.Request.Body).Decode(dest)
	}
}

func (c *Context) Body() ([]byte, error) {
	return io.ReadAll(c.Request.Body)
}

// --- Request Info ---

func (c *Context) Method() string {
	return c.Request.Method
}

func (c *Context) Path() string {
	return c.Request.URL.Path
}

func (c *Context) IP() string {
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return c.Request.RemoteAddr
}

func (c *Context) Header(name string) string {
	return c.Request.Header.Get(name)
}

func (c *Context) SetHeader(name, value string) {
	c.Response.Header().Set(name, value)
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// --- Request-Scoped Storage ---

func (c *Context) Set(key string, value any) {
	c.store[key] = value
}

func (c *Context) Get(key string) any {
	return c.store[key]
}

func (c *Context) GetString(key string) string {
	if v, ok := c.store[key].(string); ok {
		return v
	}
	return ""
}

func (c *Context) GetInt(key string) int {
	if v, ok := c.store[key].(int); ok {
		return v
	}
	return 0
}

// --- View Rendering ---

func (c *Context) View(name string, data Map) error {
	if c.app == nil || c.app.renderer == nil {
		return c.HTML(200, "Template engine not configured")
	}
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(200)
	c.written = true
	return c.app.renderer.Render(c.Response, name, data)
}

func (c *Context) ViewWithCode(code int, name string, data Map) error {
	if c.app == nil || c.app.renderer == nil {
		return c.HTML(code, "Template engine not configured")
	}
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.WriteHeader(code)
	c.written = true
	return c.app.renderer.Render(c.Response, name, data)
}
