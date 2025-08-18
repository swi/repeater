package recovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/swi/repeater/pkg/errors"
)

func TestErrorReporter_StructuredLogging(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Report a categorized error
	err := errors.NewCategorizedErrorWithContext(
		fmt.Errorf("connection timeout"),
		errors.CategoryNetwork,
		errors.SeverityHigh,
		map[string]interface{}{
			"host":    "example.com",
			"port":    443,
			"timeout": "30s",
		},
	)

	reporter.ReportError(err)

	// Verify structured log output
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "connection timeout") {
		t.Error("Expected error message in log output")
	}
	if !strings.Contains(logOutput, "network") {
		t.Error("Expected category in log output")
	}
	if !strings.Contains(logOutput, "high") {
		t.Error("Expected severity in log output")
	}
	if !strings.Contains(logOutput, "example.com") {
		t.Error("Expected context in log output")
	}
}

func TestErrorReporter_JSONFormat(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)
	reporter.SetFormat(FormatJSON)

	err := errors.NewCategorizedError(
		fmt.Errorf("test error"),
		errors.CategoryCommand,
		errors.SeverityMedium,
	)

	reporter.ReportError(err)

	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Verify JSON structure
	if logEntry["message"] != "test error" {
		t.Errorf("Expected message 'test error', got %v", logEntry["message"])
	}
	if logEntry["category"] != "command" {
		t.Errorf("Expected category 'command', got %v", logEntry["category"])
	}
	if logEntry["severity"] != "medium" {
		t.Errorf("Expected severity 'medium', got %v", logEntry["severity"])
	}
	if logEntry["timestamp"] == nil {
		t.Error("Expected timestamp in JSON log")
	}
}

func TestErrorReporter_TrendAnalysis(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable trend analysis
	reporter.EnableTrendAnalysis(true)

	// Report multiple errors of the same category
	for i := 0; i < 5; i++ {
		err := errors.NewCategorizedError(
			fmt.Errorf("timeout %d", i),
			errors.CategoryTimeout,
			errors.SeverityMedium,
		)
		reporter.ReportError(err)
	}

	// Get trend analysis
	trends := reporter.GetTrends(time.Hour)
	if len(trends) == 0 {
		t.Error("Expected trend analysis results")
	}

	// Find timeout trend
	var timeoutTrend *ErrorTrend
	for _, trend := range trends {
		if trend.Category == errors.CategoryTimeout {
			timeoutTrend = &trend
			break
		}
	}

	if timeoutTrend == nil {
		t.Error("Expected timeout trend in analysis")
		return
	}

	if timeoutTrend.Count != 5 {
		t.Errorf("Expected 5 timeout errors, got %d", timeoutTrend.Count)
	}

	if timeoutTrend.Rate <= 0 {
		t.Errorf("Expected positive error rate, got %f", timeoutTrend.Rate)
	}
}

func TestErrorReporter_AlertGeneration(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable trend analysis for alert tracking
	reporter.EnableTrendAnalysis(true)

	// Set alert threshold
	reporter.SetAlertThreshold(errors.CategoryNetwork, 3, time.Minute)
	// Report errors below threshold
	for i := 0; i < 2; i++ {
		err := errors.NewCategorizedError(
			fmt.Errorf("network error %d", i),
			errors.CategoryNetwork,
			errors.SeverityHigh,
		)
		alert := reporter.ReportError(err)
		if alert != nil {
			t.Error("Expected no alert below threshold")
		}
	}

	// Report error that crosses threshold
	err := errors.NewCategorizedError(
		fmt.Errorf("network error 3"),
		errors.CategoryNetwork,
		errors.SeverityHigh,
	)
	alert := reporter.ReportError(err)

	if alert == nil {
		t.Error("Expected alert when threshold is crossed")
		return
	}

	if alert.Category != errors.CategoryNetwork {
		t.Errorf("Expected network category alert, got %v", alert.Category)
	}

	if alert.Count < 3 {
		t.Errorf("Expected alert count >= 3, got %d", alert.Count)
	}
}

