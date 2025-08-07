package scheduler

import (
	"errors"
	"math/rand"
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
	if s.stopped {
		// Return channel that will never send
		ch := make(chan time.Time)
		return ch
	}

	if !s.initialized {
		s.initialized = true

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
					if s.stopped {
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
	if s.stopped {
		return
	}
	s.stopped = true
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.done)
}
