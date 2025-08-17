# CLI Strategy Refactor: Phase 1 Analysis

## Current CLI Structure Analysis

### Existing Subcommands (Mode-Based)
```go
// From pkg/cli/cli.go normalizeSubcommand()
case "interval", "int", "i":     return "interval"
case "count", "cnt", "c":        return "count"  
case "duration", "dur", "d":     return "duration"
case "rate-limit", "rate", "rl", "r": return "rate-limit"
case "adaptive", "adapt", "a":   return "adaptive"
case "backoff", "back", "b":     return "backoff"
case "load-adaptive", "load", "la": return "load-adaptive"
case "cron", "cr":               return "cron"
```

## Strategy Categorization

### Category 1: Retry Strategies (Mathematical Algorithms)
These implement specific mathematical approaches for retry logic:

| Current Mode | New Strategy | Algorithm | Parameters |
|-------------|-------------|-----------|------------|
| `backoff` | **`exponential`** | 1s → 2s → 4s → 8s... | `--base-delay`, `--multiplier`, `--max-delay` |
| `adaptive` | **`adaptive`** | ML-based adaptation | `--base-interval`, `--learning-rate`, `--memory-window` |

**New Strategies to Add:**
| New Strategy | Algorithm | Parameters |
|-------------|-----------|------------|
| **`fibonacci`** | 1s → 1s → 2s → 3s → 5s... | `--base-delay`, `--max-delay` |
| **`linear`** | 1s → 2s → 3s → 4s... | `--increment`, `--max-delay` |
| **`polynomial`** | 1s → 1.5² → 2² → 2.5²... | `--base-delay`, `--exponent`, `--max-delay` |
| **`decorrelated-jitter`** | AWS-style smart jitter | `--base-delay`, `--multiplier`, `--max-delay` |

### Category 2: Execution Modes (Operational Patterns)
These define how/when commands are executed over time:

| Current Mode | Keep As-Is | Purpose | Parameters |
|-------------|------------|---------|------------|
| **`interval`** | ✅ Keep | Continuous execution at fixed intervals | `--every`, `--times`, `--for` |
| **`count`** | ✅ Keep | Execute N times | `--times`, `--interval` |
| **`duration`** | ✅ Keep | Execute for time period | `--for`, `--every` |
| **`cron`** | ✅ Keep | Schedule-based execution | `--cron`, `--timezone` |

### Category 3: Rate Control (Resource Management)
These manage resource usage and server-friendly execution:

| Current Mode | Keep As-Is | Purpose | Parameters |
|-------------|------------|---------|------------|
| **`rate-limit`** | ✅ Keep | Diophantine rate limiting | `--rate`, `--retry-pattern` |
| **`load-adaptive`** | ✅ Keep | System load awareness | `--target-cpu`, `--target-memory`, `--target-load` |

## Parameter Mapping Analysis

### Common Parameters Across All Strategies/Modes
```bash
# Retry/Attempt Control
--attempts, -a COUNT           # Maximum retry attempts
--timeout, -t DURATION         # Per-attempt timeout

# Pattern Matching  
--success-pattern REGEX        # Success detection
--failure-pattern REGEX        # Failure detection
--case-insensitive            # Pattern matching mode

# Output Control
--quiet, -q                   # Suppress output
--verbose, -v                 # Detailed output  
--stats-only                  # Statistics only
--stream, -s                  # Real-time streaming
```

### Strategy-Specific Parameters

#### Exponential Strategy (formerly backoff)
```bash
--base-delay, -b DURATION     # Initial delay (was --initial-delay)
--multiplier, -m FLOAT        # Growth multiplier (keep existing)
--max-delay, -x DURATION      # Maximum delay cap (was --max)
```

#### Fibonacci Strategy (new)
```bash
--base-delay, -b DURATION     # Base delay unit
--max-delay, -x DURATION      # Maximum delay cap
```

#### Linear Strategy (new)  
```bash
--increment, -i DURATION      # Linear increment amount
--max-delay, -x DURATION      # Maximum delay cap
```

#### Polynomial Strategy (new)
```bash
--base-delay, -b DURATION     # Base delay
--exponent, -e FLOAT          # Growth exponent (1.5, 2.0, 2.5, etc.)
--max-delay, -x DURATION      # Maximum delay cap
```

#### Decorrelated Jitter Strategy (new)
```bash
--base-delay, -b DURATION     # Base delay for calculations
--multiplier, -m FLOAT        # Jitter multiplier (AWS recommends 3.0)
--max-delay, -x DURATION      # Maximum delay cap
```

