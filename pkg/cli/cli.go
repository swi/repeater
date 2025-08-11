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

	// Adaptive scheduling fields
	BaseInterval     time.Duration // base interval for adaptation
	MinInterval      time.Duration // minimum interval bound
	MaxInterval      time.Duration // maximum interval bound
	SlowThreshold    float64       // threshold for slow response (multiplier)
	FastThreshold    float64       // threshold for fast response (multiplier)
	FailureThreshold float64       // circuit breaker failure threshold
	ShowMetrics      bool          // show adaptive metrics

	// Exponential backoff fields
	InitialInterval   time.Duration // initial backoff interval
	BackoffMax        time.Duration // maximum backoff interval
	BackoffMultiplier float64       // backoff multiplier
	BackoffJitter     float64       // jitter factor (0.0-1.0)

	// Load-aware adaptive fields
	TargetCPU    float64 // target CPU usage percentage (0-100)
	TargetMemory float64 // target memory usage percentage (0-100)
	TargetLoad   float64 // target load average

	// Output control fields
	Stream       bool   // stream command output in real-time
	Quiet        bool   // suppress all output
	Verbose      bool   // show detailed execution information
	StatsOnly    bool   // show only statistics, suppress command output
	OutputPrefix string // prefix for output lines
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
	// Adaptive variations
	case "adaptive", "adapt", "a":
		return "adaptive"
	// Backoff variations
	case "backoff", "back", "b":
		return "backoff"
	// Load-adaptive variations
	case "load-adaptive", "load", "la":
		return "load-adaptive"
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
		case "--base-interval", "-b":
			if err := p.parseDurationFlag(&p.config.BaseInterval); err != nil {
				return err
			}
		case "--min-interval":
			if err := p.parseDurationFlag(&p.config.MinInterval); err != nil {
				return err
			}
		case "--max-interval":
			if err := p.parseDurationFlag(&p.config.MaxInterval); err != nil {
				return err
			}
		case "--slow-threshold":
			if err := p.parseFloatFlag(&p.config.SlowThreshold); err != nil {
				return err
			}
		case "--fast-threshold":
			if err := p.parseFloatFlag(&p.config.FastThreshold); err != nil {
				return err
			}
		case "--failure-threshold":
			if err := p.parseFloatFlag(&p.config.FailureThreshold); err != nil {
				return err
			}
		case "--show-metrics", "-m":
			p.config.ShowMetrics = true
			p.pos++
		case "--initial", "-i":
			if err := p.parseDurationFlag(&p.config.InitialInterval); err != nil {
				return err
			}
		case "--max", "-x":
			if err := p.parseDurationFlag(&p.config.BackoffMax); err != nil {
				return err
			}
		case "--multiplier":
			if err := p.parseFloatFlag(&p.config.BackoffMultiplier); err != nil {
				return err
			}
		case "--jitter":
			if err := p.parseFloatFlag(&p.config.BackoffJitter); err != nil {
				return err
			}
		case "--target-cpu":
			if err := p.parseFloatFlag(&p.config.TargetCPU); err != nil {
				return err
			}
		case "--target-memory":
			if err := p.parseFloatFlag(&p.config.TargetMemory); err != nil {
				return err
			}
		case "--target-load":
			if err := p.parseFloatFlag(&p.config.TargetLoad); err != nil {
				return err
			}
		case "--stream", "-s":
			p.config.Stream = true
			p.pos++
		case "--quiet", "-q":
			p.config.Quiet = true
			p.pos++
		case "--verbose", "-v":
			p.config.Verbose = true
			p.pos++
		case "--stats-only":
			p.config.StatsOnly = true
			p.pos++
		case "--output-prefix", "-o":
			if err := p.parseStringFlag(&p.config.OutputPrefix); err != nil {
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

// parseStringFlag parses a string flag value
func (p *argParser) parseStringFlag(target *string) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	*target = p.args[p.pos+1]
	p.pos += 2
	return nil
}

