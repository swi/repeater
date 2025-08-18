package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	rprErrors "github.com/swi/repeater/pkg/errors"
)

// RetryPolicy defines the interface for retry policies
type RetryPolicy interface {
	ShouldRetry(attempt int) bool
	NextDelay(attempt int) time.Duration
	ShouldRetryError(err error) bool
}

// ExponentialBackoffPolicy implements exponential backoff retry policy
type ExponentialBackoffPolicy struct {
	maxRetries   int
	initialDelay time.Duration
	multiplier   float64
	maxDelay     time.Duration
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(maxRetries int, initialDelay time.Duration, multiplier float64, maxDelay time.Duration) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		multiplier:   multiplier,
		maxDelay:     maxDelay,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay calculates the delay for the next retry attempt
func (p *ExponentialBackoffPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := p.initialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * p.multiplier)
	}

	if delay > p.maxDelay {
		delay = p.maxDelay
	}

	return delay
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *ExponentialBackoffPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// LinearBackoffPolicy implements linear backoff retry policy
type LinearBackoffPolicy struct {
	maxRetries   int
	initialDelay time.Duration
	increment    time.Duration
}

// NewLinearBackoffPolicy creates a new linear backoff policy
func NewLinearBackoffPolicy(maxRetries int, initialDelay time.Duration, increment time.Duration) *LinearBackoffPolicy {
	return &LinearBackoffPolicy{
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		increment:    increment,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *LinearBackoffPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay calculates the delay for the next retry attempt
func (p *LinearBackoffPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	return p.initialDelay + time.Duration(attempt-1)*p.increment
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *LinearBackoffPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// FixedDelayPolicy implements fixed delay retry policy
type FixedDelayPolicy struct {
	maxRetries int
	delay      time.Duration
}

// NewFixedDelayPolicy creates a new fixed delay policy
func NewFixedDelayPolicy(maxRetries int, delay time.Duration) *FixedDelayPolicy {
	return &FixedDelayPolicy{
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *FixedDelayPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay returns the fixed delay for retry attempts
func (p *FixedDelayPolicy) NextDelay(attempt int) time.Duration {
	return p.delay
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *FixedDelayPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// ConditionalRetryPolicy implements conditional retry policy
type ConditionalRetryPolicy struct {
	maxRetries int
	delay      time.Duration
	conditions []func(error) bool
}

// NewConditionalRetryPolicy creates a new conditional retry policy
func NewConditionalRetryPolicy(maxRetries int, delay time.Duration) *ConditionalRetryPolicy {
	return &ConditionalRetryPolicy{
		maxRetries: maxRetries,
		delay:      delay,
		conditions: make([]func(error) bool, 0),
	}
}

// AddCondition adds a condition for retrying errors
func (p *ConditionalRetryPolicy) AddCondition(condition func(error) bool) {
	p.conditions = append(p.conditions, condition)
}

// ShouldRetry returns true if the attempt should be retried
func (p *ConditionalRetryPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay returns the delay for retry attempts
func (p *ConditionalRetryPolicy) NextDelay(attempt int) time.Duration {
	return p.delay
}

// ShouldRetryError returns true if the error matches any retry condition
func (p *ConditionalRetryPolicy) ShouldRetryError(err error) bool {
	for _, condition := range p.conditions {
		if condition(err) {
			return true
		}
	}
	return false
}

// FallbackFunc defines the signature for fallback functions
type FallbackFunc func(ctx context.Context, originalErr error) error

// ExecuteFunc defines the signature for functions to be executed with recovery
type ExecuteFunc func(ctx context.Context) error

// RecoveryState tracks the state of recovery operations
type RecoveryState struct {
	TotalAttempts        int           `json:"total_attempts"`
	SuccessfulRecoveries int           `json:"successful_recoveries"`
	FailedRecoveries     int           `json:"failed_recoveries"`
	ConsecutiveSuccesses int           `json:"consecutive_successes"`
	ConsecutiveFailures  int           `json:"consecutive_failures"`
	RecentFailures       []error       `json:"recent_failures"`
	LastSuccessTime      time.Time     `json:"last_success_time"`
	LastFailureTime      time.Time     `json:"last_failure_time"`
	AverageRecoveryTime  time.Duration `json:"average_recovery_time"`
}

// RecoveryManager manages retry policies and fallback strategies
type RecoveryManager struct {
	mu               sync.RWMutex
	retryPolicy      RetryPolicy
	fallback         FallbackFunc
	successThreshold int
	stateTracking    bool
	state            *RecoveryState
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		successThreshold: 1,
		stateTracking:    false,
		state: &RecoveryState{
			RecentFailures: make([]error, 0),
		},
	}
}

// SetRetryPolicy sets the retry policy
func (rm *RecoveryManager) SetRetryPolicy(policy RetryPolicy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.retryPolicy = policy
}

// SetFallback sets the fallback function
func (rm *RecoveryManager) SetFallback(fallback FallbackFunc) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.fallback = fallback
}

// SetSuccessThreshold sets the number of consecutive successes needed
func (rm *RecoveryManager) SetSuccessThreshold(threshold int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.successThreshold = threshold
}

// EnableStateTracking enables or disables state tracking
func (rm *RecoveryManager) EnableStateTracking(enabled bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.stateTracking = enabled
}

// GetRecoveryState returns the current recovery state
func (rm *RecoveryManager) GetRecoveryState() *RecoveryState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if !rm.stateTracking {
		return nil
	}

	// Return a copy to avoid race conditions
	stateCopy := *rm.state
	stateCopy.RecentFailures = make([]error, len(rm.state.RecentFailures))
	copy(stateCopy.RecentFailures, rm.state.RecentFailures)

	return &stateCopy
}

// ExecuteWithRetry executes a function with retry policy
func (rm *RecoveryManager) ExecuteWithRetry(ctx context.Context, fn ExecuteFunc) error {
	rm.mu.RLock()
	policy := rm.retryPolicy
	rm.mu.RUnlock()

	if policy == nil {
		// No retry policy, execute once
		return rm.executeAndTrack(ctx, fn)
	}

	var lastErr error
	attempt := 0

	for {
		attempt++

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := rm.executeAndTrack(ctx, fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry
		if !policy.ShouldRetry(attempt) || !policy.ShouldRetryError(err) {
			break
		}

		// Wait for retry delay
		delay := policy.NextDelay(attempt)
		if delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

// ExecuteWithFallback executes a function with fallback on failure
func (rm *RecoveryManager) ExecuteWithFallback(ctx context.Context, fn ExecuteFunc) error {
	err := rm.executeAndTrack(ctx, fn)
	if err == nil {
		return nil
	}

	rm.mu.RLock()
	fallback := rm.fallback
	rm.mu.RUnlock()

	if fallback == nil {
		return err
	}

	// Execute fallback
	return fallback(ctx, err)
}

// ExecuteWithRetryAndFallback executes a function with both retry and fallback
func (rm *RecoveryManager) ExecuteWithRetryAndFallback(ctx context.Context, fn ExecuteFunc) error {
	err := rm.ExecuteWithRetry(ctx, fn)
	if err == nil {
		return nil
	}

	rm.mu.RLock()
	fallback := rm.fallback
	rm.mu.RUnlock()

	if fallback == nil {
		return err
	}

	// Execute fallback after retries are exhausted
	return fallback(ctx, err)
}

// executeAndTrack executes a function and tracks the result
func (rm *RecoveryManager) executeAndTrack(ctx context.Context, fn ExecuteFunc) error {
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.stateTracking {
		return err
	}

	rm.state.TotalAttempts++

	if err == nil {
		// Success
		rm.state.SuccessfulRecoveries++
		rm.state.ConsecutiveSuccesses++
		rm.state.ConsecutiveFailures = 0
		rm.state.LastSuccessTime = time.Now()

		// Update average recovery time
		if rm.state.AverageRecoveryTime == 0 {
			rm.state.AverageRecoveryTime = duration
		} else {
			rm.state.AverageRecoveryTime = (rm.state.AverageRecoveryTime + duration) / 2
		}
	} else {
		// Failure
		rm.state.FailedRecoveries++
		rm.state.ConsecutiveFailures++
		rm.state.ConsecutiveSuccesses = 0
		rm.state.LastFailureTime = time.Now()

		// Track recent failures (keep last 10)
		rm.state.RecentFailures = append(rm.state.RecentFailures, err)
		if len(rm.state.RecentFailures) > 10 {
			rm.state.RecentFailures = rm.state.RecentFailures[1:]
		}
	}

	return err
}

// Circuit Breaker Implementation

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

// CircuitBreakerStatistics contains statistics about circuit breaker operations
type CircuitBreakerStatistics struct {
	Name            string              `json:"name"`
	State           CircuitBreakerState `json:"state"`
	TotalRequests   int64               `json:"total_requests"`
	SuccessCount    int64               `json:"success_count"`
	FailureCount    int64               `json:"failure_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	LastSuccessTime time.Time           `json:"last_success_time"`
	StateChangedAt  time.Time           `json:"state_changed_at"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	name             string
	failureThreshold int
	timeout          time.Duration
	resetTimeout     time.Duration
	state            CircuitBreakerState
	failureCount     int
	successCount     int64
	totalRequests    int64
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	stateChangedAt   time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold int, timeout time.Duration, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
		stateChangedAt:   time.Now(),
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn ExecuteFunc) error {
	// Check if we can execute
	if !cb.canExecute() {
		return ErrCircuitBreakerOpen
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	if err != nil {
		cb.RecordFailure()
	} else {
		cb.RecordSuccess()
	}

	return err
}

// canExecute determines if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed to transition to Half-Open
		if time.Since(cb.stateChangedAt) >= cb.timeout {
			cb.state = StateHalfOpen
			cb.stateChangedAt = time.Now()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.totalRequests++
	cb.lastSuccessTime = time.Now()

	// Reset failure count and transition to Closed if in Half-Open
	if cb.state == StateHalfOpen {
		cb.failureCount = 0
		cb.state = StateClosed
		cb.stateChangedAt = time.Now()
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.totalRequests++
	cb.lastFailureTime = time.Now()

	// Check if we should transition to Open
	if cb.failureCount >= cb.failureThreshold {
		if cb.state != StateOpen {
			cb.state = StateOpen
			cb.stateChangedAt = time.Now()
		}
	}
}

// Reset resets the circuit breaker to Closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.stateChangedAt = time.Now()
}

// Statistics returns the current statistics of the circuit breaker
func (cb *CircuitBreaker) Statistics() *CircuitBreakerStatistics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &CircuitBreakerStatistics{
		Name:            cb.name,
		State:           cb.state,
		TotalRequests:   cb.totalRequests,
		SuccessCount:    cb.successCount,
		FailureCount:    int64(cb.failureCount),
		LastFailureTime: cb.lastFailureTime,
		LastSuccessTime: cb.lastSuccessTime,
		StateChangedAt:  cb.stateChangedAt,
	}
}

// Error Reporting Implementation

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
	logEntry := map[string]interface{}{
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
