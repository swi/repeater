package plugin

import (
	"context"
	"time"
)

// Scheduler interface represents a command scheduler (imported from scheduler package concept)
type Scheduler interface {
	Next() <-chan time.Time
	Stop()
}

// SchedulerPlugin interface defines the contract for scheduler plugins
type SchedulerPlugin interface {
	Name() string
	Version() string
	ValidateConfig(config map[string]interface{}) error
	Create(config map[string]interface{}) (Scheduler, error)
}

// ExecutorPlugin interface defines the contract for executor plugins
type ExecutorPlugin interface {
	Name() string
	Execute(ctx context.Context, cmd []string, opts ExecutorOptions) (*ExecutionResult, error)
	SupportsStreaming() bool
	SupportedPlatforms() []string
}

// OutputPlugin interface defines the contract for output plugins
type OutputPlugin interface {
	Name() string
	ProcessOutput(result *ExecutionResult, config OutputConfig) error
	SupportsStreaming() bool
	RequiredConfig() []string
}

// ExecutorOptions represents options for command execution
type ExecutorOptions struct {
	Timeout time.Duration
}

// ExecutionResult represents the result of command execution
type ExecutionResult struct {
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
}

// OutputConfig represents configuration for output processing
type OutputConfig struct {
	Format string
	Target string
}
