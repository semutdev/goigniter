package admin

import (
	"full-crud/application/libs"

	"github.com/semutdev/goigniter/system/core"
)

func init() {
	core.Register(&Dashboard{}, "admin")
}

type Dashboard struct {
	core.Controller
}

func (this *Dashboard) Index() {
	if !libs.RequireGroup(this.Ctx, "admin") {
		return
	}

	data := core.Map{
		"Title": "Dashboard Admin",
	}

	this.Ctx.View("admin/inc/header", data)
	this.Ctx.View("admin/dashboard", data)
	this.Ctx.View("admin/inc/footer", data)
}
