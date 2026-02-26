package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"weekly-dashboard/config"
	"weekly-dashboard/database"
	"weekly-dashboard/models"
)

// DashboardService handles dashboard business logic
type DashboardService struct {
	sheetsService *SheetsService
}

// NewDashboardService creates a new DashboardService instance
func NewDashboardService(sheetsService *SheetsService) *DashboardService {
	return &DashboardService{
		sheetsService: sheetsService,
	}
}

// IndicatorResponse represents a single KPI indicator in the response
type IndicatorResponse struct {
	Code             string  `json:"code"`
	Department       string  `json:"department"`
	Name             string  `json:"name"`
	Target           float64 `json:"target"`
	Performance      float64 `json:"performance"`
	Percentage       float64 `json:"percentage"`
	Status           string  `json:"status"` // "green", "yellow", "red"
	IsInverse        bool    `json:"is_inverse"`
	WowChange        float64 `json:"wow_change"`
	WowDirection     string  `json:"wow_direction"` // "up", "down", "neutral"
	ExpectedProgress float64 `json:"expected_progress"`
	Variance         float64 `json:"variance"`        // % variance = ((actual - expected) / expected) * 100
	ScheduleStatus   string  `json:"schedule_status"` // "ahead", "on_schedule", "behind"
}

// OverallPerformance represents the overall dashboard performance
type OverallPerformance struct {
	Percentage  float64 `json:"percentage"`
	Status      string  `json:"status"`
	GreenCount  int     `json:"green_count"`
	YellowCount int     `json:"yellow_count"`
	RedCount    int     `json:"red_count"`
}

// WeeklyTrend represents week-over-week trend
type WeeklyTrend struct {
	Change           float64 `json:"change"`
	Direction        string  `json:"direction"`
	GreenCountChange int     `json:"green_count_change"`
}

// Period represents the selected period
type Period struct {
	Month     int    `json:"month"`
	Year      int    `json:"year"`
	MonthName string `json:"month_name"`
}

// ScheduleSummary represents overall schedule status counts
type ScheduleSummary struct {
	AheadCount      int `json:"ahead_count"`
	OnScheduleCount int `json:"on_schedule_count"`
	BehindCount     int `json:"behind_count"`
}

// DashboardResponse represents the complete dashboard response
type DashboardResponse struct {
	Period             Period              `json:"period"`
	OverallPerformance OverallPerformance  `json:"overall_performance"`
	WeeklyTrend        WeeklyTrend         `json:"weekly_trend"`
	ScheduleSummary    ScheduleSummary     `json:"schedule_summary"`
	Indicators         []IndicatorResponse `json:"indicators"`
	LastUpdated        time.Time           `json:"last_updated"`
}

// MonthOption represents an available month option
type MonthOption struct {
	Month   int    `json:"month"`
	Year    int    `json:"year"`
	Label   string `json:"label"`
	HasData bool   `json:"has_data"`
}

// MonthsResponse represents available months response
type MonthsResponse struct {
	AvailableMonths []MonthOption `json:"available_months"`
	CurrentMonth    struct {
		Month int `json:"month"`
		Year  int `json:"year"`
	} `json:"current_month"`
}

