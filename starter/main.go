package main

import (
	"log"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/middleware"
)

func main() {
	// Create application
	app := core.New()

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Load templates
	if err := app.LoadTemplates("./application/views", true); err != nil {
		log.Printf("Warning: Could not load templates: %v", err)
	}

	// Serve static files
	app.Static("/static/", "./public")

	// Welcome page at root
	app.GET("/", func(c *core.Context) error {
		return c.View("welcome", core.Map{
			"Title": "Welcome to GoIgniter!",
		})
	})

	// Register controllers for auto-routing (optional)
	// core.Register(&YourController{})
	// app.AutoRoute()

	// Start server
	log.Println("===========================================")
	log.Println("  GoIgniter - Your application is ready!")
	log.Println("===========================================")
	log.Println()
	log.Println("Server running at http://localhost:8080")
	log.Fatal(app.Run(":8080"))
}
