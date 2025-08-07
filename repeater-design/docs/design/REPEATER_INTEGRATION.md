# Repeater Integration with Patience Ecosystem

## Integration Overview

Repeater is designed as a complementary tool to patience, sharing infrastructure while serving distinct use cases. This document outlines the integration strategy, shared components, and coordination mechanisms between the two tools.

## Core Integration Principles

### 1. Complementary Functionality
- **patience**: Retry until success (stops on success)
- **repeater**: Continuous execution (continues on success)
- **Combined Usage**: `rpr interval -- patience exponential -- command`

### 2. Shared Infrastructure
- **Daemon**: Single shared daemon for rate limiting coordination
- **Configuration**: Compatible configuration formats and environment variables
- **Libraries**: Reuse core packages for consistency and maintenance

### 3. Independent Operation
- **Standalone**: Each tool works independently without the other
- **Graceful Degradation**: Functionality preserved when shared components unavailable
- **Separate Binaries**: Distinct executables with separate release cycles

## Shared Components

### 1. Daemon Integration

#### Shared Daemon Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                    Shared Patience Daemon                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   Rate Limiter  │    │   Resource       │                   │
│  │   Coordinator   │    │   Manager        │                   │
│  └─────────────────┘    └──────────────────┘                   │
│           │                       │                            │
│           ▼                       ▼                            │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   Diophantine   │    │   Metrics        │                   │
│  │   Algorithms    │    │   Aggregator     │                   │
│  └─────────────────┘    └──────────────────┘                   │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                        Client Connections                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐                   │
│  │   patience      │    │   repeater       │                   │
│  │   clients       │    │   clients        │                   │
│  └─────────────────┘    └──────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
```

#### Daemon Protocol Extensions
```go
// Extend existing protocol for repeater-specific operations
type RepeaterRequest struct {
    Type        string            `json:"type"`
    ResourceID  string            `json:"resource_id"`
    RateLimit   *RateLimitConfig  `json:"rate_limit,omitempty"`
    Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
    Metrics     *MetricsRequest   `json:"metrics,omitempty"`
}

type ScheduleConfig struct {
    Type        string        `json:"type"`         // "interval", "count", "duration"
    Interval    time.Duration `json:"interval"`
    Count       int64         `json:"count,omitempty"`
    Duration    time.Duration `json:"duration,omitempty"`
    Coordination bool         `json:"coordination"` // Enable multi-instance coordination
}

type MetricsRequest struct {
    Action      string        `json:"action"`       // "register", "update", "query"
    ExecutionID string        `json:"execution_id"`
    Stats       *ExecutionStats `json:"stats,omitempty"`
}
```

#### Resource Coordination
```go
// pkg/daemon/resource_coordinator.go
type ResourceCoordinator struct {
    resources map[string]*SharedResource
    mutex     sync.RWMutex
}

type SharedResource struct {
    ID              string
    Type            ResourceType
    RateLimit       *RateLimitState
    ActiveClients   map[string]*ClientState
    TotalExecutions int64
    LastActivity    time.Time
}

type ResourceType int
const (
    ResourceTypeAPI ResourceType = iota
    ResourceTypeDatabase
    ResourceTypeFileSystem
    ResourceTypeNetwork
)

type ClientState struct {
    ClientID        string
    Tool            string // "patience" or "repeater"
    LastExecution   time.Time
    ExecutionCount  int64
    RateLimitTokens int64
}
```

### 2. Configuration Integration

#### Shared Configuration Format
```toml
# ~/.config/patience/config.toml (shared by both tools)
[daemon]
socket_path = "/var/run/patience/daemon.sock"
log_level = "info"
metrics_enabled = true

[rate_limiting]
default_algorithm = "diophantine"
coordination_enabled = true

[repeater]
default_interval = "30s"
default_timeout = "60s"
continue_on_error = true

[patience]
default_strategy = "exponential"
max_attempts = 10
```

#### Environment Variable Compatibility
```bash
# Shared variables
PATIENCE_DAEMON_SOCKET=/var/run/patience/daemon.sock
PATIENCE_CONFIG_FILE=~/.config/patience/config.toml
PATIENCE_LOG_LEVEL=info

# Tool-specific variables (with fallbacks)
RPR_DAEMON_SOCKET=${PATIENCE_DAEMON_SOCKET}
RPR_CONFIG_FILE=${PATIENCE_CONFIG_FILE}
RPR_LOG_LEVEL=${PATIENCE_LOG_LEVEL}
```

### 3. Shared Libraries

#### Package Reuse Strategy
```go
// Repeater imports from patience
import (
    // Core shared packages
    "github.com/shaneisley/patience/pkg/backoff"
    "github.com/shaneisley/patience/pkg/daemon"
    "github.com/shaneisley/patience/pkg/config"
    "github.com/shaneisley/patience/pkg/metrics"
    
    // Repeater-specific packages
    "github.com/shaneisley/repeater/pkg/scheduler"
    "github.com/shaneisley/repeater/pkg/output"
    "github.com/shaneisley/repeater/pkg/conditions"
)
```

#### Shared Package Modifications
```go
// pkg/daemon/client.go - Enhanced for repeater support
type DaemonClient struct {
    conn     net.Conn
    encoder  *json.Encoder
    decoder  *json.Decoder
    clientID string
    tool     string // "patience" or "repeater"
}