// GetDashboardData fetches and calculates dashboard data
func (s *DashboardService) GetDashboardData(ctx context.Context, user *models.User, month, year int) (*DashboardResponse, error) {
	// Get all active indicators
	var indicators []models.Indicator
	if err := database.DB.Where("is_active = ?", true).Order("display_order").Find(&indicators).Error; err != nil {
		return nil, err
	}

	// Only fetch from spreadsheet if year matches configured spreadsheet year
	var kpiDataList []KPIData
	if year == config.AppConfig.SpreadsheetYear {
		var err error
		kpiDataList, err = s.sheetsService.FetchKPIData(ctx, user, indicators, month)
		if err != nil {
			log.Printf("Warning: Error fetching sheet data: %v", err)
			// Continue with empty data
		}
	} else {
		log.Printf("Skipping spreadsheet fetch: requested year %d != configured year %d", year, config.AppConfig.SpreadsheetYear)
	}

	// Get previous week's data for WoW comparison
	prevSnapshots := s.getPreviousWeekSnapshots(month, year)

	// Build indicator responses
	var indicatorResponses []IndicatorResponse
	greenCount, yellowCount, redCount := 0, 0, 0

	for _, kpiData := range kpiDataList {
		// Calculate percentage: always performance / target * 100
		// For inverse metrics (Non Billable Cost, Turn Over), the color thresholds
		// are reversed in calculateStatus, but the percentage calculation is the same.
		// e.g. Non Billable Cost: actual 8M / target max 10M = 80% (good, under budget)
		// e.g. Turn Over: actual 3 / target max 5 = 60% (good, low turnover)
		var calculatedPercentage float64
		if kpiData.Target != 0 {
			calculatedPercentage = (kpiData.Performance / kpiData.Target) * 100
		} else {
			calculatedPercentage = 0
		}

		// Apply capping logic — cap at 999% to avoid display issues
		if calculatedPercentage > 999 {
			calculatedPercentage = 999
		}

		status := calculateStatus(calculatedPercentage, kpiData.IsInverse)

		switch status {
		case "green", "supergreen":
			greenCount++
		case "yellow":
			yellowCount++
		case "red":
			redCount++
		}

		// Calculate WoW change
		wowChange, wowDirection := s.calculateWoWChange(kpiData.IndicatorCode, calculatedPercentage, prevSnapshots)

		// Calculate expected progress and variance
		expectedProgress, variance, scheduleStatus := calculateVariance(kpiData.Target, kpiData.Performance, kpiData.IsInverse, month, year)

		indicatorResponses = append(indicatorResponses, IndicatorResponse{
			Code:             kpiData.IndicatorCode,
			Department:       kpiData.Department,
			Name:             kpiData.Name,
			Target:           kpiData.Target,
			Performance:      kpiData.Performance,
			Percentage:       calculatedPercentage,
			Status:           status,
			IsInverse:        kpiData.IsInverse,
			WowChange:        wowChange,
			WowDirection:     wowDirection,
			ExpectedProgress: expectedProgress,
			Variance:         variance,
			ScheduleStatus:   scheduleStatus,
		})
	}

	// Calculate overall performance
	totalIndicators := greenCount + yellowCount + redCount
	overallPercentage := 0.0
	if totalIndicators > 0 {
		overallPercentage = (float64(greenCount) / float64(totalIndicators)) * 100
	}

	overallStatus := "red"
	if overallPercentage > 85 {
		overallStatus = "green"
	} else if overallPercentage > 55 {
		overallStatus = "yellow"
	}

	// Calculate weekly trend (difference between current and previous overall percentage)
	weeklyChange, weeklyDirection, prevGreenCount := s.calculateWeeklyTrend(month, year, overallPercentage)
	greenCountChange := greenCount - prevGreenCount

	// Calculate schedule summary
	aheadCount, onScheduleCount, behindCount := 0, 0, 0
	for _, ind := range indicatorResponses {
		switch ind.ScheduleStatus {
		case "ahead":
			aheadCount++
		case "on_schedule":
			onScheduleCount++
		case "behind":
			behindCount++
		}
	}

	response := &DashboardResponse{
		Period: Period{
			Month:     month,
			Year:      year,
			MonthName: getMonthName(month),
		},
		OverallPerformance: OverallPerformance{
			Percentage:  overallPercentage,
			Status:      overallStatus,
			GreenCount:  greenCount,
			YellowCount: yellowCount,
			RedCount:    redCount,
		},
		WeeklyTrend: WeeklyTrend{
			Change:           weeklyChange,
			Direction:        weeklyDirection,
			GreenCountChange: greenCountChange,
		},
		ScheduleSummary: ScheduleSummary{
			AheadCount:      aheadCount,
			OnScheduleCount: onScheduleCount,
			BehindCount:     behindCount,
		},
		Indicators:  indicatorResponses,
		LastUpdated: time.Now(),
	}

	return response, nil
}

// GetAvailableMonths returns list of available months for the dashboard
func (s *DashboardService) GetAvailableMonths() *MonthsResponse {
	now := time.Now()
	currentMonth := int(now.Month())
	spreadsheetYear := config.AppConfig.SpreadsheetYear

	// Generate months for the configured spreadsheet year
	var months []MonthOption

	for m := 1; m <= 12; m++ {
		months = append(months, MonthOption{
			Month:   m,
			Year:    spreadsheetYear,
			Label:   getMonthName(m) + " " + strconv.Itoa(spreadsheetYear),
			HasData: s.hasDataForMonth(m, spreadsheetYear),
		})
	}

	// Determine current month: if we're in the spreadsheet year, use actual month
	// Otherwise default to January
	displayMonth := 1
	if now.Year() == spreadsheetYear {
		displayMonth = currentMonth
	}

	return &MonthsResponse{
		AvailableMonths: months,
		CurrentMonth: struct {
			Month int `json:"month"`
			Year  int `json:"year"`
		}{
			Month: displayMonth,
			Year:  spreadsheetYear,
		},
	}
}

