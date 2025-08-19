package main

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

// TestExecuteCommand tests the main executeCommand function
func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name        string
		config      *cli.Config
		expectError bool
		expectedErr string
	}{
		{
			name: "successful interval execution",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      100 * time.Millisecond,
				Times:      2,
				Command:    []string{"echo", "success"},
				Quiet:      true, // Suppress output for test
			},
			expectError: false,
		},
		{
			name: "successful count execution",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				Quiet:      true,
			},
			expectError: false,
		},
		{
			name: "successful duration execution",
			config: &cli.Config{
				Subcommand: "duration",
				For:        200 * time.Millisecond,
				Every:      50 * time.Millisecond,
				Command:    []string{"echo", "duration-test"},
				Quiet:      true,
			},
			expectError: false,
		},
		{
			name: "command failure should return error",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"false"}, // Always fails
				Quiet:      true,
			},
			expectError: true,
			expectedErr: "some commands failed",
		},
		{
			name: "nonexistent command should return error",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"nonexistent-command-12345"},
				Quiet:      true,
			},
			expectError: true,
		},
		{
			name: "exponential strategy execution",
			config: &cli.Config{
				Subcommand: "exponential",
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Times:      1, // Just 1 execution to test strategy
				Command:    []string{"echo", "exponential-test"},
				Quiet:      true,
			},
			expectError: false,
		},
		{
			name: "fibonacci strategy execution",
			config: &cli.Config{
				Subcommand: "fibonacci",
				BaseDelay:  10 * time.Millisecond,
				MaxDelay:   100 * time.Millisecond,
				Times:      1, // Just 1 execution to test strategy
				Command:    []string{"echo", "fibonacci-test"},
				Quiet:      true,
			},
			expectError: false,
		},
		{
			name: "linear strategy execution",
			config: &cli.Config{
				Subcommand: "linear",
				Increment:  10 * time.Millisecond,
				MaxDelay:   50 * time.Millisecond,
				Times:      1, // Just 1 execution to test strategy
				Command:    []string{"echo", "linear-test"},
				Quiet:      true,
			},
			expectError: false,
		},
		{
			name: "verbose mode execution",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "verbose-test"},
				Verbose:    true, // Test verbose output
			},
			expectError: false,
		},
		{
			name: "stats-only mode execution",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "stats-test"},
				StatsOnly:  true, // Test stats-only output
			},
			expectError: false,
		},
		{
			name: "stream mode execution",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "stream-test"},
				Stream:     true, // Test stream mode
			},
			expectError: false,
		},
		{
			name: "adaptive strategy execution",
			config: &cli.Config{
				Subcommand:   "adaptive",
				BaseInterval: 50 * time.Millisecond,
				MinInterval:  20 * time.Millisecond,
				MaxInterval:  200 * time.Millisecond,
				Times:        2,
				Command:      []string{"echo", "adaptive-test"},
				Quiet:        true,
			},
			expectError: false,
		},
		{
			name: "load-adaptive strategy execution",
			config: &cli.Config{
				Subcommand:   "load-adaptive",
				BaseInterval: 50 * time.Millisecond,
				TargetCPU:    70.0,
				TargetMemory: 80.0,
				Times:        1,
				Command:      []string{"echo", "load-adaptive-test"},
				Quiet:        true,
			},
			expectError: false,
		},
		{
			name: "rate-limit strategy execution",
			config: &cli.Config{
				Subcommand: "rate-limit",
				RateSpec:   "10/1m", // 10 per minute
				Times:      1,
				Command:    []string{"echo", "rate-limit-test"},
				Quiet:      true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set a timeout to prevent hanging tests
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Create a channel to receive the result
			errChan := make(chan error, 1)

			// Run executeCommand in a goroutine with timeout
			go func() {
				err := executeCommand(tt.config)
				errChan <- err
			}()

			// Wait for result or timeout
			select {
			case err := <-errChan:
				if tt.expectError {
					assert.Error(t, err, "Expected error for test case: %s", tt.name)
					if tt.expectedErr != "" {
						assert.Contains(t, err.Error(), tt.expectedErr)
					}
				} else {
					assert.NoError(t, err, "Expected no error for test case: %s", tt.name)
				}
			case <-ctx.Done():
				t.Fatalf("Test timed out for case: %s", tt.name)
			}
		})
	}
}

