package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/swi/repeater/pkg/httpaware"
	"github.com/swi/repeater/pkg/patterns"
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

	// Exponential backoff fields (legacy - for backward compatibility)
	InitialInterval   time.Duration // initial backoff interval (deprecated: use BaseDelay)
	BackoffMax        time.Duration // maximum backoff interval (deprecated: use MaxDelay)
	BackoffMultiplier float64       // backoff multiplier (deprecated: use Multiplier)
	BackoffJitter     float64       // jitter factor (0.0-1.0)

	// New unified strategy fields
	BaseDelay  time.Duration // base/initial delay for all strategies
	Increment  time.Duration // linear increment (linear strategy)
	Multiplier float64       // exponential/jitter multiplier
	Exponent   float64       // polynomial exponent
	MaxDelay   time.Duration // maximum delay cap for all strategies

	// Load-aware adaptive fields
	TargetCPU    float64 // target CPU usage percentage (0-100)
	TargetMemory float64 // target memory usage percentage (0-100)
	TargetLoad   float64 // target load average

	// Cron scheduling fields
	CronExpression string // cron expression for scheduling
	Timezone       string // timezone for cron scheduling

	// Output control fields
	Stream       bool   // stream command output in real-time
	Quiet        bool   // suppress all output
	Verbose      bool   // show detailed execution information
	StatsOnly    bool   // show only statistics, suppress command output
	OutputPrefix string // prefix for output lines

	// Pattern matching fields
	SuccessPattern  string // regex pattern indicating success in output
	FailurePattern  string // regex pattern indicating failure in output
	CaseInsensitive bool   // make pattern matching case-insensitive

	// HTTP-aware scheduling fields
	HTTPAware        bool          // enable HTTP-aware intelligent scheduling
	HTTPMaxDelay     time.Duration // maximum delay cap for HTTP timing
	HTTPMinDelay     time.Duration // minimum delay floor for HTTP timing
	HTTPParseJSON    bool          // parse JSON response bodies for timing
	HTTPParseHeaders bool          // parse HTTP headers for timing
	HTTPTrustClient  bool          // trust 4xx client error timing
	HTTPCustomFields []string      // custom JSON fields to check for timing

	// Config file fields (loaded from TOML)
	Timeout        time.Duration // command execution timeout
	MaxRetries     int           // maximum retry attempts
	LogLevel       string        // logging level
	MetricsEnabled bool          // enable metrics collection
	MetricsPort    int           // metrics server port
	HealthEnabled  bool          // enable health check endpoint
	HealthPort     int           // health check server port
}

// GetPatternConfig returns a patterns.PatternConfig from the CLI config
func (c *Config) GetPatternConfig() *patterns.PatternConfig {
	if c.SuccessPattern == "" && c.FailurePattern == "" {
		return nil
	}

	return &patterns.PatternConfig{
		SuccessPattern:  c.SuccessPattern,
		FailurePattern:  c.FailurePattern,
		CaseInsensitive: c.CaseInsensitive,
	}
}

