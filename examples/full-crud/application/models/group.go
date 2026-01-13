package models

// Group model untuk role-based access
type Group struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:20;uniqueIndex" json:"name"`
	Description string `gorm:"size:100" json:"description"`
}
