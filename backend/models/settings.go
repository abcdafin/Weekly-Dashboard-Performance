package models

import (
	"gorm.io/gorm"
)

// AppSetting stores application configuration as key-value pairs
type AppSetting struct {
	gorm.Model
	Key   string `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value string `gorm:"size:500;not null" json:"value"`
}

// TableName specifies the table name for AppSetting model
func (AppSetting) TableName() string {
	return "app_settings"
}

// Setting keys constants
const (
	SettingSpreadsheetID = "spreadsheet_id"
	SettingSheetName     = "sheet_name"
)
