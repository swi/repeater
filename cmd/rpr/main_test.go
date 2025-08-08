package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

// TestAdaptiveSubcommandIntegration tests the integration between CLI and adaptive scheduler
func TestAdaptiveSubcommandIntegration(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, config *cli.Config)
	}{
		{
			name: "adaptive subcommand creates adaptive scheduler",
			args: []string{"adaptive", "--base-interval", "1s", "--times", "1", "--", "echo", "test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "adaptive", config.Subcommand)
				assert.Equal(t, time.Second, config.BaseInterval)
				assert.Equal(t, []string{"echo", "test"}, config.Command)
				assert.Equal(t, int64(1), config.Times)
			},
		},
		{
			name: "adaptive with custom parameters",
			args: []string{"adaptive", "--base-interval", "2s", "--min-interval", "500ms",
				"--max-interval", "10s", "--slow-threshold", "3.0", "--fast-threshold", "0.3",
				"--failure-threshold", "0.2", "--show-metrics", "--times", "1", "--", "echo", "test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "adaptive", config.Subcommand)
				assert.Equal(t, 2*time.Second, config.BaseInterval)
				assert.Equal(t, 500*time.Millisecond, config.MinInterval)
				assert.Equal(t, 10*time.Second, config.MaxInterval)
				assert.Equal(t, 3.0, config.SlowThreshold)
				assert.Equal(t, 0.3, config.FastThreshold)
				assert.Equal(t, 0.2, config.FailureThreshold)
				assert.True(t, config.ShowMetrics)
				assert.Equal(t, int64(1), config.Times)
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse args
			config, err := cli.ParseArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, cli.ValidateConfig(config))

			// Validate config
			if tt.validate != nil {
				tt.validate(t, config)
			}

			// This should fail initially because executeCommand doesn't handle "adaptive" subcommand
			err = executeCommand(config)

			// For now, we expect this to fail because adaptive isn't implemented yet
			// Once implemented, this should succeed
			if config.Subcommand == "adaptive" {
				// This test should fail initially (RED phase)
				assert.NoError(t, err, "adaptive subcommand should be handled by executeCommand")
			}
		})
	}
}

// TestAdaptiveExecutionFlow tests the full execution flow with adaptive scheduling
func TestAdaptiveExecutionFlow(t *testing.T) {
	// Skip this test if we're in a CI environment or if it takes too long
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &cli.Config{
		Subcommand:       "adaptive",
		BaseInterval:     100 * time.Millisecond,
		MinInterval:      50 * time.Millisecond,
		MaxInterval:      500 * time.Millisecond,
		SlowThreshold:    2.0,
		FastThreshold:    0.5,
		FailureThreshold: 0.3,
		ShowMetrics:      true,
		Command:          []string{"echo", "adaptive-test"},
		Times:            3, // Limit executions for test
	}

	// This should fail initially because the runner doesn't know about adaptive scheduling
	err := executeCommand(config)

	// For now, we expect this to fail (RED phase)
	// Once implemented, this should succeed and we can verify adaptive behavior
	assert.NoError(t, err, "adaptive execution should work end-to-end")
}

// TestShowExecutionInfoAdaptive tests that adaptive subcommand shows proper execution info
func TestShowExecutionInfoAdaptive(t *testing.T) {
	// Capture stdout to verify output
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	config := &cli.Config{
		Subcommand:       "adaptive",
		BaseInterval:     time.Second,
		MinInterval:      100 * time.Millisecond,
		MaxInterval:      10 * time.Second,
		SlowThreshold:    2.0,
		FastThreshold:    0.5,
		FailureThreshold: 0.3,
		ShowMetrics:      true,
		Command:          []string{"echo", "test"},
	}

	// This should fail initially because showExecutionInfo doesn't handle "adaptive"
	// We're testing that the function doesn't panic and eventually shows adaptive info
	showExecutionInfo(config)

	// For now, this will not show adaptive-specific info (RED phase)
	// Once implemented, we should see adaptive scheduling information
}

// TestBackoffSubcommandIntegration tests the integration between CLI and backoff scheduler
func TestBackoffSubcommandIntegration(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, config *cli.Config)
	}{
		{
			name: "backoff subcommand creates backoff scheduler",
			args: []string{"backoff", "--initial", "100ms", "--max", "5s", "--times", "2", "--", "echo", "test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "backoff", config.Subcommand)
				assert.Equal(t, 100*time.Millisecond, config.InitialInterval)
				assert.Equal(t, 5*time.Second, config.BackoffMax)
				assert.Equal(t, []string{"echo", "test"}, config.Command)
				assert.Equal(t, int64(2), config.Times)
			},
		},
		{
			name: "backoff with custom parameters",
			args: []string{"backoff", "--initial", "200ms", "--max", "10s", "--multiplier", "1.5", "--jitter", "0.1", "--times", "1", "--", "echo", "backoff-test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "backoff", config.Subcommand)
				assert.Equal(t, 200*time.Millisecond, config.InitialInterval)
				assert.Equal(t, 10*time.Second, config.BackoffMax)
				assert.Equal(t, 1.5, config.BackoffMultiplier)
				assert.Equal(t, 0.1, config.BackoffJitter)
				assert.Equal(t, int64(1), config.Times)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse args
			config, err := cli.ParseArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, cli.ValidateConfig(config))

			// Validate config
			if tt.validate != nil {
				tt.validate(t, config)
			}

			// Test execution
			err = executeCommand(config)
			assert.NoError(t, err, "backoff subcommand should be handled by executeCommand")
		})
	}
}

// TestLoadAdaptiveSubcommandIntegration tests the integration between CLI and load-adaptive scheduler
func TestLoadAdaptiveSubcommandIntegration(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		validate func(t *testing.T, config *cli.Config)
	}{
		{
			name: "load-adaptive subcommand creates load-aware scheduler",
			args: []string{"load-adaptive", "--base-interval", "1s", "--target-cpu", "60", "--target-memory", "70", "--times", "2", "--", "echo", "test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "load-adaptive", config.Subcommand)
				assert.Equal(t, time.Second, config.BaseInterval)
				assert.Equal(t, 60.0, config.TargetCPU)
				assert.Equal(t, 70.0, config.TargetMemory)
				assert.Equal(t, []string{"echo", "test"}, config.Command)
				assert.Equal(t, int64(2), config.Times)
			},
		},
		{
			name: "load-adaptive with custom parameters",
			args: []string{"load-adaptive", "--base-interval", "500ms", "--target-cpu", "80", "--target-memory", "90", "--target-load", "1.5", "--times", "1", "--", "echo", "load-test"},
			validate: func(t *testing.T, config *cli.Config) {
				assert.Equal(t, "load-adaptive", config.Subcommand)
				assert.Equal(t, 500*time.Millisecond, config.BaseInterval)
				assert.Equal(t, 80.0, config.TargetCPU)
				assert.Equal(t, 90.0, config.TargetMemory)
				assert.Equal(t, 1.5, config.TargetLoad)
				assert.Equal(t, int64(1), config.Times)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse args
			config, err := cli.ParseArgs(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NoError(t, cli.ValidateConfig(config))

			// Validate config
			if tt.validate != nil {
				tt.validate(t, config)
			}

			// Test execution
			err = executeCommand(config)
			assert.NoError(t, err, "load-adaptive subcommand should be handled by executeCommand")
		})
	}
}
