package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/swi/repeater/pkg/patterns"
)

// Pool for reusing buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// Common errors to avoid allocations
var (
	errEmptyCommand = errors.New("command cannot be empty")
)

// OptimizedExecutor is a performance-optimized version of Executor
type OptimizedExecutor struct {
	*Executor // Embed original for compatibility
}

// NewOptimizedExecutor creates a performance-optimized executor
func NewOptimizedExecutor(options ...Option) (*OptimizedExecutor, error) {
	executor, err := NewExecutor(options...)
	if err != nil {
		return nil, err
	}
	return &OptimizedExecutor{Executor: executor}, nil
}

// ExecuteOptimized is an optimized version of Execute with reduced allocations
func (e *OptimizedExecutor) ExecuteOptimized(ctx context.Context, command []string) (*ExecutionResult, error) {
	if len(command) == 0 {
		return nil, errEmptyCommand
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

	// Get buffers from pool to reduce allocations
	stdout := bufferPool.Get().(*bytes.Buffer)
	stderr := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		stdout.Reset()
		stderr.Reset()
		bufferPool.Put(stdout)
		bufferPool.Put(stderr)
	}()

	// Standard execution (optimized for common case)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	duration := time.Since(start)

	// Handle exit code
	originalExitCode := 0
	if err != nil {
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

	// Get strings once to avoid multiple allocations
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Apply pattern matching if configured (avoid string concatenation)
	var finalResult patterns.EvaluationResult
	if e.patternMatcher != nil {
		// Create combined output only if pattern matching is needed
		var combinedOutput string
		if len(stdoutStr) > 0 && len(stderrStr) > 0 {
			// Pre-allocate with known size to avoid reallocations
			combined := make([]byte, 0, len(stdoutStr)+len(stderrStr))
			combined = append(combined, stdoutStr...)
			combined = append(combined, stderrStr...)
			combinedOutput = string(combined)
		} else if len(stdoutStr) > 0 {
			combinedOutput = stdoutStr
		} else {
			combinedOutput = stderrStr
		}
		finalResult = e.patternMatcher.EvaluateResult(combinedOutput, originalExitCode)
	} else {
		// No pattern matching - use original exit code
		finalResult = patterns.EvaluationResult{
			Success:  originalExitCode == 0,
			ExitCode: originalExitCode,
			Reason:   "exit code used",
		}
	}

	// Create result with optimized output handling
	var combinedOutput string
	if len(stdoutStr) > 0 && len(stderrStr) > 0 {
		// Only concatenate if both have content and it's needed
		combinedOutput = stdoutStr + stderrStr
	} else if len(stdoutStr) > 0 {
		combinedOutput = stdoutStr
	} else {
		combinedOutput = stderrStr
	}

	result := &ExecutionResult{
		ExitCode: finalResult.ExitCode,
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		Duration: duration,
		Success:  finalResult.Success,
		Reason:   finalResult.Reason,
		Output:   combinedOutput,
	}

	return result, nil
}