func NewDaemonClient(tool string) (*DaemonClient, error) {
    // Tool-specific initialization
    client := &DaemonClient{
        clientID: generateClientID(),
        tool:     tool,
    }
    return client, client.connect()
}

// Tool-specific methods
func (c *DaemonClient) RegisterRepeaterSchedule(resourceID string, config *ScheduleConfig) error
func (c *DaemonClient) UpdateRepeaterMetrics(resourceID string, stats *ExecutionStats) error
func (c *DaemonClient) QueryResourceState(resourceID string) (*ResourceState, error)
```

## Rate Limiting Coordination

### 1. Multi-Instance Coordination

#### Scenario: Multiple Repeater Instances
```bash
# Terminal 1: API monitoring
rpr interval --every 30s --resource-id api-monitor --daemon -- curl api.example.com

# Terminal 2: API load testing  
rpr count --times 1000 --every 1s --resource-id api-load --daemon -- curl api.example.com

# Terminal 3: Combined with patience
rpr interval --every 60s --resource-id api-health --daemon -- \
    patience exponential --resource-id api-retry -- curl api.example.com/health
```

#### Daemon Coordination Logic
```go
// pkg/daemon/rate_coordinator.go
type RateCoordinator struct {
    globalLimits map[string]*GlobalRateLimit
    mutex        sync.RWMutex
}

type GlobalRateLimit struct {
    Limit           int64
    Window          time.Duration
    CurrentUsage    int64
    WindowStart     time.Time
    ClientAllocations map[string]int64
}

func (rc *RateCoordinator) AllocateTokens(resourceID, clientID string, requested int64) (allocated int64, delay time.Duration) {
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    
    limit := rc.globalLimits[resourceID]
    if limit == nil {
        return requested, 0 // No global limit
    }
    
    // Reset window if expired
    if time.Since(limit.WindowStart) > limit.Window {
        limit.CurrentUsage = 0
        limit.WindowStart = time.Now()
        limit.ClientAllocations = make(map[string]int64)
    }
    
    // Calculate available tokens
    available := limit.Limit - limit.CurrentUsage
    if available <= 0 {
        return 0, limit.Window - time.Since(limit.WindowStart)
    }
    
    // Allocate tokens fairly among clients
    allocated = min(requested, available)
    limit.CurrentUsage += allocated
    limit.ClientAllocations[clientID] += allocated
    
    return allocated, 0
}
```

### 2. Resource-Based Rate Limiting

#### Resource Identification
```go
// pkg/ratelimit/resource.go
type ResourceIdentifier struct {
    Type     ResourceType
    Target   string
    Category string
}

func (r ResourceIdentifier) String() string {
    return fmt.Sprintf("%s:%s:%s", r.Type, r.Category, r.Target)
}

// Examples:
// api:external:api.example.com
// database:postgres:user_db
// filesystem:local:/var/log
// network:tcp:192.168.1.100:8080
```

#### Automatic Resource Detection
```go
// pkg/ratelimit/detector.go
func DetectResourceFromCommand(command []string) *ResourceIdentifier {
    if len(command) == 0 {
        return nil
    }
    
    switch command[0] {
    case "curl", "wget", "http":
        return detectHTTPResource(command)
    case "psql", "mysql", "mongo":
        return detectDatabaseResource(command)
    case "ssh", "scp", "rsync":
        return detectNetworkResource(command)
    default:
        return &ResourceIdentifier{
            Type:     ResourceTypeGeneric,
            Category: "command",
            Target:   command[0],
        }
    }
}
```

## Combined Usage Patterns

### 1. Repeater + Patience Integration

#### Health Check with Retry
```bash
# Repeater runs health checks every 30s, patience retries failures
rpr interval --every 30s --for 8h --daemon --resource-id health-check -- \
    patience exponential --max-attempts 3 --resource-id health-retry -- \
    curl -f --max-time 10 https://api.example.com/health
```

#### Load Testing with Resilience
```bash
# Repeater generates load, patience handles individual request failures
rpr count --times 10000 --every 100ms --daemon --resource-id load-test -- \
    patience linear --max-attempts 2 --resource-id request-retry -- \
    curl -f --max-time 5 https://api.example.com/endpoint
