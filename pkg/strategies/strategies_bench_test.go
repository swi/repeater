package strategies

import (
	"testing"
	"time"
)

// BenchmarkExponentialStrategy_NextDelay benchmarks exponential backoff calculations
func BenchmarkExponentialStrategy_NextDelay(b *testing.B) {
	strategy := NewExponentialStrategy(100*time.Millisecond, 2.0, 60*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strategy.NextDelay(5, 50*time.Millisecond) // Test with attempt 5, last duration 50ms
	}
}

// BenchmarkFibonacciStrategy_NextDelay benchmarks fibonacci sequence calculations
func BenchmarkFibonacciStrategy_NextDelay(b *testing.B) {
	strategy := NewFibonacciStrategy(100*time.Millisecond, 60*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strategy.NextDelay(10, 200*time.Millisecond) // Test with attempt 10, last duration 200ms
	}
}

// BenchmarkLinearStrategy_NextDelay benchmarks linear backoff calculations
func BenchmarkLinearStrategy_NextDelay(b *testing.B) {
	strategy := NewLinearStrategy(100*time.Millisecond, 60*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strategy.NextDelay(5, 50*time.Millisecond) // Test with attempt 5, last duration 50ms
	}
}

// BenchmarkPolynomialStrategy_NextDelay benchmarks polynomial backoff calculations
func BenchmarkPolynomialStrategy_NextDelay(b *testing.B) {
	strategy := NewPolynomialStrategy(100*time.Millisecond, 2.5, 60*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strategy.NextDelay(5, 50*time.Millisecond) // Test with attempt 5, last duration 50ms
	}
}

// BenchmarkDecorrelatedJitterStrategy_NextDelay benchmarks decorrelated jitter calculations
func BenchmarkDecorrelatedJitterStrategy_NextDelay(b *testing.B) {
	strategy := NewDecorrelatedJitterStrategy(100*time.Millisecond, 3.0, 60*time.Second)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = strategy.NextDelay(5, 50*time.Millisecond) // Test with attempt 5, last duration 50ms
	}
}

// BenchmarkAllStrategies_Comparison compares all strategies performance
func BenchmarkAllStrategies_Comparison(b *testing.B) {
	strategies := map[string]Strategy{
		"Exponential":        NewExponentialStrategy(100*time.Millisecond, 2.0, 60*time.Second),
		"Fibonacci":          NewFibonacciStrategy(100*time.Millisecond, 60*time.Second),
		"Linear":             NewLinearStrategy(100*time.Millisecond, 60*time.Second),
		"Polynomial":         NewPolynomialStrategy(100*time.Millisecond, 2.5, 60*time.Second),
		"DecorrelatedJitter": NewDecorrelatedJitterStrategy(100*time.Millisecond, 3.0, 60*time.Second),
	}

	for name, strategy := range strategies {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = strategy.NextDelay(5, 50*time.Millisecond)
			}
		})
	}
}
