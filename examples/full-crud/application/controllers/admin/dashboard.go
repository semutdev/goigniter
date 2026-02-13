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

func (d *Dashboard) Index() {
	if !libs.RequireGroup(d.Ctx, "admin") {
		return
	}

	d.Ctx.View("admin/dashboard", core.Map{
		"Title": "Dashboard Admin",
	})
}
