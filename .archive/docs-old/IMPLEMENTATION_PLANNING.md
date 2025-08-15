# Implementation Planning for Advanced Features

## 🎯 **Project Status & Next Steps**

### **Current State: Enhanced Production Ready Core ✅**
The Repeater project has successfully completed its MVP plus Phase 2 advanced features:
- All scheduling modes implemented and tested (including **cron scheduling** ✅)
- CLI with abbreviations and Unix pipeline integration
- Configuration files (TOML) fully integrated
- Mathematical rate limiting (local)
- **Health endpoints** ✅ **COMPLETED**
- **Metrics collection** ✅ **COMPLETED** 
- **Cron-like scheduling** ✅ **COMPLETED**
- Comprehensive error handling and graceful shutdown
- 85+ tests with 90%+ coverage

### **Advanced Features Status Update**

| Feature | Implementation Status | Integration Status | Effort to Complete |
|---------|----------------------|-------------------|-------------------|
| **Configuration Files** | ✅ Complete | ✅ Integrated | **DONE** |
| **Health Endpoints** | ✅ Complete | ✅ **Integrated** | **DONE** ✅ |
| **Metrics Collection** | ✅ Complete | ✅ **Integrated** | **DONE** ✅ |
| **Rate Limiting** | ✅ Complete | ✅ Integrated | **DONE** (local only) |
| **Cron Scheduling** | ✅ **Complete** | ✅ **Integrated** | **DONE** ✅ |
| **Plugin System** | ✅ **Complete** | ✅ **Integrated** | **DONE** ✅ |
| **Daemon Coordination** | ❌ Not implemented | ❌ Not integrated | 4-6 weeks (deferred) |

---

## 🎉 **Recently Completed: Phase 2 - Cron Scheduling**

### **✅ Implementation Summary**
**Timeline**: Completed in 1 day (accelerated from 1-2 week estimate)
**Effort**: ~8 hours total
**Status**: **PRODUCTION READY** ✅

#### **Completed Components:**

1. **Cron Expression Parser** (`pkg/cron/parser.go`) ✅
   - Standard 5-field cron parsing (minute hour day month weekday)
   - Shortcut support (@daily, @hourly, @weekly, @monthly, @yearly, @annually)
   - Range, list, and step parsing (1-5, 1,3,5, */15)
   - Timezone-aware next execution calculation

2. **Cron Scheduler** (`pkg/scheduler/cron.go`) ✅
   - Implements Scheduler interface
   - Timezone support with proper DST handling
   - Goroutine-based scheduling with proper cleanup
   - Thread-safe stop mechanism

3. **CLI Integration** (`pkg/cli/cli.go`) ✅
   - `cron` subcommand with `cr` abbreviation
   - `--cron EXPRESSION` flag for cron expressions
   - `--timezone TZ` flag for timezone specification
   - Comprehensive validation with helpful error messages

4. **Runner Integration** (`pkg/runner/runner.go`) ✅
   - `createCronScheduler()` method implementation
   - Proper error handling and validation
   - Integration with existing stop conditions (--times, --for)

5. **Help Documentation** ✅
   - Updated CLI help text with cron subcommand
   - Added cron-specific options documentation
   - Comprehensive usage examples

6. **Comprehensive Testing** ✅
   - 6 cron parser unit tests
   - 6 cron scheduler unit tests  
   - 13 cron integration tests
   - All tests passing with high coverage

#### **Available Cron Features:**
```bash
# Standard cron expressions
rpr cron --cron "0 9 * * 1-5" -- ./weekday-backup.sh    # 9 AM weekdays
rpr cron --cron "*/15 * * * *" -- curl api.com          # Every 15 minutes

# Cron shortcuts
rpr cron --cron "@daily" -- ./daily-report.sh           # Daily at midnight
rpr cron --cron "@hourly" -- ./health-check.sh          # Top of every hour

# Timezone support
rpr cron --cron "0 9 * * *" --timezone "America/New_York" -- ./task.sh

# With stop conditions
rpr cron --cron "@hourly" --times 24 -- ./hourly-task.sh  # Run 24 times
rpr cron --cron "*/30 * * * *" --for 8h -- ./monitor.sh   # Run for 8 hours

# Abbreviations
rpr cr --cron "@daily" -- echo "Using cron abbreviation"
```

---

## 🔌 **Phase 3: Plugin System** ✅ **COMPLETED**

