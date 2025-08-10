package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLoad_FromTOMLFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "repeater.toml")

	tomlContent := `
[defaults]
timeout = "30s"
max_retries = 3
log_level = "info"

[scheduling]
default_interval = "10s"
jitter_percent = 5.0

[observability]
metrics_enabled = true
metrics_port = 9090
health_check_port = 8080
`

	err := os.WriteFile(configFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading configuration
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	// Verify loaded values
	if config.Defaults.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Defaults.Timeout)
	}

	if config.Defaults.MaxRetries != 3 {
		t.Errorf("Expected max_retries 3, got %d", config.Defaults.MaxRetries)
	}

	if config.Defaults.LogLevel != "info" {
		t.Errorf("Expected log_level 'info', got %s", config.Defaults.LogLevel)
	}

	if config.Scheduling.DefaultInterval != 10*time.Second {
		t.Errorf("Expected default_interval 10s, got %v", config.Scheduling.DefaultInterval)
	}

	if config.Scheduling.JitterPercent != 5.0 {
		t.Errorf("Expected jitter_percent 5.0, got %f", config.Scheduling.JitterPercent)
	}

	if !config.Observability.MetricsEnabled {
		t.Error("Expected metrics_enabled true, got false")
	}

	if config.Observability.MetricsPort != 9090 {
		t.Errorf("Expected metrics_port 9090, got %d", config.Observability.MetricsPort)
	}

	if config.Observability.HealthCheckPort != 8080 {
		t.Errorf("Expected health_check_port 8080, got %d", config.Observability.HealthCheckPort)
	}
}

func TestConfigLoad_WithEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("RPR_TIMEOUT", "60s")
	_ = os.Setenv("RPR_MAX_RETRIES", "5")
	_ = os.Setenv("RPR_LOG_LEVEL", "debug")
	_ = os.Setenv("RPR_METRICS_PORT", "9091")
	defer func() {
		_ = os.Unsetenv("RPR_TIMEOUT")
		_ = os.Unsetenv("RPR_MAX_RETRIES")
		_ = os.Unsetenv("RPR_LOG_LEVEL")
		_ = os.Unsetenv("RPR_METRICS_PORT")
	}()

	// Create config file with different values
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "repeater.toml")

	tomlContent := `
[defaults]
timeout = "30s"
max_retries = 3
log_level = "info"

[observability]
metrics_port = 9090
`

	err := os.WriteFile(configFile, []byte(tomlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load config - environment should override file values
	config, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	// Verify environment overrides took effect
	if config.Defaults.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s (from env), got %v", config.Defaults.Timeout)
	}

	if config.Defaults.MaxRetries != 5 {
		t.Errorf("Expected max_retries 5 (from env), got %d", config.Defaults.MaxRetries)
	}

	if config.Defaults.LogLevel != "debug" {
		t.Errorf("Expected log_level 'debug' (from env), got %s", config.Defaults.LogLevel)
	}

	if config.Observability.MetricsPort != 9091 {
		t.Errorf("Expected metrics_port 9091 (from env), got %d", config.Observability.MetricsPort)
	}
}

func TestConfigLoad_DefaultValues(t *testing.T) {
	// Test loading with no config file - should use defaults
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Expected no error with empty config path, got: %v", err)
	}

	// Verify default values
	if config.Defaults.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", config.Defaults.Timeout)
	}

	if config.Defaults.MaxRetries != 3 {
		t.Errorf("Expected default max_retries 3, got %d", config.Defaults.MaxRetries)
	}

	if config.Defaults.LogLevel != "info" {
		t.Errorf("Expected default log_level 'info', got %s", config.Defaults.LogLevel)
	}

	if config.Scheduling.DefaultInterval != 10*time.Second {
		t.Errorf("Expected default interval 10s, got %v", config.Scheduling.DefaultInterval)
	}

	if config.Scheduling.JitterPercent != 0.0 {
		t.Errorf("Expected default jitter_percent 0.0, got %f", config.Scheduling.JitterPercent)
	}

	if config.Observability.MetricsEnabled {
		t.Error("Expected default metrics_enabled false, got true")
	}

	if config.Observability.MetricsPort != 9090 {
		t.Errorf("Expected default metrics_port 9090, got %d", config.Observability.MetricsPort)
	}

	if config.Observability.HealthCheckPort != 8080 {
		t.Errorf("Expected default health_check_port 8080, got %d", config.Observability.HealthCheckPort)
	}
}

func TestConfigLoad_InvalidTOMLFile(t *testing.T) {
	// Create invalid TOML file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.toml")

	invalidContent := `
[defaults
timeout = "30s"
invalid syntax here
`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Should return error for invalid TOML
	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("Expected error for invalid TOML file, got nil")
	}
}

func TestConfigLoad_NonexistentFile(t *testing.T) {
	// Should return error for nonexistent file
	_, err := LoadConfig("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent config file, got nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				Defaults: DefaultsConfig{
					Timeout:    30 * time.Second,
					MaxRetries: 3,
					LogLevel:   "info",
				},
				Observability: ObservabilityConfig{
					MetricsPort:     9090,
					HealthCheckPort: 8080,
				},
			},
			expectError: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Defaults: DefaultsConfig{
					LogLevel: "invalid",
				},
			},
			expectError: true,
		},
		{
			name: "invalid port numbers",
			config: Config{
				Observability: ObservabilityConfig{
					MetricsPort:     -1,
					HealthCheckPort: 70000,
				},
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			config: Config{
				Defaults: DefaultsConfig{
					Timeout: -1 * time.Second,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}
