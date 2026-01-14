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

	// Initialize helpers
	helpers.Init("http://localhost:8080")

	// Initialize session
	session.Init(session.Config{
		Secret:   "your-secret-key-change-this",
		MaxAge:   86400,
		HttpOnly: true,
	})

	// Load templates with helpers
	if err := app.LoadTemplatesWithFuncs("./views", true, helpers.AllTemplateFuncs()); err != nil {
		log.Printf("Warning: Could not load templates: %v", err)
	}

	// Global middleware
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())

	// Basic routes
	app.GET("/", func(c *core.Context) error {
		return c.JSON(200, core.Map{
			"message":  "Welcome to GoIgniter!",
			"base_url": helpers.BaseURL(),
		})
	})

	app.GET("/hello/:name", func(c *core.Context) error {
		name := c.Param("name")
		return c.JSON(200, core.Map{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// HTML response
	app.GET("/html", func(c *core.Context) error {
		return c.HTML(200, `
			<!DOCTYPE html>
			<html>
			<head><title>GoIgniter</title></head>
			<body>
				<h1>GoIgniter</h1>
				<p>HTTP layer berbasis net/http stdlib</p>
				<ul>
					<li><a href="/">JSON Response</a></li>
					<li><a href="/hello/world">Path Params</a></li>
					<li><a href="/users?page=1&limit=10">Query Params</a></li>
					<li><a href="/api/v1/products">API Group</a></li>
					<li><a href="/session">Session Demo</a></li>
				</ul>
			</body>
			</html>
		`)
	})

	// Query params example
	app.GET("/users", func(c *core.Context) error {
		page := c.QueryIntDefault("page", 1)
		limit := c.QueryIntDefault("limit", 10)
		return c.JSON(200, core.Map{
			"page":  page,
			"limit": limit,
			"users": []core.Map{
				{"id": 1, "name": "John"},
				{"id": 2, "name": "Jane"},
			},
		})
	})

	// POST example
	app.POST("/users", func(c *core.Context) error {
		var input struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := c.Bind(&input); err != nil {
			return c.JSON(400, core.Map{"error": err.Error()})
		}
		return c.JSON(201, core.Map{
			"message": "User created",
			"user":    input,
		})
	})

	// Session demo
	app.GET("/session", func(c *core.Context) error {
		sess := session.Get(c)

		// Increment visit counter
		visits := sess.GetInt("visits") + 1
		sess.Set("visits", visits)
		sess.Save(c)

		return c.JSON(200, core.Map{
			"session_id": sess.ID,
			"visits":     visits,
			"message":    "Refresh to see visit counter increase",
		})
	})

	// Flash message demo
	app.GET("/flash/set", func(c *core.Context) error {
		session.SetFlash(c, "success", "This is a flash message!")
		return c.JSON(200, core.Map{
			"message": "Flash message set. Visit /flash/get to see it.",
		})
	})

	app.GET("/flash/get", func(c *core.Context) error {
		flash := session.GetFlash(c, "success")
		return c.JSON(200, core.Map{
			"flash":   flash,
			"message": "Flash message is now cleared. Refresh to see it's gone.",
		})
	})

	// Route groups
	api := app.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/products", func(c *core.Context) error {
				return c.JSON(200, core.Map{
					"products": []core.Map{
						{"id": 1, "name": "Product A", "price": 100},
						{"id": 2, "name": "Product B", "price": 200},
					},
				})
			})

			v1.GET("/products/:id", func(c *core.Context) error {
				id := c.Param("id")
				return c.JSON(200, core.Map{
					"id":    id,
					"name":  "Product " + id,
					"price": 150,
				})
			})
		}
	}

	// Protected route with auth middleware
	admin := app.Group("/admin", middleware.BasicAuth(func(user, pass string) bool {
		return user == "admin" && pass == "secret"
	}))
	{
		admin.GET("/dashboard", func(c *core.Context) error {
			return c.JSON(200, core.Map{
				"message": "Welcome to admin dashboard",
			})
		})
	}

	// Panic recovery demo
	app.GET("/panic", func(c *core.Context) error {
		panic("This is a test panic!")
	})

	// Start server
	port := ":8080"
	fmt.Println("=================================")
	fmt.Println("  GoIgniter - Test Server")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Server running at http://localhost" + port)
	fmt.Println()
	fmt.Println("Test endpoints:")
	fmt.Println("  GET  /              - JSON welcome")
	fmt.Println("  GET  /html          - HTML page")
	fmt.Println("  GET  /hello/:name   - Path params")
	fmt.Println("  GET  /users         - Query params")
	fmt.Println("  POST /users         - Create user")
	fmt.Println("  GET  /session       - Session demo")
	fmt.Println("  GET  /flash/set     - Set flash message")
	fmt.Println("  GET  /flash/get     - Get flash message")
	fmt.Println("  GET  /api/v1/products     - API group")
	fmt.Println("  GET  /admin/dashboard     - Protected (admin:secret)")
	fmt.Println("  GET  /panic         - Panic recovery")
	fmt.Println()

	log.Fatal(app.Run(port))
}
