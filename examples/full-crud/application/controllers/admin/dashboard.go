package admin

import (
	"full-crud/application/libs"

	"goigniter/system/core"
)

func init() {
	core.Register(&DashboardController{}, "admin")
}

type DashboardController struct {
	core.Controller
}

func (d *DashboardController) Index() {
	if !libs.RequireGroup(d.Ctx, "admin") {
		return
	}

	d.Ctx.View("admin/dashboard", core.Map{
		"Title": "Dashboard Admin",
	})
}
