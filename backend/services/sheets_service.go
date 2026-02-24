package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"weekly-dashboard/config"
	"weekly-dashboard/models"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// DiscoveredLayout holds auto-discovered spreadsheet column and row positions
type DiscoveredLayout struct {
	// MonthColumns: month number (1-12) → [targetIdx, laggingIdx, percentIdx, perfIdx] (0-based column indices)
	MonthColumns map[int][4]int
	// KPIRows: KPI name (lowercase, trimmed) → row numbers (1-based, may have duplicates)
	KPIRows     map[string][]int
	LastRefresh time.Time
}

// SheetsService handles Google Sheets API operations
type SheetsService struct {
	authService *AuthService
	layout      *DiscoveredLayout
	layoutMu    sync.RWMutex
}

// NewSheetsService creates a new SheetsService instance
func NewSheetsService(authService *AuthService) *SheetsService {
	return &SheetsService{
		authService: authService,
	}
}

// KPIData represents data fetched for a single KPI
type KPIData struct {
	IndicatorCode string
	Department    string
	Name          string
	Target        float64
	Performance   float64
	Percentage    float64
	IsInverse     bool
}

// monthNames maps month numbers to their full names (lowercase)
var monthNames = map[int]string{
	1: "january", 2: "february", 3: "march", 4: "april",
	5: "may", 6: "june", 7: "july", 8: "august",
	9: "september", 10: "october", 11: "november", 12: "december",
}

// nameToMonth maps lowercase month names to month numbers
var nameToMonth = map[string]int{
	"january": 1, "february": 2, "march": 3, "april": 4,
	"may": 5, "june": 6, "july": 7, "august": 8,
	"september": 9, "october": 10, "november": 11, "december": 12,
}

// formatSheetName wraps the sheet name in single quotes if it contains spaces
func formatSheetName(name string) string {
	if strings.Contains(name, " ") {
		return fmt.Sprintf("'%s'", name)
	}
	return name
}

// matchMonthFromHeader parses a header cell like "January Target", "January Lagging",
// "% January Performance", or standalone "January" and returns:
//   - month number (1-12), or 0 if no match
//   - column type: "target", "lagging", "percent", "perf", or "" if no match
func matchMonthFromHeader(header string) (int, string) {
	h := strings.TrimSpace(header)
	lower := strings.ToLower(h)

	// Pattern: "% January Performance" → percent column
	if strings.HasPrefix(lower, "% ") && strings.HasSuffix(lower, " performance") {
		mid := strings.TrimPrefix(lower, "% ")
		mid = strings.TrimSuffix(mid, " performance")
		mid = strings.TrimSpace(mid)
		if m, ok := nameToMonth[mid]; ok {
			return m, "percent"
		}
	}

	// Pattern: "January Target" → target column
	for name, m := range nameToMonth {
		if lower == name+" target" {
			return m, "target"
		}
	}

	// Pattern: "January Lagging" → lagging column
	for name, m := range nameToMonth {
		if lower == name+" lagging" {
			return m, "lagging"
		}
	}

	// Pattern: standalone "January" (exact match, no suffix) → performance/actual column
	if m, ok := nameToMonth[lower]; ok {
		return m, "perf"
	}

	return 0, ""
}

// discoverColumns reads the header row (row 1) and discovers month column positions.
// Returns map[month][4]int where indices are: [target, lagging, percent, perf] (0-based).
func discoverColumns(srv *sheets.Service, spreadsheetID, sheetName string) (map[int][4]int, error) {
	// Fetch row 1 (header) — wide range to cover all columns up to BT
	rangeStr := fmt.Sprintf("%s!1:1", formatSheetName(sheetName))
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch header row: %w", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return nil, fmt.Errorf("header row is empty")
	}

	result := make(map[int][4]int)

	log.Printf("[Discovery] Header row has %d columns", len(resp.Values[0]))

	for colIdx, cell := range resp.Values[0] {
		headerText, ok := cell.(string)
		if !ok {
			continue
		}

		month, colType := matchMonthFromHeader(headerText)
		if month == 0 {
			continue
		}

		entry := result[month]
		switch colType {
		case "target":
			entry[0] = colIdx
		case "lagging":
			entry[1] = colIdx
		case "percent":
			entry[2] = colIdx
		case "perf":
			entry[3] = colIdx
		}
		result[month] = entry
		log.Printf("[Discovery] Col %d (%s) = '%s' → month=%d type=%s", colIdx, indexToCol(colIdx), headerText, month, colType)
	}

	return result, nil
}

