package strategies

import (
	"errors"
	"math/rand"
	"time"
)

// DecorrelatedJitterStrategy implements AWS-recommended decorrelated jitter backoff
// Each delay is calculated based on the previous delay rather than attempt number
// This provides better distribution than simple jitter and prevents thundering herd
type DecorrelatedJitterStrategy struct {
	baseDelay     time.Duration
	multiplier    float64
	maxDelay      time.Duration
	previousDelay time.Duration
	rng           *rand.Rand
}

// NewDecorrelatedJitterStrategy creates a new decorrelated jitter backoff strategy
func NewDecorrelatedJitterStrategy(baseDelay time.Duration, multiplier float64, maxDelay time.Duration) *DecorrelatedJitterStrategy {
	return &DecorrelatedJitterStrategy{
		baseDelay:     baseDelay,
		multiplier:    multiplier,
		maxDelay:      maxDelay,
		previousDelay: baseDelay,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Name returns the strategy name
func (d *DecorrelatedJitterStrategy) Name() string {
	return "decorrelated-jitter"
}

// NextDelay calculates the next delay using decorrelated jitter algorithm
func (d *DecorrelatedJitterStrategy) NextDelay(attempt int, lastDuration time.Duration) time.Duration {
	if attempt <= 1 {
		d.previousDelay = d.baseDelay
		return d.baseDelay
	}

	// AWS decorrelated jitter algorithm:
	// next_delay = random(base_delay, previous_delay * multiplier)
	upperBound := time.Duration(float64(d.previousDelay) * d.multiplier)

	// Ensure we don't go below base delay
	if upperBound < d.baseDelay {
		upperBound = d.baseDelay
	}

	// Calculate random delay between base_delay and upper_bound
	diff := upperBound - d.baseDelay
	if diff <= 0 {
		d.previousDelay = d.baseDelay
		return d.baseDelay
	}

	randomDelay := time.Duration(d.rng.Int63n(int64(diff)))
	delay := d.baseDelay + randomDelay

	// Apply maximum delay cap
	if d.maxDelay > 0 && delay > d.maxDelay {
		delay = d.maxDelay
	}

	// Store for next calculation
	d.previousDelay = delay

	return delay
}

// ShouldRetry determines if we should continue retrying
func (d *DecorrelatedJitterStrategy) ShouldRetry(attempt int, err error, output string) bool {
	// This is handled by the main retry logic based on MaxAttempts
	// Strategy just provides the delay calculation
	return true
}

// ValidateConfig validates the decorrelated jitter strategy configuration
func (d *DecorrelatedJitterStrategy) ValidateConfig(config *StrategyConfig) error {
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.Multiplier > 10.0 {
		return errors.New("multiplier must be <= 10.0 to prevent excessive delays")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	if config.MaxAttempts <= 0 {
		return errors.New("attempts must be positive")
	}

	return nil
}
