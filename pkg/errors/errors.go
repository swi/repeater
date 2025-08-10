package errors

import (
	"context"
	"strings"
	"sync"
	"time"
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	CategoryUnknown ErrorCategory = iota
	CategoryTimeout
	CategoryNetwork
	CategoryPermission
	CategoryResource
	CategoryCommand
	CategorySystem
)

// String returns the string representation of the error category
func (c ErrorCategory) String() string {
	switch c {
	case CategoryTimeout:
		return "timeout"
	case CategoryNetwork:
		return "network"
	case CategoryPermission:
		return "permission"
	case CategoryResource:
		return "resource"
	case CategoryCommand:
		return "command"
	case CategorySystem:
		return "system"
	default:
		return "unknown"
	}
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String returns the string representation of the error severity
func (s ErrorSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// CategorizedError represents an error with category, severity, and context
type CategorizedError struct {
	err       error
	category  ErrorCategory
	severity  ErrorSeverity
	context   map[string]interface{}
	timestamp time.Time
}

// NewCategorizedError creates a new categorized error
func NewCategorizedError(err error, category ErrorCategory, severity ErrorSeverity) *CategorizedError {
	return &CategorizedError{
		err:       err,
		category:  category,
		severity:  severity,
		timestamp: time.Now(),
	}
}

// NewCategorizedErrorWithContext creates a new categorized error with context
func NewCategorizedErrorWithContext(err error, category ErrorCategory, severity ErrorSeverity, context map[string]interface{}) *CategorizedError {
	return &CategorizedError{
		err:       err,
		category:  category,
		severity:  severity,
		context:   context,
		timestamp: time.Now(),
	}
}

// Error implements the error interface
func (e *CategorizedError) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error
func (e *CategorizedError) Unwrap() error {
	return e.err
}

// Category returns the error category
func (e *CategorizedError) Category() ErrorCategory {
	return e.category
}

// Severity returns the error severity
func (e *CategorizedError) Severity() ErrorSeverity {
	return e.severity
}

// Context returns the error context
func (e *CategorizedError) Context() map[string]interface{} {
	return e.context
}

// Timestamp returns when the error occurred
func (e *CategorizedError) Timestamp() time.Time {
	return e.timestamp
}

// ClassifyError automatically classifies an error based on its content
func ClassifyError(err error) ErrorCategory {
	if err == nil {
		return CategoryUnknown
	}

	errMsg := strings.ToLower(err.Error())

	// Check for timeout errors
	if err == context.DeadlineExceeded || strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
		return CategoryTimeout
	}

	// Check for network errors
	if strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "refused") || strings.Contains(errMsg, "unreachable") {
		return CategoryNetwork
	}

	// Check for permission errors
	if strings.Contains(errMsg, "permission") || strings.Contains(errMsg, "denied") ||
		strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "forbidden") {
		return CategoryPermission
	}

	// Check for resource errors
	if strings.Contains(errMsg, "memory") || strings.Contains(errMsg, "disk") ||
		strings.Contains(errMsg, "space") || strings.Contains(errMsg, "resource") {
		return CategoryResource
	}

	// Check for command errors
	if strings.Contains(errMsg, "command not found") || strings.Contains(errMsg, "executable") ||
		strings.Contains(errMsg, "no such file") {
		return CategoryCommand
	}

	// Check for system errors
	if strings.Contains(errMsg, "system") || strings.Contains(errMsg, "kernel") ||
		strings.Contains(errMsg, "panic") {
		return CategorySystem
	}

	return CategoryUnknown
}

// DetermineSeverity automatically determines error severity based on content
func DetermineSeverity(err error) ErrorSeverity {
	if err == nil {
		return SeverityInfo
	}

	errMsg := strings.ToLower(err.Error())

	// Critical errors
	if strings.Contains(errMsg, "panic") || strings.Contains(errMsg, "kernel") ||
		strings.Contains(errMsg, "critical") || strings.Contains(errMsg, "fatal") {
		return SeverityCritical
	}

	// High severity errors
	if strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "access denied") {
		return SeverityHigh
	}

	// Info level (check before low severity to catch "completed with warnings")
	if strings.Contains(errMsg, "completed") || strings.Contains(errMsg, "info") {
		return SeverityInfo
	}

	// Medium severity errors
	if err == context.DeadlineExceeded || strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "connection") {
		return SeverityMedium
	}

	// Low severity errors
	if strings.Contains(errMsg, "warning") || strings.Contains(errMsg, "deprecated") {
		return SeverityLow
	}
	return SeverityMedium // Default to medium
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	Category  ErrorCategory `json:"category"`
	Count     int           `json:"count"`
	TimeSpan  time.Duration `json:"time_span"`
	FirstSeen time.Time     `json:"first_seen"`
	LastSeen  time.Time     `json:"last_seen"`
}

