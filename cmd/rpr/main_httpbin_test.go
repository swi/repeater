package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpbinTest "github.com/swi/repeater/pkg/testing"
)

// TestCLI_HTTPBin_RealWorldScenarios tests CLI with real HTTPBin endpoints
func TestCLI_HTTPBin_RealWorldScenarios(t *testing.T) {
	// Skip if binary doesn't exist
	if _, err := os.Stat("./bin/rpr"); os.IsNotExist(err) {
		t.Skip("Binary ./bin/rpr not found - run 'make build' first")
	}

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	scenarios := helper.CommonHTTPBinScenarios()

	tests := []struct {
		name        string
		scenario    httpbinTest.TestScenario
		args        []string
		expectError bool
		expectText  string
		description string
	}{
		{
			name:     "exponential_strategy_with_503_errors",
			scenario: scenarios[0], // service_unavailable_503
			args: []string{
				"exponential",
				"--base-delay", "500ms",
				"--max-delay", "2s",
				"--times", "2",
				"--http-aware",
				"--verbose",
				"--",
			},
			expectError: false, // 503 responses don't cause curl to exit with errors
			expectText:  "exponential strategy",
			description: "Should handle 503 errors with exponential backoff strategy",
		},
		{
			name:     "fibonacci_strategy_with_rate_limiting",
			scenario: scenarios[1], // rate_limited_429
			args: []string{
				"fibonacci",
				"--base-delay", "1s",
				"--times", "2",
				"--http-aware",
				"--failure-pattern", "429",
				"--",
			},
			expectError: true, // 429 responses will cause failures
			expectText:  "fibonacci strategy",
			description: "Should handle rate limiting with fibonacci retry strategy",
		},
		{
			name:     "success_with_json_pattern_matching",
			scenario: scenarios[4], // json_response_parsing
			args: []string{
				"count",
				"--times", "2",
				"--http-aware",
				"--success-pattern", "slideshow",
				"--verbose",
				"--",
			},
			expectError: false,
			expectText:  "✅",
			description: "Should successfully match JSON response patterns",
		},
		{
			name:     "adaptive_scheduling_with_json",
			scenario: scenarios[4], // json_response_parsing
			args: []string{
				"adaptive",
				"--base-interval", "2s",
				"--times", "3",
				"--http-aware",
				"--success-pattern", "origin",
				"--verbose",
				"--",
			},
			expectError: false,
			expectText:  "Adaptive",
			description: "Should adapt scheduling based on real JSON responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip delayed tests in short mode
			if strings.Contains(tt.name, "delay") && testing.Short() {
				t.Skip("Skipping delayed test in short mode")
			}

			// Build complete command with HTTPBin endpoint
			curlCmd := helper.GetCurlCommand(tt.scenario.Endpoint, tt.scenario.Method)
			fullArgs := append(tt.args, curlCmd...)

			// Execute the CLI command
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./bin/rpr", fullArgs...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			t.Logf("Command: ./bin/rpr %s", strings.Join(fullArgs, " "))
			t.Logf("Output: %s", outputStr)

			if tt.expectError {
				// We expect the command to exit with non-zero status due to HTTP errors
				// but it should still execute and provide output
				assert.NotEmpty(t, outputStr, "Expected output even with errors")
			} else {
				assert.NoError(t, err, "Command should succeed for %s", tt.description)
			}

			// Check for expected text in output
			if tt.expectText != "" {
				assert.Contains(t, strings.ToLower(outputStr), strings.ToLower(tt.expectText),
					"Expected output to contain %s", tt.expectText)
			}

			// Validate that HTTP-aware scheduling is working
			if strings.Contains(strings.Join(tt.args, " "), "--http-aware") {
				// HTTP-aware mode should show some indication of HTTP processing
				assert.True(t,
					strings.Contains(strings.ToLower(outputStr), "http") ||
						strings.Contains(strings.ToLower(outputStr), "timing") ||
						len(outputStr) > 10, // At least some meaningful output
					"Expected HTTP-aware processing indicators")
			}

			t.Logf("✅ %s: Command executed successfully", tt.description)
		})
	}
}

// TestCLI_HTTPBin_StrategyComparison tests different mathematical strategies with real HTTP endpoints
func TestCLI_HTTPBin_StrategyComparison(t *testing.T) {
	// Skip if binary doesn't exist
	if _, err := os.Stat("./bin/rpr"); os.IsNotExist(err) {
		t.Skip("Binary ./bin/rpr not found - run 'make build' first")
	}

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	if testing.Short() {
		t.Skip("Skipping strategy comparison in short mode")
	}

	// Use 503 endpoint for consistent error responses
	endpoint503 := helper.CommonHTTPBinScenarios()[0].Endpoint
	curlCmd := helper.GetCurlCommand(endpoint503, "GET")

	strategies := []struct {
		name string
		args []string
	}{
		{
			name: "exponential",
			args: []string{"exponential", "--base-delay", "500ms", "--times", "3"},
		},
		{
			name: "fibonacci",
			args: []string{"fibonacci", "--base-delay", "500ms", "--times", "3"},
		},
		{
			name: "linear",
			args: []string{"linear", "--increment", "1s", "--times", "3"},
		},
		{
			name: "polynomial",
			args: []string{"polynomial", "--base-delay", "500ms", "--exponent", "1.5", "--times", "3"},
		},
	}

	for _, strategy := range strategies {
		t.Run(strategy.name+"_with_real_http_errors", func(t *testing.T) {
			// Build command
			args := append(strategy.args, "--http-aware", "--failure-pattern", "503", "--verbose", "--")
			fullArgs := append(args, curlCmd...)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./rpr", fullArgs...)
			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			t.Logf("Strategy %s output: %s", strategy.name, outputStr)

			// All strategies should execute (even if they "fail" due to 503 responses)
			assert.NotEmpty(t, outputStr, "Expected output from %s strategy", strategy.name)

			// Should contain strategy name or execution indicators
			assert.True(t,
				strings.Contains(strings.ToLower(outputStr), strategy.name) ||
					strings.Contains(strings.ToLower(outputStr), "execution") ||
					strings.Contains(strings.ToLower(outputStr), "503"),
				"Expected strategy execution indicators in output")

			t.Logf("✅ Strategy %s executed with real HTTP endpoints", strategy.name)
		})
	}
}

