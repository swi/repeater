package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swi/repeater/pkg/cli"
)

// TestConfigFileIntegration tests end-to-end config file integration
func TestConfigFileIntegration(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		args           []string
		expectedConfig func(t *testing.T, config *cli.Config)
		expectError    bool
	}{
		{
			name: "config file with default timeout should be applied",
			configContent: `
[defaults]
timeout = "60s"
max_retries = 5
log_level = "debug"

[observability]
metrics_enabled = true
metrics_port = 9091
health_enabled = true
health_check_port = 8081
`,
			args: []string{"--config", "CONFIG_FILE", "interval", "--every", "1s", "--times", "3", "--", "echo", "test"},
			expectedConfig: func(t *testing.T, config *cli.Config) {
				// These should come from config file
				assert.Equal(t, 60*time.Second, config.Timeout)
				assert.Equal(t, 5, config.MaxRetries)
				assert.Equal(t, "debug", config.LogLevel)
				assert.True(t, config.MetricsEnabled)
				assert.Equal(t, 9091, config.MetricsPort)
				assert.True(t, config.HealthEnabled)
				assert.Equal(t, 8081, config.HealthPort)

				// These should come from CLI args
				assert.Equal(t, "interval", config.Subcommand)
				assert.Equal(t, time.Second, config.Every)
				assert.Equal(t, int64(3), config.Times)
				assert.Equal(t, []string{"echo", "test"}, config.Command)
			},
		},
		{
			name: "environment variables should override config file",
			configContent: `
[defaults]
timeout = "30s"
log_level = "info"
`,
			args: []string{"--config", "CONFIG_FILE", "count", "--times", "2", "--", "echo", "test"},
			expectedConfig: func(t *testing.T, config *cli.Config) {
				// Environment should override config file
				assert.Equal(t, 90*time.Second, config.Timeout)
				assert.Equal(t, "error", config.LogLevel)
			},
		}, {
			name: "invalid config file should return error",
			configContent: `
[defaults
timeout = "30s"
invalid syntax
`,
			args:        []string{"--config", "CONFIG_FILE", "interval", "--every", "1s", "--", "echo", "test"},
			expectError: true,
		},
		{
			name:        "nonexistent config file should return error",
			args:        []string{"--config", "/nonexistent/config.toml", "interval", "--every", "1s", "--", "echo", "test"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the environment override test
			if tt.name == "environment variables should override config file" {
				t.Setenv("RPR_TIMEOUT", "90s")
				t.Setenv("RPR_LOG_LEVEL", "error")
			}

			var configFile string

			if tt.configContent != "" {
				// Create temporary config file
				tmpDir := t.TempDir()
				configFile = filepath.Join(tmpDir, "test-config.toml")
				err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
				require.NoError(t, err)
			}

			// Replace CONFIG_FILE placeholder with actual path
			args := make([]string, len(tt.args))
			for i, arg := range tt.args {
				if arg == "CONFIG_FILE" {
					args[i] = configFile
				} else {
					args[i] = arg
				}
			}

			// Parse CLI args
			config, err := cli.ParseArgs(args)
			require.NoError(t, err, "CLI parsing should not fail")

			// Load and apply config file
			err = applyConfigFile(config)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Validate final configuration
			if tt.expectedConfig != nil {
				tt.expectedConfig(t, config)
			}
		})
	}
}

// TestConfigFileWithRunner tests that config file settings are applied to runner
func TestConfigFileWithRunner(t *testing.T) {
	// Create config file with observability settings
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "runner-config.toml")

	configContent := `
[defaults]
timeout = "45s"

[observability]
metrics_enabled = true
metrics_port = 9092
health_enabled = true
health_check_port = 8082
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Parse CLI with config file
	args := []string{"--config", configFile, "count", "--times", "1", "--", "echo", "test"}
	config, err := cli.ParseArgs(args)
	require.NoError(t, err)

	// Apply config file
	err = applyConfigFile(config)
	require.NoError(t, err)

	// Verify config was applied
	assert.Equal(t, 45*time.Second, config.Timeout)
	assert.True(t, config.MetricsEnabled)
	assert.Equal(t, 9092, config.MetricsPort)
	assert.True(t, config.HealthEnabled)
	assert.Equal(t, 8082, config.HealthPort)

	// Test that runner can be created with config (this will fail until we implement it)
	_, err = createRunnerWithConfig(config)
	require.NoError(t, err, "Runner should be created successfully with config file settings")
}

// createRunnerWithConfig creates a runner with config file settings applied
// This function doesn't exist yet - we need to implement it
func createRunnerWithConfig(config *cli.Config) (interface{}, error) {
	// TODO: Implement this function
	// This should create a runner that respects config file settings like:
	// - Timeout
	// - Metrics enabled/port
	// - Health check enabled/port
	return nil, nil
}
