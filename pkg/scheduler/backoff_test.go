package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExponentialBackoffScheduler tests the core exponential backoff algorithm
func TestExponentialBackoffScheduler(t *testing.T) {
	tests := []struct {
		name        string
		initial     time.Duration
		multiplier  float64
		maxInterval time.Duration
		jitter      float64
		failures    int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{
			name:        "basic exponential backoff",
			initial:     100 * time.Millisecond,
			multiplier:  2.0,
			maxInterval: 10 * time.Second,
			jitter:      0.0,
			failures:    3,
			expectedMin: 800 * time.Millisecond, // 100ms * 2^3
			expectedMax: 800 * time.Millisecond,
		},
		{
			name:        "with maximum cap",
			initial:     1 * time.Second,
			multiplier:  2.0,
			maxInterval: 5 * time.Second,
			jitter:      0.0,
			failures:    5,
			expectedMin: 5 * time.Second, // Capped at max
			expectedMax: 5 * time.Second,
		},
		{
			name:        "with jitter",
			initial:     1 * time.Second,
			multiplier:  2.0,
			maxInterval: 10 * time.Second,
			jitter:      0.1, // 10% jitter
			failures:    2,
			expectedMin: 3600 * time.Millisecond, // 4s - 10%
			expectedMax: 4400 * time.Millisecond, // 4s + 10%
		},
		{
			name:        "no failures should use initial interval",
			initial:     500 * time.Millisecond,
			multiplier:  2.0,
			maxInterval: 10 * time.Second,
			jitter:      0.0,
			failures:    0,
			expectedMin: 500 * time.Millisecond,
			expectedMax: 500 * time.Millisecond,
		},
		{
			name:        "single failure",
			initial:     200 * time.Millisecond,
			multiplier:  1.5,
			maxInterval: 10 * time.Second,
			jitter:      0.0,
			failures:    1,
			expectedMin: 300 * time.Millisecond, // 200ms * 1.5^1
			expectedMax: 300 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail initially - no ExponentialBackoffScheduler exists
			scheduler := NewExponentialBackoffScheduler(tt.initial, tt.multiplier, tt.maxInterval, tt.jitter)
			require.NotNil(t, scheduler)

			// Simulate failures
			for i := 0; i < tt.failures; i++ {
				scheduler.RecordFailure()
			}

			interval := scheduler.GetCurrentInterval()
			assert.GreaterOrEqual(t, interval, tt.expectedMin)
			assert.LessOrEqual(t, interval, tt.expectedMax)
		})
	}
}

// TestExponentialBackoffSchedulerReset tests that success resets the backoff
func TestExponentialBackoffSchedulerReset(t *testing.T) {
	initial := 100 * time.Millisecond
	multiplier := 2.0
	maxInterval := 10 * time.Second
	jitter := 0.0

	scheduler := NewExponentialBackoffScheduler(initial, multiplier, maxInterval, jitter)
	require.NotNil(t, scheduler)

	// Record several failures to increase backoff
	for i := 0; i < 3; i++ {
		scheduler.RecordFailure()
	}

	// Interval should be increased
	interval := scheduler.GetCurrentInterval()
	assert.Equal(t, 800*time.Millisecond, interval) // 100ms * 2^3

	// Record success - should reset to initial
	scheduler.RecordSuccess()
	interval = scheduler.GetCurrentInterval()
	assert.Equal(t, initial, interval)
}

// TestExponentialBackoffSchedulerInterface tests the scheduler interface methods
func TestExponentialBackoffSchedulerInterface(t *testing.T) {
	scheduler := NewExponentialBackoffScheduler(100*time.Millisecond, 2.0, 10*time.Second, 0.0)
	require.NotNil(t, scheduler)

	// Test Next() returns a channel
	nextChan := scheduler.Next()
	assert.NotNil(t, nextChan)

	// Test Stop() doesn't panic
	scheduler.Stop()
}

// TestExponentialBackoffSchedulerConcurrency tests concurrent access
func TestExponentialBackoffSchedulerConcurrency(t *testing.T) {
	scheduler := NewExponentialBackoffScheduler(100*time.Millisecond, 2.0, 10*time.Second, 0.0)
	require.NotNil(t, scheduler)

	// Test concurrent access to RecordFailure and GetCurrentInterval
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 10; i++ {
			scheduler.RecordFailure()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			_ = scheduler.GetCurrentInterval()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have some failures recorded
	interval := scheduler.GetCurrentInterval()
	assert.Greater(t, interval, 100*time.Millisecond)

	scheduler.Stop()
}

// TestExponentialBackoffSchedulerJitterRange tests jitter stays within bounds
func TestExponentialBackoffSchedulerJitterRange(t *testing.T) {
	initial := 1 * time.Second
	multiplier := 2.0
	maxInterval := 10 * time.Second
	jitter := 0.2 // 20% jitter

	scheduler := NewExponentialBackoffScheduler(initial, multiplier, maxInterval, jitter)
	require.NotNil(t, scheduler)

	// Record one failure to get 2s base interval
	scheduler.RecordFailure()

	// Test multiple times to check jitter range
	for i := 0; i < 10; i++ {
		interval := scheduler.GetCurrentInterval()
		// With 20% jitter on 2s, should be between 1.6s and 2.4s
		assert.GreaterOrEqual(t, interval, 1600*time.Millisecond)
		assert.LessOrEqual(t, interval, 2400*time.Millisecond)
	}

	scheduler.Stop()
}
