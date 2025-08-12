# Advanced Features Implementation Plan

## 🎯 **Overview**

This document outlines the implementation plan for advanced and future features in the Repeater project. The core functionality is complete and production-ready. This plan focuses on enhancing observability, scheduling capabilities, and extensibility.

## 📊 **Current Status Summary**

### ✅ **Production Ready (Core Features)**
- All scheduling modes (interval, count, duration, cron, adaptive, backoff, load-adaptive, rate-limit)
- CLI with abbreviations and Unix pipeline integration
- Configuration files (TOML) with environment overrides
- Mathematical rate limiting (local only)
- Comprehensive error handling and recovery
- Signal handling and graceful shutdown
- **Plugin System**: Extensible architecture for custom schedulers and executors ✅
- **Health endpoints**: HTTP server integrated ✅
- **Metrics collection**: HTTP server integrated ✅
- **Cron-like scheduling**: Time-based patterns with timezone support ✅

### 🚧 **Requires Development**
- Distributed multi-node coordination (future)
- Advanced plugin types (output processors, custom executors)
- Enhanced observability (Grafana dashboards, alerting)

---

## 🚀 **Phase 1: Observability Integration (Quick Wins)**

**Goal**: Integrate existing health and metrics servers into the runner
**Effort**: 1-2 days
**Priority**: HIGH

### 1.1 Health Endpoints Integration

**Current State**: Full HTTP server implementation exists in `pkg/health/`
**Gap**: Runner doesn't start the health server

#### Implementation Steps:

1. **Modify Runner to Start Health Server** (`pkg/runner/runner.go`)
```go
// In NewRunner function, after validation
if config.HealthEnabled {
    healthServer := health.NewHealthServer(config.HealthPort)
    r.healthServer = healthServer
}

// In Run function, start health server
if r.healthServer != nil {
    go func() {
        if err := r.healthServer.Start(ctx); err != nil {
            // Log error but don't fail execution
        }
    }()
    defer r.healthServer.Stop()
}
```

2. **Update Health Server with Execution Stats**
```go
// In execution loop, update health server
if r.healthServer != nil {
    r.healthServer.SetExecutionStats(health.ExecutionStats{
        TotalExecutions:      stats.TotalExecutions,
        SuccessfulExecutions: stats.SuccessfulExecutions,
        FailedExecutions:     stats.FailedExecutions,
        // ... other stats
    })
}
```

3. **Add Health Server Field to Runner Struct**
```go
type Runner struct {
    config       *cli.Config
    healthServer *health.HealthServer  // Add this field
}
```

#### Testing:
- Start rpr with `--config health-config.toml` where `health_enabled = true`
- Verify `curl http://localhost:8080/health` returns JSON response
- Verify `/ready` and `/live` endpoints work
- Test health server shutdown on Ctrl+C

### 1.2 Metrics Server Integration

**Current State**: Full Prometheus-compatible server exists in `pkg/metrics/`
**Gap**: Runner doesn't start the metrics server

#### Implementation Steps:

1. **Modify Runner to Start Metrics Server** (similar pattern to health)
```go
// Add metrics server field to Runner struct
type Runner struct {
    config        *cli.Config
    healthServer  *health.HealthServer
    metricsServer *metrics.MetricsServer  // Add this field
}

// In NewRunner function
if config.MetricsEnabled {
    metricsServer := metrics.NewMetricsServer(config.MetricsPort)
    r.metricsServer = metricsServer
}

// In Run function
if r.metricsServer != nil {
    go func() {
        if err := r.metricsServer.Start(ctx); err != nil {
            // Log error but don't fail execution
        }
    }()
    defer r.metricsServer.Stop()
}
```

2. **Update Metrics with Execution Data**
```go
// In execution loop, update metrics
if r.metricsServer != nil {
    r.metricsServer.RecordExecution(result.Duration, result.ExitCode == 0)
    r.metricsServer.UpdateCurrentInterval(nextInterval)
}
```

#### Testing:
- Start rpr with metrics enabled in config
- Verify `curl http://localhost:9090/metrics` returns Prometheus format
- Verify metrics update during execution
- Test concurrent access to metrics endpoint

### 1.3 Integration Tests

Create comprehensive integration tests for observability:

