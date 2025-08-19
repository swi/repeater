package cli

import (
	"testing"
	"time"
)

// TestMathematicalStrategyValidation tests all mathematical strategy validation functions
func TestMathematicalStrategyValidation(t *testing.T) {
	tests := []struct {
		name          string
		subcommand    string
		config        *Config
		expectedError string
	}{
		// Exponential Strategy Tests
		{
			name:       "exponential_with_valid_config",
			subcommand: "exponential",
			config: &Config{
				Subcommand: "exponential",
				BaseDelay:  time.Second,
				MaxDelay:   10 * time.Second,
				Multiplier: 2.0,
				MaxRetries: 5,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name:       "exponential_missing_base_delay",
			subcommand: "exponential",
			config: &Config{
				Subcommand: "exponential",
				Command:    []string{"echo", "test"},
			},
			expectedError: "--base-delay is required for exponential strategy",
		},
		{
			name:       "exponential_invalid_multiplier",
			subcommand: "exponential",
			config: &Config{
				Subcommand: "exponential",
				BaseDelay:  time.Second,
				Multiplier: 0.5, // Invalid: must be > 1
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid exponential config: multiplier must be greater than 1.0",
		},
		{
			name:       "exponential_max_delay_less_than_base",
			subcommand: "exponential",
			config: &Config{
				Subcommand: "exponential",
				BaseDelay:  10 * time.Second,
				MaxDelay:   5 * time.Second, // Invalid: less than base delay
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid exponential config: max-delay must be greater than base-delay",
		},

		// Fibonacci Strategy Tests
		{
			name:       "fibonacci_with_valid_config",
			subcommand: "fibonacci",
			config: &Config{
				Subcommand: "fibonacci",
				BaseDelay:  time.Second,
				MaxDelay:   30 * time.Second,
				MaxRetries: 8,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name:       "fibonacci_missing_base_delay",
			subcommand: "fibonacci",
			config: &Config{
				Subcommand: "fibonacci",
				Command:    []string{"echo", "test"},
			},
			expectedError: "--base-delay is required for fibonacci strategy",
		},
		{
			name:       "fibonacci_max_delay_less_than_base",
			subcommand: "fibonacci",
			config: &Config{
				Subcommand: "fibonacci",
				BaseDelay:  10 * time.Second,
				MaxDelay:   5 * time.Second,
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid fibonacci config: max-delay must be greater than base-delay",
		},

		// Linear Strategy Tests
		{
			name:       "linear_with_valid_config",
			subcommand: "linear",
			config: &Config{
				Subcommand: "linear",
				Increment:  2 * time.Second,
				MaxDelay:   20 * time.Second,
				MaxRetries: 5,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name:       "linear_missing_increment",
			subcommand: "linear",
			config: &Config{
				Subcommand: "linear",
				Command:    []string{"echo", "test"},
			},
			expectedError: "--increment is required for linear strategy",
		},
		{
			name:       "linear_negative_increment",
			subcommand: "linear",
			config: &Config{
				Subcommand: "linear",
				Increment:  -time.Second, // Invalid: negative
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid linear config: increment must be positive",
		},
		{
			name:       "linear_max_delay_less_than_increment",
			subcommand: "linear",
			config: &Config{
				Subcommand: "linear",
				Increment:  10 * time.Second,
				MaxDelay:   5 * time.Second,
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid linear config: max-delay must be greater than increment",
		},

		// Polynomial Strategy Tests
		{
			name:       "polynomial_with_valid_config",
			subcommand: "polynomial",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  time.Second,
				Exponent:   1.5,
				MaxDelay:   60 * time.Second,
				MaxRetries: 4,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name:       "polynomial_missing_base_delay",
			subcommand: "polynomial",
			config: &Config{
				Subcommand: "polynomial",
				Exponent:   2.0,
				Command:    []string{"echo", "test"},
			},
			expectedError: "--base-delay is required for polynomial strategy",
		},
		{
			name:       "polynomial_large_exponent",
			subcommand: "polynomial",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  time.Second,
				Exponent:   15.0, // Invalid: must be <= 10
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid polynomial config: exponent must be <= 10.0 to prevent overflow",
		},
		{
			name:       "polynomial_max_delay_less_than_base",
			subcommand: "polynomial",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  10 * time.Second,
				Exponent:   2.0,
				MaxDelay:   5 * time.Second,
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid polynomial config: max-delay must be greater than base-delay",
		},

		// Decorrelated Jitter Strategy Tests
		{
			name:       "decorrelated_jitter_with_valid_config",
			subcommand: "decorrelated-jitter",
			config: &Config{
				Subcommand: "decorrelated-jitter",
				BaseDelay:  time.Second,
				Multiplier: 3.0,
				MaxDelay:   120 * time.Second,
				MaxRetries: 6,
				Command:    []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name:       "decorrelated_jitter_missing_base_delay",
			subcommand: "decorrelated-jitter",
			config: &Config{
				Subcommand: "decorrelated-jitter",
				Command:    []string{"echo", "test"},
			},
			expectedError: "--base-delay is required for decorrelated-jitter strategy",
		},
		{
			name:       "decorrelated_jitter_invalid_multiplier",
			subcommand: "decorrelated-jitter",
			config: &Config{
				Subcommand: "decorrelated-jitter",
				BaseDelay:  time.Second,
				Multiplier: 0.8, // Invalid: must be > 1
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid decorrelated-jitter config: multiplier must be greater than 1.0",
		},
		{
			name:       "decorrelated_jitter_max_delay_less_than_base",
			subcommand: "decorrelated-jitter",
			config: &Config{
				Subcommand: "decorrelated-jitter",
				BaseDelay:  10 * time.Second,
				MaxDelay:   5 * time.Second,
				Command:    []string{"echo", "test"},
			},
			expectedError: "invalid decorrelated-jitter config: max-delay must be greater than base-delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

// TestCronValidation tests cron expression validation
func TestCronValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError string
	}{
		{
			name: "cron_with_valid_expression",
			config: &Config{
				Subcommand:     "cron",
				CronExpression: "0 9 * * *", // Every day at 9 AM
				Command:        []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "cron_with_shortcut_expression",
			config: &Config{
				Subcommand:     "cron",
				CronExpression: "@daily",
				Command:        []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "cron_missing_expression",
			config: &Config{
				Subcommand: "cron",
				Command:    []string{"echo", "test"},
			},
			expectedError: "--cron is required for cron subcommand",
		},
		{
			name: "cron_with_timezone",
			config: &Config{
				Subcommand:     "cron",
				CronExpression: "0 9 * * *",
				Timezone:       "America/New_York",
				Command:        []string{"echo", "test"},
			},
			expectedError: "",
		},
		{
			name: "cron_with_invalid_timezone",
			config: &Config{
				Subcommand:     "cron",
				CronExpression: "0 9 * * *",
				Timezone:       "Invalid/Timezone",
				Command:        []string{"echo", "test"},
			},
			expectedError: "", // Current implementation doesn't validate timezone
		},
		{
			name: "cron_with_invalid_expression",
			config: &Config{
				Subcommand:     "cron",
				CronExpression: "invalid cron",
				Command:        []string{"echo", "test"},
			},
			expectedError: "invalid cron config: cron expression must have 5 fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

// TestRateSpecValidation tests rate specification validation in detail
func TestRateSpecValidation(t *testing.T) {
	tests := []struct {
		name          string
		rateSpec      string
		expectedError string
	}{
		{
			name:          "valid_rate_spec_per_hour",
			rateSpec:      "100/1h",
			expectedError: "",
		},
		{
			name:          "valid_rate_spec_per_minute",
			rateSpec:      "10/1m",
			expectedError: "",
		},
		{
			name:          "valid_rate_spec_per_second",
			rateSpec:      "5/1s",
			expectedError: "",
		},
		{
			name:          "empty_rate_spec",
			rateSpec:      "",
			expectedError: "rate spec cannot be empty",
		},
		{
			name:          "missing_slash",
			rateSpec:      "100",
			expectedError: "rate spec must be in format 'rate/period' (e.g., '10/1h')",
		},
		{
			name:          "multiple_slashes",
			rateSpec:      "100/1h/extra",
			expectedError: "rate spec must be in format 'rate/period' (e.g., '10/1h')",
		},
		{
			name:          "invalid_rate_number",
			rateSpec:      "abc/1h",
			expectedError: "invalid rate number: abc",
		},
		{
			name:          "invalid_period_duration",
			rateSpec:      "100/invalid",
			expectedError: "invalid period duration: invalid",
		},
		{
			name:          "negative_rate",
			rateSpec:      "-100/1h",
			expectedError: "",
		},
		{
			name:          "zero_rate",
			rateSpec:      "0/1h",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRateSpec(tt.rateSpec)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.expectedError)
				} else if !contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing '%s', but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

// TestValidateConfigEdgeCases tests edge cases in main ValidateConfig function
func TestValidateConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError string
	}{
		{
			name: "help_flag_skips_validation",
			config: &Config{
				Help: true,
				// Missing required fields, but should be ignored
			},
			expectedError: "",
		},
		{
			name: "version_flag_skips_validation",
			config: &Config{
				Version: true,
				// Missing required fields, but should be ignored
			},
			expectedError: "",
		},
		{
			name: "missing_subcommand",
			config: &Config{
				Command: []string{"echo", "test"},
			},
			expectedError: "subcommand required",
		},
		{
			name: "missing_command",
			config: &Config{
				Subcommand: "interval",
				Every:      time.Second,
			},
			expectedError: "command required after --",
		},
		{
			name: "unknown_subcommand",
			config: &Config{
				Subcommand: "unknown",
				Command:    []string{"echo", "test"},
			},
			expectedError: "", // Unknown subcommands should not error in ValidateConfig
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error '%s', but got no error", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("Expected error '%s', but got '%s'", tt.expectedError, err.Error())
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
