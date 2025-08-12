package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CronExpression represents a parsed cron expression
type CronExpression struct {
	Minute     []int // 0-59
	Hour       []int // 0-23
	DayOfMonth []int // 1-31
	Month      []int // 1-12
	DayOfWeek  []int // 0-6 (Sunday=0)
}

// ParseCron parses a cron expression string into a CronExpression
func ParseCron(expr string) (*CronExpression, error) {
	// Handle shortcuts first
	if shortcut, ok := expandShortcut(expr); ok {
		expr = shortcut
	}

	// Split the expression into fields
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields))
	}

	cronExpr := &CronExpression{}

	// Parse each field
	var err error
	cronExpr.Minute, err = parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("invalid minute field: %w", err)
	}

	cronExpr.Hour, err = parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("invalid hour field: %w", err)
	}

	cronExpr.DayOfMonth, err = parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("invalid day of month field: %w", err)
	}

	cronExpr.Month, err = parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("invalid month field: %w", err)
	}

	cronExpr.DayOfWeek, err = parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("invalid day of week field: %w", err)
	}

	return cronExpr, nil
}

// expandShortcut expands cron shortcuts like @daily, @hourly, etc.
func expandShortcut(expr string) (string, bool) {
	shortcuts := map[string]string{
		"@yearly":   "0 0 1 1 *",
		"@annually": "0 0 1 1 *",
		"@monthly":  "0 0 1 * *",
		"@weekly":   "0 0 * * 0",
		"@daily":    "0 0 * * *",
		"@hourly":   "0 * * * *",
	}

	if expanded, exists := shortcuts[expr]; exists {
		return expanded, true
	}
	return expr, false
}

// parseField parses a single cron field (minute, hour, etc.)
func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		// Return all valid values for this field
		result := make([]int, max-min+1)
		for i := min; i <= max; i++ {
			result[i-min] = i
		}
		return result, nil
	}

	// Handle step values (e.g., */15)
	if strings.Contains(field, "/") {
		return parseStepField(field, min, max)
	}

	// Handle ranges (e.g., 1-5)
	if strings.Contains(field, "-") {
		return parseRangeField(field, min, max)
	}

	// Handle lists (e.g., 1,3,5)
	if strings.Contains(field, ",") {
		return parseListField(field, min, max)
	}

	// Handle single value
	value, err := strconv.Atoi(field)
	if err != nil {
		return nil, fmt.Errorf("invalid value: %s", field)
	}

	if value < min || value > max {
		return nil, fmt.Errorf("value %d out of range [%d-%d]", value, min, max)
	}

	return []int{value}, nil
}

// parseStepField parses step fields like */15 or 2-10/3
func parseStepField(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid step format: %s", field)
	}

	step, err := strconv.Atoi(parts[1])
	if err != nil || step <= 0 {
		return nil, fmt.Errorf("invalid step value: %s", parts[1])
	}

	var baseValues []int
	if parts[0] == "*" {
		// Generate all values from min to max
		for i := min; i <= max; i++ {
			baseValues = append(baseValues, i)
		}
	} else {
		// Parse the base range
		baseValues, err = parseField(parts[0], min, max)
		if err != nil {
			return nil, err
		}
	}

	// Apply step
	var result []int
	for i, value := range baseValues {
		if i%step == 0 {
			result = append(result, value)
		}
	}

	return result, nil
}

// parseRangeField parses range fields like 1-5
func parseRangeField(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", field)
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %s", parts[0])
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %s", parts[1])
	}

	if start < min || start > max || end < min || end > max {
		return nil, fmt.Errorf("range [%d-%d] out of bounds [%d-%d]", start, end, min, max)
	}

	if start > end {
		return nil, fmt.Errorf("invalid range: start %d > end %d", start, end)
	}

	var result []int
	for i := start; i <= end; i++ {
		result = append(result, i)
	}

	return result, nil
}

// parseListField parses list fields like 1,3,5
func parseListField(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, ",")
	var result []int

	for _, part := range parts {
		values, err := parseField(strings.TrimSpace(part), min, max)
		if err != nil {
			return nil, err
		}
		result = append(result, values...)
	}

	return result, nil
}

// NextExecution calculates the next execution time after the given time
func (c *CronExpression) NextExecution(from time.Time) time.Time {
	// Start from the next minute (truncate seconds and add 1 minute)
	next := from.Truncate(time.Minute).Add(time.Minute)

	// Find the next valid time
	for attempts := 0; attempts < 366*24*60; attempts++ { // Prevent infinite loops
		if c.matches(next) {
			return next
		}
		next = next.Add(time.Minute)
	}

	// Fallback - should not happen with valid cron expressions
	return next
}

// matches checks if the given time matches the cron expression
func (c *CronExpression) matches(t time.Time) bool {
	minute := t.Minute()
	hour := t.Hour()
	day := t.Day()
	month := int(t.Month())
	weekday := int(t.Weekday())

	return contains(c.Minute, minute) &&
		contains(c.Hour, hour) &&
		contains(c.DayOfMonth, day) &&
		contains(c.Month, month) &&
		contains(c.DayOfWeek, weekday)
}

// contains checks if a slice contains a value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