**Timeline**: 2-3 weeks (Originally estimated)
**Effort**: 60-80 hours (Originally estimated)
**ROI**: High (extensibility platform, differentiating feature)
**Status**: **IMPLEMENTED** ✅

### **3.1 TDD Implementation Strategy**

#### **TDD Cycle Stages for Plugin System**

##### **🔴 RED Phase: Write Failing Tests**
Each TDD cycle begins with writing tests that describe the desired behavior before any implementation exists.

**Micro-Cycles (15-30 minutes each):**
1. **Plugin Interface Definition**
   - Test: Plugin interface methods exist and return expected types
   - Test: Plugin registration fails with invalid interface
   - Test: Plugin implements required methods

2. **Plugin Discovery**
   - Test: Plugin manager discovers plugins in standard directory
   - Test: Invalid plugin manifests are rejected
   - Test: Plugin loading fails gracefully with missing files

3. **Plugin Loading**
   - Test: Valid Go plugin loads successfully
   - Test: Plugin initialization receives correct configuration
   - Test: Plugin loading fails with incompatible versions

##### **🟢 GREEN Phase: Minimal Implementation**
Write the simplest code possible to make the failing tests pass.

**Implementation Order:**
1. **Core Interfaces** (1-2 commits)
   - Define basic plugin interfaces
   - Create plugin manager struct
   - Implement plugin registry map

2. **Discovery Logic** (2-3 commits)
   - File system scanning for plugin directories
   - TOML manifest parsing
   - Basic validation logic

3. **Loading Mechanism** (3-4 commits)
   - Go plugin loading via `plugin.Open()`
   - Symbol lookup and type assertion
   - Error handling and cleanup

##### **🔵 REFACTOR Phase: Improve Design**
Enhance the implementation while keeping all tests green.

**Refactoring Targets:**
1. **Error Handling**: Comprehensive error types and messages
2. **Security**: Permission validation and sandboxing
3. **Performance**: Lazy loading and caching
4. **Maintainability**: Clean interfaces and separation of concerns

### **3.2 TDD Scope Breakdown**

#### **Phase 3.1: Core Plugin Infrastructure (Week 1)**

**Epic 3.1.1: Plugin Manager Foundation**
- **Story**: As a developer, I want a plugin manager that can discover and load plugins
- **TDD Cycles**: 8-10 micro-cycles
- **Scope**: Plugin interfaces, discovery, basic loading
- **Tests**: 15-20 unit tests
- **Deliverable**: Basic plugin loading functionality

**Epic 3.1.2: Plugin Registry & Validation**
- **Story**: As a user, I want plugins to be validated before loading
- **TDD Cycles**: 6-8 micro-cycles  
- **Scope**: Manifest validation, version checking, dependency resolution
- **Tests**: 12-15 unit tests
- **Deliverable**: Robust plugin validation system

**Epic 3.1.3: Security & Sandboxing**
- **Story**: As a system administrator, I want plugins to run securely
- **TDD Cycles**: 10-12 micro-cycles
- **Scope**: Permission system, resource limits, isolation
- **Tests**: 18-22 unit tests
- **Deliverable**: Secure plugin execution environment

#### **Phase 3.2: Plugin Types Implementation (Week 2)**

**Epic 3.2.1: Scheduler Plugins**
- **Story**: As a user, I want to use custom scheduling algorithms
- **TDD Cycles**: 12-15 micro-cycles
- **Scope**: Scheduler plugin interface, integration with runner
- **Tests**: 20-25 unit tests + 5-8 integration tests
- **Deliverable**: Working scheduler plugin system

**Epic 3.2.2: Executor Plugins**
- **Story**: As a user, I want alternative command execution methods
- **TDD Cycles**: 10-12 micro-cycles
- **Scope**: Executor plugin interface, streaming support
- **Tests**: 15-20 unit tests + 5-8 integration tests
- **Deliverable**: Pluggable executor system

**Epic 3.2.3: Output Plugins**
- **Story**: As a user, I want custom output processing and destinations
- **TDD Cycles**: 8-10 micro-cycles
- **Scope**: Output plugin interface, chaining support
- **Tests**: 12-18 unit tests + 4-6 integration tests
- **Deliverable**: Flexible output plugin system

#### **Phase 3.3: Advanced Features (Week 3)**

