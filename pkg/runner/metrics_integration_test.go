package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestMetricsServerIntegration(t *testing.T) {
	tests := []struct {
		name           string
		metricsEnabled bool
		metricsPort    int
		expectServer   bool
		expectEndpoint bool
	}{
		{
			name:           "metrics server enabled should start HTTP server",
			metricsEnabled: true,
			metricsPort:    0, // Use random port for testing
			expectServer:   true,
			expectEndpoint: true,
		},
		{
			name:           "metrics server disabled should not start HTTP server",
			metricsEnabled: false,
			metricsPort:    0,
			expectServer:   false,
			expectEndpoint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with metrics settings
			config := &cli.Config{
				Subcommand:     "count",
				Times:          2,
				Command:        []string{"echo", "metrics-test"},
				MetricsEnabled: tt.metricsEnabled,
				MetricsPort:    tt.metricsPort,
			}

			// Create runner
			runner, err := NewRunner(config)
			require.NoError(t, err)

			// Check if metrics server field exists and is properly initialized
			if tt.expectServer {
				assert.NotNil(t, runner.metricsServer, "Metrics server should be initialized when enabled")
			} else {
				assert.Nil(t, runner.metricsServer, "Metrics server should be nil when disabled")
			}

			// Run the basic functionality to ensure it still works
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err, "Runner should complete successfully")
			assert.Equal(t, 2, stats.TotalExecutions, "Should execute 2 times")
			assert.Equal(t, 2, stats.SuccessfulExecutions, "All executions should succeed")

			// This section will be implemented once metrics server integration is complete
		})
	}
}

func TestMetricsServerStatsUpdate(t *testing.T) {
	// This test will be implemented once metrics server integration is complete
	// It should test that execution metrics are properly recorded in the metrics endpoint

	// Create config with metrics enabled
	config := &cli.Config{
		Subcommand:     "count",
		Times:          3,
		Command:        []string{"echo", "metrics-stats-test"},
		MetricsEnabled: true,
		MetricsPort:    0, // Random port
	}

	// Create runner
	runner, err := NewRunner(config)
	require.NoError(t, err)

	// For now, just verify runner works
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stats, err := runner.Run(ctx)
	require.NoError(t, err, "Runner should complete successfully")
	assert.Equal(t, 3, stats.TotalExecutions, "Should execute 3 times")
	assert.Equal(t, 3, stats.SuccessfulExecutions, "All executions should succeed")

	// Start runner in goroutine for real-time metrics monitoring
	resultCh := make(chan ExecutionStats, 1)
	errorCh := make(chan error, 1)

	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stats, err := runner.Run(runCtx)
		if err != nil {
			errorCh <- err
			return
		}
		resultCh <- *stats
	}()

	// Get metrics server port
	require.NotNil(t, runner.metricsServer, "Metrics server should be initialized")
	port := runner.metricsServer.GetPort()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make HTTP requests to /metrics endpoint during execution
	client := &http.Client{Timeout: 1 * time.Second}
	metricsURL := fmt.Sprintf("http://localhost:%d/metrics", port)

	// Test /metrics endpoint during execution
	metricsResp, err := client.Get(metricsURL)
	require.NoError(t, err, "Metrics endpoint should be accessible")
	require.Equal(t, http.StatusOK, metricsResp.StatusCode)

	// Read metrics response
	defer func() { _ = metricsResp.Body.Close() }()

	// Wait for execution to complete and verify stats were updated
	select {
	case stats := <-resultCh:
		assert.Equal(t, 3, stats.TotalExecutions)
		assert.Equal(t, 3, stats.SuccessfulExecutions)

		// Make another metrics request to verify stats are updated
		metricsResp2, err := client.Get(metricsURL)
		require.NoError(t, err)
		defer func() { _ = metricsResp2.Body.Close() }()

		// Read and verify Prometheus metrics
		body, err := io.ReadAll(metricsResp2.Body)
		require.NoError(t, err, "Should read metrics response")

		metricsText := string(body)

		// Verify Prometheus metrics format and content
		// Note: metrics are cumulative across multiple requests, so we check for existence rather than exact counts
		assert.Contains(t, metricsText, "rpr_executions_total{status=\"success\"}")
		assert.Contains(t, metricsText, "rpr_executions_total{status=\"failure\"} 0")
		assert.Contains(t, metricsText, "rpr_execution_duration_seconds_count")
		assert.Contains(t, metricsText, "# TYPE rpr_executions_total counter")
		assert.Contains(t, metricsText, "# TYPE rpr_execution_duration_seconds histogram")

		// Verify the metrics show at least the executions we know happened
		assert.Contains(t, metricsText, "# HELP rpr_executions_total Total number of command executions")
		assert.Contains(t, metricsText, "# HELP rpr_execution_duration_seconds Duration of command executions")

	case err := <-errorCh:
		t.Fatalf("Runner execution failed: %v", err)
	case <-time.After(6 * time.Second):
		t.Fatal("Test timeout waiting for runner completion")
	}
}
