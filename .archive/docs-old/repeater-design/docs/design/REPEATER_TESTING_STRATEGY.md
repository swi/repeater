# Repeater Testing Strategy and Quality Assurance

## Overview

This document outlines the comprehensive testing strategy for the repeater project, ensuring high quality, reliability, and performance. The strategy follows Test-Driven Development (TDD) principles and includes multiple testing layers from unit tests to production monitoring.

## Testing Philosophy

### Core Principles

1. **Test-First Development**: Write tests before implementation
2. **Comprehensive Coverage**: Unit, integration, and end-to-end testing
3. **Performance Validation**: Timing accuracy and resource efficiency
4. **Real-World Scenarios**: Test with actual commands and environments
5. **Continuous Quality**: Automated testing in CI/CD pipeline
6. **Production Monitoring**: Ongoing quality validation in live environments

### Quality Goals

- **Reliability**: 99.9% uptime for continuous execution scenarios
- **Accuracy**: ±1% timing precision for scheduled executions
- **Performance**: <10ms overhead per execution
- **Memory Efficiency**: <50MB baseline memory usage
- **Error Handling**: Graceful degradation in all failure scenarios

## Testing Pyramid

```
                    ┌─────────────────────┐
                    │   Manual Testing    │ (5%)
                    │  - Exploratory      │
                    │  - User Acceptance  │
                    └─────────────────────┘
                ┌─────────────────────────────┐
                │    End-to-End Tests         │ (15%)
                │  - Full CLI workflows      │
                │  - Cross-tool integration  │
                │  - Production scenarios    │
                └─────────────────────────────┘
            ┌─────────────────────────────────────┐
            │        Integration Tests            │ (30%)
            │  - Component interactions          │
            │  - Daemon communication           │
            │  - File system operations         │
            │  - Network interactions           │
            └─────────────────────────────────────┘
        ┌─────────────────────────────────────────────┐
        │              Unit Tests                     │ (50%)
        │  - Individual functions and methods        │
        │  - Scheduler algorithms                    │
        │  - Configuration parsing                   │
        │  - Error handling logic                    │
        └─────────────────────────────────────────────┘
```

## Unit Testing Strategy

### 1. Scheduler Testing

#### Interval Scheduler Tests
```go
// pkg/scheduler/interval_test.go
func TestIntervalScheduler(t *testing.T) {
    tests := []struct {
        name     string
        interval time.Duration
        jitter   float64
        count    int
        expected time.Duration
        tolerance time.Duration
    }{
        {
            name:      "Fixed 1 second interval",
            interval:  time.Second,
            jitter:    0,
            count:     10,
            expected:  10 * time.Second,
            tolerance: 100 * time.Millisecond,
        },
        {
            name:      "1 second with 10% jitter",
            interval:  time.Second,
            jitter:    0.1,
            count:     100,
            expected:  100 * time.Second,
            tolerance: 10 * time.Second, // Allow for jitter variance
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scheduler := NewIntervalScheduler(tt.interval, tt.jitter, false)
            
            start := time.Now()
            for i := 0; i < tt.count; i++ {
                <-scheduler.Next()
            }
            elapsed := time.Since(start)
            
            assert.InDelta(t, tt.expected.Seconds(), elapsed.Seconds(), 
                tt.tolerance.Seconds(), "Timing accuracy")
            
            scheduler.Stop()
        })
    }
}

func TestIntervalSchedulerJitter(t *testing.T) {
    scheduler := NewIntervalScheduler(time.Second, 0.2, false)
    defer scheduler.Stop()
    
    intervals := make([]time.Duration, 50)
    lastTime := time.Now()
    
    for i := 0; i < 50; i++ {
        <-scheduler.Next()
        now := time.Now()
        intervals[i] = now.Sub(lastTime)
        lastTime = now
    }
    
    // Verify jitter distribution
    var sum time.Duration
    for _, interval := range intervals {
        sum += interval
    }
    avgInterval := sum / time.Duration(len(intervals))
    
    // Average should be close to 1 second
    assert.InDelta(t, time.Second.Seconds(), avgInterval.Seconds(), 0.1)
    
    // Verify variance (jitter should create spread)
    var variance float64
    for _, interval := range intervals {
        diff := interval.Seconds() - avgInterval.Seconds()
        variance += diff * diff
    }
    variance /= float64(len(intervals))
    
    // With 20% jitter, we expect some variance
    assert.Greater(t, variance, 0.01, "Jitter should create timing variance")
}
```

