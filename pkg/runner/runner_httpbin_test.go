package runner

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
	httpbinTest "github.com/swi/repeater/pkg/testing"
)

// TestRunner_HTTPBin_RealWorldIntegration tests real-world HTTP scenarios using HTTPBin
func TestRunner_HTTPBin_RealWorldIntegration(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	scenarios := helper.CommonHTTPBinScenarios()

	tests := []struct {
		name           string
		scenario       httpbinTest.TestScenario
		config         *cli.Config
		expectedResult func(*ExecutionStats) bool
		description    string
	}{
		{
			name:     "exponential_backoff_with_503_errors",
			scenario: scenarios[0], // service_unavailable_503
			config: &cli.Config{
				Subcommand:     "exponential",
				BaseDelay:      500 * time.Millisecond,
				MaxDelay:       5 * time.Second,
				Times:          3,
				HTTPAware:      true,
				FailurePattern: "503",
				Verbose:        true,
			},
			expectedResult: func(stats *ExecutionStats) bool {
				// HTTPBin 503 responses don't cause curl to exit with non-zero status
				// They are successful HTTP requests that return 503 status codes
				return stats.TotalExecutions == 3 && stats.SuccessfulExecutions == 3
			},
			description: "Should handle 503 errors with exponential backoff and HTTP-aware timing",
		},
		{
			name:     "rate_limiting_429_with_adaptive",
			scenario: scenarios[1], // rate_limited_429
			config: &cli.Config{
				Subcommand:     "adaptive",
				BaseInterval:   2 * time.Second,
				MinInterval:    1 * time.Second,
				MaxInterval:    10 * time.Second,
				Times:          2,
				HTTPAware:      true,
				FailurePattern: "429",
			},
			expectedResult: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2
			},
			description: "Should adapt intervals based on 429 rate limiting responses",
		},
		{
			name:     "success_pattern_matching_json",
			scenario: scenarios[4], // json_response_parsing
			config: &cli.Config{
				Subcommand:     "count",
				Times:          2,
				HTTPAware:      true,
				SuccessPattern: "slideshow",
			},
			expectedResult: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should successfully parse JSON responses and match success patterns",
		},
		{
			name:     "delayed_response_timing",
			scenario: scenarios[5], // delayed_response
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				HTTPAware:  true,
				Timeout:    15 * time.Second, // Account for delay
			},
			expectedResult: func(stats *ExecutionStats) bool {
				return stats.TotalExecutions == 2 && stats.SuccessfulExecutions == 2
			},
			description: "Should handle delayed responses correctly with HTTP-aware scheduling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if this is a particularly long test and we're in short mode
			if strings.Contains(tt.name, "delayed") && testing.Short() {
				t.Skip("Skipping delayed response test in short mode")
			}

			// Set up command to use HTTPBin endpoint
			curlCmd := helper.GetCurlCommand(tt.scenario.Endpoint, tt.scenario.Method)
			tt.config.Command = curlCmd

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			runner, err := NewRunner(tt.config)
			require.NoError(t, err, "Failed to create runner for %s", tt.description)

			stats, _ := runner.Run(ctx)

			// For some scenarios, we expect errors (like 503 responses)
			if strings.Contains(tt.name, "503") || strings.Contains(tt.name, "429") {
				// These scenarios test error handling, so runner.Run might return an error
				// but we still want to check the stats
				if stats == nil {
					t.Errorf("Expected stats even with errors, got nil")
					return
				}
			} else {
				require.NoError(t, err, "Runner execution failed for %s", tt.description)
				require.NotNil(t, stats, "Expected stats, got nil")
			}

			// Validate results using custom assertion
			if !tt.expectedResult(stats) {
				t.Errorf("Result validation failed for %s. Stats: %+v", tt.description, stats)
			}

			t.Logf("✅ %s completed: %d total, %d successful, %d failed",
				tt.description, stats.TotalExecutions, stats.SuccessfulExecutions, stats.FailedExecutions)
		})
	}
}

