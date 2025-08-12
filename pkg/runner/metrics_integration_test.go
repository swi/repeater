package runner

import (
	"context"
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

	// TODO: Once metrics server integration is implemented:
	// 1. Start runner in goroutine
	// 2. Make HTTP requests to /metrics endpoint during execution
	// 3. Verify that Prometheus metrics are updated in real-time
	// 4. Check that metrics include rpr_executions_total{status="success"}, etc.
}
