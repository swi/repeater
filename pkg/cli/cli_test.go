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
