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

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name

		if isInternalMethod(methodName) {
			continue
		}

		routePath := resolveRoutePath(basePath, methodName)
		httpMethods := resolveHTTPMethods(methodName, allowedMethods)

		handler := createControllerHandler(factory, methodName, controllerMiddleware, methodMiddleware[methodName])

		// Register route for each HTTP method
		for _, httpMethod := range httpMethods {
			app.router.Add(httpMethod, routePath, handler)
		}
	}
}

// resolveRoutePath returns the route path for a method (without HTTP method)
func resolveRoutePath(basePath, methodName string) string {
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

func resolveRoute(basePath, methodName string) (httpMethod, routePath string) {
	return "GET", "/" + basePath + "/" + strings.ToLower(methodName)
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
