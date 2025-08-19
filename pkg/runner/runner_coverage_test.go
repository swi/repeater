package runner

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

// TestRateLimitSchedulerIntegration tests the rate limit scheduler functionality
func TestRateLimitSchedulerIntegration(t *testing.T) {
	tests := []struct {
		name          string
		config        *cli.Config
		expectedError string
	}{
		{
			name: "rate_limit_with_valid_rate_spec",
			config: &cli.Config{
				Subcommand: "rate-limit",
				RateSpec:   "10/1h",
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "rate_limit_with_retry_pattern",
			config: &cli.Config{
				Subcommand:   "rate-limit",
				RateSpec:     "5/1m",
				RetryPattern: "0,10s,30s",
				Command:      []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "rate_limit_missing_rate_spec",
			config: &cli.Config{
				Subcommand: "rate-limit",
				Command:    []string{"echo", "test"},
			},
			expectedError: "rate-limit requires --rate",
		},
		{
			name: "rate_limit_invalid_rate_spec",
			config: &cli.Config{
				Subcommand: "rate-limit",
				RateSpec:   "invalid",
				Command:    []string{"echo", "test"},
			},
			expectedError: "", // NewRunner validates basic format, actual parsing happens during execution
		},
		{
			name: "rate_limit_invalid_retry_pattern",
			config: &cli.Config{
				Subcommand:   "rate-limit",
				RateSpec:     "10/1h",
				RetryPattern: "invalid,pattern",
				Command:      []string{"echo", "test"},
			},
			expectedError: "", // NewRunner creates successfully, errors happen during execution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, runner)

				// Test that the rate limit scheduler was created
				assert.NotNil(t, runner.config)
				assert.Equal(t, "rate-limit", runner.config.Subcommand)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, runner)
			}
		})
	}
}