func TestErrorReporter_RecoveryTracking(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable recovery tracking
	reporter.EnableRecoveryTracking(true)

	// Simulate recovery scenario
	manager := NewRecoveryManager()
	policy := NewFixedDelayPolicy(2, 10*time.Millisecond)
	manager.SetRetryPolicy(policy)

	// Track recovery attempt
	attempts := 0
	err := manager.ExecuteWithRetry(context.Background(), func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			failErr := fmt.Errorf("failure %d", attempts)
			reporter.ReportRecoveryAttempt(failErr, attempts, false)
			return failErr
		}
		reporter.ReportRecoveryAttempt(nil, attempts, true)
		return nil
	})

	if err != nil {
		t.Errorf("Expected successful recovery, got %v", err)
	}

	// Verify recovery was tracked
	recoveryStats := reporter.GetRecoveryStatistics()
	if recoveryStats.TotalRecoveryAttempts != 3 {
		t.Errorf("Expected 3 recovery attempts, got %d", recoveryStats.TotalRecoveryAttempts)
	}

	if recoveryStats.SuccessfulRecoveries != 1 {
		t.Errorf("Expected 1 successful recovery, got %d", recoveryStats.SuccessfulRecoveries)
	}
}

func TestErrorReporter_CircuitBreakerIntegration(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Report circuit breaker state changes
	reporter.ReportCircuitBreakerStateChange("test-service", StateClosed, StateOpen)

	// Verify state change was logged
	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, "circuit_breaker_state_change") {
		t.Error("Expected circuit breaker state change in log")
	}
	if !strings.Contains(logOutput, "test-service") {
		t.Error("Expected service name in log")
	}
	if !strings.Contains(logOutput, "closed") {
		t.Error("Expected old state in log")
	}
	if !strings.Contains(logOutput, "open") {
		t.Error("Expected new state in log")
	}
}

func TestErrorReporter_MetricsIntegration(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable metrics integration
	reporter.EnableMetricsIntegration(true)

	// Report errors and verify metrics are updated
	for i := 0; i < 3; i++ {
		err := errors.NewCategorizedError(
			fmt.Errorf("error %d", i),
			errors.CategoryTimeout,
			errors.SeverityMedium,
		)
		reporter.ReportError(err)
	}

	// Get metrics
	metrics := reporter.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be available")
		return
	}

	if metrics.TotalErrors != 3 {
		t.Errorf("Expected 3 total errors, got %d", metrics.TotalErrors)
	}

	if metrics.ErrorsByCategory[errors.CategoryTimeout] != 3 {
		t.Errorf("Expected 3 timeout errors, got %d", metrics.ErrorsByCategory[errors.CategoryTimeout])
	}
}

func TestErrorReporter_HealthIntegration(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable health integration
	reporter.EnableHealthIntegration(true)

	// Report critical errors
	for i := 0; i < 2; i++ {
		err := errors.NewCategorizedError(
			fmt.Errorf("critical error %d", i),
			errors.CategorySystem,
			errors.SeverityCritical,
		)
		reporter.ReportError(err)
	}

	// Check health status
	healthStatus := reporter.GetHealthStatus()
	if healthStatus.Healthy {
		t.Error("Expected unhealthy status after critical errors")
	}

	if len(healthStatus.Issues) == 0 {
		t.Error("Expected health issues to be reported")
	}

	// Find critical error issue
	foundCritical := false
	for _, issue := range healthStatus.Issues {
		if strings.Contains(issue, "critical") {
			foundCritical = true
			break
		}
	}

	if !foundCritical {
		t.Error("Expected critical error issue in health status")
	}
}

func TestErrorReporter_ConcurrentReporting(t *testing.T) {
	var logBuffer bytes.Buffer
	reporter := NewErrorReporter(&logBuffer)

	// Enable all features for concurrent testing
	reporter.EnableTrendAnalysis(true)
	reporter.EnableRecoveryTracking(true)
	reporter.EnableMetricsIntegration(true)

	// Report errors concurrently
	const numGoroutines = 10
	const errorsPerGoroutine = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < errorsPerGoroutine; j++ {
				err := errors.NewCategorizedError(
					fmt.Errorf("concurrent error %d-%d", id, j),
					errors.CategoryNetwork,
					errors.SeverityMedium,
				)
				reporter.ReportError(err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify metrics
	metrics := reporter.GetMetrics()
	expectedTotal := numGoroutines * errorsPerGoroutine
	if metrics.TotalErrors != int64(expectedTotal) {
		t.Errorf("Expected %d total errors, got %d", expectedTotal, metrics.TotalErrors)
	}

	// Verify trends
	trends := reporter.GetTrends(time.Hour)
	if len(trends) == 0 {
		t.Error("Expected trend analysis results")
	}
}
