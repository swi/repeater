package main

import (
	"github.com/swi/repeater/pkg/cli"
	configpkg "github.com/swi/repeater/pkg/config"
)

// applyConfigFile loads and applies config file settings to CLI config
func applyConfigFile(config *cli.Config) error {
	if config.ConfigFile == "" {
		return nil // No config file specified
	}

	// Load config file using pkg/config
	fileConfig, err := configpkg.LoadConfig(config.ConfigFile)
	if err != nil {
		return err
	}

	// Apply config file settings to CLI config
	config.Timeout = fileConfig.Defaults.Timeout
	config.MaxRetries = fileConfig.Defaults.MaxRetries
	config.LogLevel = fileConfig.Defaults.LogLevel
	config.MetricsEnabled = fileConfig.Observability.MetricsEnabled
	config.MetricsPort = fileConfig.Observability.MetricsPort
	config.HealthEnabled = fileConfig.Observability.HealthEnabled
	config.HealthPort = fileConfig.Observability.HealthCheckPort

	return nil
}