```go
// cmd/rpr/observability_integration_test.go
func TestObservabilityIntegration(t *testing.T) {
    // Test health + metrics + execution together
    // Verify endpoints respond correctly
    // Test graceful shutdown
}
```

---

## 🕐 **Phase 2: Cron-like Scheduling (New Feature)**

**Goal**: Add time-based scheduling patterns like cron
**Effort**: 1-2 weeks
**Priority**: MEDIUM

### 2.1 Cron Expression Parser

Create new package `pkg/cron/` with cron expression parsing:

```go
// pkg/cron/parser.go
type CronExpression struct {
    Minute     []int  // 0-59
    Hour       []int  // 0-23
    DayOfMonth []int  // 1-31
    Month      []int  // 1-12
    DayOfWeek  []int  // 0-6 (Sunday=0)
}

func ParseCron(expr string) (*CronExpression, error) {
    // Parse "0 9 * * 1-5" (9 AM weekdays)
    // Parse "*/15 * * * *" (every 15 minutes)
    // Parse "@daily", "@hourly" shortcuts
}

func (c *CronExpression) NextExecution(from time.Time) time.Time {
    // Calculate next execution time
}
```

### 2.2 Cron Scheduler Implementation

```go
// pkg/scheduler/cron.go
type CronScheduler struct {
    expression *cron.CronExpression
    timezone   *time.Location
    stopCh     chan struct{}
}

func NewCronScheduler(expr string, tz *time.Location) (*CronScheduler, error) {
    cronExpr, err := cron.ParseCron(expr)
    if err != nil {
        return nil, err
    }
    
    return &CronScheduler{
        expression: cronExpr,
        timezone:   tz,
        stopCh:     make(chan struct{}),
    }, nil
}

func (s *CronScheduler) Next() <-chan time.Time {
    ch := make(chan time.Time, 1)
    go func() {
        defer close(ch)
        
        for {
            now := time.Now().In(s.timezone)
            next := s.expression.NextExecution(now)
            
            select {
            case <-time.After(next.Sub(now)):
                ch <- next
            case <-s.stopCh:
                return
            }
        }
    }()
    return ch
}
```

### 2.3 CLI Integration

Add new subcommand `cron`:

```go
// pkg/cli/cli.go - Add to subcommand parsing
case "cron", "cr":
    config.Subcommand = "cron"
    // Parse --cron flag
    case "--cron":
        config.CronExpression = args[i+1]
        i++
    case "--timezone", "--tz":
        config.Timezone = args[i+1]
        i++
```

### 2.4 Usage Examples

```bash
# Run every weekday at 9 AM
rpr cron --cron "0 9 * * 1-5" -- ./backup.sh

# Run every 15 minutes
rpr cron --cron "*/15 * * * *" -- curl health-check.com

# Run daily at midnight in specific timezone
rpr cron --cron "@daily" --timezone "America/New_York" -- ./daily-report.sh

# Run with stop conditions
rpr cron --cron "0 */2 * * *" --for 24h -- ./periodic-task.sh
```

### 2.5 Testing Strategy

```go
func TestCronScheduler(t *testing.T) {
    tests := []struct {
        name       string
        expression string
        timezone   string
        from       time.Time
        expected   time.Time
    }{
        {
            name:       "daily at 9 AM",
            expression: "0 9 * * *",
            timezone:   "UTC",
            from:       time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
            expected:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
        },
        // More test cases...
    }
}
```

---

## 🔌 **Phase 3: Plugin System (Extensibility)** ✅ **COMPLETED**

**Goal**: Allow custom schedulers via plugin architecture
**Effort**: 2-3 weeks (Originally estimated)
**Priority**: MEDIUM
**Status**: **IMPLEMENTED** ✅

### 3.1 Plugin Interface Design

```go
// pkg/plugin/interface.go
type SchedulerPlugin interface {
    Name() string
    Version() string
    Description() string
    
    // Create scheduler with config
    NewScheduler(config map[string]interface{}) (scheduler.Scheduler, error)
    
    // Validate plugin configuration
    ValidateConfig(config map[string]interface{}) error
    
    // Plugin metadata
    ConfigSchema() *ConfigSchema
}

type ConfigSchema struct {
    Fields []ConfigField `json:"fields"`
}

type ConfigField struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"`        // "string", "int", "duration", "bool"
    Required    bool        `json:"required"`
    Default     interface{} `json:"default,omitempty"`
    Description string      `json:"description"`
}
```

### 3.2 Plugin Manager

```go
// pkg/plugin/manager.go
type PluginManager struct {
    plugins    map[string]SchedulerPlugin
    pluginDirs []string
}