// GetHTTPAwareConfig returns HTTP-aware configuration from the CLI config
func (c *Config) GetHTTPAwareConfig() *httpaware.HTTPAwareConfig {
	if !c.HTTPAware {
		return nil
	}

	// Set defaults
	config := &httpaware.HTTPAwareConfig{
		MaxDelay:          c.HTTPMaxDelay,
		MinDelay:          c.HTTPMinDelay,
		ParseJSON:         c.HTTPParseJSON,
		ParseHeaders:      c.HTTPParseHeaders,
		TrustClientErrors: c.HTTPTrustClient,
		JSONFields:        c.HTTPCustomFields,
	}

	// Apply defaults if not set
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Minute // Default max delay
	}
	if config.MinDelay == 0 {
		config.MinDelay = 1 * time.Second // Default min delay
	}
	if config.JSONFields == nil {
		config.JSONFields = []string{"retry_after", "retryAfter"} // Default fields
	}

	return config
}

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

	// LEGACY SUPPORT (Backward Compatibility)
	case "backoff", "back", "b":
		return "backoff"
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
		case "--initial-delay", "-i":
			if err := p.parseDurationFlag(&p.config.InitialInterval); err != nil {
				return err
			}
		case "--max", "-x":
			if err := p.parseDurationFlag(&p.config.BackoffMax); err != nil {
				return err
			}
		case "--multiplier":
			if err := p.parseFloatFlag(&p.config.Multiplier); err != nil {
				return err
			}
			// Also set legacy field for backward compatibility
			p.config.BackoffMultiplier = p.config.Multiplier
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
		case "--cron":
			if err := p.parseStringFlag(&p.config.CronExpression); err != nil {
				return err
			}
		case "--timezone", "--tz":
			if err := p.parseStringFlag(&p.config.Timezone); err != nil {
				return err
			}
		case "--success-pattern":
			if err := p.parseStringFlag(&p.config.SuccessPattern); err != nil {
				return err
			}
		case "--failure-pattern":
			if err := p.parseStringFlag(&p.config.FailurePattern); err != nil {
				return err
			}
		case "--case-insensitive":
			p.config.CaseInsensitive = true
			p.pos++
		case "--http-aware":
			p.config.HTTPAware = true
			p.pos++
		case "--http-max-delay":
			if err := p.parseDurationFlag(&p.config.HTTPMaxDelay); err != nil {
				return err
			}
		case "--http-min-delay":
			if err := p.parseDurationFlag(&p.config.HTTPMinDelay); err != nil {
				return err
			}
		case "--http-parse-json":
			p.config.HTTPParseJSON = true
			p.pos++
		case "--http-no-parse-json":
			p.config.HTTPParseJSON = false
			p.pos++
		case "--http-parse-headers":
			p.config.HTTPParseHeaders = true
			p.pos++
		case "--http-no-parse-headers":
			p.config.HTTPParseHeaders = false
			p.pos++
		case "--http-trust-client":
			p.config.HTTPTrustClient = true
			p.pos++
		case "--http-custom-fields":
			if err := p.parseStringSliceFlag(&p.config.HTTPCustomFields); err != nil {
				return err
			}
		case "--attempts", "-a":
			if err := p.parseIntFlag(&p.config.MaxRetries); err != nil {
				return err
			}

		// NEW STRATEGY PARAMETERS
		case "--base-delay", "-bd":
			if err := p.parseDurationFlag(&p.config.BaseDelay); err != nil {
				return err
			}
		case "--increment", "-inc":
			if err := p.parseDurationFlag(&p.config.Increment); err != nil {
				return err
			}
		case "--exponent", "-exp":
			if err := p.parseFloatFlag(&p.config.Exponent); err != nil {
				return err
			}
		case "--max-delay", "-md":
			if err := p.parseDurationFlag(&p.config.MaxDelay); err != nil {
				return err
			}
		// Note: --multiplier already exists at line 356, update it to use new field
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