// discoverRows reads column C (index 2, "Leading Indicators") and matches KPI names
// from the indicator definitions. Returns map[lowercase_name][]row_numbers (1-based).
// Stores all occurrences to handle duplicate names (e.g., "Customer Satisfaction" appears twice).
func discoverRows(srv *sheets.Service, spreadsheetID, sheetName string) (map[string][]int, error) {
	// Fetch column C for all rows
	rangeStr := fmt.Sprintf("%s!C:C", formatSheetName(sheetName))
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch column C: %w", err)
	}

	result := make(map[string][]int)
	if len(resp.Values) == 0 {
		return result, nil
	}

	log.Printf("[Discovery] Column C has %d rows", len(resp.Values))

	for rowIdx, row := range resp.Values {
		if len(row) == 0 {
			continue
		}
		cellStr, ok := row[0].(string)
		if !ok || cellStr == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(cellStr))
		result[key] = append(result[key], rowIdx+1) // 1-based row number
		log.Printf("[Discovery] Row %d: '%s'", rowIdx+1, cellStr)
	}

	return result, nil
}

// DiscoverLayout performs full auto-discovery of spreadsheet layout.
func DiscoverLayout(srv *sheets.Service, spreadsheetID, sheetName string) (*DiscoveredLayout, error) {
	monthCols, err := discoverColumns(srv, spreadsheetID, sheetName)
	if err != nil {
		return nil, fmt.Errorf("column discovery failed: %w", err)
	}

	kpiRows, err := discoverRows(srv, spreadsheetID, sheetName)
	if err != nil {
		return nil, fmt.Errorf("row discovery failed: %w", err)
	}

	layout := &DiscoveredLayout{
		MonthColumns: monthCols,
		KPIRows:      kpiRows,
		LastRefresh:  time.Now(),
	}

	log.Printf("Discovered %d month columns and %d KPI rows", len(monthCols), len(kpiRows))

	// Log detailed month → column mapping
	for m := 1; m <= 12; m++ {
		if cols, ok := monthCols[m]; ok {
			log.Printf("[Discovery] Month %2d (%s): target=%s(%d) lagging=%s(%d) percent=%s(%d) perf=%s(%d)",
				m, monthNames[m],
				indexToCol(cols[0]), cols[0],
				indexToCol(cols[1]), cols[1],
				indexToCol(cols[2]), cols[2],
				indexToCol(cols[3]), cols[3])
		}
	}

	// Log discovered KPI rows
	for name, rows := range kpiRows {
		log.Printf("[Discovery] KPI '%s' → rows %v", name, rows)
	}

	return layout, nil
}

// GetLayout returns the cached layout or triggers discovery if cache is empty/expired.
func (s *SheetsService) GetLayout(ctx context.Context, user *models.User) (*DiscoveredLayout, error) {
	s.layoutMu.RLock()
	if s.layout != nil && time.Since(s.layout.LastRefresh) < 5*time.Minute {
		defer s.layoutMu.RUnlock()
		return s.layout, nil
	}
	s.layoutMu.RUnlock()

	// Need to discover or refresh
	s.layoutMu.Lock()
	defer s.layoutMu.Unlock()

	// Double-check after acquiring write lock
	if s.layout != nil && time.Since(s.layout.LastRefresh) < 5*time.Minute {
		return s.layout, nil
	}

	srv, err := s.CreateSheetsClient(ctx, user)
	if err != nil {
		// If we have a cached layout, return it despite error
		if s.layout != nil {
			log.Printf("Warning: failed to refresh layout, using cached: %v", err)
			return s.layout, nil
		}
		return nil, err
	}

	layout, err := DiscoverLayout(srv, config.AppConfig.SpreadsheetID, config.AppConfig.SheetName)
	if err != nil {
		if s.layout != nil {
			log.Printf("Warning: layout discovery failed, using cached: %v", err)
			return s.layout, nil
		}
		return nil, err
	}

	s.layout = layout
	return layout, nil
}

// InvalidateLayout clears the cached layout, forcing re-discovery on next request.
func (s *SheetsService) InvalidateLayout() {
	s.layoutMu.Lock()
	defer s.layoutMu.Unlock()
	s.layout = nil
}