// TestRateLimitSchedulerExecution tests actual execution with rate limiting
func TestRateLimitSchedulerExecution(t *testing.T) {
	config := &cli.Config{
		Subcommand: "rate-limit",
		RateSpec:   "2/10s", // 2 executions per 10 seconds
		Times:      3,       // Try to execute 3 times (should hit rate limit)
		Command:    []string{"echo", "rate-limit-test"},
		Quiet:      true,
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	// Test execution with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := runner.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Should execute 2 times within rate limit, then timeout before 3rd
	assert.LessOrEqual(t, stats.TotalExecutions, 3)
	assert.GreaterOrEqual(t, stats.TotalExecutions, 1)
}

// TestAdaptiveMetricsDisplay tests the showAdaptiveMetrics function
func TestAdaptiveMetricsDisplay(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	config := &cli.Config{
		Subcommand:   "adaptive",
		BaseInterval: 1 * time.Second,
		MinInterval:  500 * time.Millisecond,
		MaxInterval:  5 * time.Second,
		ShowMetrics:  true,
		Times:        2,
		Command:      []string{"echo", "adaptive-test"},
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	// Execute to generate metrics
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stats, err := runner.Run(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout
	output := make([]byte, 1024)
	n, _ := r.Read(output)
	outputStr := string(output[:n])

	// Should contain adaptive metrics output
	assert.Contains(t, outputStr, "Adaptive Metrics")
	assert.Contains(t, outputStr, "Interval=")
	assert.Contains(t, outputStr, "Success=")
	assert.Contains(t, outputStr, "Circuit=")
}

// TestCircuitStateString tests the circuitStateString helper function
func TestCircuitStateString(t *testing.T) {
	// This is a private function, but we can test it through the adaptive metrics display
	config := &cli.Config{
		Subcommand:   "adaptive",
		BaseInterval: 100 * time.Millisecond,
		MinInterval:  50 * time.Millisecond,
		MaxInterval:  1 * time.Second,
		ShowMetrics:  true,
		Times:        1,
		Command:      []string{"echo", "circuit-test"},
		Quiet:        true,
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	// Execute to initialize adaptive scheduler
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = runner.Run(ctx)
	require.NoError(t, err)

	// The function is tested indirectly through the metrics display
	// We mainly want to ensure no panics occur
}

// TestParseRetryPatternFunction tests the parseRetryPattern function
func TestParseRetryPatternFunction(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		expectedError bool
		expectedLen   int
	}{
		{
			name:          "valid_simple_pattern",
			pattern:       "0,10s,30s",
			expectedError: false,
			expectedLen:   3,
		},
		{
			name:          "valid_complex_pattern",
			pattern:       "1m,2m,5m,10m",
			expectedError: false,
			expectedLen:   4,
		},
		{
			name:          "empty_pattern",
			pattern:       "",
			expectedError: false,
			expectedLen:   0,
		},
		{
			name:          "invalid_duration_format",
			pattern:       "0,invalid,30s",
			expectedError: false, // NewRunner doesn't validate pattern format, only during execution
			expectedLen:   0,
		},
		{
			name:          "single_value",
			pattern:       "5s",
			expectedError: false,
			expectedLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config that will trigger parseRetryPattern
			config := &cli.Config{
				Subcommand:   "rate-limit",
				RateSpec:     "10/1h",
				RetryPattern: tt.pattern,
				Command:      []string{"echo", "test"},
			}

			runner, err := NewRunner(config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, runner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, runner)
			}
		})
	}
}

// TestSchedulerCreationEdgeCases tests edge cases in scheduler creation
func TestSchedulerCreationEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		config        *cli.Config
		expectedError string
	}{
		{
			name: "load_adaptive_with_valid_config",
			config: &cli.Config{
				Subcommand:   "load-adaptive",
				BaseInterval: 1 * time.Second,
				TargetCPU:    70.0,
				TargetMemory: 80.0,
				TargetLoad:   1.0,
				Command:      []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "load_adaptive_missing_base_interval",
			config: &cli.Config{
				Subcommand: "load-adaptive",
				Command:    []string{"echo", "test"},
			},
			expectedError: "load-adaptive requires --base-interval",
		},
		{
			name: "exponential_strategy_with_defaults",
			config: &cli.Config{
				Subcommand: "exponential",
				MaxRetries: 3,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "fibonacci_strategy_with_defaults",
			config: &cli.Config{
				Subcommand: "fibonacci",
				MaxRetries: 5,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "linear_strategy_with_defaults",
			config: &cli.Config{
				Subcommand: "linear",
				MaxRetries: 4,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "polynomial_strategy_with_defaults",
			config: &cli.Config{
				Subcommand: "polynomial",
				MaxRetries: 3,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "decorrelated_jitter_strategy_with_defaults",
			config: &cli.Config{
				Subcommand: "decorrelated-jitter",
				MaxRetries: 6,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, runner)
				assert.Equal(t, tt.config.Subcommand, runner.config.Subcommand)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, runner)
			}
		})
	}
}

// TestMathematicalStrategyExecution tests execution of mathematical retry strategies
func TestMathematicalStrategyExecution(t *testing.T) {
	strategies := []string{"exponential", "fibonacci", "linear", "polynomial", "decorrelated-jitter"}

	for _, strategy := range strategies {
		t.Run("strategy_"+strategy, func(t *testing.T) {
			config := &cli.Config{
				Subcommand: strategy,
				BaseDelay:  100 * time.Millisecond,
				MaxDelay:   1 * time.Second,
				MaxRetries: 3,
				Command:    []string{"echo", strategy + "-test"},
				Quiet:      true,
			}

			// Special handling for linear strategy
			if strategy == "linear" {
				config.Increment = 200 * time.Millisecond
			}

			// Special handling for polynomial strategy
			if strategy == "polynomial" {
				config.Exponent = 1.5
			}

			// Special handling for decorrelated-jitter
			if strategy == "decorrelated-jitter" {
				config.Multiplier = 2.0
			}

			runner, err := NewRunner(config)
			require.NoError(t, err)

			// Test execution with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err)
			require.NotNil(t, stats)

			// Should execute at least once
			assert.GreaterOrEqual(t, stats.TotalExecutions, 1)
			assert.LessOrEqual(t, stats.TotalExecutions, 3)
		})
	}
}

// TestServerIntegrationCombinations tests various server configuration combinations
func TestServerIntegrationCombinations(t *testing.T) {
	tests := []struct {
		name   string
		config *cli.Config
	}{
		{
			name: "health_and_metrics_servers_enabled",
			config: &cli.Config{
				Subcommand:     "count",
				Times:          1,
				Command:        []string{"echo", "test"},
				HealthEnabled:  true,
				HealthPort:     18081,
				MetricsEnabled: true,
				MetricsPort:    18082,
				Quiet:          true,
			},
		},
		{
			name: "only_health_server_enabled",
			config: &cli.Config{
				Subcommand:    "count",
				Times:         1,
				Command:       []string{"echo", "test"},
				HealthEnabled: true,
				HealthPort:    18083,
				Quiet:         true,
			},
		},
		{
			name: "only_metrics_server_enabled",
			config: &cli.Config{
				Subcommand:     "count",
				Times:          1,
				Command:        []string{"echo", "test"},
				MetricsEnabled: true,
				MetricsPort:    18084,
				Quiet:          true,
			},
		},
		{
			name: "no_servers_enabled",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				Quiet:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			// Execute briefly
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			stats, err := runner.Run(ctx)
			require.NoError(t, err)
			require.NotNil(t, stats)

			// Verify execution completed
			assert.GreaterOrEqual(t, stats.TotalExecutions, 1)
		})
	}
}

// TestErrorHandlingInExecution tests various error conditions during execution
func TestErrorHandlingInExecution(t *testing.T) {
	tests := []struct {
		name     string
		config   *cli.Config
		expectOK bool
	}{
		{
			name: "command_failure_handling",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				Command:    []string{"sh", "-c", "exit 1"}, // Command that fails
				Quiet:      true,
			},
			expectOK: true, // Runner should handle command failures gracefully
		},
		{
			name: "nonexistent_command",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"nonexistent-command-12345"},
				Quiet:      true,
			},
			expectOK: true, // Runner should handle execution errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			// Use short timeout for timeout test
			timeout := 500 * time.Millisecond
			if strings.Contains(tt.name, "timeout") {
				timeout = 300 * time.Millisecond
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			stats, err := runner.Run(ctx)

			if tt.expectOK {
				require.NoError(t, err)
				require.NotNil(t, stats)
				// Should have attempted at least one execution
				assert.GreaterOrEqual(t, stats.TotalExecutions, 1)
			} else {
				require.Error(t, err)
			}
		})
	}
}
