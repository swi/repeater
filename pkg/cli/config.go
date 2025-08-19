package cli

import (
	"time"

	"github.com/swi/repeater/pkg/httpaware"
	"github.com/swi/repeater/pkg/patterns"
)

// Config represents the parsed CLI configuration
type Config struct {
	Subcommand     string
	Every          time.Duration
	Times          int64
	For            time.Duration
	Command        []string
	Help           bool
	SubcommandHelp bool // help for specific subcommand
	Version        bool
	ConfigFile     string

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
