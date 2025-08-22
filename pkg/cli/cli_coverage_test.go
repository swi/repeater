package cli

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseIntFlag tests the parseIntFlag function specifically
func TestParseIntFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		pos      int
		expected int
		wantErr  bool
	}{
		{
			name:     "valid integer flag",
			args:     []string{"--port", "8080"},
			pos:      0,
			expected: 8080,
			wantErr:  false,
		},
		{
			name:     "negative integer flag",
			args:     []string{"--count", "-5"},
			pos:      0,
			expected: -5,
			wantErr:  false,
		},
		{
			name:    "missing value",
			args:    []string{"--port"},
			pos:     0,
			wantErr: true,
		},
		{
			name:    "invalid integer value",
			args:    []string{"--port", "invalid"},
			pos:     0,
			wantErr: true,
		},
		{
			name:    "empty string value",
			args:    []string{"--port", ""},
			pos:     0,
			wantErr: true,
		},
		{
			name:     "zero value",
			args:     []string{"--port", "0"},
			pos:      0,
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{
				args: tt.args,
				pos:  tt.pos,
			}

			var result int
			err := parser.parseIntFlag(&result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestNormalizeSubcommandEdgeCases tests edge cases in normalizeSubcommand function
func TestNormalizeSubcommandEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Test all the abbreviations that might not be covered
		{
			name:     "count abbreviation cnt",
			input:    "cnt",
			expected: "count",
		},
		{
			name:     "duration abbreviation dur",
			input:    "dur",
			expected: "duration",
		},
		{
			name:     "cron abbreviation cr",
			input:    "cr",
			expected: "cron",
		},
		{
			name:     "adaptive abbreviation adapt",
			input:    "adapt",
			expected: "adaptive",
		},
		{
			name:     "load-adaptive abbreviation la",
			input:    "la",
			expected: "load-adaptive",
		},
		{
			name:     "rate-limit abbreviation rate",
			input:    "rate",
			expected: "rate-limit",
		},
		{
			name:     "rate-limit abbreviation rl",
			input:    "rl",
			expected: "rate-limit",
		},
		{
			name:     "exponential abbreviation exp",
			input:    "exp",
			expected: "exponential",
		},
		{
			name:     "fibonacci abbreviation fib",
			input:    "fib",
			expected: "fibonacci",
		},
		{
			name:     "linear abbreviation lin",
			input:    "lin",
			expected: "linear",
		},
		{
			name:     "polynomial abbreviation poly",
			input:    "poly",
			expected: "polynomial",
		},
		{
			name:     "decorrelated-jitter abbreviation dj",
			input:    "dj",
			expected: "decorrelated-jitter",
		},
		{
			name:     "unknown abbreviation returns empty",
			input:    "unknown",
			expected: "",
		},
		{
			name:     "empty string stays empty",
			input:    "",
			expected: "",
		},
		{
			name:     "already normalized stays the same",
			input:    "interval",
			expected: "interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSubcommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseSubcommandFlagsEdgeCases tests uncovered paths in parseSubcommandFlags
func TestParseSubcommandFlagsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		config  *Config
		wantErr bool
	}{
		{
			name: "unknown flag should return error",
			args: []string{"interval", "--unknown-flag", "value", "--", "echo", "test"},
			config: &Config{
				Subcommand: "interval",
			},
			wantErr: true,
		},
		{
			name: "flag without double dash separator",
			args: []string{"interval", "--every", "1s", "echo", "test"},
			config: &Config{
				Subcommand: "interval",
			},
			wantErr: true,
		},
		{
			name: "multiple unknown flags",
			args: []string{"interval", "--unknown1", "val1", "--unknown2", "val2", "--", "echo", "test"},
			config: &Config{
				Subcommand: "interval",
			},
			wantErr: true,
		},
		{
			name: "flag at end without value",
			args: []string{"interval", "--every"},
			config: &Config{
				Subcommand: "interval",
			},
			wantErr: true,
		},
		{
			name: "valid minimal interval config",
			args: []string{"interval", "--every", "1s", "--", "echo", "test"},
			config: &Config{
				Subcommand: "interval",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{
				args: tt.args,
				pos:  1, // Start after subcommand
			}

			parser.config = tt.config
			err := parser.parseSubcommandFlags()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateFibonacciConfigEdgeCases tests uncovered cases in validateFibonacciConfig
func TestValidateFibonacciConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "missing base-delay should error",
			config: &Config{
				Subcommand: "fibonacci",
				MaxDelay:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero base-delay should error",
			config: &Config{
				Subcommand: "fibonacci",
				BaseDelay:  0,
				MaxDelay:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "base-delay equal to max-delay should work",
			config: &Config{
				Subcommand: "fibonacci",
				BaseDelay:  30 * time.Second,
				MaxDelay:   30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid fibonacci config with defaults",
			config: &Config{
				Subcommand: "fibonacci",
				BaseDelay:  1 * time.Second,
				MaxDelay:   60 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFibonacciConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePolynomialConfigEdgeCases tests uncovered cases in validatePolynomialConfig
func TestValidatePolynomialConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "missing base-delay should error",
			config: &Config{
				Subcommand: "polynomial",
				Exponent:   2.0,
				MaxDelay:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero exponent gets default (no error)",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   0.0,
				MaxDelay:   30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "negative exponent should error",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   -1.5,
				MaxDelay:   30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "exponent of 1.0 should work (linear case)",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   1.0,
				MaxDelay:   30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid polynomial config with fractional exponent",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   1.5,
				MaxDelay:   60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "high exponent should work",
			config: &Config{
				Subcommand: "polynomial",
				BaseDelay:  1 * time.Second,
				Exponent:   3.0,
				MaxDelay:   120 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePolynomialConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestParseTimesFlag tests edge cases for parseTimesFlag to improve coverage
func TestParseTimesFlagEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		pos      int
		expected int64
		wantErr  bool
	}{
		{
			name:     "large number",
			args:     []string{"--times", "999999"},
			pos:      0,
			expected: 999999,
			wantErr:  false,
		},
		{
			name:     "negative number should parse",
			args:     []string{"--times", "-5"},
			pos:      0,
			expected: -5,
			wantErr:  false,
		},
		{
			name:     "zero should parse",
			args:     []string{"--times", "0"},
			pos:      0,
			expected: 0,
			wantErr:  false,
		},

		{
			name:    "invalid format",
			args:    []string{"--times", "abc"},
			pos:     0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{
				args:   tt.args,
				pos:    tt.pos,
				config: &Config{},
			}

			err := parser.parseTimesFlag()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, parser.config.Times)
			}
		})
	}
}

// TestParseFloatFlagEdgeCases tests edge cases for parseFloatFlag to improve coverage
func TestParseFloatFlagEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		pos      int
		expected float64
		wantErr  bool
	}{
		{
			name:     "scientific notation",
			args:     []string{"--multiplier", "1.5e2"},
			pos:      0,
			expected: 150.0,
			wantErr:  false,
		},
		{
			name:     "very small decimal",
			args:     []string{"--threshold", "0.001"},
			pos:      0,
			expected: 0.001,
			wantErr:  false,
		},
		{
			name:     "infinity should parse as infinity",
			args:     []string{"--value", "inf"},
			pos:      0,
			expected: math.Inf(1),
			wantErr:  false,
		},
		{
			name:    "not a number should error",
			args:    []string{"--value", "notfloat"},
			pos:     0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{
				args: tt.args,
				pos:  tt.pos,
			}

			var result float64
			err := parser.parseFloatFlag(&result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestParseStringSliceFlagEdgeCases tests edge cases for parseStringSliceFlag
func TestParseStringSliceFlagEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		pos      int
		expected []string
		wantErr  bool
	}{
		{
			name:     "empty values in slice",
			args:     []string{"--fields", "field1,,field3"},
			pos:      0,
			expected: []string{"field1", "", "field3"},
			wantErr:  false,
		},
		{
			name:     "single value",
			args:     []string{"--fields", "single"},
			pos:      0,
			expected: []string{"single"},
			wantErr:  false,
		},
		{
			name:     "values with spaces",
			args:     []string{"--fields", "field with spaces,another field"},
			pos:      0,
			expected: []string{"field with spaces", "another field"},
			wantErr:  false,
		},
		{
			name:    "missing value should error",
			args:    []string{"--fields"},
			pos:     0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &argParser{
				args: tt.args,
				pos:  tt.pos,
			}

			var result []string
			err := parser.parseStringSliceFlag(&result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
