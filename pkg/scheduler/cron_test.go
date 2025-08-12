package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronScheduler_Creation(t *testing.T) {
	tests := []struct {
		name        string
		expression  string
		timezone    string
		expectError bool
	}{
		{
			name:       "valid daily expression",
			expression: "0 9 * * *",
			timezone:   "UTC",
		},
		{
			name:       "valid weekday expression",
			expression: "0 9 * * 1-5",
			timezone:   "UTC",
		},
		{
			name:       "valid shortcut",
			expression: "@daily",
			timezone:   "UTC",
		},
		{
			name:       "valid with timezone",
			expression: "0 9 * * *",
			timezone:   "America/New_York",
		},
		{
			name:        "invalid expression",
			expression:  "invalid",
			timezone:    "UTC",
			expectError: true,
		},
		{
			name:        "invalid timezone",
			expression:  "0 9 * * *",
			timezone:    "Invalid/Timezone",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewCronScheduler(tt.expression, tt.timezone)

			if tt.expectError {
				assert.Error(t, err, "Expected error for invalid input")
				assert.Nil(t, scheduler, "Scheduler should be nil on error")
			} else {
				require.NoError(t, err, "Should create scheduler successfully")
				require.NotNil(t, scheduler, "Scheduler should not be nil")

				// Verify it implements the Scheduler interface
				var _ Scheduler = scheduler
			}
		})
	}
}

func TestCronScheduler_Interface(t *testing.T) {
	// Test that CronScheduler implements the Scheduler interface
	scheduler, err := NewCronScheduler("0 9 * * *", "UTC")
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	// Should implement Scheduler interface
	var _ Scheduler = scheduler

	// Test Next() method returns a channel
	nextCh := scheduler.Next()
	assert.NotNil(t, nextCh, "Next() should return a channel")

	// Test Stop() method doesn't panic
	assert.NotPanics(t, func() {
		scheduler.Stop()
	}, "Stop() should not panic")
}

func TestCronScheduler_NextExecution(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		timezone   string
		testTime   time.Time
		expectNext time.Time
	}{
		{
			name:       "daily at 9 AM - before time",
			expression: "0 9 * * *",
			timezone:   "UTC",
			testTime:   time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			expectNext: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		},
		{
			name:       "daily at 9 AM - after time",
			expression: "0 9 * * *",
			timezone:   "UTC",
			testTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			expectNext: time.Date(2024, 1, 2, 9, 0, 0, 0, time.UTC),
		},
		{
			name:       "every minute",
			expression: "* * * * *",
			timezone:   "UTC",
			testTime:   time.Date(2024, 1, 1, 12, 30, 30, 0, time.UTC),
			expectNext: time.Date(2024, 1, 1, 12, 31, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewCronScheduler(tt.expression, tt.timezone)
			require.NoError(t, err)

			// This test will fail until we implement the scheduler
			// For now, just verify the scheduler was created
			assert.NotNil(t, scheduler, "Scheduler should be created")

			// TODO: Once implemented, test:
			// 1. Set a mock time or use the scheduler's internal time calculation
			// 2. Get the next execution time
			// 3. Verify it matches the expected time
		})
	}
}

func TestCronScheduler_Timing(t *testing.T) {
	// Test that the scheduler actually waits for the correct time
	// This is an integration test that verifies timing behavior

	// Use a cron expression that should trigger soon
	// For testing, we'll use "every minute" but in practice this would be too slow
	// So we'll create a more frequent test case

	scheduler, err := NewCronScheduler("* * * * *", "UTC") // Every minute
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	// Start the scheduler
	nextCh := scheduler.Next()
	assert.NotNil(t, nextCh, "Next() should return a channel")

	// For this test, we'll just verify the channel exists
	// In a real test, we might wait for a short time to see if it triggers
	// but that would make tests slow and flaky

	// Clean up
	scheduler.Stop()

	// TODO: Once implemented, add timing tests:
	// 1. Test with expressions that trigger quickly (for testing)
	// 2. Verify the timing is accurate within reasonable bounds
	// 3. Test timezone handling
}

func TestCronScheduler_Stop(t *testing.T) {
	scheduler, err := NewCronScheduler("0 9 * * *", "UTC")
	require.NoError(t, err)

	// Start the scheduler
	nextCh := scheduler.Next()
	assert.NotNil(t, nextCh, "Next() should return a channel")

	// Stop the scheduler
	scheduler.Stop()

	// TODO: Once implemented, verify:
	// 1. The channel is closed or stops sending values
	// 2. Multiple calls to Stop() don't panic
	// 3. Calling Next() after Stop() behaves correctly
}

func TestCronScheduler_Timezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{
			name:     "UTC timezone",
			timezone: "UTC",
		},
		{
			name:     "New York timezone",
			timezone: "America/New_York",
		},
		{
			name:     "Tokyo timezone",
			timezone: "Asia/Tokyo",
		},
		{
			name:     "London timezone",
			timezone: "Europe/London",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewCronScheduler("0 9 * * *", tt.timezone)
			require.NoError(t, err, "Should create scheduler with valid timezone")
			require.NotNil(t, scheduler, "Scheduler should not be nil")

			// Verify the scheduler was created successfully
			nextCh := scheduler.Next()
			assert.NotNil(t, nextCh, "Next() should return a channel")

			scheduler.Stop()

			// TODO: Once implemented, verify:
			// 1. The scheduler respects the timezone for calculations
			// 2. DST transitions are handled correctly
			// 3. Time calculations are accurate for the specified timezone
		})
	}
}
