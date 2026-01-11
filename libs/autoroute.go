package libs

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// AutoRoute mendaftarkan handler untuk semua controller yang terdaftar
func AutoRoute(e *echo.Echo, registry map[string]interface{}) {
	handler := func(c echo.Context) error {
		path := c.Param("*")

		// Hapus leading slash jika ada
		path = strings.TrimPrefix(path, "/")

		if path == "" {
			return echo.ErrNotFound
		}

		controller, method, id := parsePath(path, registry)

		// Default method = Index
		if method == "" {
			method = "Index"
		}

		ctrl, exists := registry[controller]
		if !exists {
			return echo.ErrNotFound
		}

		return callMethod(ctrl, method, c, id)
	}

	// Wildcard route untuk menangkap semua path
	e.Any("/*", handler)
}

// parsePath memparse path menjadi controller, method, dan id
// Mencoba match dari path terpanjang ke terpendek
func parsePath(path string, registry map[string]interface{}) (controller, method, id string) {
	parts := strings.Split(path, "/")

	// Coba dari path terpanjang
	// Contoh: "admin/users/edit/123"
	// Cek: "admin/users/edit/123" → "admin/users/edit" → "admin/users" → "admin"

	for i := len(parts); i > 0; i-- {
		tryController := strings.Join(parts[:i], "/")

		if _, exists := registry[tryController]; exists {
			controller = tryController
			remaining := parts[i:]

			if len(remaining) > 0 {
				method = remaining[0]
			}
			if len(remaining) > 1 {
				id = remaining[1]
			}

			return
		}
	}

	// Fallback: asumsi format lama controller/method/id
	if len(parts) >= 1 {
		controller = parts[0]
	}
	if len(parts) >= 2 {
		method = parts[1]
	}
	if len(parts) >= 3 {
		id = parts[2]
	}

	return
}

// callMethod memanggil method pada controller dengan reflection
func callMethod(ctrl interface{}, methodName string, c echo.Context, id string) error {
	refVal := reflect.ValueOf(ctrl)

	// Title case method name
	methodName = strings.Title(strings.ToLower(methodName))

	method := refVal.MethodByName(methodName)
	if !method.IsValid() {
		return echo.NewHTTPError(http.StatusNotFound, "Method "+methodName+" not found")
	}

	// Cek HTTP method restriction
	if restrictor, ok := ctrl.(MethodRestrictor); ok {
		allowed := restrictor.AllowedMethods()[methodName]
		if len(allowed) > 0 {
			httpMethod := c.Request().Method
			if !contains(allowed, httpMethod) {
				return echo.ErrMethodNotAllowed
			}
		}
	}

	methodType := method.Type()
	var args []reflect.Value

	// Param 1: echo.Context
	args = append(args, reflect.ValueOf(c))

	// Param 2: inject id jika method punya parameter kedua
	if methodType.NumIn() == 2 && id != "" {
		paramType := methodType.In(1)

		switch paramType.Kind() {
		case reflect.String:
			args = append(args, reflect.ValueOf(id))
		case reflect.Int:
			idInt, _ := strconv.Atoi(id)
			args = append(args, reflect.ValueOf(idInt))
		case reflect.Int64:
			idInt, _ := strconv.ParseInt(id, 10, 64)
			args = append(args, reflect.ValueOf(idInt))
		case reflect.Uint:
			idUint, _ := strconv.ParseUint(id, 10, 64)
			args = append(args, reflect.ValueOf(uint(idUint)))
		}
	}

	// Set id ke param context agar bisa diakses via c.Param("id")
	if id != "" {
		c.SetParamNames("id")
		c.SetParamValues(id)
	}

	results := method.Call(args)

	if len(results) > 0 && !results[0].IsNil() {
		return results[0].Interface().(error)
	}

	return nil
}

// contains cek apakah slice berisi string tertentu
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
