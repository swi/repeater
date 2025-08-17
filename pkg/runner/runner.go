package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/swi/repeater/pkg/adaptive"
	"github.com/swi/repeater/pkg/cli"
	"github.com/swi/repeater/pkg/executor"
	"github.com/swi/repeater/pkg/health"
	"github.com/swi/repeater/pkg/httpaware"
	"github.com/swi/repeater/pkg/metrics"
	"github.com/swi/repeater/pkg/ratelimit"
	"github.com/swi/repeater/pkg/scheduler"
	"github.com/swi/repeater/pkg/strategies"
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
	config             *cli.Config
	healthServer       *health.HealthServer
	metricsServer      *metrics.MetricsServer
	httpAwareScheduler httpaware.HTTPAwareScheduler // HTTP-aware scheduler if enabled
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
	// RETRY STRATEGIES - require base-delay or fallback to 1s default
	case "exponential":
		// Will use base-delay or default to 1s if not specified
	case "fibonacci":
		// Will use base-delay or default to 1s if not specified
	case "linear":
		// Will use increment or default to 1s if not specified
	case "polynomial":
		// Will use base-delay and exponent or defaults if not specified
	case "decorrelated-jitter":
		// Will use base-delay and multiplier or defaults if not specified

	// EXECUTION MODES - have specific requirements
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
	case "cron":
		if config.CronExpression == "" {
			return nil, errors.New("cron requires --cron")
		}
	case "adaptive":
		if config.BaseInterval == 0 {
			return nil, errors.New("adaptive requires --base-interval")
		}

	// RATE CONTROL
	case "rate-limit":
		if config.RateSpec == "" {
			return nil, errors.New("rate-limit requires --rate")
		}
	case "load-adaptive":
		if config.BaseInterval == 0 {
			return nil, errors.New("load-adaptive requires --base-interval")
		}

	// LEGACY SUPPORT
	case "backoff":
		if config.InitialInterval == 0 {
			return nil, errors.New("backoff requires --initial-delay")
		}
	default:
		return nil, errors.New("unknown subcommand: " + config.Subcommand)
	}

	// Initialize health server if enabled
	var healthServer *health.HealthServer
	if config.HealthEnabled {
		healthServer = health.NewHealthServer(config.HealthPort)
	}

	// Initialize metrics server if enabled
	var metricsServer *metrics.MetricsServer
	if config.MetricsEnabled {
		metricsServer = metrics.NewMetricsServer(config.MetricsPort)
	}

	return &Runner{
		config:        config,
		healthServer:  healthServer,
		metricsServer: metricsServer,
	}, nil
}

