package httpaware

import (
	"time"

	"github.com/swi/repeater/pkg/scheduler"
)

// httpAwareScheduler implements HTTPAwareScheduler interface
type httpAwareScheduler struct {
	config            HTTPAwareConfig
	parser            HTTPResponseParser
	fallbackScheduler scheduler.Scheduler
	lastResponse      string
	lastTimingInfo    *TimingInfo
	nextCh            chan time.Time
	stopped           bool
}

// NewHTTPAwareScheduler creates a new HTTP-aware scheduler with default configuration
func NewHTTPAwareScheduler(config HTTPAwareConfig) HTTPAwareScheduler {
	return &httpAwareScheduler{
		config: config,
		parser: NewHTTPResponseParser(),
		nextCh: make(chan time.Time, 1),
	}
}

// NewHTTPAwareSchedulerWithConfig creates a new HTTP-aware scheduler with custom configuration
func NewHTTPAwareSchedulerWithConfig(config HTTPAwareConfig) HTTPAwareScheduler {
	return &httpAwareScheduler{
		config: config,
		parser: NewHTTPResponseParserWithConfig(config),
		nextCh: make(chan time.Time, 1),
	}
}

// Next returns a channel that will send the next execution time
func (s *httpAwareScheduler) Next() <-chan time.Time {
	// TODO: Implement actual scheduling logic
	// This is a stub implementation that will make tests fail (RED phase)
	if s.fallbackScheduler != nil {
		return s.fallbackScheduler.Next()
	}
	return s.nextCh
}

// Stop stops the scheduler
func (s *httpAwareScheduler) Stop() {
	s.stopped = true
	if s.fallbackScheduler != nil {
		s.fallbackScheduler.Stop()
	}
	if s.nextCh != nil {
		close(s.nextCh)
	}
}

// SetLastResponse sets the last HTTP response for timing extraction
func (s *httpAwareScheduler) SetLastResponse(response string) {
	s.lastResponse = response

	// Parse the response to extract timing information
	if timingInfo, err := s.parser.ParseResponse(response); err == nil && timingInfo != nil {
		// Apply constraints
		delay := timingInfo.Delay

		// Apply minimum delay
		if delay < s.config.MinDelay {
			delay = s.config.MinDelay
		}

		// Apply maximum delay cap
		if s.config.MaxDelay > 0 && delay > s.config.MaxDelay {
			delay = s.config.MaxDelay
		}

		// Update timing info with constrained delay
		s.lastTimingInfo = &TimingInfo{
			Delay:      delay,
			Source:     timingInfo.Source,
			Confidence: timingInfo.Confidence,
		}
	} else {
		s.lastTimingInfo = nil
	}
}

// SetFallbackScheduler sets the fallback scheduler to use when no HTTP timing is available
func (s *httpAwareScheduler) SetFallbackScheduler(fallback scheduler.Scheduler) {
	s.fallbackScheduler = fallback
}

// GetTimingInfo returns the last extracted timing information
func (s *httpAwareScheduler) GetTimingInfo() *TimingInfo {
	return s.lastTimingInfo
}

// NextDelay returns the next delay duration (for testing purposes)
func (s *httpAwareScheduler) NextDelay() time.Duration {
	// If we have HTTP timing information, use it
	if s.lastTimingInfo != nil {
		return s.lastTimingInfo.Delay
	}

	// Otherwise, use fallback scheduler
	if s.fallbackScheduler != nil {
		// For testing, we need to simulate getting the delay from fallback
		// This is a simplified approach - in real usage, we'd use the Next() channel
		return s.simulateFallbackDelay()
	}

	return 1 * time.Second
}

// TestableScheduler interface for testing purposes
type TestableScheduler interface {
	scheduler.Scheduler
	GetNextDelay() time.Duration
}

// simulateFallbackDelay simulates getting delay from fallback scheduler for testing
func (s *httpAwareScheduler) simulateFallbackDelay() time.Duration {
	// This is a testing helper - in real usage we'd use the scheduler's Next() channel

	// Check if it's a testable scheduler
	if testable, ok := s.fallbackScheduler.(TestableScheduler); ok {
		return testable.GetNextDelay()
	}

	// Default fallback
	return 10 * time.Second
}
