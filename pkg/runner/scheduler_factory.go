package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/swi/repeater/pkg/adaptive"
	"github.com/swi/repeater/pkg/cli"
	"github.com/swi/repeater/pkg/httpaware"
	"github.com/swi/repeater/pkg/interfaces"
	"github.com/swi/repeater/pkg/ratelimit"
	"github.com/swi/repeater/pkg/scheduler"
	"github.com/swi/repeater/pkg/strategies"
)

// SchedulerFactory handles creation of all scheduler types
type SchedulerFactory struct {
	config             *cli.Config
	httpAwareScheduler httpaware.HTTPAwareScheduler
}

// NewSchedulerFactory creates a new scheduler factory
func NewSchedulerFactory(config *cli.Config) *SchedulerFactory {
	return &SchedulerFactory{
		config: config,
	}
}

// CreateScheduler creates the appropriate scheduler based on configuration
func (f *SchedulerFactory) CreateScheduler() (interfaces.Scheduler, error) {
	const immediateInterval = 1 * time.Millisecond
	const noJitter = 0.0
	const immediateStart = true

	var baseScheduler interfaces.Scheduler
	var err error

	switch f.config.Subcommand {
	// NEW RETRY STRATEGIES
	case "exponential":
		baseScheduler, err = f.createStrategyScheduler("exponential")
	case "fibonacci":
		baseScheduler, err = f.createStrategyScheduler("fibonacci")
	case "linear":
		baseScheduler, err = f.createStrategyScheduler("linear")
	case "polynomial":
		baseScheduler, err = f.createStrategyScheduler("polynomial")
	case "decorrelated-jitter":
		baseScheduler, err = f.createStrategyScheduler("decorrelated-jitter")

	// EXISTING EXECUTION MODES
	case "interval":
		baseScheduler, err = scheduler.NewIntervalScheduler(f.config.Every, noJitter, immediateStart)
	case "count", "duration":
		interval := f.config.Every
		if interval == 0 {
			interval = immediateInterval // Immediate execution for count/duration without --every
		}
		baseScheduler, err = scheduler.NewIntervalScheduler(interval, noJitter, immediateStart)
	case "cron":
		baseScheduler, err = f.createCronScheduler()
	case "adaptive":
		baseScheduler, err = f.createAdaptiveScheduler()

	// EXISTING RATE CONTROL
	case "rate-limit":
		baseScheduler, err = f.createRateLimitScheduler()
	case "load-adaptive":
		baseScheduler, err = f.createLoadAdaptiveScheduler()

	default:
		return nil, fmt.Errorf("unknown subcommand: %s", f.config.Subcommand)
	}

	if err != nil {
		return nil, err
	}

	// Wrap with HTTP-aware scheduler if enabled
	return f.wrapWithHTTPAware(baseScheduler)
}

// wrapWithHTTPAware wraps a scheduler with HTTP-aware functionality if enabled
func (f *SchedulerFactory) wrapWithHTTPAware(baseScheduler interfaces.Scheduler) (interfaces.Scheduler, error) {
	httpConfig := f.config.GetHTTPAwareConfig()
	if httpConfig == nil {
		return baseScheduler, nil
	}

	// Create HTTP-aware scheduler with the base scheduler as fallback
	f.httpAwareScheduler = httpaware.NewHTTPAwareScheduler(*httpConfig)

	// For now, we'll use the base scheduler for timing and integrate HTTP-aware logic
	// in the execution loop. This allows us to maintain compatibility with existing
	// scheduler interfaces while adding HTTP-aware intelligence.
	return baseScheduler, nil
}

// createRateLimitScheduler creates a rate-limit aware scheduler
func (f *SchedulerFactory) createRateLimitScheduler() (interfaces.Scheduler, error) {
	// Parse rate specification
	rate, period, err := ratelimit.ParseRateSpec(f.config.RateSpec)
	if err != nil {
		return nil, fmt.Errorf("invalid rate spec: %w", err)
	}

	// Parse retry pattern if provided
	var retryPattern []time.Duration
	if f.config.RetryPattern != "" {
		retryPattern, err = f.parseRetryPattern(f.config.RetryPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid retry pattern: %w", err)
		}
	} else {
		retryPattern = []time.Duration{0} // Default: single attempt, no retries
	}

	// Create Diophantine rate limiter
	limiter := ratelimit.NewDiophantineRateLimiter(rate, period, retryPattern)

	// Create a scheduler that respects the rate limiter
	return NewRateLimitScheduler(limiter, f.config.ShowNext), nil
}

// parseRetryPattern parses retry pattern string like "0,10m,30m"
func (f *SchedulerFactory) parseRetryPattern(pattern string) ([]time.Duration, error) {
	parts := strings.Split(pattern, ",")
	retryPattern := make([]time.Duration, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "0" {
			retryPattern[i] = 0
			continue
		}

		duration, err := time.ParseDuration(part)
		if err != nil {
			return nil, fmt.Errorf("invalid duration '%s': %w", part, err)
		}
		retryPattern[i] = duration
	}

	return retryPattern, nil
}

