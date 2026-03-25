package database

import (
	"log"
	"time"

	"github.com/semutdev/goigniter/system/libraries/database"
	"golang.org/x/crypto/bcrypt"
)

// Seed runs the database seeder
func Seed(db *database.DB) {
	log.Println("Running database seeder...")

	// 1. Create tables
	createTables(db)

	// 2. Create groups
	seedGroups(db)

	// 3. Create default admin user
	seedAdminUser(db)

	log.Println("Database seeding completed!")
}

func createTables(db *database.DB) {
	// MySQL schema
	tables := []string{
		`CREATE TABLE IF NOT EXISTS groups (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(20) NOT NULL UNIQUE,
			description VARCHAR(100)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			ip_address VARCHAR(45),
			username VARCHAR(100),
			password VARCHAR(255) NOT NULL,
			email VARCHAR(254) NOT NULL UNIQUE,
			activation_selector VARCHAR(255),
			activation_code VARCHAR(255),
			forgotten_password_selector VARCHAR(255),
			forgotten_password_code VARCHAR(255),
			forgotten_password_time INT,
			remember_selector VARCHAR(255),
			remember_code VARCHAR(255),
			created_on INT,
			last_login INT,
			active TINYINT(1) DEFAULT 0,
			first_name VARCHAR(50),
			last_name VARCHAR(50),
			company VARCHAR(100),
			phone VARCHAR(20)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS users_groups (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			group_id INT NOT NULL,
			UNIQUE KEY unique_user_group (user_id, group_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS login_attempts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			ip_address VARCHAR(45),
			login VARCHAR(100),
			time INT
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(15,2) NOT NULL,
			stock INT DEFAULT 0,
			image VARCHAR(255) DEFAULT '',
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
}

func runMigrations(db *database.DB) {
	// Add image column to products table if not exists
	var imageColumnExists int64
	db.Query(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'products'
		AND COLUMN_NAME = 'image'
	`).Get(&imageColumnExists)

	if imageColumnExists == 0 {
		_, err := db.Exec("ALTER TABLE products ADD COLUMN image VARCHAR(255) DEFAULT '' AFTER stock")
		if err != nil {
			log.Printf("Error adding image column: %v\n", err)
		} else {
			log.Println("Added image column to products table")
		}
	}
}

func seedGroups(db *database.DB) {
	groups := []struct {
		Name        string
		Description string
	}{
		{"admin", "Administrator"},
		{"members", "General User"},
	}

	for _, g := range groups {
		// Check if exists
		var count int64
		db.Query("SELECT COUNT(*) FROM groups WHERE name = ?", g.Name).Get(&count)
		if count == 0 {
			_, err := db.Exec("INSERT INTO groups (name, description) VALUES (?, ?)", g.Name, g.Description)
			if err != nil {
				log.Printf("Error creating group %s: %v\n", g.Name, err)
			} else {
				log.Printf("Created group: %s\n", g.Name)
			}
		}
	}
}

func seedAdminUser(db *database.DB) {
	// Check if admin exists
	var count int64
	db.Query("SELECT COUNT(*) FROM users WHERE email = ?", "admin@admin.com").Get(&count)
	if count > 0 {
		log.Println("Admin user already exists, skipping...")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		return
	}

	firstName := "Admin"
	lastName := "istrator"
	username := "administrator"
	now := time.Now().Unix()

	// Insert admin user
	result, err := db.Exec(`
		INSERT INTO users (email, username, password, active, first_name, last_name, created_on, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "admin@admin.com", username, string(hashedPassword), 1, firstName, lastName, now, "127.0.0.1")
	if err != nil {
		log.Printf("Error creating admin user: %v\n", err)
		return
	}

	userID, _ := result.LastInsertId()

	// Get group IDs
	var adminGroupID, membersGroupID int64
	db.Query("SELECT id FROM groups WHERE name = ?", "admin").Get(&adminGroupID)
	db.Query("SELECT id FROM groups WHERE name = ?", "members").Get(&membersGroupID)

	// Assign user to groups
	db.Exec("INSERT INTO users_groups (user_id, group_id) VALUES (?, ?)", userID, adminGroupID)
	db.Exec("INSERT INTO users_groups (user_id, group_id) VALUES (?, ?)", userID, membersGroupID)

	log.Println("Created admin user: admin@admin.com / password")
}