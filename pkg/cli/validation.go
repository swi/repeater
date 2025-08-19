package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/swi/repeater/pkg/patterns"
)

// ValidateConfig validates the parsed configuration
func ValidateConfig(config *Config) error {
	if config.Help || config.Version {
		return nil // Help and version don't need validation
	}

	if config.Subcommand == "" {
		return errors.New("subcommand required")
	}

	if len(config.Command) == 0 {
		return errors.New("command required after --")
	}

	// Validate output control flags
	if err := validateOutputFlags(config); err != nil {
		return err
	}

	// Validate pattern matching configuration
	if err := validatePatterns(config); err != nil {
		return err
	}

	// Validate subcommand-specific requirements
	switch config.Subcommand {
	case "interval":
		if config.Every == 0 {
			return errors.New("--every is required for interval subcommand")
		}
	case "count":
		if config.Times == 0 {
			return errors.New("--times is required for count subcommand")
		}
	case "duration":
		if config.For == 0 {
			return errors.New("--for is required for duration subcommand")
		}
	case "rate-limit":
		if config.RateSpec == "" {
			return errors.New("--rate is required for rate-limit subcommand")
		}
		// Validate rate spec format
		if err := validateRateSpec(config.RateSpec); err != nil {
			return fmt.Errorf("invalid rate spec: %w", err)
		}
	case "adaptive":
		if config.BaseInterval == 0 {
			return errors.New("--base-interval is required for adaptive subcommand")
		}
		// Validate adaptive configuration
		if err := validateAdaptiveConfig(config); err != nil {
			return fmt.Errorf("invalid adaptive config: %w", err)
		}
	case "load-adaptive":
		if config.BaseInterval == 0 {
			return errors.New("--base-interval is required for load-adaptive subcommand")
		}
		// Validate load-adaptive configuration
		if err := validateLoadAdaptiveConfig(config); err != nil {
			return fmt.Errorf("invalid load-adaptive config: %w", err)
		}
	case "cron":
		if config.CronExpression == "" {
			return errors.New("--cron is required for cron subcommand")
		}
		// Validate cron configuration
		if err := validateCronConfig(config); err != nil {
			return fmt.Errorf("invalid cron config: %w", err)
		}
	case "exponential":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for exponential strategy")
		}
		// Validate exponential strategy configuration
		if err := validateExponentialConfig(config); err != nil {
			return fmt.Errorf("invalid exponential config: %w", err)
		}
	case "fibonacci":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for fibonacci strategy")
		}
		// Validate fibonacci strategy configuration
		if err := validateFibonacciConfig(config); err != nil {
			return fmt.Errorf("invalid fibonacci config: %w", err)
		}
	case "linear":
		if config.Increment == 0 {
			return errors.New("--increment is required for linear strategy")
		}
		// Validate linear strategy configuration
		if err := validateLinearConfig(config); err != nil {
			return fmt.Errorf("invalid linear config: %w", err)
		}
	case "polynomial":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for polynomial strategy")
		}
		// Validate polynomial strategy configuration
		if err := validatePolynomialConfig(config); err != nil {
			return fmt.Errorf("invalid polynomial config: %w", err)
		}
	case "decorrelated-jitter":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for decorrelated-jitter strategy")
		}
		// Validate decorrelated-jitter strategy configuration
		if err := validateDecorrelatedJitterConfig(config); err != nil {
			return fmt.Errorf("invalid decorrelated-jitter config: %w", err)
		}
	}

	return nil
}

// validateRateSpec validates the rate specification format
func validateRateSpec(spec string) error {
	// Use the ParseRateSpec function from ratelimit package to validate
	// For now, do basic validation here to avoid circular imports
	if spec == "" {
		return errors.New("rate spec cannot be empty")
	}

	// Basic format check: should contain "/"
	if !strings.Contains(spec, "/") {
		return errors.New("rate spec must be in format 'rate/period' (e.g., '10/1h')")
	}

	parts := strings.Split(spec, "/")
	if len(parts) != 2 {
		return errors.New("rate spec must be in format 'rate/period' (e.g., '10/1h')")
	}

	// Validate rate part is a number
	if _, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err != nil {
		return fmt.Errorf("invalid rate number: %s", parts[0])
	}

	// Validate period part is a valid duration
	if _, err := time.ParseDuration(strings.TrimSpace(parts[1])); err != nil {
		return fmt.Errorf("invalid period duration: %s", parts[1])
	}

	return nil
}

// validateAdaptiveConfig validates the adaptive configuration
func validateAdaptiveConfig(config *Config) error {
	// Set defaults if not provided
	if config.MinInterval == 0 {
		config.MinInterval = 100 * time.Millisecond
	}
	if config.MaxInterval == 0 {
		config.MaxInterval = 30 * time.Second
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 2.0
	}
	if config.FastThreshold == 0 {
		config.FastThreshold = 0.5
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 0.3
	}

	// Validate bounds
	if config.MinInterval >= config.MaxInterval {
		return errors.New("min-interval must be less than max-interval")
	}

	if config.BaseInterval < config.MinInterval || config.BaseInterval > config.MaxInterval {
		return errors.New("base-interval must be between min-interval and max-interval")
	}

	if config.SlowThreshold <= 1.0 {
		return errors.New("slow-threshold must be greater than 1.0")
	}

	if config.FastThreshold <= 0 || config.FastThreshold >= 1.0 {
		return errors.New("fast-threshold must be between 0 and 1.0")
	}

	if config.FailureThreshold <= 0 || config.FailureThreshold >= 1.0 {
		return errors.New("failure-threshold must be between 0 and 1.0")
	}

	return nil
}

