package interfaces

import "time"

// Scheduler defines the interface for all scheduler implementations
// This interface is used by the runner to manage execution timing
type Scheduler interface {
	// Next returns a channel that will deliver the next execution time
	Next() <-chan time.Time

	// Stop stops the scheduler and releases any resources
	Stop()
}
