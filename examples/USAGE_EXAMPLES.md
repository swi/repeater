# Repeater (rpr) - Usage Examples

## 🚀 **Quick Start Examples**

### **Basic Interval Execution**
```bash
# Execute every 30 seconds, 10 times
rpr interval --every 30s --times 10 -- curl https://api.example.com/health

# Abbreviated form (same as above)
rpr i -e 30s -t 10 -- curl https://api.example.com/health

# Ultra-compact form
rpr i -e30s -t10 -- curl https://api.example.com/health
```

### **Count-Based Execution**
```bash
# Execute exactly 5 times with 2-second intervals
rpr count --times 5 --every 2s -- echo "Execution #$(date)"

# Abbreviated
rpr c -t 5 -e 2s -- echo "Execution #$(date)"
```

### **Duration-Based Execution**
```bash
# Execute for 5 minutes with 10-second intervals
rpr duration --for 5m --every 10s -- ./monitor-script.sh

# Abbreviated
rpr d -f 5m -e 10s -- ./monitor-script.sh
```

## 📅 **Cron-Based Scheduling**

### **Standard Cron Expressions**
```bash
# Every 5 minutes
rpr cron --expression "*/5 * * * *" -- backup-logs.sh

# Every weekday at 9 AM
rpr cron --expression "0 9 * * 1-5" -- daily-report.sh

# Every hour on the hour
rpr cron --expression "0 * * * *" -- hourly-cleanup.sh
```

### **Cron Shortcuts**
```bash
# Daily at midnight
rpr cron --expression "@daily" -- daily-backup.sh

# Every hour
rpr cron --expression "@hourly" -- log-rotation.sh

# Weekly on Sunday
rpr cron --expression "@weekly" -- weekly-maintenance.sh
```

### **Timezone Support**
```bash
# Execute in New York timezone
rpr cron --expression "0 9 * * *" --timezone "America/New_York" -- morning-report.sh

# Execute in Tokyo timezone
rpr cron --expression "0 18 * * *" --timezone "Asia/Tokyo" -- evening-summary.sh
```

## 🧠 **Adaptive Scheduling**

### **Basic Adaptive Execution**
```bash
# Start with 1-second intervals, adapt based on performance
rpr adaptive --base-interval 1s -- curl https://api.example.com

# With custom adaptation range
rpr adaptive --base-interval 2s --min-interval 500ms --max-interval 30s -- health-check.sh
```

### **Advanced Adaptive Configuration**
```bash
# Fine-tuned adaptive parameters
rpr adaptive \
  --base-interval 1s \
  --min-interval 100ms \
  --max-interval 60s \
  --success-threshold 0.8 \
  --response-threshold 2s \
  -- api-performance-test.sh
```

## 📈 **Backoff Scheduling**

### **Exponential Backoff**
```bash
# Start with 1s, double on each failure
rpr backoff --initial-interval 1s --multiplier 2.0 -- flaky-service-call.sh

# With maximum cap and jitter
rpr backoff \
  --initial-interval 500ms \
  --multiplier 1.5 \
  --max-interval 30s \
  --jitter 0.1 \
  -- unreliable-api.sh
```

### **Linear Backoff**
```bash
# Increase by 2 seconds each failure
rpr backoff --initial-interval 2s --multiplier 1.0 --increment 2s -- retry-operation.sh
```

## 🖥️ **Load-Aware Scheduling**

### **System Resource Monitoring**
```bash
# Adjust based on CPU usage (target 70%)
rpr load-adaptive --base-interval 1s --target-cpu 0.7 -- cpu-intensive-task.sh

# Adjust based on memory usage (target 80%)
rpr load-adaptive --base-interval 2s --target-memory 0.8 -- memory-intensive-task.sh

# Monitor both CPU and memory
rpr load-adaptive \
  --base-interval 1s \
  --target-cpu 0.6 \
  --target-memory 0.7 \
  --min-interval 500ms \
  --max-interval 10s \
  -- resource-heavy-operation.sh
```

## 🚦 **Rate Limiting**

### **Basic Rate Limiting**
```bash
# 10 requests per minute
rpr rate-limit --rate "10/1m" -- api-call.sh

# 100 requests per hour with burst
rpr rate-limit --rate "100/1h" --burst 10 -- batch-operation.sh

# 5 requests per 30 seconds
rpr rate-limit --rate "5/30s" -- frequent-check.sh
```

### **Rate Limiting with Retry Patterns**
```bash
# Exponential backoff on rate limit
rpr rate-limit \
  --rate "20/1m" \
  --retry-pattern exponential \
  --max-retries 5 \
  -- rate-limited-api.sh

# Linear backoff with custom delays
rpr rate-limit \
  --rate "50/1h" \
  --retry-pattern linear \
  --retry-delay 30s \
  -- slow-api-endpoint.sh
```

## 📊 **Output Control & Monitoring**

### **Streaming Output**
```bash
# Stream command output in real-time
rpr i -e 5s -t 10 --stream -- tail -f /var/log/app.log

# Stream with custom prefix
rpr i -e 2s --stream --output-prefix "[MONITOR]" -- system-check.sh
```

### **Quiet and Verbose Modes**
```bash
# Suppress all output except errors
rpr i -e 30s -t 5 --quiet -- background-task.sh

# Verbose output with execution details
rpr i -e 10s -t 3 --verbose -- debug-script.sh

# Statistics only (no command output)
rpr i -e 5s -t 10 --stats-only -- performance-test.sh
```