#### Count Scheduler Tests
```go
func TestCountScheduler(t *testing.T) {
    tests := []struct {
        name     string
        count    int64
        interval time.Duration
        parallel int
    }{
        {"Sequential execution", 10, 100 * time.Millisecond, 1},
        {"Parallel execution", 20, 50 * time.Millisecond, 5},
        {"Single execution", 1, 0, 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scheduler := NewCountScheduler(tt.count, tt.interval, tt.parallel)
            
            executions := 0
            start := time.Now()
            
            for range scheduler.Next() {
                executions++
            }
            
            elapsed := time.Since(start)
            
            assert.Equal(t, int(tt.count), executions)
            
            if tt.interval > 0 && tt.parallel == 1 {
                expectedDuration := time.Duration(tt.count-1) * tt.interval
                assert.InDelta(t, expectedDuration.Seconds(), elapsed.Seconds(), 0.1)
            }
        })
    }
}
```

#### Duration Scheduler Tests
```go
func TestDurationScheduler(t *testing.T) {
    duration := 2 * time.Second
    interval := 200 * time.Millisecond
    
    scheduler := NewDurationScheduler(duration, interval)
    
    executions := 0
    start := time.Now()
    
    for range scheduler.Next() {
        executions++
    }
    
    elapsed := time.Since(start)
    
    // Should run for approximately the specified duration
    assert.InDelta(t, duration.Seconds(), elapsed.Seconds(), 0.1)
    
    // Should execute approximately duration/interval times
    expectedExecutions := int(duration / interval)
    assert.InDelta(t, expectedExecutions, executions, 2) // Allow some variance
}
```

### 2. Command Executor Testing

#### Basic Execution Tests
```go
// pkg/executor/executor_test.go
func TestExecutor_Execute(t *testing.T) {
    executor := NewExecutor(ExecutorConfig{
        Timeout:    30 * time.Second,
        WorkingDir: "/tmp",
    })

    tests := []struct {
        name        string
        command     []string
        expectError bool
        expectCode  int
    }{
        {
            name:        "Successful command",
            command:     []string{"echo", "hello"},
            expectError: false,
            expectCode:  0,
        },
        {
            name:        "Failed command",
            command:     []string{"false"},
            expectError: false, // Command runs, but exits with non-zero
            expectCode:  1,
        },
        {
            name:        "Non-existent command",
            command:     []string{"nonexistent-command"},
            expectError: true,
            expectCode:  -1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result, err := executor.Execute(ctx, tt.command)

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectCode, result.ExitCode)
            }
        })
    }
}

func TestExecutor_Timeout(t *testing.T) {
    executor := NewExecutor(ExecutorConfig{
        Timeout: 100 * time.Millisecond,
    })

    ctx := context.Background()
    start := time.Now()
    
    result, err := executor.Execute(ctx, []string{"sleep", "1"})
    elapsed := time.Since(start)

    assert.NoError(t, err) // Timeout is not an error, just kills process
    assert.NotEqual(t, 0, result.ExitCode) // Should be killed
    assert.Less(t, elapsed, 200*time.Millisecond) // Should timeout quickly
}
```

#### Output Capture Tests
```go
func TestExecutor_OutputCapture(t *testing.T) {
    executor := NewExecutor(ExecutorConfig{})

    tests := []struct {
        name           string
        command        []string
        expectedStdout string
        expectedStderr string
    }{
        {
            name:           "Stdout capture",
            command:        []string{"echo", "hello world"},
            expectedStdout: "hello world\n",
            expectedStderr: "",
        },
        {
            name:           "Stderr capture",
            command:        []string{"sh", "-c", "echo error >&2"},
            expectedStdout: "",
            expectedStderr: "error\n",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result, err := executor.Execute(ctx, tt.command)

            assert.NoError(t, err)
            assert.Equal(t, tt.expectedStdout, result.Stdout)
            assert.Equal(t, tt.expectedStderr, result.Stderr)
        })
    }
}
```

