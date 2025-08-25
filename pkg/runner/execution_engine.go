package runner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/swi/repeater/pkg/adaptive"
	"github.com/swi/repeater/pkg/cli"
	"github.com/swi/repeater/pkg/executor"
	"github.com/swi/repeater/pkg/httpaware"
	"github.com/swi/repeater/pkg/interfaces"
	"github.com/swi/repeater/pkg/patterns"
)

// ExecutionEngine handles the core execution loop and statistics
type ExecutionEngine struct {
	config             *cli.Config
	httpAwareScheduler httpaware.HTTPAwareScheduler
	executor           *executor.Executor
	patternMatcher     *patterns.PatternMatcher
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(config *cli.Config, httpAwareScheduler httpaware.HTTPAwareScheduler) (*ExecutionEngine, error) {
	exec, err := executor.NewExecutor()
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create pattern matcher if patterns are configured
	var patternMatcher *patterns.PatternMatcher
	if config.SuccessPattern != "" || config.FailurePattern != "" {
		matcher, err := patterns.NewPatternMatcher(patterns.PatternConfig{
			SuccessPattern:  config.SuccessPattern,
			FailurePattern:  config.FailurePattern,
			CaseInsensitive: config.CaseInsensitive,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create pattern matcher: %w", err)
		}
		patternMatcher = matcher
	}

	return &ExecutionEngine{
		config:             config,
		httpAwareScheduler: httpAwareScheduler,
		executor:           exec,
		patternMatcher:     patternMatcher,
	}, nil
}

// ExecuteWithScheduler runs the execution loop with the given scheduler
func (e *ExecutionEngine) ExecuteWithScheduler(ctx context.Context, scheduler interfaces.Scheduler) (*ExecutionStats, error) {
	stats := &ExecutionStats{
		StartTime:  time.Now(),
		Executions: make([]ExecutionRecord, 0),
	}

	// Create execution context with timeout if specified
	execCtx, cancel := e.createExecutionContext(ctx)
	defer cancel()

	for {
		select {
		case <-execCtx.Done():
			stats.EndTime = time.Now()
			stats.Duration = stats.EndTime.Sub(stats.StartTime)
			return stats, execCtx.Err()

		case nextTime := <-scheduler.Next():
			if nextTime.IsZero() {
				// Scheduler indicates completion
				stats.EndTime = time.Now()
				stats.Duration = stats.EndTime.Sub(stats.StartTime)
				return stats, nil
			}

			// Execute the command
			result, err := e.executor.Execute(execCtx, e.config.Command)
			if err != nil {
				// Log execution error but continue
				if !e.config.Quiet {
					fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
				}
			}

			// Update statistics
			e.updateStats(stats, result)

			// Process result for HTTP-aware scheduling
			if e.httpAwareScheduler != nil {
				// Pass execution result to HTTP-aware scheduler for timing adjustments
				if result.Stdout != "" {
					e.httpAwareScheduler.SetLastResponse(result.Stdout)
				} else if result.Stderr != "" {
					// Fallback to stderr for HTTP error responses
					e.httpAwareScheduler.SetLastResponse(result.Stderr)
				}
			}

			// Check for pattern matching
			success := e.checkPatternMatch(result)

			// Update adaptive scheduler if available
			if adaptiveWrapper, ok := scheduler.(*AdaptiveSchedulerWrapper); ok {
				// Update the adaptive scheduler with execution results
				var execError error
				if !success {
					execError = fmt.Errorf("command failed with exit code %d", result.ExitCode)
				}

				adaptiveResult := adaptive.ExecutionResult{
					Timestamp:    time.Now(),
					ResponseTime: result.Duration,
					Success:      success,
					StatusCode:   result.ExitCode,
					Error:        execError,
				}
				adaptiveWrapper.scheduler.UpdateFromResult(adaptiveResult)
			}

			// Show progress if verbose
			if e.config.Verbose {
				e.showExecutionProgress(stats, result)
			}

			// Check stop conditions
			if e.shouldStop(stats, stats.StartTime) {
				stats.EndTime = time.Now()
				stats.Duration = stats.EndTime.Sub(stats.StartTime)
				scheduler.Stop()
				return stats, nil
			}
		}
	}
}

// createExecutionContext creates a context with timeout if specified
func (e *ExecutionEngine) createExecutionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if e.config.For > 0 {
		return context.WithTimeout(ctx, e.config.For)
	}
	return ctx, func() {} // No-op cancel function
}

// updateStats updates execution statistics with the result
func (e *ExecutionEngine) updateStats(stats *ExecutionStats, result *executor.ExecutionResult) {
	stats.TotalExecutions++

	// Record execution details
	startTime := time.Now().Add(-result.Duration)
	record := ExecutionRecord{
		ExecutionNumber: stats.TotalExecutions,
		ExitCode:        result.ExitCode,
		Duration:        result.Duration,
		Stdout:          result.Stdout,
		Stderr:          result.Stderr,
		StartTime:       startTime,
		EndTime:         startTime.Add(result.Duration),
	}
	stats.Executions = append(stats.Executions, record)

	// Update success/failure counts
	if result.ExitCode == 0 {
		stats.SuccessfulExecutions++
	} else {
		stats.FailedExecutions++
	}
}

// checkPatternMatch checks if the result matches success/failure patterns
func (e *ExecutionEngine) checkPatternMatch(result *executor.ExecutionResult) bool {
	// If pattern matcher is configured, use it for evaluation
	if e.patternMatcher != nil {
		// Combine stdout and stderr for pattern matching
		output := result.Stdout
		if result.Stderr != "" {
			if output != "" {
				output += "\n" + result.Stderr
			} else {
				output = result.Stderr
			}
		}

		evalResult := e.patternMatcher.EvaluateResult(output, result.ExitCode)
		return evalResult.Success
	}

	// Fall back to exit code as success indicator
	return result.ExitCode == 0
}

// showExecutionProgress shows progress information if verbose mode is enabled
func (e *ExecutionEngine) showExecutionProgress(stats *ExecutionStats, result *executor.ExecutionResult) {
	fmt.Printf("Execution %d: exit code %d, duration %v\n",
		stats.TotalExecutions, result.ExitCode, result.Duration)

	if result.ExitCode == 0 {
		fmt.Printf("✓ Success\n")
	} else {
		fmt.Printf("✗ Failed\n")
	}
}

// shouldStop determines if execution should stop based on configuration
func (e *ExecutionEngine) shouldStop(stats *ExecutionStats, startTime time.Time) bool {
	// Check count-based stopping condition
	if e.config.Times > 0 && int64(stats.TotalExecutions) >= e.config.Times {
		return true
	}

	// Check time-based stopping condition (handled by context timeout)
	// Duration-based stopping is managed by the execution context timeout

	return false
}

// ShowFinalStats displays final execution statistics
func (e *ExecutionEngine) ShowFinalStats(stats *ExecutionStats) {
	if e.config.Quiet {
		return
	}

	fmt.Printf("\n=== Execution Summary ===\n")
	fmt.Printf("Total executions: %d\n", stats.TotalExecutions)
	fmt.Printf("Successful: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("Failed: %d\n", stats.FailedExecutions)
	fmt.Printf("Duration: %v\n", stats.Duration)

	if stats.TotalExecutions > 0 {
		successRate := float64(stats.SuccessfulExecutions) / float64(stats.TotalExecutions) * 100
		fmt.Printf("Success rate: %.1f%%\n", successRate)
	}
}

// ShowAdaptiveMetrics shows adaptive scheduler metrics if available
func (e *ExecutionEngine) ShowAdaptiveMetrics(metrics *adaptive.AdaptiveMetrics) {
	if e.config.Quiet {
		return
	}

	fmt.Printf("\n=== Adaptive Metrics ===\n")
	fmt.Printf("Current interval: %v\n", metrics.CurrentInterval)
	fmt.Printf("Success rate: %.2f\n", metrics.SuccessRate)
	fmt.Printf("Circuit state: %s\n", e.circuitStateString(metrics.CircuitState))
	fmt.Printf("Total executions: %d\n", metrics.TotalExecutions)
	fmt.Printf("Average response time: %v\n", metrics.AverageResponseTime)
}

func (e *ExecutionEngine) circuitStateString(state adaptive.CircuitState) string {
	switch state {
	case adaptive.CircuitClosed:
		return "closed"
	case adaptive.CircuitOpen:
		return "open"
	case adaptive.CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