// createCronScheduler creates a cron-based scheduler
func (f *SchedulerFactory) createCronScheduler() (interfaces.Scheduler, error) {
	cronExpr := f.config.CronExpression
	if cronExpr == "" {
		return nil, fmt.Errorf("cron expression is required")
	}

	// Create cron scheduler with UTC timezone by default
	return scheduler.NewCronScheduler(cronExpr, "UTC")
}

// createAdaptiveScheduler creates an adaptive scheduler
func (f *SchedulerFactory) createAdaptiveScheduler() (interfaces.Scheduler, error) {
	config := adaptive.DefaultAdaptiveConfig()

	// Override with command-line settings if provided
	if f.config.Every > 0 {
		config.BaseInterval = f.config.Every
	}

	adaptiveScheduler := adaptive.NewAdaptiveScheduler(config)

	// Wrap it to implement Scheduler interface
	return NewAdaptiveSchedulerWrapper(adaptiveScheduler, f.config), nil
}

// createLoadAdaptiveScheduler creates a load-adaptive scheduler
func (f *SchedulerFactory) createLoadAdaptiveScheduler() (interfaces.Scheduler, error) {
	interval := f.config.Every
	if interval == 0 {
		interval = 1 * time.Second // Default interval
	}

	// Use default target values: 70% CPU, 80% memory, 0.5 load
	return scheduler.NewLoadAwareScheduler(interval, 0.7, 0.8, 0.5), nil
}

// createStrategyScheduler creates a strategy-based scheduler for mathematical retry patterns
func (f *SchedulerFactory) createStrategyScheduler(strategyName string) (interfaces.Scheduler, error) {
	var strategy strategies.Strategy
	var err error

	// Get configuration values
	baseDelay := f.getBaseDelay()
	maxDelay := f.getMaxDelay()
	maxAttempts := int(f.config.Times)
	if maxAttempts <= 0 {
		maxAttempts = 3 // Default attempts
	}

	// Create strategy config
	strategyConfig := &strategies.StrategyConfig{
		BaseDelay:   baseDelay,
		MaxDelay:    maxDelay,
		MaxAttempts: maxAttempts,
	}

	// Create the appropriate strategy
	switch strategyName {
	case "exponential":
		multiplier := f.getMultiplier()
		if multiplier <= 1.0 {
			multiplier = 2.0 // Default exponential multiplier
		}
		strategyConfig.Multiplier = multiplier
		strategy = strategies.NewExponentialStrategy(baseDelay, multiplier, maxDelay)

	case "fibonacci":
		strategy = strategies.NewFibonacciStrategy(baseDelay, maxDelay)

	case "linear":
		increment := f.getIncrement()
		if increment == 0 {
			increment = baseDelay // Default increment
		}
		strategyConfig.Increment = increment
		strategy = strategies.NewLinearStrategy(increment, maxDelay)

	case "polynomial":
		exponent := f.getExponent()
		if exponent <= 1.0 {
			exponent = 2.0 // Default quadratic growth
		}
		strategyConfig.Exponent = exponent
		strategy = strategies.NewPolynomialStrategy(baseDelay, exponent, maxDelay)

	case "decorrelated-jitter":
		multiplier := f.getMultiplier()
		if multiplier <= 1.0 {
			multiplier = 3.0 // Default decorrelated jitter multiplier
		}
		strategyConfig.Multiplier = multiplier
		strategy = strategies.NewDecorrelatedJitterStrategy(baseDelay, multiplier, maxDelay)

	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategyName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s strategy: %w", strategyName, err)
	}

	return scheduler.NewStrategyScheduler(strategy, strategyConfig)
}

// Configuration helper methods

func (f *SchedulerFactory) getBaseDelay() time.Duration {
	if f.config.BaseDelay > 0 {
		return f.config.BaseDelay
	}
	return 1 * time.Second // Default base delay
}

func (f *SchedulerFactory) getIncrement() time.Duration {
	if f.config.Increment > 0 {
		return f.config.Increment
	}
	return 1 * time.Second // Default increment
}

func (f *SchedulerFactory) getMultiplier() float64 {
	if f.config.Multiplier > 0 {
		return f.config.Multiplier
	}
	return 2.0 // Default multiplier
}

func (f *SchedulerFactory) getExponent() float64 {
	if f.config.Exponent > 0 {
		return f.config.Exponent
	}
	return 2.0 // Default exponent
}

func (f *SchedulerFactory) getMaxDelay() time.Duration {
	if f.config.MaxDelay > 0 {
		return f.config.MaxDelay
	}
	return 60 * time.Second // Default max delay
}

// GetHTTPAwareScheduler returns the HTTP-aware scheduler if created
func (f *SchedulerFactory) GetHTTPAwareScheduler() httpaware.HTTPAwareScheduler {
	return f.httpAwareScheduler
}