// getPreviousWeekSnapshots gets snapshots from previous week
func (s *DashboardService) getPreviousWeekSnapshots(month, year int) map[string]float64 {
	snapshots := make(map[string]float64)

	var records []models.WeeklySnapshot
	database.DB.Where("month = ? AND year = ?", month, year).
		Order("snapshot_date desc").
		Find(&records)

	// Get the most recent snapshot for each indicator
	seen := make(map[string]bool)
	for _, record := range records {
		if !seen[record.IndicatorID] {
			snapshots[record.IndicatorID] = record.Percentage
			seen[record.IndicatorID] = true
		}
	}

	return snapshots
}

// calculateWoWChange calculates week-over-week change
func (s *DashboardService) calculateWoWChange(indicatorCode string, currentPercentage float64, prevSnapshots map[string]float64) (float64, string) {
	prevPercentage, exists := prevSnapshots[indicatorCode]
	if !exists || prevPercentage == 0 {
		return 0, "neutral"
	}

	change := currentPercentage - prevPercentage
	direction := "neutral"
	if change > 0.5 {
		direction = "up"
	} else if change < -0.5 {
		direction = "down"
	}

	return change, direction
}

// calculateWeeklyTrendFromIndicators calculates weekly trend as average of all indicator WoW changes
func (s *DashboardService) calculateWeeklyTrendFromIndicators(indicators []IndicatorResponse) (float64, string) {
	if len(indicators) == 0 {
		return 0, "neutral"
	}

	var totalChange float64
	var validCount int

	for _, indicator := range indicators {
		// Only count indicators that have WoW data
		if indicator.WowDirection != "neutral" || indicator.WowChange != 0 {
			totalChange += indicator.WowChange
			validCount++
		} else {
			// Include all indicators in average
			validCount++
		}
	}

	if validCount == 0 {
		return 0, "neutral"
	}

	averageChange := totalChange / float64(validCount)

	direction := "neutral"
	if averageChange > 0.5 {
		direction = "up"
	} else if averageChange < -0.5 {
		direction = "down"
	}

	return averageChange, direction
}

// getPreviousGreenCount gets the green count from the last snapshot

// calculateWeeklyTrend calculates overall weekly trend (legacy, kept for reference)
func (s *DashboardService) calculateWeeklyTrend(month, year int, currentPercentage float64) (float64, string, int) {
	// Get last week's overall performance from snapshots
	var lastSnapshot models.WeeklySnapshot
	result := database.DB.Where("month = ? AND year = ?", month, year).
		Order("snapshot_date desc").
		First(&lastSnapshot)

	if result.Error != nil || result.RowsAffected == 0 {
		return 0, "neutral", 0
	}

	// Calculate total green count from that week's snapshots
	var snapshots []models.WeeklySnapshot
	database.DB.Where("snapshot_date = ?", lastSnapshot.SnapshotDate).Find(&snapshots)

	if len(snapshots) == 0 {
		return 0, "neutral", 0
	}

	// Get all indicators to check for inverse metrics
	var indicators []models.Indicator
	database.DB.Find(&indicators)
	inverseMap := make(map[string]bool)
	for _, ind := range indicators {
		inverseMap[ind.Code] = ind.IsInverse
	}

	// Count green indicators from previous snapshot
	prevGreenCount := 0
	for _, snap := range snapshots {
		isInverse := inverseMap[snap.IndicatorID]
		status := calculateStatus(snap.Percentage, isInverse)
		if status == "green" || status == "supergreen" {
			prevGreenCount++
		}
	}

	prevOverallPercentage := (float64(prevGreenCount) / float64(len(snapshots))) * 100
	change := currentPercentage - prevOverallPercentage

	direction := "neutral"
	if change > 0.5 {
		direction = "up"
	} else if change < -0.5 {
		direction = "down"
	}

	return change, direction, prevGreenCount
}

// hasDataForMonth checks if there's data available for a given month
func (s *DashboardService) hasDataForMonth(month, year int) bool {
	var count int64
	database.DB.Model(&models.WeeklySnapshot{}).
		Where("month = ? AND year = ?", month, year).
		Count(&count)
	return count > 0
}

