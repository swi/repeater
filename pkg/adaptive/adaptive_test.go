package adaptive

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAIMDAdapter_BasicAdaptation tests basic AIMD interval adjustment
func TestAIMDAdapter_BasicAdaptation(t *testing.T) {
	tests := []struct {
		name           string
		baseInterval   time.Duration
		responseTime   time.Duration
		success        bool
		expectedChange string // "increase", "decrease", "same"
	}{
		{
			name:           "slow response increases interval",
			baseInterval:   time.Second,
			responseTime:   3 * time.Second, // 3x base = slow
			success:        true,
			expectedChange: "increase",
		},
		{
			name:           "fast response decreases interval",
			baseInterval:   time.Second,
			responseTime:   200 * time.Millisecond, // 0.2x base = fast
			success:        true,
			expectedChange: "decrease",
		},
		{
			name:           "failure decreases interval significantly",
			baseInterval:   time.Second,
			responseTime:   500 * time.Millisecond,
			success:        false,
			expectedChange: "decrease",
		},
		{
			name:           "normal response keeps interval stable",
			baseInterval:   time.Second,
			responseTime:   time.Second, // 1x base = normal
			success:        true,
			expectedChange: "same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultAIMDConfig()
			config.BaseInterval = tt.baseInterval

			adapter := NewAIMDAdapter(config)
			require.NotNil(t, adapter)

			initialInterval := adapter.GetCurrentInterval()

			// Update with response
			adapter.UpdateInterval(tt.responseTime, tt.success)

			newInterval := adapter.GetCurrentInterval()

			switch tt.expectedChange {
			case "increase":
				assert.Greater(t, newInterval, initialInterval, "interval should increase for slow responses")
			case "decrease":
				assert.Less(t, newInterval, initialInterval, "interval should decrease for fast responses or failures")
			case "same":
				assert.InDelta(t, initialInterval.Nanoseconds(), newInterval.Nanoseconds(),
					float64(100*time.Millisecond.Nanoseconds()), "interval should remain stable for normal responses")
			}
		})
	}
}

// TestAIMDAdapter_BoundaryConditions tests interval bounds
func TestAIMDAdapter_BoundaryConditions(t *testing.T) {
	config := DefaultAIMDConfig()
	config.BaseInterval = time.Second
	config.MinInterval = 100 * time.Millisecond
	config.MaxInterval = 30 * time.Second

	adapter := NewAIMDAdapter(config)
	require.NotNil(t, adapter)

	// Test minimum bound
	for i := 0; i < 20; i++ {
		adapter.UpdateInterval(50*time.Millisecond, false) // Fast failures
	}

	interval := adapter.GetCurrentInterval()
	assert.GreaterOrEqual(t, interval, config.MinInterval, "interval should not go below minimum")

	// Reset and test maximum bound
	adapter = NewAIMDAdapter(config)
	for i := 0; i < 20; i++ {
		adapter.UpdateInterval(10*time.Second, true) // Very slow responses
	}

	interval = adapter.GetCurrentInterval()
	assert.LessOrEqual(t, interval, config.MaxInterval, "interval should not exceed maximum")
}

// TestAIMDAdapter_EWMASmoothing tests exponential weighted moving average
func TestAIMDAdapter_EWMASmoothing(t *testing.T) {
	config := DefaultAIMDConfig()
	config.EWMAAlpha = 0.5 // High alpha for faster adaptation in tests

	adapter := NewAIMDAdapter(config)
	require.NotNil(t, adapter)

	// Send consistent fast responses
	for i := 0; i < 10; i++ {
		adapter.UpdateInterval(100*time.Millisecond, true)
	}

	avgResponseTime := adapter.GetAverageResponseTime()
	assert.InDelta(t, 100*time.Millisecond, avgResponseTime, float64(50*time.Millisecond),
		"EWMA should converge to consistent response time")
}

// TestBayesianPredictor_SuccessRatePrediction tests Bayesian success prediction
func TestBayesianPredictor_SuccessRatePrediction(t *testing.T) {
	config := DefaultBayesianConfig()
	predictor := NewBayesianPredictor(config)
	require.NotNil(t, predictor)

	// Initially should have neutral prediction (around 0.5 with uniform prior)
	initialRate := predictor.GetSuccessProbability()
	assert.InDelta(t, 0.5, initialRate, 0.2, "initial success rate should be around 0.5")

	// Add successful observations
	for i := 0; i < 10; i++ {
		predictor.UpdatePattern(true)
	}

	successRate := predictor.GetSuccessProbability()
	assert.Greater(t, successRate, 0.7, "success rate should increase after successful observations")

	// Add failures
	for i := 0; i < 15; i++ {
		predictor.UpdatePattern(false)
	}

	successRate = predictor.GetSuccessProbability()
	assert.Less(t, successRate, 0.4, "success rate should decrease after failures")
}

