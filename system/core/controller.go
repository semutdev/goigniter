package core

// Controller is the base controller that user controllers should embed.
type Controller struct {
	Ctx  *Context
	Load *Loader
	Data Map
}

func (c *Controller) SetContext(ctx *Context) {
	c.Ctx = ctx
	c.Load = &Loader{controller: c}
	c.Data = make(Map)
}

func (c *Controller) Middleware() []Middleware {
	return nil
}

func (c *Controller) MiddlewareFor() map[string][]Middleware {
	return nil
}

// Loader provides helper methods for loading resources.
type Loader struct {
	controller *Controller
	models     map[string]any
	libraries  map[string]any
}

func (l *Loader) Model(name string) any {
	if l.models == nil {
		l.models = make(map[string]any)
	}
	return l.models[name]
}

func (l *Loader) Library(name string) any {
	if l.libraries == nil {
		l.libraries = make(map[string]any)
	}
	return l.libraries[name]
}

func (l *Loader) View(name string, data Map) error {
	merged := make(Map)
	for k, v := range l.controller.Data {
		merged[k] = v
	}
	for k, v := range data {
		merged[k] = v
	}
	return l.controller.Ctx.View(name, merged)
}

func (l *Loader) Helper(name string) {
	// Helpers in Go are imported packages
}