// TestExecuteCommandExitCodes tests specific exit code scenarios
func TestExecuteCommandExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		config       *cli.Config
		expectedCode int
	}{
		{
			name: "successful execution should not return ExitError",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "success"},
				Quiet:      true,
			},
			expectedCode: 0, // No error expected
		},
		{
			name: "command failure should return exit code 1",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"false"},
				Quiet:      true,
			},
			expectedCode: 1,
		},
		{
			name: "runner creation failure should return exit code 1",
			config: &cli.Config{
				Subcommand: "invalid-subcommand",
				Command:    []string{"echo", "test"},
				Quiet:      true,
			},
			expectedCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCommand(tt.config)

			if tt.expectedCode == 0 {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				// The executeCommand function returns different error types
				// For runner creation failures, it returns a regular error, not ExitError
				// The main() function would convert this to ExitError
				if exitErr, ok := err.(*ExitError); ok {
					assert.Equal(t, tt.expectedCode, exitErr.Code)
				} else {
					// This is expected for runner creation failures
					assert.Contains(t, err.Error(), "failed to create runner")
				}
			}
		})
	}
}

// TestMainLevelIntegration tests main-level functionality without calling main() directly
func TestMainLevelIntegration(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectHelp  bool
		expectVer   bool
		expectError bool
	}{
		{
			name:       "help flag should trigger help display",
			args:       []string{"--help"},
			expectHelp: true,
		},
		{
			name:       "help shorthand should trigger help display",
			args:       []string{"-h"},
			expectHelp: true,
		},
		{
			name:      "version flag should trigger version display",
			args:      []string{"--version"},
			expectVer: true,
		},
		{
			name:      "version shorthand should trigger version display",
			args:      []string{"-v"},
			expectVer: true,
		},
		{
			name:        "invalid arguments should return error",
			args:        []string{"--invalid-flag"},
			expectError: true,
		},
		{
			name:        "missing command should return error",
			args:        []string{"interval", "--every", "1s"},
			expectError: true,
		},
		{
			name: "valid configuration should parse successfully",
			args: []string{"interval", "--every", "1s", "--times", "1", "--", "echo", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test CLI parsing (this is what main() does first)
			config, err := cli.ParseArgs(tt.args)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Test special flag handling (this is what main() does next)
			if tt.expectHelp {
				assert.True(t, config.Help)
				// In main(), this would call showHelp() and return
				return
			}

			if tt.expectVer {
				assert.True(t, config.Version)
				// In main(), this would call showVersion() and return
				return
			}

			// Test configuration validation (this is what main() does next)
			err = cli.ValidateConfig(config)
			assert.NoError(t, err, "Valid configuration should pass validation")

			// Test that executeCommand can handle the config
			// (we don't run it fully to avoid test execution overhead)
			assert.NotNil(t, config.Command, "Command should be set for execution")
		})
	}
}

// TestConfigFileIntegrationExecution tests config file + execution integration
func TestConfigFileIntegrationExecution(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test-config.toml"

	configContent := `
[defaults]
timeout = "30s"
max_retries = 3
log_level = "info"

[observability]
metrics_enabled = false
health_enabled = false
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test full integration: parse args -> apply config -> validate -> execute
	args := []string{"--config", configFile, "count", "--times", "1", "--", "echo", "config-test"}
	config, err := cli.ParseArgs(args)
	require.NoError(t, err)

	// Apply config file (this is what main() does)
	err = applyConfigFile(config)
	require.NoError(t, err)

	// Verify config file was applied
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "info", config.LogLevel)
	assert.False(t, config.MetricsEnabled)
	assert.False(t, config.HealthEnabled)

	// Validate config (this is what main() does)
	err = cli.ValidateConfig(config)
	require.NoError(t, err)

	// Execute (this is what main() does) - with quiet mode for test
	config.Quiet = true
	err = executeCommand(config)
	assert.NoError(t, err, "Config file integration execution should succeed")
}

// TestStreamingBehavior tests Unix pipeline behavior
func TestStreamingBehavior(t *testing.T) {
	tests := []struct {
		name           string
		config         *cli.Config
		expectedStream bool
	}{
		{
			name: "quiet mode should not enable streaming",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				Quiet:      true,
			},
			expectedStream: false,
		},
		{
			name: "stats-only mode should not enable streaming",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				StatsOnly:  true,
			},
			expectedStream: false,
		},
		{
			name: "explicit stream mode should enable streaming",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				Stream:     true,
			},
			expectedStream: true,
		},
		{
			name: "default mode should enable streaming for Unix pipeline",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "test"},
				// No special flags - should default to streaming
			},
			expectedStream: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			testConfig := *tt.config

			// Create a minimal mock execution to test streaming behavior
			// We'll capture the config after executeCommand applies its logic
			originalStream := testConfig.Stream

			// The executeCommand function applies Unix pipeline behavior
			// We can test this by checking if Stream is set correctly
			err := executeCommand(&testConfig)

			// For successful tests, check streaming behavior
			if err == nil || (err != nil && strings.Contains(err.Error(), "some commands failed")) {
				if tt.expectedStream {
					assert.True(t, testConfig.Stream, "Stream should be enabled for this configuration")
				} else if originalStream == false {
					// Only check if stream wasn't explicitly set
					assert.Equal(t, tt.expectedStream, testConfig.Stream)
				}
			}
		})
	}
}
