package errors

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestErrorCategory_Classification(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorCategory
	}{
		{
			name:     "timeout error",
			err:      context.DeadlineExceeded,
			expected: CategoryTimeout,
		},
		{
			name:     "network connection error",
			err:      fmt.Errorf("connection refused"),
			expected: CategoryNetwork,
		},
		{
			name:     "permission denied error",
			err:      fmt.Errorf("permission denied"),
			expected: CategoryPermission,
		},
		{
			name:     "resource exhausted error",
			err:      fmt.Errorf("out of memory"),
			expected: CategoryResource,
		},
		{
			name:     "command not found error",
			err:      fmt.Errorf("command not found"),
			expected: CategoryCommand,
		},
		{
			name:     "generic system error",
			err:      fmt.Errorf("system error"),
			expected: CategorySystem,
		},
		{
			name:     "unknown error",
			err:      fmt.Errorf("some random error"),
			expected: CategoryUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := ClassifyError(tt.err)
			if category != tt.expected {
				t.Errorf("Expected category %v, got %v", tt.expected, category)
			}
		})
	}
}

func TestErrorSeverity_Assignment(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorSeverity
	}{
		{
			name:     "critical system failure",
			err:      fmt.Errorf("kernel panic"),
			expected: SeverityCritical,
		},
		{
			name:     "high severity permission error",
			err:      fmt.Errorf("permission denied"),
			expected: SeverityHigh,
		},
		{
			name:     "medium severity timeout",
			err:      context.DeadlineExceeded,
			expected: SeverityMedium,
		},
		{
			name:     "low severity warning",
			err:      fmt.Errorf("deprecated feature used"),
			expected: SeverityLow,
		},
		{
			name:     "info level message",
			err:      fmt.Errorf("operation completed with warnings"),
			expected: SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := DetermineSeverity(tt.err)
			if severity != tt.expected {
				t.Errorf("Expected severity %v, got %v", tt.expected, severity)
			}
		})
	}
}

func TestCategorizedError_Creation(t *testing.T) {
	originalErr := fmt.Errorf("connection timeout")

	catErr := NewCategorizedError(originalErr, CategoryNetwork, SeverityHigh)

	if catErr.Error() != originalErr.Error() {
		t.Errorf("Expected error message '%s', got '%s'", originalErr.Error(), catErr.Error())
	}

	if catErr.Category() != CategoryNetwork {
		t.Errorf("Expected category %v, got %v", CategoryNetwork, catErr.Category())
	}

	if catErr.Severity() != SeverityHigh {
		t.Errorf("Expected severity %v, got %v", SeverityHigh, catErr.Severity())
	}

	if catErr.Unwrap() != originalErr {
		t.Error("Expected Unwrap() to return original error")
	}
}

func TestCategorizedError_WithContext(t *testing.T) {
	originalErr := fmt.Errorf("command failed")
	context := map[string]interface{}{
		"command":     "curl",
		"args":        []string{"-v", "https://example.com"},
		"exit_code":   1,
		"duration":    "5s",
		"retry_count": 2,
	}

	catErr := NewCategorizedErrorWithContext(originalErr, CategoryCommand, SeverityMedium, context)

	if catErr.Context() == nil {
		t.Error("Expected context to be set")
	}

	if catErr.Context()["command"] != "curl" {
		t.Errorf("Expected command 'curl', got %v", catErr.Context()["command"])
	}

	if catErr.Context()["retry_count"] != 2 {
		t.Errorf("Expected retry_count 2, got %v", catErr.Context()["retry_count"])
	}
}

func TestErrorPattern_Detection(t *testing.T) {
	detector := NewErrorPatternDetector()

	// Simulate repeated timeout errors
	timeoutErr := context.DeadlineExceeded
	for i := 0; i < 5; i++ {
		detector.RecordError(NewCategorizedError(timeoutErr, CategoryTimeout, SeverityMedium))
	}

	patterns := detector.DetectPatterns(time.Minute)

	if len(patterns) == 0 {
		t.Error("Expected to detect timeout pattern")
	}

	found := false
	for _, pattern := range patterns {
		if pattern.Category == CategoryTimeout && pattern.Count >= 5 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find timeout pattern with count >= 5")
	}
}