// TestBayesianPredictor_CircuitBreaker tests circuit breaker functionality
func TestBayesianPredictor_CircuitBreaker(t *testing.T) {
	config := DefaultBayesianConfig()
	config.FailureThreshold = 0.3  // Open circuit at 30% failure rate
	config.RecoveryThreshold = 0.8 // Close circuit at 80% success rate

	predictor := NewBayesianPredictor(config)
	require.NotNil(t, predictor)

	// Initially circuit should be closed
	assert.Equal(t, CircuitClosed, predictor.GetCircuitState())

	// Add many failures to open circuit
	for i := 0; i < 20; i++ {
		predictor.UpdatePattern(false)
	}

	assert.Equal(t, CircuitOpen, predictor.GetCircuitState(), "circuit should open after many failures")

	// Add successes to recover
	for i := 0; i < 30; i++ {
		predictor.UpdatePattern(true)
	}

	state := predictor.GetCircuitState()
	assert.True(t, state == CircuitClosed || state == CircuitHalfOpen,
		"circuit should recover after many successes")
}

// TestAdaptiveScheduler_Integration tests full adaptive scheduler
func TestAdaptiveScheduler_Integration(t *testing.T) {
	config := DefaultAdaptiveConfig()
	config.BaseInterval = 500 * time.Millisecond

	scheduler := NewAdaptiveScheduler(config)
	require.NotNil(t, scheduler)

	// Test initial interval
	initialInterval := scheduler.GetCurrentInterval()
	assert.Equal(t, config.BaseInterval, initialInterval, "initial interval should match base")

	// Simulate slow responses
	result := ExecutionResult{
		Timestamp:    time.Now(),
		ResponseTime: 2 * time.Second,
		Success:      true,
		StatusCode:   200,
	}

	scheduler.UpdateFromResult(result)

	newInterval := scheduler.GetCurrentInterval()
	assert.Greater(t, newInterval, initialInterval, "interval should increase after slow response")

	// Test metrics
	metrics := scheduler.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalExecutions, "should track total executions")
	assert.Equal(t, int64(1), metrics.SuccessfulExecutions, "should track successful executions")
}

// TestAdaptiveScheduler_PatternLearning tests pattern learning integration
func TestAdaptiveScheduler_PatternLearning(t *testing.T) {
	config := DefaultAdaptiveConfig()
	config.WindowSize = 10

	scheduler := NewAdaptiveScheduler(config)
	require.NotNil(t, scheduler)

	// Add pattern of alternating success/failure
	for i := 0; i < 20; i++ {
		result := ExecutionResult{
			Timestamp:    time.Now(),
			ResponseTime: 500 * time.Millisecond,
			Success:      i%2 == 0, // Alternating pattern
			StatusCode:   200,
		}
		scheduler.UpdateFromResult(result)
	}

	// Should detect pattern and adjust accordingly
	successRate := scheduler.GetSuccessProbability()
	assert.InDelta(t, 0.5, successRate, 0.2, "should detect 50% success pattern")
}

// TestAdaptiveScheduler_ConfigValidation tests configuration validation
func TestAdaptiveScheduler_ConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *AdaptiveConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    DefaultAdaptiveConfig(),
			wantError: false,
		},
		{
			name: "invalid min > max interval",
			config: &AdaptiveConfig{
				BaseInterval: time.Second,
				MinInterval:  2 * time.Second,
				MaxInterval:  time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid EWMA alpha",
			config: &AdaptiveConfig{
				BaseInterval:      time.Second,
				MinInterval:       100 * time.Millisecond,
				MaxInterval:       30 * time.Second,
				ResponseTimeAlpha: 1.5, // Invalid: > 1.0
			},
			wantError: true,
		},
		{
			name: "invalid threshold",
			config: &AdaptiveConfig{
				BaseInterval:     time.Second,
				MinInterval:      100 * time.Millisecond,
				MaxInterval:      30 * time.Second,
				FailureThreshold: -0.1, // Invalid: < 0
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewAdaptiveSchedulerWithValidation(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, scheduler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scheduler)
			}
		})
	}
}
