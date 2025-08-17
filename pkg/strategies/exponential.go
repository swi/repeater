package strategies

import (
	"errors"
	"math"
	"time"
)

// ExponentialStrategy implements exponential backoff: base_delay * multiplier^(attempt-1)
// Industry standard for network operations and API calls: 1s, 2s, 4s, 8s, 16s...
type ExponentialStrategy struct {
	baseDelay  time.Duration
	multiplier float64
	maxDelay   time.Duration
}

// NewExponentialStrategy creates a new exponential backoff strategy
func NewExponentialStrategy(baseDelay time.Duration, multiplier float64, maxDelay time.Duration) *ExponentialStrategy {
	return &ExponentialStrategy{
		baseDelay:  baseDelay,
		multiplier: multiplier,
		maxDelay:   maxDelay,
	}
}

// Name returns the strategy name
func (e *ExponentialStrategy) Name() string {
	return "exponential"
}

// NextDelay calculates the next delay using exponential growth
func (e *ExponentialStrategy) NextDelay(attempt int, lastDuration time.Duration) time.Duration {
	if attempt <= 1 {
		return e.baseDelay
	}

	// Calculate exponential growth: base_delay * multiplier^(attempt-1)
	multiplier := math.Pow(e.multiplier, float64(attempt-1))
	delay := time.Duration(float64(e.baseDelay) * multiplier)

	// Apply maximum delay cap
	if e.maxDelay > 0 && delay > e.maxDelay {
		delay = e.maxDelay
	}

	return delay
}

// ShouldRetry determines if we should continue retrying
func (e *ExponentialStrategy) ShouldRetry(attempt int, err error, output string) bool {
	// This is handled by the main retry logic based on MaxAttempts
	// Strategy just provides the delay calculation
	return true
}

// ValidateConfig validates the exponential strategy configuration
func (e *ExponentialStrategy) ValidateConfig(config *StrategyConfig) error {
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.Multiplier > 10.0 {
		return errors.New("multiplier must be <= 10.0 to prevent overflow")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	if config.MaxAttempts <= 0 {
		return errors.New("attempts must be positive")
	}

	return nil
}
