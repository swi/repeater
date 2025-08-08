package runner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/swi/repeater/pkg/cli"
	"github.com/swi/repeater/pkg/executor"
	"github.com/swi/repeater/pkg/ratelimit"
	"github.com/swi/repeater/pkg/scheduler"
)

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

// Runner orchestrates the execution of commands using schedulers and executors
type Runner struct {
	config *cli.Config
}

// NewRunner creates a new runner with the given configuration
func NewRunner(config *cli.Config) (*Runner, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	if len(config.Command) == 0 {
		return nil, errors.New("command cannot be empty")
	}

	// Validate subcommand-specific requirements
	switch config.Subcommand {
	case "interval":
		if config.Every == 0 {
			return nil, errors.New("interval requires --every")
		}
	case "count":
		if config.Times == 0 {
			return nil, errors.New("count requires --times")
		}
	case "duration":
		if config.For == 0 {
			return nil, errors.New("duration requires --for")
		}
	case "rate-limit":
		if config.RateSpec == "" {
			return nil, errors.New("rate-limit requires --rate")
		}
	default:
		return nil, errors.New("unknown subcommand: " + config.Subcommand)
	}

	return &Runner{
		config: config,
	}, nil
}

// Run executes the configured command according to the scheduling rules
func (r *Runner) Run(ctx context.Context) (*ExecutionStats, error) {
	startTime := time.Now()

	// Create executor
	exec, err := executor.NewExecutor()
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create scheduler based on subcommand
	sched, err := r.createScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	defer sched.Stop()

	// Create execution context with stop conditions
	execCtx, cancel := r.createExecutionContext(ctx)
	defer cancel()

	// Initialize statistics
	stats := &ExecutionStats{
		StartTime:  startTime,
		Executions: make([]ExecutionRecord, 0),
	}

	// Main execution loop
	executionNumber := 1
	for {
		select {
		case <-execCtx.Done():
			// Context canceled (timeout, signal, or stop condition)
			stats.EndTime = time.Now()
			stats.Duration = stats.EndTime.Sub(stats.StartTime)

			if execCtx.Err() == context.Canceled {
				return stats, fmt.Errorf("execution stopped: %w", context.Canceled)
			}
			return stats, nil

		case tick := <-sched.Next():
			// Check stop conditions before execution
			if r.shouldStop(stats, startTime) {
				stats.EndTime = time.Now()
				stats.Duration = stats.EndTime.Sub(stats.StartTime)
				return stats, nil
			}

			// Execute command
			execStart := time.Now()
			result, execErr := exec.Execute(execCtx, r.config.Command)
			execEnd := time.Now()

			// Record execution
			record := ExecutionRecord{
				ExecutionNumber: executionNumber,
				StartTime:       execStart,
				EndTime:         execEnd,
				Duration:        execEnd.Sub(execStart),
			}

			if execErr != nil {
				// Command failed or was canceled
				if execCtx.Err() != nil {
					// Context was canceled during execution
					stats.EndTime = time.Now()
					stats.Duration = stats.EndTime.Sub(stats.StartTime)
					return stats, fmt.Errorf("execution canceled: %w", execCtx.Err())
				}

				// Command failed but we continue
				record.ExitCode = 1 // Default failure code
				record.Stderr = execErr.Error()
				stats.FailedExecutions++
			} else {
				// Command succeeded
				record.ExitCode = result.ExitCode
				record.Stdout = result.Stdout
				record.Stderr = result.Stderr

				if result.ExitCode == 0 {
					stats.SuccessfulExecutions++
				} else {
					stats.FailedExecutions++
				}
			}

			stats.Executions = append(stats.Executions, record)
			stats.TotalExecutions++
			executionNumber++

			// Update tick time for scheduler
			_ = tick
		}
	}
}

// Scheduler interface for type safety
type Scheduler interface {
	Next() <-chan time.Time
	Stop()
}

// RateLimitScheduler implements Scheduler using Diophantine rate limiting
type RateLimitScheduler struct {
	limiter  *ratelimit.DiophantineRateLimiter
	showNext bool
	nextChan chan time.Time
	stopChan chan struct{}
	stopped  bool
	started  bool
}

