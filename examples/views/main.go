package main

import (
	"fmt"
	"log"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/helpers"
	"github.com/semutdev/goigniter/system/libraries/session"
	"github.com/semutdev/goigniter/system/middleware"
)

func main() {
	app := core.New()

	// Initialize helpers with base URL
	helpers.Init("http://localhost:8080")

	// Initialize session
	session.Init(session.Config{
		Secret: "your-secret-key-change-this",
		MaxAge: 86400,
	})

	// Load templates with helper functions
	if err := app.LoadTemplatesWithFuncs("./views", true, helpers.AllTemplateFuncs()); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Serve static files
	app.Static("/public/", "./public")

	// === ROUTES ===

	// Home page
	app.GET("/", func(c *core.Context) error {
		return c.View("pages/home", core.Map{
			"title":   "Home",
			"message": "Welcome to GoIgniter!",
		})
	})

	// About page
	app.GET("/about", func(c *core.Context) error {
		return c.View("pages/about", core.Map{
			"title": "About Us",
			"team": []core.Map{
				{"name": "John Doe", "role": "Developer"},
				{"name": "Jane Smith", "role": "Designer"},
				{"name": "Bob Wilson", "role": "Manager"},
			},
		})
	})

	// Products page with data
	app.GET("/products", func(c *core.Context) error {
		products := []core.Map{
			{"id": 1, "name": "Laptop", "price": 15000000, "stock": 10},
			{"id": 2, "name": "Mouse", "price": 250000, "stock": 50},
			{"id": 3, "name": "Keyboard", "price": 750000, "stock": 30},
			{"id": 4, "name": "Monitor", "price": 3500000, "stock": 15},
		}
		return c.View("pages/products", core.Map{
			"title":    "Products",
			"products": products,
		})
	})

	// Product detail
	app.GET("/products/:id", func(c *core.Context) error {
		id := c.Param("id")
		// Simulate fetching product from database
		product := core.Map{
			"id":          id,
			"name":        "Product " + id,
			"price":       1500000,
			"stock":       25,
			"description": "This is a detailed description for product " + id,
		}
		return c.View("pages/product_detail", core.Map{
			"title":   "Product Detail",
			"product": product,
		})
	})

	// Contact page with flash message demo
	app.GET("/contact", func(c *core.Context) error {
		flash := session.GetFlash(c, "success")
		return c.View("pages/contact", core.Map{
			"title": "Contact Us",
			"flash": flash,
		})
	})

	// Handle contact form submission
	app.POST("/contact", func(c *core.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		message := c.FormValue("message")

		// Here you would normally process the form (send email, save to DB, etc.)
		_ = name
		_ = email
		_ = message

		// Set flash message
		session.SetFlash(c, "success", "Thank you for your message! We'll get back to you soon.")

		// Redirect back to contact page
		return c.Redirect(302, "/contact")
	})

	// Start server
	port := ":8080"
	fmt.Println("=================================")
	fmt.Println("  GoIgniter - Views Example")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Server running at http://localhost" + port)
	fmt.Println()
	fmt.Println("Pages:")
	fmt.Println("  GET  /              - Home page")
	fmt.Println("  GET  /about         - About page")
	fmt.Println("  GET  /products      - Products list")
	fmt.Println("  GET  /products/:id  - Product detail")
	fmt.Println("  GET  /contact       - Contact form")
	fmt.Println("  POST /contact       - Submit contact form")
	fmt.Println()

	log.Fatal(app.Run(port))
}
