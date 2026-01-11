package admin

import (
	"goigniter/libs"

	"github.com/labstack/echo/v4"
)

func init() {
	libs.Register("admin/users", &Users{})
}

type Users struct{}

func (d *Users) Index(c echo.Context) error {
	return c.String(200, "Admin Users")
}
