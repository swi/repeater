package recovery

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	rprErrors "github.com/swi/repeater/pkg/errors"
)

// LogFormat represents the format for error logging
type LogFormat int

const (
	FormatText LogFormat = iota
	FormatJSON
)

// ErrorTrend represents a trend in error occurrences
type ErrorTrend struct {
	Category  rprErrors.ErrorCategory `json:"category"`
	Count     int                     `json:"count"`
	Rate      float64                 `json:"rate"` // errors per hour
	TimeSpan  time.Duration           `json:"time_span"`
	FirstSeen time.Time               `json:"first_seen"`
	LastSeen  time.Time               `json:"last_seen"`
}

// RecoveryStatistics represents recovery operation statistics
type RecoveryStatistics struct {
	TotalRecoveryAttempts int           `json:"total_recovery_attempts"`
	SuccessfulRecoveries  int           `json:"successful_recoveries"`
	FailedRecoveries      int           `json:"failed_recoveries"`
	AverageRecoveryTime   time.Duration `json:"average_recovery_time"`
}

// ErrorMetrics represents error metrics for reporting
type ErrorMetrics struct {
	TotalErrors      int64                             `json:"total_errors"`
	ErrorsByCategory map[rprErrors.ErrorCategory]int64 `json:"errors_by_category"`
	ErrorsBySeverity map[rprErrors.ErrorSeverity]int64 `json:"errors_by_severity"`
	LastReportTime   time.Time                         `json:"last_report_time"`
}

