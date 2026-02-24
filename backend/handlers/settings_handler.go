package handlers

import (
	"log"
	"net/http"
	"strings"

	"weekly-dashboard/config"
	"weekly-dashboard/database"
	"weekly-dashboard/models"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles settings endpoints
type SettingsHandler struct{}

// NewSettingsHandler creates a new SettingsHandler instance
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// SpreadsheetSettingsResponse represents the spreadsheet settings response
type SpreadsheetSettingsResponse struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
}

// UpdateSpreadsheetRequest represents the request to update spreadsheet settings
type UpdateSpreadsheetRequest struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
}

// GetSpreadsheetSettings returns current spreadsheet configuration
func (h *SettingsHandler) GetSpreadsheetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": SpreadsheetSettingsResponse{
			SpreadsheetID: config.AppConfig.SpreadsheetID,
			SheetName:     config.AppConfig.SheetName,
		},
	})
}

// UpdateSpreadsheetSettings updates spreadsheet configuration
func (h *SettingsHandler) UpdateSpreadsheetSettings(c *gin.Context) {
	var req UpdateSpreadsheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Extract spreadsheet ID from URL if full URL is provided
	spreadsheetID := extractSpreadsheetID(req.SpreadsheetID)
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Spreadsheet ID or URL is required",
		})
		return
	}

	sheetName := req.SheetName
	if sheetName == "" {
		sheetName = config.AppConfig.SheetName // Keep existing if not provided
	}

	// Save to database
	if err := upsertSetting(models.SettingSpreadsheetID, spreadsheetID); err != nil {
		log.Printf("Failed to save spreadsheet_id setting: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save settings",
		})
		return
	}

	if err := upsertSetting(models.SettingSheetName, sheetName); err != nil {
		log.Printf("Failed to save sheet_name setting: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save settings",
		})
		return
	}

	// Update runtime config
	config.AppConfig.SpreadsheetID = spreadsheetID
	config.AppConfig.SheetName = sheetName

	log.Printf("Spreadsheet settings updated: ID=%s, Sheet=%s", spreadsheetID, sheetName)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Spreadsheet settings updated successfully",
		"data": SpreadsheetSettingsResponse{
			SpreadsheetID: spreadsheetID,
			SheetName:     sheetName,
		},
	})
}

// extractSpreadsheetID extracts the spreadsheet ID from a full Google Sheets URL or returns as-is if already an ID
func extractSpreadsheetID(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	// Check if it's a full URL
	if strings.Contains(input, "docs.google.com/spreadsheets") {
		// URL format: https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/...
		parts := strings.Split(input, "/d/")
		if len(parts) >= 2 {
			idPart := parts[1]
			// Remove anything after the ID (e.g., /edit, /view, etc.)
			if slashIdx := strings.Index(idPart, "/"); slashIdx != -1 {
				idPart = idPart[:slashIdx]
			}
			return idPart
		}
	}

	// Already an ID
	return input
}

// upsertSetting creates or updates a setting in the database
func upsertSetting(key, value string) error {
	var setting models.AppSetting
	result := database.DB.Where("key = ?", key).First(&setting)

	if result.Error != nil {
		// Create new setting
		setting = models.AppSetting{
			Key:   key,
			Value: value,
		}
		return database.DB.Create(&setting).Error
	}

	// Update existing setting
	setting.Value = value
	return database.DB.Save(&setting).Error
}

// LoadSettingsFromDB loads settings from database and overrides AppConfig
func LoadSettingsFromDB() {
	var settings []models.AppSetting
	result := database.DB.Find(&settings)
	if result.Error != nil {
		log.Printf("Warning: Failed to load settings from database: %v", result.Error)
		return
	}

	for _, setting := range settings {
		switch setting.Key {
		case models.SettingSpreadsheetID:
			if setting.Value != "" {
				config.AppConfig.SpreadsheetID = setting.Value
				log.Printf("Loaded spreadsheet_id from database: %s", setting.Value)
			}
		case models.SettingSheetName:
			if setting.Value != "" {
				config.AppConfig.SheetName = setting.Value
				log.Printf("Loaded sheet_name from database: %s", setting.Value)
			}
		}
	}
}
