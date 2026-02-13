package main

import (
	"fmt"
	"log"
	"os"

	"full-crud/application/models"
	"full-crud/config"
	"full-crud/database"

	// Import controllers untuk auto-register
	_ "full-crud/application/controllers"
	_ "full-crud/application/controllers/admin"

	"github.com/joho/godotenv"
	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/helpers"
	"github.com/semutdev/goigniter/system/libraries/session"
	"github.com/semutdev/goigniter/system/middleware"
)

func main() {
	// Load env
	godotenv.Load()

	// Connect to DB
	config.ConnectDB()

	// Auto migrate tables
	config.DB.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.LoginAttempt{},
		&models.Product{},
	)

	// Run seeder jika DB_SEED=true
	if os.Getenv("DB_SEED") == "true" {
		database.Seed(config.DB)
	}

	// Create app
	app := core.New()

	// Initialize helpers
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = ":8080"
	}
	helpers.Init("http://localhost" + port)

	// Initialize session
	sessionSecret := os.Getenv("APP_KEY")
	if sessionSecret == "" {
		sessionSecret = "goigniter-default-secret-key-32"
	}
	session.Init(session.Config{
		Secret: sessionSecret,
		MaxAge: 86400,
	})

	// Load templates
	if err := app.LoadTemplatesWithFuncs("./application/views", true, helpers.AllTemplateFuncs()); err != nil {
		log.Printf("Warning: Could not load templates: %v", err)
	}

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Static files
	app.Static("/static/", "./public")

	// Auto-route dari registered controllers
	app.AutoRoute()

	// Default route
	app.GET("/", func(c *core.Context) error {
		return c.Redirect(302, "/welcome")
	})

	// Custom routes untuk auth dengan parameter
	app.GET("/auth/activate/:selector/:code", func(c *core.Context) error {
		// Forward to auth controller activate
		return c.Redirect(302, "/auth/activate")
	})

	app.GET("/auth/reset/:selector/:code", func(c *core.Context) error {
		return c.Redirect(302, "/auth/reset")
	})

	// Start server
	fmt.Println("=================================")
	fmt.Println("  GoIgniter - Full CRUD Example")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Server running at http://localhost" + port)
	fmt.Println()
	fmt.Println("Default admin login:")
	fmt.Println("  Email: admin@admin.com")
	fmt.Println("  Password: password")
	fmt.Println()
	fmt.Println("Note: Run with DB_SEED=true to create admin user")
	fmt.Println()

	log.Fatal(app.Run(port))
}
