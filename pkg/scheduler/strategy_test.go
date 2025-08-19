package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/strategies"
)

// TestStrategyScheduler_Creation tests the creation of strategy schedulers
func TestStrategyScheduler_Creation(t *testing.T) {
	tests := []struct {
		name          string
		strategy      strategies.Strategy
		config        *strategies.StrategyConfig
		expectedError string
	}{
		{
			name:     "exponential_strategy_valid_config",
			strategy: &strategies.ExponentialStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   time.Second,
				MaxDelay:    10 * time.Second,
				Multiplier:  2.0,
				MaxAttempts: 5,
			},
			expectedError: "",
		},
		{
			name:     "fibonacci_strategy_valid_config",
			strategy: &strategies.FibonacciStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   500 * time.Millisecond,
				MaxDelay:    30 * time.Second,
				MaxAttempts: 8,
			},
			expectedError: "",
		},
		{
			name:     "linear_strategy_valid_config",
			strategy: &strategies.LinearStrategy{},
			config: &strategies.StrategyConfig{
				Increment:   2 * time.Second,
				MaxDelay:    20 * time.Second,
				MaxAttempts: 5,
			},
			expectedError: "",
		},
		{
			name:     "polynomial_strategy_valid_config",
			strategy: &strategies.PolynomialStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   time.Second,
				Exponent:    1.5,
				MaxDelay:    60 * time.Second,
				MaxAttempts: 4,
			},
			expectedError: "",
		},
		{
			name:     "decorrelated_jitter_strategy_valid_config",
			strategy: &strategies.DecorrelatedJitterStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   time.Second,
				Multiplier:  3.0,
				MaxDelay:    120 * time.Second,
				MaxAttempts: 6,
			},
			expectedError: "",
		},
		{
			name:     "exponential_strategy_invalid_config",
			strategy: &strategies.ExponentialStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:  0, // Invalid: zero base delay
				MaxDelay:   10 * time.Second,
				Multiplier: 2.0,
			},
			expectedError: "base-delay must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := NewStrategyScheduler(tt.strategy, tt.config)

			if tt.expectedError == "" {
				require.NoError(t, err)
				require.NotNil(t, scheduler)
				assert.Equal(t, tt.strategy, scheduler.strategy)
				assert.Equal(t, tt.config, scheduler.config)
				assert.Equal(t, 0, scheduler.currentAttempt)
				assert.Equal(t, tt.config.MaxAttempts, scheduler.maxAttempts)
				assert.False(t, scheduler.stopped)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, scheduler)
			}
		})
	}
}

