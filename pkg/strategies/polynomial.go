package strategies

import (
	"errors"
	"math"
	"time"
)

// PolynomialStrategy implements polynomial backoff: base_delay * attempt^exponent
// Provides customizable growth patterns with configurable exponent
type PolynomialStrategy struct {
	baseDelay time.Duration
	exponent  float64
	maxDelay  time.Duration
}

// NewPolynomialStrategy creates a new polynomial backoff strategy
func NewPolynomialStrategy(baseDelay time.Duration, exponent float64, maxDelay time.Duration) *PolynomialStrategy {
	return &PolynomialStrategy{
		baseDelay: baseDelay,
		exponent:  exponent,
		maxDelay:  maxDelay,
	}
}

// Name returns the strategy name
func (p *PolynomialStrategy) Name() string {
	return "polynomial"
}

// NextDelay calculates the next delay using polynomial growth
func (p *PolynomialStrategy) NextDelay(attempt int, lastDuration time.Duration) time.Duration {
	if attempt <= 0 {
		return p.baseDelay
	}

	// Calculate polynomial growth: base_delay * attempt^exponent
	multiplier := math.Pow(float64(attempt), p.exponent)
	delay := time.Duration(float64(p.baseDelay) * multiplier)

	// Apply maximum delay cap
	if p.maxDelay > 0 && delay > p.maxDelay {
		delay = p.maxDelay
	}

	return delay
}

// ShouldRetry determines if we should continue retrying
func (p *PolynomialStrategy) ShouldRetry(attempt int, err error, output string) bool {
	// This is handled by the main retry logic based on MaxAttempts
	// Strategy just provides the delay calculation
	return true
}

// ValidateConfig validates the polynomial strategy configuration
func (p *PolynomialStrategy) ValidateConfig(config *StrategyConfig) error {
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Exponent <= 0 {
		return errors.New("exponent must be positive")
	}

	if config.Exponent > 10 {
		return errors.New("exponent must be <= 10 to prevent overflow")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	if config.MaxAttempts <= 0 {
		return errors.New("attempts must be positive")
	}

	return nil
}
