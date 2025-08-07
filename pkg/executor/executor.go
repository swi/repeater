package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// ExecutionResult represents the result of a command execution
type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Executor handles command execution with configurable options
type Executor struct {
	timeout time.Duration
}

// Option represents a configuration option for the executor
type Option func(*Executor) error

// WithTimeout sets the execution timeout for commands
func WithTimeout(timeout time.Duration) Option {
	return func(e *Executor) error {
		if timeout <= 0 {
			return errors.New("timeout must be positive")
		}
		e.timeout = timeout
		return nil
	}
}

// NewExecutor creates a new command executor with the given options
func NewExecutor(options ...Option) (*Executor, error) {
	executor := &Executor{
		timeout: 30 * time.Second, // Default timeout
	}

	for _, option := range options {
		if err := option(executor); err != nil {
			return nil, err
		}
	}

	return executor, nil
}

// Execute runs a command and returns the execution result
func (e *Executor) Execute(ctx context.Context, command []string) (*ExecutionResult, error) {
	if len(command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	start := time.Now()

	// Create context with timeout if specified
	execCtx := ctx
	if e.timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, e.timeout)
		defer cancel()
	}

	// Create the command
	cmd := exec.CommandContext(execCtx, command[0], command[1:]...)

	// Prepare output buffers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()
	duration := time.Since(start)

	// Create result with common fields
	result := &ExecutionResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Handle different types of errors
	if err != nil {
		return e.handleExecutionError(err, execCtx, result)
	}

	// Successful execution
	result.ExitCode = 0
	return result, nil
}

// handleExecutionError processes different types of execution errors
func (e *Executor) handleExecutionError(err error, execCtx context.Context, result *ExecutionResult) (*ExecutionResult, error) {
	// Check if it's a context cancellation
	if execCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("command timeout after %v", e.timeout)
	}
	if execCtx.Err() == context.Canceled {
		return nil, fmt.Errorf("command execution canceled: %w", context.Canceled)
	}

	// Check if it's an exit error (non-zero exit code)
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			result.ExitCode = status.ExitStatus()
			return result, nil
		}
	}

	// Other errors (command not found, etc.)
	return nil, fmt.Errorf("command execution failed: %w", err)
}
