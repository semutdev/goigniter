package controllers

import (
	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&WelcomeController{})
}

type WelcomeController struct {
	core.Controller
}

func (w *WelcomeController) Index() {
	w.Ctx.View("welcome", core.Map{
		"Title": "Welcome to GoIgniter!",
	})
}
