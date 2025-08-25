package recovery

import (
	"context"
	"sync"
	"time"
)

// RecoveryManager manages retry policies and fallback strategies
type RecoveryManager struct {
	mu               sync.RWMutex
	retryPolicy      RetryPolicy
	fallback         FallbackFunc
	successThreshold int
	stateTracking    bool
	state            *RecoveryState
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		successThreshold: 1,
		stateTracking:    false,
		state: &RecoveryState{
			RecentFailures: make([]error, 0),
		},
	}
}

// SetRetryPolicy sets the retry policy
func (rm *RecoveryManager) SetRetryPolicy(policy RetryPolicy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.retryPolicy = policy
}

// SetFallback sets the fallback function
func (rm *RecoveryManager) SetFallback(fallback FallbackFunc) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.fallback = fallback
}

// SetSuccessThreshold sets the number of consecutive successes needed
func (rm *RecoveryManager) SetSuccessThreshold(threshold int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.successThreshold = threshold
}

// EnableStateTracking enables or disables state tracking
func (rm *RecoveryManager) EnableStateTracking(enabled bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.stateTracking = enabled
}

// GetRecoveryState returns the current recovery state
func (rm *RecoveryManager) GetRecoveryState() *RecoveryState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if !rm.stateTracking {
		return nil
	}

	// Return a copy to avoid race conditions
	stateCopy := *rm.state
	stateCopy.RecentFailures = make([]error, len(rm.state.RecentFailures))
	copy(stateCopy.RecentFailures, rm.state.RecentFailures)

	return &stateCopy
}

// ExecuteWithRetry executes a function with retry policy
func (rm *RecoveryManager) ExecuteWithRetry(ctx context.Context, fn ExecuteFunc) error {
	rm.mu.RLock()
	policy := rm.retryPolicy
	rm.mu.RUnlock()

	if policy == nil {
		// No retry policy, execute once
		return rm.executeAndTrack(ctx, fn)
	}

	var lastErr error
	attempt := 0

	for {
		attempt++

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := rm.executeAndTrack(ctx, fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry
		if !policy.ShouldRetry(attempt) || !policy.ShouldRetryError(err) {
			break
		}

		// Wait for retry delay
		delay := policy.NextDelay(attempt)
		if delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

// ExecuteWithFallback executes a function with fallback on failure
func (rm *RecoveryManager) ExecuteWithFallback(ctx context.Context, fn ExecuteFunc) error {
	err := rm.executeAndTrack(ctx, fn)
	if err == nil {
		return nil
	}

	rm.mu.RLock()
	fallback := rm.fallback
	rm.mu.RUnlock()

	if fallback == nil {
		return err
	}

	// Execute fallback
	return fallback(ctx, err)
}

// ExecuteWithRetryAndFallback executes a function with both retry and fallback
func (rm *RecoveryManager) ExecuteWithRetryAndFallback(ctx context.Context, fn ExecuteFunc) error {
	err := rm.ExecuteWithRetry(ctx, fn)
	if err == nil {
		return nil
	}

	rm.mu.RLock()
	fallback := rm.fallback
	rm.mu.RUnlock()

	if fallback == nil {
		return err
	}

	// Execute fallback after retries are exhausted
	return fallback(ctx, err)
}

// executeAndTrack executes a function and tracks the result
func (rm *RecoveryManager) executeAndTrack(ctx context.Context, fn ExecuteFunc) error {
	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.stateTracking {
		return err
	}

	rm.state.TotalAttempts++

	if err == nil {
		// Success
		rm.state.SuccessfulRecoveries++
		rm.state.ConsecutiveSuccesses++
		rm.state.ConsecutiveFailures = 0
		rm.state.LastSuccessTime = time.Now()

		// Update average recovery time
		if rm.state.AverageRecoveryTime == 0 {
			rm.state.AverageRecoveryTime = duration
		} else {
			rm.state.AverageRecoveryTime = (rm.state.AverageRecoveryTime + duration) / 2
		}
	} else {
		// Failure
		rm.state.FailedRecoveries++
		rm.state.ConsecutiveFailures++
		rm.state.ConsecutiveSuccesses = 0
		rm.state.LastFailureTime = time.Now()

		// Track recent failures (keep last 10)
		rm.state.RecentFailures = append(rm.state.RecentFailures, err)
		if len(rm.state.RecentFailures) > 10 {
			rm.state.RecentFailures = rm.state.RecentFailures[1:]
		}
	}

	return err
}
