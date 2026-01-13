// Package core provides the main components of GoIgniter framework.
// Built on net/http stdlib with zero external dependencies for HTTP handling.
package core

// Map is a shorthand for map[string]any, used for view data and JSON responses.
type Map map[string]any

// HandlerFunc defines the handler function signature.
type HandlerFunc func(c *Context) error

// Middleware defines the middleware function signature.
// Middleware wraps a handler and returns a new handler.
type Middleware func(next HandlerFunc) HandlerFunc

// ControllerInterface defines the interface that all controllers must implement.
type ControllerInterface interface {
	// SetContext sets the context for the controller.
	SetContext(ctx *Context)

	// Middleware returns controller-level middleware.
	Middleware() []Middleware

	// MiddlewareFor returns method-specific middleware.
	MiddlewareFor() map[string][]Middleware
}

// ControllerFactory is a function that creates a new controller instance.
type ControllerFactory func() ControllerInterface

// New creates a new Application instance.
func New() *Application {
	app := &Application{
		router:      newRouter(),
		middlewares: make([]Middleware, 0),
	}
	return app
}

// Register registers a controller with the global registry.
// Optional prefix can be provided for nested controllers (e.g., "admin").
func Register(controller ControllerInterface, prefix ...string) {
	globalRegistry.Register(controller, prefix...)
}