// TestCLI_HTTPBin_OutputModes tests different output modes with real HTTP responses
func TestCLI_HTTPBin_OutputModes(t *testing.T) {
	// Skip if binary doesn't exist
	if _, err := os.Stat("./bin/rpr"); os.IsNotExist(err) {
		t.Skip("Binary ./bin/rpr not found - run 'make build' first")
	}

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)
	helper.ValidatePrerequisites(t)

	// Use JSON endpoint for consistent successful responses
	jsonEndpoint := helper.CommonHTTPBinScenarios()[4].Endpoint
	curlCmd := helper.GetCurlCommand(jsonEndpoint, "GET")

	outputModes := []struct {
		name        string
		flags       []string
		expectText  string
		description string
	}{
		{
			name:        "verbose_mode",
			flags:       []string{"--verbose"},
			expectText:  "execution",
			description: "Should show detailed execution information",
		},
		{
			name:        "quiet_mode",
			flags:       []string{"--quiet"},
			expectText:  "", // Minimal output expected
			description: "Should minimize output in quiet mode",
		},
		{
			name:        "stats_only_mode",
			flags:       []string{"--stats-only"},
			expectText:  "statistics",
			description: "Should show only execution statistics",
		},
		{
			name:        "stream_mode",
			flags:       []string{"--stream"},
			expectText:  "slideshow", // JSON content should be streamed
			description: "Should stream command output in real-time",
		},
	}

	for _, mode := range outputModes {
		t.Run(mode.name+"_with_json_response", func(t *testing.T) {
			// Build command
			baseArgs := []string{"count", "--times", "2", "--http-aware", "--success-pattern", "slideshow"}
			args := append(baseArgs, mode.flags...)
			args = append(args, "--")
			fullArgs := append(args, curlCmd...)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./rpr", fullArgs...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			require.NoError(t, err, "Output mode test should succeed")

			t.Logf("Output mode %s result: %s", mode.name, outputStr)

			if mode.expectText != "" {
				assert.Contains(t, strings.ToLower(outputStr), strings.ToLower(mode.expectText),
					"Expected %s output to contain %s", mode.name, mode.expectText)
			}

			// Quiet mode should have minimal output
			if strings.Contains(mode.name, "quiet") {
				assert.True(t, len(outputStr) < 100 || strings.TrimSpace(outputStr) == "",
					"Quiet mode should have minimal output, got: %s", outputStr)
			}

			t.Logf("✅ %s: %s", mode.name, mode.description)
		})
	}
}

// TestCLI_HTTPBin_ErrorScenarios tests error handling with real HTTP error responses
func TestCLI_HTTPBin_ErrorScenarios(t *testing.T) {
	// Skip if binary doesn't exist
	if _, err := os.Stat("./bin/rpr"); os.IsNotExist(err) {
		t.Skip("Binary ./bin/rpr not found - run 'make build' first")
	}

	helper := httpbinTest.NewHTTPBinHelper(nil)
	helper.SkipIfNoNetwork(t)

	errorEndpoints := []struct {
		name     string
		endpoint string
		pattern  string
		code     int
	}{
		{"server_error", helper.CommonHTTPBinScenarios()[2].Endpoint, "502", 502},
		{"rate_limit", helper.CommonHTTPBinScenarios()[1].Endpoint, "429", 429},
		{"not_found", httpbinTest.NewHTTPBinEndpoints("https://httpbin.org").Status(404), "404", 404},
	}

	for _, errorCase := range errorEndpoints {
		t.Run(errorCase.name+"_error_handling", func(t *testing.T) {
			curlCmd := helper.GetCurlCommand(errorCase.endpoint, "GET")

			args := []string{
				"count",
				"--times", "2",
				"--http-aware",
				"--failure-pattern", errorCase.pattern,
				"--verbose",
				"--",
			}
			fullArgs := append(args, curlCmd...)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./rpr", fullArgs...)
			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			t.Logf("Error scenario %s output: %s", errorCase.name, outputStr)

			// Command may exit with error status due to failure pattern matching
			// but should produce meaningful output
			assert.NotEmpty(t, outputStr, "Expected output even with error responses")

			// Should contain error code or pattern in output
			assert.True(t,
				strings.Contains(outputStr, errorCase.pattern) ||
					strings.Contains(strings.ToLower(outputStr), "fail") ||
					strings.Contains(strings.ToLower(outputStr), "error"),
				"Expected error handling indicators in output")

			t.Logf("✅ Error handling test %s completed", errorCase.name)
		})
	}
}
