package controllers

import (
	"goigniter/libs"
	"net/http"

	"github.com/labstack/echo/v4"
)

func init() {
	libs.Register("welcome", &Welcome{})
}

type Welcome struct{}

func (w *Welcome) Index(c echo.Context) error {
	data := map[string]interface{}{
		"Title": "Welcome to Goigniter!",
	}

	return c.Render(http.StatusOK, "welcome", data)
}