func NewPluginManager(dirs []string) *PluginManager {
    return &PluginManager{
        plugins:    make(map[string]SchedulerPlugin),
        pluginDirs: dirs,
    }
}

func (pm *PluginManager) LoadPlugins() error {
    // Scan plugin directories for .so files
    // Load plugins using Go's plugin package
    // Register discovered plugins
}

func (pm *PluginManager) GetPlugin(name string) (SchedulerPlugin, error) {
    plugin, exists := pm.plugins[name]
    if !exists {
        return nil, fmt.Errorf("plugin %s not found", name)
    }
    return plugin, nil
}

func (pm *PluginManager) ListPlugins() []PluginInfo {
    // Return list of available plugins with metadata
}
```

### 3.3 Plugin Development Kit

Create example plugin and development guide:

```go
// examples/plugins/fibonacci/fibonacci.go
package main

import (
    "time"
    "github.com/swi/repeater/pkg/plugin"
    "github.com/swi/repeater/pkg/scheduler"
)

type FibonacciPlugin struct{}

func (p *FibonacciPlugin) Name() string { return "fibonacci" }
func (p *FibonacciPlugin) Version() string { return "1.0.0" }
func (p *FibonacciPlugin) Description() string { 
    return "Schedules executions using Fibonacci sequence intervals" 
}

func (p *FibonacciPlugin) NewScheduler(config map[string]interface{}) (scheduler.Scheduler, error) {
    baseInterval := config["base_interval"].(time.Duration)
    maxInterval := config["max_interval"].(time.Duration)
    
    return NewFibonacciScheduler(baseInterval, maxInterval), nil
}

func (p *FibonacciPlugin) ConfigSchema() *plugin.ConfigSchema {
    return &plugin.ConfigSchema{
        Fields: []plugin.ConfigField{
            {
                Name:        "base_interval",
                Type:        "duration",
                Required:    true,
                Description: "Base interval for Fibonacci sequence",
            },
            {
                Name:        "max_interval",
                Type:        "duration",
                Required:    false,
                Default:     "1h",
                Description: "Maximum interval (resets sequence)",
            },
        },
    }
}

// Plugin entry point
var Plugin FibonacciPlugin
```

### 3.4 CLI Integration

```bash
# List available plugins
rpr plugins list

# Show plugin info
rpr plugins info fibonacci

# Use plugin scheduler
rpr plugin fibonacci --base-interval 1s --max-interval 5m -- echo "Fibonacci timing"

# Plugin configuration via config file
[plugins.fibonacci]
base_interval = "1s"
max_interval = "5m"
```

### 3.5 Plugin Security & Sandboxing

```go
// pkg/plugin/security.go
type PluginSandbox struct {
    allowedSyscalls []string
    resourceLimits  ResourceLimits
    networkAccess   bool
}

type ResourceLimits struct {
    MaxMemory   int64
    MaxCPU      float64
    MaxFileSize int64
}

func (ps *PluginSandbox) Execute(plugin SchedulerPlugin, config map[string]interface{}) error {
    // Apply resource limits
    // Restrict system calls
    // Monitor plugin behavior
}
```

---

## 📅 **Phase 4: Future Enhancements (Long-term)**

### 4.1 Distributed Multi-Node Coordination

**Goal**: Coordinate scheduling across multiple machines
**Effort**: 4-6 weeks
**Priority**: LOW (Future)

#### Architecture Options:
1. **Consensus-based** (Raft/etcd)
2. **Message queue** (Redis/RabbitMQ)
3. **Database coordination** (PostgreSQL/MySQL)

#### Implementation Approach:
```go
// pkg/distributed/coordinator.go
type DistributedCoordinator interface {
    RegisterNode(nodeID string, capabilities NodeCapabilities) error
    RequestExecution(resourceID string, priority int) (*ExecutionLease, error)
    ReleaseExecution(leaseID string) error
    GetClusterState() (*ClusterState, error)
}