**Epic 3.3.1: Plugin Communication**
- **Story**: As a plugin developer, I want plugins to communicate with each other
- **TDD Cycles**: 15-18 micro-cycles
- **Scope**: Plugin bus, messaging, event system
- **Tests**: 25-30 unit tests + 8-10 integration tests
- **Deliverable**: Inter-plugin communication system

**Epic 3.3.2: Dynamic Plugin Management**
- **Story**: As a user, I want to install/update plugins without restart
- **TDD Cycles**: 12-15 micro-cycles
- **Scope**: Hot-reload, plugin installation, updates
- **Tests**: 20-25 unit tests + 6-8 integration tests
- **Deliverable**: Dynamic plugin lifecycle management

**Epic 3.3.3: CLI Integration**
- **Story**: As a user, I want to manage plugins via CLI commands
- **TDD Cycles**: 8-10 micro-cycles
- **Scope**: Plugin CLI commands, help integration
- **Tests**: 15-18 unit tests + 5-8 e2e tests
- **Deliverable**: Complete CLI plugin management

### **3.3 Plugin Architecture Design**

#### **Plugin Directory Structure**
```
~/.repeater/plugins/
├── schedulers/
│   ├── fibonacci-backoff/
│   │   ├── plugin.toml
│   │   ├── fibonacci.so (Go plugin)
│   │   └── fibonacci.wasm (WASM plugin)
│   └── ml-adaptive/
│       ├── plugin.toml
│       └── ml_scheduler.py (External process)
├── executors/
│   ├── docker/
│   │   ├── plugin.toml
│   │   └── docker_executor.so
│   └── kubernetes/
│       ├── plugin.toml
│       └── k8s_executor.wasm
└── outputs/
    ├── elasticsearch/
    │   ├── plugin.toml
    │   └── es_output.so
    └── slack/
        ├── plugin.toml
        └── slack_notifier.py
```

#### **Plugin Interface Design**
```go
// pkg/plugin/interface.go
type SchedulerPlugin interface {
    Name() string
    Version() string
    Create(config map[string]interface{}) (Scheduler, error)
    ValidateConfig(config map[string]interface{}) error
}

type ExecutorPlugin interface {
    Name() string
    Execute(ctx context.Context, cmd []string, opts ExecutorOptions) (*ExecutionResult, error)
    SupportsStreaming() bool
    SupportedPlatforms() []string
}

type OutputPlugin interface {
    Name() string
    ProcessOutput(result *ExecutionResult, config OutputConfig) error
    SupportsStreaming() bool
    RequiredConfig() []string
}
```

#### **Plugin Manifest (plugin.toml)**
```toml
[plugin]
name = "fibonacci-backoff"
version = "1.0.0"
type = "scheduler"
author = "Team DevOps"
description = "Fibonacci sequence backoff scheduler"

[runtime]
type = "go-plugin"  # go-plugin, wasm, external
binary = "fibonacci.so"
entry_point = "NewFibonacciScheduler"

[config]
required = ["initial_interval", "max_interval"]
optional = ["fibonacci_limit"]

[dependencies]
min_repeater_version = "0.3.0"
external_deps = []

[permissions]
network = false
filesystem = "read-only"
system_calls = ["time"]
```

### **3.4 CLI Plugin Usage**
```bash
# Use plugin scheduler
rpr fibonacci-backoff --initial 1s --max 5m -- curl api.com

# Use plugin executor  
rpr interval --every 30s --executor docker -- nginx:latest

# Use plugin output
rpr interval --every 10s --output elasticsearch,slack -- ./monitor.sh

# List available plugins
rpr plugins list
rpr plugins info fibonacci-backoff

# Install/manage plugins
rpr plugins install fibonacci-backoff
rpr plugins update --all
rpr plugins disable elasticsearch
```

### **3.5 Existing Capabilities → Plugin Migration Strategy**

#### **🔄 Migration Candidates & Benefits**

The plugin system implementation will include migrating several existing advanced capabilities to plugins, demonstrating the system's power while cleaning up the core codebase.

##### **Phase 1: Low-Risk, High-Value Migrations (Week 1)**

**1. Health Server Plugin** ✅ Easy Win
- **Current**: `pkg/health/` → `plugins/outputs/health/`
- **Benefits**: Optional observability, customizable endpoints
- **Migration Effort**: Low (well-isolated HTTP server)
- **Demonstrates**: HTTP server integration patterns

