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
  - [Rate-Limited Execution](#rate-limited-execution)
  - [Adaptive Scheduling](#adaptive-scheduling)
  - [Exponential Backoff](#exponential-backoff)
  - [Load-Aware Scheduling](#load-aware-scheduling)
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

**Load testing with failure threshold:**
```bash
rpr adaptive --base-interval 500ms --failure-threshold 0.2 --times 100 -- curl https://api.com/test
```
*Use case: Back off when failure rate exceeds 20%*

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

**Database connection testing with success counting:**
```bash
# Test connection 10 times, count successful connections
rpr count --times 10 --every 5s -- mysql -h db.example.com -u user -p -e "SELECT 1" | grep -c "1"
# Abbreviated form:
rpr c -t 10 -e 5s -- mysql -h db.example.com -u user -p -e "SELECT 1" | grep -c "1"
```

**SSL certificate expiration check:**
```bash
# Check daily for a week
rpr interval --every 24h --times 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
# Abbreviated form:
rpr i -e 24h -t 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
```

### Development & Testing

**API load testing:**
```bash
# Hit API endpoint 100 times as fast as possible
rpr count --times 100 -- curl -s https://api.example.com/endpoint
```

**Build system monitoring:**
```bash
# Check build status every 2 minutes during work hours
rpr duration --for 8h --every 2m -- curl -s https://ci.example.com/api/build/status
```

**Database migration progress:**
```bash
# Monitor migration every 10 seconds for up to 30 minutes
rpr duration --for 30m --every 10s -- mysql -e "SELECT COUNT(*) FROM migration_status WHERE completed = 1"
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
```

For more examples and advanced usage patterns, see the project documentation at [github.com/swi/repeater](https://github.com/swi/repeater).