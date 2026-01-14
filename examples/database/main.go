package main

import (
	"fmt"
	"log"
	"os"

	"github.com/semutdev/goigniter/system/core"
	"github.com/semutdev/goigniter/system/libraries/database"
	_ "github.com/semutdev/goigniter/system/libraries/database/drivers" // Register drivers
	"github.com/semutdev/goigniter/system/middleware"
)

// User model
type User struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Email  string `db:"email"`
	Status string `db:"status"`
}

// Product model
type Product struct {
	ID    int64   `db:"id"`
	Name  string  `db:"name"`
	Price float64 `db:"price"`
	Stock int     `db:"stock"`
}

var db *database.DB
var dbDriver string

func main() {
	// Check environment for database driver
	// Default: SQLite
	// For MySQL: DB_DRIVER=mysql DB_DSN="user:password@tcp(localhost:3306)/dbname?parseTime=true"
	dbDriver = os.Getenv("DB_DRIVER")
	dbDSN := os.Getenv("DB_DSN")

	if dbDriver == "" {
		dbDriver = "sqlite"
	}
	if dbDSN == "" {
		if dbDriver == "sqlite" {
			dbDSN = "./app.db"
		} else {
			log.Fatal("DB_DSN environment variable is required for non-SQLite databases")
		}
	}

	var err error
	db, err = database.Open(dbDriver, dbDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set as default for global access
	database.SetDefault(db)

	// Create tables
	setupDatabase()

	// Seed sample data
	seedData()

	// Create app
	app := core.New()

	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Routes
	app.GET("/", indexHandler)
	app.GET("/users", listUsers)
	app.GET("/users/:id", showUser)
	app.GET("/products", listProducts)
	app.GET("/stats", showStats)
	app.GET("/demo/insert", demoInsert)
	app.GET("/demo/update", demoUpdate)
	app.GET("/demo/delete", demoDelete)
	app.GET("/demo/transaction", demoTransaction)
	app.GET("/demo/raw", demoRawQuery)

	fmt.Println("===========================================")
	fmt.Println("  GoIgniter - Database Query Builder Demo")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Printf("Database driver: %s\n", dbDriver)
	fmt.Println("Server running at http://localhost:8080")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET /              - API info")
	fmt.Println("  GET /users         - List all users")
	fmt.Println("  GET /users/:id     - Get user by ID")
	fmt.Println("  GET /products      - List products with filters")
	fmt.Println("  GET /stats         - Aggregate functions demo")
	fmt.Println("  GET /demo/insert   - Insert demo")
	fmt.Println("  GET /demo/update   - Update demo")
	fmt.Println("  GET /demo/delete   - Delete demo")
	fmt.Println("  GET /demo/transaction - Transaction demo")
	fmt.Println("  GET /demo/raw      - Raw query demo")
	fmt.Println()

	log.Fatal(app.Run(":8080"))
}

func setupDatabase() {
	if dbDriver == "mysql" {
		// MySQL syntax
		db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				email VARCHAR(255) NOT NULL UNIQUE,
				status VARCHAR(50) DEFAULT 'active'
			)
		`)

		db.Exec(`
			CREATE TABLE IF NOT EXISTS products (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				price DECIMAL(15,2) NOT NULL,
				stock INT DEFAULT 0
			)
		`)

		db.Exec(`
			CREATE TABLE IF NOT EXISTS orders (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT NOT NULL,
				product_id INT NOT NULL,
				quantity INT NOT NULL,
				total DECIMAL(15,2) NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (product_id) REFERENCES products(id)
			)
		`)
	} else {
		// SQLite syntax
		db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT NOT NULL UNIQUE,
				status TEXT DEFAULT 'active'
			)
		`)

		db.Exec(`
			CREATE TABLE IF NOT EXISTS products (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				price REAL NOT NULL,
				stock INTEGER DEFAULT 0
			)
		`)

		db.Exec(`
			CREATE TABLE IF NOT EXISTS orders (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				product_id INTEGER NOT NULL,
				quantity INTEGER NOT NULL,
				total REAL NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (product_id) REFERENCES products(id)
			)
		`)
	}
}

