package handlers

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"weekly-dashboard/database"
	"weekly-dashboard/middleware"
	"weekly-dashboard/models"

	"github.com/gin-gonic/gin"
)

// ScreenshotHandler handles screenshot endpoints
type ScreenshotHandler struct{}

// NewScreenshotHandler creates a new ScreenshotHandler instance
func NewScreenshotHandler() *ScreenshotHandler {
	return &ScreenshotHandler{}
}

// ScreenshotResponse represents the response for screenshot list
type ScreenshotResponse struct {
	ID        uint      `json:"id"`
	Month     int       `json:"month"`
	Year      int       `json:"year"`
	Week      int       `json:"week"`
	Filename  string    `json:"filename"`
	SizeBytes int64     `json:"size_bytes"`
	SavedAt   time.Time `json:"saved_at"`
}

// UploadScreenshot handles PNG screenshot upload and saves to database
func (h *ScreenshotHandler) UploadScreenshot(c *gin.Context) {
	_, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse form data
	monthStr := c.PostForm("month")
	yearStr := c.PostForm("year")
	weekStr := c.PostForm("week")

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid month value (1-12)",
		})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid year value",
		})
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid week value (1-5)",
		})
		return
	}

	// Get uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No file uploaded",
		})
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read uploaded file",
		})
		return
	}

	// Generate filename
	monthName := time.Month(month).String()
	filename := monthName + "_" + strconv.Itoa(year) + "_Week_" + strconv.Itoa(week) + ".png"

	// Check if screenshot already exists (upsert)
	var existingScreenshot models.Screenshot
	result := database.DB.Where("month = ? AND year = ? AND week = ?", month, year, week).First(&existingScreenshot)

	now := time.Now()
	if result.Error == nil {
		// Update existing screenshot
		existingScreenshot.ImageData = imageData
		existingScreenshot.SizeBytes = int64(len(imageData))
		existingScreenshot.SavedAt = now
		existingScreenshot.Filename = filename

		if err := database.DB.Save(&existingScreenshot).Error; err != nil {
			log.Printf("Failed to update screenshot: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to update screenshot",
			})
			return
		}
		log.Printf("Screenshot updated: %s (%d bytes)", filename, len(imageData))
	} else {
		// Create new screenshot
		screenshot := models.Screenshot{
			Month:     month,
			Year:      year,
			Week:      week,
			Filename:  filename,
			ImageData: imageData,
			MimeType:  "image/png",
			SizeBytes: int64(len(imageData)),
			SavedAt:   now,
		}

		if err := database.DB.Create(&screenshot).Error; err != nil {
			log.Printf("Failed to save screenshot: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to save screenshot",
			})
			return
		}
		log.Printf("Screenshot saved: %s (%d bytes)", filename, len(imageData))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Screenshot saved successfully",
		"data": gin.H{
			"filename":   filename,
			"month":      month,
			"year":       year,
			"week":       week,
			"size_bytes": len(imageData),
			"saved_at":   now.Format(time.RFC3339),
		},
	})
}

// GetScreenshots returns list of saved screenshots for a month/year
func (h *ScreenshotHandler) GetScreenshots(c *gin.Context) {
	_, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	var screenshots []models.Screenshot
	query := database.DB.Select("id, month, year, week, filename, size_bytes, saved_at, created_at, updated_at")

	if month >= 1 && month <= 12 && year > 0 {
		query = query.Where("month = ? AND year = ?", month, year)
	}

	query.Order("week ASC").Find(&screenshots)

	// Convert to response format
	var response []ScreenshotResponse
	for _, s := range screenshots {
		response = append(response, ScreenshotResponse{
			ID:        s.ID,
			Month:     s.Month,
			Year:      s.Year,
			Week:      s.Week,
			Filename:  s.Filename,
			SizeBytes: s.SizeBytes,
			SavedAt:   s.SavedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetScreenshotImage returns the actual image data for a screenshot
func (h *ScreenshotHandler) GetScreenshotImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid screenshot ID",
		})
		return
	}

	var screenshot models.Screenshot
	if err := database.DB.First(&screenshot, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Screenshot not found",
		})
		return
	}

	// Return image as base64 data URL
	base64Data := base64.StdEncoding.EncodeToString(screenshot.ImageData)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"filename":  screenshot.Filename,
			"mime_type": screenshot.MimeType,
			"image":     "data:" + screenshot.MimeType + ";base64," + base64Data,
		},
	})
}

// ServeScreenshotImage serves the raw image file
func (h *ScreenshotHandler) ServeScreenshotImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var screenshot models.Screenshot
	if err := database.DB.First(&screenshot, id).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Type", screenshot.MimeType)
	c.Header("Content-Disposition", "inline; filename=\""+screenshot.Filename+"\"")
	c.Data(http.StatusOK, screenshot.MimeType, screenshot.ImageData)
}
