package admin

import (
	"goigniter/libs"

	"github.com/labstack/echo/v4"
)

type Dashboard struct{}

func init() {
	libs.Register("admin/dashboard", &Dashboard{})
}

func (d *Dashboard) Index(c echo.Context) error {
	// protectt dashboard harus login admin
	if !libs.RequireGroup(c, "admin") {
		return nil
	}
	return c.String(200, "Admin Dashboard")
}