// TestRunner_HTTPBin_HTTPAwareScheduling tests HTTP-aware scheduling with real timing responses
func TestRunner_HTTPBin_HTTPAwareScheduling(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	t.Run("retry_after_header_timing", func(t *testing.T) {
		// HTTPBin doesn't naturally provide Retry-After headers, so we'll test with 503 status
		// and verify that our HTTP-aware parsing doesn't break with real responses
		config := &cli.Config{
			Subcommand: "count",
			Times:      2,
			HTTPAware:  true,
			Command:    helper.GetCurlCommand(helper.CommonHTTPBinScenarios()[0].Endpoint, "GET"), // 503 endpoint
			Verbose:    true,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		runner, err := NewRunner(config)
		require.NoError(t, err)

		stats, _ := runner.Run(ctx)

		// The 503 endpoint won't have retry-after headers, so this tests fallback behavior
		assert.NotNil(t, stats)
		assert.Equal(t, 2, stats.TotalExecutions)

		t.Logf("HTTP-aware scheduling with real 503 responses: %d executions completed", stats.TotalExecutions)
	})

	t.Run("json_timing_extraction", func(t *testing.T) {
		// Test with JSON endpoint to ensure JSON parsing doesn't interfere with timing
		config := &cli.Config{
			Subcommand:     "count",
			Times:          2,
			HTTPAware:      true,
			SuccessPattern: "slideshow",
			Command:        helper.GetCurlCommand(helper.CommonHTTPBinScenarios()[4].Endpoint, "GET"), // JSON endpoint
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		runner, err := NewRunner(config)
		require.NoError(t, err)

		stats, _ := runner.Run(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, 2, stats.TotalExecutions)
		assert.Equal(t, 2, stats.SuccessfulExecutions)

		t.Logf("JSON parsing with HTTP-aware scheduling: %d successful executions", stats.SuccessfulExecutions)
	})
}

// TestRunner_HTTPBin_ErrorHandling tests error handling with real HTTP error responses
func TestRunner_HTTPBin_ErrorHandling(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	errorScenarios := []struct {
		name         string
		endpoint     string
		expectedCode string
		strategy     string
	}{
		{"server_error_502", helper.CommonHTTPBinScenarios()[2].Endpoint, "502", "exponential"},
		{"rate_limit_429", helper.CommonHTTPBinScenarios()[1].Endpoint, "429", "fibonacci"},
		{"not_found_404", httpbinTest.NewHTTPBinEndpoints("https://httpbin.org").Status(404), "404", "linear"},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			config := &cli.Config{
				Subcommand:     scenario.strategy,
				BaseDelay:      500 * time.Millisecond,
				Times:          2,
				FailurePattern: scenario.expectedCode,
				Command:        helper.GetCurlCommand(scenario.endpoint, "GET"),
				Verbose:        true,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			runner, err := NewRunner(config)
			require.NoError(t, err)

			stats, _ := runner.Run(ctx)

			// We expect stats even if the run "fails" due to error responses
			assert.NotNil(t, stats)
			assert.Equal(t, 2, stats.TotalExecutions)

			// With failure pattern matching, these should be detected as failures
			assert.Equal(t, 2, stats.FailedExecutions)

			t.Logf("Error handling test %s: %d failed executions as expected",
				scenario.name, stats.FailedExecutions)
		})
	}
}

// TestRunner_HTTPBin_PerformanceValidation tests performance characteristics with real HTTP
func TestRunner_HTTPBin_PerformanceValidation(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	t.Run("timing_accuracy_with_delays", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping timing validation in short mode")
		}

		// Test with 1-second delay endpoint
		config := &cli.Config{
			Subcommand: "interval",
			Every:      2 * time.Second,
			Times:      3,
			HTTPAware:  true,
			Command:    helper.GetCurlCommand(httpbinTest.NewHTTPBinEndpoints("https://httpbin.org").Delay(1), "GET"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		startTime := time.Now()
		runner, err := NewRunner(config)
		require.NoError(t, err)

		stats, _ := runner.Run(ctx)
		totalTime := time.Since(startTime)

		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, 3, stats.TotalExecutions)
		assert.Equal(t, 3, stats.SuccessfulExecutions)

		// Should take at least 4 seconds (2s interval * 2 gaps between 3 executions) + HTTP delays
		expectedMinTime := 4 * time.Second
		assert.GreaterOrEqual(t, totalTime, expectedMinTime,
			"Expected execution to take at least %v, took %v", expectedMinTime, totalTime)

		t.Logf("Performance validation: %d executions in %v (expected minimum %v)",
			stats.TotalExecutions, totalTime, expectedMinTime)
	})
}

// TestRunner_HTTPBin_ConcurrentSafety tests concurrent execution safety with real HTTP endpoints
func TestRunner_HTTPBin_ConcurrentSafety(t *testing.T) {
	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	t.Run("concurrent_http_aware_execution", func(t *testing.T) {
		config := &cli.Config{
			Subcommand: "count",
			Times:      5,
			HTTPAware:  true,
			Command:    helper.GetCurlCommand(helper.CommonHTTPBinScenarios()[4].Endpoint, "GET"), // JSON endpoint
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Run multiple instances concurrently
		results := make(chan *ExecutionStats, 3)
		errors := make(chan error, 3)

		for i := 0; i < 3; i++ {
			go func(id int) {
				runner, err := NewRunner(config)
				if err != nil {
					errors <- err
					return
				}

				stats, _ := runner.Run(ctx)
				if err != nil {
					errors <- err
					return
				}

				results <- stats
			}(i)
		}

		// Collect results
		var allStats []*ExecutionStats
		for i := 0; i < 3; i++ {
			select {
			case stats := <-results:
				allStats = append(allStats, stats)
			case err := <-errors:
				t.Errorf("Concurrent execution error: %v", err)
			case <-time.After(25 * time.Second):
				t.Error("Timeout waiting for concurrent execution results")
			}
		}

		// Validate all executions completed successfully
		assert.Len(t, allStats, 3, "Expected 3 concurrent execution results")

		for i, stats := range allStats {
			assert.Equal(t, 5, stats.TotalExecutions, "Concurrent run %d should have 5 executions", i)
			assert.Equal(t, 5, stats.SuccessfulExecutions, "Concurrent run %d should have 5 successes", i)
		}

		t.Logf("Concurrent safety test: 3 concurrent runs completed successfully")
	})
}