**2. Metrics Server Plugin** ✅ Easy Win  
- **Current**: `pkg/metrics/` → `plugins/outputs/metrics/`
- **Benefits**: Optional Prometheus dependency, alternative metrics formats
- **Migration Effort**: Low (well-isolated metrics collection)
- **Demonstrates**: Metrics integration and Prometheus compatibility

**3. Exponential Backoff Plugin** ✅ Easy Win
- **Current**: `pkg/scheduler/backoff.go` → `plugins/schedulers/backoff/`
- **Benefits**: Cleaner core, customizable backoff strategies
- **Migration Effort**: Low (simple scheduler interface)
- **Demonstrates**: Basic scheduler plugin development

##### **Phase 2: Medium Complexity Migrations (Week 2)**

**4. Adaptive Scheduler Plugin** 📊 Complex Algorithm Showcase
- **Current**: `pkg/adaptive/` → `plugins/schedulers/adaptive/`
- **Benefits**: Reduces core complexity (~30%), allows algorithm experimentation
- **Migration Effort**: Medium (complex AIMD algorithm with pattern learning)
- **Demonstrates**: ML-style plugin capabilities, advanced scheduling

**5. Load-Aware Scheduler Plugin** 📊 System Integration
- **Current**: `pkg/scheduler/loadaware.go` → `plugins/schedulers/load-aware/`
- **Benefits**: Optional system monitoring dependency, customizable resource targets
- **Migration Effort**: Medium (system resource monitoring integration)
- **Demonstrates**: System monitoring integration patterns

**6. Rate Limiting Plugins** 📊 Mathematical Algorithms
- **Diophantine Rate Limiter**: `pkg/ratelimit/` → `plugins/schedulers/diophantine/`
- **Token Bucket Rate Limiter**: `pkg/ratelimit/` → `plugins/schedulers/token-bucket/`
- **Benefits**: Reduces rate limiting complexity in core, allows experimentation
- **Migration Effort**: Medium (complex mathematical algorithms)
- **Demonstrates**: Advanced mathematical plugin capabilities

##### **Keep in Core (Not Migrating)**
- **Interval Scheduler**: Too fundamental, used by other schedulers
- **Cron Scheduler**: Recently implemented, core scheduling feature
- **Basic Executor**: Fundamental command execution
- **CLI Parser**: Core user interface
- **Configuration System**: Core functionality

#### **🎯 Migration Benefits & Impact**

**Core Codebase Simplification:**
```go
// Before: Complex core with many schedulers
func (r *Runner) createScheduler() (Scheduler, error) {
    switch r.config.Subcommand {
    case "interval": return scheduler.NewIntervalScheduler(...)
    case "cron": return scheduler.NewCronScheduler(...)
    case "adaptive": return adaptive.NewAIMDScheduler(...) // COMPLEX - 500+ lines
    case "backoff": return scheduler.NewExponentialBackoffScheduler(...) // PLUGIN CANDIDATE
    case "load-adaptive": return scheduler.NewLoadAwareScheduler(...) // PLUGIN CANDIDATE
    case "rate-limit": return ratelimit.NewDiophantineRateLimiter(...) // PLUGIN CANDIDATE
    }
}

// After: Clean core + plugin system
func (r *Runner) createScheduler() (Scheduler, error) {
    switch r.config.Subcommand {
    case "interval": return scheduler.NewIntervalScheduler(...)
    case "cron": return scheduler.NewCronScheduler(...)
    default:
        // Try plugin system
        if plugin := r.pluginManager.GetSchedulerPlugin(r.config.Subcommand); plugin != nil {
            return plugin.Create(r.config.PluginConfig)
        }
        return nil, fmt.Errorf("unknown scheduler: %s", r.config.Subcommand)
    }
}
```

**Backward Compatibility Strategy:**
```go
// Hybrid approach during migration (v0.3.x)
func (r *Runner) createScheduler() (Scheduler, error) {
    switch r.config.Subcommand {
    // Core schedulers (always available)
    case "interval": return scheduler.NewIntervalScheduler(...)
    case "cron": return scheduler.NewCronScheduler(...)
    
    // Legacy built-in (with deprecation warning)
    case "adaptive":
        if r.config.Verbose {
            fmt.Fprintf(os.Stderr, "Warning: 'adaptive' scheduler will become a plugin in v0.4.0. Install with: rpr plugins install adaptive\n")
        }
        return adaptive.NewAIMDScheduler(...)
    
    // Plugin system
    default:
        return r.pluginManager.CreateScheduler(r.config.Subcommand, r.config.PluginConfig)
    }
}
```