```

### 2. Configuration Coordination

#### Shared Rate Limits
```toml
# ~/.config/patience/config.toml
[rate_limiting.resources]
"api:external:api.example.com" = { limit = 1000, window = "1h" }
"database:postgres:prod_db" = { limit = 100, window = "1m" }

[repeater.defaults]
daemon_enabled = true
resource_detection = true

[patience.defaults]
daemon_enabled = true
resource_detection = true
```

## Daemon Lifecycle Management

### 1. Daemon Discovery and Startup

#### Automatic Daemon Management
```go
// pkg/daemon/manager.go
type DaemonManager struct {
    socketPath   string
    autoStart    bool
    startTimeout time.Duration
}

func (dm *DaemonManager) EnsureDaemon() error {
    // Check if daemon is running
    if dm.isDaemonRunning() {
        return nil
    }
    
    if !dm.autoStart {
        return ErrDaemonNotRunning
    }
    
    // Start daemon if not running
    return dm.startDaemon()
}

func (dm *DaemonManager) startDaemon() error {
    cmd := exec.Command("patience-daemon", "--socket", dm.socketPath)
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start daemon: %w", err)
    }
    
    // Wait for daemon to be ready
    return dm.waitForDaemon(dm.startTimeout)
}
```

### 2. Graceful Degradation

#### Fallback Behavior
```go
// pkg/ratelimit/fallback.go
type FallbackRateLimiter struct {
    localLimiter *LocalRateLimiter
    daemonClient *DaemonClient
    fallbackMode bool
}

func (f *FallbackRateLimiter) Allow() bool {
    if f.fallbackMode {
        return f.localLimiter.Allow()
    }
    
    allowed, err := f.daemonClient.RequestToken()
    if err != nil {
        log.Warn("Daemon unavailable, falling back to local rate limiting")
        f.fallbackMode = true
        return f.localLimiter.Allow()
    }
    
    return allowed
}
```

## Metrics and Observability Integration

### 1. Unified Metrics Collection

#### Shared Metrics Schema
```go
// pkg/metrics/shared.go
type ExecutionMetrics struct {
    Tool            string            `json:"tool"`             // "patience" or "repeater"
    ResourceID      string            `json:"resource_id"`
    Command         []string          `json:"command"`
    StartTime       time.Time         `json:"start_time"`
    Duration        time.Duration     `json:"duration"`
    Success         bool              `json:"success"`
    ExitCode        int               `json:"exit_code"`
    AttemptNumber   int               `json:"attempt_number"`   // For patience
    ExecutionNumber int64             `json:"execution_number"` // For repeater
    Labels          map[string]string `json:"labels"`
}

type AggregatedMetrics struct {
    ResourceID       string        `json:"resource_id"`
    TotalExecutions  int64         `json:"total_executions"`
    SuccessfulRuns   int64         `json:"successful_runs"`
    FailedRuns       int64         `json:"failed_runs"`
    AverageDuration  time.Duration `json:"average_duration"`
    LastExecution    time.Time     `json:"last_execution"`
    ActiveClients    []string      `json:"active_clients"`
}
```

### 2. Cross-Tool Metrics Queries

#### Daemon Metrics API
```go
// pkg/daemon/metrics_api.go
type MetricsAPI struct {
    storage MetricsStorage
}

func (m *MetricsAPI) QueryMetrics(req *MetricsQuery) (*MetricsResponse, error) {
    switch req.Type {
    case "resource_summary":
        return m.getResourceSummary(req.ResourceID)
    case "tool_comparison":
        return m.getToolComparison(req.ResourceID)
    case "rate_limit_usage":
        return m.getRateLimitUsage(req.ResourceID)
    default:
        return nil, ErrInvalidQueryType
    }
}