func seedData() {
	// Check if data exists
	count, _ := db.Table("users").Count()
	if count > 0 {
		return
	}

	// Seed users
	users := []map[string]any{
		{"name": "John Doe", "email": "john@example.com", "status": "active"},
		{"name": "Jane Smith", "email": "jane@example.com", "status": "active"},
		{"name": "Bob Wilson", "email": "bob@example.com", "status": "inactive"},
		{"name": "Alice Brown", "email": "alice@example.com", "status": "active"},
		{"name": "Charlie Davis", "email": "charlie@example.com", "status": "pending"},
	}
	for _, u := range users {
		db.Table("users").Insert(u)
	}

	// Seed products
	products := []map[string]any{
		{"name": "Laptop", "price": 15000000, "stock": 10},
		{"name": "Mouse", "price": 250000, "stock": 50},
		{"name": "Keyboard", "price": 750000, "stock": 30},
		{"name": "Monitor", "price": 3500000, "stock": 15},
		{"name": "Headphone", "price": 500000, "stock": 25},
	}
	for _, p := range products {
		db.Table("products").Insert(p)
	}

	// Seed orders
	orders := []map[string]any{
		{"user_id": 1, "product_id": 1, "quantity": 1, "total": 15000000},
		{"user_id": 1, "product_id": 2, "quantity": 2, "total": 500000},
		{"user_id": 2, "product_id": 3, "quantity": 1, "total": 750000},
		{"user_id": 3, "product_id": 4, "quantity": 1, "total": 3500000},
		{"user_id": 4, "product_id": 5, "quantity": 3, "total": 1500000},
	}
	for _, o := range orders {
		db.Table("orders").Insert(o)
	}

	log.Println("Sample data seeded successfully!")
}

// Handlers

func indexHandler(c *core.Context) error {
	return c.JSON(200, core.Map{
		"message": "GoIgniter Query Builder Demo",
		"endpoints": core.Map{
			"GET /users":            "List all users",
			"GET /users/:id":        "Get user by ID",
			"GET /products":         "List products (supports ?status=active, ?min_price=1000)",
			"GET /stats":            "Aggregate functions demo",
			"GET /demo/insert":      "Insert demo",
			"GET /demo/update":      "Update demo",
			"GET /demo/delete":      "Delete demo",
			"GET /demo/transaction": "Transaction demo",
			"GET /demo/raw":         "Raw query demo",
		},
	})
}

func listUsers(c *core.Context) error {
	status := c.Query("status")

	builder := db.Table("users")

	// Filter by status if provided
	if status != "" {
		builder = builder.Where("status", status)
	}

	// Get as struct
	var users []User
	err := builder.OrderBy("name", "ASC").Get(&users)
	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	// Also show the SQL for learning
	sql := db.Table("users").OrderBy("name", "ASC").ToSQL()

	return c.JSON(200, core.Map{
		"users": users,
		"count": len(users),
		"sql":   sql,
	})
}

func showUser(c *core.Context) error {
	id := c.Param("id")

	var users []User
	err := db.Table("users").Where("id", id).First(&users)
	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	if len(users) == 0 {
		return c.JSON(404, core.Map{"error": "User not found"})
	}

	// Get user's orders with join
	orders, _ := db.Table("orders").
		Select("orders.*", "products.name as product_name").
		Join("products", "orders.product_id", "=", "products.id").
		Where("orders.user_id", id).
		GetMap()

	return c.JSON(200, core.Map{
		"user":   users[0],
		"orders": orders,
	})
}

func listProducts(c *core.Context) error {
	minPrice := c.Query("min_price")
	inStock := c.Query("in_stock")

	builder := db.Table("products")

	if minPrice != "" {
		builder = builder.Where("price", ">=", minPrice)
	}

	if inStock == "true" {
		builder = builder.Where("stock", ">", 0)
	}

	var products []Product
	err := builder.OrderBy("price", "DESC").Get(&products)
	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	return c.JSON(200, core.Map{
		"products": products,
		"count":    len(products),
	})
}

func showStats(c *core.Context) error {
	// Count
	totalUsers, _ := db.Table("users").Count()
	activeUsers, _ := db.Table("users").Where("status", "active").Count()

	// Sum
	totalRevenue, _ := db.Table("orders").Sum("total")

	// Avg
	avgPrice, _ := db.Table("products").Avg("price")

	// Min & Max
	minPrice, _ := db.Table("products").Min("price")
	maxPrice, _ := db.Table("products").Max("price")

	// Total stock
	totalStock, _ := db.Table("products").Sum("stock")

	return c.JSON(200, core.Map{
		"users": core.Map{
			"total":  totalUsers,
			"active": activeUsers,
		},
		"products": core.Map{
			"avg_price":   avgPrice,
			"min_price":   minPrice,
			"max_price":   maxPrice,
			"total_stock": totalStock,
		},
		"orders": core.Map{
			"total_revenue": totalRevenue,
		},
	})
}

