package scheduler

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// ExponentialBackoffScheduler implements exponential backoff with jitter
type ExponentialBackoffScheduler struct {
	mu              sync.RWMutex
	initial         time.Duration
	multiplier      float64
	maxInterval     time.Duration
	jitter          float64
	failures        int
	currentInterval time.Duration
	nextChan        chan time.Time
	stopChan        chan struct{}
	stopped         bool
}

// NewExponentialBackoffScheduler creates a new exponential backoff scheduler
func NewExponentialBackoffScheduler(initial time.Duration, multiplier float64, maxInterval time.Duration, jitter float64) *ExponentialBackoffScheduler {
	s := &ExponentialBackoffScheduler{
		initial:         initial,
		multiplier:      multiplier,
		maxInterval:     maxInterval,
		jitter:          jitter,
		failures:        0,
		currentInterval: initial,
		nextChan:        make(chan time.Time, 1),
		stopChan:        make(chan struct{}),
		stopped:         false,
	}

	go s.scheduleLoop()
	return s
}

// RecordFailure records a failure and updates the backoff interval
func (s *ExponentialBackoffScheduler) RecordFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failures++
	s.updateInterval()
}

// RecordSuccess resets the backoff to initial interval
func (s *ExponentialBackoffScheduler) RecordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failures = 0
	s.currentInterval = s.initial
}

// GetCurrentInterval returns the current backoff interval
func (s *ExponentialBackoffScheduler) GetCurrentInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentInterval
}

// updateInterval calculates the new interval with exponential backoff and jitter
func (s *ExponentialBackoffScheduler) updateInterval() {
	// Calculate exponential backoff: initial * multiplier^failures
	backoff := float64(s.initial) * math.Pow(s.multiplier, float64(s.failures))

	// Apply maximum cap
	if backoff > float64(s.maxInterval) {
		backoff = float64(s.maxInterval)
	}

	// Apply jitter if configured
	if s.jitter > 0 {
		jitterAmount := backoff * s.jitter
		jitterOffset := (rand.Float64()*2 - 1) * jitterAmount // -jitter to +jitter
		backoff += jitterOffset

		// Ensure we don't go below 0 or above max
		if backoff < 0 {
			backoff = float64(s.initial)
		}
		if backoff > float64(s.maxInterval) {
			backoff = float64(s.maxInterval)
		}
	}

	s.currentInterval = time.Duration(backoff)
}

// Next returns a channel that delivers the next execution time
func (s *ExponentialBackoffScheduler) Next() <-chan time.Time {
	return s.nextChan
}

// scheduleLoop continuously schedules the next execution
func (s *ExponentialBackoffScheduler) scheduleLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			interval := s.GetCurrentInterval()

			select {
			case <-time.After(interval):
				select {
				case s.nextChan <- time.Now():
					// Successfully sent
				case <-s.stopChan:
					return
				}
			case <-s.stopChan:
				return
			}
		}
	}
}

// Stop stops the scheduler
func (s *ExponentialBackoffScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.stopped {
		close(s.stopChan)
		s.stopped = true
	}
}
