package strategies

import (
	"errors"
	"time"
)

// FibonacciStrategy implements Fibonacci backoff: 1, 1, 2, 3, 5, 8, 13, 21...
// Provides moderate growth between linear and exponential backoff
type FibonacciStrategy struct {
	baseDelay time.Duration
	maxDelay  time.Duration
}

// NewFibonacciStrategy creates a new Fibonacci backoff strategy
func NewFibonacciStrategy(baseDelay, maxDelay time.Duration) *FibonacciStrategy {
	return &FibonacciStrategy{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

// Name returns the strategy name
func (f *FibonacciStrategy) Name() string {
	return "fibonacci"
}

// NextDelay calculates the next delay using Fibonacci sequence
func (f *FibonacciStrategy) NextDelay(attempt int, lastDuration time.Duration) time.Duration {
	if attempt <= 0 {
		return f.baseDelay
	}

	// Calculate Fibonacci number for the attempt
	fibNumber := f.fibonacci(attempt)

	// Apply to base delay
	delay := time.Duration(fibNumber) * f.baseDelay

	// Apply maximum delay cap
	if f.maxDelay > 0 && delay > f.maxDelay {
		delay = f.maxDelay
	}

	return delay
}

// ShouldRetry determines if we should continue retrying
func (f *FibonacciStrategy) ShouldRetry(attempt int, err error, output string) bool {
	// This is handled by the main retry logic based on MaxAttempts
	// Strategy just provides the delay calculation
	return true
}

// ValidateConfig validates the Fibonacci strategy configuration
func (f *FibonacciStrategy) ValidateConfig(config *StrategyConfig) error {
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	if config.MaxAttempts <= 0 {
		return errors.New("attempts must be positive")
	}

	return nil
}

// fibonacci calculates the nth Fibonacci number (1-based indexing)
// fibonacci(1) = 1, fibonacci(2) = 1, fibonacci(3) = 2, fibonacci(4) = 3, etc.
func (f *FibonacciStrategy) fibonacci(n int) int64 {
	if n <= 0 {
		return 1
	}
	if n == 1 || n == 2 {
		return 1
	}

	// Calculate iteratively to avoid stack overflow for large n
	a, b := int64(1), int64(1)
	for i := 3; i <= n; i++ {
		a, b = b, a+b
	}

	return b
}
