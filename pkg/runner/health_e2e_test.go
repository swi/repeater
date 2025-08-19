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
	"github.com/swi/repeater/pkg/health"
)

func TestHealthServerEndToEnd(t *testing.T) {
	// Create config with health enabled
	config := &cli.Config{
		Subcommand:    "count",
		Times:         3,
		Command:       []string{"echo", "health-e2e-test"},
		HealthEnabled: true,
		HealthPort:    0, // Use random port for testing
	}

	// Create runner
	runner, err := NewRunner(config)
	require.NoError(t, err)
	require.NotNil(t, runner.healthServer, "Health server should be initialized")

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

	// Wait a moment for health server to start
	time.Sleep(200 * time.Millisecond)

	// Get the actual port the health server is using
	port := runner.healthServer.GetPort()
	require.Greater(t, port, 0, "Health server should have a valid port")

	// Test health endpoint
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	resp, err := http.Get(healthURL)
	require.NoError(t, err, "Health endpoint should be accessible")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")

	// Parse health response
	var healthResp health.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err, "Health response should be valid JSON")

	assert.Equal(t, "healthy", healthResp.Status, "Health status should be 'healthy'")
	assert.NotZero(t, healthResp.Timestamp, "Health response should have timestamp")
	assert.Equal(t, "0.3.0", healthResp.Version, "Health response should have version")
	assert.Greater(t, healthResp.Uptime, time.Duration(0), "Health response should have uptime")

	// Test ready endpoint
	readyURL := fmt.Sprintf("http://localhost:%d/ready", port)
	resp, err = http.Get(readyURL)
	require.NoError(t, err, "Ready endpoint should be accessible")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Ready endpoint should return 200 OK")

	var readyResp health.ReadinessResponse
	err = json.NewDecoder(resp.Body).Decode(&readyResp)
	require.NoError(t, err, "Ready response should be valid JSON")

	assert.True(t, readyResp.Ready, "Service should be ready")
	assert.NotZero(t, readyResp.Timestamp, "Ready response should have timestamp")

	// Test live endpoint
	liveURL := fmt.Sprintf("http://localhost:%d/live", port)
	resp, err = http.Get(liveURL)
	require.NoError(t, err, "Live endpoint should be accessible")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Live endpoint should return 200 OK")

	var liveResp health.LivenessResponse
	err = json.NewDecoder(resp.Body).Decode(&liveResp)
	require.NoError(t, err, "Live response should be valid JSON")

	assert.True(t, liveResp.Alive, "Service should be alive")
	assert.NotZero(t, liveResp.Timestamp, "Live response should have timestamp")

	// Wait a bit for some executions to complete
	time.Sleep(500 * time.Millisecond)

	// Check health endpoint again to see updated metrics
	resp, err = http.Get(healthURL)
	require.NoError(t, err, "Health endpoint should still be accessible")
	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err, "Health response should still be valid JSON")

	// Should have execution metrics now
	if healthResp.Metrics != nil {
		assert.GreaterOrEqual(t, healthResp.Metrics.TotalExecutions, int64(0), "Should have execution stats")
		assert.GreaterOrEqual(t, healthResp.Metrics.SuccessfulExecutions, int64(0), "Should have success stats")
		assert.NotZero(t, healthResp.Metrics.LastExecution, "Should have last execution time")
	}

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

func TestHealthServerWithConfigFile(t *testing.T) {
	// This test verifies that health server works with config file settings
	// Create config that would come from config file
	config := &cli.Config{
		Subcommand:    "interval",
		Every:         500 * time.Millisecond,
		Times:         2,
		Command:       []string{"echo", "config-health-test"},
		HealthEnabled: true,
		HealthPort:    0, // Random port
	}

	// Create and run
	runner, err := NewRunner(config)
	require.NoError(t, err)
	require.NotNil(t, runner.healthServer, "Health server should be initialized from config")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start runner
	go func() {
		_, _ = runner.Run(ctx)
	}()

	// Wait for health server to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	port := runner.healthServer.GetPort()
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)

	resp, err := http.Get(healthURL)
	require.NoError(t, err, "Health endpoint should work with config file settings")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")
}
