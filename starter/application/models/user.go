package models

import (
	"time"

	"github.com/semutdev/goigniter/system/libraries/database"
)

// User model
type User struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name
func (User) TableName() string {
	return "users"
}

// UserModel provides user-related database operations (Fat Model)
type UserModel struct{}

// NewUserModel creates a new UserModel instance
func NewUserModel() *UserModel {
	return &UserModel{}
}

// All retrieves all users
func (m *UserModel) All() ([]User, error) {
	var users []User
	err := database.Table("users").OrderBy("id", "desc").Get(&users)
	return users, err
}

// Find finds a user by ID
func (m *UserModel) Find(id int64) (*User, error) {
	var user User
	err := database.Table("users").Where("id", id).First(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (m *UserModel) FindByEmail(email string) (*User, error) {
	var user User
	err := database.Table("users").Where("email", email).First(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (m *UserModel) Create(name, email string) (*User, error) {
	now := time.Now()
	id, err := database.Table("users").InsertGetId(map[string]any{
		"name":       name,
		"email":      email,
		"created_at": now,
		"updated_at": now,
	})
	if err != nil {
		return nil, err
	}
	return m.Find(id)
}

// Update updates a user
func (m *UserModel) Update(id int64, name, email string) error {
	return database.Table("users").Where("id", id).Update(map[string]any{
		"name":       name,
		"email":      email,
		"updated_at": time.Now(),
	})
}

// Delete deletes a user
func (m *UserModel) Delete(id int64) error {
	return database.Table("users").Where("id", id).Delete()
}

// Count returns total number of users
func (m *UserModel) Count() (int64, error) {
	return database.Table("users").Count()
}

// Exists checks if email already exists (optionally exclude a user ID)
func (m *UserModel) Exists(email string, excludeID ...int64) (bool, error) {
	query := database.Table("users").Where("email", email)
	if len(excludeID) > 0 {
		query = query.Where("id !=", excludeID[0])
	}
	count, err := query.Count()
	return count > 0, err
}