package admin

import (
	"goigniter/libs"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Dashboard struct{}

func init() {
	libs.Register("admin/dashboard", &Dashboard{})
}

func (d *Dashboard) Index(c echo.Context) error {
	// protect dashboard harus login admin
	if !libs.RequireGroup(c, "admin") {
		return nil
	}

	data := map[string]interface{}{
		"Title": "Dashboard admin",
	}

	return c.Render(http.StatusOK, "admin/dashboard", data)
}