### 3. Configuration Testing

#### Configuration Parsing Tests
```go
// pkg/config/config_test.go
func TestConfig_LoadFromFile(t *testing.T) {
    configContent := `
[defaults]
continue_on_error = true
timeout = "30s"
output_file = "/var/log/repeater.log"

[interval]
jitter = "10%"
immediate = true

[daemon]
socket_path = "/var/run/patience/daemon.sock"
enabled = true
`

    tmpFile, err := os.CreateTemp("", "config-*.toml")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())

    _, err = tmpFile.WriteString(configContent)
    require.NoError(t, err)
    tmpFile.Close()

    config, err := LoadConfig(tmpFile.Name())
    require.NoError(t, err)

    assert.True(t, config.Defaults.ContinueOnError)
    assert.Equal(t, 30*time.Second, config.Defaults.Timeout)
    assert.Equal(t, "/var/log/repeater.log", config.Defaults.OutputFile)
    assert.Equal(t, 0.1, config.Interval.Jitter)
    assert.True(t, config.Interval.Immediate)
    assert.True(t, config.Daemon.Enabled)
}

func TestConfig_EnvironmentOverrides(t *testing.T) {
    os.Setenv("RPR_CONTINUE_ON_ERROR", "false")
    os.Setenv("RPR_TIMEOUT", "60s")
    defer os.Unsetenv("RPR_CONTINUE_ON_ERROR")
    defer os.Unsetenv("RPR_TIMEOUT")

    config := NewDefaultConfig()
    err := config.LoadFromEnvironment()
    require.NoError(t, err)

    assert.False(t, config.Defaults.ContinueOnError)
    assert.Equal(t, 60*time.Second, config.Defaults.Timeout)
}
```

### 4. Stop Condition Testing

#### Stop Condition Manager Tests
```go
// pkg/conditions/conditions_test.go
func TestStopConditionManager(t *testing.T) {
    manager := NewStopConditionManager()
    
    // Add conditions
    manager.AddCondition(NewCountCondition(5))
    manager.AddCondition(NewDurationCondition(2 * time.Second))
    manager.AddCondition(NewFailureCondition(3, 2))

    stats := ExecutionStats{
        TotalExecutions:     3,
        SuccessfulRuns:     2,
        FailedRuns:         1,
        ConsecutiveFailures: 1,
        StartTime:          time.Now().Add(-1 * time.Second),
    }

    // Should not stop yet
    assert.False(t, manager.ShouldStop(stats))

    // Update stats to trigger count condition
    stats.TotalExecutions = 5
    assert.True(t, manager.ShouldStop(stats))
}

func TestFailureCondition(t *testing.T) {
    condition := NewFailureCondition(5, 3) // Max 5 total, 3 consecutive

    tests := []struct {
        name                string
        totalFailures       int
        consecutiveFailures int
        shouldStop          bool
    }{
        {"Under limits", 2, 1, false},
        {"At consecutive limit", 2, 3, true},
        {"At total limit", 5, 2, true},
        {"Over both limits", 6, 4, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            stats := ExecutionStats{
                FailedRuns:          int64(tt.totalFailures),
                ConsecutiveFailures: tt.consecutiveFailures,
            }
            
            assert.Equal(t, tt.shouldStop, condition.ShouldStop(stats))
        })
    }
}
```

## Integration Testing Strategy

### 1. CLI Integration Tests

#### Subcommand Integration Tests
```go
// tests/integration/cli_test.go
func TestCLI_IntervalSubcommand(t *testing.T) {
    tests := []struct {
        name        string
        args        []string
        expectError bool
        expectRuns  int
    }{
        {
            name:        "Basic interval execution",
            args:        []string{"interval", "--every", "100ms", "--times", "3", "--", "echo", "test"},
            expectError: false,
            expectRuns:  3,
        },
        {
            name:        "Missing required flag",
            args:        []string{"interval", "--times", "3", "--", "echo", "test"},
            expectError: true,
            expectRuns:  0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command("rpr", tt.args...)
            output, err := cmd.CombinedOutput()

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                
                // Count occurrences of "test" in output
                runs := strings.Count(string(output), "test")
                assert.Equal(t, tt.expectRuns, runs)
            }
        })
    }
}

func TestCLI_ConfigFile(t *testing.T) {
    configContent := `
