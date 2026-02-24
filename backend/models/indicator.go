package models

import (
	"gorm.io/gorm"
)

// Indicator represents a KPI indicator master data
type Indicator struct {
	gorm.Model
	Code            string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Department      string `gorm:"size:50;not null" json:"department"`
	Name            string `gorm:"size:100;not null" json:"name"`
	UnitOfMeasure   string `gorm:"size:20" json:"unit_of_measure"`
	SpreadsheetName string `gorm:"size:100" json:"spreadsheet_name"` // Exact name in spreadsheet col C
	SpreadsheetRow  int    `gorm:"not null" json:"spreadsheet_row"`  // Fallback row number
	IsInverse       bool   `gorm:"default:false" json:"is_inverse"`  // For metrics like turnover where lower is better
	DisplayOrder    int    `json:"display_order"`
	IsActive        bool   `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for Indicator model
func (Indicator) TableName() string {
	return "indicators"
}

// GetDefaultIndicators returns the 12 main KPI indicators based on requirements
func GetDefaultIndicators() []Indicator {
	return []Indicator{
		{Code: "KPI-01", Department: "FINANCE", Name: "Revenue Group", UnitOfMeasure: "B", SpreadsheetName: "Revenue Group", SpreadsheetRow: 3, IsInverse: false, DisplayOrder: 1, IsActive: true},
		{Code: "KPI-02", Department: "MARKETING", Name: "MQL-SQL Conversion Rate", UnitOfMeasure: "%", SpreadsheetName: "MQL - SQL CR", SpreadsheetRow: 12, IsInverse: false, DisplayOrder: 2, IsActive: true},
		{Code: "KPI-03", Department: "SALES", Name: "Total Sales", UnitOfMeasure: "B", SpreadsheetName: "Total Sales", SpreadsheetRow: 14, IsInverse: false, DisplayOrder: 3, IsActive: true},
		{Code: "KPI-04", Department: "OPERATIONS", Name: "COGS & OPEX", UnitOfMeasure: "B", SpreadsheetName: "COGS & OPEX", SpreadsheetRow: 20, IsInverse: false, DisplayOrder: 4, IsActive: true},
		{Code: "KPI-05", Department: "FINANCE", Name: "% Collection (Ontime)", UnitOfMeasure: "%", SpreadsheetName: "% Collection (Ontime)", SpreadsheetRow: 22, IsInverse: false, DisplayOrder: 5, IsActive: true},
		{Code: "KPI-06", Department: "IT OPERATIONS", Name: "System Uptime", UnitOfMeasure: "%", SpreadsheetName: "System Uptime", SpreadsheetRow: 23, IsInverse: false, DisplayOrder: 6, IsActive: true},
		{Code: "KPI-07", Department: "PS", Name: "Non Billable Cost", UnitOfMeasure: "IDR", SpreadsheetName: "Non Billable Cost Ratio (max)", SpreadsheetRow: 27, IsInverse: true, DisplayOrder: 7, IsActive: true},
		{Code: "KPI-08", Department: "PS", Name: "Ontime Timesheet Collection", UnitOfMeasure: "%", SpreadsheetName: "Ontime Timesheet Approval Colledtion", SpreadsheetRow: 29, IsInverse: false, DisplayOrder: 8, IsActive: true},
		{Code: "KPI-09", Department: "DELIVERY", Name: "Customer Satisfaction", UnitOfMeasure: "score", SpreadsheetName: "Customer Satisfaction", SpreadsheetRow: 36, IsInverse: false, DisplayOrder: 9, IsActive: true},
		{Code: "KPI-10", Department: "HC", Name: "Turn Over", UnitOfMeasure: "people", SpreadsheetName: "Turn Over (max / up to)", SpreadsheetRow: 42, IsInverse: true, DisplayOrder: 10, IsActive: true},
		{Code: "KPI-11", Department: "BD", Name: "MQL Outbound", UnitOfMeasure: "leads", SpreadsheetName: "MQL Outbound", SpreadsheetRow: 47, IsInverse: false, DisplayOrder: 11, IsActive: true},
		{Code: "KPI-12", Department: "TA", Name: "PS Talents Placement", UnitOfMeasure: "people", SpreadsheetName: "PS Talents Placement", SpreadsheetRow: 60, IsInverse: false, DisplayOrder: 12, IsActive: true},
	}
}
