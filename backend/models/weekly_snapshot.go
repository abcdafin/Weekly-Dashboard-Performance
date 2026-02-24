package models

import (
	"time"

	"gorm.io/gorm"
)

// WeeklySnapshot stores historical KPI data for week-over-week comparison
type WeeklySnapshot struct {
	gorm.Model
	IndicatorID      string    `gorm:"size:50;not null;index" json:"indicator_id"`
	Department       string    `gorm:"size:50;not null" json:"department"`
	IndicatorName    string    `gorm:"size:100;not null" json:"indicator_name"`
	TargetValue      float64   `gorm:"type:decimal(15,2)" json:"target_value"`
	PerformanceValue float64   `gorm:"type:decimal(15,2)" json:"performance_value"`
	Percentage       float64   `gorm:"type:decimal(5,2)" json:"percentage"`
	SnapshotDate     time.Time `gorm:"not null;index" json:"snapshot_date"`
	Month            int       `gorm:"not null;index" json:"month"` // 1-12
	WeekNumber       int       `gorm:"not null" json:"week_number"`
	Year             int       `gorm:"not null;index" json:"year"`
}

// TableName specifies the table name for WeeklySnapshot model
func (WeeklySnapshot) TableName() string {
	return "weekly_snapshots"
}