[defaults]
continue_on_error = true
timeout = "10s"
`

    configFile, err := os.CreateTemp("", "rpr-config-*.toml")
    require.NoError(t, err)
    defer os.Remove(configFile.Name())

    _, err = configFile.WriteString(configContent)
    require.NoError(t, err)
    configFile.Close()

    cmd := exec.Command("rpr", "--config", configFile.Name(), 
        "count", "--times", "2", "--", "false") // Command that fails
    
    err = cmd.Run()
    // Should not fail because continue_on_error is true in config
    assert.NoError(t, err)
}
```

### 2. Daemon Integration Tests

#### Daemon Communication Tests
```go
// tests/integration/daemon_test.go
func TestDaemonIntegration(t *testing.T) {
    // Start test daemon
    daemon := startTestDaemon(t)
    defer daemon.Stop()

    t.Run("Basic daemon communication", func(t *testing.T) {
        client, err := daemon.NewClient("repeater")
        require.NoError(t, err)
        defer client.Close()

        // Test rate limiting request
        allowed, err := client.RequestRateLimit("test-resource", 1)
        assert.NoError(t, err)
        assert.True(t, allowed)
    })

    t.Run("Rate limit coordination", func(t *testing.T) {
        // Set up rate limit
        daemon.SetResourceLimit("test-api", 5, time.Minute)

        // Start multiple repeater instances
        var wg sync.WaitGroup
        results := make(chan int, 3)

        for i := 0; i < 3; i++ {
            wg.Add(1)
            go func(instance int) {
                defer wg.Done()
                
                cmd := exec.Command("rpr", "count", "--times", "10", "--every", "1s",
                    "--daemon", "--resource-id", "test-api", "--",
                    "echo", fmt.Sprintf("instance-%d", instance))
                
                output, err := cmd.CombinedOutput()
                if err == nil {
                    results <- strings.Count(string(output), fmt.Sprintf("instance-%d", instance))
                } else {
                    results <- 0
                }
            }(i)
        }

        wg.Wait()
        close(results)

        // Verify total executions respect rate limit
        totalExecutions := 0
        for result := range results {
            totalExecutions += result
        }

        // Should be limited by rate limit (5 per minute)
        assert.LessOrEqual(t, totalExecutions, 5)
    })
}
```

### 3. Cross-Tool Integration Tests

#### Repeater + Patience Integration
```go
func TestRepeaterPatienceIntegration(t *testing.T) {
    // Test server that fails first few requests
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Simulate intermittent failures
        if rand.Float32() < 0.3 {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }))
    defer server.Close()

    // Run repeater with patience for resilient requests
    cmd := exec.Command("rpr", "count", "--times", "10", "--every", "500ms", "--continue-on-error", "--",
        "patience", "exponential", "--max-attempts", "3", "--",
        "curl", "-f", server.URL)

    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)

    // Should have some successful requests despite failures
    successCount := strings.Count(string(output), "success")
    assert.Greater(t, successCount, 5) // At least half should succeed with retries
}
```

## End-to-End Testing Strategy

### 1. Real-World Scenario Tests

#### Production-Like Monitoring Test
```go
// tests/e2e/monitoring_test.go
func TestE2E_HealthMonitoring(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Start a test service
    service := startTestService(t, 8080)
    defer service.Stop()

    // Create temporary log file
    logFile, err := os.CreateTemp("", "health-monitor-*.log")
    require.NoError(t, err)
    defer os.Remove(logFile.Name())
    logFile.Close()

    // Run health monitoring for 30 seconds
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "rpr", "interval", 
        "--every", "2s", 
        "--continue-on-error",
        "--output-file", logFile.Name(),
        "--", "curl", "-f", "http://localhost:8080/health")

    err = cmd.Run()
    assert.NoError(t, err)

    // Verify log file contains health check results
    logContent, err := os.ReadFile(logFile.Name())
    require.NoError(t, err)

    // Should have approximately 15 health checks (30s / 2s)
    healthChecks := strings.Count(string(logContent), "HTTP/1.1 200 OK")
    assert.InDelta(t, 15, healthChecks, 3) // Allow some variance
}
```