### **Metrics and Health Monitoring**
```bash
# Enable Prometheus metrics on port 8080
rpr i -e 10s --enable-metrics --metrics-port 8080 -- api-health-check.sh

# Enable health endpoint on port 8081
rpr i -e 5s --enable-health --health-port 8081 -- service-monitor.sh

# Both metrics and health
rpr adaptive \
  --base-interval 2s \
  --enable-metrics \
  --enable-health \
  -- comprehensive-monitoring.sh
```

## 🔧 **Configuration Files**

### **TOML Configuration Example**
```toml
# ~/.config/rpr/config.toml
[default]
timeout = "30s"
log_level = "info"
enable_metrics = true
metrics_port = 8080
enable_health = true
health_port = 8081

[scheduler]
default_interval = "10s"
max_jitter = 0.1

[adaptive]
success_threshold = 0.85
response_threshold = "2s"
ewma_alpha = 0.3
```

### **Using Configuration Files**
```bash
# Use default config file
rpr i -e 5s -t 10 -- monitoring-script.sh

# Use custom config file
rpr --config /path/to/custom.toml i -e 5s -t 10 -- script.sh

# Override config with environment variables
RPR_TIMEOUT=60s rpr i -e 10s -t 5 -- long-running-task.sh
```

## 🔄 **Complex Workflows**

### **Multi-Stage Monitoring**
```bash
# Stage 1: Quick health checks
rpr i -e 5s -t 12 --stats-only -- quick-health-check.sh

# Stage 2: Detailed monitoring if issues found
rpr adaptive --base-interval 30s --enable-metrics -- detailed-diagnostics.sh

# Stage 3: Recovery operations
rpr backoff --initial-interval 10s --max-interval 300s -- recovery-script.sh
```

### **API Testing Pipeline**
```bash
# Load testing with rate limiting
rpr rate-limit --rate "100/1m" --burst 20 --stats-only -- load-test-api.sh

# Performance monitoring with adaptive intervals
rpr adaptive \
  --base-interval 1s \
  --min-interval 100ms \
  --max-interval 10s \
  --enable-metrics \
  --verbose \
  -- performance-monitor.sh

# Stress testing with backoff on failures
rpr backoff \
  --initial-interval 500ms \
  --multiplier 2.0 \
  --max-interval 60s \
  --stream \
  -- stress-test-endpoint.sh
```

### **System Maintenance Automation**
```bash
# Daily maintenance at 2 AM
rpr cron --expression "0 2 * * *" --quiet -- daily-maintenance.sh

# Hourly log rotation
rpr cron --expression "@hourly" --output-prefix "[LOG-ROTATE]" -- rotate-logs.sh

# Load-aware cleanup (only when system is idle)
rpr load-adaptive \
  --base-interval 300s \
  --target-cpu 0.3 \
  --target-memory 0.5 \
  --quiet \
  -- system-cleanup.sh
```

## 🐛 **Debugging & Troubleshooting**

### **Verbose Debugging**
```bash
# Maximum verbosity for troubleshooting
rpr i -e 5s -t 3 --verbose --stream -- problematic-script.sh

# Debug with metrics collection
rpr adaptive \
  --base-interval 2s \
  --verbose \
  --enable-metrics \
  --enable-health \
  -- debug-target.sh
```

### **Error Analysis**
```bash
# Capture all output for analysis
rpr i -e 10s -t 5 --stream --output-prefix "[DEBUG]" -- error-prone-command.sh 2>&1 | tee debug.log

# Test failure recovery
rpr backoff \
  --initial-interval 1s \
  --multiplier 2.0 \
  --max-interval 30s \
  --verbose \
  -- failing-service.sh
```

## 📱 **Integration Examples**

### **Docker Container Monitoring**
```bash
# Monitor container health
rpr i -e 30s --stream -- docker exec mycontainer health-check.sh

# Adaptive container scaling trigger
rpr adaptive --base-interval 60s --enable-metrics -- check-container-load.sh
```

### **Kubernetes Health Checks**
```bash
# Pod readiness monitoring
rpr i -e 10s --quiet -- kubectl get pods -l app=myapp --no-headers | grep -v Running

# Service endpoint testing
rpr rate-limit --rate "10/1m" -- kubectl exec -it pod -- curl localhost:8080/health
```

### **CI/CD Pipeline Integration**
```bash
# Build status monitoring
rpr i -e 60s -t 30 --stats-only -- check-build-status.sh

# Deployment health verification
rpr adaptive \
  --base-interval 30s \
  --min-interval 10s \
  --max-interval 300s \
  --enable-health \
  -- verify-deployment.sh
```

## 🎯 **Best Practices**

### **Performance Optimization**
- Use `--stats-only` for monitoring without output overhead
- Enable metrics only when needed for monitoring
- Use adaptive scheduling for variable-load scenarios
- Implement proper timeout values for commands

### **Reliability**
- Always use `--quiet` in production scripts
- Implement proper error handling in monitored commands
- Use backoff scheduling for unreliable services
- Monitor system resources with load-aware scheduling

### **Monitoring & Observability**
- Enable metrics for production monitoring
- Use health endpoints for service discovery
- Implement structured logging in monitored commands
- Use verbose mode only for debugging

### **Security**
- Validate all command inputs
- Use configuration files for sensitive parameters
- Implement proper timeout values
- Monitor resource usage to prevent abuse