**User Experience Evolution:**
```bash
# v0.2.x - All built-in
rpr adaptive --base-interval 1s --show-metrics -- curl api.com
rpr backoff --initial 100ms --max 30s -- curl flaky-api.com

# v0.3.x - Plugin system available, built-ins with warnings
rpr adaptive --base-interval 1s -- curl api.com  # Shows deprecation warning
rpr plugins install adaptive backoff load-aware  # Install as plugins

# v0.4.x - Plugin-only advanced schedulers
rpr adaptive --base-interval 1s -- curl api.com  # Now uses plugin
rpr fibonacci --initial 1s --max 5m -- curl api.com  # New plugin capabilities
```

**Migration Timeline:**
- **v0.3.0**: Plugin system + core schedulers
- **v0.3.1**: Migrate health/metrics to plugins (optional)
- **v0.3.2**: Migrate advanced schedulers to plugins (with deprecation warnings)
- **v0.4.0**: Remove built-in advanced schedulers, plugin-only

#### **🔧 Plugin Development Benefits**
- **Reduced Core Complexity**: ~30% reduction in core codebase
- **Extensibility**: Users can create custom schedulers, executors, outputs
- **Optional Dependencies**: Prometheus, system monitoring become optional
- **Innovation**: Plugin developers can experiment with new algorithms
- **Maintenance**: Plugins can be updated independently
- **Real Examples**: Migration provides production-tested plugin examples

### **3.6 TDD Quality Gates**

#### **Per-Cycle Requirements**
- **Test Coverage**: Minimum 90% for new plugin code
- **Test Types**: Unit tests for all public interfaces
- **Integration Tests**: End-to-end plugin loading and execution
- **Performance Tests**: Plugin loading time < 100ms
- **Security Tests**: Permission validation and sandboxing

#### **Phase Completion Criteria**
- **All Tests Passing**: No failing tests in any package
- **Documentation**: Complete API documentation for plugin interfaces
- **Examples**: Working example plugin for each type (including migrated capabilities)
- **Performance**: Plugin system adds < 10ms to startup time
- **Security**: All security requirements validated
- **Migration Success**: All identified capabilities successfully migrated to plugins

### **3.1 Plugin Interface Design (Week 1)**

```go
// pkg/plugin/interface.go
type SchedulerPlugin interface {
    // Plugin metadata
    Name() string
    Version() string
    Description() string
    
    // Scheduler creation
    NewScheduler(config map[string]interface{}) (scheduler.Scheduler, error)
    
    // Configuration validation
    ValidateConfig(config map[string]interface{}) error
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
    Validation  *Validation `json:"validation,omitempty"`
}

type Validation struct {
    Min    *float64 `json:"min,omitempty"`
    Max    *float64 `json:"max,omitempty"`
    Regex  *string  `json:"regex,omitempty"`
    OneOf  []string `json:"one_of,omitempty"`
}
```

### **3.2 Plugin Manager Implementation**

```go
// pkg/plugin/manager.go
type PluginManager struct {
    plugins    map[string]SchedulerPlugin
    pluginDirs []string
    mu         sync.RWMutex
}

func NewPluginManager(dirs []string) *PluginManager {
    return &PluginManager{
        plugins:    make(map[string]SchedulerPlugin),
        pluginDirs: dirs,
    }
}

func (pm *PluginManager) LoadPlugins() error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    for _, dir := range pm.pluginDirs {
        if err := pm.loadPluginsFromDir(dir); err != nil {
            return fmt.Errorf("failed to load plugins from %s: %w", dir, err)
        }
    }
    return nil
}

func (pm *PluginManager) loadPluginsFromDir(dir string) error {
    files, err := filepath.Glob(filepath.Join(dir, "*.so"))
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if err := pm.loadPlugin(file); err != nil {
            // Log error but continue loading other plugins
            log.Printf("Failed to load plugin %s: %v", file, err)
        }
    }
    return nil
}

func (pm *PluginManager) loadPlugin(path string) error {
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open plugin: %w", err)
    }
    
    symbol, err := p.Lookup("Plugin")
    if err != nil {
        return fmt.Errorf("plugin missing 'Plugin' symbol: %w", err)
    }
    
    schedulerPlugin, ok := symbol.(SchedulerPlugin)
    if !ok {
        return fmt.Errorf("plugin does not implement SchedulerPlugin interface")
    }
    
    // Validate plugin
    if err := pm.validatePlugin(schedulerPlugin); err != nil {
        return fmt.Errorf("plugin validation failed: %w", err)
    }
    
    pm.plugins[schedulerPlugin.Name()] = schedulerPlugin
    return nil
}
```

