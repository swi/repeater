package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputControlFlags tests the new output control flags
func TestOutputControlFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "interval with stream flag",
			args: []string{"interval", "--every", "1s", "--stream", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      1 * time.Second,
				Stream:     true,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "interval with quiet flag",
			args: []string{"interval", "--every", "1s", "--quiet", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      1 * time.Second,
				Quiet:      true,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "interval with verbose flag",
			args: []string{"interval", "--every", "1s", "--verbose", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      1 * time.Second,
				Verbose:    true,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "count with stream and verbose flags",
			args: []string{"count", "--times", "5", "--stream", "--verbose", "--", "curl", "api.com"},
			expected: Config{
				Subcommand: "count",
				Times:      5,
				Stream:     true,
				Verbose:    true,
				Command:    []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "duration with output prefix",
			args: []string{"duration", "--for", "1m", "--stream", "--output-prefix", "[CMD]", "--", "echo", "test"},
			expected: Config{
				Subcommand:   "duration",
				For:          1 * time.Minute,
				Stream:       true,
				OutputPrefix: "[CMD]",
				Command:      []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name:     "conflicting quiet and stream flags",
			args:     []string{"interval", "--every", "1s", "--quiet", "--stream", "--", "echo", "test"},
			expected: Config{},
			wantErr:  true,
		},
		{
			name:     "conflicting quiet and verbose flags",
			args:     []string{"interval", "--every", "1s", "--quiet", "--verbose", "--", "echo", "test"},
			expected: Config{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.Times, config.Times)
			assert.Equal(t, tt.expected.For, config.For)
			assert.Equal(t, tt.expected.Stream, config.Stream)
			assert.Equal(t, tt.expected.Quiet, config.Quiet)
			assert.Equal(t, tt.expected.Verbose, config.Verbose)
			assert.Equal(t, tt.expected.OutputPrefix, config.OutputPrefix)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

// TestRateLimitSubcommand tests the new rate-limit subcommand parsing
func TestRateLimitSubcommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "rate-limit with basic rate spec",
			args: []string{"rate-limit", "--rate", "10/1h", "--", "curl", "api.com"},
			expected: Config{
				Subcommand: "rate-limit",
				RateSpec:   "10/1h",
				Command:    []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "rate-limit with retry pattern",
			args: []string{"rate-limit", "--rate", "100/1h", "--retry-pattern", "0,10m,30m", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:   "rate-limit",
				RateSpec:     "100/1h",
				RetryPattern: "0,10m,30m",
				Command:      []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "rate-limit with show-next flag",
			args: []string{"rate-limit", "--rate", "5/1m", "--show-next", "--", "echo", "test"},
			expected: Config{
				Subcommand: "rate-limit",
				RateSpec:   "5/1m",
				ShowNext:   true,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "rate-limit abbreviations",
			args: []string{"rl", "--rate", "50/1h", "--", "curl", "api.com"},
			expected: Config{
				Subcommand: "rate-limit",
				RateSpec:   "50/1h",
				Command:    []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "rate-limit short flags",
			args: []string{"rate-limit", "-r", "20/1m", "-p", "0,5m", "-n", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:   "rate-limit",
				RateSpec:     "20/1m",
				RetryPattern: "0,5m",
				ShowNext:     true,
				Command:      []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name:    "rate-limit missing rate spec",
			args:    []string{"rate-limit", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "rate-limit invalid rate spec",
			args:    []string{"rate-limit", "--rate", "invalid", "--", "curl", "api.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				// For error cases, check both parsing and validation
				if err == nil {
					err = ValidateConfig(config)
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, ValidateConfig(config))
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.RateSpec, config.RateSpec)
			assert.Equal(t, tt.expected.RetryPattern, config.RetryPattern)
			assert.Equal(t, tt.expected.ShowNext, config.ShowNext)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

func TestCLIParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "interval subcommand with basic flags",
			args: []string{"interval", "--every", "30s", "--times", "5", "--", "echo", "hello"},
			expected: Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
				Times:      5,
				Command:    []string{"echo", "hello"},
			},
			wantErr: false,
		},
		{
			name: "count subcommand with command",
			args: []string{"count", "--times", "3", "--", "date"},
			expected: Config{
				Subcommand: "count",
				Times:      3,
				Command:    []string{"date"},
			},
			wantErr: false,
		},
		{
			name: "duration subcommand",
			args: []string{"duration", "--for", "2m", "--every", "10s", "--", "curl", "http://example.com"},
			expected: Config{
				Subcommand: "duration",
				For:        2 * time.Minute,
				Every:      10 * time.Second,
				Command:    []string{"curl", "http://example.com"},
			},
			wantErr: false,
		},
		{
			name: "help flag",
			args: []string{"--help"},
			expected: Config{
				Help: true,
			},
			wantErr: false,
		},
		{
			name: "version flag",
			args: []string{"--version"},
			expected: Config{
				Version: true,
			},
			wantErr: false,
		},
		{
			name: "config file flag",
			args: []string{"--config", "/path/to/config.toml", "interval", "--every", "1s", "--", "echo", "test"},
			expected: Config{
				ConfigFile: "/path/to/config.toml",
				Subcommand: "interval",
				Every:      1 * time.Second,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.Times, config.Times)
			assert.Equal(t, tt.expected.For, config.For)
			assert.Equal(t, tt.expected.Command, config.Command)
			assert.Equal(t, tt.expected.Help, config.Help)
			assert.Equal(t, tt.expected.Version, config.Version)
			assert.Equal(t, tt.expected.ConfigFile, config.ConfigFile)
		})
	}
}

func TestCLIValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing subcommand",
			args:    []string{},
			wantErr: true,
			errMsg:  "subcommand required",
		},
		{
			name:    "invalid subcommand",
			args:    []string{"invalid", "--", "echo", "test"},
			wantErr: true,
			errMsg:  "unknown subcommand",
		},
		{
			name:    "missing command after separator",
			args:    []string{"interval", "--every", "1s", "--"},
			wantErr: true,
			errMsg:  "command required after --",
		},
		{
			name:    "invalid duration format",
			args:    []string{"interval", "--every", "invalid", "--", "echo", "test"},
			wantErr: true,
			errMsg:  "invalid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseArgs(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCommandSeparation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "simple command",
			args:     []string{"interval", "--every", "1s", "--", "echo", "hello"},
			expected: []string{"echo", "hello"},
		},
		{
			name:     "command with flags",
			args:     []string{"count", "--times", "3", "--", "curl", "-v", "http://example.com"},
			expected: []string{"curl", "-v", "http://example.com"},
		},
		{
			name:     "complex command with pipes",
			args:     []string{"duration", "--for", "30s", "--", "bash", "-c", "echo hello | grep hello"},
			expected: []string{"bash", "-c", "echo hello | grep hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, config.Command)
		})
	}
}

func TestSubcommandAbbreviations(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedCmd string
		wantErr     bool
	}{
		// Interval abbreviations
		{
			name:        "interval full name",
			args:        []string{"interval", "--every", "30s", "--", "echo", "test"},
			expectedCmd: "interval",
			wantErr:     false,
		},
		{
			name:        "interval primary abbreviation (int)",
			args:        []string{"int", "--every", "30s", "--", "echo", "test"},
			expectedCmd: "interval",
			wantErr:     false,
		},
		{
			name:        "interval minimal abbreviation (i)",
			args:        []string{"i", "--every", "30s", "--", "echo", "test"},
			expectedCmd: "interval",
			wantErr:     false,
		},
		// Count abbreviations
		{
			name:        "count full name",
			args:        []string{"count", "--times", "5", "--", "echo", "test"},
			expectedCmd: "count",
			wantErr:     false,
		},
		{
			name:        "count primary abbreviation (cnt)",
			args:        []string{"cnt", "--times", "5", "--", "echo", "test"},
			expectedCmd: "count",
			wantErr:     false,
		},
		{
			name:        "count minimal abbreviation (c)",
			args:        []string{"c", "--times", "5", "--", "echo", "test"},
			expectedCmd: "count",
			wantErr:     false,
		},
		// Duration abbreviations
		{
			name:        "duration full name",
			args:        []string{"duration", "--for", "1m", "--", "echo", "test"},
			expectedCmd: "duration",
			wantErr:     false,
		},
		{
			name:        "duration primary abbreviation (dur)",
			args:        []string{"dur", "--for", "1m", "--", "echo", "test"},
			expectedCmd: "duration",
			wantErr:     false,
		},
		{
			name:        "duration minimal abbreviation (d)",
			args:        []string{"d", "--for", "1m", "--", "echo", "test"},
			expectedCmd: "duration",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCmd, config.Subcommand)
		})
	}
}

