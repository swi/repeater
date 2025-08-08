package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestRunner_EndToEndExecution(t *testing.T) {
	tests := []struct {
		name           string
		config         *cli.Config
		expectedRuns   int
		maxDuration    time.Duration
		expectSuccess  bool
		validateOutput func(t *testing.T, stats *ExecutionStats)
	}{
		{
			name: "interval execution with times limit",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      100 * time.Millisecond,
				Times:      3,
				Command:    []string{"echo", "test"},
			},
			expectedRuns:  3,
			maxDuration:   1 * time.Second,
			expectSuccess: true,
			validateOutput: func(t *testing.T, stats *ExecutionStats) {
				assert.Equal(t, 3, stats.TotalExecutions)
				assert.Equal(t, 3, stats.SuccessfulExecutions)
				assert.Equal(t, 0, stats.FailedExecutions)
				assert.True(t, stats.Duration < 1*time.Second)
			},
		},
		{
			name: "interval execution with duration limit",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      50 * time.Millisecond,
				For:        200 * time.Millisecond,
				Command:    []string{"echo", "duration-test"},
			},
			expectedRuns:  4, // Should run ~4 times in 200ms with 50ms intervals
			maxDuration:   500 * time.Millisecond,
			expectSuccess: true,
			validateOutput: func(t *testing.T, stats *ExecutionStats) {
				assert.True(t, stats.TotalExecutions >= 3 && stats.TotalExecutions <= 5)
				assert.Equal(t, stats.TotalExecutions, stats.SuccessfulExecutions)
				assert.Equal(t, 0, stats.FailedExecutions)
			},
		},
		{
			name: "count execution",
			config: &cli.Config{
				Subcommand: "count",
				Times:      5,
				Every:      10 * time.Millisecond,
				Command:    []string{"echo", "count-test"},
			},
			expectedRuns:  5,
			maxDuration:   200 * time.Millisecond,
			expectSuccess: true,
			validateOutput: func(t *testing.T, stats *ExecutionStats) {
				assert.Equal(t, 5, stats.TotalExecutions)
				assert.Equal(t, 5, stats.SuccessfulExecutions)
				assert.Equal(t, 0, stats.FailedExecutions)
			},
		},
		{
			name: "duration execution",
			config: &cli.Config{
				Subcommand: "duration",
				For:        150 * time.Millisecond,
				Every:      30 * time.Millisecond,
				Command:    []string{"echo", "duration-only"},
			},
			expectedRuns:  5, // Should run ~5 times in 150ms with 30ms intervals
			maxDuration:   300 * time.Millisecond,
			expectSuccess: true,
			validateOutput: func(t *testing.T, stats *ExecutionStats) {
				assert.True(t, stats.TotalExecutions >= 4 && stats.TotalExecutions <= 6)
				assert.Equal(t, stats.TotalExecutions, stats.SuccessfulExecutions)
			},
		},
		{
			name: "execution with command failures",
			config: &cli.Config{
				Subcommand: "count",
				Times:      3,
				Command:    []string{"sh", "-c", "exit 1"}, // Always fails
			},
			expectedRuns:  3,
			maxDuration:   1 * time.Second,
			expectSuccess: true, // Runner should succeed even if commands fail
			validateOutput: func(t *testing.T, stats *ExecutionStats) {
				assert.Equal(t, 3, stats.TotalExecutions)
				assert.Equal(t, 0, stats.SuccessfulExecutions)
				assert.Equal(t, 3, stats.FailedExecutions)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.maxDuration)
			defer cancel()

			start := time.Now()
			stats, err := runner.Run(ctx)
			duration := time.Since(start)

			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, stats)
				assert.True(t, duration <= tt.maxDuration)
				tt.validateOutput(t, stats)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRunner_StopConditions(t *testing.T) {
	tests := []struct {
		name        string
		config      *cli.Config
		stopAfter   time.Duration
		expectStop  bool
		description string
	}{
		{
			name: "times limit stops execution",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      10 * time.Millisecond,
				Times:      2,
				Command:    []string{"echo", "times-limit"},
			},
			stopAfter:   100 * time.Millisecond,
			expectStop:  true,
			description: "Should stop after 2 executions",
		},
		{
			name: "duration limit stops execution",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      10 * time.Millisecond,
				For:        50 * time.Millisecond,
				Command:    []string{"echo", "duration-limit"},
			},
			stopAfter:   200 * time.Millisecond,
			expectStop:  true,
			description: "Should stop after 50ms duration",
		},
		{
			name: "both limits - times reached first",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      10 * time.Millisecond,
				Times:      2,
				For:        1 * time.Second, // Much longer than needed for 2 executions
				Command:    []string{"echo", "times-first"},
			},
			stopAfter:   200 * time.Millisecond,
			expectStop:  true,
			description: "Should stop when times limit reached first",
		},
		{
			name: "both limits - duration reached first",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      10 * time.Millisecond,
				Times:      100,                   // Much more than can run in 50ms
				For:        50 * time.Millisecond, // Short duration
				Command:    []string{"echo", "duration-first"},
			},
			stopAfter:   200 * time.Millisecond,
			expectStop:  true,
			description: "Should stop when duration limit reached first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.stopAfter)
			defer cancel()

			start := time.Now()
			stats, err := runner.Run(ctx)
			duration := time.Since(start)

			require.NoError(t, err)
			require.NotNil(t, stats)

			if tt.expectStop {
				// Should stop before the context timeout
				assert.True(t, duration < tt.stopAfter, tt.description)
			}
		})
	}
}

