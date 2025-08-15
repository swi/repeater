# Repeater (rpr) Usage Guide

A comprehensive guide to using the `rpr` command-line tool for continuous command execution with intelligent scheduling and Unix pipeline integration.

> **✅ All examples in this guide are tested and working with the current implementation (v0.3.0)**

## Table of Contents

- [Installation](#installation)
- [Basic Syntax](#basic-syntax)
- [Unix Pipeline Integration](#unix-pipeline-integration)
- [Output Modes](#output-modes)
- [Command Abbreviations](#command-abbreviations)
- [Core Commands](#core-commands)
- [Advanced Scheduling](#advanced-scheduling)
- [Pattern Matching](#pattern-matching)
- [HTTP-Aware Intelligence](#http-aware-intelligence)
- [Configuration](#configuration)
- [Real-World Use Cases](#real-world-use-cases)
- [Integration Patterns](#integration-patterns)
- [Exit Codes for Scripting](#exit-codes-for-scripting)
- [Performance Considerations](#performance-considerations)
- [Troubleshooting](#troubleshooting)
- [Tips & Best Practices](#tips--best-practices)

## Installation

### Build from Source
```bash
git clone https://github.com/swi/repeater
cd repeater
go build -o rpr ./cmd/rpr
```

### Install Directly
```bash
go install github.com/swi/repeater/cmd/rpr@latest
```

### Verify Installation
```bash
rpr --version
rpr --help
```

## Basic Syntax

The basic syntax for `rpr` is:
```bash
rpr [GLOBAL OPTIONS] <SUBCOMMAND> [OPTIONS] -- <COMMAND>
```

**Key Points:**
- Use `--` to separate repeater options from the command you want to run
- Commands are executed exactly as you would run them manually
- **Unix-friendly by default**: Clean output perfect for pipes and scripts
- **Streaming output**: Real-time command output for immediate processing
- **Standard exit codes**: 0 (success), 1 (failures), 2 (usage error), 130 (interrupted)

## Unix Pipeline Integration

Repeater is designed to work seamlessly with Unix pipelines and standard tools:

```bash
# Count successful API responses
rpr interval --every 5s --times 10 -- curl -s https://api.example.com | grep -c "success"

# Monitor disk usage and log changes
rpr duration --for 1h --every 5m -- df -h / | awk '{print $5}' | tee disk-usage.log

# Extract data from repeated API calls
rpr count --times 5 -- curl -s https://api.github.com/user | jq -r '.login'

# Test response times and analyze
rpr i -e 1s -t 20 -- curl -w "%{time_total}\n" -o /dev/null -s https://api.com | sort -n
```

## Output Modes

Repeater provides different output modes for various use cases:

### Default Mode (Pipeline-Friendly)
Clean command output with no decorative elements, perfect for Unix pipelines:
```bash
rpr interval --every 2s --times 3 -- echo "test"
# Output:
# test
# test  
# test
```

### Quiet Mode (`--quiet`, `-q`)
Suppresses all command output, shows only tool errors:
```bash
rpr interval --every 2s --times 3 --quiet -- echo "test"
# No output unless there's a tool error
```

### Verbose Mode (`--verbose`, `-v`)
Shows full execution information plus command output:
```bash
rpr interval --every 2s --times 3 --verbose -- echo "test"
# Shows detailed execution info with command output
```

### Stats-Only Mode (`--stats-only`)
Shows only execution statistics, suppresses command output:
```bash
rpr interval --every 2s --times 3 --stats-only -- echo "test"
# Shows final execution statistics only
```

## Command Abbreviations

Repeater supports multiple levels of abbreviations for faster typing:

### Subcommand Abbreviations

| Full Command | Primary | Minimal | Example |
|--------------|---------|---------|---------|
| `interval` | `int` | `i` | `rpr i -e 30s -- curl api.com` |
| `count` | `cnt` | `c` | `rpr c -t 5 -- echo hello` |
| `duration` | `dur` | `d` | `rpr d -f 1m -- date` |
| `cron` | `cr` | `cr` | `rpr cr --cron "0 9 * * *" -- ./backup.sh` |
| `rate-limit` | `rate` | `rl` | `rpr rl -r 10/1h -- curl api.com` |
| `adaptive` | `adapt` | `a` | `rpr a -b 1s -- curl api.com` |
| `backoff` | `back` | `b` | `rpr b -i 100ms -- curl api.com` |
| `load-adaptive` | `load` | `la` | `rpr la -b 1s -- ./task.sh` |

### Flag Abbreviations

| Full Flag | Short | Example |
|-----------|-------|---------|
| `--every DURATION` | `-e DURATION` | `-e 30s` |
| `--times COUNT` | `-t COUNT` | `-t 10` |
| `--for DURATION` | `-f DURATION` | `-f 2m` |
| `--cron EXPRESSION` | `--cron` | `--cron "0 9 * * *"` |
| `--timezone TZ` | `--tz TZ` | `--tz "America/New_York"` |

## Core Commands

### Interval Execution

Execute a command at regular time intervals.

```bash
# Basic interval execution
rpr interval --every 30s -- curl -I https://example.com

# With stop conditions
rpr interval --every 10s --times 6 -- echo "Status check"
rpr interval --every 2s --for 1m -- date

# Abbreviated forms
rpr i -e 30s -t 10 -- curl -f https://api.com
rpr i -e 5s -f 5m -- ./health-check.sh
```

### Count-Based Execution

Execute a command a specific number of times.

```bash
# Basic count execution
rpr count --times 10 -- curl -f https://api.example.com/health

# With intervals
rpr count --times 5 --every 30s -- systemctl status nginx

# Abbreviated
rpr c -t 100 -e 5s -- npm test
```

### Duration-Based Execution

Execute a command continuously for a specific time period.

```bash
# Basic duration execution
rpr duration --for 10m -- top -n 1 | head -20

# With intervals
rpr duration --for 5m --every 30s -- free -h

# Abbreviated
rpr d -f 1h -e 5m -- ./system-monitor.sh
```

## Advanced Scheduling

### Cron-like Scheduling

Execute commands based on cron expressions with timezone support.

```bash
# Daily backup at 9 AM
rpr cron --cron "0 9 * * *" -- ./daily-backup.sh

# Weekday reports at 9 AM EST
rpr cron --cron "0 9 * * 1-5" --timezone "America/New_York" -- ./generate-report.sh

# Every 15 minutes
rpr cron --cron "*/15 * * * *" -- curl -f https://api.example.com/health

# Using shortcuts
rpr cron --cron "@daily" -- ./cleanup.sh
rpr cron --cron "@hourly" -- ./log-rotation.sh
```

#### Cron Expression Format
```
┌───────────── minute (0 - 59)
│ ┌─────────── hour (0 - 23)
│ │ ┌───────── day of month (1 - 31)
│ │ │ ┌─────── month (1 - 12)
│ │ │ │ ┌───── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Adaptive Scheduling

Automatically adjust execution intervals based on command response times and success rates.

```bash
# API monitoring with adaptive intervals
rpr adaptive --base-interval 1s --show-metrics -- curl https://api.example.com/health

# Database health check with bounds
rpr adaptive --base-interval 30s --min-interval 10s --max-interval 5m -- mysql -e "SELECT 1"

# Abbreviated
rpr a -b 1s --show-metrics -- curl https://api.com
```

### Exponential Backoff

Implement exponential backoff for resilient execution against unreliable services.

```bash
# Retry unreliable API with backoff
rpr backoff --initial 100ms --max 30s --multiplier 2.0 -- curl https://flaky-api.com

# Database connection with jitter
rpr backoff --initial 1s --max 60s --jitter 0.1 --times 10 -- mysql -e "SELECT 1"

# Abbreviated
rpr b -i 100ms --max 30s -- curl https://unreliable-service.com
```

### Load-Aware Scheduling

Automatically adjust execution frequency based on system resource usage.

```bash
# CPU-intensive task with load awareness
rpr load-adaptive --base-interval 1s --target-cpu 70 -- ./cpu-intensive-task.sh

# Memory-sensitive processing
rpr load-adaptive --base-interval 30s --target-memory 80 --target-load 1.5 -- ./process-data.sh

# Abbreviated
rpr la -b 1s --target-cpu 60 -- npm test
```

### Rate-Limited Execution

Execute commands with server-friendly rate limiting.

```bash
# API calls with hourly rate limit
rpr rate-limit --rate 100/1h -- curl https://api.github.com/user

# Database queries with per-minute limit
rpr rate-limit --rate 10/1m -- mysql -e "SELECT COUNT(*) FROM users"

# Abbreviated
rpr rl -r 50/1h -- curl https://api.example.com
```

## Pattern Matching

Pattern matching allows you to define success and failure conditions based on command output rather than just exit codes.

### Success Patterns

Define regex patterns that indicate success regardless of the command's exit code.

```bash
# Monitor deployment success
rpr interval --every 30s --times 10 --success-pattern "deployment successful" -- ./deploy.sh

# API health check with JSON response
rpr count --times 5 --success-pattern '"status":\s*"ok"' -- curl -s https://api.example.com/health

# Database connection verification
rpr interval --every 1m --success-pattern "1" -- mysql -e "SELECT 1"
```

### Failure Patterns

Define regex patterns that indicate failure regardless of the command's exit code.

```bash
# Detect errors in application logs
rpr duration --for 1h --every 5m --failure-pattern "(?i)error|exception|failed" -- tail -n 10 /var/log/app.log

# API monitoring with error detection
rpr interval --every 30s --failure-pattern "(?i)timeout|unavailable|error" -- curl -s https://api.example.com/status

# Service health monitoring
rpr count --times 10 --failure-pattern "down|inactive|failed" -- systemctl status nginx
```

### Case-Insensitive Matching

```bash
# Case-insensitive success detection
rpr interval --every 1m --case-insensitive --success-pattern "healthy|ok|running" -- ./health-check.sh

# Case-insensitive error detection
rpr duration --for 30m --every 2m --case-insensitive --failure-pattern "error|warning|critical" -- ./system-check.sh
```

### Pattern Precedence

When both success and failure patterns are specified, failure patterns take precedence.

```bash
# Comprehensive monitoring with both patterns
rpr interval --every 30s --success-pattern "status.*ok" --failure-pattern "(?i)error|timeout|failed" -- curl -s https://api.example.com/health

# Log monitoring with precedence
rpr duration --for 1h --every 5m --success-pattern "completed" --failure-pattern "(?i)error|exception" -- tail -n 5 /var/log/process.log
```

## HTTP-Aware Intelligence

HTTP-aware intelligence automatically parses HTTP responses to extract timing information, making API monitoring significantly more efficient.

### Basic HTTP-Aware Scheduling

```bash
# Basic HTTP-aware API monitoring (respects Retry-After headers)
rpr interval --every 30s --http-aware -- curl -s https://api.github.com/user

# Combine with adaptive scheduling for maximum intelligence
rpr adaptive --base-interval 1s --http-aware --verbose -- curl -s https://api.example.com

# Use with any scheduler type
rpr count --times 10 --http-aware -- curl -s https://httpbin.org/status/503
```

### Configuration Options

```bash
# Set delay constraints to prevent excessive waits
rpr i -e 30s --http-aware --http-max-delay 10m --http-min-delay 2s -- curl -s https://api.com

# Parse custom JSON timing fields
rpr i -e 1m --http-aware --http-custom-fields "custom_retry,backoff_seconds" -- curl -s https://api.com

# Disable JSON parsing, only use HTTP headers
rpr i -e 30s --http-aware --http-no-parse-json -- curl -s https://api.com

# Trust client error timing (4xx responses with Retry-After)
rpr i -e 30s --http-aware --http-trust-client -- curl -s https://api.com
```

### Real-World API Examples

```bash
# GitHub API with automatic rate limit handling
rpr i -e 1m --http-aware --verbose -- curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user

# AWS API with throttling support
rpr adaptive --base-interval 2s --http-aware --http-max-delay 5m -- aws s3 ls s3://my-bucket/

# Stripe API monitoring with constraints
rpr i -e 30s --http-aware --http-max-delay 2m -- curl -H "Authorization: Bearer $STRIPE_KEY" https://api.stripe.com/v1/charges

# Discord API with fractional retry timing
rpr i -e 1s --http-aware -- curl -H "Authorization: Bot $DISCORD_TOKEN" https://discord.com/api/v10/gateway
```

## Configuration

### TOML Configuration Files

Create configuration files to set default options:

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

[http_aware]
max_delay = "10m"
min_delay = "1s"
parse_json = true
parse_headers = true
```

### Environment Variables

Override configuration with environment variables:

```bash
RPR_TIMEOUT=60s rpr i -e 10s -t 5 -- long-running-task.sh
RPR_ENABLE_METRICS=true rpr i -e 30s -- monitoring-script.sh
```

### Using Configuration Files

```bash
# Use default config file
rpr i -e 5s -t 10 -- monitoring-script.sh

# Use custom config file
rpr --config /path/to/custom.toml i -e 5s -t 10 -- script.sh
```

## Real-World Use Cases

### Monitoring & Health Checks

```bash
# Website uptime monitoring with logging
rpr duration --for 1h --every 30s -- curl -w "%{http_code}\n" -s -o /dev/null https://mysite.com | tee uptime.log

# Database connection testing with pattern matching
rpr count --times 10 --every 5s --success-pattern "1" -- mysql -h db.example.com -u user -p -e "SELECT 1"

# SSL certificate expiration check
rpr interval --every 24h --times 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
```

### Development & Testing

```bash
# API load testing with success detection
rpr count --times 100 --success-pattern '"status":\s*"success"' -- curl -s https://api.example.com/endpoint

# Build system monitoring with pattern matching
rpr duration --for 8h --every 2m --success-pattern "passed|success" --failure-pattern "(?i)failed|error" -- curl -s https://ci.example.com/api/build/status

# Database migration progress monitoring
rpr duration --for 30m --every 10s --success-pattern "migration completed" --failure-pattern "(?i)error|failed" -- ./check-migration-status.sh
```

### System Administration

```bash
# Log rotation monitoring
rpr interval --every 1h -- du -sh /var/log/*.log

# Service restart verification
rpr count --times 5 --every 12s -- systemctl is-active my-service

# Disk cleanup verification
rpr duration --for 10m --every 30s -- df -h /tmp
```

### Data Processing

```bash
# Batch job monitoring
rpr duration --for 2h --every 5m -- qstat -u $USER

# File processing pipeline
rpr duration --for 1h --every 10s -- find /incoming -name "*.csv" -exec ./process.sh {} \;

# Data synchronization
rpr count --times 8 --every 15m -- rsync -av /local/data/ remote:/backup/data/
```

## Integration Patterns

### With Monitoring Systems

```bash
# Prometheus metrics collection
rpr interval --every 30s -- curl -s http://localhost:8080/metrics | grep -E '^(cpu_usage|memory_usage)' | awk '{print $1 " " $2 " " systime()}' >> /var/lib/prometheus/node_exporter/repeater.prom

# Grafana dashboard data
rpr i -e 1m -- sh -c 'echo "$(date +%s),$(df / | tail -1 | awk "{print \$5}" | sed "s/%//")"; sleep 1' | tee -a /var/log/disk-usage.csv
```

### With CI/CD Pipelines

```bash
# GitHub Actions health check
rpr count --times 5 --every 10s --quiet -- curl -f https://staging.example.com/health

# Jenkins pipeline health verification
rpr interval --every 30s --times 10 --quiet -- curl -f ${DEPLOYMENT_URL}/health
```

### With Container Orchestration

```bash
# Kubernetes liveness probe
rpr count --times 1 --quiet -- curl -f http://localhost:8080/health

# Docker container monitoring
rpr i -e 30s --stream -- docker exec mycontainer health-check.sh
```

## Exit Codes for Scripting

Repeater follows Unix conventions for exit codes:

- **0**: All commands executed successfully
- **1**: Some commands failed during execution  
- **2**: Usage error (invalid arguments, configuration issues)
- **130**: Interrupted by user (Ctrl+C, SIGINT, SIGTERM)

### Scripting Examples

```bash
# Basic success/failure handling
if rpr interval --every 5s --times 3 --quiet -- curl -f https://api.example.com; then
    echo "API is healthy"
else
    echo "API check failed with exit code $?"
fi

# Chain with other Unix tools
rpr count --times 5 -- curl -s https://api.example.com | jq .status && echo "Success" || echo "Failed"

# Capture and handle different exit codes
rpr interval --every 10s --times 5 --quiet -- curl -f https://api.example.com
exit_code=$?

case $exit_code in
    0)   echo "All health checks passed" ;;
    1)   echo "Some health checks failed" ;;
    2)   echo "Configuration error" ;;
    130) echo "Interrupted by user" ;;
    *)   echo "Unexpected exit code: $exit_code" ;;
esac
```

## Performance Considerations

### Choosing Scheduling Modes

- **Interval Mode**: Best for regular monitoring, predictable resource usage
- **Adaptive Mode**: Best for variable workloads, automatically adjusts to conditions
- **Load-Adaptive Mode**: Best for resource-aware execution, scales with system resources
- **Backoff Mode**: Best for unreliable services, reduces load on failing services

### Resource Usage Guidelines

```bash
# Light resource usage
rpr interval --every 60s --quiet -- curl -I https://api.com

# Moderate resource usage
rpr adaptive --base-interval 1s -- curl https://api.com | jq . | awk '{print $1}'

# CPU-intensive (use load-adaptive)
rpr load-adaptive --base-interval 5s --target-cpu 60 -- ./cpu-intensive-task.sh
```

### Performance Tuning Tips

```bash
# Optimize command execution
rpr interval --every 30s -- sh -c 'curl api1.com & curl api2.com & curl api3.com & wait'

# Reduce output processing overhead
rpr interval --every 1s -- curl -s https://api.com/large-response | jq -r '.status'

# Use appropriate output modes
rpr interval --every 100ms --quiet -- ./fast-check.sh
rpr interval --every 1m --stats-only -- ./performance-test.sh
```

## Troubleshooting

### Common Issues

```bash
# Problem: Pipeline not working as expected
rpr count --times 3 -- echo "test" | wc -l  # Returns 1 instead of 3

# Solution: Use shell wrapper
rpr count --times 3 -- sh -c 'echo "test"' | wc -l  # Returns 3

# Problem: High CPU usage
rpr interval --every 100ms -- curl https://api.com
# Solution: Increase interval or use load-adaptive
rpr load-adaptive --base-interval 100ms --target-cpu 70 -- curl https://api.com

# Problem: Commands not executing
rpr interval --every 30s -- ./slow-script.sh
# Solution: Add timeout
rpr interval --every 30s -- timeout 25s ./slow-script.sh
```

### Debugging Tips

```bash
# Test commands manually first
./your-command.sh
echo "Exit code: $?"

# Use verbose mode for debugging
rpr count --times 3 --verbose -- curl https://api.example.com

# Check exit codes
rpr count --times 5 --quiet -- ./your-command.sh
echo "Repeater exit code: $?"
```

## Tips & Best Practices

### Duration Formats
```bash
30s     # 30 seconds
5m      # 5 minutes  
2h      # 2 hours
1d      # 1 day
1h30m   # 1 hour 30 minutes
```

### Command Best Practices

```bash
# Use absolute paths for scripts
rpr interval --every 1m -- /home/user/scripts/backup.sh

# Quote complex commands
rpr count --times 5 -- sh -c "echo 'Current time:' && date"

# Handle command failures gracefully
rpr interval --every 30s -- sh -c "curl -f https://api.com || echo 'API down at $(date)'"
```

### Performance Considerations

```bash
# For high-frequency execution
rpr interval --every 100ms -- echo "tick"

# For long-running commands
rpr interval --every 5m -- ./long-running-backup.sh

# Capture output to file
rpr interval --every 1m -- date >> /tmp/timestamps.log

# Add timestamps to output
rpr count --times 10 -- sh -c "echo '[$(date)]' && curl -s https://api.com"
```

### Signal Handling

- **Ctrl+C** (SIGINT): Gracefully stops execution after current command completes
- **SIGTERM**: Gracefully stops execution (useful in scripts and containers)

## Getting Help

```bash
# Show general help
rpr --help

# Show version
rpr --version

# Show subcommand help
rpr interval --help
rpr i -h
rpr count --help
rpr c -h

# Show examples
rpr examples
```

For more information and advanced usage patterns, see the project documentation at [github.com/swi/repeater](https://github.com/swi/repeater).