// Run executes the configured command according to the scheduling rules
func (r *Runner) Run(ctx context.Context) (*ExecutionStats, error) {
	startTime := time.Now()

	// Create executor with configuration including pattern matching
	executorConfig := executor.ExecutorConfig{
		Timeout:       r.config.Timeout,
		Streaming:     r.config.Stream,
		StreamWriter:  os.Stdout,
		QuietMode:     r.config.Quiet || r.config.StatsOnly,
		VerboseMode:   r.config.Verbose,
		OutputPrefix:  r.config.OutputPrefix,
		PatternConfig: r.config.GetPatternConfig(),
	}

	// Set default timeout if not specified
	if executorConfig.Timeout <= 0 {
		executorConfig.Timeout = 30 * time.Second
	}

	exec, err := executor.NewExecutorWithConfig(executorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create scheduler based on subcommand
	sched, err := r.createScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	defer sched.Stop()

	// Start health server if enabled
	if r.healthServer != nil {
		go func() {
			if err := r.healthServer.Start(ctx); err != nil {
				// Log error but don't fail execution
				if r.config.Verbose {
					fmt.Fprintf(os.Stderr, "Health server error: %v\n", err)
				}
			}
		}()
		r.healthServer.SetReady(true)
	}

	// Start metrics server if enabled
	if r.metricsServer != nil {
		go func() {
			if err := r.metricsServer.Start(ctx); err != nil {
				// Log error but don't fail execution
				if r.config.Verbose {
					fmt.Fprintf(os.Stderr, "Metrics server error: %v\n", err)
				}
			}
		}()
	}

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
				// Command executed - use pattern matching result if available
				record.ExitCode = result.ExitCode
				record.Stdout = result.Stdout
				record.Stderr = result.Stderr

				// Use the Success field from ExecutionResult which includes pattern matching
				if result.Success {
					stats.SuccessfulExecutions++
				} else {
					stats.FailedExecutions++
				}
			}

			stats.Executions = append(stats.Executions, record)
			stats.TotalExecutions++
			executionNumber++

			// Update health server stats if enabled
			if r.healthServer != nil {
				r.healthServer.SetExecutionStats(health.ExecutionStats{
					TotalExecutions:      int64(stats.TotalExecutions),
					SuccessfulExecutions: int64(stats.SuccessfulExecutions),
					FailedExecutions:     int64(stats.FailedExecutions),
					AverageResponseTime:  time.Duration(0), // TODO: Calculate if needed
					LastExecution:        time.Now(),
				})
			}

			// Update metrics server if enabled
			if r.metricsServer != nil {
				success := (execErr == nil && result != nil && result.Success)
				r.metricsServer.RecordExecution(success, record.Duration)
			}

			// Update adaptive scheduler if applicable
			if adaptiveWrapper, ok := sched.(*AdaptiveSchedulerWrapper); ok {
				// Determine success based on execution result
				success := (execErr == nil && result != nil && result.Success)
				adaptiveWrapper.UpdateFromExecution(record, success)

				// Record scheduler interval in metrics if enabled
				if r.metricsServer != nil {
					metrics := adaptiveWrapper.GetMetrics()
					r.metricsServer.RecordSchedulerInterval(metrics.CurrentInterval)
				}

				// Show metrics if requested
				if r.config.ShowMetrics {
					r.showAdaptiveMetrics(adaptiveWrapper.GetMetrics())
				}
			}

			// Update HTTP-aware scheduler if enabled
			if r.httpAwareScheduler != nil && result != nil {
				// Pass the full command output (stdout + stderr) to HTTP-aware scheduler
				fullOutput := result.Stdout
				if result.Stderr != "" {
					if fullOutput != "" {
						fullOutput += "\n"
					}
					fullOutput += result.Stderr
				}

				// Set the last response for HTTP-aware analysis
				r.httpAwareScheduler.SetLastResponse(fullOutput)

				// Show HTTP timing info if verbose mode is enabled
				if r.config.Verbose {
					if timingInfo := r.httpAwareScheduler.GetTimingInfo(); timingInfo != nil {
						fmt.Fprintf(os.Stderr, "HTTP-aware: Found %s timing: %v\n",
							timingInfo.Source, timingInfo.Delay)
					}
				}
			}

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

	var baseScheduler Scheduler
	var err error

	switch r.config.Subcommand {
	// NEW RETRY STRATEGIES
	case "exponential":
		baseScheduler, err = r.createStrategyScheduler("exponential")
	case "fibonacci":
		baseScheduler, err = r.createStrategyScheduler("fibonacci")
	case "linear":
		baseScheduler, err = r.createStrategyScheduler("linear")
	case "polynomial":
		baseScheduler, err = r.createStrategyScheduler("polynomial")
	case "decorrelated-jitter":
		baseScheduler, err = r.createStrategyScheduler("decorrelated-jitter")

	// EXISTING EXECUTION MODES
	case "interval":
		baseScheduler, err = scheduler.NewIntervalScheduler(r.config.Every, noJitter, immediateStart)
	case "count", "duration":
		interval := r.config.Every
		if interval == 0 {
			interval = immediateInterval // Immediate execution for count/duration without --every
		}
		baseScheduler, err = scheduler.NewIntervalScheduler(interval, noJitter, immediateStart)
	case "cron":
		baseScheduler, err = r.createCronScheduler()
	case "adaptive":
		baseScheduler, err = r.createAdaptiveScheduler()

	// EXISTING RATE CONTROL
	case "rate-limit":
		baseScheduler, err = r.createRateLimitScheduler()
	case "load-adaptive":
		baseScheduler, err = r.createLoadAdaptiveScheduler()

	// LEGACY SUPPORT
	case "backoff":
		baseScheduler, err = r.createBackoffScheduler()
	default:
		return nil, fmt.Errorf("unknown subcommand: %s", r.config.Subcommand)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with HTTP-aware scheduler if enabled
	return r.wrapWithHTTPAware(baseScheduler)
}

// wrapWithHTTPAware wraps a scheduler with HTTP-aware functionality if enabled
func (r *Runner) wrapWithHTTPAware(baseScheduler Scheduler) (Scheduler, error) {
	httpConfig := r.config.GetHTTPAwareConfig()
	if httpConfig == nil {
		return baseScheduler, nil
	}

	// Create HTTP-aware scheduler with the base scheduler as fallback
	r.httpAwareScheduler = httpaware.NewHTTPAwareScheduler(*httpConfig)

	// For now, we'll use the base scheduler for timing and integrate HTTP-aware logic
	// in the execution loop. This allows us to maintain compatibility with existing
	// scheduler interfaces while adding HTTP-aware intelligence.
	return baseScheduler, nil
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

// AdaptiveSchedulerWrapper wraps adaptive.AdaptiveScheduler to implement Scheduler interface
type AdaptiveSchedulerWrapper struct {
	scheduler *adaptive.AdaptiveScheduler
	config    *cli.Config
	nextChan  chan time.Time
	stopChan  chan struct{}
	stopped   bool
}

// NewAdaptiveSchedulerWrapper creates a new adaptive scheduler wrapper
func NewAdaptiveSchedulerWrapper(scheduler *adaptive.AdaptiveScheduler, config *cli.Config) *AdaptiveSchedulerWrapper {
	w := &AdaptiveSchedulerWrapper{
		scheduler: scheduler,
		config:    config,
		nextChan:  make(chan time.Time, 1),
		stopChan:  make(chan struct{}),
		stopped:   false,
	}

	// Start the scheduling goroutine
	go w.scheduleLoop()

	return w
}

// Next returns a channel that delivers the next execution time
func (w *AdaptiveSchedulerWrapper) Next() <-chan time.Time {
	return w.nextChan
}

// scheduleLoop continuously schedules the next execution based on adaptive intervals
func (w *AdaptiveSchedulerWrapper) scheduleLoop() {
	for {
		select {
		case <-w.stopChan:
			return
		default:
			// Get current interval from adaptive scheduler
			interval := w.scheduler.GetCurrentInterval()

			// Wait for the interval
			select {
			case <-time.After(interval):
				// Send next execution time
				select {
				case w.nextChan <- time.Now():
					// Successfully sent, continue to next iteration
				case <-w.stopChan:
					return
				}
			case <-w.stopChan:
				return
			}
		}
	}
}

// Stop stops the scheduler
func (w *AdaptiveSchedulerWrapper) Stop() {
	if !w.stopped {
		close(w.stopChan)
		w.stopped = true
	}
}

// UpdateFromExecution updates the adaptive scheduler with execution results
func (w *AdaptiveSchedulerWrapper) UpdateFromExecution(record ExecutionRecord, success bool) {
	result := adaptive.ExecutionResult{
		Timestamp:    record.StartTime,
		ResponseTime: record.Duration,
		Success:      success, // Use pattern matching result
		StatusCode:   record.ExitCode,
		Error:        nil,
	}

	if !success {
		result.Error = fmt.Errorf("command failed (exit code %d)", record.ExitCode)
	}

	w.scheduler.UpdateFromResult(result)
}

// GetMetrics returns current adaptive metrics
func (w *AdaptiveSchedulerWrapper) GetMetrics() *adaptive.AdaptiveMetrics {
	return w.scheduler.GetMetrics()
}

// showAdaptiveMetrics displays current adaptive scheduling metrics
func (r *Runner) showAdaptiveMetrics(metrics *adaptive.AdaptiveMetrics) {
	fmt.Printf("📊 Adaptive Metrics: Interval=%v, Success=%.1f%%, Circuit=%s\n",
		metrics.CurrentInterval.Round(time.Millisecond),
		metrics.SuccessRate*100,
		r.circuitStateString(metrics.CircuitState))
}

// circuitStateString converts circuit state to readable string
func (r *Runner) circuitStateString(state adaptive.CircuitState) string {
	switch state {
	case adaptive.CircuitClosed:
		return "CLOSED"
	case adaptive.CircuitOpen:
		return "OPEN"
	case adaptive.CircuitHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// createAdaptiveScheduler creates an adaptive scheduler
func (r *Runner) createAdaptiveScheduler() (Scheduler, error) {
	// Create adaptive configuration from CLI config
	adaptiveConfig := &adaptive.AdaptiveConfig{
		BaseInterval:           r.config.BaseInterval,
		MinInterval:            r.config.MinInterval,
		MaxInterval:            r.config.MaxInterval,
		AdditiveIncrease:       200 * time.Millisecond, // Default
		MultiplicativeDecrease: 0.6,                    // Default
		ResponseTimeAlpha:      0.1,                    // Default
		SlowThresholdFactor:    r.config.SlowThreshold,
		FastThresholdFactor:    r.config.FastThreshold,
		PriorAlpha:             1.0,  // Default
		PriorBeta:              1.0,  // Default
		DecayRate:              0.95, // Default
		FailureThreshold:       r.config.FailureThreshold,
		RecoveryThreshold:      0.8, // Default
		WindowSize:             100, // Default
		MinSamples:             10,  // Default
	}

	// Create adaptive scheduler
	scheduler, err := adaptive.NewAdaptiveSchedulerWithValidation(adaptiveConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create adaptive scheduler: %w", err)
	}

	// Wrap it to implement Scheduler interface
	return NewAdaptiveSchedulerWrapper(scheduler, r.config), nil
}

// createBackoffScheduler creates an exponential backoff scheduler
func (r *Runner) createBackoffScheduler() (Scheduler, error) {
	return scheduler.NewExponentialBackoffScheduler(
		r.config.InitialInterval,
		r.config.BackoffMultiplier,
		r.config.BackoffMax,
		r.config.BackoffJitter,
	), nil
}

// createLoadAdaptiveScheduler creates a load-aware adaptive scheduler
func (r *Runner) createLoadAdaptiveScheduler() (Scheduler, error) {
	return scheduler.NewLoadAwareSchedulerWithBounds(
		r.config.BaseInterval,
		r.config.TargetCPU,
		r.config.TargetMemory,
		r.config.TargetLoad,
		r.config.MinInterval,
		r.config.MaxInterval,
	), nil
}

// createCronScheduler creates a cron-based scheduler
func (r *Runner) createCronScheduler() (Scheduler, error) {
	if r.config.CronExpression == "" {
		return nil, fmt.Errorf("cron expression is required for cron subcommand")
	}

	// Create cron scheduler with timezone
	cronScheduler, err := scheduler.NewCronScheduler(r.config.CronExpression, r.config.Timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron scheduler: %w", err)
	}

	return cronScheduler, nil
}

// createStrategyScheduler creates a strategy-based scheduler using mathematical retry algorithms
func (r *Runner) createStrategyScheduler(strategyName string) (Scheduler, error) {
	// Create strategy configuration from CLI config
	config := &strategies.StrategyConfig{
		MaxAttempts:     r.config.MaxRetries,
		Timeout:         r.config.Timeout,
		SuccessPattern:  r.config.SuccessPattern,
		FailurePattern:  r.config.FailurePattern,
		CaseInsensitive: r.config.CaseInsensitive,
	}

	// Set default values if not specified
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3 // Default retry attempts
	}

	// Create the specific strategy based on the strategy name
	var strategy strategies.Strategy

	switch strategyName {
	case "exponential":
		baseDelay := r.getBaseDelay()
		multiplier := r.getMultiplier()
		maxDelay := r.getMaxDelay()
		config.BaseDelay = baseDelay
		config.Multiplier = multiplier
		config.MaxDelay = maxDelay
		strategy = strategies.NewExponentialStrategy(baseDelay, multiplier, maxDelay)

	case "fibonacci":
		baseDelay := r.getBaseDelay()
		maxDelay := r.getMaxDelay()
		config.BaseDelay = baseDelay
		config.MaxDelay = maxDelay
		strategy = strategies.NewFibonacciStrategy(baseDelay, maxDelay)

	case "linear":
		increment := r.getIncrement()
		maxDelay := r.getMaxDelay()
		config.Increment = increment
		config.MaxDelay = maxDelay
		strategy = strategies.NewLinearStrategy(increment, maxDelay)

	case "polynomial":
		baseDelay := r.getBaseDelay()
		exponent := r.getExponent()
		maxDelay := r.getMaxDelay()
		config.BaseDelay = baseDelay
		config.Exponent = exponent
		config.MaxDelay = maxDelay
		strategy = strategies.NewPolynomialStrategy(baseDelay, exponent, maxDelay)

	case "decorrelated-jitter":
		baseDelay := r.getBaseDelay()
		multiplier := r.getMultiplier()
		maxDelay := r.getMaxDelay()
		config.BaseDelay = baseDelay
		config.Multiplier = multiplier
		config.MaxDelay = maxDelay
		strategy = strategies.NewDecorrelatedJitterStrategy(baseDelay, multiplier, maxDelay)

	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyName)
	}

	// Create and return the strategy scheduler
	return scheduler.NewStrategyScheduler(strategy, config)
}

// Helper functions to get strategy parameters with defaults
func (r *Runner) getBaseDelay() time.Duration {
	if r.config.BaseDelay > 0 {
		return r.config.BaseDelay
	}
	// Fallback to legacy field for backward compatibility
	if r.config.InitialInterval > 0 {
		return r.config.InitialInterval
	}
	return 1 * time.Second // Default
}

func (r *Runner) getIncrement() time.Duration {
	if r.config.Increment > 0 {
		return r.config.Increment
	}
	return 1 * time.Second // Default
}

func (r *Runner) getMultiplier() float64 {
	if r.config.Multiplier > 0 {
		return r.config.Multiplier
	}
	// Fallback to legacy field for backward compatibility
	if r.config.BackoffMultiplier > 0 {
		return r.config.BackoffMultiplier
	}
	return 2.0 // Default
}

func (r *Runner) getExponent() float64 {
	if r.config.Exponent > 0 {
		return r.config.Exponent
	}
	return 2.0 // Default
}

func (r *Runner) getMaxDelay() time.Duration {
	if r.config.MaxDelay > 0 {
		return r.config.MaxDelay
	}
	// Fallback to legacy field for backward compatibility
	if r.config.BackoffMax > 0 {
		return r.config.BackoffMax
	}
	return 60 * time.Second // Default
}
