package cli

import (
	"errors"
	"time"
)

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
