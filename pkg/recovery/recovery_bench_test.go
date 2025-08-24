package recovery

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

// BenchmarkRecoveryManagerMemory tests memory usage of RecoveryManager operations
func BenchmarkRecoveryManagerMemory(b *testing.B) {
	manager := NewRecoveryManager()
	policy := NewExponentialBackoffPolicy(5, 10*time.Millisecond, 2.0, time.Second)
	manager.SetRetryPolicy(policy)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()

		// Test memory allocation for retry operations
		execFunc := func(ctx context.Context) error {
			if i%3 == 0 {
				return nil // Success case
			}
			return errors.New("test error")
		}

		_ = manager.ExecuteWithRetry(ctx, execFunc)
	}
}

// BenchmarkRetryPolicyMemory tests memory usage of retry policy operations
func BenchmarkRetryPolicyMemory(b *testing.B) {
	policy := NewExponentialBackoffPolicy(10, time.Millisecond, 2.0, 10*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		attempt := i%10 + 1 // Cycle through attempts 1-10
		shouldRetry := policy.ShouldRetry(attempt)
		delay := policy.NextDelay(attempt)
		testErr := errors.New("test error")
		shouldRetryErr := policy.ShouldRetryError(testErr)
		_ = shouldRetry
		_ = delay
		_ = shouldRetryErr
	}
}

// BenchmarkCircuitBreakerMemory tests memory usage of circuit breaker operations
func BenchmarkCircuitBreakerMemory(b *testing.B) {
	cb := NewCircuitBreaker("test", 5, 30*time.Second, time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Test circuit breaker state checking
		state := cb.State()
		_ = state

		// Simulate success/failure recording
		if i%2 == 0 {
			cb.RecordSuccess()
		} else {
			cb.RecordFailure()
		}
	}
}

// BenchmarkErrorReporterMemory tests memory usage of error reporting
func BenchmarkErrorReporterMemory(b *testing.B) {
	var buf bytes.Buffer
	reporter := NewErrorReporter(&buf)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testErr := errors.New("benchmark test error")
		reporter.ReportError(testErr)
	}
}

// BenchmarkRecoveryManagerConcurrentMemory tests memory usage under concurrent access
func BenchmarkRecoveryManagerConcurrentMemory(b *testing.B) {
	manager := NewRecoveryManager()
	policy := NewExponentialBackoffPolicy(3, time.Millisecond, 2.0, 100*time.Millisecond)
	manager.SetRetryPolicy(policy)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		attempt := 0
		for pb.Next() {
			attempt++
			ctx := context.Background()

			execFunc := func(ctx context.Context) error {
				if attempt%4 == 0 {
					return nil // Occasional success
				}
				return errors.New("concurrent test error")
			}

			_ = manager.ExecuteWithRetry(ctx, execFunc)
		}
	})
}
