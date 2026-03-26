package main

import (
	"log"
	"os"

	"myapp/application/config"
	"myapp/application/controllers"
	"myapp/database"

	"github.com/joho/godotenv"
	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/helpers"
	"github.com/semutdev/goigniter/system/middleware"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Connect to database
	config.ConnectDB()

	// Run seeder if DB_SEED=true (creates tables)
	if os.Getenv("DB_SEED") == "true" {
		database.Seed(config.DB)
	}

	// Create application
	app := core.New()

	// Initialize helpers
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":8080"
	}
	helpers.Init("http://localhost" + port)

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Load templates with helper functions
	if err := app.LoadTemplatesWithFuncs("./application/views", true, helpers.AllTemplateFuncs()); err != nil {
		log.Printf("Warning: Could not load templates: %v", err)
	}

	// Serve static files
	app.Static("/static/", "./public")

	// Register controllers
	core.Register(&controllers.Welcome{})

	// Enable auto-routing
	// GET /welcome -> Welcome.Index()
	app.AutoRoute()

	// Root redirect to /welcome
	app.GET("/", func(c *core.Context) error {
		return c.Redirect(302, "/welcome/index")
	})

	// Start server
	log.Fatal(app.Run(port))
}