type NodeCapabilities struct {
    MaxConcurrentJobs int
    AvailableResources map[string]int
    Location          string
    Tags              []string
}
```

### 4.2 Advanced Scheduling Algorithms

**Goal**: Add more sophisticated scheduling options
**Examples**:
- **Genetic Algorithm Scheduler**: Evolve optimal timing patterns
- **Machine Learning Scheduler**: Learn from execution patterns
- **Chaos Scheduler**: Introduce controlled randomness for testing

### 4.3 Enhanced Observability

**Goal**: Advanced monitoring and alerting
**Features**:
- **Grafana dashboards**: Pre-built visualization
- **Alert manager integration**: Threshold-based alerts
- **Distributed tracing**: OpenTelemetry support
- **Log aggregation**: Structured logging with correlation IDs

---

## 📋 **Implementation Priorities & Effort Estimates**

### **Phase 1: Observability Integration** ✅ **COMPLETED**
- **Effort**: 1-2 days
- **Priority**: HIGH
- **ROI**: Very High (minimal effort, major functionality gain)
- **Dependencies**: None
- **Status**: **DONE** ✅

### **Phase 2: Cron-like Scheduling** ✅ **COMPLETED**
- **Effort**: 1-2 weeks  
- **Priority**: MEDIUM
- **ROI**: High (common use case, differentiating feature)
- **Dependencies**: None
- **Status**: **DONE** ✅

### **Phase 3: Plugin System** ✅ **COMPLETED**
- **Effort**: 2-3 weeks
- **Priority**: MEDIUM
- **ROI**: Medium (extensibility for power users)
- **Dependencies**: None
- **Status**: **DONE** ✅

### **Phase 4: Distributed Coordination**
- **Effort**: 4-6 weeks
- **Priority**: LOW
- **ROI**: Low (complex, niche use case)
- **Dependencies**: Requires significant architecture changes

---

## 🎯 **Recommended Implementation Order**

### **Immediate (Next Sprint)**
1. **Health Endpoints Integration** (4-6 hours)
2. **Metrics Server Integration** (4-6 hours)
3. **Integration Testing** (4-8 hours)

### **Short Term (Next Month)**
4. **Cron Expression Parser** (1 week)
5. **Cron Scheduler Implementation** (3-4 days)
6. **Cron CLI Integration & Testing** (2-3 days)

### **Medium Term (Next Quarter)**
7. **Plugin Interface Design** (1 week)
8. **Plugin Manager Implementation** (1 week)
9. **Example Plugins & Documentation** (1 week)

### **Long Term (Future Releases)**
10. **Distributed Coordination** (Major feature, separate planning needed)
11. **Advanced Scheduling Algorithms** (Research & experimentation)
12. **Enhanced Observability** (Incremental improvements)

---

## 🧪 **Testing Strategy**

### **Integration Tests**
- End-to-end tests for each new feature
- Performance tests for scheduling accuracy
- Stress tests for concurrent execution

### **Compatibility Tests**
- Ensure new features don't break existing functionality
- Test configuration file backward compatibility
- Verify CLI argument parsing remains stable

### **Plugin Tests**
- Plugin loading and unloading
- Plugin security and sandboxing
- Plugin configuration validation

---

## 📚 **Documentation Plan**

### **User Documentation**
- Update README with new features
- Add configuration examples
- Create plugin development guide

### **Developer Documentation**
- Architecture decision records (ADRs)
- Plugin API documentation
- Contributing guidelines for new schedulers

### **Operational Documentation**
- Deployment guides for observability
- Monitoring and alerting setup
- Troubleshooting guides

---

## 🎉 **Success Metrics**

### **Phase 1 Success Criteria**
- Health endpoints respond correctly during execution
- Metrics are collected and exposed via HTTP
- No performance impact on core execution

### **Phase 2 Success Criteria**
- Cron expressions parse correctly
- Scheduled executions happen at precise times
- Timezone handling works across different regions

### **Phase 3 Success Criteria**
- Plugins can be loaded dynamically
- Custom schedulers work seamlessly with existing CLI
- Plugin development is well-documented and accessible

This implementation plan provides a clear roadmap for enhancing Repeater while maintaining its production-ready core functionality.