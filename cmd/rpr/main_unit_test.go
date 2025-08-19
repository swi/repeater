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
	"github.com/swi/repeater/pkg/runner"
)

// TestExitError tests the ExitError type and Error method
func TestExitError(t *testing.T) {
	tests := []struct {
		name     string
		exitErr  *ExitError
		expected string
	}{
		{
			name:     "basic error message",
			exitErr:  &ExitError{Code: 1, Message: "execution failed"},
			expected: "execution failed",
		},
		{
			name:     "usage error",
			exitErr:  &ExitError{Code: 2, Message: "invalid arguments"},
			expected: "invalid arguments",
		},
		{
			name:     "interrupt error",
			exitErr:  &ExitError{Code: 130, Message: "interrupted by signal"},
			expected: "interrupted by signal",
		},
		{
			name:     "empty message",
			exitErr:  &ExitError{Code: 1, Message: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.exitErr.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowHelp tests the showHelp function output
func TestShowHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call showHelp
	showHelp()

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout
	output, err := io.ReadAll(r)
	require.NoError(t, err)

	outputStr := string(output)

	// Verify key sections are present
	assert.Contains(t, outputStr, "Repeater (rpr) - Continuous Command Execution Tool")
	assert.Contains(t, outputStr, "USAGE:")
	assert.Contains(t, outputStr, "GLOBAL OPTIONS:")
	assert.Contains(t, outputStr, "EXECUTION MODES:")
	assert.Contains(t, outputStr, "MATHEMATICAL RETRY STRATEGIES:")
	assert.Contains(t, outputStr, "ADAPTIVE SCHEDULING:")
	assert.Contains(t, outputStr, "EXAMPLES:")
	assert.Contains(t, outputStr, "EXIT CODES:")

	// Verify specific commands are documented
	assert.Contains(t, outputStr, "interval, int, i")
	assert.Contains(t, outputStr, "exponential, exp")
	assert.Contains(t, outputStr, "fibonacci, fib")
	assert.Contains(t, outputStr, "adaptive, adapt, a")
	assert.Contains(t, outputStr, "rate-limit, rate, rl")

	// Verify examples are present
	assert.Contains(t, outputStr, "rpr interval --every 30s --times 10")
	assert.Contains(t, outputStr, "rpr exponential --base-delay 1s --attempts 5")
	assert.Contains(t, outputStr, "rpr adaptive --base-interval 1s --show-metrics")

	// Verify exit codes documentation
	assert.Contains(t, outputStr, "0   All commands succeeded")
	assert.Contains(t, outputStr, "1   Some commands failed")
	assert.Contains(t, outputStr, "2   Usage error")
	assert.Contains(t, outputStr, "130 Interrupted (Ctrl+C)")
}

// TestShowVersion tests the showVersion function output
func TestShowVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call showVersion
	showVersion()

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout
	output, err := io.ReadAll(r)
	require.NoError(t, err)

	outputStr := string(output)

	// Verify version output format
	assert.Contains(t, outputStr, "rpr version 0.5.0")
	assert.True(t, strings.HasPrefix(outputStr, "rpr version"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(outputStr), "0.5.0"))
}

// TestShowExecutionResults tests the showExecutionResults function
func TestShowExecutionResults(t *testing.T) {
	tests := []struct {
		name     string
		stats    *runner.ExecutionStats
		expected []string
	}{
		{
			name:     "nil stats should not panic",
			stats:    nil,
			expected: []string{}, // No output expected
		},
		{
			name: "successful execution stats",
			stats: &runner.ExecutionStats{
				TotalExecutions:      5,
				SuccessfulExecutions: 5,
				FailedExecutions:     0,
				Duration:             2*time.Second + 500*time.Millisecond,
			},
			expected: []string{
				"✅ Execution completed!",
				"📊 Statistics:",
				"Total executions: 5",
				"Successful: 5",
				"Failed: 0",
				"Duration: 2.5s",
			},
		},
		{
			name: "mixed success and failure stats",
			stats: &runner.ExecutionStats{
				TotalExecutions:      10,
				SuccessfulExecutions: 7,
				FailedExecutions:     3,
				Duration:             1*time.Minute + 30*time.Second,
			},
			expected: []string{
				"✅ Execution completed!",
				"📊 Statistics:",
				"Total executions: 10",
				"Successful: 7",
				"Failed: 3",
				"Duration: 1m30s",
				"⚠️  Some executions failed. Check command output above.",
			},
		},
		{
			name: "all failed execution stats",
			stats: &runner.ExecutionStats{
				TotalExecutions:      3,
				SuccessfulExecutions: 0,
				FailedExecutions:     3,
				Duration:             100 * time.Millisecond,
			},
			expected: []string{
				"✅ Execution completed!",
				"📊 Statistics:",
				"Total executions: 3",
				"Successful: 0",
				"Failed: 3",
				"Duration: 100ms",
				"⚠️  Some executions failed. Check command output above.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call showExecutionResults
			showExecutionResults(tt.stats)

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

			// For nil stats, output should be empty
			if tt.stats == nil {
				assert.Empty(t, strings.TrimSpace(outputStr))
			}
		})
	}
}

// TestShowExecutionInfo tests the showExecutionInfo function for all subcommands
func TestShowExecutionInfo(t *testing.T) {
	tests := []struct {
		name     string
		config   *cli.Config
		expected []string
	}{
		{
			name: "interval subcommand with times and duration",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
				Times:      5,
				For:        2 * time.Minute,
				Command:    []string{"curl", "api.com"},
			},
			expected: []string{
				"🕐 Interval execution: every 30s, 5 times, for 2m0s",
				"📋 Command: [curl api.com]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "count subcommand with interval",
			config: &cli.Config{
				Subcommand: "count",
				Times:      10,
				Every:      1 * time.Second,
				Command:    []string{"echo", "test"},
			},
			expected: []string{
				"🔢 Count execution: 10 times, every 1s",
				"📋 Command: [echo test]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "duration subcommand with interval",
			config: &cli.Config{
				Subcommand: "duration",
				For:        5 * time.Minute,
				Every:      10 * time.Second,
				Command:    []string{"ping", "google.com"},
			},
			expected: []string{
				"⏱️  Duration execution: for 5m0s, every 10s",
				"📋 Command: [ping google.com]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "adaptive subcommand with metrics",
			config: &cli.Config{
				Subcommand:   "adaptive",
				BaseInterval: 1 * time.Second,
				MinInterval:  100 * time.Millisecond,
				MaxInterval:  10 * time.Second,
				ShowMetrics:  true,
				Command:      []string{"curl", "api.example.com"},
			},
			expected: []string{
				"🧠 Adaptive execution: base interval 1s, range 100ms-10s, with metrics",
				"📋 Command: [curl api.example.com]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "exponential strategy with multiplier",
			config: &cli.Config{
				Subcommand: "exponential",
				BaseDelay:  1 * time.Second,
				MaxDelay:   30 * time.Second,
				Multiplier: 2.5,
				Command:    []string{"curl", "flaky-api.com"},
			},
			expected: []string{
				"📈 Exponential strategy: base delay 1s, max 30s, multiplier 2.5x",
				"📋 Command: [curl flaky-api.com]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "fibonacci strategy",
			config: &cli.Config{
				Subcommand: "fibonacci",
				BaseDelay:  500 * time.Millisecond,
				MaxDelay:   20 * time.Second,
				Command:    []string{"./retry-script.sh"},
			},
			expected: []string{
				"🌀 Fibonacci strategy: base delay 500ms, max 20s",
				"📋 Command: [./retry-script.sh]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "linear strategy",
			config: &cli.Config{
				Subcommand: "linear",
				Increment:  2 * time.Second,
				MaxDelay:   15 * time.Second,
				Command:    []string{"ping", "unreliable-host.com"},
			},
			expected: []string{
				"📏 Linear strategy: increment 2s, max 15s",
				"📋 Command: [ping unreliable-host.com]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "polynomial strategy with exponent",
			config: &cli.Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   1.5,
				MaxDelay:   60 * time.Second,
				Command:    []string{"complex", "command"},
			},
			expected: []string{
				"🔢 Polynomial strategy: base delay 1s, exponent 1.5, max 1m0s",
				"📋 Command: [complex command]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "decorrelated-jitter strategy",
			config: &cli.Config{
				Subcommand: "decorrelated-jitter",
				BaseDelay:  800 * time.Millisecond,
				Multiplier: 3.0,
				MaxDelay:   45 * time.Second,
				Command:    []string{"aws", "s3", "ls"},
			},
			expected: []string{
				"🎲 Decorrelated jitter strategy: base delay 800ms, multiplier 3.0x, max 45s",
				"📋 Command: [aws s3 ls]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "load-adaptive strategy with all targets",
			config: &cli.Config{
				Subcommand:   "load-adaptive",
				BaseInterval: 2 * time.Second,
				TargetCPU:    75.0,
				TargetMemory: 85.0,
				TargetLoad:   1.2,
				Command:      []string{"intensive", "task"},
			},
			expected: []string{
				"⚖️  Load-adaptive execution: base interval 2s, target CPU 75%, memory 85%, load 1.2",
				"📋 Command: [intensive task]",
				"🚀 Starting execution...",
			},
		},
		{
			name: "unknown subcommand should not crash",
			config: &cli.Config{
				Subcommand: "unknown-mode",
				Command:    []string{"some", "command"},
			},
			expected: []string{
				"📋 Command: [some command]",
				"🚀 Starting execution...",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call showExecutionInfo
			showExecutionInfo(tt.config)

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
