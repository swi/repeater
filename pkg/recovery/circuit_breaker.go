package recovery

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitBreakerOpen is returned when the circuit breaker is open
var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

// CircuitBreakerStatistics contains statistics about circuit breaker operations
type CircuitBreakerStatistics struct {
	Name            string              `json:"name"`
	State           CircuitBreakerState `json:"state"`
	TotalRequests   int64               `json:"total_requests"`
	SuccessCount    int64               `json:"success_count"`
	FailureCount    int64               `json:"failure_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	LastSuccessTime time.Time           `json:"last_success_time"`
	StateChangedAt  time.Time           `json:"state_changed_at"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	name             string
	failureThreshold int
	timeout          time.Duration
	resetTimeout     time.Duration
	state            CircuitBreakerState
	failureCount     int
	successCount     int64
	totalRequests    int64
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	stateChangedAt   time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold int, timeout time.Duration, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
		stateChangedAt:   time.Now(),
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn ExecuteFunc) error {
	// Check if we can execute
	if !cb.canExecute() {
		return ErrCircuitBreakerOpen
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	if err != nil {
		cb.RecordFailure()
	} else {
		cb.RecordSuccess()
	}

	return err
}

// canExecute determines if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed to transition to Half-Open
		if time.Since(cb.stateChangedAt) >= cb.timeout {
			cb.state = StateHalfOpen
			cb.stateChangedAt = time.Now()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.totalRequests++
	cb.lastSuccessTime = time.Now()

	// Reset failure count and transition to Closed if in Half-Open
	if cb.state == StateHalfOpen {
		cb.failureCount = 0
		cb.state = StateClosed
		cb.stateChangedAt = time.Now()
	}
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.totalRequests++
	cb.lastFailureTime = time.Now()

	// Check if we should transition to Open
	if cb.failureCount >= cb.failureThreshold {
		if cb.state != StateOpen {
			cb.state = StateOpen
			cb.stateChangedAt = time.Now()
		}
	}
}

// Reset resets the circuit breaker to Closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.stateChangedAt = time.Now()
}

// Statistics returns the current statistics of the circuit breaker
func (cb *CircuitBreaker) Statistics() *CircuitBreakerStatistics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return &CircuitBreakerStatistics{
		Name:            cb.name,
		State:           cb.state,
		TotalRequests:   cb.totalRequests,
		SuccessCount:    cb.successCount,
		FailureCount:    int64(cb.failureCount),
		LastFailureTime: cb.lastFailureTime,
		LastSuccessTime: cb.lastSuccessTime,
		StateChangedAt:  cb.stateChangedAt,
	}
}
