package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
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

	// Rate limiting fields
	RateSpec     string // e.g., "10/1h", "100/1m"
	RetryPattern string // e.g., "0,10m,30m"
	ShowNext     bool   // show next allowed time
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
	normalizedSubcommand := normalizeSubcommand(subcommand)
	if normalizedSubcommand == "" {
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}

	p.config.Subcommand = normalizedSubcommand
	p.pos++
	return nil
}

// normalizeSubcommand converts abbreviations to full subcommand names
func normalizeSubcommand(cmd string) string {
	switch cmd {
	// Interval variations
	case "interval", "int", "i":
		return "interval"
	// Count variations
	case "count", "cnt", "c":
		return "count"
	// Duration variations
	case "duration", "dur", "d":
		return "duration"
	// Rate limit variations
	case "rate-limit", "rate", "rl", "r":
		return "rate-limit"
	default:
		return ""
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
		case "--every", "-e":
			if err := p.parseDurationFlag(&p.config.Every); err != nil {
				return err
			}
		case "--times", "-t":
			if err := p.parseTimesFlag(); err != nil {
				return err
			}
		case "--for", "-f":
			if err := p.parseDurationFlag(&p.config.For); err != nil {
				return err
			}
		case "--rate", "-r":
			if err := p.parseStringFlag(&p.config.RateSpec); err != nil {
				return err
			}
		case "--retry-pattern", "-p":
			if err := p.parseStringFlag(&p.config.RetryPattern); err != nil {
				return err
			}
		case "--show-next", "-n":
			p.config.ShowNext = true
			p.pos++
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

// parseStringFlag parses a string flag value
func (p *argParser) parseStringFlag(target *string) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	*target = p.args[p.pos+1]
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
	case "rate-limit":
		if config.RateSpec == "" {
			return errors.New("--rate is required for rate-limit subcommand")
		}
		// Validate rate spec format
		if err := validateRateSpec(config.RateSpec); err != nil {
			return fmt.Errorf("invalid rate spec: %w", err)
		}
	}

	return nil
}

// validateRateSpec validates the rate specification format
func validateRateSpec(spec string) error {
	// Use the ParseRateSpec function from ratelimit package to validate
	// For now, do basic validation here to avoid circular imports
	if spec == "" {
		return errors.New("rate spec cannot be empty")
	}

	// Basic format check: should contain "/"
	if !strings.Contains(spec, "/") {
		return errors.New("rate spec must be in format 'rate/period' (e.g., '10/1h')")
	}

	parts := strings.Split(spec, "/")
	if len(parts) != 2 {
		return errors.New("rate spec must be in format 'rate/period' (e.g., '10/1h')")
	}

	// Validate rate part is a number
	if _, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err != nil {
		return fmt.Errorf("invalid rate number: %s", parts[0])
	}

	// Validate period part is a valid duration
	if _, err := time.ParseDuration(strings.TrimSpace(parts[1])); err != nil {
		return fmt.Errorf("invalid period duration: %s", parts[1])
	}

	return nil
}