#### Adaptive Strategy (keep existing)
```bash
--base-interval, -b DURATION  # Base interval for adaptation
--learning-rate, -r FLOAT     # Learning rate (0.01-1.0)
--memory-window, -w INT       # Number of outcomes to remember
--min-interval DURATION       # Minimum interval bound
--max-interval DURATION       # Maximum interval bound
--slow-threshold FLOAT        # Slow response threshold
--fast-threshold FLOAT        # Fast response threshold
--failure-threshold FLOAT     # Circuit breaker threshold
--show-metrics, -m           # Show adaptive metrics
```

## Interface Transformation Examples

### Before (Current Mode-Based)
```bash
# Retry with exponential backoff
rpr backoff --initial-delay 1s --multiplier 2.0 --max 30s --attempts 5 -- curl api.com

# Continuous execution
rpr interval --every 30s --times 10 -- health-check.sh

# Rate limiting
rpr rate-limit --rate 100/1h --retry-pattern "0,10m,30m" -- api-call.sh
```

### After (Strategy-Based)
```bash
# Retry with exponential backoff  
rpr exponential --base-delay 1s --multiplier 2.0 --max-delay 30s --attempts 5 -- curl api.com

# Continuous execution (unchanged - execution mode)
rpr interval --every 30s --times 10 -- health-check.sh

# Rate limiting (unchanged - rate control)
rpr rate-limit --rate 100/1h --retry-pattern "0,10m,30m" -- api-call.sh

# New mathematical strategies
rpr fibonacci --base-delay 1s --max-delay 60s --attempts 5 -- flaky-service.sh
rpr linear --increment 2s --max-delay 30s --attempts 4 -- database-connect.sh
rpr polynomial --base-delay 1s --exponent 1.5 --max-delay 45s --attempts 6 -- api-call.sh
rpr decorrelated-jitter --base-delay 1s --multiplier 3.0 --max-delay 60s --attempts 8 -- aws-api.sh
```

## Help System Redesign

### New Main Help Organization
```bash
$ rpr --help

Repeater (rpr) - Intelligent Command Execution Tool

USAGE:
  rpr [GLOBAL OPTIONS] <STRATEGY|MODE> [OPTIONS] -- <COMMAND>

RETRY STRATEGIES (Mathematical Algorithms):
  exponential, exp       Exponential backoff (1s, 2s, 4s, 8s...)
  fibonacci, fib         Fibonacci backoff (1s, 1s, 2s, 3s, 5s...)
  linear, lin           Linear backoff (1s, 2s, 3s, 4s...)
  polynomial, poly      Polynomial backoff (customizable growth)
  decorrelated-jitter, dj AWS-style decorrelated jitter
  adaptive, adapt       Machine learning adaptive strategy

EXECUTION MODES (Operational Patterns):
  interval, int         Execute at regular intervals  
  count, cnt           Execute a specific number of times
  duration, dur        Execute for a time duration
  cron, cr             Execute on cron schedule

RATE CONTROL (Resource Management):
  rate-limit, rl       Diophantine rate limiting
  load-adaptive, la    Load-aware adaptive execution

COMMON OPTIONS:
  --attempts, -a COUNT        Maximum retry attempts
  --timeout, -t DURATION      Per-attempt timeout
  --success-pattern REGEX     Success detection pattern
  --failure-pattern REGEX     Failure detection pattern
  --case-insensitive         Case-insensitive pattern matching
  --quiet, -q                Suppress command output
  --verbose, -v              Show detailed execution info
  --stats-only               Show only execution statistics

EXAMPLES:
  # Retry strategies
  rpr exponential --base-delay 1s --attempts 5 -- curl api.com
  rpr fibonacci --base-delay 500ms --attempts 3 -- flaky-command
  rpr linear --increment 2s --attempts 4 -- database-connection
  
  # Execution modes  
  rpr interval --every 30s --times 10 -- health-check
  rpr cron --cron '@daily' -- backup-script
  
  # Rate control
  rpr rate-limit --rate 100/1h -- api-batch-job
```

### Strategy-Specific Help Examples
```bash
$ rpr exponential --help

Exponential Backoff Strategy

DESCRIPTION:
  Doubles delay after each failure: 1s, 2s, 4s, 8s, 16s...
  Industry standard for network operations and API calls.

USAGE:
  rpr exponential [OPTIONS] -- <COMMAND>

STRATEGY OPTIONS:
  --base-delay, -b DURATION   Initial delay (default: 1s)
  --multiplier, -m FLOAT      Growth multiplier (default: 2.0)
  --max-delay, -x DURATION    Maximum delay cap (default: 60s)

COMMON OPTIONS:
  --attempts, -a COUNT        Maximum attempts (default: 3)
  --timeout, -t DURATION      Per-attempt timeout
  --success-pattern REGEX     Success detection
  --failure-pattern REGEX     Failure detection  

EXAMPLES:
  rpr exponential --base-delay 500ms --attempts 5 -- curl api.com
  rpr exp -b 1s -m 1.5 -x 30s -a 3 -- database-connect
  rpr exponential --attempts 5 --success-pattern "OK" -- deployment.sh
```

