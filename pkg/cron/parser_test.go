package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronExpressionParsing(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		expectError bool
		expected    *CronExpression
	}{
		{
			name:       "basic every minute",
			expression: "* * * * *",
			expected: &CronExpression{
				Minute:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59},
				Hour:       []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
				DayOfMonth: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				Month:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				DayOfWeek:  []int{0, 1, 2, 3, 4, 5, 6},
			},
		},
		{
			name:       "specific time - 9 AM weekdays",
			expression: "0 9 * * 1-5",
			expected: &CronExpression{
				Minute:     []int{0},
				Hour:       []int{9},
				DayOfMonth: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				Month:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				DayOfWeek:  []int{1, 2, 3, 4, 5},
			},
		},
		{
			name:       "every 15 minutes",
			expression: "*/15 * * * *",
			expected: &CronExpression{
				Minute:     []int{0, 15, 30, 45},
				Hour:       []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
				DayOfMonth: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				Month:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				DayOfWeek:  []int{0, 1, 2, 3, 4, 5, 6},
			},
		},
		{
			name:       "specific days and times",
			expression: "30 14 * * 1,3,5",
			expected: &CronExpression{
				Minute:     []int{30},
				Hour:       []int{14},
				DayOfMonth: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
				Month:      []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
				DayOfWeek:  []int{1, 3, 5},
			},
		},
		{
			name:        "invalid expression - too few fields",
			expression:  "* * *",
			expectError: true,
		},
		{
			name:        "invalid expression - too many fields",
			expression:  "* * * * * * *",
			expectError: true,
		},
		{
			name:        "invalid minute value",
			expression:  "60 * * * *",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCron(tt.expression)

			if tt.expectError {
				assert.Error(t, err, "Expected error for invalid expression")
				assert.Nil(t, result, "Result should be nil for invalid expression")
			} else {
				require.NoError(t, err, "Should parse valid cron expression")
				require.NotNil(t, result, "Result should not be nil")

				assert.Equal(t, tt.expected.Minute, result.Minute, "Minute field should match")
				assert.Equal(t, tt.expected.Hour, result.Hour, "Hour field should match")
				assert.Equal(t, tt.expected.DayOfMonth, result.DayOfMonth, "DayOfMonth field should match")
				assert.Equal(t, tt.expected.Month, result.Month, "Month field should match")
				assert.Equal(t, tt.expected.DayOfWeek, result.DayOfWeek, "DayOfWeek field should match")
			}
		})
	}
}

func TestCronShortcuts(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{
			name:       "@daily should expand to midnight",
			expression: "@daily",
			expected:   "0 0 * * *",
		},
		{
			name:       "@hourly should expand to top of hour",
			expression: "@hourly",
			expected:   "0 * * * *",
		},
		{
			name:       "@weekly should expand to Sunday midnight",
			expression: "@weekly",
			expected:   "0 0 * * 0",
		},
		{
			name:       "@monthly should expand to first of month",
			expression: "@monthly",
			expected:   "0 0 1 * *",
		},
		{
			name:       "@yearly should expand to January 1st",
			expression: "@yearly",
			expected:   "0 0 1 1 *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the shortcut
			shortcutResult, err := ParseCron(tt.expression)
			require.NoError(t, err, "Should parse cron shortcut")

			// Parse the expected expansion
			expectedResult, err := ParseCron(tt.expected)
			require.NoError(t, err, "Should parse expected expansion")

			// They should be equivalent
			assert.Equal(t, expectedResult.Minute, shortcutResult.Minute, "Minute should match expansion")
			assert.Equal(t, expectedResult.Hour, shortcutResult.Hour, "Hour should match expansion")
			assert.Equal(t, expectedResult.DayOfMonth, shortcutResult.DayOfMonth, "DayOfMonth should match expansion")
			assert.Equal(t, expectedResult.Month, shortcutResult.Month, "Month should match expansion")
			assert.Equal(t, expectedResult.DayOfWeek, shortcutResult.DayOfWeek, "DayOfWeek should match expansion")
		})
	}
}

func TestNextExecution(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		from       time.Time
		expected   time.Time
	}{
		{
			name:       "next minute",
			expression: "* * * * *",
			from:       time.Date(2024, 1, 1, 12, 30, 30, 0, time.UTC),
			expected:   time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC),
		},
		{
			name:       "daily at 9 AM",
			expression: "0 9 * * *",
			from:       time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			expected:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			name:       "daily at 9 AM - after time",
			expression: "0 9 * * *",
			from:       time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expected:   time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
		},
		{
			name:       "weekdays at 9 AM - Monday",
			expression: "0 9 * * 1-5",
			from:       time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC), // Monday
			expected:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			name:       "weekdays at 9 AM - Saturday",
			expression: "0 9 * * 1-5",
			from:       time.Date(2024, 1, 6, 8, 0, 0, 0, time.UTC), // Saturday
			expected:   time.Date(2024, 1, 8, 9, 0, 0, 0, time.UTC), // Next Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cronExpr, err := ParseCron(tt.expression)
			require.NoError(t, err, "Should parse cron expression")

			result := cronExpr.NextExecution(tt.from)
			assert.Equal(t, tt.expected, result, "Next execution time should match expected")
		})
	}
}