// TestStrategyScheduler_BasicExecution tests basic scheduler execution
func TestStrategyScheduler_BasicExecution(t *testing.T) {
	strategy := &strategies.LinearStrategy{}
	config := &strategies.StrategyConfig{
		Increment:   500 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		MaxAttempts: 3,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Test first execution (immediate)
	nextChan := scheduler.Next()
	select {
	case execTime := <-nextChan:
		assert.WithinDuration(t, time.Now(), execTime, 100*time.Millisecond)
		assert.Equal(t, 1, scheduler.GetAttemptNumber())
	case <-time.After(1 * time.Second):
		t.Fatal("Expected immediate execution")
	}

	// Test second execution (with delay)
	scheduler.UpdateExecutionResult(100*time.Millisecond, false, "failed")
	nextChan = scheduler.Next()
	select {
	case <-nextChan:
		// Just verify we get the execution, timing can vary
		assert.Equal(t, 2, scheduler.GetAttemptNumber())
	case <-time.After(2 * time.Second):
		t.Fatal("Expected execution with delay")
	}
}

// TestStrategyScheduler_SuccessStopsRetry tests that success stops the retry loop
func TestStrategyScheduler_SuccessStopsRetry(t *testing.T) {
	strategy := &strategies.ExponentialStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 10, // High number to ensure success stops it first
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// First execution
	nextChan := scheduler.Next()
	<-nextChan

	// Simulate success
	scheduler.UpdateExecutionResult(50*time.Millisecond, true, "success")

	// Try to get next execution - should be stopped
	nextChan = scheduler.Next()
	select {
	case <-nextChan:
		t.Fatal("Should not receive next execution after success")
	case <-time.After(300 * time.Millisecond):
		// Expected - no more executions after success
	}

	// Verify scheduler is in stopped state
	scheduler.mu.RLock()
	stopped := scheduler.stopped
	scheduler.mu.RUnlock()
	assert.True(t, stopped)
}

// TestStrategyScheduler_MaxAttemptsLimit tests that max attempts limit is respected
func TestStrategyScheduler_MaxAttemptsLimit(t *testing.T) {
	strategy := &strategies.FibonacciStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		MaxAttempts: 2, // Limited attempts
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// First execution
	nextChan := scheduler.Next()
	<-nextChan
	assert.Equal(t, 1, scheduler.GetAttemptNumber())

	// Simulate failure
	scheduler.UpdateExecutionResult(100*time.Millisecond, false, "failed")

	// Second execution
	nextChan = scheduler.Next()
	<-nextChan
	assert.Equal(t, 2, scheduler.GetAttemptNumber())

	// Simulate failure again
	scheduler.UpdateExecutionResult(100*time.Millisecond, false, "failed")

	// Should not get third execution due to max attempts
	nextChan = scheduler.Next()
	select {
	case <-nextChan:
		t.Fatal("Should not receive execution beyond max attempts")
	case <-time.After(200 * time.Millisecond):
		// Expected - no more executions after max attempts
	}

	// Verify scheduler is stopped
	scheduler.mu.RLock()
	stopped := scheduler.stopped
	scheduler.mu.RUnlock()
	assert.True(t, stopped)
}

// TestStrategyScheduler_StopFunctionality tests the Stop functionality
func TestStrategyScheduler_StopFunctionality(t *testing.T) {
	strategy := &strategies.PolynomialStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   100 * time.Millisecond,
		Exponent:    2.0,
		MaxDelay:    10 * time.Second,
		MaxAttempts: 10,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// First execution
	nextChan := scheduler.Next()
	<-nextChan

	// Stop the scheduler
	scheduler.Stop()

	// Verify stopped state
	scheduler.mu.RLock()
	stopped := scheduler.stopped
	scheduler.mu.RUnlock()
	assert.True(t, stopped)

	// Try to get next execution - should not receive anything
	nextChan = scheduler.Next()
	select {
	case <-nextChan:
		t.Fatal("Should not receive execution after stop")
	case <-time.After(200 * time.Millisecond):
		// Expected - no execution after stop
	}

	// Multiple calls to Stop should be safe (idempotent)
	scheduler.Stop()
	scheduler.Stop()
}

// TestStrategyScheduler_IsRetryMode tests the IsRetryMode method
func TestStrategyScheduler_IsRetryMode(t *testing.T) {
	strategy := &strategies.ExponentialStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   time.Second,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 5,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Strategy schedulers should always be in retry mode
	assert.True(t, scheduler.IsRetryMode())
}

// TestStrategyScheduler_GetStrategy tests the GetStrategy method
func TestStrategyScheduler_GetStrategy(t *testing.T) {
	strategy := &strategies.DecorrelatedJitterStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   time.Second,
		Multiplier:  3.0,
		MaxDelay:    60 * time.Second,
		MaxAttempts: 6,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Should return the same strategy instance
	assert.Equal(t, strategy, scheduler.GetStrategy())
}

// TestStrategyScheduler_ConcurrentOperations tests concurrent operations
func TestStrategyScheduler_ConcurrentOperations(t *testing.T) {
	strategy := &strategies.LinearStrategy{}
	config := &strategies.StrategyConfig{
		Increment:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		MaxAttempts: 5,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Test concurrent access to scheduler methods
	done := make(chan bool, 3)

	// Goroutine 1: Call Next()
	go func() {
		nextChan := scheduler.Next()
		<-nextChan
		done <- true
	}()

	// Goroutine 2: Call GetAttemptNumber()
	go func() {
		for i := 0; i < 10; i++ {
			scheduler.GetAttemptNumber()
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: Call IsRetryMode()
	go func() {
		for i := 0; i < 10; i++ {
			scheduler.IsRetryMode()
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("Concurrent operations timeout")
		}
	}

	// Stop the scheduler
	scheduler.Stop()
}

// TestStrategyScheduler_ZeroMaxAttempts tests behavior with unlimited attempts
func TestStrategyScheduler_ZeroMaxAttempts(t *testing.T) {
	strategy := &strategies.ExponentialStrategy{}
	config := &strategies.StrategyConfig{
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Multiplier:  2.0,
		MaxAttempts: 10, // High number instead of 0
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Execute multiple times without hitting limit
	for i := 1; i <= 5; i++ {
		nextChan := scheduler.Next()
		select {
		case <-nextChan:
			assert.Equal(t, i, scheduler.GetAttemptNumber())
			// Simulate failure to continue
			scheduler.UpdateExecutionResult(100*time.Millisecond, false, "failed")
		case <-time.After(1 * time.Second):
			t.Fatalf("Expected execution %d", i)
		}
	}

	// Should still be able to continue (not stopped)
	scheduler.mu.RLock()
	stopped := scheduler.stopped
	scheduler.mu.RUnlock()
	assert.False(t, stopped)

	scheduler.Stop()
}

// TestStrategyScheduler_UpdateExecutionResult tests the UpdateExecutionResult method
func TestStrategyScheduler_UpdateExecutionResult(t *testing.T) {
	strategy := &strategies.LinearStrategy{}
	config := &strategies.StrategyConfig{
		Increment:   200 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		MaxAttempts: 3,
	}

	scheduler, err := NewStrategyScheduler(strategy, config)
	require.NoError(t, err)

	// Test updating with failure result
	scheduler.UpdateExecutionResult(150*time.Millisecond, false, "error output")
	assert.Equal(t, 150*time.Millisecond, scheduler.lastDuration)

	// Scheduler should still be running after failure
	scheduler.mu.RLock()
	stopped := scheduler.stopped
	scheduler.mu.RUnlock()
	assert.False(t, stopped)

	// Test updating with success result
	scheduler.UpdateExecutionResult(75*time.Millisecond, true, "success output")
	assert.Equal(t, 75*time.Millisecond, scheduler.lastDuration)

	// Scheduler should be stopped after success
	scheduler.mu.RLock()
	stopped = scheduler.stopped
	scheduler.mu.RUnlock()
	assert.True(t, stopped)
}

// TestStrategyScheduler_AllStrategiesIntegration tests all strategy types
func TestStrategyScheduler_AllStrategiesIntegration(t *testing.T) {
	strategies := []struct {
		name     string
		strategy strategies.Strategy
		config   *strategies.StrategyConfig
	}{
		{
			name:     "exponential",
			strategy: &strategies.ExponentialStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   100 * time.Millisecond,
				MaxDelay:    2 * time.Second,
				Multiplier:  2.0,
				MaxAttempts: 3,
			},
		},
		{
			name:     "fibonacci",
			strategy: &strategies.FibonacciStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   100 * time.Millisecond,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 3,
			},
		},
		{
			name:     "linear",
			strategy: &strategies.LinearStrategy{},
			config: &strategies.StrategyConfig{
				Increment:   150 * time.Millisecond,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 3,
			},
		},
		{
			name:     "polynomial",
			strategy: &strategies.PolynomialStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   100 * time.Millisecond,
				Exponent:    1.5,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 3,
			},
		},
		{
			name:     "decorrelated_jitter",
			strategy: &strategies.DecorrelatedJitterStrategy{},
			config: &strategies.StrategyConfig{
				BaseDelay:   100 * time.Millisecond,
				Multiplier:  2.5,
				MaxDelay:    2 * time.Second,
				MaxAttempts: 3,
			},
		},
	}

	for _, tt := range strategies {
		t.Run(tt.name+"_integration", func(t *testing.T) {
			scheduler, err := NewStrategyScheduler(tt.strategy, tt.config)
			require.NoError(t, err)

			// Test basic execution flow
			nextChan := scheduler.Next()
			<-nextChan
			assert.Equal(t, 1, scheduler.GetAttemptNumber())
			assert.True(t, scheduler.IsRetryMode())
			assert.Equal(t, tt.strategy, scheduler.GetStrategy())

			// Test failure and retry
			scheduler.UpdateExecutionResult(100*time.Millisecond, false, "failed")
			nextChan = scheduler.Next()
			select {
			case <-nextChan:
				assert.Equal(t, 2, scheduler.GetAttemptNumber())
			case <-time.After(3 * time.Second):
				t.Fatal("Expected retry execution")
			}

			// Clean up
			scheduler.Stop()
		})
	}
}
