package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntervalScheduler_Creation(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		jitter   float64
		wantErr  bool
	}{
		{
			name:     "valid 1 second interval",
			interval: time.Second,
			jitter:   0,
			wantErr:  false,
		},
		{
			name:     "valid 100ms interval",
			interval: 100 * time.Millisecond,
			jitter:   0,
			wantErr:  false,
		},
		{
			name:     "valid interval with 10% jitter",
			interval: time.Second,
			jitter:   0.1,
			wantErr:  false,
		},
		{
			name:     "zero interval should error",
			interval: 0,
			jitter:   0,
			wantErr:  true,
		},
		{
			name:     "negative interval should error",
			interval: -time.Second,
			jitter:   0,
			wantErr:  true,
		},
		{
			name:     "invalid jitter should error",
			interval: time.Second,
			jitter:   1.5, // >100%
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewIntervalScheduler(tt.interval, tt.jitter, false)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, scheduler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scheduler)

				// Cleanup
				if scheduler != nil {
					scheduler.Stop()
				}
			}
		})
	}
}

func TestIntervalScheduler_Timing(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		count     int
		tolerance time.Duration
	}{
		{
			name:      "100ms interval timing",
			interval:  100 * time.Millisecond,
			count:     5,
			tolerance: 20 * time.Millisecond,
		},
		{
			name:      "500ms interval timing",
			interval:  500 * time.Millisecond,
			count:     3,
			tolerance: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewIntervalScheduler(tt.interval, 0, false)
			require.NoError(t, err)
			defer scheduler.Stop()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			start := time.Now()
			ticks := 0

			for ticks < tt.count {
				select {
				case <-scheduler.Next():
					ticks++
				case <-ctx.Done():
					t.Fatalf("timeout waiting for ticks, got %d/%d", ticks, tt.count)
				}
			}

			elapsed := time.Since(start)
			expected := time.Duration(tt.count-1) * tt.interval // First tick is immediate

			assert.InDelta(t, expected.Nanoseconds(), elapsed.Nanoseconds(),
				float64(tt.tolerance.Nanoseconds()),
				"timing accuracy within tolerance")
		})
	}
}

func TestIntervalScheduler_ImmediateExecution(t *testing.T) {
	scheduler, err := NewIntervalScheduler(time.Second, 0, true)
	require.NoError(t, err)
	defer scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should get immediate tick
	select {
	case <-scheduler.Next():
		// Success - got immediate tick
	case <-ctx.Done():
		t.Fatal("should have received immediate tick")
	}
}

func TestIntervalScheduler_Stop(t *testing.T) {
	scheduler, err := NewIntervalScheduler(100*time.Millisecond, 0, false)
	require.NoError(t, err)

	// Get first tick
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-scheduler.Next():
		// Got first tick
	case <-ctx.Done():
		t.Fatal("should have received first tick")
	}

	// Stop scheduler
	scheduler.Stop()

	// Should not receive more ticks
	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()

	select {
	case <-scheduler.Next():
		t.Fatal("should not receive tick after stop")
	case <-ctx2.Done():
		// Expected - no more ticks after stop
	}
}

func TestIntervalScheduler_Jitter(t *testing.T) {
	scheduler, err := NewIntervalScheduler(100*time.Millisecond, 0.2, false) // 20% jitter
	require.NoError(t, err)
	defer scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	intervals := make([]time.Duration, 0, 10)
	lastTime := time.Now()

	for len(intervals) < 10 {
		select {
		case <-scheduler.Next():
			now := time.Now()
			intervals = append(intervals, now.Sub(lastTime))
			lastTime = now
		case <-ctx.Done():
			t.Fatalf("timeout waiting for ticks, got %d/10", len(intervals))
		}
	}

	// Calculate variance to ensure jitter is working
	var sum time.Duration
	for _, interval := range intervals {
		sum += interval
	}
	avg := sum / time.Duration(len(intervals))

	// With 20% jitter, we should see some variance
	var variance float64
	for _, interval := range intervals {
		diff := float64(interval - avg)
		variance += diff * diff
	}
	variance /= float64(len(intervals))

	// Should have some variance due to jitter
	assert.Greater(t, variance, 0.0, "jitter should create timing variance")
}
