package recovery

import (
	"time"
)

// RetryPolicy defines the interface for retry policies
type RetryPolicy interface {
	ShouldRetry(attempt int) bool
	NextDelay(attempt int) time.Duration
	ShouldRetryError(err error) bool
}

// ExponentialBackoffPolicy implements exponential backoff retry policy
type ExponentialBackoffPolicy struct {
	maxRetries   int
	initialDelay time.Duration
	multiplier   float64
	maxDelay     time.Duration
}

// NewExponentialBackoffPolicy creates a new exponential backoff policy
func NewExponentialBackoffPolicy(maxRetries int, initialDelay time.Duration, multiplier float64, maxDelay time.Duration) *ExponentialBackoffPolicy {
	return &ExponentialBackoffPolicy{
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		multiplier:   multiplier,
		maxDelay:     maxDelay,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *ExponentialBackoffPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay calculates the delay for the next retry attempt
func (p *ExponentialBackoffPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	delay := p.initialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * p.multiplier)
	}

	if delay > p.maxDelay {
		delay = p.maxDelay
	}

	return delay
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *ExponentialBackoffPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// LinearBackoffPolicy implements linear backoff retry policy
type LinearBackoffPolicy struct {
	maxRetries   int
	initialDelay time.Duration
	increment    time.Duration
}

// NewLinearBackoffPolicy creates a new linear backoff policy
func NewLinearBackoffPolicy(maxRetries int, initialDelay time.Duration, increment time.Duration) *LinearBackoffPolicy {
	return &LinearBackoffPolicy{
		maxRetries:   maxRetries,
		initialDelay: initialDelay,
		increment:    increment,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *LinearBackoffPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay calculates the delay for the next retry attempt
func (p *LinearBackoffPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	return p.initialDelay + time.Duration(attempt-1)*p.increment
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *LinearBackoffPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// FixedDelayPolicy implements fixed delay retry policy
type FixedDelayPolicy struct {
	maxRetries int
	delay      time.Duration
}

// NewFixedDelayPolicy creates a new fixed delay policy
func NewFixedDelayPolicy(maxRetries int, delay time.Duration) *FixedDelayPolicy {
	return &FixedDelayPolicy{
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// ShouldRetry returns true if the attempt should be retried
func (p *FixedDelayPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay returns the fixed delay for retry attempts
func (p *FixedDelayPolicy) NextDelay(attempt int) time.Duration {
	return p.delay
}

// ShouldRetryError returns true if the error should be retried (default: all errors)
func (p *FixedDelayPolicy) ShouldRetryError(err error) bool {
	return true // Default: retry all errors
}

// ConditionalRetryPolicy implements conditional retry policy
type ConditionalRetryPolicy struct {
	maxRetries int
	delay      time.Duration
	conditions []func(error) bool
}

// NewConditionalRetryPolicy creates a new conditional retry policy
func NewConditionalRetryPolicy(maxRetries int, delay time.Duration) *ConditionalRetryPolicy {
	return &ConditionalRetryPolicy{
		maxRetries: maxRetries,
		delay:      delay,
		conditions: make([]func(error) bool, 0),
	}
}

// AddCondition adds a condition for retrying errors
func (p *ConditionalRetryPolicy) AddCondition(condition func(error) bool) {
	p.conditions = append(p.conditions, condition)
}

// ShouldRetry returns true if the attempt should be retried
func (p *ConditionalRetryPolicy) ShouldRetry(attempt int) bool {
	return attempt <= p.maxRetries
}

// NextDelay returns the delay for retry attempts
func (p *ConditionalRetryPolicy) NextDelay(attempt int) time.Duration {
	return p.delay
}

// ShouldRetryError returns true if the error matches any retry condition
func (p *ConditionalRetryPolicy) ShouldRetryError(err error) bool {
	for _, condition := range p.conditions {
		if condition(err) {
			return true
		}
	}
	return false
}
