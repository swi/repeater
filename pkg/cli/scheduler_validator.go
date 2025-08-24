package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

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