func TestErrorPattern_ThresholdAlert(t *testing.T) {
	detector := NewErrorPatternDetector()

	// Set threshold for network errors
	detector.SetThreshold(CategoryNetwork, 3, time.Minute)

	networkErr := fmt.Errorf("connection refused")

	// Record errors below threshold
	for i := 0; i < 2; i++ {
		alert := detector.RecordError(NewCategorizedError(networkErr, CategoryNetwork, SeverityHigh))
		if alert != nil {
			t.Error("Expected no alert below threshold")
		}
	}

	// Record error that crosses threshold
	alert := detector.RecordError(NewCategorizedError(networkErr, CategoryNetwork, SeverityHigh))
	if alert == nil {
		t.Error("Expected alert when threshold is crossed")
		return
	}

	if alert.Category != CategoryNetwork {
		t.Errorf("Expected alert category %v, got %v", CategoryNetwork, alert.Category)
	}

	if alert.Count != 3 {
		t.Errorf("Expected alert count 3, got %d", alert.Count)
	}
}

func TestErrorAggregation_Statistics(t *testing.T) {
	aggregator := NewErrorAggregator()

	// Record various errors
	errors := []struct {
		err      error
		category ErrorCategory
		severity ErrorSeverity
	}{
		{fmt.Errorf("timeout 1"), CategoryTimeout, SeverityMedium},
		{fmt.Errorf("timeout 2"), CategoryTimeout, SeverityMedium},
		{fmt.Errorf("network 1"), CategoryNetwork, SeverityHigh},
		{fmt.Errorf("permission 1"), CategoryPermission, SeverityHigh},
		{fmt.Errorf("timeout 3"), CategoryTimeout, SeverityLow},
	}

	for _, e := range errors {
		aggregator.RecordError(NewCategorizedError(e.err, e.category, e.severity))
	}

	stats := aggregator.GetStatistics(time.Hour)

	// Check total count
	if stats.TotalErrors != 5 {
		t.Errorf("Expected 5 total errors, got %d", stats.TotalErrors)
	}

	// Check category breakdown
	if stats.ByCategory[CategoryTimeout] != 3 {
		t.Errorf("Expected 3 timeout errors, got %d", stats.ByCategory[CategoryTimeout])
	}

	if stats.ByCategory[CategoryNetwork] != 1 {
		t.Errorf("Expected 1 network error, got %d", stats.ByCategory[CategoryNetwork])
	}

	// Check severity breakdown
	if stats.BySeverity[SeverityHigh] != 2 {
		t.Errorf("Expected 2 high severity errors, got %d", stats.BySeverity[SeverityHigh])
	}

	if stats.BySeverity[SeverityMedium] != 2 {
		t.Errorf("Expected 2 medium severity errors, got %d", stats.BySeverity[SeverityMedium])
	}
}

func TestErrorCategory_String(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{CategoryTimeout, "timeout"},
		{CategoryNetwork, "network"},
		{CategoryPermission, "permission"},
		{CategoryResource, "resource"},
		{CategoryCommand, "command"},
		{CategorySystem, "system"},
		{CategoryUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.category.String() != tt.expected {
				t.Errorf("Expected string '%s', got '%s'", tt.expected, tt.category.String())
			}
		})
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{SeverityCritical, "critical"},
		{SeverityHigh, "high"},
		{SeverityMedium, "medium"},
		{SeverityLow, "low"},
		{SeverityInfo, "info"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.severity.String() != tt.expected {
				t.Errorf("Expected string '%s', got '%s'", tt.expected, tt.severity.String())
			}
		})
	}
}
