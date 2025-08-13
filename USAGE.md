# Repeater (rpr) Usage Guide

A comprehensive guide to using the `rpr` command-line tool for continuous command execution with intelligent scheduling and Unix pipeline integration.

> **✅ All examples in this guide are tested and working with the current implementation (v0.2.0 Unix Pipeline Ready)**

## Table of Contents

- [Quick Start](#quick-start)
- [Unix Pipeline Integration](#unix-pipeline-integration)
- [Output Modes](#output-modes)
- [Command Abbreviations](#command-abbreviations)
- [Basic Commands](#basic-commands)
  - [Interval Execution](#interval-execution)
  - [Count-Based Execution](#count-based-execution)
  - [Duration-Based Execution](#duration-based-execution)
- [Advanced Scheduling](#advanced-scheduling)
  - [Cron-like Scheduling](#cron-like-scheduling)
  - [Rate-Limited Execution](#rate-limited-execution)
  - [Adaptive Scheduling](#adaptive-scheduling)
  - [Exponential Backoff](#exponential-backoff)
  - [Load-Aware Scheduling](#load-aware-scheduling)
  - [Plugin-Based Scheduling](#plugin-based-scheduling)
- [Pattern Matching](#pattern-matching)
  - [Success Patterns](#success-patterns)
  - [Failure Patterns](#failure-patterns)
  - [Case-Insensitive Matching](#case-insensitive-matching)
  - [Pattern Precedence](#pattern-precedence)
- [Advanced Usage](#advanced-usage)
  - [Combining Parameters](#combining-parameters)
  - [Working with Complex Commands](#working-with-complex-commands)
  - [Error Handling](#error-handling)
- [Real-World Use Cases](#real-world-use-cases)
  - [Monitoring & Health Checks](#monitoring--health-checks)
  - [Development & Testing](#development--testing)
  - [System Administration](#system-administration)
  - [Data Processing](#data-processing)
- [Exit Codes for Scripting](#exit-codes-for-scripting)
- [Integration Patterns](#integration-patterns)
- [Performance Considerations](#performance-considerations)
- [Troubleshooting](#troubleshooting)
- [Tips & Best Practices](#tips--best-practices)

## Quick Start

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
# Output:
# 🕐 Interval execution: every 2s, 3 times
# 📋 Command: [echo test]
# 🚀 Starting execution...
# test
# test
# test
# ✅ Execution completed!
# 📊 Statistics:
#    Total executions: 3
#    Successful: 3
#    Failed: 0
#    Duration: 4.002s
```

### Stats-Only Mode (`--stats-only`)
Shows only execution statistics, suppresses command output:
```bash
rpr interval --every 2s --times 3 --stats-only -- echo "test"
# Output:
# ✅ Execution completed!
# 📊 Statistics:
#    Total executions: 3
#    Successful: 3
#    Failed: 0
#    Duration: 4.002s
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

### Abbreviation Examples

**Ultra-compact form:**
```bash
# Instead of: rpr interval --every 30s --times 5 -- curl http://api.com
rpr i -e 30s -t 5 -- curl http://api.com
```

**Mixed abbreviations:**
```bash
# Primary subcommand with short flags
rpr int -e 1m -f 10m -- ./health-check.sh

# Minimal subcommand with full flags  
rpr c --times 100 --every 5s -- npm test
```

**Power user shortcuts:**
```bash
# Monitor API every 10 seconds for 5 minutes
rpr i -e 10s -f 5m -- curl -f https://api.example.com/health

# Run tests 50 times with 2-second intervals
rpr c -t 50 -e 2s -- go test ./...

# Check disk space every minute for an hour
rpr d -f 1h -e 1m -- df -h /
```

## Basic Commands

### Interval Execution

Execute a command at regular time intervals.

#### Basic Syntax
```bash
rpr interval --every <duration> -- <command>
```

#### Examples

**Monitor a website every 30 seconds:**
```bash
rpr interval --every 30s -- curl -I https://example.com
```
*Use case: Check if your website is responding*

**Check disk space every 5 minutes:**
```bash
rpr interval --every 5m -- df -h /
```
*Use case: Monitor disk usage on a server*

**Ping a server every second:**
```bash
rpr interval --every 1s -- ping -c 1 google.com
```
*Use case: Test network connectivity*

#### With Limits

**Run for a specific number of times:**
```bash
rpr interval --every 10s --times 6 -- echo "Status check"
```
*Use case: Run 6 status checks over 1 minute*

**Run for a specific duration:**
```bash
rpr interval --every 2s --for 1m -- date
```
*Use case: Show timestamps every 2 seconds for 1 minute*

### Count-Based Execution

Execute a command a specific number of times.

#### Basic Syntax
```bash
rpr count --times <number> -- <command>
```

#### Examples

**Run a backup script 3 times:**
```bash
rpr count --times 3 -- ./backup.sh
```
*Use case: Ensure backup completes successfully with retries*

**Test a flaky API endpoint 10 times:**
```bash
rpr count --times 10 -- curl -f https://api.example.com/health
```
*Use case: Test API reliability*

#### With Intervals

**Run 5 times with 30-second delays:**
```bash
rpr count --times 5 --every 30s -- systemctl status nginx
```
*Use case: Check service status multiple times with delays*

### Duration-Based Execution

Execute a command continuously for a specific time period.

#### Basic Syntax
```bash
rpr duration --for <duration> -- <command>
```

#### Examples

**Monitor CPU usage for 10 minutes:**
```bash
rpr duration --for 10m -- top -n 1 | head -20
```
*Use case: Collect system performance data*

**Watch log file for 1 hour:**
```bash
rpr duration --for 1h -- tail -n 5 /var/log/app.log
```
*Use case: Monitor application logs during deployment*

#### With Intervals

**Check memory every 30 seconds for 5 minutes:**
```bash
rpr duration --for 5m --every 30s -- free -h
```
*Use case: Monitor memory usage during a specific operation*

## Advanced Scheduling

### Cron-like Scheduling

Execute commands based on cron expressions with timezone support.

#### Basic Syntax
```bash
rpr cron --cron <expression> [--timezone <tz>] -- <command>
```

#### Examples

**Daily backup at 9 AM:**
```bash
rpr cron --cron "0 9 * * *" -- ./daily-backup.sh
```
*Use case: Automated daily maintenance tasks*

**Weekday reports at 9 AM EST:**
```bash
rpr cron --cron "0 9 * * 1-5" --timezone "America/New_York" -- ./generate-report.sh
```
*Use case: Business day reporting with timezone awareness*

**Every 15 minutes:**
```bash
rpr cron --cron "*/15 * * * *" -- curl -f https://api.example.com/health
```
*Use case: Regular health checks using cron syntax*

**Using cron shortcuts:**
```bash
# Daily at midnight
rpr cron --cron "@daily" -- ./cleanup.sh

# Every hour
rpr cron --cron "@hourly" -- ./log-rotation.sh

# Weekly on Sunday at midnight
rpr cron --cron "@weekly" -- ./weekly-backup.sh
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

**Supported shortcuts:**
- `@yearly` or `@annually` - Run once a year at midnight on January 1st
- `@monthly` - Run once a month at midnight on the first day
- `@weekly` - Run once a week at midnight on Sunday
- `@daily` or `@midnight` - Run once a day at midnight
- `@hourly` - Run once an hour at the beginning of the hour

#### With Stop Conditions
```bash
# Run daily for a week
rpr cron --cron "@daily" --times 7 -- ./daily-task.sh

# Run hourly for 8 hours
rpr cron --cron "@hourly" --for 8h -- ./hourly-check.sh
```

### Rate-Limited Execution

Execute commands with server-friendly rate limiting to avoid overwhelming APIs or services.

#### Basic Syntax
```bash
rpr rate-limit --rate <rate_spec> -- <command>
```

#### Examples

**API calls with hourly rate limit:**
```bash
rpr rate-limit --rate 100/1h -- curl https://api.github.com/user
```
*Use case: Stay within API rate limits*

**Database queries with per-minute limit:**
```bash
rpr rate-limit --rate 10/1m -- mysql -e "SELECT COUNT(*) FROM users"
```
*Use case: Avoid overwhelming database with queries*

**With retry pattern:**
```bash
rpr rate-limit --rate 50/1h --retry-pattern 0,5m,15m -- curl https://api.example.com
```
*Use case: Retry failed requests with exponential delays*

### Adaptive Scheduling

Automatically adjust execution intervals based on command response times and success rates.

#### Basic Syntax
```bash
rpr adaptive --base-interval <duration> [OPTIONS] -- <command>
```

#### Examples

**API monitoring with adaptive intervals:**
```bash
rpr adaptive --base-interval 1s --show-metrics -- curl https://api.example.com/health
```
*Use case: Increase frequency when API is fast, decrease when slow*

**Database health check with bounds:**
```bash
rpr adaptive --base-interval 30s --min-interval 10s --max-interval 5m -- mysql -e "SELECT 1"
```
*Use case: Adaptive monitoring with safety bounds*

### Plugin-Based Scheduling

Use custom scheduling algorithms via the plugin system.

#### Basic Syntax
```bash
rpr <plugin-name> [PLUGIN-OPTIONS] -- <command>
```

#### Examples

**Fibonacci backoff scheduler:**
```bash
rpr fibonacci --base-interval 1s --max-interval 5m -- curl https://api.example.com
```
*Use case: Custom backoff pattern using Fibonacci sequence*

**Machine learning adaptive scheduler:**
```bash
rpr ml-adaptive --learning-rate 0.1 --history-window 100 -- ./performance-test.sh
```
*Use case: AI-driven scheduling based on historical performance*

**Chaos scheduler for testing:**
```bash
rpr chaos --min-interval 100ms --max-interval 10s --randomness 0.3 -- ./resilience-test.sh
```
*Use case: Introduce controlled randomness for chaos engineering*

#### Plugin Management
```bash
# List available plugins
rpr plugins list

# Show plugin information
rpr plugins info fibonacci

# Install new plugins
rpr plugins install fibonacci-scheduler

# Update plugins
rpr plugins update --all
```

#### Plugin Configuration
```toml
# ~/.repeater/config.toml
[plugins.fibonacci]
base_interval = "1s"
max_interval = "5m"
reset_threshold = 10

[plugins.ml-adaptive]
learning_rate = 0.1
history_window = 100
```

**Load testing with failure threshold:**
```bash
rpr adaptive --base-interval 500ms --failure-threshold 0.2 --times 100 -- curl https://api.com/test
```
*Use case: Back off when failure rate exceeds 20%*

## Pattern Matching

Pattern matching allows you to define success and failure conditions based on command output rather than just exit codes. This is particularly useful for commands that don't follow standard Unix exit code conventions or when you need to detect specific conditions in the output.

### Success Patterns

Define regex patterns that indicate success regardless of the command's exit code.

#### Basic Syntax
```bash
rpr <subcommand> --success-pattern <regex> -- <command>
```

#### Examples

**Monitor deployment success:**
```bash
rpr interval --every 30s --times 10 --success-pattern "deployment successful" -- ./deploy.sh
```
*Use case: Treat deployment as successful when output contains "deployment successful", even if exit code is non-zero*

**API health check with JSON response:**
```bash
rpr count --times 5 --success-pattern '"status":\s*"ok"' -- curl -s https://api.example.com/health
```
*Use case: Consider API healthy when JSON response contains `"status": "ok"`*

**Database connection verification:**
```bash
rpr interval --every 1m --success-pattern "1" -- mysql -e "SELECT 1"
```
*Use case: Verify database connectivity by checking for "1" in output*

### Failure Patterns

Define regex patterns that indicate failure regardless of the command's exit code.

#### Basic Syntax
```bash
rpr <subcommand> --failure-pattern <regex> -- <command>
```

#### Examples

**Detect errors in application logs:**
```bash
rpr duration --for 1h --every 5m --failure-pattern "(?i)error|exception|failed" -- tail -n 10 /var/log/app.log
```
*Use case: Monitor logs for error conditions even when tail command succeeds*

**API monitoring with error detection:**
```bash
rpr interval --every 30s --failure-pattern "(?i)timeout|unavailable|error" -- curl -s https://api.example.com/status
```
*Use case: Detect API issues from response content, not just HTTP status*

**Service health monitoring:**
```bash
rpr count --times 10 --failure-pattern "down|inactive|failed" -- systemctl status nginx
```
*Use case: Detect service problems from status output*

### Case-Insensitive Matching

Make pattern matching case-insensitive for more flexible detection.

#### Basic Syntax
```bash
rpr <subcommand> --case-insensitive --success-pattern <pattern> -- <command>
# or
rpr <subcommand> --failure-pattern "(?i)<pattern>" -- <command>
```

#### Examples

**Case-insensitive success detection:**
```bash
rpr interval --every 1m --case-insensitive --success-pattern "healthy|ok|running" -- ./health-check.sh
```
*Use case: Match "HEALTHY", "Ok", "Running", etc.*

**Case-insensitive error detection:**
```bash
rpr duration --for 30m --every 2m --case-insensitive --failure-pattern "error|warning|critical" -- ./system-check.sh
```
*Use case: Catch "ERROR", "Warning", "CRITICAL", etc.*

### Pattern Precedence

When both success and failure patterns are specified, failure patterns take precedence. This ensures that errors are always detected even when success patterns also match.

#### Examples

**Comprehensive monitoring with both patterns:**
```bash
rpr interval --every 30s --success-pattern "status.*ok" --failure-pattern "(?i)error|timeout|failed" -- curl -s https://api.example.com/health
```
*Use case: Consider successful when status is ok, but always fail on errors*

**Log monitoring with precedence:**
```bash
rpr duration --for 1h --every 5m --success-pattern "completed" --failure-pattern "(?i)error|exception" -- tail -n 5 /var/log/process.log
```
*Use case: Success when process completes, but failure patterns override for errors*

### Advanced Pattern Examples

**Complex regex patterns:**
```bash
# Monitor HTTP status codes
rpr count --times 10 --success-pattern "HTTP/1\.[01] [23][0-9][0-9]" -- curl -I https://api.example.com

# Detect specific error codes
rpr interval --every 1m --failure-pattern "HTTP/1\.[01] [45][0-9][0-9]" -- curl -I https://api.example.com

# Monitor numeric thresholds
rpr duration --for 10m --every 30s --failure-pattern "[89][0-9]|100" -- sh -c "df / | tail -1 | awk '{print \$5}' | sed 's/%//'"
```

**Integration with adaptive scheduling:**
```bash
# Adaptive scheduling with pattern-based success detection
rpr adaptive --base-interval 1s --min-interval 500ms --max-interval 10s \
    --success-pattern "healthy" --failure-pattern "(?i)error|down" \
    --show-metrics -- ./service-check.sh
```
*Use case: Adaptive scheduling adjusts based on pattern matching results, not just exit codes*

**Pattern matching with output control:**
```bash
# Quiet mode with pattern matching for automation
rpr interval --every 30s --times 20 --quiet \
    --success-pattern "backup completed" --failure-pattern "(?i)error|failed" \
    -- ./backup-script.sh

# Stats-only mode to see pattern matching effectiveness
rpr count --times 50 --stats-only \
    --success-pattern "ok" --failure-pattern "(?i)error|timeout" \
    -- curl -s https://api.example.com/health
```

### Exponential Backoff

Implement exponential backoff for resilient execution against unreliable services.

#### Basic Syntax
```bash
rpr backoff --initial <duration> [OPTIONS] -- <command>
```

#### Examples

**Retry unreliable API with backoff:**
```bash
rpr backoff --initial 100ms --max 30s --multiplier 2.0 -- curl https://flaky-api.com
```
*Use case: Resilient API calls with exponential delays*

**Database connection with jitter:**
```bash
rpr backoff --initial 1s --max 60s --jitter 0.1 --times 10 -- mysql -e "SELECT 1"
```
*Use case: Avoid thundering herd with randomized delays*

### Load-Aware Scheduling

Automatically adjust execution frequency based on system resource usage (CPU, memory, load average).

#### Basic Syntax
```bash
rpr load-adaptive --base-interval <duration> [OPTIONS] -- <command>
```

#### Examples

**CPU-intensive task with load awareness:**
```bash
rpr load-adaptive --base-interval 1s --target-cpu 70 -- ./cpu-intensive-task.sh
```
*Use case: Scale back when CPU usage is high*

**Memory-sensitive processing:**
```bash
rpr load-adaptive --base-interval 30s --target-memory 80 --target-load 1.5 -- ./process-data.sh
```
*Use case: Adjust frequency based on memory pressure and system load*

**Development environment monitoring:**
```bash
rpr load-adaptive --base-interval 5s --target-cpu 60 --show-metrics -- npm test
```
*Use case: Run tests more frequently when system is idle*

## Advanced Usage

### Combining Parameters

You can combine different parameters for sophisticated execution patterns:

**Limited interval execution:**
```bash
rpr interval --every 1m --times 10 --for 15m -- curl -s https://api.example.com/metrics
```
*Use case: Collect metrics every minute, but stop after 10 calls or 15 minutes (whichever comes first)*

### Working with Complex Commands

**Commands with pipes and redirections:**
```bash
rpr interval --every 5s -- sh -c "ps aux | grep nginx | wc -l"
```
*Use case: Count nginx processes every 5 seconds*

**Commands with multiple arguments:**
```bash
rpr count --times 3 -- rsync -av --progress /source/ user@server:/backup/
```
*Use case: Retry file synchronization up to 3 times*

**Commands with environment variables:**
```bash
rpr interval --every 1m -- sh -c 'echo "$(date): $USER logged in"'
```
*Use case: Log user activity with timestamps*

### Error Handling

Repeater continues execution even if individual commands fail:

**Test unreliable service:**
```bash
rpr interval --every 10s --times 20 -- curl -f --max-time 5 https://unreliable-api.com
```
*Use case: Test service reliability over time, allowing for individual failures*

## Real-World Use Cases

### Monitoring & Health Checks

**Website uptime monitoring with logging:**
```bash
# Check every 30 seconds for 1 hour, log status codes
rpr duration --for 1h --every 30s -- curl -w "%{http_code}\n" -s -o /dev/null https://mysite.com | tee uptime.log
# Abbreviated form:
rpr d -f 1h -e 30s -- curl -w "%{http_code}\n" -s -o /dev/null https://mysite.com | tee uptime.log
```

**Database connection testing with pattern matching:**
```bash
# Test connection 10 times using pattern matching
rpr count --times 10 --every 5s --success-pattern "1" -- mysql -h db.example.com -u user -p -e "SELECT 1"
# Abbreviated form:
rpr c -t 10 -e 5s --success-pattern "1" -- mysql -h db.example.com -u user -p -e "SELECT 1"
```

**SSL certificate expiration check:**
```bash
# Check daily for a week
rpr interval --every 24h --times 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
# Abbreviated form:
rpr i -e 24h -t 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
```

### Development & Testing

**API load testing with success detection:**
```bash
# Hit API endpoint 100 times, detect success from response
rpr count --times 100 --success-pattern '"status":\s*"success"' -- curl -s https://api.example.com/endpoint
```

**Build system monitoring with pattern matching:**
```bash
# Check build status every 2 minutes, detect failures
rpr duration --for 8h --every 2m --success-pattern "passed|success" --failure-pattern "(?i)failed|error" -- curl -s https://ci.example.com/api/build/status
```

**Database migration progress with completion detection:**
```bash
# Monitor migration with pattern-based completion detection
rpr duration --for 30m --every 10s --success-pattern "migration completed" --failure-pattern "(?i)error|failed" -- ./check-migration-status.sh
```

### System Administration

**Log rotation monitoring:**
```bash
# Check log file sizes every hour
rpr interval --every 1h -- du -sh /var/log/*.log
```

**Service restart verification:**
```bash
# Verify service is running after restart, check 5 times over 1 minute
rpr count --times 5 --every 12s -- systemctl is-active my-service
```

**Disk cleanup verification:**
```bash
# Monitor disk space during cleanup operation
rpr duration --for 10m --every 30s -- df -h /tmp
```

### Data Processing

**Batch job monitoring:**
```bash
# Check job queue every 5 minutes for 2 hours
rpr duration --for 2h --every 5m -- qstat -u $USER
```

**File processing pipeline:**
```bash
# Process files as they arrive, check every 10 seconds for 1 hour
rpr duration --for 1h --every 10s -- find /incoming -name "*.csv" -exec ./process.sh {} \;
```

**Data synchronization:**
```bash
# Sync data every 15 minutes, up to 8 times (2 hours)
rpr count --times 8 --every 15m -- rsync -av /local/data/ remote:/backup/data/
```

## Integration Patterns

Repeater integrates seamlessly with existing Unix toolchains and modern infrastructure. Here are common integration patterns:

### With Monitoring Systems

**Prometheus metrics collection:**
```bash
# Collect metrics and format for Prometheus
rpr interval --every 30s -- curl -s http://localhost:8080/metrics | \
  grep -E '^(cpu_usage|memory_usage)' | \
  awk '{print $1 " " $2 " " systime()}' >> /var/lib/prometheus/node_exporter/repeater.prom
```

**Grafana dashboard data:**
```bash
# Generate time-series data for Grafana
rpr i -e 1m -- sh -c 'echo "$(date +%s),$(df / | tail -1 | awk "{print \$5}" | sed "s/%//")"; sleep 1' | \
  tee -a /var/log/disk-usage.csv
```

**Nagios/Icinga integration:**
```bash
# Health check with Nagios-compatible exit codes
rpr count --times 3 --quiet -- curl -f --max-time 10 https://api.example.com/health
case $? in
    0) echo "OK - API is healthy"; exit 0 ;;
    1) echo "CRITICAL - API health check failed"; exit 2 ;;
    *) echo "UNKNOWN - Unexpected error"; exit 3 ;;
esac
```

### With CI/CD Pipelines

**GitHub Actions integration:**
```yaml
# .github/workflows/health-check.yml
- name: API Health Check
  run: |
    rpr count --times 5 --every 10s --quiet -- \
      curl -f https://staging.example.com/health
    if [ $? -ne 0 ]; then
      echo "::error::Staging environment health check failed"
      exit 1
    fi
```

**Jenkins pipeline:**
```groovy
pipeline {
    stages {
        stage('Health Check') {
            steps {
                sh '''
                    rpr interval --every 30s --times 10 --quiet -- \
                      curl -f ${DEPLOYMENT_URL}/health
                    if [ $? -ne 0 ]; then
                        echo "Deployment health check failed"
                        exit 1
                    fi
                '''
            }
        }
    }
}
```

**GitLab CI integration:**
```yaml
health_check:
  script:
    - rpr count --times 3 --every 5s --stats-only -- curl -f $API_URL/health
    - echo "Health check completed with exit code $?"
  only:
    - deploy
```

### With Container Orchestration

**Docker health checks:**
```dockerfile
# Dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD rpr count --times 1 --quiet -- curl -f http://localhost:8080/health || exit 1
```

**Kubernetes liveness probe:**
```yaml
# k8s-deployment.yaml
livenessProbe:
  exec:
    command:
    - /bin/sh
    - -c
    - rpr count --times 1 --quiet -- curl -f http://localhost:8080/health
  initialDelaySeconds: 30
  periodSeconds: 30
```

**Docker Compose monitoring:**
```yaml
# docker-compose.yml
services:
  monitor:
    image: alpine:latest
    command: >
      sh -c "apk add --no-cache curl &&
             rpr interval --every 60s -- 
               curl -f http://app:8080/health || 
               docker-compose restart app"
    depends_on:
      - app
```

### With Log Aggregation

**ELK Stack integration:**
```bash
# Send structured logs to Elasticsearch
rpr interval --every 5m -- sh -c '
  status=$(curl -s -w "%{http_code}" -o /dev/null https://api.example.com)
  echo "{\"timestamp\":\"$(date -Iseconds)\",\"service\":\"api\",\"status\":$status}" | \
  curl -X POST "http://elasticsearch:9200/health-checks/_doc" \
       -H "Content-Type: application/json" -d @-
'
```

**Fluentd log shipping:**
```bash
# Generate logs in fluentd-compatible format
rpr i -e 30s -- sh -c '
  response_time=$(curl -w "%{time_total}" -s -o /dev/null https://api.com)
  echo "$(date -Iseconds) health.api response_time=$response_time"
' | tee -a /var/log/fluentd/health.log
```

### With Parallel Processing

**GNU Parallel integration:**
```bash
# Monitor multiple services in parallel
echo "api.com db.com cache.com" | tr ' ' '\n' | \
parallel -j3 'rpr count --times 5 --quiet -- curl -f https://{}/health && echo "{} OK" || echo "{} FAILED"'
```

**xargs parallel execution:**
```bash
# Process multiple URLs concurrently
echo "url1 url2 url3" | xargs -n1 -P3 -I{} \
  rpr count --times 1 --quiet -- curl -f {}
```

### With System Services

**systemd service monitoring:**
```ini
# /etc/systemd/system/api-monitor.service
[Unit]
Description=API Health Monitor
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rpr interval --every 60s --quiet -- curl -f http://localhost:8080/health
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**cron integration:**
```bash
# /etc/crontab - Run health check every 5 minutes
*/5 * * * * root rpr count --times 1 --quiet -- /usr/local/bin/health-check.sh || logger "Health check failed"
```

## Performance Considerations

Understanding performance characteristics helps you choose the right scheduling mode and configuration for your use case.

### Choosing Scheduling Modes

#### **Interval Mode** - Best for regular monitoring
```bash
# Good for: Regular health checks, periodic data collection
rpr interval --every 30s -- curl https://api.example.com/health

# Performance: Predictable resource usage, consistent timing
# Use when: You need regular, evenly-spaced execution
```

#### **Adaptive Mode** - Best for variable workloads
```bash
# Good for: API monitoring, services with variable response times
rpr adaptive --base-interval 1s --show-metrics -- curl https://api.example.com

# Performance: Automatically adjusts to system conditions
# Use when: Command execution time varies significantly
```

#### **Load-Adaptive Mode** - Best for resource-aware execution
```bash
# Good for: CPU/memory intensive tasks, shared systems
rpr load-adaptive --base-interval 30s --target-cpu 70 -- ./process-data.sh

# Performance: Scales with available system resources
# Use when: You want to avoid overwhelming the system
```

#### **Backoff Mode** - Best for unreliable services
```bash
# Good for: External APIs, flaky network services
rpr backoff --initial 100ms --max 30s -- curl https://external-api.com

# Performance: Reduces load on failing services
# Use when: Dealing with unreliable external dependencies
```

### Resource Usage Guidelines

#### **Memory Usage**
```bash
# Low memory usage (< 10MB)
rpr interval --every 1s --quiet -- echo "test"

# Moderate memory usage (10-50MB) 
rpr interval --every 1s -- curl -s https://api.com | jq .

# High memory usage (> 50MB)
rpr interval --every 1s -- ./data-processing-script.sh
```

#### **CPU Usage**
```bash
# Minimal CPU impact
rpr interval --every 10s --quiet -- curl -f https://api.com

# Moderate CPU usage
rpr adaptive --base-interval 1s -- curl https://api.com | jq . | awk '{print $1}'

# CPU-intensive (use load-adaptive)
rpr load-adaptive --base-interval 5s --target-cpu 60 -- ./cpu-intensive-task.sh
```

#### **Network Usage**
```bash
# Light network usage
rpr interval --every 60s -- curl -I https://api.com

# Heavy network usage (consider rate limiting)
rpr rate-limit --rate 10/1m -- curl -s https://api.com/large-data

# Burst network usage
rpr count --times 100 --every 100ms -- curl -s https://api.com/small-endpoint
```

### Scaling Recommendations

#### **Single Instance Limits**
- **Maximum frequency**: ~100 executions/second (depends on command complexity)
- **Recommended intervals**: ≥ 100ms for simple commands, ≥ 1s for complex commands
- **Memory limit**: Scales with command output size and frequency

#### **Multi-Instance Coordination**
```bash
# Use different intervals to avoid thundering herd
# Instance 1:
rpr interval --every 60s -- ./health-check.sh

# Instance 2 (offset by 30s):
sleep 30 && rpr interval --every 60s -- ./health-check.sh
```

#### **Horizontal Scaling Patterns**
```bash
# Distribute load across multiple hosts
# Host 1: Monitor services 1-10
rpr interval --every 30s -- ./monitor-services.sh 1 10

# Host 2: Monitor services 11-20  
rpr interval --every 30s -- ./monitor-services.sh 11 20
```

### Performance Tuning Tips

#### **Optimize Command Execution**
```bash
# Slow: Multiple separate calls
rpr interval --every 30s -- sh -c 'curl api1.com && curl api2.com && curl api3.com'

# Fast: Single call with parallel processing
rpr interval --every 30s -- sh -c 'curl api1.com & curl api2.com & curl api3.com & wait'
```

#### **Reduce Output Processing**
```bash
# Slow: Processing large output
rpr interval --every 1s -- curl -s https://api.com/large-response | jq .

# Fast: Filter at source
rpr interval --every 1s -- curl -s https://api.com/large-response | jq -r '.status'
```

#### **Use Appropriate Output Modes**
```bash
# For automation: Use --quiet to reduce overhead
rpr interval --every 100ms --quiet -- ./fast-check.sh

# For debugging: Use --verbose only when needed
rpr interval --every 5s --verbose -- ./debug-command.sh

# For monitoring: Use --stats-only for metrics
rpr interval --every 1m --stats-only -- ./performance-test.sh
```

## Troubleshooting

Common issues and their solutions when using Repeater in production environments.

### Common Issues

#### **Pipeline Not Working**
```bash
# Problem: Command works manually but not in pipeline
rpr count --times 3 -- echo "test" | wc -l
# Returns: 1 (instead of 3)

# Solution: Check command quoting and shell interpretation
rpr count --times 3 -- sh -c 'echo "test"' | wc -l
# Returns: 3 (correct)

# Explanation: Without sh -c, each execution is a separate pipeline
```

#### **High CPU Usage**
```bash
# Problem: CPU usage is unexpectedly high
rpr interval --every 100ms -- curl https://api.com

# Solution 1: Increase interval
rpr interval --every 1s -- curl https://api.com

# Solution 2: Use load-adaptive mode
rpr load-adaptive --base-interval 100ms --target-cpu 70 -- curl https://api.com

# Solution 3: Use --quiet to reduce output processing
rpr interval --every 100ms --quiet -- curl https://api.com
```

#### **Memory Issues**
```bash
# Problem: Memory usage grows over time
rpr interval --every 1s -- ./script-with-memory-leak.sh

# Solution 1: Use --quiet to reduce output buffering
rpr interval --every 1s --quiet -- ./script-with-memory-leak.sh

# Solution 2: Restart periodically using external wrapper
while true; do
    timeout 1h rpr interval --every 1s --quiet -- ./script.sh
    sleep 5
done
```

#### **Commands Not Executing**
```bash
# Problem: Commands appear to hang or not execute
rpr interval --every 30s -- ./slow-script.sh

# Solution 1: Check if command is actually slow
time ./slow-script.sh

# Solution 2: Add timeout to prevent hanging
rpr interval --every 30s -- timeout 25s ./slow-script.sh

# Solution 3: Use verbose mode to see what's happening
rpr interval --every 30s --verbose -- ./slow-script.sh
```

### Debugging Tips

#### **Test Commands Manually First**
```bash
# Always test your command manually before using with rpr
./your-command.sh
echo "Exit code: $?"

# Then test with rpr
rpr count --times 1 --verbose -- ./your-command.sh
```

#### **Use Verbose Mode for Debugging**
```bash
# See detailed execution information
rpr count --times 3 --verbose -- curl https://api.example.com

# Check timing and statistics
rpr interval --every 5s --times 3 --verbose -- ./test-command.sh
```

#### **Check Exit Codes**
```bash
# Capture and examine exit codes
rpr count --times 5 --quiet -- ./your-command.sh
echo "Repeater exit code: $?"

# Test individual command exit codes
./your-command.sh
echo "Command exit code: $?"
```

#### **Isolate Pipeline Issues**
```bash
# Test without pipeline first
rpr count --times 3 -- echo "test"

# Then add pipeline components one by one
rpr count --times 3 -- echo "test" | cat
rpr count --times 3 -- echo "test" | wc -l
```

### Error Scenarios and Solutions

#### **Network Timeouts**
```bash
# Problem: Network requests timing out
rpr interval --every 30s -- curl https://slow-api.com

# Solution: Add explicit timeout and retry logic
rpr interval --every 30s -- sh -c '
    curl --max-time 10 --retry 2 --retry-delay 1 https://slow-api.com || 
    echo "Request failed at $(date)"
'
```

#### **Permission Issues**
```bash
# Problem: Permission denied errors
rpr interval --every 1m -- ./script.sh
# Error: Permission denied

# Solution 1: Check script permissions
chmod +x ./script.sh

# Solution 2: Use absolute paths
rpr interval --every 1m -- /full/path/to/script.sh

# Solution 3: Run with appropriate user
sudo -u appuser rpr interval --every 1m -- ./script.sh
```

#### **Environment Variable Issues**
```bash
# Problem: Environment variables not available
rpr interval --every 1m -- echo $HOME
# Output: (empty)

# Solution: Use sh -c to preserve environment
rpr interval --every 1m -- sh -c 'echo $HOME'

# Or export variables explicitly
export HOME=/home/user
rpr interval --every 1m -- sh -c 'echo $HOME'
```

#### **Signal Handling Issues**
```bash
# Problem: Repeater not stopping cleanly
rpr interval --every 1s -- ./long-running-command.sh
# Ctrl+C doesn't stop immediately

# Solution: Commands should handle signals properly
rpr interval --every 1s -- timeout 30s ./long-running-command.sh

# Or use shorter intervals
rpr interval --every 5s -- ./quick-command.sh
```

### Performance Debugging

#### **Identify Bottlenecks**
```bash
# Monitor system resources while running
rpr interval --every 1s --verbose -- ./test-command.sh &
top -p $!

# Check I/O usage
iotop -p $(pgrep rpr)

# Monitor network usage
nethogs
```

#### **Profile Command Execution**
```bash
# Time individual executions
rpr count --times 10 --verbose -- time ./your-command.sh

# Use stats-only mode to see execution patterns
rpr interval --every 1s --times 60 --stats-only -- ./your-command.sh
```

#### **Memory Leak Detection**
```bash
# Monitor memory usage over time
rpr interval --every 10s -- sh -c 'ps -o pid,vsz,rss,comm -p $(pgrep rpr)' &
rpr interval --every 1s --times 1000 -- ./test-command.sh
```

### Getting Help

#### **Enable Debug Output**
```bash
# Use verbose mode for detailed information
rpr interval --every 5s --times 3 --verbose -- ./debug-me.sh

# Check system logs
journalctl -f | grep rpr
```

#### **Collect Diagnostic Information**
```bash
# System information
uname -a
rpr --version

# Resource usage
free -h
df -h
ps aux | grep rpr

# Network connectivity (if applicable)
ping -c 3 api.example.com
curl -I https://api.example.com
```

## Tips & Best Practices

### Duration Formats

Repeater supports various duration formats:
- `s` - seconds: `30s`, `45s`
- `m` - minutes: `5m`, `30m`
- `h` - hours: `2h`, `24h`
- Combined: `1h30m`, `2m30s`

### Command Best Practices

**Use absolute paths for scripts:**
```bash
rpr interval --every 1m -- /home/user/scripts/backup.sh
```

**Quote complex commands:**
```bash
rpr count --times 5 -- sh -c "echo 'Current time:' && date"
```

**Handle command failures gracefully:**
```bash
rpr interval --every 30s -- sh -c "curl -f https://api.com || echo 'API down at $(date)'"
```

### Performance Considerations

**For high-frequency execution:**
```bash
# Use shorter commands for sub-second intervals
rpr interval --every 100ms -- echo "tick"
```

**For long-running commands:**
```bash
# Ensure intervals are longer than command execution time
rpr interval --every 5m -- ./long-running-backup.sh
```

### Debugging and Logging

**Capture output to file:**
```bash
rpr interval --every 1m -- date >> /tmp/timestamps.log
```

**Add timestamps to output:**
```bash
rpr count --times 10 -- sh -c "echo '[$(date)]' && curl -s https://api.com"
```

**Separate success and error logs:**
```bash
rpr interval --every 30s -- sh -c "curl -s https://api.com >> success.log 2>> error.log"
```

### Signal Handling

- **Ctrl+C** (SIGINT): Gracefully stops execution after current command completes
- **SIGTERM**: Gracefully stops execution (useful in scripts and containers)

## Exit Codes for Scripting

Repeater follows Unix conventions for exit codes, making it perfect for use in scripts and automation:

### Exit Code Reference

- **0**: All commands executed successfully
- **1**: Some commands failed during execution  
- **2**: Usage error (invalid arguments, configuration issues)
- **130**: Interrupted by user (Ctrl+C, SIGINT, SIGTERM)

### Scripting Examples

**Basic success/failure handling:**
```bash
if rpr interval --every 5s --times 3 --quiet -- curl -f https://api.example.com; then
    echo "API is healthy"
else
    echo "API check failed with exit code $?"
fi
```

**Chain with other Unix tools:**
```bash
rpr count --times 5 -- curl -s https://api.example.com | jq .status && echo "Success" || echo "Failed"
```

**Use in conditional execution:**
```bash
# Only proceed if health check passes
rpr i -e 2s -t 3 --quiet -- curl -f https://api.com/health && ./deploy.sh
```

**Capture and handle different exit codes:**
```bash
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

---

## Getting Help

```bash
# Show general help
rpr --help

# Show version
rpr --version

# Show subcommand help
rpr interval --help
rpr count --help
rpr duration --help
rpr cron --help
rpr adaptive --help
rpr backoff --help
rpr load-adaptive --help
rpr rate-limit --help

# Show plugin help
rpr plugins --help
rpr plugins list
rpr plugins info <plugin-name>
```

For more examples and advanced usage patterns, see the project documentation at [github.com/swi/repeater](https://github.com/swi/repeater).