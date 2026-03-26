package database

import (
	"log"

	"github.com/semutdev/goigniter/system/libraries/database"
)

// Seed runs the database seeder
func Seed(db *database.DB) {
	log.Println("Running database seeder...")

	// Create tables
	createTables(db)

	log.Println("Database seeding completed!")
}

func createTables(db *database.DB) {
	// MySQL schema - customize as needed
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, sql := range tables {
		_, err := db.Exec(sql)
		if err != nil {
			log.Printf("Error creating table: %v\n", err)
		}
	}

	// Run migrations for existing tables
	runMigrations(db)

	log.Println("Tables created successfully!")
}

func runMigrations(db *database.DB) {
	// Add future migrations here
	// Example: Add column if not exists
	// var columnExists int64
	// db.Query(`SELECT COUNT(*) FROM information_schema.COLUMNS WHERE ...`).Get(&columnExists)
	// if columnExists == 0 {
	//     db.Exec("ALTER TABLE users ADD COLUMN ...")
	// }
}