func TestFlagAbbreviations(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "every flag abbreviation (-e)",
			args: []string{"interval", "-e", "30s", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "times flag abbreviation (-t)",
			args: []string{"count", "-t", "10", "--", "echo", "test"},
			expected: Config{
				Subcommand: "count",
				Times:      10,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "for flag abbreviation (-f)",
			args: []string{"duration", "-f", "2m", "--", "echo", "test"},
			expected: Config{
				Subcommand: "duration",
				For:        2 * time.Minute,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "mixed full and abbreviated flags",
			args: []string{"interval", "-e", "1s", "--times", "5", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      1 * time.Second,
				Times:      5,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "all abbreviated flags",
			args: []string{"interval", "-e", "30s", "-t", "3", "-f", "1m", "--", "echo", "test"},
			expected: Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
				Times:      3,
				For:        1 * time.Minute,
				Command:    []string{"echo", "test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.Times, config.Times)
			assert.Equal(t, tt.expected.For, config.For)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

func TestMixedAbbreviations(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "minimal subcommand with abbreviated flags",
			args: []string{"i", "-e", "30s", "-t", "5", "--", "curl", "http://example.com"},
			expected: Config{
				Subcommand: "interval",
				Every:      30 * time.Second,
				Times:      5,
				Command:    []string{"curl", "http://example.com"},
			},
			wantErr: false,
		},
		{
			name: "primary subcommand with mixed flags",
			args: []string{"cnt", "-t", "10", "--every", "2s", "--", "echo", "hello"},
			expected: Config{
				Subcommand: "count",
				Times:      10,
				Every:      2 * time.Second,
				Command:    []string{"echo", "hello"},
			},
			wantErr: false,
		},
		{
			name: "ultra-compact form",
			args: []string{"d", "-f", "5m", "-e", "10s", "--", "date"},
			expected: Config{
				Subcommand: "duration",
				For:        5 * time.Minute,
				Every:      10 * time.Second,
				Command:    []string{"date"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.Every, config.Every)
			assert.Equal(t, tt.expected.Times, config.Times)
			assert.Equal(t, tt.expected.For, config.For)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

// TestAdaptiveSubcommand tests the new adaptive subcommand parsing
func TestAdaptiveSubcommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "adaptive with basic base interval",
			args: []string{"adaptive", "--base-interval", "1s", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "adaptive",
				BaseInterval:     time.Second,
				MinInterval:      100 * time.Millisecond, // default
				MaxInterval:      30 * time.Second,       // default
				SlowThreshold:    2.0,                    // default
				FastThreshold:    0.5,                    // default
				FailureThreshold: 0.3,                    // default
				Command:          []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "adaptive with all parameters",
			args: []string{"adaptive", "--base-interval", "2s", "--min-interval", "500ms",
				"--max-interval", "1m", "--slow-threshold", "3.0", "--fast-threshold", "0.3",
				"--failure-threshold", "0.2", "--show-metrics", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "adaptive",
				BaseInterval:     2 * time.Second,
				MinInterval:      500 * time.Millisecond,
				MaxInterval:      time.Minute,
				SlowThreshold:    3.0,
				FastThreshold:    0.3,
				FailureThreshold: 0.2,
				ShowMetrics:      true,
				Command:          []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "adaptive abbreviations",
			args: []string{"adapt", "-b", "1s", "-m", "--", "echo", "test"},
			expected: Config{
				Subcommand:       "adaptive",
				BaseInterval:     time.Second,
				MinInterval:      100 * time.Millisecond, // default
				MaxInterval:      30 * time.Second,       // default
				SlowThreshold:    2.0,                    // default
				FastThreshold:    0.5,                    // default
				FailureThreshold: 0.3,                    // default
				ShowMetrics:      true,
				Command:          []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "adaptive minimal abbreviation",
			args: []string{"a", "-b", "500ms", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:       "adaptive",
				BaseInterval:     500 * time.Millisecond,
				MinInterval:      100 * time.Millisecond, // default
				MaxInterval:      30 * time.Second,       // default
				SlowThreshold:    2.0,                    // default
				FastThreshold:    0.5,                    // default
				FailureThreshold: 0.3,                    // default
				Command:          []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name:    "adaptive missing base interval",
			args:    []string{"adaptive", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "adaptive invalid min > max",
			args:    []string{"adaptive", "-b", "1s", "--min-interval", "2s", "--max-interval", "1s", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "adaptive invalid threshold",
			args:    []string{"adaptive", "-b", "1s", "--slow-threshold", "0.5", "--", "curl", "api.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				// For error cases, check both parsing and validation
				if err == nil {
					err = ValidateConfig(config)
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, ValidateConfig(config))
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.BaseInterval, config.BaseInterval)
			assert.Equal(t, tt.expected.MinInterval, config.MinInterval)
			assert.Equal(t, tt.expected.MaxInterval, config.MaxInterval)
			assert.Equal(t, tt.expected.SlowThreshold, config.SlowThreshold)
			assert.Equal(t, tt.expected.FastThreshold, config.FastThreshold)
			assert.Equal(t, tt.expected.FailureThreshold, config.FailureThreshold)
			assert.Equal(t, tt.expected.ShowMetrics, config.ShowMetrics)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

// TestBackoffSubcommand tests the new backoff subcommand parsing
func TestBackoffSubcommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "backoff with basic parameters",
			args: []string{"backoff", "--initial", "100ms", "--max", "10s", "--multiplier", "2.0", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:        "backoff",
				InitialInterval:   100 * time.Millisecond,
				BackoffMax:        10 * time.Second,
				BackoffMultiplier: 2.0,
				BackoffJitter:     0.0, // default
				Command:           []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "backoff with all parameters",
			args: []string{"backoff", "--initial", "200ms", "--max", "30s", "--multiplier", "1.5", "--jitter", "0.1", "--", "echo", "test"},
			expected: Config{
				Subcommand:        "backoff",
				InitialInterval:   200 * time.Millisecond,
				BackoffMax:        30 * time.Second,
				BackoffMultiplier: 1.5,
				BackoffJitter:     0.1,
				Command:           []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "backoff abbreviations",
			args: []string{"back", "-i", "500ms", "-x", "5s", "--", "echo", "test"},
			expected: Config{
				Subcommand:        "backoff",
				InitialInterval:   500 * time.Millisecond,
				BackoffMax:        5 * time.Second,
				BackoffMultiplier: 2.0, // default
				BackoffJitter:     0.0, // default
				Command:           []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "backoff minimal abbreviation",
			args: []string{"b", "-i", "1s", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:        "backoff",
				InitialInterval:   time.Second,
				BackoffMax:        30 * time.Second, // default
				BackoffMultiplier: 2.0,              // default
				BackoffJitter:     0.0,              // default
				Command:           []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name:    "backoff missing initial interval",
			args:    []string{"backoff", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "backoff invalid multiplier",
			args:    []string{"backoff", "-i", "1s", "--multiplier", "0.5", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "backoff invalid jitter",
			args:    []string{"backoff", "-i", "1s", "--jitter", "1.5", "--", "curl", "api.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				// For error cases, check both parsing and validation
				if err == nil {
					err = ValidateConfig(config)
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, ValidateConfig(config))
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.InitialInterval, config.InitialInterval)
			assert.Equal(t, tt.expected.BackoffMax, config.BackoffMax)
			assert.Equal(t, tt.expected.BackoffMultiplier, config.BackoffMultiplier)
			assert.Equal(t, tt.expected.BackoffJitter, config.BackoffJitter)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}

// TestLoadAdaptiveSubcommand tests the new load-adaptive subcommand parsing
func TestLoadAdaptiveSubcommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "load-adaptive with basic parameters",
			args: []string{"load-adaptive", "--base-interval", "1s", "--target-cpu", "70", "--target-memory", "80", "--target-load", "1.0", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:   "load-adaptive",
				BaseInterval: time.Second,
				TargetCPU:    70.0,
				TargetMemory: 80.0,
				TargetLoad:   1.0,
				MinInterval:  100 * time.Millisecond, // base/10
				MaxInterval:  10 * time.Second,       // base*10
				Command:      []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name: "load-adaptive with all parameters",
			args: []string{"load-adaptive", "--base-interval", "2s", "--target-cpu", "60", "--target-memory", "70", "--target-load", "0.8", "--min-interval", "500ms", "--max-interval", "30s", "--", "echo", "test"},
			expected: Config{
				Subcommand:   "load-adaptive",
				BaseInterval: 2 * time.Second,
				TargetCPU:    60.0,
				TargetMemory: 70.0,
				TargetLoad:   0.8,
				MinInterval:  500 * time.Millisecond,
				MaxInterval:  30 * time.Second,
				Command:      []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "load-adaptive abbreviations",
			args: []string{"load", "--base-interval", "1s", "--", "echo", "test"},
			expected: Config{
				Subcommand:   "load-adaptive",
				BaseInterval: time.Second,
				TargetCPU:    70.0,                   // default
				TargetMemory: 80.0,                   // default
				TargetLoad:   1.0,                    // default
				MinInterval:  100 * time.Millisecond, // default
				MaxInterval:  10 * time.Second,       // default
				Command:      []string{"echo", "test"},
			},
			wantErr: false,
		},
		{
			name: "load-adaptive minimal abbreviation",
			args: []string{"la", "--base-interval", "500ms", "--", "curl", "api.com"},
			expected: Config{
				Subcommand:   "load-adaptive",
				BaseInterval: 500 * time.Millisecond,
				TargetCPU:    70.0,                  // default
				TargetMemory: 80.0,                  // default
				TargetLoad:   1.0,                   // default
				MinInterval:  50 * time.Millisecond, // default
				MaxInterval:  5 * time.Second,       // default
				Command:      []string{"curl", "api.com"},
			},
			wantErr: false,
		},
		{
			name:    "load-adaptive missing base interval",
			args:    []string{"load-adaptive", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "load-adaptive invalid target cpu",
			args:    []string{"load-adaptive", "--base-interval", "1s", "--target-cpu", "150", "--", "curl", "api.com"},
			wantErr: true,
		},
		{
			name:    "load-adaptive invalid target memory",
			args:    []string{"load-adaptive", "--base-interval", "1s", "--target-memory", "-10", "--", "curl", "api.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseArgs(tt.args)

			if tt.wantErr {
				// For error cases, check both parsing and validation
				if err == nil {
					err = ValidateConfig(config)
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NoError(t, ValidateConfig(config))
			assert.Equal(t, tt.expected.Subcommand, config.Subcommand)
			assert.Equal(t, tt.expected.BaseInterval, config.BaseInterval)
			assert.Equal(t, tt.expected.TargetCPU, config.TargetCPU)
			assert.Equal(t, tt.expected.TargetMemory, config.TargetMemory)
			assert.Equal(t, tt.expected.TargetLoad, config.TargetLoad)
			assert.Equal(t, tt.expected.MinInterval, config.MinInterval)
			assert.Equal(t, tt.expected.MaxInterval, config.MaxInterval)
			assert.Equal(t, tt.expected.Command, config.Command)
		})
	}
}