// parseFloatFlag parses a float flag value
func (p *argParser) parseFloatFlag(target *float64) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value, err := strconv.ParseFloat(p.args[p.pos+1], 64)
	if err != nil {
		return fmt.Errorf("invalid float value: %s", p.args[p.pos+1])
	}

	*target = value
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

	// Validate output control flags
	if err := validateOutputFlags(config); err != nil {
		return err
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
	case "adaptive":
		if config.BaseInterval == 0 {
			return errors.New("--base-interval is required for adaptive subcommand")
		}
		// Validate adaptive configuration
		if err := validateAdaptiveConfig(config); err != nil {
			return fmt.Errorf("invalid adaptive config: %w", err)
		}
	case "backoff":
		if config.InitialInterval == 0 {
			return errors.New("--initial is required for backoff subcommand")
		}
		// Validate backoff configuration
		if err := validateBackoffConfig(config); err != nil {
			return fmt.Errorf("invalid backoff config: %w", err)
		}
	case "load-adaptive":
		if config.BaseInterval == 0 {
			return errors.New("--base-interval is required for load-adaptive subcommand")
		}
		// Validate load-adaptive configuration
		if err := validateLoadAdaptiveConfig(config); err != nil {
			return fmt.Errorf("invalid load-adaptive config: %w", err)
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

// validateAdaptiveConfig validates the adaptive configuration
func validateAdaptiveConfig(config *Config) error {
	// Set defaults if not provided
	if config.MinInterval == 0 {
		config.MinInterval = 100 * time.Millisecond
	}
	if config.MaxInterval == 0 {
		config.MaxInterval = 30 * time.Second
	}
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 2.0
	}
	if config.FastThreshold == 0 {
		config.FastThreshold = 0.5
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 0.3
	}

	// Validate bounds
	if config.MinInterval >= config.MaxInterval {
		return errors.New("min-interval must be less than max-interval")
	}

	if config.BaseInterval < config.MinInterval || config.BaseInterval > config.MaxInterval {
		return errors.New("base-interval must be between min-interval and max-interval")
	}

	if config.SlowThreshold <= 1.0 {
		return errors.New("slow-threshold must be greater than 1.0")
	}

	if config.FastThreshold <= 0 || config.FastThreshold >= 1.0 {
		return errors.New("fast-threshold must be between 0 and 1.0")
	}

	if config.FailureThreshold <= 0 || config.FailureThreshold >= 1.0 {
		return errors.New("failure-threshold must be between 0 and 1.0")
	}

	return nil
}

// validateBackoffConfig validates the backoff configuration
func validateBackoffConfig(config *Config) error {
	// Set defaults if not provided
	if config.BackoffMax == 0 {
		config.BackoffMax = 30 * time.Second
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2.0
	}
	if config.BackoffJitter < 0 {
		config.BackoffJitter = 0.0
	}

	// Validate bounds
	if config.InitialInterval >= config.BackoffMax {
		return errors.New("initial interval must be less than max interval")
	}

	if config.BackoffMultiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.BackoffJitter < 0 || config.BackoffJitter > 1.0 {
		return errors.New("jitter must be between 0.0 and 1.0")
	}

	return nil
}

// validateLoadAdaptiveConfig validates the load-adaptive configuration
func validateLoadAdaptiveConfig(config *Config) error {
	// Set defaults if not provided
	if config.TargetCPU == 0 {
		config.TargetCPU = 70.0 // Default 70% CPU target
	}
	if config.TargetMemory == 0 {
		config.TargetMemory = 80.0 // Default 80% memory target
	}
	if config.TargetLoad == 0 {
		config.TargetLoad = 1.0 // Default load average of 1.0
	}
	if config.MinInterval == 0 {
		config.MinInterval = config.BaseInterval / 10
	}
	if config.MaxInterval == 0 {
		config.MaxInterval = config.BaseInterval * 10
	}

	// Validate bounds
	if config.TargetCPU <= 0 || config.TargetCPU > 100 {
		return errors.New("target-cpu must be between 0 and 100")
	}

	if config.TargetMemory <= 0 || config.TargetMemory > 100 {
		return errors.New("target-memory must be between 0 and 100")
	}

	if config.TargetLoad <= 0 {
		return errors.New("target-load must be greater than 0")
	}

	if config.MinInterval >= config.MaxInterval {
		return errors.New("min-interval must be less than max-interval")
	}

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

// validateOutputFlags validates output control flags for conflicts
func validateOutputFlags(config *Config) error {
	// Check for conflicting flags
	if config.Quiet && config.Stream {
		return errors.New("--quiet and --stream flags are mutually exclusive")
	}

	if config.Quiet && config.Verbose {
		return errors.New("--quiet and --verbose flags are mutually exclusive")
	}

	if config.StatsOnly && config.Stream {
		return errors.New("--stats-only and --stream flags are mutually exclusive")
	}

	if config.StatsOnly && config.Verbose {
		return errors.New("--stats-only and --verbose flags are mutually exclusive")
	}

	if config.StatsOnly && config.Quiet {
		return errors.New("--stats-only and --quiet flags are mutually exclusive")
	}

	// Note: --stream and --verbose can be used together for detailed streaming

	return nil
}
