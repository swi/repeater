// Package interfaces defines contracts for core Repeater components.
//
// This package contains only interface definitions and requires no tests.
// Interface implementations are tested in their respective packages:
//   - pkg/scheduler/* - Basic scheduling implementations
//   - pkg/adaptive/* - Adaptive scheduling implementations
//   - pkg/strategies/* - Mathematical retry strategies
package interfaces

import "time"

// Scheduler defines the interface for all scheduler implementations.
// This interface is used by the runner to manage execution timing.
//
// All Scheduler implementations must be thread-safe and handle
// concurrent calls to Next() and Stop().
type Scheduler interface {
	// Next returns a channel that will deliver the next execution time
	Next() <-chan time.Time

	// Stop stops the scheduler and releases any resources
	Stop()
}
