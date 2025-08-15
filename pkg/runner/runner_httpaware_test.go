package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestRunner_HTTPAwareIntegration(t *testing.T) {
	tests := []struct {
		name          string
		config        *cli.Config
		expectedStats func(*ExecutionStats) bool
		description   string
	}{
		{
			name: "http-aware with retry-after header",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				HTTPAware:  true,
				Verbose:    true, // Enable verbose to see HTTP timing info
				Command:    []string{"sh", "-c", "echo 'HTTP/1.1 503 Service Unavailable\r\nRetry-After: 1\r\n\r\n'; exit 0"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should parse HTTP responses and show timing info in verbose mode",
		},
		{
			name: "http-aware with json response",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				HTTPAware:  true,
				Command:    []string{"sh", "-c", "echo 'HTTP/1.1 429 Too Many Requests\r\n\r\n{\"retry_after\": 1}'; exit 0"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should parse JSON retry information from HTTP responses",
		},
		{
			name: "http-aware with non-http response",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				HTTPAware:  true,
				Command:    []string{"echo", "regular output"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should handle non-HTTP responses gracefully",
		},
		{
			name: "http-aware disabled",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				HTTPAware:  false, // Disabled
				Command:    []string{"sh", "-c", "echo 'HTTP/1.1 503 Service Unavailable\r\nRetry-After: 60\r\n\r\n'; exit 0"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should work normally when HTTP-aware is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err)

			assert.True(t, tt.expectedStats(stats),
				"Stats validation failed for %s: %+v", tt.description, stats)
		})
	}
}

func TestRunner_HTTPAwareWithAdaptiveScheduler(t *testing.T) {
	config := &cli.Config{
		Subcommand:   "adaptive",
		BaseInterval: 100 * time.Millisecond,
		MinInterval:  50 * time.Millisecond,  // Required for adaptive
		MaxInterval:  500 * time.Millisecond, // Required for adaptive
		Times:        3,
		HTTPAware:    true,
		Verbose:      true,
		Command:      []string{"sh", "-c", "echo 'HTTP/1.1 503 Service Unavailable\r\nRetry-After: 1\r\n\r\n'; exit 0"},
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stats, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should execute successfully with both adaptive and HTTP-aware functionality
	assert.Equal(t, 3, stats.TotalExecutions)
	assert.Equal(t, 3, stats.SuccessfulExecutions)
	assert.Equal(t, 0, stats.FailedExecutions)
}

func TestRunner_HTTPAwareConfigurationOptions(t *testing.T) {
	tests := []struct {
		name   string
		config *cli.Config
	}{
		{
			name: "http-aware with custom delays",
			config: &cli.Config{
				Subcommand:   "count",
				Times:        1,
				HTTPAware:    true,
				HTTPMaxDelay: 5 * time.Minute,
				HTTPMinDelay: 2 * time.Second,
				Command:      []string{"echo", "test"},
			},
		},
		{
			name: "http-aware with parsing options",
			config: &cli.Config{
				Subcommand:       "count",
				Times:            1,
				HTTPAware:        true,
				HTTPParseJSON:    false, // Disable JSON parsing
				HTTPParseHeaders: true,
				HTTPTrustClient:  true,
				Command:          []string{"echo", "test"},
			},
		},
		{
			name: "http-aware with custom fields",
			config: &cli.Config{
				Subcommand:       "count",
				Times:            1,
				HTTPAware:        true,
				HTTPCustomFields: []string{"custom_retry", "backoff_seconds"},
				Command:          []string{"echo", "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			// Verify that HTTP-aware configuration is properly set
			httpConfig := tt.config.GetHTTPAwareConfig()
			require.NotNil(t, httpConfig)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err)
			assert.Equal(t, 1, stats.TotalExecutions)
		})
	}
}
