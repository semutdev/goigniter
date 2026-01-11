package models

// LoginAttempt untuk tracking failed login attempts
type LoginAttempt struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	IPAddress string `gorm:"size:45" json:"ip_address"`
	Login     string `gorm:"size:100" json:"login"`
	Time      int64  `json:"time"`
}
