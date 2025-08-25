package cli

import (
	"errors"
	"fmt"
)

// ValidateConfig validates the parsed configuration
func ValidateConfig(config *Config) error {
	if config.Help || config.Version || config.SubcommandHelp {
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
