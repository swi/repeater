package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the complete configuration structure
type Config struct {
	Defaults      DefaultsConfig      `toml:"defaults"`
	Scheduling    SchedulingConfig    `toml:"scheduling"`
	Observability ObservabilityConfig `toml:"observability"`
}

// DefaultsConfig contains default execution parameters
type DefaultsConfig struct {
	Timeout    time.Duration `toml:"timeout"`
	MaxRetries int           `toml:"max_retries"`
	LogLevel   string        `toml:"log_level"`
}

// SchedulingConfig contains scheduling-related configuration
type SchedulingConfig struct {
	DefaultInterval time.Duration `toml:"default_interval"`
	JitterPercent   float64       `toml:"jitter_percent"`
}

// ObservabilityConfig contains monitoring and metrics configuration
type ObservabilityConfig struct {
	MetricsEnabled  bool `toml:"metrics_enabled"`
	MetricsPort     int  `toml:"metrics_port"`
	HealthCheckPort int  `toml:"health_check_port"`
	HealthEnabled   bool `toml:"health_enabled"`
}

// LoadConfig loads configuration from TOML file with environment variable overrides
func LoadConfig(configPath string) (*Config, error) {
	// Start with default configuration
	config := &Config{
		Defaults: DefaultsConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			LogLevel:   "info",
		},
		Scheduling: SchedulingConfig{
			DefaultInterval: 10 * time.Second,
			JitterPercent:   0.0,
		},
		Observability: ObservabilityConfig{
			MetricsEnabled:  false,
			MetricsPort:     9090,
			HealthCheckPort: 8080,
			HealthEnabled:   false,
		},
	}

	// Load from TOML file if provided and exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err != nil {
			return nil, fmt.Errorf("config file not found: %w", err)
		}

		if _, err := toml.DecodeFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML config: %w", err)
		}
	}

	// Apply environment variable overrides
	if err := applyEnvironmentOverrides(config); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// applyEnvironmentOverrides applies environment variable overrides to config
func applyEnvironmentOverrides(config *Config) error {
	// Defaults section
	if val := os.Getenv("RPR_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_TIMEOUT: %w", err)
		}
		config.Defaults.Timeout = duration
	}

	if val := os.Getenv("RPR_MAX_RETRIES"); val != "" {
		retries, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_MAX_RETRIES: %w", err)
		}
		config.Defaults.MaxRetries = retries
	}

	if val := os.Getenv("RPR_LOG_LEVEL"); val != "" {
		config.Defaults.LogLevel = val
	}

	// Scheduling section
	if val := os.Getenv("RPR_DEFAULT_INTERVAL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_DEFAULT_INTERVAL: %w", err)
		}
		config.Scheduling.DefaultInterval = duration
	}

	if val := os.Getenv("RPR_JITTER_PERCENT"); val != "" {
		jitter, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid RPR_JITTER_PERCENT: %w", err)
		}
		config.Scheduling.JitterPercent = jitter
	}

	// Observability section
	if val := os.Getenv("RPR_METRICS_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_METRICS_ENABLED: %w", err)
		}
		config.Observability.MetricsEnabled = enabled
	}

	if val := os.Getenv("RPR_METRICS_PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_METRICS_PORT: %w", err)
		}
		config.Observability.MetricsPort = port
	}

	if val := os.Getenv("RPR_HEALTH_CHECK_PORT"); val != "" {
		port, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_HEALTH_CHECK_PORT: %w", err)
		}
		config.Observability.HealthCheckPort = port
	}

	if val := os.Getenv("RPR_HEALTH_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid RPR_HEALTH_ENABLED: %w", err)
		}
		config.Observability.HealthEnabled = enabled
	}

	return nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if strings.ToLower(c.Defaults.LogLevel) == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level '%s', must be one of: %v", c.Defaults.LogLevel, validLogLevels)
	}

	// Validate timeout
	if c.Defaults.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative: %v", c.Defaults.Timeout)
	}

	// Validate max retries
	if c.Defaults.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative: %d", c.Defaults.MaxRetries)
	}

	// Validate default interval
	if c.Scheduling.DefaultInterval < 0 {
		return fmt.Errorf("default_interval cannot be negative: %v", c.Scheduling.DefaultInterval)
	}

	// Validate jitter percent
	if c.Scheduling.JitterPercent < 0 || c.Scheduling.JitterPercent > 100 {
		return fmt.Errorf("jitter_percent must be between 0 and 100: %f", c.Scheduling.JitterPercent)
	}

	// Validate port numbers
	if c.Observability.MetricsPort < 1 || c.Observability.MetricsPort > 65535 {
		return fmt.Errorf("metrics_port must be between 1 and 65535: %d", c.Observability.MetricsPort)
	}

	if c.Observability.HealthCheckPort < 1 || c.Observability.HealthCheckPort > 65535 {
		return fmt.Errorf("health_check_port must be between 1 and 65535: %d", c.Observability.HealthCheckPort)
	}

	return nil
}
