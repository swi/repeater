package scheduler

import (
	"time"

	"github.com/swi/repeater/pkg/strategies"
)

// StrategyScheduler implements retry scheduling using mathematical strategies
type StrategyScheduler struct {
	strategy       strategies.Strategy
	config         *strategies.StrategyConfig
	currentAttempt int
	lastDuration   time.Duration
	maxAttempts    int
	nextChan       chan time.Time
	stopChan       chan struct{}
	stopped        bool
}

// NewStrategyScheduler creates a new strategy-based scheduler
func NewStrategyScheduler(strategy strategies.Strategy, config *strategies.StrategyConfig) (*StrategyScheduler, error) {
	// Validate the strategy configuration
	if err := strategy.ValidateConfig(config); err != nil {
		return nil, err
	}

	return &StrategyScheduler{
		strategy:       strategy,
		config:         config,
		currentAttempt: 0,
		maxAttempts:    config.MaxAttempts,
		nextChan:       make(chan time.Time, 1),
		stopChan:       make(chan struct{}),
		stopped:        false,
	}, nil
}

// Next returns a channel that delivers the next execution time
func (s *StrategyScheduler) Next() <-chan time.Time {
	if s.stopped {
		return s.nextChan
	}

	go func() {
		s.currentAttempt++

		// Check if we've exceeded max attempts
		if s.maxAttempts > 0 && s.currentAttempt > s.maxAttempts {
			s.Stop()
			return
		}

		var delay time.Duration
		if s.currentAttempt == 1 {
			// First attempt - execute immediately
			delay = 0
		} else {
			// Calculate retry delay using the strategy
			delay = s.strategy.NextDelay(s.currentAttempt-1, s.lastDuration)
		}

		// Schedule the next execution
		go func() {
			if delay > 0 {
				timer := time.NewTimer(delay)
				defer timer.Stop()

				select {
				case <-timer.C:
					if !s.stopped {
						select {
						case s.nextChan <- time.Now():
						case <-s.stopChan:
							return
						}
					}
				case <-s.stopChan:
					return
				}
			} else {
				// Immediate execution
				if !s.stopped {
					select {
					case s.nextChan <- time.Now():
					case <-s.stopChan:
						return
					}
				}
			}
		}()
	}()

	return s.nextChan
}

// Stop stops the scheduler
func (s *StrategyScheduler) Stop() {
	if !s.stopped {
		s.stopped = true
		close(s.stopChan)
	}
}

// UpdateExecutionResult updates the scheduler with the result of the last execution
// This allows adaptive strategies to learn from execution results
func (s *StrategyScheduler) UpdateExecutionResult(duration time.Duration, success bool, output string) {
	s.lastDuration = duration

	// If the command succeeded and we're in retry mode, we should stop
	if success {
		s.Stop()
	}
}

// IsRetryMode returns true if this scheduler is designed for retry (until success)
func (s *StrategyScheduler) IsRetryMode() bool {
	return true // Strategy schedulers are designed for retry-until-success
}

// GetAttemptNumber returns the current attempt number
func (s *StrategyScheduler) GetAttemptNumber() int {
	return s.currentAttempt
}

// GetStrategy returns the underlying strategy
func (s *StrategyScheduler) GetStrategy() strategies.Strategy {
	return s.strategy
}
