package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
