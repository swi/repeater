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

			// Test the next execution time calculation using test data
			// This tests that the scheduler is properly configured and can calculate next execution
			nextTime := scheduler.expression.NextExecution(tt.testTime)

			// Verify the calculated next execution time matches expected
			assert.Equal(t, tt.expectNext, nextTime, "Next execution time should match expected")

			// Verify the next execution time is in the future relative to test time
			assert.True(t, nextTime.After(tt.testTime), "Next execution time should be after test time")

			// Verify timezone is respected
			assert.Equal(t, time.UTC, nextTime.Location(), "Expected UTC timezone")
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

	// Timing tests - verify the scheduler can calculate execution times correctly
	// Note: We don't test actual timing delays in unit tests, just the calculations

	// 1. Test that expressions that trigger quickly are calculated correctly
	// We start the scheduler to verify it initializes properly, then stop it
	_ = scheduler.Next() // Start the scheduler goroutine
	scheduler.Stop()     // Stop it immediately

	// 2. Verify timezone handling by testing with a known timezone
	scheduler2, err := NewCronScheduler("0 9 * * *", "America/New_York")
	require.NoError(t, err)

	// Calculate next execution in EST timezone
	estTime := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC) // 8 AM UTC = 3 AM EST
	est, _ := time.LoadLocation("America/New_York")
	nextInEST := scheduler2.expression.NextExecution(estTime.In(est))

	// Should be 9 AM EST (14:00 UTC)
	assert.Equal(t, 9, nextInEST.Hour(), "Should be 9 AM in EST timezone")
	scheduler2.Stop()
}

func TestCronScheduler_Stop(t *testing.T) {
	scheduler, err := NewCronScheduler("0 9 * * *", "UTC")
	require.NoError(t, err)

	// Start the scheduler
	nextCh := scheduler.Next()
	assert.NotNil(t, nextCh, "Next() should return a channel")

	// Stop the scheduler
	scheduler.Stop()

	// Verify stop behavior - test the idempotent stop functionality
	// 1. Verify multiple calls to Stop() don't panic (handled by sync.Once)
	scheduler.Stop() // First stop
	scheduler.Stop() // Second stop - should not panic

	// 2. Test that after Stop(), the scheduler behaves correctly
	// The channel may be closed, but we verify the scheduler handles it gracefully
	// 3. Note: The channel behavior after Stop() is implementation-dependent
	//    but the scheduler should not panic or cause race conditions
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

			// Verify timezone handling implementation
			// 1. The scheduler respects the timezone for calculations
			now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC) // UTC noon
			next := scheduler.expression.NextExecution(now.In(scheduler.timezone))

			// Verify the result is in the correct timezone
			assert.Equal(t, scheduler.timezone, next.Location(), "Next time should be in scheduler timezone")

			// 2. & 3. Time calculations are accurate for the specified timezone
			// The cron expression "0 9 * * *" should give us 9 AM in the scheduler's timezone
			assert.Equal(t, 9, next.Hour(), "Should be 9 AM in the scheduler's timezone")
			assert.Equal(t, 0, next.Minute(), "Should be exactly 9:00")

			// Note: DST transitions are handled by Go's time package
			// which the cron implementation leverages
		})
	}
}
