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
	if len(args) == 0 {
		return nil, errors.New("subcommand required")
	}

	parser := &argParser{args: args, config: &Config{}}
	return parser.parse()
}

// argParser handles the parsing logic
type argParser struct {
	args   []string
	config *Config
	pos    int
}

// parse orchestrates the parsing process
func (p *argParser) parse() (*Config, error) {
	// Parse global flags first
	if err := p.parseGlobalFlags(); err != nil {
		return nil, err
	}

	// Early return for help/version
	if p.config.Help || p.config.Version {
		return p.config, nil
	}

	// Parse subcommand
	if err := p.parseSubcommand(); err != nil {
		return nil, err
	}

	// Parse subcommand flags
	if err := p.parseSubcommandFlags(); err != nil {
		return nil, err
	}

	// Parse command after --
	if err := p.parseCommand(); err != nil {
		return nil, err
	}

	return p.config, nil
}

// parseGlobalFlags parses global flags like --help, --version, --config
func (p *argParser) parseGlobalFlags() error {
	for p.pos < len(p.args) {
		arg := p.args[p.pos]

		switch arg {
		case "--help", "-h":
			p.config.Help = true
			return nil
		case "--version", "-v":
			p.config.Version = true
			return nil
		case "--config":
			return p.parseConfigFlag()
		default:
			// Not a global flag, continue to subcommand parsing
			return nil
		}
	}
	return nil
}

// parseConfigFlag parses the --config flag and its value
func (p *argParser) parseConfigFlag() error {
	if p.pos+1 >= len(p.args) {
		return errors.New("--config requires a value")
	}
	p.config.ConfigFile = p.args[p.pos+1]
	p.pos += 2
	return nil
}

// parseSubcommand parses and validates the subcommand
func (p *argParser) parseSubcommand() error {
	if p.pos >= len(p.args) {
		return errors.New("subcommand required")
	}

	subcommand := p.args[p.pos]
	switch subcommand {
	case "interval", "count", "duration":
		p.config.Subcommand = subcommand
		p.pos++
		return nil
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// parseSubcommandFlags parses flags specific to subcommands
func (p *argParser) parseSubcommandFlags() error {
	for p.pos < len(p.args) {
		arg := p.args[p.pos]

		if arg == "--" {
			p.pos++ // Skip the separator
			return nil
		}

		switch arg {
		case "--every":
			if err := p.parseDurationFlag(&p.config.Every); err != nil {
				return err
			}
		case "--times":
			if err := p.parseTimesFlag(); err != nil {
				return err
			}
		case "--for":
			if err := p.parseDurationFlag(&p.config.For); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown flag: %s", arg)
		}
	}
	return nil
}

// parseDurationFlag parses a duration flag value
func (p *argParser) parseDurationFlag(target *time.Duration) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	duration, err := time.ParseDuration(p.args[p.pos+1])
	if err != nil {
		return fmt.Errorf("invalid duration: %s", p.args[p.pos+1])
	}

	*target = duration
	p.pos += 2
	return nil
}

// parseTimesFlag parses the --times flag value
func (p *argParser) parseTimesFlag() error {
	if p.pos+1 >= len(p.args) {
		return errors.New("--times requires a value")
	}

	times, err := strconv.ParseInt(p.args[p.pos+1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid times value: %s", p.args[p.pos+1])
	}

	p.config.Times = times
	p.pos += 2
	return nil
}

// parseCommand parses the command after the -- separator
func (p *argParser) parseCommand() error {
	if p.pos >= len(p.args) {
		return errors.New("command required after --")
	}

	p.config.Command = p.args[p.pos:]
	return nil
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
