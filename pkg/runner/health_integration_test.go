package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
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

	// TODO: Once health server integration is implemented:
	// 1. Start runner in goroutine
	// 2. Make HTTP requests to health endpoint during execution
	// 3. Verify that execution stats are updated in real-time
	// 4. Check that metrics include TotalExecutions, SuccessfulExecutions, etc.
}