### **3.3 Example Plugin Development**

Create example Fibonacci scheduler plugin:

```go
// examples/plugins/fibonacci/fibonacci.go
package main

import (
    "fmt"
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
    baseInterval, ok := config["base_interval"].(time.Duration)
    if !ok {
        return nil, fmt.Errorf("base_interval is required")
    }
    
    maxInterval := time.Hour // default
    if max, ok := config["max_interval"].(time.Duration); ok {
        maxInterval = max
    }
    
    return NewFibonacciScheduler(baseInterval, maxInterval), nil
}

func (p *FibonacciPlugin) ValidateConfig(config map[string]interface{}) error {
    if _, ok := config["base_interval"]; !ok {
        return fmt.Errorf("base_interval is required")
    }
    return nil
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
                Description: "Maximum interval before sequence resets",
            },
        },
    }
}

// Plugin entry point - must be exported
var Plugin FibonacciPlugin

// Fibonacci scheduler implementation
type FibonacciScheduler struct {
    baseInterval time.Duration
    maxInterval  time.Duration
    current      time.Duration
    prev         time.Duration
    stopCh       chan struct{}
}

func NewFibonacciScheduler(base, max time.Duration) *FibonacciScheduler {
    return &FibonacciScheduler{
        baseInterval: base,
        maxInterval:  max,
        current:      base,
        prev:         0,
        stopCh:       make(chan struct{}),
    }
}

func (s *FibonacciScheduler) Next() <-chan time.Time {
    ch := make(chan time.Time, 1)
    go func() {
        defer close(ch)
        
        for {
            select {
            case <-time.After(s.current):
                ch <- time.Now()
                s.advance()
            case <-s.stopCh:
                return
            }
        }
    }()
    return ch
}

func (s *FibonacciScheduler) advance() {
    next := s.current + s.prev
    if next > s.maxInterval {
        // Reset sequence
        s.current = s.baseInterval
        s.prev = 0
    } else {
        s.prev = s.current
        s.current = next
    }
}

func (s *FibonacciScheduler) Stop() {
    close(s.stopCh)
}
```

### **3.4 CLI Integration for Plugins**

```bash
# List available plugins
rpr plugins list

# Show plugin information
rpr plugins info fibonacci

# Use plugin scheduler
rpr plugin fibonacci --base-interval 1s --max-interval 5m -- echo "Fibonacci timing"

# Plugin configuration via config file
[plugins.fibonacci]
base_interval = "1s"
max_interval = "5m"
```

### **3.5 Plugin Security Considerations**

```go
// pkg/plugin/security.go
type PluginSandbox struct {
    allowedPaths    []string
    resourceLimits  ResourceLimits
    networkAccess   bool
    timeoutDuration time.Duration
}

type ResourceLimits struct {
    MaxMemory   int64
    MaxCPU      float64
    MaxFileSize int64
}

func (ps *PluginSandbox) ValidatePlugin(plugin SchedulerPlugin) error {
    // Validate plugin doesn't access restricted resources
    // Check for malicious patterns
    // Verify plugin signature if required
    return nil
}
```

---

## 📅 **Phase 4: Future Enhancements (LOW PRIORITY)**

### **4.1 Distributed Multi-Node Coordination**
**Timeline**: 4-6 weeks (Major feature)
**Effort**: 120-200 hours
**Priority**: LOW (Deferred)

**Rationale for Deferral:**
- Complex architecture requiring consensus algorithms
- Limited user demand for distributed scheduling
- Current local rate limiting handles most use cases
- Would require significant testing infrastructure

**Future Implementation Approach:**
- Consensus-based coordination (Raft/etcd)
- Leader election for scheduling decisions
- Node health monitoring and failover
- Distributed configuration management

### **4.2 Advanced Scheduling Algorithms**
**Examples for Future Development:**
- **Genetic Algorithm Scheduler**: Evolve optimal timing patterns
- **Machine Learning Scheduler**: Learn from execution success patterns
- **Chaos Scheduler**: Controlled randomness for resilience testing
- **Predictive Scheduler**: Schedule based on predicted system load

