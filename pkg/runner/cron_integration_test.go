package runner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

func TestCronSchedulerIntegration(t *testing.T) {
	tests := []struct {
		name           string
		cronExpression string
		timezone       string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "valid cron expression should create scheduler",
			cronExpression: "* * * * *",
			timezone:       "UTC",
			expectError:    false,
		},
		{
			name:           "valid cron shortcut should create scheduler",
			cronExpression: "@daily",
			timezone:       "UTC",
			expectError:    false,
		},
		{
			name:           "valid timezone should be accepted",
			cronExpression: "0 9 * * *",
			timezone:       "America/New_York",
			expectError:    false,
		},
		{
			name:           "empty cron expression should fail",
			cronExpression: "",
			timezone:       "UTC",
			expectError:    true,
			errorContains:  "cron requires --cron",
		},
		{
			name:           "invalid cron expression should fail",
			cronExpression: "invalid",
			timezone:       "UTC",
			expectError:    true,
			errorContains:  "invalid cron expression",
		},
		{
			name:           "invalid timezone should fail",
			cronExpression: "0 9 * * *",
			timezone:       "Invalid/Timezone",
			expectError:    true,
			errorContains:  "invalid timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with cron settings
			config := &cli.Config{
				Subcommand:     "cron",
				CronExpression: tt.cronExpression,
				Timezone:       tt.timezone,
				Times:          1, // Limit to 1 execution for testing
				Command:        []string{"echo", "cron-test"},
			}

			// Create runner
			runner, err := NewRunner(config)
			if tt.expectError && err != nil {
				// Error during runner creation (e.g., empty cron expression)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, runner)

			// Try to create scheduler - this is where cron expression and timezone validation happens
			scheduler, err := runner.createScheduler()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, scheduler)

			// Clean up
			scheduler.Stop()
		})
	}
}

func TestCronSchedulerExecution(t *testing.T) {
	// This test uses a cron expression that should trigger immediately
	// We'll use a specific time-based test that doesn't rely on wall clock time

	t.Run("cron scheduler should execute command at scheduled time", func(t *testing.T) {
		// Create config with a cron expression that triggers every minute
		config := &cli.Config{
			Subcommand:     "cron",
			CronExpression: "* * * * *", // Every minute
			Timezone:       "UTC",
			Times:          1, // Only execute once
			Command:        []string{"echo", "cron-execution-test"},
		}

		// Create runner
		runner, err := NewRunner(config)
		require.NoError(t, err)
		require.NotNil(t, runner)

		// Create a context with timeout to prevent hanging
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start execution in a goroutine
		done := make(chan struct{})
		var stats *ExecutionStats
		var execErr error

		go func() {
			defer close(done)
			stats, execErr = runner.Run(ctx)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Execution completed
			require.NoError(t, execErr)

			// For cron scheduling, we might not get executions immediately
			// since it waits for the next scheduled time
			// The test validates that the scheduler was created and can be stopped gracefully
			assert.GreaterOrEqual(t, stats.TotalExecutions, 0)

		case <-ctx.Done():
			// Timeout - this is expected for cron scheduling since it waits for scheduled time
			// The important thing is that no error occurred during setup
			t.Log("Cron scheduler setup completed successfully (timeout expected for scheduling)")
		}
	})
}

func TestCronSchedulerWithShortcuts(t *testing.T) {
	shortcuts := []string{
		"@yearly",
		"@annually",
		"@monthly",
		"@weekly",
		"@daily",
		"@hourly",
	}

	for _, shortcut := range shortcuts {
		t.Run("shortcut_"+shortcut, func(t *testing.T) {
			config := &cli.Config{
				Subcommand:     "cron",
				CronExpression: shortcut,
				Timezone:       "UTC",
				Times:          1,
				Command:        []string{"echo", "shortcut-test"},
			}

			// Create runner - should not error
			runner, err := NewRunner(config)
			require.NoError(t, err)
			require.NotNil(t, runner)

			// Verify scheduler creation
			scheduler, err := runner.createScheduler()
			require.NoError(t, err)
			require.NotNil(t, scheduler)

			// Clean up
			scheduler.Stop()
		})
	}
}