## Backward Compatibility Strategy

### Legacy Mode Mapping
```go
// Support old interface with deprecation warnings
var legacyModeMap = map[string]string{
    "backoff": "exponential",
    "back":    "exponential",
    "b":       "exponential",
}

var legacyParamMap = map[string]string{
    "initial-delay": "base-delay",
    "initial":       "base-delay",
    "i":            "base-delay", // When used with backoff
    "max":          "max-delay",
}
```

### Migration Examples
```bash
# Legacy command (still works with warnings)
$ rpr backoff --initial-delay 1s --max 30s --attempts 5 -- command
⚠️  'backoff' is deprecated, use 'exponential' instead
⚠️  '--initial-delay' is deprecated, use '--base-delay' instead  
⚠️  '--max' is deprecated, use '--max-delay' instead
[command executes normally]

# New equivalent command
$ rpr exponential --base-delay 1s --max-delay 30s --attempts 5 -- command
```

## Implementation Architecture

### Config Structure Updates
```go
type Config struct {
    // Change from Subcommand to Strategy for clarity
    Strategy string  // exponential, fibonacci, linear, etc.
    Mode     string  // interval, count, duration, cron (for execution modes)
    
    // Unified retry parameters
    Attempts         int           // --attempts (was MaxRetries)
    Timeout          time.Duration // --timeout (per attempt)
    SuccessPattern   string        // --success-pattern
    FailurePattern   string        // --failure-pattern  
    CaseInsensitive  bool          // --case-insensitive
    
    // Strategy-specific parameters (unified naming)
    BaseDelay        time.Duration // --base-delay (was InitialInterval)
    Increment        time.Duration // --increment (linear strategy)
    Multiplier       float64       // --multiplier (keep existing)
    Exponent         float64       // --exponent (polynomial strategy)
    MaxDelay         time.Duration // --max-delay (was BackoffMax)
    
    // Keep existing execution mode parameters
    Every            time.Duration // --every
    Times            int64         // --times
    For              time.Duration // --for
    // ... rest unchanged
}
```

### Strategy Interface Design
```go
// pkg/strategies/
type Strategy interface {
    Name() string
    NextDelay(attempt int, lastDuration time.Duration) time.Duration
    ShouldRetry(attempt int, err error, output string) bool
    ValidateConfig(config *Config) error
}

type ExponentialStrategy struct {
    BaseDelay  time.Duration
    Multiplier float64
    MaxDelay   time.Duration
}

type FibonacciStrategy struct {
    BaseDelay time.Duration
    MaxDelay  time.Duration
}

type LinearStrategy struct {
    Increment time.Duration
    MaxDelay  time.Duration
}

type PolynomialStrategy struct {
    BaseDelay time.Duration
    Exponent  float64
    MaxDelay  time.Duration
}

type DecorrelatedJitterStrategy struct {
    BaseDelay  time.Duration
    Multiplier float64
    MaxDelay   time.Duration
}
```

## Migration Timeline

### Phase 1: Add New Strategies (Backward Compatible)
- Add fibonacci, linear, polynomial, decorrelated-jitter strategies
- Keep existing backoff mode working unchanged
- No breaking changes

### Phase 2: Add Strategy Aliases (Backward Compatible)  
- Add "exponential" as alias for "backoff"
- Add parameter aliases (--base-delay for --initial-delay)
- Still no breaking changes

### Phase 3: Deprecation Warnings (Gradual Migration)
- Show warnings for old modes and parameters
- Update documentation to show new preferred interface
- Give users time to migrate

### Phase 4: New Default Interface (Breaking Change Window)
- Make strategy-based interface the primary documentation
- Legacy interface still works but with prominent warnings
- Update all examples and help text

### Phase 5: Legacy Removal (Future Breaking Change)
- Remove legacy modes after 6-12 month deprecation period
- Clean up codebase of backward compatibility code

## Success Metrics

### User Experience Improvements
- ✅ Strategy selection matches mathematical mental model
- ✅ Consistent parameter naming across strategies  
- ✅ Enhanced discoverability via categorized help
- ✅ Intuitive command construction: `rpr fibonacci` vs `rpr backoff --fibonacci`

### Technical Quality Maintenance
- ✅ No functionality lost during refactor
- ✅ All existing tests continue to pass
- ✅ New strategies thoroughly tested
- ✅ Performance maintained or improved
- ✅ Backward compatibility preserved during transition

This analysis provides the foundation for implementing the strategy-based CLI refactor while maintaining all existing functionality and providing a smooth migration path.