// ErrorAlert represents an alert triggered by error patterns
type ErrorAlert struct {
	Category  ErrorCategory `json:"category"`
	Count     int           `json:"count"`
	Threshold int           `json:"threshold"`
	TimeSpan  time.Duration `json:"time_span"`
	Timestamp time.Time     `json:"timestamp"`
}

// ErrorPatternDetector detects patterns in error occurrences
type ErrorPatternDetector struct {
	mu         sync.RWMutex
	errors     []*CategorizedError
	thresholds map[ErrorCategory]struct {
		count    int
		timeSpan time.Duration
	}
}

// NewErrorPatternDetector creates a new error pattern detector
func NewErrorPatternDetector() *ErrorPatternDetector {
	return &ErrorPatternDetector{
		errors: make([]*CategorizedError, 0),
		thresholds: make(map[ErrorCategory]struct {
			count    int
			timeSpan time.Duration
		}),
	}
}

// RecordError records an error and checks for threshold violations
func (d *ErrorPatternDetector) RecordError(err *CategorizedError) *ErrorAlert {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.errors = append(d.errors, err)

	// Check if threshold is set for this category
	threshold, exists := d.thresholds[err.Category()]
	if !exists {
		return nil
	}

	// Count errors of this category within the time span
	cutoff := time.Now().Add(-threshold.timeSpan)
	count := 0
	for _, e := range d.errors {
		if e.Category() == err.Category() && e.Timestamp().After(cutoff) {
			count++
		}
	}

	// Check if threshold is exceeded
	if count >= threshold.count {
		return &ErrorAlert{
			Category:  err.Category(),
			Count:     count,
			Threshold: threshold.count,
			TimeSpan:  threshold.timeSpan,
			Timestamp: time.Now(),
		}
	}

	return nil
}

// SetThreshold sets an alert threshold for a specific error category
func (d *ErrorPatternDetector) SetThreshold(category ErrorCategory, count int, timeSpan time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.thresholds[category] = struct {
		count    int
		timeSpan time.Duration
	}{count: count, timeSpan: timeSpan}
}

// DetectPatterns detects error patterns within a given time window
func (d *ErrorPatternDetector) DetectPatterns(timeWindow time.Duration) []ErrorPattern {
	d.mu.RLock()
	defer d.mu.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	patterns := make(map[ErrorCategory]*ErrorPattern)

	for _, err := range d.errors {
		if err.Timestamp().Before(cutoff) {
			continue
		}

		category := err.Category()
		if pattern, exists := patterns[category]; exists {
			pattern.Count++
			if err.Timestamp().After(pattern.LastSeen) {
				pattern.LastSeen = err.Timestamp()
			}
			if err.Timestamp().Before(pattern.FirstSeen) {
				pattern.FirstSeen = err.Timestamp()
			}
		} else {
			patterns[category] = &ErrorPattern{
				Category:  category,
				Count:     1,
				TimeSpan:  timeWindow,
				FirstSeen: err.Timestamp(),
				LastSeen:  err.Timestamp(),
			}
		}
	}

	result := make([]ErrorPattern, 0, len(patterns))
	for _, pattern := range patterns {
		result = append(result, *pattern)
	}

	return result
}

// ErrorStatistics represents aggregated error statistics
type ErrorStatistics struct {
	TotalErrors int                   `json:"total_errors"`
	ByCategory  map[ErrorCategory]int `json:"by_category"`
	BySeverity  map[ErrorSeverity]int `json:"by_severity"`
	TimeWindow  time.Duration         `json:"time_window"`
	GeneratedAt time.Time             `json:"generated_at"`
}

// ErrorAggregator aggregates error statistics
type ErrorAggregator struct {
	mu     sync.RWMutex
	errors []*CategorizedError
}

// NewErrorAggregator creates a new error aggregator
func NewErrorAggregator() *ErrorAggregator {
	return &ErrorAggregator{
		errors: make([]*CategorizedError, 0),
	}
}

// RecordError records an error for aggregation
func (a *ErrorAggregator) RecordError(err *CategorizedError) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errors = append(a.errors, err)
}

// GetStatistics returns aggregated error statistics for a time window
func (a *ErrorAggregator) GetStatistics(timeWindow time.Duration) *ErrorStatistics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cutoff := time.Now().Add(-timeWindow)
	stats := &ErrorStatistics{
		ByCategory:  make(map[ErrorCategory]int),
		BySeverity:  make(map[ErrorSeverity]int),
		TimeWindow:  timeWindow,
		GeneratedAt: time.Now(),
	}

	for _, err := range a.errors {
		if err.Timestamp().Before(cutoff) {
			continue
		}

		stats.TotalErrors++
		stats.ByCategory[err.Category()]++
		stats.BySeverity[err.Severity()]++
	}

	return stats
}
