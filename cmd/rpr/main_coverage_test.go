package main

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

// TestShowSubcommandHelp tests the showSubcommandHelp function for all strategy subcommands
func TestShowSubcommandHelp(t *testing.T) {
	tests := []struct {
		name       string
		subcommand string
		expected   []string
	}{
		{
			name:       "exponential strategy help",
			subcommand: "exponential",
			expected: []string{
				"Exponential Backoff Strategy",
				"USAGE:",
				"rpr exponential [OPTIONS] -- <COMMAND>",
				"rpr exp [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"exponential backoff delays",
				"OPTIONS:",
				"--base-delay, -bd DURATION",
				"--multiplier FLOAT",
				"--max-delay, -md DURATION",
				"--attempts, -a COUNT",
				"EXAMPLES:",
				"rpr exp --base-delay 500ms",
				"rpr exponential --base-delay 1s",
			},
		},
		{
			name:       "fibonacci strategy help",
			subcommand: "fibonacci",
			expected: []string{
				"Fibonacci Backoff Strategy",
				"USAGE:",
				"rpr fibonacci [OPTIONS] -- <COMMAND>",
				"rpr fib [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"Fibonacci sequence delays",
				"1s, 1s, 2s, 3s, 5s, 8s",
				"OPTIONS:",
				"--base-delay, -bd DURATION",
				"--max-delay, -md DURATION",
				"--attempts, -a COUNT",
				"EXAMPLES:",
				"rpr fib --base-delay 1s",
				"rpr fibonacci --base-delay 500ms",
			},
		},
		{
			name:       "linear strategy help",
			subcommand: "linear",
			expected: []string{
				"Linear Backoff Strategy",
				"USAGE:",
				"rpr linear [OPTIONS] -- <COMMAND>",
				"rpr lin [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"linear incremental delays",
				"OPTIONS:",
				"--increment, -inc DURATION",
				"--max-delay, -md DURATION",
				"--attempts, -a COUNT",
				"EXAMPLES:",
				"rpr lin --increment 2s",
			},
		},
		{
			name:       "polynomial strategy help",
			subcommand: "polynomial",
			expected: []string{
				"Polynomial Backoff Strategy",
				"USAGE:",
				"rpr polynomial [OPTIONS] -- <COMMAND>",
				"rpr poly [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"polynomial growth delays",
				"OPTIONS:",
				"--base-delay, -bd DURATION",
				"--exponent, -exp FLOAT",
				"--max-delay, -md DURATION",
				"--attempts, -a COUNT",
				"EXAMPLES:",
				"rpr poly --base-delay 1s",
			},
		},
		{
			name:       "decorrelated-jitter strategy help",
			subcommand: "decorrelated-jitter",
			expected: []string{
				"Decorrelated Jitter Strategy",
				"USAGE:",
				"rpr decorrelated-jitter [OPTIONS] -- <COMMAND>",
				"rpr dj [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"AWS-recommended",
				"OPTIONS:",
				"--base-delay, -bd DURATION",
				"--multiplier FLOAT",
				"--max-delay, -md DURATION",
				"--attempts, -a COUNT",
				"EXAMPLES:",
				"rpr dj --base-delay 1s",
			},
		},
		{
			name:       "interval execution mode help",
			subcommand: "interval",
			expected: []string{
				"Interval Execution Mode",
				"USAGE:",
				"rpr interval [OPTIONS] -- <COMMAND>",
				"rpr int [OPTIONS] -- <COMMAND>",
				"rpr i [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"regular intervals",
				"OPTIONS:",
				"--every, -e DURATION",
				"--times, -t COUNT",
				"--for, -f DURATION",
				"EXAMPLES:",
				"rpr i -e 30s -t 10",
			},
		},
		{
			name:       "adaptive execution mode help",
			subcommand: "adaptive",
			expected: []string{
				"Adaptive Execution Mode",
				"USAGE:",
				"rpr adaptive [OPTIONS] -- <COMMAND>",
				"rpr adapt [OPTIONS] -- <COMMAND>",
				"rpr a [OPTIONS] -- <COMMAND>",
				"DESCRIPTION:",
				"AI-driven AIMD algorithm",
				"OPTIONS:",
				"--base-interval, -b DURATION",
				"--show-metrics, -m",
				"EXAMPLES:",
				"rpr adaptive --base-interval 1s",
			},
		},
		{
			name:       "unknown subcommand help",
			subcommand: "unknown",
			expected: []string{
				"Help not available for subcommand: unknown",
				"rpr --help",
			},
		},
		{
			name:       "empty subcommand help",
			subcommand: "",
			expected: []string{
				"Help not available for subcommand:",
				"rpr --help",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call showSubcommandHelp
			showSubcommandHelp(tt.subcommand)

			// Restore stdout and read captured output
			_ = w.Close()
			os.Stdout = oldStdout
			output, err := io.ReadAll(r)
			require.NoError(t, err)

			outputStr := string(output)

			// Verify expected content
			for _, expected := range tt.expected {
				assert.Contains(t, outputStr, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestExecuteCommandEdgeCases tests edge cases in executeCommand function
func TestExecuteCommandEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.Config
		wantErr bool
	}{
		{
			name: "empty command should handle gracefully",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{},
			},
			wantErr: true,
		},
		{
			name: "invalid duration values should be handled",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      0, // Invalid duration
				Times:      1,
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
		},
		{
			name: "valid minimal configuration should work",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "minimal"},
			},
			wantErr: false,
		},
		{
			name: "signal handling path should work",
			config: &cli.Config{
				Subcommand: "count",
				Times:      1,
				Command:    []string{"echo", "signal-test"},
				Timeout:    1 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCommand(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMainFunctionIntegration tests main function behavior through CLI parsing
func TestMainFunctionIntegration(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T)
	}{
		{
			name: "help flag should trigger help display",
			args: []string{"--help"},
			validate: func(t *testing.T) {
				config, err := cli.ParseArgs([]string{"--help"})
				require.NoError(t, err)
				assert.True(t, config.Help)
			},
		},
		{
			name: "version flag should trigger version display",
			args: []string{"--version"},
			validate: func(t *testing.T) {
				config, err := cli.ParseArgs([]string{"--version"})
				require.NoError(t, err)
				assert.True(t, config.Version)
			},
		},
		{
			name: "subcommand help should be detected",
			args: []string{"exponential", "--help"},
			validate: func(t *testing.T) {
				config, err := cli.ParseArgs([]string{"exponential", "--help"})
				require.NoError(t, err)
				assert.True(t, config.SubcommandHelp)
				assert.Equal(t, "exponential", config.Subcommand)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

// TestApplyConfigFileFunction tests the applyConfigFile function specifically
func TestApplyConfigFileFunction(t *testing.T) {
	// Create a temporary config file
	configContent := `[defaults]
timeout = "45s"
max_retries = 5
log_level = "debug"

[observability]
metrics_enabled = true
metrics_port = 9090
health_enabled = true
health_check_port = 8080
`
	tmpFile, err := os.CreateTemp("", "test-config-*.toml")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Test config file application
	config := &cli.Config{
		ConfigFile: tmpFile.Name(),
		Subcommand: "interval",
		Command:    []string{"echo", "test"},
	}

	err = applyConfigFile(config)
	assert.NoError(t, err)

	// Verify config was applied (only the fields that applyConfigFile actually sets)
	assert.Equal(t, 45*time.Second, config.Timeout)
	assert.Equal(t, 5, config.MaxRetries)
	assert.Equal(t, "debug", config.LogLevel)
	assert.True(t, config.MetricsEnabled)
	assert.Equal(t, 9090, config.MetricsPort)
	assert.True(t, config.HealthEnabled)
	assert.Equal(t, 8080, config.HealthPort)
}

// TestConfigFileErrors tests error handling in config file processing
func TestConfigFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.Config
		wantErr bool
	}{
		{
			name: "non-existent config file should error",
			config: &cli.Config{
				ConfigFile: "/non/existent/file.toml",
				Subcommand: "interval",
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
		},
		{
			name: "invalid config file format should error",
			config: func() *cli.Config {
				tmpFile, _ := os.CreateTemp("", "invalid-config-*.toml")
				_, _ = tmpFile.WriteString("invalid toml content [[[")
				_ = tmpFile.Close()
				return &cli.Config{
					ConfigFile: tmpFile.Name(),
					Subcommand: "interval",
					Command:    []string{"echo", "test"},
				}
			}(),
			wantErr: true,
		},
		{
			name: "empty config file path should not error",
			config: &cli.Config{
				ConfigFile: "",
				Subcommand: "interval",
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyConfigFile(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Clean up temp files
			if strings.Contains(tt.config.ConfigFile, "invalid-config") {
				_ = os.Remove(tt.config.ConfigFile)
			}
		})
	}
}

// TestVersionConstant tests that version constant is properly set
func TestVersionConstant(t *testing.T) {
	assert.NotEmpty(t, version, "Version constant should not be empty")
	assert.Contains(t, version, "0.5", "Version should contain current version number")

	// Test that version is used in showVersion
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showVersion()

	_ = w.Close()
	os.Stdout = oldStdout
	output, err := io.ReadAll(r)
	require.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, version)
}

// TestExitErrorImplementation tests ExitError implementation details
func TestExitErrorImplementation(t *testing.T) {
	err := &ExitError{Code: 1, Message: "test error"}

	// Test that it implements error interface
	var _ error = err

	// Test Error method behavior
	assert.Equal(t, "test error", err.Error())

	// Test with empty message
	emptyErr := &ExitError{Code: 2, Message: ""}
	assert.Equal(t, "", emptyErr.Error())

	// Test that Code field is accessible
	assert.Equal(t, 1, err.Code)
	assert.Equal(t, 2, emptyErr.Code)
}
