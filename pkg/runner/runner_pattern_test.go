package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestRunner_WithPatternMatching(t *testing.T) {
	tests := []struct {
		name          string
		config        *cli.Config
		expectedStats func(*ExecutionStats) bool
		description   string
	}{
		{
			name: "success pattern should override exit code",
			config: &cli.Config{
				Subcommand:     "interval",
				Every:          100 * time.Millisecond,
				Times:          3,
				SuccessPattern: "deployment successful",
				Command:        []string{"sh", "-c", "echo 'deployment successful'; exit 1"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.SuccessfulExecutions == 3 && stats.FailedExecutions == 0
			},
			description: "Commands that print success but exit with 1 should be treated as successful",
		},
		{
			name: "failure pattern should override exit code",
			config: &cli.Config{
				Subcommand:     "interval",
				Every:          100 * time.Millisecond,
				Times:          3,
				FailurePattern: "(?i)error",
				Command:        []string{"sh", "-c", "echo 'ERROR occurred'; exit 0"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.SuccessfulExecutions == 0 && stats.FailedExecutions == 3
			}, description: "Commands that print errors but exit with 0 should be treated as failed",
		},
		{
			name: "case insensitive pattern matching",
			config: &cli.Config{
				Subcommand:      "count",
				Times:           2,
				SuccessPattern:  "SUCCESS",
				CaseInsensitive: true,
				Command:         []string{"echo", "Process completed with success"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.SuccessfulExecutions == 2 && stats.FailedExecutions == 0
			},
			description: "Case insensitive matching should work correctly",
		},
		{
			name: "pattern precedence - failure overrides success",
			config: &cli.Config{
				Subcommand:     "count",
				Times:          2,
				SuccessPattern: "completed",
				FailurePattern: "error",
				Command:        []string{"sh", "-c", "echo 'Process completed with error'"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.SuccessfulExecutions == 0 && stats.FailedExecutions == 2
			},
			description: "Failure patterns should take precedence over success patterns",
		},
		{
			name: "no patterns should use exit code",
			config: &cli.Config{
				Subcommand: "count",
				Times:      2,
				Command:    []string{"echo", "test"},
			},
			expectedStats: func(stats *ExecutionStats) bool {
				return stats.SuccessfulExecutions == 2 && stats.FailedExecutions == 0
			},
			description: "When no patterns are configured, should use exit code",
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

func TestRunner_PatternMatchingWithAdaptiveScheduler(t *testing.T) {
	config := &cli.Config{
		Subcommand:     "adaptive",
		BaseInterval:   100 * time.Millisecond,
		MinInterval:    50 * time.Millisecond,
		MaxInterval:    500 * time.Millisecond,
		Times:          5,
		SuccessPattern: "success",
		FailurePattern: "error",
		Command:        []string{"sh", "-c", "if [ $((RANDOM % 2)) -eq 0 ]; then echo 'success'; else echo 'error'; fi"},
	}

	runner, err := NewRunner(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stats, err := runner.Run(ctx)
	require.NoError(t, err)

	// Should have executed 5 times with mixed results based on patterns
	assert.Equal(t, 5, stats.TotalExecutions)
	assert.Equal(t, 5, stats.SuccessfulExecutions+stats.FailedExecutions)
}