#### Load Testing Scenario
```go
func TestE2E_LoadTesting(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Start test API server
    server := startTestAPIServer(t)
    defer server.Stop()

    // Run load test
    cmd := exec.Command("rpr", "count", 
        "--times", "100", 
        "--every", "50ms",
        "--parallel", "5",
        "--quiet",
        "--", "curl", "-w", "%{http_code},%{time_total}\\n", 
        "-o", "/dev/null", "-s", server.URL+"/api/test")

    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)

    // Analyze results
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    successCount := 0
    var totalTime float64

    for _, line := range lines {
        if strings.HasPrefix(line, "200,") {
            successCount++
            parts := strings.Split(line, ",")
            if len(parts) == 2 {
                if time, err := strconv.ParseFloat(parts[1], 64); err == nil {
                    totalTime += time
                }
            }
        }
    }

    // Verify load test results
    assert.Greater(t, successCount, 90) // At least 90% success rate
    avgResponseTime := totalTime / float64(successCount)
    assert.Less(t, avgResponseTime, 1.0) // Average response time < 1s
}
```

### 2. Error Scenario Tests

#### Network Failure Handling
```go
func TestE2E_NetworkFailureHandling(t *testing.T) {
    // Start server that will be stopped mid-test
    server := startTestService(t, 8081)

    // Start monitoring
    cmd := exec.Command("rpr", "interval",
        "--every", "1s",
        "--for", "10s",
        "--continue-on-error",
        "--max-failures", "5",
        "--", "curl", "-f", "--max-time", "2", "http://localhost:8081/health")

    // Stop server after 3 seconds to simulate network failure
    go func() {
        time.Sleep(3 * time.Second)
        server.Stop()
    }()

    output, err := cmd.CombinedOutput()
    
    // Should complete without error due to continue-on-error
    assert.NoError(t, err)
    
    // Should have some successful requests before failure
    successCount := strings.Count(string(output), "200 OK")
    assert.GreaterOrEqual(t, successCount, 2)
    
    // Should have some failures after server stops
    failureCount := strings.Count(string(output), "curl: ")
    assert.Greater(t, failureCount, 0)
}
```

## Performance Testing Strategy

### 1. Timing Accuracy Tests

#### Scheduler Precision Benchmarks
```go
// benchmarks/scheduler_bench_test.go
func BenchmarkIntervalScheduler(b *testing.B) {
    scheduler := NewIntervalScheduler(time.Millisecond, 0, false)
    defer scheduler.Stop()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        <-scheduler.Next()
    }
}

func TestSchedulerAccuracy(t *testing.T) {
    intervals := []time.Duration{
        10 * time.Millisecond,
        100 * time.Millisecond,
        time.Second,
    }

    for _, interval := range intervals {
        t.Run(fmt.Sprintf("Interval_%v", interval), func(t *testing.T) {
            scheduler := NewIntervalScheduler(interval, 0, false)
            defer scheduler.Stop()

            measurements := make([]time.Duration, 100)
            lastTime := time.Now()

            for i := 0; i < 100; i++ {
                <-scheduler.Next()
                now := time.Now()
                measurements[i] = now.Sub(lastTime)
                lastTime = now
            }

            // Calculate statistics
            var sum time.Duration
            for _, m := range measurements {
                sum += m
            }
            avg := sum / time.Duration(len(measurements))

            // Verify accuracy (within 5% of target)
            tolerance := time.Duration(float64(interval) * 0.05)
            assert.InDelta(t, interval.Nanoseconds(), avg.Nanoseconds(), 
                float64(tolerance.Nanoseconds()))
        })
    }
}
```

### 2. Resource Usage Tests

#### Memory Usage Benchmarks
```go
func TestMemoryUsage(t *testing.T) {
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)

    // Run repeater for extended period
    scheduler := NewIntervalScheduler(10*time.Millisecond, 0, false)
    executor := NewExecutor(ExecutorConfig{})

    for i := 0; i < 1000; i++ {
        <-scheduler.Next()
        _, _ = executor.Execute(context.Background(), []string{"echo", "test"})
    }

    scheduler.Stop()
    runtime.GC()
    runtime.ReadMemStats(&m2)

    // Memory growth should be minimal
    memoryGrowth := m2.Alloc - m1.Alloc
    assert.Less(t, memoryGrowth, uint64(10*1024*1024)) // Less than 10MB growth
}

func BenchmarkExecutorOverhead(b *testing.B) {
    executor := NewExecutor(ExecutorConfig{})
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = executor.Execute(ctx, []string{"true"})
    }
}
```