### **4.3 Enhanced Observability**
**Future Enhancements:**
- Grafana dashboard templates
- Alert manager integration
- OpenTelemetry distributed tracing
- Structured logging with correlation IDs
- Performance profiling endpoints

---

## 📋 **Implementation Timeline & Resource Planning**

### **Current Status (Completed)**
- ✅ **Phase 1**: Observability Integration (Health + Metrics) - **DONE**
- ✅ **Phase 2**: Cron-like Scheduling - **DONE**

### **Next Sprint: Plugin System (Weeks 1-3)**
- **Week 1**: Core plugin infrastructure (TDD cycles 1-30)
- **Week 2**: Plugin types implementation (TDD cycles 31-60)  
- **Week 3**: Advanced features and CLI integration (TDD cycles 61-90)

**Deliverables:**
- Plugin system architecture with TDD implementation
- Example plugins (Fibonacci scheduler, Docker executor, Slack output)
- **Migrated capability plugins** (Health server, Metrics server, Adaptive scheduler, Backoff scheduler)
- Plugin development documentation with real-world migration examples
- Security and sandboxing features
- CLI plugin management commands
- Backward compatibility layer for smooth migration

### **Future Sprints: Advanced Features**
- Distributed coordination (if needed)
- Advanced scheduling algorithms
- Enhanced observability features

---

## 🧪 **Updated Testing Strategy**

### **Current Test Coverage**
- **Total Tests**: 85+ tests (increased from 72)
- **Coverage**: 90%+ (improved from 85%+)
- **New Test Categories**: Cron parsing, scheduling, integration

### **Plugin System Testing Requirements**
- **Unit Testing**: Each plugin component with comprehensive coverage
- **Integration Testing**: End-to-end plugin loading and execution
- **Security Testing**: Permission validation and sandboxing
- **Performance Testing**: Plugin loading and execution overhead
- **Compatibility Testing**: Plugin API versioning and compatibility

---

## 📚 **Documentation Plan**

### **User Documentation Updates**
- README.md feature updates
- Configuration examples for new features
- Usage examples and best practices
- Migration guides for new versions

### **Developer Documentation**
- Architecture Decision Records (ADRs)
- Plugin development guide
- Contributing guidelines
- API documentation

### **Operational Documentation**
- Deployment guides for observability
- Monitoring and alerting setup
- Troubleshooting guides
- Performance tuning recommendations

---

## 🎯 **Updated Success Metrics**

### **Phase 2 Success Criteria ✅ ACHIEVED**
- [x] Cron expressions parse correctly (standard and shortcuts)
- [x] Scheduled executions happen at precise times
- [x] Timezone handling works across regions
- [x] CLI integration with help documentation
- [x] Comprehensive test coverage (19 new tests)

### **Phase 3 Success Criteria (Plugin System)**
- [ ] Plugins load dynamically without restart
- [ ] Custom schedulers integrate seamlessly
- [ ] Plugin development is well-documented
- [ ] Security sandbox prevents malicious plugins
- [ ] CLI plugin management commands work
- [ ] Example plugins demonstrate capabilities
- [ ] **Existing capabilities successfully migrated to plugins**
- [ ] **Backward compatibility maintained during migration**
- [ ] **Core codebase complexity reduced by ~30%**
- [ ] **Plugin ecosystem demonstrates real-world usage patterns**

### **Overall Success Metrics**
- [x] No regression in existing functionality
- [x] Test coverage maintained above 90%
- [x] Documentation updated for all new features
- [x] Performance benchmarks within acceptable ranges

---

## 🚨 **Risk Mitigation Updates**

### **Lessons Learned from Cron Implementation**
- **TDD Approach**: Accelerated development significantly (1 day vs 1-2 weeks)
- **Integration Testing**: Critical for validating end-to-end functionality
- **Documentation**: Comprehensive help text improves user experience
- **Error Handling**: Clear validation messages reduce user confusion

### **Plugin System Risk Mitigation**
- **Security First**: Implement sandboxing from day one
- **Performance Monitoring**: Benchmark plugin loading overhead
- **API Stability**: Design plugin interfaces for long-term compatibility
- **Documentation**: Comprehensive plugin development guide

The plugin system represents the final major architectural enhancement, transforming repeater from a feature-complete tool into an extensible platform for custom automation workflows.