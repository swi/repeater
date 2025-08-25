package cli

import (
	"errors"
	"fmt"

	"github.com/swi/repeater/pkg/patterns"
)

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
