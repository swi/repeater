package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/patterns"
)

func TestCLI_PatternMatchingFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{
			name: "success pattern flag",
			args: []string{"interval", "--every", "30s", "--success-pattern", "deployment successful", "--", "echo", "test"},
			expected: Config{
				Subcommand:     "interval",
				Every:          30 * time.Second,
				SuccessPattern: "deployment successful",
				Command:        []string{"echo", "test"},
			},
		},
		{
			name: "failure pattern flag",
			args: []string{"interval", "--every", "30s", "--failure-pattern", "(?i)error|failed", "--", "echo", "test"},
			expected: Config{
				Subcommand:     "interval",
				Every:          30 * time.Second,
				FailurePattern: "(?i)error|failed",
				Command:        []string{"echo", "test"},
			},
		},
		{
			name: "both patterns with case insensitive",
			args: []string{"interval", "--every", "30s", "--success-pattern", "success", "--failure-pattern", "error", "--case-insensitive", "--", "echo", "test"},
			expected: Config{
				Subcommand:      "interval",
				Every:           30 * time.Second,
				SuccessPattern:  "success",
				FailurePattern:  "error",
				CaseInsensitive: true,
				Command:         []string{"echo", "test"},
			},
		},
		{
			name: "adaptive with patterns",
			args: []string{"adaptive", "--base-interval", "1s", "--success-pattern", "completed", "--", "build.sh"},
			expected: Config{
				Subcommand:     "adaptive",
				BaseInterval:   1 * time.Second,
				SuccessPattern: "completed",
				Command:        []string{"build.sh"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.BaseInterval, config.BaseInterval)
			assert.Equal(t, tt.expected.SuccessPattern, config.SuccessPattern)
			assert.Equal(t, tt.expected.FailurePattern, config.FailurePattern)
			assert.Equal(t, tt.expected.CaseInsensitive, config.CaseInsensitive)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

func TestCLI_PatternValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid success pattern regex",
			args:        []string{"interval", "--every", "30s", "--success-pattern", "[invalid", "--", "echo", "test"},
			expectError: true,
			errorMsg:    "invalid success pattern",
		},
		{
			name:        "invalid failure pattern regex",
			args:        []string{"interval", "--every", "30s", "--failure-pattern", "*invalid", "--", "echo", "test"},
			expectError: true,
			errorMsg:    "invalid failure pattern",
		},
		{
			name:        "valid patterns should not error",
			args:        []string{"interval", "--every", "30s", "--success-pattern", "success", "--failure-pattern", "(?i)error", "--", "echo", "test"},
			expectError: false,
		},
		{
			name:        "case insensitive with valid patterns",
			args:        []string{"interval", "--every", "30s", "--success-pattern", "SUCCESS", "--case-insensitive", "--", "echo", "test"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetPatternConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected *patterns.PatternConfig
	}{
		{
			name: "no patterns configured",
			config: Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
			},
			expected: nil,
		},
		{
			name: "success pattern only",
			config: Config{
				Subcommand:     "interval",
				Every:          30 * time.Second,
				SuccessPattern: "deployment successful",
			},
			expected: &patterns.PatternConfig{
				SuccessPattern: "deployment successful",
			},
		},
		{
			name: "failure pattern only",
			config: Config{
				Subcommand:     "interval",
				Every:          30 * time.Second,
				FailurePattern: "(?i)error",
			},
			expected: &patterns.PatternConfig{
				FailurePattern: "(?i)error",
			},
		},
		{
			name: "both patterns with case insensitive",
			config: Config{
				Subcommand:      "interval",
				Every:           30 * time.Second,
				SuccessPattern:  "success",
				FailurePattern:  "error",
				CaseInsensitive: true,
			},
			expected: &patterns.PatternConfig{
				SuccessPattern:  "success",
				FailurePattern:  "error",
				CaseInsensitive: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPatternConfig()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.SuccessPattern, result.SuccessPattern)
				assert.Equal(t, tt.expected.FailurePattern, result.FailurePattern)
				assert.Equal(t, tt.expected.CaseInsensitive, result.CaseInsensitive)
			}
		})
	}
}
