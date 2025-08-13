package httpaware

import (
	"time"

	"github.com/swi/repeater/pkg/scheduler"
)

// TimingSource indicates where timing information came from
type TimingSource int

const (
	TimingSourceRetryAfterHeader TimingSource = iota
	TimingSourceJSONRetryAfter
	TimingSourceJSONRateLimit
	TimingSourceJSONBackoff
)

// String returns a string representation of the timing source
func (ts TimingSource) String() string {
	switch ts {
	case TimingSourceRetryAfterHeader:
		return "retry-after-header"
	case TimingSourceJSONRetryAfter:
		return "json-retry-after"
	case TimingSourceJSONRateLimit:
		return "json-rate-limit"
	case TimingSourceJSONBackoff:
		return "json-backoff"
	default:
		return "unknown"
	}
}

// TimingInfo represents extracted timing information from HTTP responses
type TimingInfo struct {
	Delay      time.Duration // How long to wait before next execution
	Source     TimingSource  // Where the timing info came from
	Confidence float64       // How confident we are in this timing (0.0-1.0)
}

// HTTPResponseParser extracts timing information from HTTP responses
type HTTPResponseParser interface {
	// ParseResponse extracts timing information from an HTTP response
	ParseResponse(response string) (*TimingInfo, error)

	// SupportsResponse returns true if the response appears to be HTTP
	SupportsResponse(response string) bool
}

// HTTPAwareScheduler combines HTTP intelligence with fallback scheduling
type HTTPAwareScheduler interface {
	scheduler.Scheduler
	SetLastResponse(response string)
	SetFallbackScheduler(fallback scheduler.Scheduler)
	GetTimingInfo() *TimingInfo
	NextDelay() time.Duration // For testing purposes
}

// HTTPAwareConfig configures HTTP-aware scheduling behavior
type HTTPAwareConfig struct {
	MaxDelay         time.Duration // Maximum delay cap
	MinDelay         time.Duration // Minimum delay floor
	FallbackStrategy string        // "interval", "exponential", etc.
	FallbackConfig   interface{}   // Config for fallback scheduler

	// HTTP parsing options
	ParseJSON         bool // Whether to parse JSON responses
	ParseHeaders      bool // Whether to parse HTTP headers
	TrustClientErrors bool // Whether to trust 4xx retry timing

	// Timing extraction patterns
	JSONFields  []string // Custom JSON fields to check
	HeaderNames []string // Custom header names to check
}
