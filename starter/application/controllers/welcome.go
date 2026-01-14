package controllers

import "github.com/semutdev/goigniter/system/core"

// WelcomeController - Controller default
type WelcomeController struct {
	core.Controller
}

// Index - GET /welcome
func (w *WelcomeController) Index() {
	w.Ctx.View("welcome", core.Map{
		"Title": "Welcome to GoIgniter!",
	})
}