### 3. Concurrency Tests

#### Parallel Execution Tests
```go
func TestParallelExecution(t *testing.T) {
    executor := NewExecutor(ExecutorConfig{})
    
    const numGoroutines = 100
    const executionsPerGoroutine = 10
    
    var wg sync.WaitGroup
    results := make(chan *ExecutionResult, numGoroutines*executionsPerGoroutine)
    
    start := time.Now()
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < executionsPerGoroutine; j++ {
                result, err := executor.Execute(context.Background(), 
                    []string{"echo", fmt.Sprintf("goroutine-%d-exec-%d", id, j)})
                if err == nil {
                    results <- result
                }
            }
        }(i)
    }
    
    wg.Wait()
    close(results)
    elapsed := time.Since(start)
    
    // Verify all executions completed
    resultCount := 0
    for range results {
        resultCount++
    }
    
    assert.Equal(t, numGoroutines*executionsPerGoroutine, resultCount)
    
    // Parallel execution should be significantly faster than sequential
    // (This is a rough estimate, actual timing depends on system)
    maxSequentialTime := time.Duration(numGoroutines*executionsPerGoroutine) * 10 * time.Millisecond
    assert.Less(t, elapsed, maxSequentialTime)
}
```

## Continuous Integration Strategy

### 1. CI Pipeline Configuration

#### GitHub Actions Workflow
```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.21, 1.22]
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Run unit tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.22
    
    - name: Build binary
      run: go build -o rpr ./cmd/rpr
    
    - name: Run integration tests
      run: go test -v -tags=integration ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.22
    
    - name: Build binary
      run: go build -o rpr ./cmd/rpr
    
    - name: Run E2E tests
      run: go test -v -tags=e2e -timeout=10m ./tests/e2e/...

  performance-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.22
    
    - name: Run benchmarks
      run: |
        go test -bench=. -benchmem -count=3 ./... > benchmark.txt
        
    - name: Performance regression check
      run: |
        # Compare with baseline benchmarks
        # Fail if performance degrades significantly
        ./scripts/check-performance-regression.sh
```

### 2. Quality Gates

#### Coverage Requirements
```go
// scripts/coverage-check.sh
#!/bin/bash

COVERAGE_THRESHOLD=80
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold $COVERAGE_THRESHOLD%"
    exit 1
fi

echo "Coverage $COVERAGE% meets threshold"
```

#### Performance Regression Detection
```go
// scripts/check-performance-regression.sh
#!/bin/bash

# Compare current benchmarks with baseline
go test -bench=. -count=5 ./... > current_bench.txt

# Check for significant performance regressions
benchcmp baseline_bench.txt current_bench.txt | grep -E "(slower|worse)" | while read line; do
    # Extract performance change percentage
    change=$(echo "$line" | grep -oE '[0-9]+\.[0-9]+x' | sed 's/x//')
    
    if (( $(echo "$change > 1.2" | bc -l) )); then
        echo "Performance regression detected: $line"
        exit 1
    fi
done
```

## Production Monitoring Strategy

### 1. Health Checks

#### Application Health Monitoring
```go
// pkg/health/health.go
type HealthChecker struct {
    checks map[string]HealthCheck
    mutex  sync.RWMutex
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) error
}

type SchedulerHealthCheck struct{}

func (s *SchedulerHealthCheck) Name() string {
    return "scheduler"
}

func (s *SchedulerHealthCheck) Check(ctx context.Context) error {
    // Verify scheduler is responsive
    scheduler := NewIntervalScheduler(time.Millisecond, 0, false)
    defer scheduler.Stop()
    
    select {
    case <-scheduler.Next():
        return nil
    case <-time.After(100 * time.Millisecond):
        return errors.New("scheduler not responsive")
    }
}
```

### 2. Metrics Collection

