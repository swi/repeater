package executor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
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
	timeout      time.Duration
	streamWriter io.Writer
	quietMode    bool
	verboseMode  bool
	outputPrefix string
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

// WithStreaming enables real-time output streaming to the provided writer
func WithStreaming(writer io.Writer) Option {
	return func(e *Executor) error {
		e.streamWriter = writer
		return nil
	}
}

// WithQuietMode suppresses all output streaming
func WithQuietMode() Option {
	return func(e *Executor) error {
		e.quietMode = true
		return nil
	}
}

// WithVerboseMode enables verbose output with additional details
func WithVerboseMode() Option {
	return func(e *Executor) error {
		e.verboseMode = true
		return nil
	}
}

// WithOutputPrefix sets a prefix for all streamed output lines
func WithOutputPrefix(prefix string) Option {
	return func(e *Executor) error {
		e.outputPrefix = prefix
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
	var err error

	// Set up streaming if enabled and not in quiet mode
	if e.streamWriter != nil && !e.quietMode {
		// Create pipes for real-time streaming
		stdoutPipe, pipeErr := cmd.StdoutPipe()
		if pipeErr != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", pipeErr)
		}
		stderrPipe, pipeErr := cmd.StderrPipe()
		if pipeErr != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", pipeErr)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command: %w", err)
		}

		// Stream output in real-time while capturing
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			e.streamOutput(stdoutPipe, &stdout, "stdout", command)
		}()

		go func() {
			defer wg.Done()
			e.streamOutput(stderrPipe, &stderr, "stderr", command)
		}()

		// Wait for command to complete
		err = cmd.Wait()

		// Wait for streaming to complete
		wg.Wait()
	} else {
		// Standard execution without streaming
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
	}

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

// streamOutput handles real-time streaming of command output
func (e *Executor) streamOutput(pipe io.ReadCloser, buffer *bytes.Buffer, streamType string, command []string) {
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Write to buffer for result capture
		buffer.WriteString(line + "\n")

		// Stream to writer if enabled
		if e.streamWriter != nil && !e.quietMode {
			output := strings.TrimSpace(line)
			// Add verbose information first
			if e.verboseMode {
				cmdName := command[0]
				output = fmt.Sprintf("[%s:%s] %s", streamType, cmdName, output)
			}

			// Add prefix if specified
			if e.outputPrefix != "" {
				if strings.HasSuffix(e.outputPrefix, " ") {
					output = e.outputPrefix + output
				} else {
					output = e.outputPrefix + " " + output
				}
			}
			fmt.Fprintln(e.streamWriter, output)
		}
	}
}
