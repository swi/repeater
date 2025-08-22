# Repeater Technical Architecture

## System Overview

Repeater is designed as a comprehensive, extensible Go application with advanced scheduling capabilities, plugin system, and production-ready features. The architecture emphasizes modularity, testability, performance, and extensibility for continuous execution scenarios.

> 📖 **User Perspective:** For practical usage examples of these architectural components, see the [Usage Guide](USAGE.md). For contributing to this architecture, see [Contributing Guidelines](CONTRIBUTING.md#key-architecture-patterns).

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Repeater (rpr)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   CLI Parser    │───▶│   Config Loader  │                   │
│  │   (Multi-level) │    │   (TOML/Env)     │                   │
│  └─────────────────┘    └──────────────────┘                   │
│           │                       │                            │
│           ▼                       ▼                            │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   Subcommand    │───▶│   Scheduler      │                   │
│  │   Router        │    │   Factory        │                   │
│  └─────────────────┘    └──────────────────┘                   │
│                                   │                            │
│                                   ▼                            │
│           ┌──────────────────────────────────────────┐         │
│           │            Scheduler Engine              │         │
│           ├──────────────────────────────────────────┤         │
│           │  ┌─────────────┐  ┌─────────────────────┐│         │
│           │  │  Interval   │  │    HTTP-Aware       ││         │
│           │  │  Scheduler  │  │   Intelligence      ││         │
│           │  └─────────────┘  └─────────────────────┘│         │
│           │  ┌─────────────┐  ┌─────────────────────┐│         │
│           │  │   Cron      │  │   Pattern Matching  ││         │
│           │  │  Scheduler  │  │     Engine          ││         │
│           │  └─────────────┘  └─────────────────────┘│         │
│           │  ┌─────────────┐  ┌─────────────────────┐│         │
│           │  │  Adaptive   │  │    Plugin System    ││         │
│           │  │  Scheduler  │  │                     ││         │
│           │  └─────────────┘  └─────────────────────┘│         │
│           └──────────────────────────────────────────┘         │
│                                   │                            │
│                                   ▼                            │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   Command       │───▶│   Output         │                   │
│  │   Executor      │    │   Manager        │                   │
│  └─────────────────┘    └──────────────────┘                   │
│           │                       │                            │
│           ▼                       ▼                            │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   Metrics &     │    │   Recovery &     │                   │
│  │   Health        │    │   Error Handling │                   │
│  └─────────────────┘    └──────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. CLI Parser (`cmd/rpr` & `pkg/cli`)

**Responsibility**: Parse command-line arguments with multi-level abbreviations and route to appropriate subcommand handlers.

```go
type Config struct {
    Subcommand      string
    Every           time.Duration
    Times           int64
    For             time.Duration
    Quiet           bool
    Verbose         bool
    StatsOnly       bool
    SuccessPattern  string
    FailurePattern  string
    CaseInsensitive bool
    HTTPAware       bool
    Command         []string
}
```

**Key Features**:
- Multi-level abbreviations (`interval`/`int`/`i`)
- Global option inheritance
- Comprehensive help system with examples
- Configuration file and environment variable support
- Pattern matching configuration
- HTTP-aware settings

### 2. Scheduler Engine (`pkg/scheduler` & related packages)

**Responsibility**: Core execution scheduling logic with pluggable scheduler implementations.

```go
type Scheduler interface {
    Next() <-chan time.Time
    Stop()
}

// Core scheduler implementations
type IntervalScheduler struct {
    interval  time.Duration
    jitter    float64
    immediate bool
    ticker    *time.Ticker
}

type CronScheduler struct {
    expression *cron.CronExpression
    timezone   *time.Location
    stopCh     chan struct{}
}

type AdaptiveScheduler struct {
    baseInterval time.Duration
    current      time.Duration
    aimd         *AIMDController
    metrics      *AdaptiveMetrics
}
```

**Scheduler Types**:
1. **Interval Scheduler** - Fixed intervals with optional jitter
2. **Cron Scheduler** - Time-based with timezone support
3. **Adaptive Scheduler** - AI-driven interval adjustment
4. **Backoff Scheduler** - Exponential backoff with jitter
5. **Load-Aware Scheduler** - System resource monitoring
6. **Rate-Limited Scheduler** - Mathematical rate limiting
7. **Count Scheduler** - Execute N times
8. **Duration Scheduler** - Execute for time period

### 3. Command Executor (`pkg/executor`)

**Responsibility**: Execute commands with timeout, output capture, and error handling.

```go
type Executor struct {
    timeout     time.Duration
    workingDir  string
    env         []string
    streaming   bool
}

type ExecutionResult struct {
    Command     []string
    StartTime   time.Time
    Duration    time.Duration
    ExitCode    int
    Stdout      string
    Stderr      string
    Success     bool
    Error       error
}

func (e *Executor) Execute(ctx context.Context, command []string) (*ExecutionResult, error)
```

**Features**:
- Context-aware execution with cancellation
- Streaming output support
- Timeout handling
- Environment variable support
- Working directory specification

### 4. Pattern Matching Engine (`pkg/patterns`)

**Responsibility**: Success/failure detection via regex patterns with precedence rules.

```go
type PatternMatcher struct {
    successPattern  *regexp.Regexp
    failurePattern  *regexp.Regexp
    caseInsensitive bool
}

func (pm *PatternMatcher) EvaluateResult(result *ExecutionResult) PatternResult {
    // Failure patterns take precedence over success patterns
    if pm.failurePattern != nil && pm.failurePattern.MatchString(result.Stdout) {
        return PatternResult{Success: false, Reason: "failure pattern matched"}
    }
    
    if pm.successPattern != nil && pm.successPattern.MatchString(result.Stdout) {
        return PatternResult{Success: true, Reason: "success pattern matched"}
    }
    
    // Fall back to exit code
    return PatternResult{Success: result.ExitCode == 0, Reason: "exit code"}
}
```

### 5. HTTP-Aware Intelligence (`pkg/httpaware`)

**Responsibility**: Parse HTTP responses to extract timing information for optimal API scheduling.

> 📖 **Usage Examples:** See [HTTP-Aware Intelligence](USAGE.md#http-aware-intelligence) in the Usage Guide for practical configuration and real-world API examples

```go
type HTTPAwareParser struct {
    maxDelay        time.Duration
    minDelay        time.Duration
    parseJSON       bool
    parseHeaders    bool
    customFields    []string
    trustClient     bool
}

func (p *HTTPAwareParser) ParseTiming(response *http.Response, body []byte) *TimingInfo {
    // Priority: Headers > Custom JSON > Standard JSON > Nested structures
    if timing := p.parseRetryAfterHeader(response); timing != nil {
        return timing
    }
    
    if p.parseJSON {
        if timing := p.parseJSONTiming(body); timing != nil {
            return timing
        }
    }
    
    return nil
}
```

**Supported Sources**:
- **Retry-After Headers** (highest priority)
- **Custom JSON fields** (configurable)
- **Standard JSON fields** (`retry_after`, `retryAfter`)
- **Rate limit structures** (`rate_limit.reset_in`)
- **Nested timing** (`error.retry_after`)

### 6. Plugin System (`pkg/plugin`)

**Responsibility**: Extensible architecture for custom schedulers, executors, and outputs.

> 🤝 **Plugin Development:** See [Plugin Development Guide](CONTRIBUTING.md#plugin-development) for creating custom plugins and integration examples

```go
type SchedulerPlugin interface {
    Name() string
    Version() string
    Description() string
    
    NewScheduler(config map[string]interface{}) (Scheduler, error)
    ValidateConfig(config map[string]interface{}) error
    ConfigSchema() *ConfigSchema
}

type PluginManager struct {
    plugins    map[string]SchedulerPlugin
    pluginDirs []string
    mu         sync.RWMutex
}

func (pm *PluginManager) LoadPlugins() error {
    // Scan plugin directories for .so files
    // Load plugins using Go's plugin package
    // Register discovered plugins with validation
}
```

### 7. Recovery & Error Handling (`pkg/recovery`)

**Responsibility**: Circuit breakers, retry policies, and comprehensive error management.

```go
type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    state          CircuitState
    failures       int
    lastFailureTime time.Time
}

type RetryPolicy struct {
    maxRetries   int
    backoff      BackoffStrategy
    retryableErrors []error
}
```

### 8. Runner Orchestration (`pkg/runner`)

**Responsibility**: Coordinate all components for end-to-end execution.

```go
type Runner struct {
    config         *cli.Config
    scheduler      Scheduler
    executor       *Executor
    patternMatcher *PatternMatcher
    httpParser     *HTTPAwareParser
    healthServer   *HealthServer
    metricsServer  *MetricsServer
}

func (r *Runner) Run(ctx context.Context) error {
    // Initialize all components
    // Start health and metrics servers
    // Begin execution loop
    // Handle signals and stop conditions
    // Cleanup resources
}
```

## Data Flow

### 1. Initialization Flow
```
CLI Args → Config Loading → Plugin Loading → Component Creation → Validation
```

### 2. Execution Flow
```
Scheduler.Next() → Command Execution → Pattern Evaluation → HTTP Parsing → 
Metrics Collection → Stop Condition Check → Next Iteration
```

### 3. Shutdown Flow
```
Signal Received → Graceful Scheduler Stop → Wait for Active Commands → 
Cleanup Resources → Exit with Status
```

## Advanced Features

### Adaptive Scheduling Algorithm

The adaptive scheduler uses an Additive Increase Multiplicative Decrease (AIMD) algorithm:

```go
type AIMDController struct {
    current         time.Duration
    baseInterval    time.Duration
    minInterval     time.Duration
    maxInterval     time.Duration
    successFactor   float64  // Additive increase
    failureFactor   float64  // Multiplicative decrease
    responseThreshold time.Duration
}

func (a *AIMDController) Adjust(success bool, responseTime time.Duration) time.Duration {
    if success && responseTime < a.responseThreshold {
        // Additive increase - more frequent execution
        a.current = time.Duration(float64(a.current) * (1 - a.successFactor))
    } else {
        // Multiplicative decrease - less frequent execution
        a.current = time.Duration(float64(a.current) * a.failureFactor)
    }
    
    // Apply bounds
    if a.current < a.minInterval {
        a.current = a.minInterval
    }
    if a.current > a.maxInterval {
        a.current = a.maxInterval
    }
    
    return a.current
}
```

### Load-Aware Scheduling

Monitors system resources and adjusts timing:

```go
type LoadAwareScheduler struct {
    baseInterval   time.Duration
    targetCPU      float64
    targetMemory   float64
    targetLoad     float64
    currentInterval time.Duration
}

func (s *LoadAwareScheduler) adjustForLoad() time.Duration {
    cpuUsage := getCurrentCPUUsage()
    memUsage := getCurrentMemoryUsage()
    loadAvg := getLoadAverage()
    
    // Calculate adjustment factor based on resource usage
    factor := 1.0
    if cpuUsage > s.targetCPU {
        factor *= (cpuUsage / s.targetCPU)
    }
    if memUsage > s.targetMemory {
        factor *= (memUsage / s.targetMemory)
    }
    if loadAvg > s.targetLoad {
        factor *= (loadAvg / s.targetLoad)
    }
    
    return time.Duration(float64(s.baseInterval) * factor)
}
```

## Concurrency Model

### Goroutine Architecture

```go
func (r *Runner) Run(ctx context.Context) error {
    // Goroutine 1: Scheduler ticker
    scheduleChan := r.scheduler.Next()
    
    // Goroutine 2: Signal handling
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Goroutine 3: Health server (if enabled)
    if r.healthServer != nil {
        go r.healthServer.Start(ctx)
    }
    
    // Goroutine 4: Metrics server (if enabled)
    if r.metricsServer != nil {
        go r.metricsServer.Start(ctx)
    }
    
    // Main execution loop
    for {
        select {
        case <-scheduleChan:
            go r.executeCommand(ctx) // Goroutine per execution
        case <-signalChan:
            return r.gracefulShutdown()
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### Thread Safety

- **Scheduler**: Thread-safe with internal synchronization
- **Metrics Collection**: Atomic operations and mutexes
- **Pattern Matching**: Stateless evaluation (thread-safe)
- **HTTP Parsing**: Stateless parsing (thread-safe)
- **Plugin System**: Read-write locks for plugin registry

## Performance Characteristics

### Memory Management
- **Bounded Output Buffers**: Prevent memory growth during long runs
- **Metrics Rotation**: Automatic cleanup of historical data
- **Goroutine Pooling**: Reuse execution goroutines when possible
- **Plugin Caching**: Cache loaded plugins to avoid repeated loading

### CPU Optimization
- **Efficient Scheduling**: Minimal overhead between executions (<1ms)
- **Lazy Initialization**: Create resources only when needed
- **Batch Operations**: Group related operations for efficiency
- **Pattern Compilation**: Compile regex patterns once at startup

### Timing Precision
- **Microsecond Accuracy**: Uses Go's high-precision timers
- **Jitter Support**: Configurable randomization to prevent thundering herd
- **Drift Correction**: Compensates for execution time in interval calculations
- **Timezone Handling**: Proper DST transitions for cron scheduling

## Quality Assurance

### Testing Architecture

```go
// Unit Testing
func TestIntervalScheduler(t *testing.T) {
    scheduler := NewIntervalScheduler(100 * time.Millisecond)
    start := time.Now()
    
    // Test timing accuracy
    <-scheduler.Next()
    elapsed := time.Since(start)
    assert.InDelta(t, 100*time.Millisecond, elapsed, float64(5*time.Millisecond))
}

// Integration Testing
func TestEndToEndExecution(t *testing.T) {
    config := &Config{
        Subcommand: "interval",
        Every:      time.Second,
        Times:      3,
        Command:    []string{"echo", "test"},
    }
    
    runner := NewRunner(config)
    result := runner.Run(context.Background())
    
    assert.NoError(t, result)
    assert.Equal(t, 3, runner.Stats().TotalExecutions)
}

// Performance Testing
func BenchmarkSchedulerOverhead(b *testing.B) {
    scheduler := NewIntervalScheduler(time.Microsecond)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        <-scheduler.Next()
    }
}
```

### Quality Metrics
- **Test Coverage**: 90%+ across all packages
- **Benchmark Tests**: Performance regression detection
- **Race Testing**: `go test -race` for concurrency safety
- **Integration Tests**: End-to-end workflow validation
- **Linting**: golangci-lint compliance

## Project Structure

```
cmd/rpr/                 # Main application entry point
├── main.go              # CLI application with signal handling
├── config.go            # Configuration integration
└── *_test.go            # Integration tests

pkg/                     # Core library packages
├── cli/                 # Command-line interface and parsing
├── scheduler/           # Scheduling algorithms (8 types)
├── executor/            # Command execution engine
├── runner/              # Main execution orchestrator
├── config/              # Configuration management
├── metrics/             # Prometheus metrics server
├── health/              # Health check endpoints
├── recovery/            # Circuit breaker and retry logic
├── ratelimit/           # Rate limiting algorithms
├── plugin/              # Plugin system architecture
├── adaptive/            # AI-driven adaptive scheduling
├── cron/                # Cron expression parsing
├── patterns/            # Pattern matching engine
├── httpaware/           # HTTP-aware intelligence
└── errors/              # Error categorization and handling
```

## Deployment Considerations

### Binary Distribution
- **Single Binary**: No external dependencies
- **Cross-Platform**: Linux, macOS, Windows support
- **Size Optimization**: Minimal binary size with build flags

### Configuration Management
- **Default Locations**: `~/.config/rpr/config.toml`
- **Environment Override**: `RPR_*` environment variables
- **Runtime Configuration**: CLI flags take precedence

### Monitoring Integration
- **Metrics Export**: Prometheus-compatible endpoints
- **Health Checks**: HTTP endpoints for monitoring
- **Structured Logging**: JSON logs with correlation IDs

This architecture provides a robust, extensible foundation for continuous command execution with advanced scheduling capabilities and production-ready features.

## See Also

### Related Documentation
- 📖 **[README.md](README.md)** - Project overview and quick start
- 📚 **[USAGE.md](USAGE.md)** - Comprehensive usage guide with practical examples
- 🤝 **[CONTRIBUTING.md](CONTRIBUTING.md)** - Development guidelines for contributing to this architecture
- 📋 **[FEATURES.md](FEATURES.md)** - Implementation roadmap and technical achievements
- 📝 **[CHANGELOG.md](CHANGELOG.md)** - Architectural evolution and version history

### Implementation Examples
- 🔧 **[Plugin Development](CONTRIBUTING.md#plugin-development)** - Extending the architecture with custom components
- 📊 **[Performance Benchmarks](FEATURES.md#quality-metrics)** - Real-world performance metrics
- 🧪 **[Testing Strategy](CONTRIBUTING.md#tdd-workflow-mandatory)** - Quality assurance methodology