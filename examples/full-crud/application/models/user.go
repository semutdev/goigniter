package models

import "time"

// User model based on Ion Auth 3
type User struct {
	ID                        int64   `json:"id" db:"id"`
	IPAddress                 string  `json:"ip_address" db:"ip_address"`
	Username                  *string `json:"username" db:"username"`
	Password                  string  `json:"-" db:"password"`
	Email                     string  `json:"email" db:"email"`
	ActivationSelector        *string `json:"-" db:"activation_selector"`
	ActivationCode            *string `json:"-" db:"activation_code"`
	ForgottenPasswordSelector *string `json:"-" db:"forgotten_password_selector"`
	ForgottenPasswordCode     *string `json:"-" db:"forgotten_password_code"`
	ForgottenPasswordTime     *int64  `json:"-" db:"forgotten_password_time"`
	RememberSelector          *string `json:"-" db:"remember_selector"`
	RememberCode              *string `json:"-" db:"remember_code"`
	CreatedOn                 int64   `json:"created_on" db:"created_on"`
	LastLogin                 *int64  `json:"last_login" db:"last_login"`
	Active                    bool    `json:"active" db:"active"`
	FirstName                 *string `json:"first_name" db:"first_name"`
	LastName                  *string `json:"last_name" db:"last_name"`
	Company                   *string `json:"company" db:"company"`
	Phone                     *string `json:"phone" db:"phone"`

	// Loaded separately, not from DB
	Groups []Group `json:"groups"`
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	first := ""
	last := ""
	if u.FirstName != nil {
		first = *u.FirstName
	}
	if u.LastName != nil {
		last = *u.LastName
	}
	if first == "" && last == "" {
		return u.Email
	}
	return first + " " + last
}

// Group model for role-based access
type Group struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// LoginAttempt for tracking failed login attempts
type LoginAttempt struct {
	ID        int64  `json:"id" db:"id"`
	IPAddress string `json:"ip_address" db:"ip_address"`
	Login     string `json:"login" db:"login"`
	Time      int64  `json:"time" db:"time"`
}

// Product model
type Product struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Price     float64   `json:"price" db:"price"`
	Stock     int       `json:"stock" db:"stock"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for each model
func (User) TableName() string         { return "users" }
func (Group) TableName() string        { return "groups" }
func (LoginAttempt) TableName() string { return "login_attempts" }
func (Product) TableName() string      { return "products" }
