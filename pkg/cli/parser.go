package cli

import (
	"errors"
	"fmt"
)

// ParseArgs parses command line arguments and returns a Config
func ParseArgs(args []string) (*Config, error) {
	if len(args) == 0 {
		return nil, errors.New("subcommand required")
	}

	parser := &argParser{args: args, config: &Config{
		// Set HTTP-aware defaults
		HTTPParseJSON:    true,  // Enable JSON parsing by default
		HTTPParseHeaders: true,  // Enable header parsing by default
		HTTPTrustClient:  false, // Don't trust client errors by default
	}}
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

	// Apply configuration file defaults if specified
	if err := p.applyConfigDefaults(); err != nil {
		return nil, err
	}

	// Validate the final configuration
	if err := ValidateConfig(p.config); err != nil {
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
	// NEW RETRY STRATEGIES (Mathematical Algorithms)
	case "exponential", "exp":
		return "exponential"
	case "fibonacci", "fib":
		return "fibonacci"
	case "linear", "lin":
		return "linear"
	case "polynomial", "poly":
		return "polynomial"
	case "decorrelated-jitter", "dj":
		return "decorrelated-jitter"

	// EXISTING EXECUTION MODES (Operational Patterns)
	case "interval", "int", "i":
		return "interval"
	case "count", "cnt", "c":
		return "count"
	case "duration", "dur", "d":
		return "duration"
	case "cron", "cr":
		return "cron"
	case "adaptive", "adapt", "a":
		return "adaptive"

	// EXISTING RATE CONTROL (Resource Management)
	case "rate-limit", "rate", "rl", "r":
		return "rate-limit"
	case "load-adaptive", "load", "la":
		return "load-adaptive"
	default:
		return ""
	}
}

// parseCommand parses the command after the -- separator
func (p *argParser) parseCommand() error {
	if p.pos >= len(p.args) {
		return errors.New("command required after --")
	}

	p.config.Command = p.args[p.pos:]
	return nil
}

// applyConfigDefaults applies configuration file defaults to CLI config
func (p *argParser) applyConfigDefaults() error {
	if p.config.ConfigFile == "" {
		return nil // No config file specified
	}

	// For now, just store the config file path for later use
	// The actual loading and application of defaults will be done
	// by the runner when it needs the configuration
	// This allows the CLI parsing to succeed without requiring
	// the config file to exist at parse time

	return nil
}
