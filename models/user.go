package models

// User model berdasarkan Ion Auth 3
type User struct {
	ID                        uint    `gorm:"primaryKey" json:"id"`
	IPAddress                 string  `gorm:"size:45" json:"ip_address"`
	Username                  *string `gorm:"size:100" json:"username"`
	Password                  string  `gorm:"size:255" json:"-"`
	Email                     string  `gorm:"size:254;uniqueIndex" json:"email"`
	ActivationSelector        *string `gorm:"size:255" json:"-"`
	ActivationCode            *string `gorm:"size:255" json:"-"`
	ForgottenPasswordSelector *string `gorm:"size:255" json:"-"`
	ForgottenPasswordCode     *string `gorm:"size:255" json:"-"`
	ForgottenPasswordTime     *int64  `json:"-"`
	RememberSelector          *string `gorm:"size:255" json:"-"`
	RememberCode              *string `gorm:"size:255" json:"-"`
	CreatedOn                 int64   `json:"created_on"`
	LastLogin                 *int64  `json:"last_login"`
	Active                    bool    `gorm:"default:false" json:"active"`
	FirstName                 *string `gorm:"size:50" json:"first_name"`
	LastName                  *string `gorm:"size:50" json:"last_name"`
	Company                   *string `gorm:"size:100" json:"company"`
	Phone                     *string `gorm:"size:20" json:"phone"`

	// Relasi many-to-many dengan Group
	Groups []Group `gorm:"many2many:users_groups;" json:"groups"`
}

// FullName return nama lengkap user
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
