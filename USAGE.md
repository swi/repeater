# Repeater (rpr) Usage Guide

A comprehensive guide to using the `rpr` command-line tool for continuous command execution with intelligent scheduling.

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Commands](#basic-commands)
  - [Interval Execution](#interval-execution)
  - [Count-Based Execution](#count-based-execution)
  - [Duration-Based Execution](#duration-based-execution)
- [Advanced Usage](#advanced-usage)
  - [Combining Parameters](#combining-parameters)
  - [Working with Complex Commands](#working-with-complex-commands)
  - [Error Handling](#error-handling)
- [Real-World Use Cases](#real-world-use-cases)
  - [Monitoring & Health Checks](#monitoring--health-checks)
  - [Development & Testing](#development--testing)
  - [System Administration](#system-administration)
  - [Data Processing](#data-processing)
- [Tips & Best Practices](#tips--best-practices)

## Quick Start

The basic syntax for `rpr` is:
```bash
rpr [GLOBAL OPTIONS] <SUBCOMMAND> [OPTIONS] -- <COMMAND>
```

**Key Points:**
- Use `--` to separate repeater options from the command you want to run
- Commands are executed exactly as you would run them manually
- All output is preserved and displayed in real-time

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

**Website uptime monitoring:**
```bash
# Check every 30 seconds for 1 hour
rpr duration --for 1h --every 30s -- curl -f -s -o /dev/null https://mysite.com
```

**Database connection testing:**
```bash
# Test connection 10 times with 5-second intervals
rpr count --times 10 --every 5s -- mysql -h db.example.com -u user -p -e "SELECT 1"
```

**SSL certificate expiration check:**
```bash
# Check daily for a week
rpr interval --every 24h --times 7 -- openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
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

### Exit Codes

- `0`: All executions completed successfully
- `1`: Configuration or argument error
- `2`: Command execution failed (when applicable)
- `130`: Interrupted by user (Ctrl+C)

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