type MetricsQuery struct {
    Type       string    `json:"type"`
    ResourceID string    `json:"resource_id,omitempty"`
    TimeRange  TimeRange `json:"time_range,omitempty"`
    Tools      []string  `json:"tools,omitempty"`
}
```

## Testing Integration

### 1. Integration Test Suite

#### Cross-Tool Testing
```go
// tests/integration/cross_tool_test.go
func TestRepeaterPatienceIntegration(t *testing.T) {
    // Start shared daemon
    daemon := startTestDaemon(t)
    defer daemon.Stop()
    
    // Test combined usage
    t.Run("RepeaterWithPatienceRetry", func(t *testing.T) {
        cmd := exec.Command("rpr", "interval", "--every", "1s", "--times", "5", 
            "--daemon", "--resource-id", "test-resource", "--",
            "patience", "exponential", "--max-attempts", "3", "--",
            "test-command")
        
        output, err := cmd.CombinedOutput()
        assert.NoError(t, err)
        
        // Verify daemon coordination
        metrics := daemon.GetResourceMetrics("test-resource")
        assert.Equal(t, int64(5), metrics.TotalExecutions)
    })
}
```

### 2. Daemon Coordination Tests

#### Rate Limiting Coordination
```go
func TestMultiInstanceRateCoordination(t *testing.T) {
    daemon := startTestDaemon(t)
    defer daemon.Stop()
    
    // Configure global rate limit
    daemon.SetResourceLimit("test-api", 10, time.Minute)
    
    // Start multiple repeater instances
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(instance int) {
            defer wg.Done()
            cmd := exec.Command("rpr", "count", "--times", "20", "--every", "1s",
                "--daemon", "--resource-id", "test-api", "--",
                "echo", fmt.Sprintf("instance-%d", instance))
            cmd.Run()
        }(i)
    }
    
    wg.Wait()
    
    // Verify rate limit was respected
    metrics := daemon.GetResourceMetrics("test-api")
    assert.LessOrEqual(t, metrics.ExecutionsInLastMinute, int64(10))
}
```

## Migration and Compatibility

### 1. Backward Compatibility

#### Configuration Migration
```go
// pkg/config/migration.go
func MigrateConfig(oldConfig *OldConfig) (*NewConfig, error) {
    newConfig := &NewConfig{
        Daemon: DaemonConfig{
            SocketPath: oldConfig.DaemonSocket,
            LogLevel:   oldConfig.LogLevel,
        },
    }
    
    // Migrate patience-specific settings
    if oldConfig.Patience != nil {
        newConfig.Patience = *oldConfig.Patience
    }
    
    // Add repeater defaults
    newConfig.Repeater = RepeaterConfig{
        DefaultInterval:    "30s",
        DefaultTimeout:     "60s",
        ContinueOnError:    true,
        DaemonEnabled:      true,
        ResourceDetection:  true,
    }
    
    return newConfig, nil
}
```

### 2. Version Compatibility

#### Protocol Versioning
```go
// pkg/daemon/protocol.go
const (
    ProtocolVersionV1 = "1.0"
    ProtocolVersionV2 = "2.0" // Adds repeater support
)

type ProtocolHeader struct {
    Version   string `json:"version"`
    ClientID  string `json:"client_id"`
    Tool      string `json:"tool"`
    Timestamp int64  `json:"timestamp"`
}

func (c *DaemonClient) negotiateProtocol() error {
    // Send version request
    req := &VersionRequest{
        SupportedVersions: []string{ProtocolVersionV2, ProtocolVersionV1},
        Tool:             c.tool,
    }
    
    resp, err := c.sendRequest(req)
    if err != nil {
        return err
    }
    
    c.protocolVersion = resp.SelectedVersion
    return nil
}
```

## Deployment Considerations

### 1. Package Distribution

#### Unified Installation
```bash
# Install both tools together
curl -sSL https://install.patience.sh | sh

# Or install separately
go install github.com/shaneisley/patience@latest
go install github.com/shaneisley/repeater@latest
```

#### System Service Integration
```ini
# /etc/systemd/system/patience-daemon.service
[Unit]
Description=Patience/Repeater Shared Daemon
After=network.target

[Service]
Type=notify
ExecStart=/usr/local/bin/patience-daemon --socket /var/run/patience/daemon.sock
User=patience
Group=patience
RuntimeDirectory=patience
RuntimeDirectoryMode=0755

[Install]
WantedBy=multi-user.target
```

### 2. Configuration Management

#### Centralized Configuration
```yaml
# /etc/patience/config.yaml (system-wide)
daemon:
  socket_path: /var/run/patience/daemon.sock
  log_level: info
  metrics_enabled: true

rate_limiting:
  coordination_enabled: true
  default_algorithm: diophantine
  
  resources:
    "api:external:*":
      limit: 1000
      window: 1h
    "database:postgres:*":
      limit: 100
      window: 1m

tools:
  patience:
    default_strategy: exponential
    max_attempts: 10
  
  repeater:
    default_interval: 30s
    default_timeout: 60s
    continue_on_error: true
```

## Future Integration Opportunities

### 1. Advanced Coordination

#### Adaptive Rate Limiting
- Cross-tool learning from response patterns
- Dynamic rate limit adjustment based on system load
- Predictive scheduling based on historical data

#### Resource Optimization
- Intelligent resource allocation between tools
- Load balancing across multiple instances
- Automatic failover and recovery

### 2. Enhanced Observability

#### Unified Dashboard
- Combined metrics from both tools
- Resource utilization visualization
- Performance trend analysis

#### Alerting Integration
- Cross-tool alert correlation
- Resource exhaustion warnings
- Performance degradation detection

This integration design ensures that repeater and patience work seamlessly together while maintaining their distinct purposes and independent operation capabilities.