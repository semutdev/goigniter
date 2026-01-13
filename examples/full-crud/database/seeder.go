package database

import (
	"full-crud/application/models"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed menjalankan seeder untuk data awal
func Seed(db *gorm.DB) {
	log.Println("Running database seeder...")

	// 1. Create groups
	seedGroups(db)

	// 2. Create default admin user
	seedAdminUser(db)

	log.Println("Database seeding completed!")
}

func seedGroups(db *gorm.DB) {
	groups := []models.Group{
		{ID: 1, Name: "admin", Description: "Administrator"},
		{ID: 2, Name: "members", Description: "General User"},
	}

	for _, group := range groups {
		// Cek apakah sudah ada
		var existing models.Group
		result := db.Where("name = ?", group.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&group)
			log.Printf("Created group: %s\n", group.Name)
		}
	}
}

func seedAdminUser(db *gorm.DB) {
	// Cek apakah admin sudah ada
	var existing models.User
	result := db.Where("email = ?", "admin@admin.com").First(&existing)
	if result.Error == nil {
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

	admin := models.User{
		ID:        1,
		Email:     "admin@admin.com",
		Username:  &username,
		Password:  string(hashedPassword),
		Active:    true,
		FirstName: &firstName,
		LastName:  &lastName,
		CreatedOn: time.Now().Unix(),
		IPAddress: "127.0.0.1",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Error creating admin user: %v\n", err)
		return
	}

	// Assign admin to groups
	var adminGroup models.Group
	var membersGroup models.Group
	db.Where("name = ?", "admin").First(&adminGroup)
	db.Where("name = ?", "members").First(&membersGroup)

	db.Model(&admin).Association("Groups").Append(&adminGroup, &membersGroup)

	log.Println("Created admin user: admin@admin.com / password")
}

// Ptr helper untuk membuat pointer dari string
func Ptr(s string) *string {
	return &s
}
