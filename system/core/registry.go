package core

import (
	"reflect"
	"strings"
)

type controllerRegistry struct {
	controllers map[string]ControllerFactory
}

var globalRegistry = &controllerRegistry{
	controllers: make(map[string]ControllerFactory),
}

func (r *controllerRegistry) Register(controller ControllerInterface, prefix ...string) {
	t := reflect.TypeOf(controller)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := strings.ToLower(t.Name())

	path := name
	if len(prefix) > 0 && prefix[0] != "" {
		path = prefix[0] + "/" + name
	}

	r.controllers[path] = func() ControllerInterface {
		newCtrl := reflect.New(t).Interface().(ControllerInterface)
		return newCtrl
	}
}

func (r *controllerRegistry) AutoRoute(app *Application) {
	for path, factory := range r.controllers {
		r.registerControllerRoutes(app, path, factory)
	}
}

func (r *controllerRegistry) registerControllerRoutes(app *Application, basePath string, factory ControllerFactory) {
	ctrl := factory()
	t := reflect.TypeOf(ctrl)

	controllerMiddleware := ctrl.Middleware()
	methodMiddleware := ctrl.MiddlewareFor()
	allowedMethods := ctrl.AllowedMethods()
	customRoutes := ctrl.Routes()

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name

		if isInternalMethod(methodName) {
			continue
		}

		routePath := resolveRoutePath(basePath, methodName, customRoutes)
		httpMethods := resolveHTTPMethods(methodName, allowedMethods)

		handler := createControllerHandler(factory, methodName, controllerMiddleware, methodMiddleware[methodName])

		// Register route for each HTTP method
		for _, httpMethod := range httpMethods {
			app.router.Add(httpMethod, routePath, handler)
		}
	}
}

// resolveRoutePath returns the route path for a method.
// If a custom route is defined, it uses that; otherwise uses default pattern.
func resolveRoutePath(basePath, methodName string, customRoutes map[string]string) string {
	// Check if custom route is defined
	if customRoutes != nil {
		if route, ok := customRoutes[methodName]; ok {
			// If route starts with /, it's absolute; otherwise relative to basePath
			if len(route) > 0 && route[0] == '/' {
				return route
			}
			return "/" + basePath + "/" + route
		}
	}

	// Default pattern: /{controller}/{method}
	return "/" + basePath + "/" + strings.ToLower(methodName)
}

// resolveHTTPMethods returns allowed HTTP methods for a controller method
func resolveHTTPMethods(methodName string, allowedMethods map[string][]string) []string {
	// Check if explicitly defined in controller
	if allowedMethods != nil {
		if methods, ok := allowedMethods[methodName]; ok {
			return methods
		}
	}

	// Default: all methods allow GET and POST
	return []string{"GET", "POST"}
}

func resolveRoute(basePath, methodName string, customRoutes map[string]string) (httpMethod, routePath string) {
	return "GET", resolveRoutePath(basePath, methodName, customRoutes)
}

func isInternalMethod(name string) bool {
	internal := map[string]bool{
		"SetContext":     true,
		"Middleware":     true,
		"MiddlewareFor":  true,
		"AllowedMethods": true,
	}
	return internal[name]
}

func createControllerHandler(factory ControllerFactory, methodName string, ctrlMw, methodMw []Middleware) HandlerFunc {
	return func(c *Context) error {
		ctrl := factory()
		ctrl.SetContext(c)

		method := reflect.ValueOf(ctrl).MethodByName(methodName)
		if !method.IsValid() {
			return nil
		}

		handler := func(ctx *Context) error {
			method.Call(nil)
			return nil
		}

		allMiddleware := append(ctrlMw, methodMw...)
		finalHandler := applyMiddleware(handler, allMiddleware...)

		return finalHandler(c)
	}
}