// calculateStatus determines the status color based on percentage
// Normal metrics (higher is better): >100% supergreen, 85-100% green, 55-85% yellow, <55% red
// Inverse metrics (lower is better, e.g. Non Billable Cost, Turn Over):
//
//	Percentage = actual/target, so >100% means EXCEEDING max target = BAD
//	<55% supergreen, 55-85% green, 85-100% yellow, >=100% red
func calculateStatus(percentage float64, isInverse bool) string {
	if isInverse {
		// Inverse: lower percentage = better (under max target)
		if percentage >= 100 {
			return "red"
		} else if percentage >= 85 {
			return "yellow"
		} else if percentage >= 55 {
			return "green"
		}
		return "supergreen"
	}

	// Normal: higher percentage = better
	if percentage > 100 {
		return "supergreen"
	} else if percentage > 85 {
		return "green"
	} else if percentage > 55 {
		return "yellow"
	}
	return "red"
}

// calculateVariance calculates expected progress, % variance, and schedule status
// Expected progress is prorated linearly: target × (currentDay / totalDaysInMonth)
func calculateVariance(target, actual float64, isInverse bool, month, year int) (expectedProgress, variance float64, scheduleStatus string) {
	if target == 0 {
		return 0, 0, "on_schedule"
	}

	// Calculate day progress within the month
	now := time.Now()
	currentDay := now.Day()
	totalDays := daysInMonth(month, year)

	// Progress ratio (what fraction of the month has passed)
	progressRatio := float64(currentDay) / float64(totalDays)

	// Expected progress = target prorated by days elapsed
	expectedProgress = target * progressRatio

	// Avoid division by zero
	if expectedProgress == 0 {
		return expectedProgress, 0, "on_schedule"
	}

	// Calculate variance
	if isInverse {
		// For inverse metrics (lower is better): being below expected is good
		// Variance positive = ahead (actual < expected means doing well)
		variance = ((expectedProgress - actual) / expectedProgress) * 100
	} else {
		// For normal metrics (higher is better): being above expected is good
		variance = ((actual - expectedProgress) / expectedProgress) * 100
	}

	// Round to 1 decimal
	variance = math.Round(variance*10) / 10
	expectedProgress = math.Round(expectedProgress*100) / 100

	// Determine schedule status with ±5% threshold
	if variance > 5 {
		scheduleStatus = "ahead"
	} else if variance < -5 {
		scheduleStatus = "behind"
	} else {
		scheduleStatus = "on_schedule"
	}

	return expectedProgress, variance, scheduleStatus
}

