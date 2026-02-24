package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"weekly-dashboard/middleware"
	"weekly-dashboard/services"

	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard endpoints
type DashboardHandler struct {
	dashboardService *services.DashboardService
	sheetsService    *services.SheetsService
}

// NewDashboardHandler creates a new DashboardHandler instance
func NewDashboardHandler(dashboardService *services.DashboardService, sheetsService *services.SheetsService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		sheetsService:    sheetsService,
	}
}

// GetDashboard returns dashboard data for a specific month
// @Summary Get dashboard data
// @Description Returns KPI dashboard data for a specific month and year
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param month query int false "Month (1-12)" default(current month)
// @Param year query int false "Year" default(current year)
// @Success 200 {object} map[string]interface{} "Dashboard data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse month and year from query params
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2020 && y <= 2100 {
			year = y
		}
	}

	log.Printf("Fetching dashboard for user %s, month=%d, year=%d", user.Email, month, year)

	// Invalidate layout cache if refresh is requested
	if c.Query("refresh") == "true" {
		log.Printf("Force refresh requested, invalidating layout cache")
		h.sheetsService.InvalidateLayout()
	}

	// Test spreadsheet access first
	if err := h.sheetsService.TestConnection(c.Request.Context(), user); err != nil {
		log.Printf("User %s does not have access to spreadsheet: %v", user.Email, err)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "You do not have access to the performance spreadsheet. Please contact your administrator.",
		})
		return
	}

	// Get dashboard data
	dashboardData, err := h.dashboardService.GetDashboardData(c.Request.Context(), user, month, year)
	if err != nil {
		log.Printf("Failed to get dashboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch dashboard data. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboardData,
	})
}

// GetAvailableMonths returns list of available months
// @Summary Get available months
// @Description Returns list of months with available data
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Available months"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/months [get]
func (h *DashboardHandler) GetAvailableMonths(c *gin.Context) {
	months := h.dashboardService.GetAvailableMonths()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    months,
	})
}

// CompareDashboard returns comparison data between periods
// @Summary Compare dashboard data
// @Description Returns dashboard data with comparison to another period
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param month query int true "Month (1-12)"
// @Param year query int true "Year"
// @Param compareWith query string false "Comparison period: previous_month or previous_year" default(previous_month)
// @Success 200 {object} map[string]interface{} "Comparison data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /api/v1/dashboard/compare [get]
func (h *DashboardHandler) CompareDashboard(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse parameters
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	compareWith := c.DefaultQuery("compareWith", "previous_month")

	if monthStr == "" || yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Month and year are required",
		})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid month value",
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

	// Get current period data
	currentData, err := h.dashboardService.GetDashboardData(c.Request.Context(), user, month, year)
	if err != nil {
		log.Printf("Failed to get current dashboard data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch dashboard data",
		})
		return
	}

	// Calculate comparison period
	compareMonth := month
	compareYear := year

	switch compareWith {
	case "previous_month":
		compareMonth--
		if compareMonth < 1 {
			compareMonth = 12
			compareYear--
		}
	case "previous_year":
		compareYear--
	}

	// Get comparison period data
	comparisonData, err := h.dashboardService.GetDashboardData(c.Request.Context(), user, compareMonth, compareYear)
	if err != nil {
		log.Printf("Failed to get comparison dashboard data: %v", err)
		// Continue without comparison data
		comparisonData = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"current":    currentData,
			"comparison": comparisonData,
			"compareWith": gin.H{
				"type":  compareWith,
				"month": compareMonth,
				"year":  compareYear,
			},
		},
	})
}

// SaveSnapshot saves current dashboard data for week-over-week comparison
// @Summary Save weekly snapshot
// @Description Saves current dashboard data to database for future WoW comparison
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query int false "Month (1-12)" default(current month)
// @Param year query int false "Year" default(current year)
// @Param week query int false "Week number (1-5)" default(1)
// @Success 200 {object} map[string]interface{} "Snapshot saved"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dashboard/snapshot [post]
func (h *DashboardHandler) SaveSnapshot(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Parse month and year from query params
	now := time.Now()
	month := int(now.Month())
	year := now.Year()
	weekNumber := 1 // Default to week 1

	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2020 && y <= 2100 {
			year = y
		}
	}

	// Parse week number from query params
	if weekStr := c.Query("week"); weekStr != "" {
		if w, err := strconv.Atoi(weekStr); err == nil && w >= 1 && w <= 5 {
			weekNumber = w
		}
	}

	log.Printf("Saving snapshot for user %s, month=%d, year=%d, week=%d", user.Email, month, year, weekNumber)

	// Get current dashboard data
	dashboardData, err := h.dashboardService.GetDashboardData(c.Request.Context(), user, month, year)
	if err != nil {
		log.Printf("Failed to get dashboard data for snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch current dashboard data",
		})
		return
	}

	// Save snapshot (will delete existing data for same week first)
	if err := h.dashboardService.SaveSnapshot(dashboardData.Indicators, month, year, weekNumber); err != nil {
		log.Printf("Failed to save snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save snapshot",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Snapshot saved successfully",
		"data": gin.H{
			"month":       month,
			"year":        year,
			"week_number": weekNumber,
			"saved_at":    now.Format(time.RFC3339),
		},
	})
}

// GetSnapshotsByMonth returns all snapshots for a month grouped by indicator
// @Summary Get monthly snapshots
// @Description Returns all weekly snapshots for a month grouped by indicator
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param month query int true "Month (1-12)"
// @Param year query int true "Year"
// @Success 200 {object} map[string]interface{} "Monthly snapshots"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dashboard/snapshots [get]
func (h *DashboardHandler) GetSnapshotsByMonth(c *gin.Context) {
	user, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	// Test spreadsheet access first
	if err := h.sheetsService.TestConnection(c.Request.Context(), user); err != nil {
		log.Printf("User %s does not have access to spreadsheet: %v", user.Email, err)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "You do not have access to the performance spreadsheet. Please contact your administrator.",
		})
		return
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if monthStr := c.Query("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil && y >= 2020 && y <= 2100 {
			year = y
		}
	}

	data, err := h.dashboardService.GetSnapshotsByMonth(month, year)
	if err != nil {
		log.Printf("Failed to get monthly snapshots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch snapshot data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// DeleteSnapshot deletes all snapshot data and screenshot for a specific week
// @Summary Delete weekly snapshot
// @Description Deletes all snapshot data and screenshot for a specific month/year/week
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param month query int true "Month (1-12)"
// @Param year query int true "Year"
// @Param week query int true "Week number (1-5)"
// @Success 200 {object} map[string]interface{} "Snapshot deleted"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/dashboard/snapshot [delete]
func (h *DashboardHandler) DeleteSnapshot(c *gin.Context) {
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
	weekStr := c.Query("week")

	if monthStr == "" || yearStr == "" || weekStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Month, year, and week are required",
		})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid month value",
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
			"error":   "Invalid week value",
		})
		return
	}

	log.Printf("Deleting snapshot for month=%d, year=%d, week=%d", month, year, week)

	if err := h.dashboardService.DeleteSnapshotWeek(month, year, week); err != nil {
		log.Printf("Failed to delete snapshot: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete snapshot",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Snapshot deleted successfully",
	})
}

// HealthCheck returns API health status
// @Summary Health check
// @Description Returns API health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Router /api/v1/health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "healthy",
		"time":    time.Now().Format(time.RFC3339),
	})
}