// NewRateLimitScheduler creates a new rate-limit aware scheduler
func NewRateLimitScheduler(limiter *ratelimit.DiophantineRateLimiter, showNext bool) *RateLimitScheduler {
	s := &RateLimitScheduler{
		limiter:  limiter,
		showNext: showNext,
		nextChan: make(chan time.Time, 1),
		stopChan: make(chan struct{}),
		stopped:  false,
		started:  false,
	}

	// Start the scheduling goroutine immediately
	go s.scheduleLoop()

	return s
}

// Next returns a channel that delivers the next allowed execution time
func (s *RateLimitScheduler) Next() <-chan time.Time {
	return s.nextChan
}

// scheduleLoop continuously schedules the next allowed execution
func (s *RateLimitScheduler) scheduleLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			if s.limiter.Allow() {
				// Request is allowed now
				select {
				case s.nextChan <- time.Now():
					// Successfully sent, continue to next iteration
				case <-s.stopChan:
					return
				}
			} else {
				// Request not allowed, wait until next allowed time
				nextTime := s.limiter.NextAllowedTime()
				if s.showNext {
					fmt.Printf("Next request allowed at: %s\n", nextTime.Format("15:04:05"))
				}

				// Wait until that time
				waitDuration := time.Until(nextTime)
				if waitDuration > 0 {
					select {
					case <-time.After(waitDuration):
						// Continue loop to try again
					case <-s.stopChan:
						return
					}
				}
			}
		}
	}
}

// Stop stops the scheduler
func (s *RateLimitScheduler) Stop() {
	if !s.stopped {
		close(s.stopChan)
		s.stopped = true
	}
}

// createScheduler creates the appropriate scheduler based on the subcommand
func (r *Runner) createScheduler() (Scheduler, error) {
	const immediateInterval = 1 * time.Millisecond
	const noJitter = 0.0
	const immediateStart = true

	switch r.config.Subcommand {
	case "interval":
		return scheduler.NewIntervalScheduler(r.config.Every, noJitter, immediateStart)
	case "count", "duration":
		interval := r.config.Every
		if interval == 0 {
			interval = immediateInterval // Immediate execution for count/duration without --every
		}
		return scheduler.NewIntervalScheduler(interval, noJitter, immediateStart)
	case "rate-limit":
		return r.createRateLimitScheduler()
	default:
		return nil, fmt.Errorf("unknown subcommand: %s", r.config.Subcommand)
	}
}

// createRateLimitScheduler creates a rate-limit aware scheduler
func (r *Runner) createRateLimitScheduler() (Scheduler, error) {
	// Parse rate specification
	rate, period, err := ratelimit.ParseRateSpec(r.config.RateSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid rate spec: %w", err)
	}

	// Parse retry pattern if provided
	var retryPattern []time.Duration
	if r.config.RetryPattern != "" {
		retryPattern, err = r.parseRetryPattern(r.config.RetryPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid retry pattern: %w", err)
		}
	} else {
		retryPattern = []time.Duration{0} // Default: single attempt, no retries
	}

	// Create Diophantine rate limiter
	limiter := ratelimit.NewDiophantineRateLimiter(rate, period, retryPattern)

	// Create a scheduler that respects the rate limiter
	return NewRateLimitScheduler(limiter, r.config.ShowNext), nil
}

// parseRetryPattern parses retry pattern string like "0,10m,30m"
func (r *Runner) parseRetryPattern(pattern string) ([]time.Duration, error) {
	parts := strings.Split(pattern, ",")
	retryPattern := make([]time.Duration, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "0" {
			retryPattern[i] = 0
		} else {
			duration, err := time.ParseDuration(part)
			if err != nil {
				return nil, fmt.Errorf("invalid retry offset '%s': %w", part, err)
			}
			retryPattern[i] = duration
		}
	}

	return retryPattern, nil
}

// createExecutionContext creates a context with appropriate timeouts
func (r *Runner) createExecutionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if r.config.For > 0 {
		// Duration-based timeout
		return context.WithTimeout(ctx, r.config.For)
	}

	// No timeout, use parent context
	return context.WithCancel(ctx)
}

// shouldStop checks if execution should stop based on configured limits
func (r *Runner) shouldStop(stats *ExecutionStats, startTime time.Time) bool {
	// Check times limit
	if r.config.Times > 0 && int64(stats.TotalExecutions) >= r.config.Times {
		return true
	}

	// Check duration limit
	if r.config.For > 0 && time.Since(startTime) >= r.config.For {
		return true
	}

	return false
}
