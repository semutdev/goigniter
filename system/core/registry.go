package core

import (
	"reflect"
	"strings"
	"unicode"
)

type controllerRegistry struct {
	controllers map[string]ControllerFactory
}

var globalRegistry = &controllerRegistry{
	controllers: make(map[string]ControllerFactory),
}

// ResetRegistry clears the global controller registry (useful for testing).
func ResetRegistry() {
	globalRegistry.controllers = make(map[string]ControllerFactory)
}

func (r *controllerRegistry) Register(controller ControllerInterface, prefix ...string) {
	t := reflect.TypeOf(controller)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := strings.ToLower(t.Name())
	if trimmed := strings.TrimSuffix(name, "controller"); trimmed != "" {
		name = trimmed
	}

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

		routePaths := resolveRoutePaths(basePath, methodName, customRoutes)
		httpMethods := resolveHTTPMethods(methodName, allowedMethods)

		handler := createControllerHandler(factory, methodName, controllerMiddleware, methodMiddleware[methodName])

		// Register route for each HTTP method and resolved path
		for _, routePath := range routePaths {
			for _, httpMethod := range httpMethods {
				app.router.Add(httpMethod, routePath, handler)
			}
		}
	}
}

// resolveRoutePaths returns route paths for a controller method.
// For Index: returns both /{basePath} and /{basePath}/index.
// For PascalCase methods: returns both /{basePath}/snake_case and /{basePath}/lowercase.
func resolveRoutePaths(basePath, methodName string, customRoutes map[string]string) []string {
	if customRoutes != nil {
		if route, ok := customRoutes[methodName]; ok {
			if len(route) > 0 && route[0] == '/' {
				return []string{route}
			}
			return []string{"/" + basePath + "/" + route}
		}
	}

	lower := strings.ToLower(methodName)
	if lower == "index" {
		return []string{
			"/" + basePath,
			"/" + basePath + "/index",
		}
	}

	paths := []string{"/" + basePath + "/" + lower}
	snake := toSnakeCase(methodName)
	if snake != lower {
		paths = append([]string{"/" + basePath + "/" + snake}, paths...)
	}

	return paths
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				sb.WriteByte('_')
			}
			sb.WriteRune(unicode.ToLower(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// resolveRoutePath returns the primary route path for a method (backwards compatibility).
func resolveRoutePath(basePath, methodName string, customRoutes map[string]string) string {
	paths := resolveRoutePaths(basePath, methodName, customRoutes)
	if len(paths) > 0 {
		return paths[0]
	}
	return "/" + basePath + "/" + strings.ToLower(methodName)
}

// resolveHTTPMethods returns allowed HTTP methods for a controller method
func resolveHTTPMethods(methodName string, allowedMethods map[string][]string) []string {
	if allowedMethods != nil {
		if methods, ok := allowedMethods[methodName]; ok {
			return methods
		}
	}

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
		"Routes":         true,
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

		mType := method.Type()

		handler := func(ctx *Context) error {
			var args []reflect.Value
			if mType.NumIn() == 1 && mType.In(0) == reflect.TypeOf(ctx) {
				args = []reflect.Value{reflect.ValueOf(ctx)}
			}
			results := method.Call(args)
			if len(results) > 0 {
				last := results[len(results)-1]
				if !last.IsNil() && last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
					return last.Interface().(error)
				}
			}
			return nil
		}

		allMiddleware := append(ctrlMw, methodMw...)
		finalHandler := applyMiddleware(handler, allMiddleware...)

		return finalHandler(c)
	}
}
