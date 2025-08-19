// Package interfaces defines contracts for core Repeater components.
//
// This package contains only interface definitions and requires no tests.
// Interface implementations are tested in their respective packages:
//   - pkg/scheduler/* - Basic scheduling implementations
//   - pkg/adaptive/* - Adaptive scheduling implementations
//   - pkg/strategies/* - Mathematical retry strategies
package interfaces

import (
	"context"
	"time"
)

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

// ExecutionCoordinator defines the interface for coordinating command execution
// with scheduling, monitoring, and observability features.
type ExecutionCoordinator interface {
	// Execute runs a command with the configured scheduler and monitoring
	Execute(ctx context.Context, command []string) (*ExecutionStats, error)

	// GetScheduler returns the underlying scheduler
	GetScheduler() Scheduler

	// Stop gracefully stops the coordinator and all associated resources
	Stop() error
}

// ExecutionStats represents statistics from a complete execution run
type ExecutionStats struct {
	TotalExecutions      int
	SuccessfulExecutions int
	FailedExecutions     int
	Duration             time.Duration
	StartTime            time.Time
	EndTime              time.Time
	Executions           []ExecutionRecord
}

// ExecutionRecord represents a single command execution
type ExecutionRecord struct {
	ExecutionNumber int
	ExitCode        int
	Duration        time.Duration
	Stdout          string
	Stderr          string
	StartTime       time.Time
	EndTime         time.Time
}
