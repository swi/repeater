package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
	healthpkg "github.com/swi/repeater/pkg/health"
)

func TestHealthServerIntegration(t *testing.T) {
	tests := []struct {
		name           string
		healthEnabled  bool
		healthPort     int
		expectServer   bool
		expectEndpoint bool
	}{
		{
			name:           "health server enabled should start HTTP server",
			healthEnabled:  true,
			healthPort:     0, // Use random port for testing
			expectServer:   true,
			expectEndpoint: true,
		},
		{
			name:           "health server disabled should not start HTTP server",
			healthEnabled:  false,
			healthPort:     0,
			expectServer:   false,
			expectEndpoint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with health settings
			config := &cli.Config{
				Subcommand:    "count",
				Times:         2,
				Command:       []string{"echo", "health-test"},
				HealthEnabled: tt.healthEnabled,
				HealthPort:    tt.healthPort,
			}

			// Create runner
			runner, err := NewRunner(config)
			require.NoError(t, err)

			// Check if health server field exists and is properly initialized
			if tt.expectServer {
				assert.NotNil(t, runner.healthServer, "Health server should be initialized when enabled")
			} else {
				assert.Nil(t, runner.healthServer, "Health server should be nil when disabled")
			}

			// Run the basic functionality to ensure it still works
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err, "Runner should complete successfully")
			assert.Equal(t, 2, stats.TotalExecutions, "Should execute 2 times")
			assert.Equal(t, 2, stats.SuccessfulExecutions, "All executions should succeed")

			// This section will be implemented once health server integration is complete
		})
	}
}

func TestHealthServerStatsUpdate(t *testing.T) {
	// This test will be implemented once health server integration is complete
	// It should test that execution stats are properly updated in the health endpoint

	// Create config with health enabled
	config := &cli.Config{
		Subcommand:    "count",
		Times:         3,
		Command:       []string{"echo", "stats-test"},
		HealthEnabled: true,
		HealthPort:    0, // Random port
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

	// Start runner in goroutine for real-time health monitoring
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

	// Get health server port
	require.NotNil(t, runner.healthServer, "Health server should be initialized")
	port := runner.healthServer.GetPort()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Make HTTP requests to health endpoints during execution
	client := &http.Client{Timeout: 1 * time.Second}
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Test /health endpoint
	healthResp, err := client.Get(baseURL + "/health")
	require.NoError(t, err, "Health endpoint should be accessible")
	require.Equal(t, http.StatusOK, healthResp.StatusCode)

	var health healthpkg.HealthResponse
	err = json.NewDecoder(healthResp.Body).Decode(&health)
	require.NoError(t, err, "Should decode health response")
	_ = healthResp.Body.Close()

	assert.Equal(t, "healthy", health.Status)
	assert.NotZero(t, health.Uptime)

	// Test /ready endpoint
	readyResp, err := client.Get(baseURL + "/ready")
	require.NoError(t, err, "Ready endpoint should be accessible")
	require.Equal(t, http.StatusOK, readyResp.StatusCode)
	_ = readyResp.Body.Close()

	// Test /live endpoint
	liveResp, err := client.Get(baseURL + "/live")
	require.NoError(t, err, "Live endpoint should be accessible")
	require.Equal(t, http.StatusOK, liveResp.StatusCode)
	_ = liveResp.Body.Close()

	// Wait for execution to complete and verify stats were updated
	select {
	case stats := <-resultCh:
		assert.Equal(t, 3, stats.TotalExecutions)
		assert.Equal(t, 3, stats.SuccessfulExecutions)

		// Make another health request to verify stats are updated
		healthResp2, err := client.Get(baseURL + "/health")
		require.NoError(t, err)

		var health2 healthpkg.HealthResponse
		err = json.NewDecoder(healthResp2.Body).Decode(&health2)
		require.NoError(t, err)
		_ = healthResp2.Body.Close()

		// Verify execution stats are reflected in health response
		require.NotNil(t, health2.Metrics, "Health response should include metrics")
		assert.Equal(t, int64(3), health2.Metrics.TotalExecutions)
		assert.Equal(t, int64(3), health2.Metrics.SuccessfulExecutions)
		assert.Equal(t, int64(0), health2.Metrics.FailedExecutions)
		assert.NotZero(t, health2.Metrics.LastExecution)

	case err := <-errorCh:
		t.Fatalf("Runner execution failed: %v", err)
	case <-time.After(6 * time.Second):
		t.Fatal("Test timeout waiting for runner completion")
	}
}
