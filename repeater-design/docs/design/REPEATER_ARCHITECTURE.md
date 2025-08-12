# Repeater Technical Architecture

## System Overview

Repeater is designed as a comprehensive, extensible Go application with advanced scheduling capabilities, plugin system, and optional integration with shared infrastructure. The architecture emphasizes modularity, testability, performance, and extensibility for continuous execution scenarios.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Repeater (rpr)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   CLI Parser    │───▶│   Config Loader  │                   │
│  │   (Cobra)       │    │   (TOML/Env)     │                   │
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
│           │  │  Interval   │  │    Rate Limiter     ││         │
│           │  │  Scheduler  │  │   (Diophantine)     ││         │
│           │  └─────────────┘  └─────────────────────┘│         │
│           │  ┌─────────────┐  ┌─────────────────────┐│         │
│           │  │   Count     │  │   Stop Condition    ││         │
│           │  │  Scheduler  │  │     Manager         ││         │
│           │  └─────────────┘  └─────────────────────┘│         │
│           │  ┌─────────────┐  ┌─────────────────────┐│         │
│           │  │  Duration   │  │    Signal Handler   ││         │
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
│  │   Metrics       │    │   Daemon Client  │                   │
│  │   Collector     │    │   (Optional)     │                   │
│  └─────────────────┘    └──────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. CLI Parser (cmd/rpr)
**Responsibility**: Parse command-line arguments and route to appropriate subcommand handlers.

```go
// cmd/rpr/main.go
type RootCommand struct {
    Version    bool
    ConfigFile string
    Verbose    bool
}

type GlobalOptions struct {
    Quiet           bool
    Verbose         bool
    OutputFile      string
    ContinueOnError bool
    MaxFailures     int
    Timeout         time.Duration
    WorkingDir      string
}
```

**Key Features**:
- Cobra-based CLI with subcommand architecture
- Global option inheritance
- Comprehensive help system with examples
- Configuration file and environment variable support

### 2. Scheduler Engine (pkg/scheduler)
**Responsibility**: Core execution scheduling logic with pluggable scheduler implementations.

```go
// pkg/scheduler/scheduler.go
type Scheduler interface {
    Next() <-chan time.Time
    Stop()
    Stats() SchedulerStats
}

type SchedulerStats struct {
    ExecutionsPlanned int64
    ExecutionsSkipped int64
    AverageInterval   time.Duration
    NextExecution     time.Time
}

// Concrete implementations
type IntervalScheduler struct {
    interval  time.Duration
    jitter    float64
    immediate bool
    ticker    *time.Ticker
}

type CountScheduler struct {
    remaining int64
    interval  time.Duration
    parallel  int
}

type DurationScheduler struct {
    endTime   time.Time
    interval  time.Duration
    ticker    *time.Ticker
}
```

### 3. Command Executor (pkg/executor)
**Responsibility**: Execute commands with timeout, output capture, and error handling.

```go
// pkg/executor/executor.go
type Executor struct {
    timeout     time.Duration
    workingDir  string
    env         []string
    outputMgr   *OutputManager
    metrics     *MetricsCollector
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

### 4. Output Manager (pkg/output)
**Responsibility**: Handle command output streaming, aggregation, and logging.

```go
// pkg/output/manager.go
type OutputManager struct {
    mode        OutputMode
    file        *os.File
    aggregator  *OutputAggregator
    quiet       bool
    verbose     bool
}

type OutputMode int
const (
    OutputModeStream OutputMode = iota
    OutputModeAggregate
    OutputModeSuppress
)

type OutputAggregator struct {
    executions []ExecutionOutput
    mutex      sync.RWMutex
}
```

### 5. Stop Condition Manager (pkg/conditions)
**Responsibility**: Track and evaluate stop conditions across different criteria.

```go
// pkg/conditions/manager.go
type StopConditionManager struct {
    conditions []StopCondition
    stats      ExecutionStats
}

type StopCondition interface {
    ShouldStop(stats ExecutionStats) bool
    Description() string
}

// Implementations
type CountCondition struct {
    maxCount int64
}

type DurationCondition struct {
    maxDuration time.Duration
    startTime   time.Time
}

type FailureCondition struct {
    maxFailures        int
    consecutiveFailures int
}
```

### 6. Rate Limiter Integration (pkg/ratelimit)
**Responsibility**: Mathematical rate limiting using algorithms from patience.

```go
// pkg/ratelimit/limiter.go
type RateLimiter interface {
    Allow() bool
    Wait(ctx context.Context) error
    Stats() RateLimitStats
}