func TestRunner_SignalHandling(t *testing.T) {
	t.Run("graceful shutdown on SIGINT", func(t *testing.T) {
		config := &cli.Config{
			Subcommand: "interval",
			Every:      50 * time.Millisecond,
			Command:    []string{"echo", "signal-test"},
		}

		runner, err := NewRunner(config)
		require.NoError(t, err)

		// Create a context that we can cancel to simulate signal
		ctx, cancel := context.WithCancel(context.Background())

		// Start runner in goroutine
		var stats *ExecutionStats
		var runErr error
		done := make(chan struct{})

		go func() {
			defer close(done)
			stats, runErr = runner.Run(ctx)
		}()

		// Let it run for a bit
		time.Sleep(120 * time.Millisecond)

		// Simulate signal by canceling context
		cancel()

		// Wait for graceful shutdown
		select {
		case <-done:
			// Good, it shut down
		case <-time.After(1 * time.Second):
			t.Fatal("Runner did not shut down gracefully within timeout")
		}

		// Should have stopped gracefully
		assert.Error(t, runErr) // Context cancellation should cause error
		assert.Contains(t, runErr.Error(), "context canceled")
		require.NotNil(t, stats)
		assert.True(t, stats.TotalExecutions >= 2) // Should have run at least twice
	})
}

func TestRunner_ExecutionStatistics(t *testing.T) {
	t.Run("statistics collection", func(t *testing.T) {
		config := &cli.Config{
			Subcommand: "count",
			Times:      3,
			Every:      20 * time.Millisecond,
			Command:    []string{"echo", "stats-test"},
		}

		runner, err := NewRunner(config)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		start := time.Now()
		stats, err := runner.Run(ctx)
		totalDuration := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, stats)

		// Validate statistics
		assert.Equal(t, 3, stats.TotalExecutions)
		assert.Equal(t, 3, stats.SuccessfulExecutions)
		assert.Equal(t, 0, stats.FailedExecutions)
		assert.True(t, stats.Duration > 0)
		assert.True(t, stats.Duration <= totalDuration)
		assert.NotNil(t, stats.StartTime)
		assert.NotNil(t, stats.EndTime)
		assert.True(t, stats.EndTime.After(stats.StartTime))

		// Should have execution details
		assert.Len(t, stats.Executions, 3)
		for i, exec := range stats.Executions {
			assert.Equal(t, i+1, exec.ExecutionNumber)
			assert.Equal(t, 0, exec.ExitCode) // echo should succeed
			assert.True(t, exec.Duration > 0)
			assert.Contains(t, exec.Stdout, "stats-test")
			assert.Empty(t, exec.Stderr)
		}
	})

	t.Run("statistics with mixed success/failure", func(t *testing.T) {
		config := &cli.Config{
			Subcommand: "count",
			Times:      4,
			Command:    []string{"sh", "-c", "if [ $((RANDOM % 2)) -eq 0 ]; then echo 'success'; else echo 'failure' >&2; exit 1; fi"},
		}

		runner, err := NewRunner(config)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		stats, err := runner.Run(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Should have run all 4 times
		assert.Equal(t, 4, stats.TotalExecutions)
		assert.Equal(t, stats.SuccessfulExecutions+stats.FailedExecutions, stats.TotalExecutions)
		assert.Len(t, stats.Executions, 4)

		// Validate individual execution records
		for _, exec := range stats.Executions {
			assert.True(t, exec.ExecutionNumber >= 1 && exec.ExecutionNumber <= 4)
			assert.True(t, exec.Duration > 0)
			// Exit code should be either 0 (success) or 1 (failure)
			assert.True(t, exec.ExitCode == 0 || exec.ExitCode == 1)
		}
	})
}

func TestRunner_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "empty command",
			config: &cli.Config{
				Subcommand: "interval",
				Every:      1 * time.Second,
				Command:    []string{},
			},
			wantErr: true,
			errMsg:  "command cannot be empty",
		},
		{
			name: "invalid subcommand",
			config: &cli.Config{
				Subcommand: "invalid",
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
			errMsg:  "unknown subcommand",
		},
		{
			name: "interval without every",
			config: &cli.Config{
				Subcommand: "interval",
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
			errMsg:  "interval requires --every",
		},
		{
			name: "count without times",
			config: &cli.Config{
				Subcommand: "count",
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
			errMsg:  "count requires --times",
		},
		{
			name: "duration without for",
			config: &cli.Config{
				Subcommand: "duration",
				Command:    []string{"echo", "test"},
			},
			wantErr: true,
			errMsg:  "duration requires --for",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, err := NewRunner(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, runner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, runner)
			}
		})
	}
}
