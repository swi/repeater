package scheduler

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type IntervalScheduler struct {
	interval    time.Duration
	jitter      float64
	immediate   bool
	ticker      *time.Ticker
	done        chan struct{}
	stopped     bool
	initialized bool
	tickCh      chan time.Time
	mu          sync.RWMutex // Protects stopped and initialized fields
	stopOnce    sync.Once    // Ensures Stop() is idempotent
}

func NewIntervalScheduler(interval time.Duration, jitter float64, immediate bool) (*IntervalScheduler, error) {
	if interval <= 0 {
		return nil, errors.New("interval must be positive")
	}

	if jitter < 0 || jitter > 1.0 {
		return nil, errors.New("jitter must be between 0 and 1.0")
	}

	return &IntervalScheduler{
		interval:  interval,
		jitter:    jitter,
		immediate: immediate,
		done:      make(chan struct{}),
		tickCh:    make(chan time.Time, 1),
	}, nil
}

func (s *IntervalScheduler) Next() <-chan time.Time {
	s.mu.RLock()
	if s.stopped {
		s.mu.RUnlock()
		// Return channel that will never send
		ch := make(chan time.Time)
		return ch
	}

	if !s.initialized {
		s.mu.RUnlock()
		s.mu.Lock()
		// Double-check after acquiring write lock
		if s.stopped {
			s.mu.Unlock()
			ch := make(chan time.Time)
			return ch
		}
		if !s.initialized {
			s.initialized = true
			s.mu.Unlock()

			// Always send immediate first tick
			s.tickCh <- time.Now()

			// Start ticker for subsequent ticks
			actualInterval := s.calculateInterval()
			s.ticker = time.NewTicker(actualInterval)

			// Start goroutine to forward ticker ticks
			go func() {
				defer s.ticker.Stop()
				for {
					select {
					case t := <-s.ticker.C:
						s.mu.RLock()
						stopped := s.stopped
						s.mu.RUnlock()
						if stopped {
							return
						}
						select {
						case s.tickCh <- t:
						case <-s.done:
							return
						}
					case <-s.done:
						return
					}
				}
			}()
		} else {
			s.mu.Unlock()
		}
	} else {
		s.mu.RUnlock()
	}

	return s.tickCh
}

func (s *IntervalScheduler) calculateInterval() time.Duration {
	actualInterval := s.interval
	if s.jitter > 0 {
		maxJitter := time.Duration(float64(s.interval) * s.jitter)
		jitterAmount := time.Duration(rand.Int63n(int64(maxJitter*2))) - maxJitter
		actualInterval += jitterAmount
		if actualInterval <= 0 {
			actualInterval = s.interval
		}
	}
	return actualInterval
}

func (s *IntervalScheduler) Stop() {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.stopped = true
		s.mu.Unlock()

		if s.ticker != nil {
			s.ticker.Stop()
		}
		close(s.done)
	})
}
