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

	"github.com/swi/repeater/pkg/patterns"
)

// ExecutionResult represents the result of a command execution
type ExecutionResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Success  bool   // Whether the command was considered successful (after pattern matching)
	Reason   string // Reason for the success/failure determination
	Output   string // Combined stdout and stderr for convenience
}

// ExecutorConfig holds configuration for the executor
type ExecutorConfig struct {
	Timeout       time.Duration
	Streaming     bool
	StreamWriter  io.Writer
	QuietMode     bool
	VerboseMode   bool
	OutputPrefix  string
	PatternConfig *patterns.PatternConfig
}

// Executor handles command execution with configurable options
type Executor struct {
	timeout        time.Duration
	streamWriter   io.Writer
	quietMode      bool
	verboseMode    bool
	outputPrefix   string
	patternMatcher *patterns.PatternMatcher
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

// NewExecutorWithConfig creates a new executor with the given configuration
func NewExecutorWithConfig(config ExecutorConfig) (*Executor, error) {
	executor := &Executor{
		timeout:      config.Timeout,
		quietMode:    config.QuietMode,
		verboseMode:  config.VerboseMode,
		outputPrefix: config.OutputPrefix,
	}

	// Set default timeout if not specified
	if executor.timeout <= 0 {
		executor.timeout = 30 * time.Second
	}

	// Set up streaming
	if config.Streaming && config.StreamWriter != nil {
		executor.streamWriter = config.StreamWriter
	}

	// Initialize pattern matcher if patterns are configured
	if config.PatternConfig != nil {
		matcher, err := patterns.NewPatternMatcher(*config.PatternConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create pattern matcher: %w", err)
		}
		executor.patternMatcher = matcher
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

	// Get original exit code
	originalExitCode := 0
	if err != nil {
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
				originalExitCode = status.ExitStatus()
			}
		} else {
			// Other errors (command not found, etc.)
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	// Combine stdout and stderr for pattern matching
	combinedOutput := stdout.String() + stderr.String()

	// Apply pattern matching if configured
	var finalResult patterns.EvaluationResult
	if e.patternMatcher != nil {
		finalResult = e.patternMatcher.EvaluateResult(combinedOutput, originalExitCode)
	} else {
		// No pattern matching - use original exit code
		finalResult = patterns.EvaluationResult{
			Success:  originalExitCode == 0,
			ExitCode: originalExitCode,
			Reason:   "exit code used",
		}
	}

	// Create result with all fields
	result := &ExecutionResult{
		ExitCode: finalResult.ExitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
		Success:  finalResult.Success,
		Reason:   finalResult.Reason,
		Output:   combinedOutput,
	}

	return result, nil
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
