package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Config represents the parsed CLI configuration
type Config struct {
	Subcommand string
	Every      time.Duration
	Times      int64
	For        time.Duration
	Command    []string
	Help       bool
	Version    bool
	ConfigFile string
}

// ParseArgs parses command line arguments and returns a Config
func ParseArgs(args []string) (*Config, error) {
	config := &Config{}

	if len(args) == 0 {
		return nil, errors.New("subcommand required")
	}

	i := 0

	// Parse global flags first
	for i < len(args) {
		arg := args[i]

		if arg == "--help" || arg == "-h" {
			config.Help = true
			return config, nil
		}

		if arg == "--version" || arg == "-v" {
			config.Version = true
			return config, nil
		}

		if arg == "--config" {
			if i+1 >= len(args) {
				return nil, errors.New("--config requires a value")
			}
			config.ConfigFile = args[i+1]
			i += 2
			continue
		}

		// If not a global flag, break to parse subcommand
		break
	}

	if i >= len(args) {
		return nil, errors.New("subcommand required")
	}

	// Parse subcommand
	subcommand := args[i]
	switch subcommand {
	case "interval", "count", "duration":
		config.Subcommand = subcommand
	default:
		return nil, fmt.Errorf("unknown subcommand: %s", subcommand)
	}
	i++

	// Parse subcommand flags
	var commandStart int = -1
	for i < len(args) {
		arg := args[i]

		if arg == "--" {
			commandStart = i + 1
			break
		}

		switch arg {
		case "--every":
			if i+1 >= len(args) {
				return nil, errors.New("--every requires a value")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration: %s", args[i+1])
			}
			config.Every = duration
			i += 2

		case "--times":
			if i+1 >= len(args) {
				return nil, errors.New("--times requires a value")
			}
			times, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid times value: %s", args[i+1])
			}
			config.Times = times
			i += 2

		case "--for":
			if i+1 >= len(args) {
				return nil, errors.New("--for requires a value")
			}
			duration, err := time.ParseDuration(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration: %s", args[i+1])
			}
			config.For = duration
			i += 2

		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	// Parse command after --
	if commandStart == -1 {
		return nil, errors.New("command required after --")
	}

	if commandStart >= len(args) {
		return nil, errors.New("command required after --")
	}

	config.Command = args[commandStart:]

	return config, nil
}

// ValidateConfig validates the parsed configuration
func ValidateConfig(config *Config) error {
	if config.Help || config.Version {
		return nil // Help and version don't need validation
	}

	if config.Subcommand == "" {
		return errors.New("subcommand required")
	}

	if len(config.Command) == 0 {
		return errors.New("command required after --")
	}

	// Validate subcommand-specific requirements
	switch config.Subcommand {
	case "interval":
		if config.Every == 0 {
			return errors.New("--every is required for interval subcommand")
		}
	case "count":
		if config.Times == 0 {
			return errors.New("--times is required for count subcommand")
		}
	case "duration":
		if config.For == 0 {
			return errors.New("--for is required for duration subcommand")
		}
	}

	return nil
}
