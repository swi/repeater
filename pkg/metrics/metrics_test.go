package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsServer_StartStop(t *testing.T) {
	server := NewMetricsServer(9090)

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:9090/metrics")
	if err != nil {
		t.Fatalf("Expected server to be running, got error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Stop server
	cancel()

	// Wait for server to stop
	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			t.Errorf("Expected clean shutdown, got error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Server did not stop within timeout")
	}
}

func TestMetricsEndpoint_PrometheusFormat(t *testing.T) {
	server := NewMetricsServer(0) // Use random port for testing

	// Record some test metrics
	server.RecordExecution(true, 150*time.Millisecond)
	server.RecordExecution(false, 300*time.Millisecond)
	server.RecordExecution(true, 100*time.Millisecond)

	// Create test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Call handler directly
	server.metricsHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify Prometheus format metrics are present
	expectedMetrics := []string{
		"# HELP rpr_executions_total Total number of command executions",
		"# TYPE rpr_executions_total counter",
		"rpr_executions_total{status=\"success\"} 2",
		"rpr_executions_total{status=\"failure\"} 1",
		"# HELP rpr_execution_duration_seconds Duration of command executions",
		"# TYPE rpr_execution_duration_seconds histogram",
		"# HELP rpr_execution_duration_seconds_sum Sum of execution durations",
		"# TYPE rpr_execution_duration_seconds_sum counter",
		"# HELP rpr_execution_duration_seconds_count Count of execution durations",
		"# TYPE rpr_execution_duration_seconds_count counter",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected metric line not found: %s", expected)
		}
	}
}

func TestMetricsServer_RecordExecution(t *testing.T) {
	server := NewMetricsServer(0)

	// Record successful execution
	server.RecordExecution(true, 100*time.Millisecond)

	// Record failed execution
	server.RecordExecution(false, 200*time.Millisecond)

	// Get metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.metricsHandler(w, req)

	body := w.Body.String()

	// Verify counters were incremented
	if !strings.Contains(body, "rpr_executions_total{status=\"success\"} 1") {
		t.Error("Success counter not incremented correctly")
	}

	if !strings.Contains(body, "rpr_executions_total{status=\"failure\"} 1") {
		t.Error("Failure counter not incremented correctly")
	}
}

func TestMetricsServer_RecordSchedulerMetrics(t *testing.T) {
	server := NewMetricsServer(0)

	// Record scheduler metrics
	server.RecordSchedulerInterval(5 * time.Second)
	server.RecordSchedulerInterval(3 * time.Second)
	server.RecordSchedulerInterval(7 * time.Second)

	// Get metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.metricsHandler(w, req)

	body := w.Body.String()

	// Verify scheduler metrics are present
	expectedMetrics := []string{
		"# HELP rpr_scheduler_interval_seconds Current scheduler interval",
		"# TYPE rpr_scheduler_interval_seconds gauge",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected scheduler metric not found: %s", expected)
		}
	}
}

func TestMetricsServer_RecordRateLimitMetrics(t *testing.T) {
	server := NewMetricsServer(0)

	// Record rate limit metrics
	server.RecordRateLimitHit()
	server.RecordRateLimitHit()
	server.RecordRateLimitAllowed()

	// Get metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.metricsHandler(w, req)

	body := w.Body.String()

	// Verify rate limit metrics are present
	expectedMetrics := []string{
		"# HELP rpr_rate_limit_total Rate limit events",
		"# TYPE rpr_rate_limit_total counter",
		"rpr_rate_limit_total{result=\"hit\"} 2",
		"rpr_rate_limit_total{result=\"allowed\"} 1",
	}

	for _, expected := range expectedMetrics {
		if !strings.Contains(body, expected) {
			t.Errorf("Expected rate limit metric not found: %s", expected)
		}
	}
}

func TestMetricsServer_ConcurrentAccess(t *testing.T) {
	server := NewMetricsServer(0)

	// Record metrics concurrently
	const numGoroutines = 10
	const recordsPerGoroutine = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < recordsPerGoroutine; j++ {
				server.RecordExecution(true, 100*time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Get metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.metricsHandler(w, req)

	body := w.Body.String()

	// Verify total count is correct
	expectedCount := numGoroutines * recordsPerGoroutine
	expectedLine := fmt.Sprintf("rpr_executions_total{status=\"success\"} %d", expectedCount)

	if !strings.Contains(body, expectedLine) {
		t.Errorf("Expected total count %d not found in metrics", expectedCount)
	}
}

func TestMetricsServer_CustomPort(t *testing.T) {
	// Test that server can be configured with custom port
	server := NewMetricsServer(9999)

	if server.port != 9999 {
		t.Errorf("Expected port 9999, got %d", server.port)
	}
}

func TestMetricsServer_Reset(t *testing.T) {
	server := NewMetricsServer(0)

	// Record some metrics
	server.RecordExecution(true, 100*time.Millisecond)
	server.RecordExecution(false, 200*time.Millisecond)

	// Reset metrics
	server.Reset()

	// Get metrics after reset
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.metricsHandler(w, req)

	body := w.Body.String()

	// Verify counters are reset
	if !strings.Contains(body, "rpr_executions_total{status=\"success\"} 0") {
		t.Error("Success counter not reset correctly")
	}

	if !strings.Contains(body, "rpr_executions_total{status=\"failure\"} 0") {
		t.Error("Failure counter not reset correctly")
	}
}