// HealthStatus represents the health status based on error patterns
type HealthStatus struct {
	Healthy   bool      `json:"healthy"`
	Issues    []string  `json:"issues"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorReporter handles comprehensive error reporting and analysis
type ErrorReporter struct {
	mu                        sync.RWMutex
	writer                    io.Writer
	format                    LogFormat
	trendAnalysisEnabled      bool
	recoveryTrackingEnabled   bool
	metricsIntegrationEnabled bool
	healthIntegrationEnabled  bool

	// Alert thresholds
	alertThresholds map[rprErrors.ErrorCategory]struct {
		count    int
		timeSpan time.Duration
	}

	// Error tracking
	errors []rprErrors.CategorizedError

	// Recovery tracking
	recoveryStats RecoveryStatistics

	// Metrics
	metrics ErrorMetrics

	// Health status
	healthStatus HealthStatus
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(writer io.Writer) *ErrorReporter {
	return &ErrorReporter{
		writer: writer,
		format: FormatText,
		alertThresholds: make(map[rprErrors.ErrorCategory]struct {
			count    int
			timeSpan time.Duration
		}),
		errors: make([]rprErrors.CategorizedError, 0),
		metrics: ErrorMetrics{
			ErrorsByCategory: make(map[rprErrors.ErrorCategory]int64),
			ErrorsBySeverity: make(map[rprErrors.ErrorSeverity]int64),
		},
		healthStatus: HealthStatus{
			Healthy: true,
			Issues:  make([]string, 0),
		},
	}
}

// SetFormat sets the logging format
func (er *ErrorReporter) SetFormat(format LogFormat) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.format = format
}

// EnableTrendAnalysis enables or disables trend analysis
func (er *ErrorReporter) EnableTrendAnalysis(enabled bool) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.trendAnalysisEnabled = enabled
}

// EnableRecoveryTracking enables or disables recovery tracking
func (er *ErrorReporter) EnableRecoveryTracking(enabled bool) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.recoveryTrackingEnabled = enabled
}

// EnableMetricsIntegration enables or disables metrics integration
func (er *ErrorReporter) EnableMetricsIntegration(enabled bool) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.metricsIntegrationEnabled = enabled
}

// EnableHealthIntegration enables or disables health integration
func (er *ErrorReporter) EnableHealthIntegration(enabled bool) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.healthIntegrationEnabled = enabled
}

// SetAlertThreshold sets an alert threshold for a specific error category
func (er *ErrorReporter) SetAlertThreshold(category rprErrors.ErrorCategory, count int, timeSpan time.Duration) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.alertThresholds[category] = struct {
		count    int
		timeSpan time.Duration
	}{count: count, timeSpan: timeSpan}
}

// ReportError reports an error and returns an alert if threshold is exceeded
func (er *ErrorReporter) ReportError(err error) *rprErrors.ErrorAlert {
	er.mu.Lock()
	defer er.mu.Unlock()

	// Convert to categorized error if needed
	var catErr *rprErrors.CategorizedError
	if ce, ok := err.(*rprErrors.CategorizedError); ok {
		catErr = ce
	} else {
		category := rprErrors.ClassifyError(err)
		severity := rprErrors.DetermineSeverity(err)
		catErr = rprErrors.NewCategorizedError(err, category, severity)
	}

	// Log the error
	er.logError(catErr)

	// Update metrics
	if er.metricsIntegrationEnabled {
		er.updateMetrics(catErr)
	}

	// Update health status
	if er.healthIntegrationEnabled {
		er.updateHealthStatus(catErr)
	}

	// Check alert thresholds before adding to errors slice
	alert := er.checkAlertThreshold(catErr)

	// Track for trend analysis
	if er.trendAnalysisEnabled {
		er.errors = append(er.errors, *catErr)
	}

	return alert
}

// ReportRecoveryAttempt reports a recovery attempt
func (er *ErrorReporter) ReportRecoveryAttempt(err error, attempt int, success bool) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.recoveryTrackingEnabled {
		return
	}

	er.recoveryStats.TotalRecoveryAttempts++

	if success {
		er.recoveryStats.SuccessfulRecoveries++
	} else {
		er.recoveryStats.FailedRecoveries++
	}

	// Log recovery attempt
	er.logRecoveryAttempt(err, attempt, success)
}

// ReportCircuitBreakerStateChange reports a circuit breaker state change
func (er *ErrorReporter) ReportCircuitBreakerStateChange(name string, oldState, newState CircuitBreakerState) {
	er.mu.Lock()
	defer er.mu.Unlock()

	// Log state change
	er.logCircuitBreakerStateChange(name, oldState, newState)
}

// GetTrends returns error trends for the specified time window
func (er *ErrorReporter) GetTrends(timeWindow time.Duration) []ErrorTrend {
	er.mu.RLock()
	defer er.mu.RUnlock()

	if !er.trendAnalysisEnabled {
		return nil
	}

	cutoff := time.Now().Add(-timeWindow)
	trends := make(map[rprErrors.ErrorCategory]*ErrorTrend)

	for _, err := range er.errors {
		if err.Timestamp().Before(cutoff) {
			continue
		}

		category := err.Category()
		if trend, exists := trends[category]; exists {
			trend.Count++
			if err.Timestamp().After(trend.LastSeen) {
				trend.LastSeen = err.Timestamp()
			}
			if err.Timestamp().Before(trend.FirstSeen) {
				trend.FirstSeen = err.Timestamp()
			}
		} else {
			trends[category] = &ErrorTrend{
				Category:  category,
				Count:     1,
				TimeSpan:  timeWindow,
				FirstSeen: err.Timestamp(),
				LastSeen:  err.Timestamp(),
			}
		}
	}

	// Calculate rates
	result := make([]ErrorTrend, 0, len(trends))
	for _, trend := range trends {
		trend.Rate = float64(trend.Count) / timeWindow.Hours()
		result = append(result, *trend)
	}

	return result
}

// GetRecoveryStatistics returns recovery statistics
func (er *ErrorReporter) GetRecoveryStatistics() *RecoveryStatistics {
	er.mu.RLock()
	defer er.mu.RUnlock()

	if !er.recoveryTrackingEnabled {
		return nil
	}

	// Return a copy
	stats := er.recoveryStats
	return &stats
}

// GetMetrics returns error metrics
func (er *ErrorReporter) GetMetrics() *ErrorMetrics {
	er.mu.RLock()
	defer er.mu.RUnlock()

	if !er.metricsIntegrationEnabled {
		return nil
	}

	// Return a copy
	metrics := ErrorMetrics{
		TotalErrors:      er.metrics.TotalErrors,
		ErrorsByCategory: make(map[rprErrors.ErrorCategory]int64),
		ErrorsBySeverity: make(map[rprErrors.ErrorSeverity]int64),
		LastReportTime:   er.metrics.LastReportTime,
	}

	for k, v := range er.metrics.ErrorsByCategory {
		metrics.ErrorsByCategory[k] = v
	}
	for k, v := range er.metrics.ErrorsBySeverity {
		metrics.ErrorsBySeverity[k] = v
	}

	return &metrics
}

// GetHealthStatus returns the current health status
func (er *ErrorReporter) GetHealthStatus() *HealthStatus {
	er.mu.RLock()
	defer er.mu.RUnlock()

	if !er.healthIntegrationEnabled {
		return nil
	}

	// Return a copy
	status := HealthStatus{
		Healthy:   er.healthStatus.Healthy,
		Issues:    make([]string, len(er.healthStatus.Issues)),
		Timestamp: er.healthStatus.Timestamp,
	}
	copy(status.Issues, er.healthStatus.Issues)

	return &status
}

// logError logs an error in the specified format
func (er *ErrorReporter) logError(err *rprErrors.CategorizedError) {
	switch er.format {
	case FormatJSON:
		er.logErrorJSON(err)
	default:
		er.logErrorText(err)
	}
}

// logErrorText logs an error in text format
func (er *ErrorReporter) logErrorText(err *rprErrors.CategorizedError) {
	if _, writeErr := fmt.Fprintf(er.writer, "[%s] %s category=%s severity=%s",
		err.Timestamp().Format(time.RFC3339),
		err.Error(),
		err.Category().String(),
		err.Severity().String()); writeErr != nil {
		// Ignore write errors to avoid cascading failures
		return
	}

	if context := err.Context(); context != nil {
		if _, writeErr := fmt.Fprintf(er.writer, " context=%v", context); writeErr != nil {
			return
		}
	}

	_, _ = fmt.Fprintln(er.writer) // Ignore error on final newline
}

// logErrorJSON logs an error in JSON format
func (er *ErrorReporter) logErrorJSON(err *rprErrors.CategorizedError) {
	logEntry := map[string]any{
		"timestamp": err.Timestamp().Format(time.RFC3339),
		"message":   err.Error(),
		"category":  err.Category().String(),
		"severity":  err.Severity().String(),
	}

	if context := err.Context(); context != nil {
		logEntry["context"] = context
	}

	jsonData, _ := json.Marshal(logEntry)
	_, _ = fmt.Fprintln(er.writer, string(jsonData)) // Ignore write error
}

// logRecoveryAttempt logs a recovery attempt
func (er *ErrorReporter) logRecoveryAttempt(err error, attempt int, success bool) {
	status := "failed"
	if success {
		status = "succeeded"
	}

	if _, writeErr := fmt.Fprintf(er.writer, "[%s] recovery_attempt attempt=%d status=%s",
		time.Now().Format(time.RFC3339), attempt, status); writeErr != nil {
		return
	}

	if err != nil {
		if _, writeErr := fmt.Fprintf(er.writer, " error=%s", err.Error()); writeErr != nil {
			return
		}
	}

	_, _ = fmt.Fprintln(er.writer) // Ignore error on final newline
}

// logCircuitBreakerStateChange logs a circuit breaker state change
func (er *ErrorReporter) logCircuitBreakerStateChange(name string, oldState, newState CircuitBreakerState) {
	_, _ = fmt.Fprintf(er.writer, "[%s] circuit_breaker_state_change service=%s old_state=%s new_state=%s\n",
		time.Now().Format(time.RFC3339), name, oldState.String(), newState.String())
}

// updateMetrics updates error metrics
func (er *ErrorReporter) updateMetrics(err *rprErrors.CategorizedError) {
	er.metrics.TotalErrors++
	er.metrics.ErrorsByCategory[err.Category()]++
	er.metrics.ErrorsBySeverity[err.Severity()]++
	er.metrics.LastReportTime = time.Now()
}

// updateHealthStatus updates health status based on error severity
func (er *ErrorReporter) updateHealthStatus(err *rprErrors.CategorizedError) {
	er.healthStatus.Timestamp = time.Now()

	// Mark as unhealthy for critical errors
	if err.Severity() == rprErrors.SeverityCritical {
		er.healthStatus.Healthy = false
		issue := fmt.Sprintf("Critical error: %s", err.Error())
		er.healthStatus.Issues = append(er.healthStatus.Issues, issue)
	}
}

// checkAlertThreshold checks if an alert threshold is exceeded
func (er *ErrorReporter) checkAlertThreshold(err *rprErrors.CategorizedError) *rprErrors.ErrorAlert {
	threshold, exists := er.alertThresholds[err.Category()]
	if !exists {
		return nil
	}

	// Count errors of this category within the time span (including current error)
	cutoff := time.Now().Add(-threshold.timeSpan)
	count := 1 // Count the current error
	for _, e := range er.errors {
		if e.Category() == err.Category() && e.Timestamp().After(cutoff) {
			count++
		}
	}

	// Check if threshold is exceeded
	if count >= threshold.count {
		return &rprErrors.ErrorAlert{
			Category:  err.Category(),
			Count:     count,
			Threshold: threshold.count,
			TimeSpan:  threshold.timeSpan,
			Timestamp: time.Now(),
		}
	}

	return nil
}
