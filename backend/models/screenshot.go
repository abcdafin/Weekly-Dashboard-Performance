package models

import (
	"time"

	"gorm.io/gorm"
)

// Screenshot represents a weekly dashboard screenshot stored in database
type Screenshot struct {
	gorm.Model
	Month     int       `gorm:"not null;index" json:"month"`
	Year      int       `gorm:"not null;index" json:"year"`
	Week      int       `gorm:"not null" json:"week"`
	Filename  string    `gorm:"size:100;not null" json:"filename"`
	ImageData []byte    `gorm:"type:bytea;not null" json:"-"` // Store image as binary, don't include in JSON
	MimeType  string    `gorm:"size:50;default:'image/png'" json:"mime_type"`
	SizeBytes int64     `gorm:"not null" json:"size_bytes"`
	SavedAt   time.Time `gorm:"not null" json:"saved_at"`
}

// TableName specifies the table name for Screenshot model
func (Screenshot) TableName() string {
	return "screenshots"
}
