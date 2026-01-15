package controllers

import "github.com/semutdev/goigniter/system/core"

// Welcome - Controller default
type Welcome struct {
	core.Controller
}

// Index - GET /welcome
func (w *Welcome) Index() {
	w.Ctx.View("welcome", core.Map{
		"Title": "Welcome to GoIgniter!",
	})
}
