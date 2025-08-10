package recovery

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 5*time.Second, 2*time.Second)

	// Initial state should be Closed
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state Closed, got %v", cb.State())
	}

	// Record failures to trigger Open state
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Should now be Open
	if cb.State() != StateOpen {
		t.Errorf("Expected state Open after failures, got %v", cb.State())
	}

	// Should reject calls in Open state
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Error("Expected error in Open state")
	}
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond, 50*time.Millisecond)

	// Trigger Open state
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected Open state, got %v", cb.State())
	}

	// Wait for timeout to trigger Half-Open
	time.Sleep(150 * time.Millisecond)

	// Next call should transition to Half-Open
	executed := false
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error in Half-Open state, got %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed in Half-Open state")
	}

	// Should now be Closed after successful execution
	if cb.State() != StateClosed {
		t.Errorf("Expected Closed state after success, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 100*time.Millisecond, 50*time.Millisecond)

	// Trigger Open state
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Execute failing function in Half-Open state
	executed := false
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return fmt.Errorf("test failure")
	})

	if err == nil {
		t.Error("Expected error from failing function")
	}
	if !executed {
		t.Error("Expected function to be executed in Half-Open state")
	}

	// Should return to Open state after failure
	if cb.State() != StateOpen {
		t.Errorf("Expected Open state after Half-Open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_SuccessfulExecution(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 5*time.Second, 2*time.Second)

	// Execute successful function
	executed := false
	result := ""
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		result = "success"
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed")
	}
	if result != "success" {
		t.Errorf("Expected result 'success', got '%s'", result)
	}

	// Should remain Closed
	if cb.State() != StateClosed {
		t.Errorf("Expected Closed state, got %v", cb.State())
	}
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 1*time.Second, 500*time.Millisecond)

	// Record failures below threshold
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
		if cb.State() != StateClosed {
			t.Errorf("Expected Closed state after %d failures, got %v", i+1, cb.State())
		}
	}

	// Record failure that crosses threshold
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("Expected Open state after threshold, got %v", cb.State())
	}
}

func TestCircuitBreaker_Statistics(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second, 500*time.Millisecond)

	// Record some successes and failures
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()

	stats := cb.Statistics()
	if stats.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", stats.Name)
	}
	if stats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
	}
	if stats.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", stats.SuccessCount)
	}
	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failure, got %d", stats.FailureCount)
	}
	if stats.State != StateClosed {
		t.Errorf("Expected Closed state, got %v", stats.State)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 1*time.Second, 500*time.Millisecond)

	// Trigger Open state
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("Expected Open state, got %v", cb.State())
	}

	// Reset circuit breaker
	cb.Reset()

	// Should be Closed and allow execution
	if cb.State() != StateClosed {
		t.Errorf("Expected Closed state after reset, got %v", cb.State())
	}

	executed := false
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed after reset")
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker("test", 10, 1*time.Second, 500*time.Millisecond)

	// Execute multiple goroutines concurrently
	const numGoroutines = 20
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond) // Simulate work
				if id%3 == 0 {
					return fmt.Errorf("failure %d", id)
				}
				return nil
			})
			results <- err
		}(i)
	}

	// Collect results
	successCount := 0
	failureCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	// Verify statistics
	stats := cb.Statistics()
	if stats.TotalRequests != numGoroutines {
		t.Errorf("Expected %d total requests, got %d", numGoroutines, stats.TotalRequests)
	}

	// Should have some successes and failures
	if successCount == 0 {
		t.Error("Expected some successful executions")
	}
	if failureCount == 0 {
		t.Error("Expected some failed executions")
	}
}

func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 1*time.Second, 500*time.Millisecond)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Execute function that takes longer than timeout
	executed := false
	err := cb.Execute(ctx, func(ctx context.Context) error {
		executed = true
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	})

	if err == nil {
		t.Error("Expected context cancellation error")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
	if !executed {
		t.Error("Expected function to be executed")
	}
}

func TestCircuitBreaker_MultipleInstances(t *testing.T) {
	cb1 := NewCircuitBreaker("service1", 2, 1*time.Second, 500*time.Millisecond)
	cb2 := NewCircuitBreaker("service2", 3, 1*time.Second, 500*time.Millisecond)

	// Trigger different states
	cb1.RecordFailure()
	cb1.RecordFailure() // Should open cb1

	cb2.RecordSuccess() // cb2 should remain closed

	// Verify independent states
	if cb1.State() != StateOpen {
		t.Errorf("Expected cb1 to be Open, got %v", cb1.State())
	}
	if cb2.State() != StateClosed {
		t.Errorf("Expected cb2 to be Closed, got %v", cb2.State())
	}

	// Verify independent execution
	err1 := cb1.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err1 != ErrCircuitBreakerOpen {
		t.Errorf("Expected cb1 to reject execution, got %v", err1)
	}

	executed := false
	err2 := cb2.Execute(context.Background(), func(ctx context.Context) error {
		executed = true
		return nil
	})
	if err2 != nil {
		t.Errorf("Expected cb2 to allow execution, got %v", err2)
	}
	if !executed {
		t.Error("Expected cb2 to execute function")
	}
}
