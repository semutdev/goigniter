package controllers

import (
	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Welcome{})
}

type Welcome struct {
	core.Controller
}

func (w *Welcome) Index() {
	w.Ctx.View("welcome", core.Map{
		"Title": "Welcome to GoIgniter!",
	})
}