#### Prometheus Metrics
```go
// pkg/metrics/prometheus.go
var (
    executionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "repeater_executions_total",
            Help: "Total number of command executions",
        },
        []string{"command", "status"},
    )
    
    executionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "repeater_execution_duration_seconds",
            Help: "Duration of command executions",
            Buckets: prometheus.DefBuckets,
        },
        []string{"command"},
    )
    
    schedulerAccuracy = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "repeater_scheduler_accuracy_seconds",
            Help: "Scheduler timing accuracy",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"scheduler_type"},
    )
)

func init() {
    prometheus.MustRegister(executionsTotal)
    prometheus.MustRegister(executionDuration)
    prometheus.MustRegister(schedulerAccuracy)
}
```

### 3. Alerting Rules

#### Prometheus Alerting Rules
```yaml
# alerts/repeater.yml
groups:
- name: repeater
  rules:
  - alert: RepeaterHighFailureRate
    expr: rate(repeater_executions_total{status="failed"}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High failure rate in repeater executions"
      description: "Failure rate is {{ $value }} failures per second"

  - alert: RepeaterSchedulerInaccuracy
    expr: histogram_quantile(0.95, rate(repeater_scheduler_accuracy_seconds_bucket[5m])) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Repeater scheduler timing inaccuracy"
      description: "95th percentile timing accuracy is {{ $value }} seconds"

  - alert: RepeaterMemoryLeak
    expr: process_resident_memory_bytes{job="repeater"} > 100 * 1024 * 1024
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "Potential memory leak in repeater"
      description: "Memory usage is {{ $value | humanizeBytes }}"
```

## Test Data Management

### 1. Test Fixtures

#### Mock Services
```go
// tests/fixtures/mock_service.go
type MockService struct {
    server     *httptest.Server
    responses  []MockResponse
    callCount  int
    mutex      sync.Mutex
}

type MockResponse struct {
    StatusCode int
    Body       string
    Delay      time.Duration
}

func NewMockService(responses []MockResponse) *MockService {
    ms := &MockService{responses: responses}
    
    ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ms.mutex.Lock()
        defer ms.mutex.Unlock()
        
        if ms.callCount < len(ms.responses) {
            response := ms.responses[ms.callCount]
            ms.callCount++
            
            if response.Delay > 0 {
                time.Sleep(response.Delay)
            }
            
            w.WriteHeader(response.StatusCode)
            w.Write([]byte(response.Body))
        } else {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("default response"))
        }
    }))
    
    return ms
}
```

### 2. Test Environment Setup

#### Docker Test Environment
```dockerfile
# tests/docker/Dockerfile.test
FROM golang:1.22-alpine

RUN apk add --no-cache curl bash

WORKDIR /app
COPY . .

RUN go build -o rpr ./cmd/rpr

# Install test dependencies
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest

CMD ["go", "test", "-v", "./..."]
```

```yaml
# tests/docker/docker-compose.test.yml
version: '3.8'

services:
  repeater-test:
    build:
      context: ../..
      dockerfile: tests/docker/Dockerfile.test
    volumes:
      - ../..:/app
    environment:
      - GO_ENV=test
    depends_on:
      - test-api
      - test-db

  test-api:
    image: nginx:alpine
    ports:
      - "8080:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf

  test-db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5432:5432"
```

## Quality Assurance Checklist

### Pre-Release Checklist

#### Functionality
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] Performance benchmarks meet requirements
- [ ] Memory usage within acceptable limits
- [ ] No race conditions detected
- [ ] Error handling covers all scenarios
- [ ] Configuration validation works correctly

#### Compatibility
- [ ] Works with supported Go versions
- [ ] Cross-platform compatibility (Linux, macOS, Windows)
- [ ] Backward compatibility with existing configurations
- [ ] Integration with patience daemon works correctly
- [ ] CLI interface matches specification

#### Documentation
- [ ] All public APIs documented
- [ ] Usage examples are accurate
- [ ] Configuration options documented
- [ ] Error messages are helpful
- [ ] Performance characteristics documented

#### Security
- [ ] No secrets in logs or output
- [ ] Input validation prevents injection attacks
- [ ] File permissions are appropriate
- [ ] Network communications are secure

This comprehensive testing strategy ensures that repeater meets high quality standards and performs reliably in production environments.