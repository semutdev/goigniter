package config

import (
	"fmt"
	"log"
	"os"

	"github.com/semutdev/goigniter/system/libraries/database"
	_ "github.com/semutdev/goigniter/system/libraries/database/drivers"
)

// DB is the global database instance
var DB *database.DB

// ConnectDB initializes the database connection
func ConnectDB() {
	driver := getEnv("DB_DRIVER", "mysql")
	dsn := buildDSN(driver)

	var err error
	DB, err = database.Open(driver, dsn)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Set as global default
	database.SetDefault(DB)

	log.Println("Database connected!")
}

func buildDSN(driver string) string {
	switch driver {
	case "sqlite":
		return getEnv("DB_DSN", "./app.db")
	default:
		// MySQL
		host := getEnv("DB_HOST", "127.0.0.1")
		port := getEnv("DB_PORT", "3306")
		user := getEnv("DB_USER", "root")
		password := os.Getenv("DB_PASSWORD")
		name := getEnv("DB_NAME", "myapp")
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			user, password, host, port, name)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}