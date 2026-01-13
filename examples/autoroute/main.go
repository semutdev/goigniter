package main

import (
	"fmt"
	"log"

	"goigniter/system/core"
	"goigniter/system/middleware"
)

func main() {
	app := core.New()

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Register controllers
	core.Register(&ProductController{})
	core.Register(&UserController{})
	core.Register(&DashboardController{}, "admin") // With prefix: /admin/dashboardcontroller

	// Enable auto-routing based on registered controllers
	app.AutoRoute()

	// You can still add manual routes alongside auto-routes
	app.GET("/", func(c *core.Context) error {
		return c.JSON(200, core.Map{
			"message": "Welcome to GoIgniter AutoRoute Example!",
			"routes": core.Map{
				"products": core.Map{
					"GET /products":          "ProductController.Index()",
					"GET /products/:id":      "ProductController.Show()",
					"GET /products/create":   "ProductController.Create()",
					"POST /products":         "ProductController.Store()",
					"GET /products/:id/edit": "ProductController.Edit()",
					"PUT /products/:id":      "ProductController.Update()",
					"DELETE /products/:id":   "ProductController.Delete()",
					"GET /products/search":   "ProductController.Search() - custom method",
				},
				"users": core.Map{
					"GET /users":     "UserController.Index()",
					"GET /users/:id": "UserController.Show()",
				},
				"admin": core.Map{
					"GET /admin/dashboardcontroller": "DashboardController.Index() (with prefix)",
				},
			},
		})
	})

	// Start server
	port := ":8080"
	fmt.Println("=================================")
	fmt.Println("  GoIgniter - AutoRoute Example")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Server running at http://localhost" + port)
	fmt.Println()
	fmt.Println("Auto-generated routes from controllers:")
	fmt.Println()
	fmt.Println("ProductController:")
	fmt.Println("  GET    /productcontroller          → Index()")
	fmt.Println("  GET    /productcontroller/:id      → Show()")
	fmt.Println("  GET    /productcontroller/create   → Create()")
	fmt.Println("  POST   /productcontroller          → Store()")
	fmt.Println("  GET    /productcontroller/:id/edit → Edit()")
	fmt.Println("  PUT    /productcontroller/:id      → Update()")
	fmt.Println("  DELETE /productcontroller/:id      → Delete()")
	fmt.Println("  GET    /productcontroller/search   → Search()")
	fmt.Println()
	fmt.Println("UserController:")
	fmt.Println("  GET    /usercontroller             → Index()")
	fmt.Println("  GET    /usercontroller/:id         → Show()")
	fmt.Println()
	fmt.Println("DashboardController (admin prefix):")
	fmt.Println("  GET    /admin/dashboardcontroller  → Index()")
	fmt.Println()

	log.Fatal(app.Run(port))
}

// =============================================================================
// Product Controller - Full CRUD Example
// =============================================================================

type ProductController struct {
	core.Controller
}

// Index - GET /productcontroller
func (p *ProductController) Index() {
	products := []core.Map{
		{"id": 1, "name": "Laptop", "price": 15000000},
		{"id": 2, "name": "Mouse", "price": 250000},
		{"id": 3, "name": "Keyboard", "price": 750000},
	}
	p.Ctx.JSON(200, core.Map{
		"message":  "List all products",
		"products": products,
	})
}

// Show - GET /productcontroller/:id
func (p *ProductController) Show() {
	id := p.Ctx.Param("id")
	p.Ctx.JSON(200, core.Map{
		"message": "Show product detail",
		"product": core.Map{
			"id":    id,
			"name":  "Product " + id,
			"price": 500000,
		},
	})
}

// Create - GET /productcontroller/create (show form)
func (p *ProductController) Create() {
	p.Ctx.JSON(200, core.Map{
		"message": "Show create product form",
		"form": core.Map{
			"action": "POST /productcontroller",
			"fields": []string{"name", "price", "description"},
		},
	})
}

// Store - POST /productcontroller (save new product)
func (p *ProductController) Store() {
	var input struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}
	if err := p.Ctx.Bind(&input); err != nil {
		p.Ctx.JSON(400, core.Map{"error": err.Error()})
		return
	}
	p.Ctx.JSON(201, core.Map{
		"message": "Product created successfully",
		"product": input,
	})
}

// Edit - GET /productcontroller/:id/edit (show edit form)
func (p *ProductController) Edit() {
	id := p.Ctx.Param("id")
	p.Ctx.JSON(200, core.Map{
		"message": "Show edit form for product " + id,
		"product": core.Map{
			"id":    id,
			"name":  "Product " + id,
			"price": 500000,
		},
		"form": core.Map{
			"action": "PUT /productcontroller/" + id,
			"fields": []string{"name", "price", "description"},
		},
	})
}

// Update - PUT /productcontroller/:id
func (p *ProductController) Update() {
	id := p.Ctx.Param("id")
	var input struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}
	if err := p.Ctx.Bind(&input); err != nil {
		p.Ctx.JSON(400, core.Map{"error": err.Error()})
		return
	}
	p.Ctx.JSON(200, core.Map{
		"message": "Product " + id + " updated successfully",
		"product": input,
	})
}

// Delete - DELETE /productcontroller/:id
func (p *ProductController) Delete() {
	id := p.Ctx.Param("id")
	p.Ctx.JSON(200, core.Map{
		"message": "Product " + id + " deleted successfully",
	})
}

// Search - GET /productcontroller/search (custom method)
func (p *ProductController) Search() {
	q := p.Ctx.Query("q")
	p.Ctx.JSON(200, core.Map{
		"message": "Search products",
		"query":   q,
		"results": []core.Map{
			{"id": 1, "name": "Laptop (matched: " + q + ")"},
		},
	})
}

// =============================================================================
// User Controller - Simple Example
// =============================================================================

type UserController struct {
	core.Controller
}

// Index - GET /usercontroller
func (u *UserController) Index() {
	u.Ctx.JSON(200, core.Map{
		"message": "List all users",
		"users": []core.Map{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Doe", "email": "jane@example.com"},
		},
	})
}

// Show - GET /usercontroller/:id
func (u *UserController) Show() {
	id := u.Ctx.Param("id")
	u.Ctx.JSON(200, core.Map{
		"message": "User detail",
		"user": core.Map{
			"id":    id,
			"name":  "User " + id,
			"email": "user" + id + "@example.com",
		},
	})
}

// =============================================================================
// Dashboard Controller - With Prefix Example
// =============================================================================

type DashboardController struct {
	core.Controller
}

// Index - GET /admin/dashboardcontroller
func (d *DashboardController) Index() {
	d.Ctx.JSON(200, core.Map{
		"message": "Admin Dashboard",
		"stats": core.Map{
			"total_users":    100,
			"total_products": 50,
			"total_orders":   250,
		},
	})
}
