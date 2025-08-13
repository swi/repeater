package patterns

import (
	"fmt"
	"regexp"
)

// PatternConfig holds configuration for pattern matching
type PatternConfig struct {
	SuccessPattern  string
	FailurePattern  string
	CaseInsensitive bool
}

// EvaluationResult represents the result of pattern evaluation
type EvaluationResult struct {
	Success  bool
	ExitCode int
	Reason   string // For debugging/logging
}

// PatternMatcher handles success/failure pattern matching
type PatternMatcher struct {
	config       PatternConfig
	successRegex *regexp.Regexp
	failureRegex *regexp.Regexp
}

// NewPatternMatcher creates a new pattern matcher with the given configuration
func NewPatternMatcher(config PatternConfig) (*PatternMatcher, error) {
	matcher := &PatternMatcher{
		config: config,
	}

	// Compile success pattern if provided
	if config.SuccessPattern != "" {
		pattern := config.SuccessPattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}

		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid success pattern: %w", err)
		}
		matcher.successRegex = regex
	}

	// Compile failure pattern if provided
	if config.FailurePattern != "" {
		pattern := config.FailurePattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}

		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid failure pattern: %w", err)
		}
		matcher.failureRegex = regex
	}

	return matcher, nil
}

// EvaluateResult evaluates command output and exit code using configured patterns
func (pm *PatternMatcher) EvaluateResult(output string, exitCode int) EvaluationResult {
	// Pattern precedence:
	// 1. Failure pattern match → Command fails (exit code 1)
	// 2. Success pattern match → Command succeeds (exit code 0)
	// 3. Exit code → Standard behavior (0 = success, non-zero = failure)

	// Check failure pattern first (highest precedence)
	if pm.failureRegex != nil && pm.failureRegex.MatchString(output) {
		return EvaluationResult{
			Success:  false,
			ExitCode: 1,
			Reason:   "failure pattern matched",
		}
	}

	// Check success pattern second
	if pm.successRegex != nil && pm.successRegex.MatchString(output) {
		return EvaluationResult{
			Success:  true,
			ExitCode: 0,
			Reason:   "success pattern matched",
		}
	}

	// Fall back to exit code
	return EvaluationResult{
		Success:  exitCode == 0,
		ExitCode: exitCode,
		Reason:   "exit code used",
	}
}