// validateLoadAdaptiveConfig validates the load-adaptive configuration
func validateLoadAdaptiveConfig(config *Config) error {
	// Set defaults if not provided
	if config.TargetCPU == 0 {
		config.TargetCPU = 70.0 // Default 70% CPU target
	}
	if config.TargetMemory == 0 {
		config.TargetMemory = 80.0 // Default 80% memory target
	}
	if config.TargetLoad == 0 {
		config.TargetLoad = 1.0 // Default load average of 1.0
	}
	if config.MinInterval == 0 {
		config.MinInterval = config.BaseInterval / 10
	}
	if config.MaxInterval == 0 {
		config.MaxInterval = config.BaseInterval * 10
	}

	// Validate bounds
	if config.TargetCPU <= 0 || config.TargetCPU > 100 {
		return errors.New("target-cpu must be between 0 and 100")
	}

	if config.TargetMemory <= 0 || config.TargetMemory > 100 {
		return errors.New("target-memory must be between 0 and 100")
	}

	if config.TargetLoad <= 0 {
		return errors.New("target-load must be greater than 0")
	}

	if config.MinInterval >= config.MaxInterval {
		return errors.New("min-interval must be less than max-interval")
	}

	return nil
}

// validateOutputFlags validates output control flags for conflicts
func validateOutputFlags(config *Config) error {
	// Check for conflicting flags
	if config.Quiet && config.Stream {
		return errors.New("--quiet and --stream flags are mutually exclusive")
	}

	if config.Quiet && config.Verbose {
		return errors.New("--quiet and --verbose flags are mutually exclusive")
	}

	if config.StatsOnly && config.Stream {
		return errors.New("--stats-only and --stream flags are mutually exclusive")
	}

	if config.StatsOnly && config.Verbose {
		return errors.New("--stats-only and --verbose flags are mutually exclusive")
	}

	if config.StatsOnly && config.Quiet {
		return errors.New("--stats-only and --quiet flags are mutually exclusive")
	}

	// Note: --stream and --verbose can be used together for detailed streaming

	return nil
}

// validateCronConfig validates the cron configuration
func validateCronConfig(config *Config) error {
	// Import cron package to validate expression
	// For now, do basic validation to avoid circular imports
	if config.CronExpression == "" {
		return errors.New("cron expression cannot be empty")
	}

	// Set default timezone if not specified
	if config.Timezone == "" {
		config.Timezone = "UTC"
	}

	// Basic validation - check if it looks like a cron expression or shortcut
	expr := strings.TrimSpace(config.CronExpression)
	if strings.HasPrefix(expr, "@") {
		// Shortcut format
		validShortcuts := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@hourly"}
		if !slices.Contains(validShortcuts, expr) {
			return fmt.Errorf("invalid cron shortcut: %s (valid shortcuts: %s)", expr, strings.Join(validShortcuts, ", "))
		}
	} else {
		// Standard cron format - should have 5 fields
		fields := strings.Fields(expr)
		if len(fields) != 5 {
			return fmt.Errorf("cron expression must have 5 fields (minute hour day month weekday), got %d", len(fields))
		}
	}

	return nil
}

// validatePatterns validates regex patterns for success/failure matching
func validatePatterns(config *Config) error {
	// Validate success pattern if provided
	if config.SuccessPattern != "" {
		pattern := config.SuccessPattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		if _, err := patterns.NewPatternMatcher(patterns.PatternConfig{
			SuccessPattern: pattern,
		}); err != nil {
			return fmt.Errorf("invalid success pattern: %w", err)
		}
	}

	// Validate failure pattern if provided
	if config.FailurePattern != "" {
		pattern := config.FailurePattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		if _, err := patterns.NewPatternMatcher(patterns.PatternConfig{
			FailurePattern: pattern,
		}); err != nil {
			return fmt.Errorf("invalid failure pattern: %w", err)
		}
	}

	return nil
}

// validateExponentialConfig validates the exponential strategy configuration
func validateExponentialConfig(config *Config) error {
	// Set defaults if not provided
	if config.Multiplier == 0 {
		config.Multiplier = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateFibonacciConfig validates the fibonacci strategy configuration
func validateFibonacciConfig(config *Config) error {
	// Set defaults if not provided
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateLinearConfig validates the linear strategy configuration
func validateLinearConfig(config *Config) error {
	// Set defaults if not provided
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.Increment <= 0 {
		return errors.New("increment must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.Increment {
		return errors.New("max-delay must be greater than increment")
	}

	return nil
}

// validatePolynomialConfig validates the polynomial strategy configuration
func validatePolynomialConfig(config *Config) error {
	// Set defaults if not provided
	if config.Exponent == 0 {
		config.Exponent = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Exponent <= 0 {
		return errors.New("exponent must be positive")
	}

	if config.Exponent > 10.0 {
		return errors.New("exponent must be <= 10.0 to prevent overflow")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateDecorrelatedJitterConfig validates the decorrelated-jitter strategy configuration
func validateDecorrelatedJitterConfig(config *Config) error {
	// Set defaults if not provided
	if config.Multiplier == 0 {
		config.Multiplier = 3.0 // AWS recommendation
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}
