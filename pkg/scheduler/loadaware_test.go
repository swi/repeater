package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemResourceMonitor tests system resource monitoring functionality
func TestSystemResourceMonitor(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "monitor should return current system metrics",
			test: func(t *testing.T) {
				monitor := NewSystemResourceMonitor()
				require.NotNil(t, monitor)

				metrics, err := monitor.GetCurrentMetrics()
				require.NoError(t, err)
				require.NotNil(t, metrics)

				// CPU usage should be between 0 and 100
				assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0)
				assert.LessOrEqual(t, metrics.CPUUsage, 100.0)

				// Memory usage should be between 0 and 100
				assert.GreaterOrEqual(t, metrics.MemoryUsage, 0.0)
				assert.LessOrEqual(t, metrics.MemoryUsage, 100.0)

				// Load average should be non-negative
				assert.GreaterOrEqual(t, metrics.LoadAverage1m, 0.0)
				assert.GreaterOrEqual(t, metrics.LoadAverage5m, 0.0)
				assert.GreaterOrEqual(t, metrics.LoadAverage15m, 0.0)
			},
		},
		{
			name: "monitor should track metrics over time",
			test: func(t *testing.T) {
				monitor := NewSystemResourceMonitor()
				require.NotNil(t, monitor)

				// Get initial metrics
				metrics1, err := monitor.GetCurrentMetrics()
				require.NoError(t, err)

				// Wait a bit and get metrics again
				time.Sleep(10 * time.Millisecond)
				metrics2, err := monitor.GetCurrentMetrics()
				require.NoError(t, err)

				// Timestamps should be different
				assert.True(t, metrics2.Timestamp.After(metrics1.Timestamp))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// TestLoadAwareScheduler tests the load-aware adaptive scheduler
func TestLoadAwareScheduler(t *testing.T) {
	tests := []struct {
		name         string
		baseInterval time.Duration
		targetCPU    float64
		targetMemory float64
		targetLoad   float64
		mockMetrics  *SystemMetrics
		expectedMin  time.Duration
		expectedMax  time.Duration
	}{
		{
			name:         "low system load should decrease interval",
			baseInterval: time.Second,
			targetCPU:    70.0,
			targetMemory: 80.0,
			targetLoad:   1.0,
			mockMetrics: &SystemMetrics{
				CPUUsage:      30.0, // Low CPU
				MemoryUsage:   40.0, // Low memory
				LoadAverage1m: 0.5,  // Low load
				Timestamp:     time.Now(),
			},
			expectedMin: 500 * time.Millisecond, // Should decrease
			expectedMax: 800 * time.Millisecond,
		},
		{
			name:         "high system load should increase interval",
			baseInterval: time.Second,
			targetCPU:    70.0,
			targetMemory: 80.0,
			targetLoad:   1.0,
			mockMetrics: &SystemMetrics{
				CPUUsage:      90.0, // High CPU
				MemoryUsage:   95.0, // High memory
				LoadAverage1m: 2.5,  // High load
				Timestamp:     time.Now(),
			},
			expectedMin: 1500 * time.Millisecond, // Should increase
			expectedMax: 3000 * time.Millisecond,
		},
		{
			name:         "target system load should maintain interval",
			baseInterval: time.Second,
			targetCPU:    70.0,
			targetMemory: 80.0,
			targetLoad:   1.0,
			mockMetrics: &SystemMetrics{
				CPUUsage:      70.0, // At target
				MemoryUsage:   80.0, // At target
				LoadAverage1m: 1.0,  // At target
				Timestamp:     time.Now(),
			},
			expectedMin: 900 * time.Millisecond, // Should stay close to base
			expectedMax: 1100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This will fail initially - no LoadAwareScheduler exists
			scheduler := NewLoadAwareScheduler(tt.baseInterval, tt.targetCPU, tt.targetMemory, tt.targetLoad)
			require.NotNil(t, scheduler)

			// Inject mock metrics
			scheduler.SetMockMetrics(tt.mockMetrics)

			// Update scheduler with mock metrics
			_ = scheduler.UpdateFromMetrics()

			// Get current interval
			interval := scheduler.GetCurrentInterval()
			assert.GreaterOrEqual(t, interval, tt.expectedMin)
			assert.LessOrEqual(t, interval, tt.expectedMax)
		})
	}
}

// TestLoadAwareSchedulerInterface tests that it implements Scheduler interface
func TestLoadAwareSchedulerInterface(t *testing.T) {
	scheduler := NewLoadAwareScheduler(time.Second, 70.0, 80.0, 1.0)
	require.NotNil(t, scheduler)

	// Test Next() returns a channel
	nextChan := scheduler.Next()
	assert.NotNil(t, nextChan)

	// Test Stop() doesn't panic
	scheduler.Stop()
}

// TestLoadAwareSchedulerConcurrency tests concurrent access
func TestLoadAwareSchedulerConcurrency(t *testing.T) {
	scheduler := NewLoadAwareScheduler(time.Second, 70.0, 80.0, 1.0)
	require.NotNil(t, scheduler)

	done := make(chan bool, 2)

	// Concurrent metric updates
	go func() {
		for i := 0; i < 10; i++ {
			mockMetrics := &SystemMetrics{
				CPUUsage:      float64(i * 10),
				MemoryUsage:   float64(i * 8),
				LoadAverage1m: float64(i) * 0.2,
				Timestamp:     time.Now(),
			}
			scheduler.SetMockMetrics(mockMetrics)
			_ = scheduler.UpdateFromMetrics()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent interval reads
	go func() {
		for i := 0; i < 10; i++ {
			_ = scheduler.GetCurrentInterval()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	scheduler.Stop()
}

// TestLoadAwareSchedulerBounds tests interval bounds
func TestLoadAwareSchedulerBounds(t *testing.T) {
	baseInterval := time.Second
	minInterval := 100 * time.Millisecond
	maxInterval := 10 * time.Second

	scheduler := NewLoadAwareSchedulerWithBounds(baseInterval, 70.0, 80.0, 1.0, minInterval, maxInterval)
	require.NotNil(t, scheduler)

	// Test extreme low load - should not go below min
	lowLoadMetrics := &SystemMetrics{
		CPUUsage:      1.0,
		MemoryUsage:   1.0,
		LoadAverage1m: 0.01,
		Timestamp:     time.Now(),
	}
	scheduler.SetMockMetrics(lowLoadMetrics)
	_ = scheduler.UpdateFromMetrics()

	interval := scheduler.GetCurrentInterval()
	assert.GreaterOrEqual(t, interval, minInterval)

	// Test extreme high load - should not go above max
	highLoadMetrics := &SystemMetrics{
		CPUUsage:      99.0,
		MemoryUsage:   99.0,
		LoadAverage1m: 10.0,
		Timestamp:     time.Now(),
	}
	scheduler.SetMockMetrics(highLoadMetrics)
	_ = scheduler.UpdateFromMetrics()

	interval = scheduler.GetCurrentInterval()
	assert.LessOrEqual(t, interval, maxInterval)

	scheduler.Stop()
}

// TestLoadAwareSchedulerMetricsHistory tests metrics history tracking
func TestLoadAwareSchedulerMetricsHistory(t *testing.T) {
	scheduler := NewLoadAwareScheduler(time.Second, 70.0, 80.0, 1.0)
	require.NotNil(t, scheduler)

	// Add several metrics
	for i := 0; i < 5; i++ {
		metrics := &SystemMetrics{
			CPUUsage:      float64(i * 20),
			MemoryUsage:   float64(i * 15),
			LoadAverage1m: float64(i) * 0.3,
			Timestamp:     time.Now().Add(time.Duration(i) * time.Second),
		}
		scheduler.SetMockMetrics(metrics)
		_ = scheduler.UpdateFromMetrics()
	}

	// Get metrics history
	history := scheduler.GetMetricsHistory()
	assert.Len(t, history, 5)

	// Verify chronological order
	for i := 1; i < len(history); i++ {
		assert.True(t, history[i].Timestamp.After(history[i-1].Timestamp))
	}

	scheduler.Stop()
}