func demoInsert(c *core.Context) error {
	// Insert with map
	err := db.Table("users").Insert(map[string]any{
		"name":   "New User",
		"email":  fmt.Sprintf("newuser%d@example.com", randomInt()),
		"status": "active",
	})
	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	// Insert and get ID
	id, _ := db.Table("products").InsertGetId(map[string]any{
		"name":  "New Product",
		"price": 999000,
		"stock": 5,
	})

	// Insert from struct
	user := User{
		Name:   "Struct User",
		Email:  fmt.Sprintf("structuser%d@example.com", randomInt()),
		Status: "pending",
	}
	db.Table("users").InsertStruct(&user)

	return c.JSON(200, core.Map{
		"message":            "Insert demo completed",
		"new_product_id":     id,
		"methods_demonstrated": []string{
			"Insert(map[string]any{})",
			"InsertGetId(map[string]any{})",
			"InsertStruct(&struct{})",
		},
	})
}

func demoUpdate(c *core.Context) error {
	// Update with map
	err := db.Table("users").
		Where("status", "pending").
		Update(map[string]any{
			"status": "active",
		})
	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	// Show the affected users
	var users []User
	db.Table("users").Where("status", "active").Get(&users)

	return c.JSON(200, core.Map{
		"message":      "Update demo completed",
		"active_users": len(users),
		"sql_example":  "UPDATE users SET status = 'active' WHERE status = 'pending'",
	})
}

func demoDelete(c *core.Context) error {
	// Count before
	countBefore, _ := db.Table("products").Count()

	// Delete products with name starting with "New"
	db.Table("products").WhereRaw("name LIKE ?", "New%").Delete()

	// Count after
	countAfter, _ := db.Table("products").Count()

	return c.JSON(200, core.Map{
		"message":       "Delete demo completed",
		"count_before":  countBefore,
		"count_after":   countAfter,
		"deleted":       countBefore - countAfter,
		"sql_example":   "DELETE FROM products WHERE name LIKE 'New%'",
	})
}

func demoTransaction(c *core.Context) error {
	// Transaction that succeeds
	err := db.Transaction(func(tx *database.DB) error {
		tx.Table("users").Insert(map[string]any{
			"name":   "Transaction User 1",
			"email":  fmt.Sprintf("txuser1_%d@example.com", randomInt()),
			"status": "active",
		})

		tx.Table("users").Insert(map[string]any{
			"name":   "Transaction User 2",
			"email":  fmt.Sprintf("txuser2_%d@example.com", randomInt()),
			"status": "active",
		})

		return nil // Commit
	})

	if err != nil {
		return c.JSON(500, core.Map{"error": err.Error()})
	}

	count, _ := db.Table("users").Count()

	return c.JSON(200, core.Map{
		"message":     "Transaction demo completed",
		"total_users": count,
		"code_example": `db.Transaction(func(tx *database.DB) error {
    tx.Table("users").Insert(...)
    tx.Table("orders").Insert(...)
    return nil // Commit
    // return errors.New("...") // Rollback
})`,
	})
}

func demoRawQuery(c *core.Context) error {
	// Raw SELECT
	var users []User
	db.Query("SELECT * FROM users WHERE status = ? ORDER BY name LIMIT ?", "active", 3).Get(&users)

	// Raw SELECT to map
	results, _ := db.Query("SELECT name, COUNT(*) as count FROM users GROUP BY status").GetMap()

	// Raw EXEC
	db.Exec("UPDATE products SET stock = stock + 1 WHERE id = ?", 1)

	return c.JSON(200, core.Map{
		"message":       "Raw query demo completed",
		"users":         users,
		"status_counts": results,
		"examples": []string{
			`db.Query("SELECT * FROM users WHERE status = ?", "active").Get(&users)`,
			`db.Query("SELECT * FROM users").GetMap()`,
			`db.Exec("UPDATE products SET stock = stock + 1 WHERE id = ?", 1)`,
		},
	})
}

func randomInt() int {
	count, _ := db.Table("users").Count()
	return int(count) + 1000
}