// daysInMonth returns the total number of days in a given month/year
func daysInMonth(month, year int) int {
	// Use the trick: day 0 of the next month = last day of this month
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// getMonthName returns the English name of a month
func getMonthName(month int) string {
	months := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	if month >= 1 && month <= 12 {
		return months[month]
	}
	return ""
}

// SaveSnapshot saves a weekly snapshot to the database (upsert - deletes existing week data first)
func (s *DashboardService) SaveSnapshot(indicators []IndicatorResponse, month, year, weekNumber int) error {
	snapshotDate := time.Now()

	// Delete existing snapshots for this month/year/week (upsert behavior)
	deleteResult := database.DB.Where("month = ? AND year = ? AND week_number = ?", month, year, weekNumber).
		Delete(&models.WeeklySnapshot{})
	if deleteResult.Error != nil {
		log.Printf("Warning: Failed to delete existing snapshots: %v", deleteResult.Error)
	} else if deleteResult.RowsAffected > 0 {
		log.Printf("Deleted %d existing snapshots for month=%d, year=%d, week=%d", deleteResult.RowsAffected, month, year, weekNumber)
	}

	for _, indicator := range indicators {
		snapshot := models.WeeklySnapshot{
			IndicatorID:      indicator.Code,
			Department:       indicator.Department,
			IndicatorName:    indicator.Name,
			TargetValue:      indicator.Target,
			PerformanceValue: indicator.Performance,
			Percentage:       indicator.Percentage,
			SnapshotDate:     snapshotDate,
			Month:            month,
			WeekNumber:       weekNumber,
			Year:             year,
		}

		if err := database.DB.Create(&snapshot).Error; err != nil {
			log.Printf("Failed to save snapshot for %s: %v", indicator.Code, err)
			return err
		}
	}

	log.Printf("Saved %d snapshots for month %d, week %d, year %d", len(indicators), month, weekNumber, year)
	return nil
}

// SnapshotWeekData represents snapshot data for a single week
type SnapshotWeekData struct {
	Week       int     `json:"week"`
	Percentage float64 `json:"percentage"`
}

// IndicatorSnapshots represents an indicator with its weekly snapshot data
type IndicatorSnapshots struct {
	Code       string             `json:"code"`
	Department string             `json:"department"`
	Name       string             `json:"name"`
	Weeks      []SnapshotWeekData `json:"weeks"`
}

// MonthlySnapshotsResponse represents the response for monthly snapshots
type MonthlySnapshotsResponse struct {
	Indicators     []IndicatorSnapshots `json:"indicators"`
	AvailableWeeks []int                `json:"available_weeks"`
	Month          int                  `json:"month"`
	Year           int                  `json:"year"`
	MonthName      string               `json:"month_name"`
}

// GetSnapshotsByMonth returns all snapshots for a month grouped by indicator
func (s *DashboardService) GetSnapshotsByMonth(month, year int) (*MonthlySnapshotsResponse, error) {
	var snapshots []models.WeeklySnapshot
	result := database.DB.Where("month = ? AND year = ? AND week_number >= 1 AND week_number <= 5", month, year).
		Order("indicator_id, week_number").
		Find(&snapshots)

	if result.Error != nil {
		return nil, result.Error
	}

	// Group by indicator, deduplicate by week_number per indicator
	indicatorMap := make(map[string]*IndicatorSnapshots)
	indicatorOrder := []string{}
	weekSet := make(map[int]bool)
	// Track which (indicator, week) pairs we've already added
	seenWeeks := make(map[string]bool)

	for _, snap := range snapshots {
		weekSet[snap.WeekNumber] = true

		if _, exists := indicatorMap[snap.IndicatorID]; !exists {
			indicatorMap[snap.IndicatorID] = &IndicatorSnapshots{
				Code:       snap.IndicatorID,
				Department: snap.Department,
				Name:       snap.IndicatorName,
				Weeks:      []SnapshotWeekData{},
			}
			indicatorOrder = append(indicatorOrder, snap.IndicatorID)
		}

		// Deduplicate: only add each week once per indicator
		key := fmt.Sprintf("%s-%d", snap.IndicatorID, snap.WeekNumber)
		if !seenWeeks[key] {
			seenWeeks[key] = true
			indicatorMap[snap.IndicatorID].Weeks = append(indicatorMap[snap.IndicatorID].Weeks, SnapshotWeekData{
				Week:       snap.WeekNumber,
				Percentage: snap.Percentage,
			})
		}
	}

	// Build ordered indicators list
	var indicators []IndicatorSnapshots
	for _, code := range indicatorOrder {
		indicators = append(indicators, *indicatorMap[code])
	}

	// Build available weeks sorted
	var availableWeeks []int
	for w := 1; w <= 5; w++ {
		if weekSet[w] {
			availableWeeks = append(availableWeeks, w)
		}
	}

	return &MonthlySnapshotsResponse{
		Indicators:     indicators,
		AvailableWeeks: availableWeeks,
		Month:          month,
		Year:           year,
		MonthName:      getMonthName(month),
	}, nil
}

// DeleteSnapshotWeek deletes all snapshots and screenshots for a specific week
func (s *DashboardService) DeleteSnapshotWeek(month, year, week int) error {
	// Delete weekly snapshots
	snapResult := database.DB.Where("month = ? AND year = ? AND week_number = ?", month, year, week).
		Delete(&models.WeeklySnapshot{})
	if snapResult.Error != nil {
		log.Printf("Failed to delete snapshots: %v", snapResult.Error)
		return snapResult.Error
	}
	log.Printf("Deleted %d snapshot records for month=%d, year=%d, week=%d", snapResult.RowsAffected, month, year, week)

	// Delete corresponding screenshot
	screenResult := database.DB.Where("month = ? AND year = ? AND week = ?", month, year, week).
		Delete(&models.Screenshot{})
	if screenResult.Error != nil {
		log.Printf("Warning: Failed to delete screenshot: %v", screenResult.Error)
		// Don't fail the whole operation if screenshot delete fails
	} else if screenResult.RowsAffected > 0 {
		log.Printf("Deleted screenshot for month=%d, year=%d, week=%d", month, year, week)
	}

	return nil
}
