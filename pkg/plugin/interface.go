package plugin

import (
	"context"
	"time"

	"github.com/swi/repeater/pkg/interfaces"
)

// Use centralized Scheduler interface from pkg/interfaces
type Scheduler = interfaces.Scheduler

// SchedulerPlugin interface defines the contract for scheduler plugins
type SchedulerPlugin interface {
	// Plugin metadata
	Name() string
	Version() string
	Description() string

	// Configuration management
	ValidateConfig(config map[string]interface{}) error
	ConfigSchema() *ConfigSchema

	// Scheduler creation
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

// ConfigSchema represents the configuration schema for a plugin
type ConfigSchema struct {
	Fields []ConfigField `json:"fields"`
}

// ConfigField represents a single configuration field
type ConfigField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "int", "duration", "bool"
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Validation  *Validation `json:"validation,omitempty"`
}

// Validation represents validation rules for a configuration field
type Validation struct {
	Min   *float64 `json:"min,omitempty"`
	Max   *float64 `json:"max,omitempty"`
	Regex *string  `json:"regex,omitempty"`
	OneOf []string `json:"one_of,omitempty"`
}
