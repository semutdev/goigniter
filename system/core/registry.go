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

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodName := method.Name

		if isInternalMethod(methodName) {
			continue
		}

		httpMethod, routePath := resolveRoute(basePath, methodName)
		if httpMethod == "" {
			continue
		}

		handler := createControllerHandler(factory, methodName, controllerMiddleware, methodMiddleware[methodName])
		app.router.Add(httpMethod, routePath, handler)
	}
}

func resolveRoute(basePath, methodName string) (httpMethod, routePath string) {
	switch methodName {
	case "Index":
		return "GET", "/" + basePath
	case "Show":
		return "GET", "/" + basePath + "/:id"
	case "Create":
		return "GET", "/" + basePath + "/create"
	case "Store":
		return "POST", "/" + basePath
	case "Edit":
		return "GET", "/" + basePath + "/:id/edit"
	case "Update":
		return "PUT", "/" + basePath + "/:id"
	case "Delete":
		return "DELETE", "/" + basePath + "/:id"
	default:
		return "GET", "/" + basePath + "/" + strings.ToLower(methodName)
	}
}

func isInternalMethod(name string) bool {
	internal := map[string]bool{
		"SetContext":    true,
		"Middleware":    true,
		"MiddlewareFor": true,
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
