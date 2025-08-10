package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/swi/repeater/pkg/errors"
)

func TestRetryPolicy_ExponentialBackoff(t *testing.T) {
	policy := NewExponentialBackoffPolicy(3, 100*time.Millisecond, 2.0, 1*time.Second)

	// Test first retry
	delay := policy.NextDelay(1)
	if delay != 100*time.Millisecond {
		t.Errorf("Expected first retry delay 100ms, got %v", delay)
	}

	// Test second retry (exponential)
	delay = policy.NextDelay(2)
	if delay != 200*time.Millisecond {
		t.Errorf("Expected second retry delay 200ms, got %v", delay)
	}

	// Test third retry (exponential)
	delay = policy.NextDelay(3)
	if delay != 400*time.Millisecond {
		t.Errorf("Expected third retry delay 400ms, got %v", delay)
	}

	// Test max retries exceeded
	if policy.ShouldRetry(4) {
		t.Error("Expected no retry after max attempts")
	}

	// Test max delay cap
	delay = policy.NextDelay(10) // Would be very large without cap
	if delay > 1*time.Second {
		t.Errorf("Expected delay capped at 1s, got %v", delay)
	}
}

func TestRetryPolicy_LinearBackoff(t *testing.T) {
	policy := NewLinearBackoffPolicy(3, 100*time.Millisecond, 50*time.Millisecond)

	// Test linear progression
	delays := []time.Duration{
		100 * time.Millisecond, // First retry
		150 * time.Millisecond, // Second retry
		200 * time.Millisecond, // Third retry
	}

	for i, expected := range delays {
		delay := policy.NextDelay(i + 1)
		if delay != expected {
			t.Errorf("Expected delay %v for attempt %d, got %v", expected, i+1, delay)
		}
	}

	// Test max retries
	if policy.ShouldRetry(4) {
		t.Error("Expected no retry after max attempts")
	}
}

func TestRetryPolicy_FixedDelay(t *testing.T) {
	policy := NewFixedDelayPolicy(5, 250*time.Millisecond)

	// Test fixed delay for all attempts
	for i := 1; i <= 5; i++ {
		delay := policy.NextDelay(i)
		if delay != 250*time.Millisecond {
			t.Errorf("Expected fixed delay 250ms for attempt %d, got %v", i, delay)
		}
	}

	// Test max retries
	if policy.ShouldRetry(6) {
		t.Error("Expected no retry after max attempts")
	}
}

func TestRetryPolicy_ConditionalRetry(t *testing.T) {
	policy := NewConditionalRetryPolicy(3, 100*time.Millisecond)

	// Add retry conditions
	policy.AddCondition(func(err error) bool {
		catErr, ok := err.(*errors.CategorizedError)
		if !ok {
			return false
		}
		return catErr.Category() == errors.CategoryTimeout || catErr.Category() == errors.CategoryNetwork
	})

	// Test retryable error
	timeoutErr := errors.NewCategorizedError(fmt.Errorf("timeout"), errors.CategoryTimeout, errors.SeverityMedium)
	if !policy.ShouldRetryError(timeoutErr) {
		t.Error("Expected timeout error to be retryable")
	}

	// Test non-retryable error
	permissionErr := errors.NewCategorizedError(fmt.Errorf("permission denied"), errors.CategoryPermission, errors.SeverityHigh)
	if policy.ShouldRetryError(permissionErr) {
		t.Error("Expected permission error to not be retryable")
	}

	// Test generic error (should not retry)
	genericErr := fmt.Errorf("generic error")
	if policy.ShouldRetryError(genericErr) {
		t.Error("Expected generic error to not be retryable")
	}
}

func TestRecoveryManager_ExecuteWithRetry(t *testing.T) {
	manager := NewRecoveryManager()

	// Configure retry policy
	policy := NewExponentialBackoffPolicy(3, 50*time.Millisecond, 2.0, 500*time.Millisecond)
	manager.SetRetryPolicy(policy)

	// Test successful execution on first try
	attempts := 0
	successFunc := func(ctx context.Context) error {
		attempts++
		return nil
	}

	err := manager.ExecuteWithRetry(context.Background(), successFunc)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}

	// Test successful execution after retries
	attempts = 0
	retryFunc := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("temporary failure")
		}
		return nil
	}

	err = manager.ExecuteWithRetry(context.Background(), retryFunc)
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Test failure after max retries
	attempts = 0
	failFunc := func(ctx context.Context) error {
		attempts++
		return fmt.Errorf("persistent failure")
	}

	err = manager.ExecuteWithRetry(context.Background(), failFunc)
	if err == nil {
		t.Error("Expected error after max retries")
	}
	if attempts != 4 { // Initial attempt + 3 retries
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
}

