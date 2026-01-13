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

	// Register controllers
	core.Register(&WelcomeController{})

	// Enable auto-routing
	app.AutoRoute()

	// Serve static files
	app.Static("/static/", "./public")

	// Start server
	log.Println("Server running at http://localhost:8080")
	log.Fatal(app.Run(":8080"))
}

// =============================================================================
// Welcome Controller
// =============================================================================

type WelcomeController struct {
	core.Controller
}

// Index - GET /welcomecontroller
func (w *WelcomeController) Index() {
	w.Ctx.View("welcome", core.Map{
		"Title":   "Welcome to GoIgniter!",
		"Message": "Your application is ready.",
	})
}