// getIndicatorRow determines the row number for an indicator using discovered layout.
// Priority: SpreadsheetName match → SpreadsheetRow fallback.
// When multiple rows match the same name, picks the one closest to SpreadsheetRow.
func getIndicatorRow(layout *DiscoveredLayout, indicator models.Indicator) (int, bool) {
	if indicator.SpreadsheetName != "" {
		key := strings.ToLower(strings.TrimSpace(indicator.SpreadsheetName))
		if rows, ok := layout.KPIRows[key]; ok && len(rows) > 0 {
			if len(rows) == 1 {
				return rows[0], true
			}
			// Multiple matches — pick the one closest to SpreadsheetRow as tie-breaker
			best := rows[0]
			bestDiff := abs(rows[0] - indicator.SpreadsheetRow)
			for _, r := range rows[1:] {
				diff := abs(r - indicator.SpreadsheetRow)
				if diff < bestDiff {
					best = r
					bestDiff = diff
				}
			}
			log.Printf("KPI '%s': name '%s' matched %d rows, using row %d (closest to fallback %d)",
				indicator.Code, indicator.SpreadsheetName, len(rows), best, indicator.SpreadsheetRow)
			return best, true
		}
		log.Printf("Warning: KPI '%s' (SpreadsheetName='%s') not found in spreadsheet column C, falling back to row %d",
			indicator.Code, indicator.SpreadsheetName, indicator.SpreadsheetRow)
	}

	// Fallback to hardcoded row
	if indicator.SpreadsheetRow > 0 {
		return indicator.SpreadsheetRow, true
	}

	log.Printf("Warning: KPI '%s' has no SpreadsheetName match and no valid SpreadsheetRow, skipping", indicator.Code)
	return 0, false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// indexToCol converts a 0-based column index to Excel-style column letters.
// 0 → A, 25 → Z, 26 → AA, etc.
func indexToCol(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

// CreateSheetsClient creates a new Google Sheets client using user's token
func (s *SheetsService) CreateSheetsClient(ctx context.Context, user *models.User) (*sheets.Service, error) {
	// Get refreshed token
	token, err := s.authService.RefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// Create OAuth2 client
	oauthConfig := s.authService.GetOAuthConfig()
	client := oauthConfig.Client(ctx, token)

	// Create Sheets service
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return srv, nil
}

// FetchKPIData fetches KPI data from Google Sheets for a specific month using batch API
func (s *SheetsService) FetchKPIData(ctx context.Context, user *models.User, indicators []models.Indicator, month int) ([]KPIData, error) {
	// Get discovered layout
	layout, err := s.GetLayout(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get layout: %w", err)
	}

	// Get month column indices from discovered layout
	monthCols, ok := layout.MonthColumns[month]
	if !ok {
		return nil, fmt.Errorf("month %d not found in discovered layout", month)
	}

	targetIdx := monthCols[0]
	percentIdx := monthCols[2]
	perfIdx := monthCols[3]

	log.Printf("[FetchKPI] Month %d (%s): targetCol=%s(%d), percentCol=%s(%d), perfCol=%s(%d)",
		month, monthNames[month], indexToCol(targetIdx), targetIdx, indexToCol(percentIdx), percentIdx, indexToCol(perfIdx), perfIdx)

	// Determine the last column letter we need to fetch
	maxCol := targetIdx
	if percentIdx > maxCol {
		maxCol = percentIdx
	}
	if perfIdx > maxCol {
		maxCol = perfIdx
	}
	lastColLetter := indexToCol(maxCol)

	srv, err := s.CreateSheetsClient(ctx, user)
	if err != nil {
		return nil, err
	}

	spreadsheetID := config.AppConfig.SpreadsheetID

	// Build all ranges for batch request
	var ranges []string
	var activeIndicators []models.Indicator

	for _, indicator := range indicators {
		if !indicator.IsActive {
			continue
		}

		row, found := getIndicatorRow(layout, indicator)
		if !found {
			continue
		}

		activeIndicators = append(activeIndicators, indicator)
		rangeStr := fmt.Sprintf("%s!A%d:%s%d", formatSheetName(config.AppConfig.SheetName), row, lastColLetter, row)
		ranges = append(ranges, rangeStr)
	}

	if len(ranges) == 0 {
		return []KPIData{}, nil
	}

	log.Printf("Batch fetching %d KPIs in single API call for month %d", len(ranges), month)

	// Use BatchGet to fetch all ranges in a single API call
	resp, err := srv.Spreadsheets.Values.BatchGet(spreadsheetID).Ranges(ranges...).Do()
	if err != nil {
		log.Printf("Error in batch fetch: %v", err)
		// Return default data for all indicators on error
		var kpiDataList []KPIData
		for _, indicator := range activeIndicators {
			kpiDataList = append(kpiDataList, KPIData{
				IndicatorCode: indicator.Code,
				Department:    indicator.Department,
				Name:          indicator.Name,
				IsInverse:     indicator.IsInverse,
			})
		}
		return kpiDataList, nil
	}

	// Parse batch response
	var kpiDataList []KPIData
	for i, valueRange := range resp.ValueRanges {
		if i >= len(activeIndicators) {
			break
		}
		indicator := activeIndicators[i]
		kpiData := s.parseKPIRow(valueRange.Values, indicator, targetIdx, percentIdx, perfIdx)
		kpiDataList = append(kpiDataList, kpiData)
	}

	log.Printf("Successfully fetched %d KPIs", len(kpiDataList))
	return kpiDataList, nil
}

// FetchSingleKPIData fetches data for a single KPI
func (s *SheetsService) FetchSingleKPIData(ctx context.Context, user *models.User, indicator models.Indicator, month int) (*KPIData, error) {
	// Get discovered layout
	layout, err := s.GetLayout(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to get layout: %w", err)
	}

	monthCols, ok := layout.MonthColumns[month]
	if !ok {
		return nil, fmt.Errorf("month %d not found in discovered layout", month)
	}

	targetIdx := monthCols[0]
	percentIdx := monthCols[2]
	perfIdx := monthCols[3]

	maxCol := targetIdx
	if percentIdx > maxCol {
		maxCol = percentIdx
	}
	if perfIdx > maxCol {
		maxCol = perfIdx
	}
	lastColLetter := indexToCol(maxCol)

	row, found := getIndicatorRow(layout, indicator)
	if !found {
		return nil, fmt.Errorf("could not determine row for KPI %s", indicator.Code)
	}

	srv, err := s.CreateSheetsClient(ctx, user)
	if err != nil {
		return nil, err
	}

	spreadsheetID := config.AppConfig.SpreadsheetID
	rangeStr := fmt.Sprintf("%s!A%d:%s%d", formatSheetName(config.AppConfig.SheetName), row, lastColLetter, row)

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeStr).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	data := s.parseKPIRow(resp.Values, indicator, targetIdx, percentIdx, perfIdx)
	return &data, nil
}

// parseKPIRow parses a row of values from the spreadsheet
func (s *SheetsService) parseKPIRow(values [][]interface{}, indicator models.Indicator, targetIdx, percentIdx, perfIdx int) KPIData {
	kpiData := KPIData{
		IndicatorCode: indicator.Code,
		Department:    indicator.Department,
		Name:          indicator.Name,
		IsInverse:     indicator.IsInverse,
	}

	if len(values) == 0 || len(values[0]) == 0 {
		return kpiData
	}

	row := values[0]
	rowLen := len(row)

	// Safely access columns based on indices
	if rowLen > targetIdx {
		kpiData.Target = parseFloat(row[targetIdx])
	}
	if rowLen > percentIdx {
		kpiData.Percentage = parseFloat(row[percentIdx])
	}
	if rowLen > perfIdx {
		kpiData.Performance = parseFloat(row[perfIdx])
	}

	return kpiData
}

// colToIndex converts Excel-style column letters to 0-based index
// A -> 0, Z -> 25, AA -> 26, AB -> 27, etc.
func colToIndex(col string) int {
	result := 0
	for i := 0; i < len(col); i++ {
		result = result*26 + int(col[i]-'A'+1)
	}
	return result - 1
}

// parseNumericValue parses numeric value from sheet response
func parseNumericValue(values [][]interface{}) float64 {
	if len(values) == 0 || len(values[0]) == 0 {
		return 0
	}
	return parseFloat(values[0][0])
}

// parseFloat converts interface to float64
func parseFloat(val interface{}) float64 {
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Remove percentage signs and commas
		s := strings.ReplaceAll(v, "%", "")
		s = strings.ReplaceAll(s, ",", "")
		s = strings.TrimSpace(s)

		if s == "" || s == "-" {
			return 0
		}

		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Printf("Warning: Failed to parse float from '%s': %v", v, err)
			return 0
		}

		return f
	default:
		return 0
	}
}

// TestConnection tests if the user has access to the spreadsheet
func (s *SheetsService) TestConnection(ctx context.Context, user *models.User) error {
	srv, err := s.CreateSheetsClient(ctx, user)
	if err != nil {
		return err
	}

	spreadsheetID := config.AppConfig.SpreadsheetID

	// Try to get spreadsheet metadata
	_, err = srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("no access to spreadsheet: %w", err)
	}

	return nil
}

// GetTokenForClient creates an oauth2 token from user's stored tokens
func (s *SheetsService) GetTokenForClient(user *models.User) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.TokenExpiry,
	}
}