func TestRecoveryManager_FallbackExecution(t *testing.T) {
	manager := NewRecoveryManager()

	// Set up fallback command
	fallbackExecuted := false
	fallback := func(ctx context.Context, originalErr error) error {
		fallbackExecuted = true
		if originalErr == nil {
			t.Error("Expected original error to be passed to fallback")
		}
		return nil
	}
	manager.SetFallback(fallback)

	// Test fallback execution when primary fails
	primaryFunc := func(ctx context.Context) error {
		return fmt.Errorf("primary failure")
	}

	err := manager.ExecuteWithFallback(context.Background(), primaryFunc)
	if err != nil {
		t.Errorf("Expected no error with successful fallback, got %v", err)
	}
	if !fallbackExecuted {
		t.Error("Expected fallback to be executed")
	}

	// Test fallback failure
	fallbackExecuted = false
	failingFallback := func(ctx context.Context, originalErr error) error {
		fallbackExecuted = true
		return fmt.Errorf("fallback also failed")
	}
	manager.SetFallback(failingFallback)

	err = manager.ExecuteWithFallback(context.Background(), primaryFunc)
	if err == nil {
		t.Error("Expected error when both primary and fallback fail")
	}
	if !fallbackExecuted {
		t.Error("Expected fallback to be attempted")
	}
}

func TestRecoveryManager_CombinedRetryAndFallback(t *testing.T) {
	manager := NewRecoveryManager()

	// Configure retry policy
	policy := NewFixedDelayPolicy(2, 10*time.Millisecond)
	manager.SetRetryPolicy(policy)

	// Configure fallback
	fallbackExecuted := false
	fallback := func(ctx context.Context, originalErr error) error {
		fallbackExecuted = true
		return nil
	}
	manager.SetFallback(fallback)

	// Test that fallback is used after retries are exhausted
	attempts := 0
	failingFunc := func(ctx context.Context) error {
		attempts++
		return fmt.Errorf("persistent failure")
	}

	err := manager.ExecuteWithRetryAndFallback(context.Background(), failingFunc)
	if err != nil {
		t.Errorf("Expected no error with fallback, got %v", err)
	}
	if attempts != 3 { // Initial + 2 retries
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
	if !fallbackExecuted {
		t.Error("Expected fallback to be executed after retries")
	}
}

func TestRecoveryManager_ContextCancellation(t *testing.T) {
	manager := NewRecoveryManager()

	// Configure retry policy with long delays
	policy := NewFixedDelayPolicy(5, 1*time.Second)
	manager.SetRetryPolicy(policy)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	attempts := 0
	slowFunc := func(ctx context.Context) error {
		attempts++
		return fmt.Errorf("failure")
	}

	start := time.Now()
	err := manager.ExecuteWithRetry(ctx, slowFunc)
	duration := time.Since(start)

	// Should fail quickly due to context cancellation
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	if duration > 200*time.Millisecond {
		t.Errorf("Expected quick failure due to context, took %v", duration)
	}
	if attempts > 2 {
		t.Errorf("Expected few attempts due to context cancellation, got %d", attempts)
	}
}

func TestRecoveryState_Tracking(t *testing.T) {
	manager := NewRecoveryManager()

	// Enable state tracking
	manager.EnableStateTracking(true)

	policy := NewFixedDelayPolicy(3, 10*time.Millisecond)
	manager.SetRetryPolicy(policy)

	// Execute function that fails then succeeds
	attempts := 0
	testFunc := func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("failure %d", attempts)
		}
		return nil
	}

	err := manager.ExecuteWithRetry(context.Background(), testFunc)
	if err != nil {
		t.Errorf("Expected success after retries, got %v", err)
	}

	// Check recovery state
	state := manager.GetRecoveryState()
	if state == nil {
		t.Fatal("Expected recovery state to be tracked")
	}

	if state.TotalAttempts != 3 {
		t.Errorf("Expected 3 total attempts, got %d", state.TotalAttempts)
	}

	if state.SuccessfulRecoveries != 1 {
		t.Errorf("Expected 1 successful recovery, got %d", state.SuccessfulRecoveries)
	}

	if len(state.RecentFailures) != 2 {
		t.Errorf("Expected 2 recent failures, got %d", len(state.RecentFailures))
	}
}

func TestRecoveryManager_SuccessThreshold(t *testing.T) {
	manager := NewRecoveryManager()

	// Set success threshold
	manager.SetSuccessThreshold(3)

	// Simulate successful executions
	successFunc := func(ctx context.Context) error {
		return nil
	}

	// Execute successfully multiple times
	for i := 0; i < 5; i++ {
		err := manager.ExecuteWithRetry(context.Background(), successFunc)
		if err != nil {
			t.Errorf("Expected success, got %v", err)
		}
	}

	// Check that success threshold affects recovery behavior
	state := manager.GetRecoveryState()
	if state != nil && state.ConsecutiveSuccesses < 5 {
		t.Errorf("Expected at least 5 consecutive successes, got %d", state.ConsecutiveSuccesses)
	}
}