// Reuse from patience
type DiophantineRateLimiter struct {
    limit      int64
    window     time.Duration
    daemon     *DaemonClient
    resourceID string
}
```

## Data Flow

### 1. Initialization Flow
```
CLI Args → Config Loading → Subcommand Routing → Scheduler Creation → Executor Setup
```

### 2. Execution Flow
```
Scheduler.Next() → Command Execution → Output Processing → Metrics Collection → Stop Condition Check
```

### 3. Shutdown Flow
```
Signal Received → Graceful Scheduler Stop → Wait for Active Commands → Cleanup Resources
```

## Concurrency Model

### Goroutine Architecture
```go
// Main execution loop
func (r *Repeater) Run(ctx context.Context) error {
    // Goroutine 1: Scheduler
    scheduleChan := r.scheduler.Next()
    
    // Goroutine 2: Signal handling
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Goroutine 3: Stop condition monitoring
    stopChan := r.stopManager.Monitor(ctx)
    
    // Main loop
    for {
        select {
        case <-scheduleChan:
            go r.executeCommand(ctx)
        case <-signalChan:
            return r.gracefulShutdown()
        case <-stopChan:
            return r.normalShutdown()
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

### Thread Safety
- **Scheduler**: Thread-safe with internal mutexes
- **Output Manager**: Synchronized writes to files/stdout
- **Metrics**: Atomic counters and protected data structures
- **Stop Conditions**: Thread-safe evaluation with read-write locks

## Error Handling Strategy

### Error Categories
1. **Configuration Errors**: Invalid CLI args, bad config files
2. **Runtime Errors**: Command execution failures, system resource issues
3. **Network Errors**: Daemon communication failures (for rate limiting)
4. **Resource Errors**: File system, memory, CPU constraints

### Error Recovery
```go
type ErrorHandler struct {
    continueOnError bool
    maxFailures     int
    backoffStrategy BackoffStrategy
}

func (eh *ErrorHandler) HandleError(err error, attempt int) (shouldContinue bool, delay time.Duration) {
    switch {
    case isConfigurationError(err):
        return false, 0 // Fatal, stop immediately
    case isCommandError(err):
        return eh.continueOnError, eh.calculateBackoff(attempt)
    case isResourceError(err):
        return true, eh.calculateResourceBackoff(err)
    default:
        return eh.continueOnError, time.Second
    }
}
```

## Performance Considerations

### Memory Management
- **Bounded Output Buffers**: Prevent memory growth during long runs
- **Metrics Rotation**: Automatic cleanup of old metrics data
- **Goroutine Pooling**: Reuse goroutines for command execution

### CPU Optimization
- **Efficient Scheduling**: Minimal CPU overhead between executions
- **Lazy Initialization**: Create resources only when needed
- **Batch Operations**: Group related operations for efficiency

### I/O Optimization
- **Buffered Output**: Reduce system call overhead
- **Async Logging**: Non-blocking log writes
- **Connection Pooling**: Reuse daemon connections

## Integration Points

### Shared Components with Patience
```go
// Reuse existing packages
import (
    "github.com/shaneisley/patience/pkg/backoff"    // Rate limiting algorithms
    "github.com/shaneisley/patience/pkg/daemon"     // Daemon client
    "github.com/shaneisley/patience/pkg/config"     // Configuration loading
    "github.com/shaneisley/patience/pkg/metrics"    // Metrics collection
)
```

### Daemon Communication
- **Protocol**: Same JSON-over-Unix-socket as patience
- **Coordination**: Share rate limiting resources across tools
- **Fallback**: Graceful degradation when daemon unavailable

## Testing Architecture

### Unit Testing
- **Scheduler Tests**: Verify timing accuracy and edge cases
- **Executor Tests**: Mock command execution for reliability
- **Output Tests**: Validate formatting and aggregation
- **Condition Tests**: Test stop condition logic

### Integration Testing
- **End-to-End**: Full CLI execution with real commands
- **Daemon Integration**: Test coordination with shared daemon
- **Performance**: Timing accuracy and resource usage
- **Error Scenarios**: Failure handling and recovery

### Benchmarking
```go
func BenchmarkSchedulerOverhead(b *testing.B) {
    scheduler := NewIntervalScheduler(time.Millisecond)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        <-scheduler.Next()
    }
}
```

## Deployment Considerations

### Binary Distribution
- **Single Binary**: No external dependencies
- **Cross-Platform**: Linux, macOS, Windows support
- **Size Optimization**: Minimal binary size with build flags

### Configuration Management
- **Default Locations**: `~/.config/repeater/config.toml`
- **Environment Override**: `RPR_*` environment variables
- **Runtime Configuration**: CLI flags take precedence

### Monitoring Integration
- **Metrics Export**: Prometheus-compatible metrics
- **Health Checks**: Built-in health endpoint
- **Logging**: Structured logging with configurable levels