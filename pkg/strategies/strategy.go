package strategies

import (
	"time"
)

// Strategy defines the interface for retry strategies
type Strategy interface {
	// Name returns the strategy name
	Name() string

	// NextDelay calculates the next delay duration for the given attempt
	// attempt: 1-based attempt number (1 = first retry, 2 = second retry, etc.)
	// lastDuration: duration of the last execution (for adaptive strategies)
	NextDelay(attempt int, lastDuration time.Duration) time.Duration

	// ShouldRetry determines if we should retry based on attempt count and result
	// attempt: 1-based attempt number
	// err: error from command execution (nil if command succeeded)
	// output: command output for pattern matching
	ShouldRetry(attempt int, err error, output string) bool

	// ValidateConfig validates the strategy configuration
	ValidateConfig(config *StrategyConfig) error
}

// StrategyConfig holds configuration for all strategies
type StrategyConfig struct {
	// Common retry parameters
	MaxAttempts     int           // maximum retry attempts
	Timeout         time.Duration // per-attempt timeout
	SuccessPattern  string        // regex pattern for success detection
	FailurePattern  string        // regex pattern for failure detection
	CaseInsensitive bool          // case-insensitive pattern matching

	// Mathematical strategy parameters
	BaseDelay  time.Duration // initial/base delay
	Increment  time.Duration // linear increment
	Multiplier float64       // exponential/jitter multiplier
	Exponent   float64       // polynomial exponent
	MaxDelay   time.Duration // maximum delay cap

	// Adaptive strategy parameters
	LearningRate     float64       // learning rate (0.01-1.0)
	MemoryWindow     int           // number of outcomes to remember
	MinInterval      time.Duration // minimum interval bound
	MaxInterval      time.Duration // maximum interval bound
	SlowThreshold    float64       // slow response threshold
	FastThreshold    float64       // fast response threshold
	FailureThreshold float64       // circuit breaker threshold
}

// DefaultConfig returns a default strategy configuration
func DefaultConfig() *StrategyConfig {
	return &StrategyConfig{
		MaxAttempts:      3,
		Timeout:          0, // no timeout by default
		BaseDelay:        1 * time.Second,
		Increment:        1 * time.Second,
		Multiplier:       2.0,
		Exponent:         2.0,
		MaxDelay:         60 * time.Second,
		LearningRate:     0.1,
		MemoryWindow:     50,
		MinInterval:      500 * time.Millisecond,
		MaxInterval:      60 * time.Second,
		SlowThreshold:    1.5,
		FastThreshold:    0.8,
		FailureThreshold: 0.5,
	}
}
