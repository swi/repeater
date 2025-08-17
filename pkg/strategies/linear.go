package strategies

import (
	"errors"
	"time"
)

// LinearStrategy implements linear backoff: increment, 2*increment, 3*increment...
// Provides predictable, incremental delays ideal for rate-limited APIs
type LinearStrategy struct {
	increment time.Duration
	maxDelay  time.Duration
}

// NewLinearStrategy creates a new linear backoff strategy
func NewLinearStrategy(increment, maxDelay time.Duration) *LinearStrategy {
	return &LinearStrategy{
		increment: increment,
		maxDelay:  maxDelay,
	}
}

// Name returns the strategy name
func (l *LinearStrategy) Name() string {
	return "linear"
}

// NextDelay calculates the next delay using linear progression
func (l *LinearStrategy) NextDelay(attempt int, lastDuration time.Duration) time.Duration {
	if attempt <= 0 {
		return l.increment
	}

	// Linear progression: attempt * increment
	delay := time.Duration(attempt) * l.increment

	// Apply maximum delay cap
	if l.maxDelay > 0 && delay > l.maxDelay {
		delay = l.maxDelay
	}

	return delay
}

// ShouldRetry determines if we should continue retrying
func (l *LinearStrategy) ShouldRetry(attempt int, err error, output string) bool {
	// This is handled by the main retry logic based on MaxAttempts
	// Strategy just provides the delay calculation
	return true
}

// ValidateConfig validates the linear strategy configuration
func (l *LinearStrategy) ValidateConfig(config *StrategyConfig) error {
	if config.Increment <= 0 {
		return errors.New("increment must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.Increment {
		return errors.New("max-delay must be greater than increment")
	}

	if config.MaxAttempts <= 0 {
		return errors.New("attempts must be positive")
	}

	return nil
}