// parseIntFlag parses an integer flag value
func (p *argParser) parseIntFlag(target *int) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value, err := strconv.Atoi(p.args[p.pos+1])
	if err != nil {
		return fmt.Errorf("invalid integer value: %s", p.args[p.pos+1])
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

	// Validate pattern matching configuration
	if err := validatePatterns(config); err != nil {
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
			return errors.New("--initial-delay is required for backoff subcommand")
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
	case "cron":
		if config.CronExpression == "" {
			return errors.New("--cron is required for cron subcommand")
		}
		// Validate cron configuration
		if err := validateCronConfig(config); err != nil {
			return fmt.Errorf("invalid cron config: %w", err)
		}
	case "exponential":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for exponential strategy")
		}
		// Validate exponential strategy configuration
		if err := validateExponentialConfig(config); err != nil {
			return fmt.Errorf("invalid exponential config: %w", err)
		}
	case "fibonacci":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for fibonacci strategy")
		}
		// Validate fibonacci strategy configuration
		if err := validateFibonacciConfig(config); err != nil {
			return fmt.Errorf("invalid fibonacci config: %w", err)
		}
	case "linear":
		if config.Increment == 0 {
			return errors.New("--increment is required for linear strategy")
		}
		// Validate linear strategy configuration
		if err := validateLinearConfig(config); err != nil {
			return fmt.Errorf("invalid linear config: %w", err)
		}
	case "polynomial":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for polynomial strategy")
		}
		// Validate polynomial strategy configuration
		if err := validatePolynomialConfig(config); err != nil {
			return fmt.Errorf("invalid polynomial config: %w", err)
		}
	case "decorrelated-jitter":
		if config.BaseDelay == 0 {
			return errors.New("--base-delay is required for decorrelated-jitter strategy")
		}
		// Validate decorrelated-jitter strategy configuration
		if err := validateDecorrelatedJitterConfig(config); err != nil {
			return fmt.Errorf("invalid decorrelated-jitter config: %w", err)
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

// validateCronConfig validates the cron configuration
func validateCronConfig(config *Config) error {
	// Import cron package to validate expression
	// For now, do basic validation to avoid circular imports
	if config.CronExpression == "" {
		return errors.New("cron expression cannot be empty")
	}

	// Set default timezone if not specified
	if config.Timezone == "" {
		config.Timezone = "UTC"
	}

	// Basic validation - check if it looks like a cron expression or shortcut
	expr := strings.TrimSpace(config.CronExpression)
	if strings.HasPrefix(expr, "@") {
		// Shortcut format
		validShortcuts := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@hourly"}
		valid := false
		for _, shortcut := range validShortcuts {
			if expr == shortcut {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid cron shortcut: %s (valid shortcuts: %s)", expr, strings.Join(validShortcuts, ", "))
		}
	} else {
		// Standard cron format - should have 5 fields
		fields := strings.Fields(expr)
		if len(fields) != 5 {
			return fmt.Errorf("cron expression must have 5 fields (minute hour day month weekday), got %d", len(fields))
		}
	}

	return nil
}

// validatePatterns validates regex patterns for success/failure matching
func validatePatterns(config *Config) error {
	// Validate success pattern if provided
	if config.SuccessPattern != "" {
		pattern := config.SuccessPattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		if _, err := patterns.NewPatternMatcher(patterns.PatternConfig{
			SuccessPattern: pattern,
		}); err != nil {
			return fmt.Errorf("invalid success pattern: %w", err)
		}
	}

	// Validate failure pattern if provided
	if config.FailurePattern != "" {
		pattern := config.FailurePattern
		if config.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		if _, err := patterns.NewPatternMatcher(patterns.PatternConfig{
			FailurePattern: pattern,
		}); err != nil {
			return fmt.Errorf("invalid failure pattern: %w", err)
		}
	}

	return nil
}

// parseStringSliceFlag parses a comma-separated string slice flag value
func (p *argParser) parseStringSliceFlag(target *[]string) error {
	if p.pos+1 >= len(p.args) {
		return fmt.Errorf("%s requires a value", p.args[p.pos])
	}

	value := p.args[p.pos+1]
	if value == "" {
		*target = []string{}
	} else {
		*target = strings.Split(value, ",")
		// Trim whitespace from each field
		for i, field := range *target {
			(*target)[i] = strings.TrimSpace(field)
		}
	}

	p.pos += 2
	return nil
}

// validateExponentialConfig validates the exponential strategy configuration
func validateExponentialConfig(config *Config) error {
	// Set defaults if not provided
	if config.Multiplier == 0 {
		config.Multiplier = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateFibonacciConfig validates the fibonacci strategy configuration
func validateFibonacciConfig(config *Config) error {
	// Set defaults if not provided
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateLinearConfig validates the linear strategy configuration
func validateLinearConfig(config *Config) error {
	// Set defaults if not provided
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.Increment <= 0 {
		return errors.New("increment must be positive")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.Increment {
		return errors.New("max-delay must be greater than increment")
	}

	return nil
}

// validatePolynomialConfig validates the polynomial strategy configuration
func validatePolynomialConfig(config *Config) error {
	// Set defaults if not provided
	if config.Exponent == 0 {
		config.Exponent = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Exponent <= 0 {
		return errors.New("exponent must be positive")
	}

	if config.Exponent > 10.0 {
		return errors.New("exponent must be <= 10.0 to prevent overflow")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}

// validateDecorrelatedJitterConfig validates the decorrelated-jitter strategy configuration
func validateDecorrelatedJitterConfig(config *Config) error {
	// Set defaults if not provided
	if config.Multiplier == 0 {
		config.Multiplier = 3.0 // AWS recommendation
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 60 * time.Second
	}

	// Validate bounds
	if config.BaseDelay <= 0 {
		return errors.New("base-delay must be positive")
	}

	if config.Multiplier <= 1.0 {
		return errors.New("multiplier must be greater than 1.0")
	}

	if config.MaxDelay > 0 && config.MaxDelay < config.BaseDelay {
		return errors.New("max-delay must be greater than base-delay")
	}

	return nil
}
