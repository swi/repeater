package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestMetricsServerEndToEnd(t *testing.T) {
	// Create config with metrics enabled
	config := &cli.Config{
		Subcommand:     "count",
		Times:          3,
		Command:        []string{"echo", "metrics-e2e-test"},
		MetricsEnabled: true,
		MetricsPort:    0, // Use random port for testing
	}

	// Create runner
	runner, err := NewRunner(config)
	require.NoError(t, err)
	require.NotNil(t, runner.metricsServer, "Metrics server should be initialized")

	// Run with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start runner in goroutine
	statsCh := make(chan *ExecutionStats, 1)
	errCh := make(chan error, 1)

	go func() {
		stats, err := runner.Run(ctx)
		statsCh <- stats
		errCh <- err
	}()

	// Wait a moment for metrics server to start
	time.Sleep(200 * time.Millisecond)

	// Get the actual port the metrics server is using
	port := runner.metricsServer.GetPort()
	require.Greater(t, port, 0, "Metrics server should have a valid port")

	// Test metrics endpoint
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", port)
	resp, err := http.Get(metricsURL)
	require.NoError(t, err, "Metrics endpoint should be accessible")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint should return 200 OK")
	assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", resp.Header.Get("Content-Type"), "Should return Prometheus format")

	// Read and parse metrics response
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read metrics response")

	metricsText := string(body)

	// Check for expected Prometheus metrics
	assert.Contains(t, metricsText, "# HELP rpr_executions_total Total number of command executions", "Should have execution counter help")
	assert.Contains(t, metricsText, "# TYPE rpr_executions_total counter", "Should have execution counter type")
	assert.Contains(t, metricsText, "rpr_executions_total{status=\"success\"}", "Should have success counter")
	assert.Contains(t, metricsText, "rpr_executions_total{status=\"failure\"}", "Should have failure counter")

	assert.Contains(t, metricsText, "# HELP rpr_execution_duration_seconds Duration of command executions", "Should have duration histogram help")
	assert.Contains(t, metricsText, "# TYPE rpr_execution_duration_seconds histogram", "Should have duration histogram type")

	assert.Contains(t, metricsText, "# HELP rpr_scheduler_interval_seconds Current scheduler interval", "Should have scheduler interval gauge")
	assert.Contains(t, metricsText, "# TYPE rpr_scheduler_interval_seconds gauge", "Should have scheduler interval type")

	// Wait a bit for some executions to complete
	time.Sleep(500 * time.Millisecond)

	// Check metrics endpoint again to see updated values
	resp, err = http.Get(metricsURL)
	require.NoError(t, err, "Metrics endpoint should still be accessible")
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read updated metrics response")

	updatedMetricsText := string(body)

	// Should have some successful executions recorded
	assert.Contains(t, updatedMetricsText, "rpr_executions_total{status=\"success\"}", "Should have success metrics")

	// Parse success count from metrics
	lines := strings.Split(updatedMetricsText, "\n")
	var successCount int
	for _, line := range lines {
		if strings.HasPrefix(line, "rpr_executions_total{status=\"success\"}") {
			_, err := fmt.Sscanf(line, "rpr_executions_total{status=\"success\"} %d", &successCount)
			if err == nil && successCount > 0 {
				break
			}
		}
	}
	assert.Greater(t, successCount, 0, "Should have recorded successful executions")

	// Wait for runner to complete
	select {
	case stats := <-statsCh:
		err := <-errCh
		require.NoError(t, err, "Runner should complete successfully")
		assert.Equal(t, 3, stats.TotalExecutions, "Should execute 3 times")
		assert.Equal(t, 3, stats.SuccessfulExecutions, "All executions should succeed")
	case <-time.After(15 * time.Second):
		t.Fatal("Runner did not complete within timeout")
	}
}

func TestMetricsServerWithAdaptiveScheduling(t *testing.T) {
	// Test metrics server with adaptive scheduling to verify scheduler interval recording
	config := &cli.Config{
		Subcommand:     "adaptive",
		BaseInterval:   100 * time.Millisecond,
		MinInterval:    50 * time.Millisecond,
		MaxInterval:    500 * time.Millisecond,
		Times:          3,
		Command:        []string{"echo", "adaptive-metrics-test"},
		MetricsEnabled: true,
		MetricsPort:    0, // Random port
	}

	// Create and run
	runner, err := NewRunner(config)
	require.NoError(t, err)
	require.NotNil(t, runner.metricsServer, "Metrics server should be initialized")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start runner
	go func() {
		_, _ = runner.Run(ctx)
	}()

	// Wait for metrics server to start
	time.Sleep(200 * time.Millisecond)

	// Test metrics endpoint
	port := runner.metricsServer.GetPort()
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", port)

	resp, err := http.Get(metricsURL)
	require.NoError(t, err, "Metrics endpoint should work with adaptive scheduling")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read metrics")

	metricsText := string(body)

	// Should have scheduler interval metrics
	assert.Contains(t, metricsText, "rpr_scheduler_interval_seconds", "Should have scheduler interval metrics")
}

func TestMetricsServerWithConfigFile(t *testing.T) {
	// This test verifies that metrics server works with config file settings
	config := &cli.Config{
		Subcommand:     "interval",
		Every:          500 * time.Millisecond,
		Times:          2,
		Command:        []string{"echo", "config-metrics-test"},
		MetricsEnabled: true,
		MetricsPort:    0, // Random port
	}

	// Create and run
	runner, err := NewRunner(config)
	require.NoError(t, err)
	require.NotNil(t, runner.metricsServer, "Metrics server should be initialized from config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start runner
	go func() {
		_, _ = runner.Run(ctx)
	}()

	// Wait for metrics server to start
	time.Sleep(100 * time.Millisecond)

	// Test metrics endpoint
	port := runner.metricsServer.GetPort()
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", port)

	resp, err := http.Get(metricsURL)
	require.NoError(t, err, "Metrics endpoint should work with config file settings")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint should return 200 OK")
}
