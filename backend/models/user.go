package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user authenticated via Google OAuth
type User struct {
	gorm.Model
	Email        string    `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Name         string    `gorm:"size:100" json:"name"`
	Picture      string    `gorm:"size:255" json:"picture"`
	AccessToken  string    `gorm:"type:text" json:"-"`
	RefreshToken string    `gorm:"type:text" json:"-"`
	TokenExpiry  time.Time `json:"-"`
	LastLogin    time.Time `json:"last_login"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
