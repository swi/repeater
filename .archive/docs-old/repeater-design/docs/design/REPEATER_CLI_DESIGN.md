# Repeater CLI Design Specification

## Design Philosophy

The repeater CLI follows these principles:
1. **Subcommand Architecture**: Different execution modes as distinct subcommands
2. **Intuitive Abbreviations**: Multiple levels of abbreviation for power users
3. **Composable Options**: Flags that work together logically
4. **Consistent Patterns**: Similar to patience and other modern CLI tools
5. **Safety First**: Require explicit stop conditions to prevent runaway processes

## Command Structure

```
rpr <subcommand> [options] -- <command> [args...]
```

## Subcommands

### Core Subcommands (MVP)

#### `interval` (aliases: `int`, `i`)
Execute command at fixed intervals.

```bash
rpr interval --every DURATION [options] -- command
rpr int --every 30s -- curl https://api.example.com
rpr i -e 30s -f 1h -- health-check.sh
```

**Required Options:**
- `--every DURATION`, `-e DURATION`: Interval between executions

**Optional Options:**
- `--for DURATION`, `-f DURATION`: Stop after duration
- `--times COUNT`, `-t COUNT`: Stop after count executions
- `--immediate`, `--now`, `-n`: Run immediately, then start intervals
- `--jitter PERCENT`, `-j PERCENT`: Add randomization (e.g., `20%`)

#### `count` (aliases: `cnt`, `c`)
Execute command a specific number of times.

```bash
rpr count --times COUNT [options] -- command
rpr cnt --times 100 -- npm test
rpr c -t 100 -e 10s -- load-test.sh
```

**Required Options:**
- `--times COUNT`, `-t COUNT`: Number of executions

**Optional Options:**
- `--every DURATION`, `-e DURATION`: Interval between executions (default: immediate)
- `--parallel N`, `-p N`: Run up to N executions in parallel
- `--for DURATION`, `-f DURATION`: Stop early if duration exceeded

#### `duration` (aliases: `dur`, `d`)
Execute command for a specific time period.

```bash
rpr duration --for DURATION [options] -- command
rpr dur --for 1h -- monitoring.sh
rpr d -f 2h -e 5m -q -- backup.sh
```

**Required Options:**
- `--for DURATION`, `-f DURATION`: How long to keep running

**Optional Options:**
- `--every DURATION`, `-e DURATION`: Interval between executions (default: immediate)
- `--until TIME`, `-u TIME`: Stop at specific time (e.g., `15:30`, `2024-01-01T00:00:00Z`)

### Future Subcommands

#### `rate-limit` (aliases: `rate`, `rl`, `r`)
Execute command within rate limits.

```bash
rpr rate-limit --limit RATE [options] -- command
rpr rate --limit 100/1h -- api-call.sh
rpr r -l 1000/1h --burst 10 -- curl https://api.example.com
```

#### `schedule` (aliases: `sched`, `s`)
Execute command on cron-like schedule.

```bash
rpr schedule --cron "0 */6 * * *" -- backup.sh
rpr s --cron "*/15 * * * *" --timezone UTC -- monitoring.sh
```

#### `adaptive` (aliases: `adapt`, `a`)
Execute command with learning-based intervals.

```bash
rpr adaptive --target-latency 100ms -- health-check.sh
rpr a --learn-from-response --min 10s --max 5m -- api-monitor.sh
```

## Global Options

### Output Control
```bash
--quiet, -q                 # Suppress command output
--verbose, -v               # Show detailed execution information
--output-file FILE, -o FILE # Redirect output to file
--aggregate, -a             # Collect all output, show summary at end
--log-file FILE             # Log execution details to file
```

### Error Handling
```bash
--continue-on-error, --keep-going, -k  # Continue even if command fails
--max-failures N                       # Stop after N consecutive failures
--success-pattern REGEX                # Define success via output pattern
--failure-pattern REGEX                # Define failure via output pattern
```

### Process Control
```bash
--timeout DURATION, -T DURATION        # Timeout per command execution
--kill-timeout DURATION                # Force kill timeout
--working-dir PATH, -C PATH             # Set working directory
--env KEY=VALUE                         # Set environment variables
```

### Advanced Options
```bash
--dry-run                   # Show what would be executed without running
--metrics                   # Collect and display execution metrics
--config FILE               # Load configuration from file
--daemon                    # Enable daemon coordination (for rate limiting)
--resource-id ID            # Resource identifier for shared coordination
```

## Duration Format

### Standard Format
```bash
30s     # 30 seconds
5m      # 5 minutes  
2h      # 2 hours
1d      # 1 day
1w      # 1 week
```

### Compound Durations
```bash
1h30m   # 1 hour 30 minutes
2d12h   # 2 days 12 hours
```

### Rate Format (Future)
```bash
100/1h      # 100 per hour
10/1m       # 10 per minute
1000/1d     # 1000 per day
```

## Usage Examples

### Basic Examples
```bash
# Health check every 30 seconds for 1 hour
rpr interval --every 30s --for 1h -- curl -f localhost:8080/health

# Run test suite 50 times with 10-second intervals
rpr count --times 50 --every 10s -- npm test

# Monitor system for 8 hours, checking every 5 minutes
rpr duration --for 8h --every 5m -- ./system-monitor.sh
```

### Abbreviated Examples
```bash
# Same examples with abbreviations
rpr i -e 30s -f 1h -- curl -f localhost:8080/health
rpr c -t 50 -e 10s -- npm test  
rpr d -f 8h -e 5m -- ./system-monitor.sh
```

### Advanced Examples
```bash
# Continue on errors, log to file, with jitter
rpr i -e 30s -f 1h -k -o health.log -j 10% -- health-check.sh

# Parallel execution with failure threshold
rpr c -t 100 -p 5 --max-failures 10 -- parallel-task.sh

# Quiet execution with custom timeout
rpr d -f 2h -e 1m -q -T 30s -- slow-command.sh
```

## Help System

### Main Help
```bash
rpr --help
rpr -h
```

### Subcommand Help
```bash
rpr interval --help
rpr i -h
rpr count --help
rpr c -h
```

### Examples Help
```bash
rpr examples                    # Show common usage patterns
rpr interval examples           # Show interval-specific examples
```

## Error Messages

### Clear Error Messages
```bash
# Missing required option
$ rpr interval -- echo hello
Error: --every is required for interval subcommand
Try: rpr interval --every 30s -- echo hello

# Invalid duration format
$ rpr interval --every 30x -- echo hello  
Error: invalid duration format "30x"
Valid formats: 30s, 5m, 2h, 1d, 1w

# Missing stop condition
$ rpr interval --every 30s -- echo hello
Error: no stop condition specified (use --for, --times, or Ctrl+C)
Try: rpr interval --every 30s --for 1h -- echo hello
```

### Helpful Suggestions
```bash
# Command not found
$ rpr inter --every 30s -- echo hello
Error: unknown subcommand "inter"
Did you mean: interval (rpr interval, rpr int, rpr i)?

# Conflicting options
$ rpr count --times 100 --for 1h --every 1s -- echo hello
Warning: --for 1h may stop execution before --times 100 is reached
```

## Configuration File Support

### Configuration File Format
```toml
# ~/.config/repeater/config.toml
[defaults]
continue_on_error = true
output_file = "/var/log/repeater.log"
timeout = "30s"

[interval]
jitter = "10%"
immediate = true

[count]  
parallel = 3

[daemon]
socket_path = "/var/run/repeater/daemon.sock"
```

### Environment Variables
```bash
RPR_CONTINUE_ON_ERROR=true
RPR_OUTPUT_FILE=/var/log/repeater.log
RPR_TIMEOUT=30s
RPR_DAEMON_SOCKET=/var/run/repeater/daemon.sock
```

## Abbreviation Reference

### Subcommands
| Full | Primary | Minimal |
|------|---------|---------|
| `interval` | `int` | `i` |
| `count` | `cnt` | `c` |
| `duration` | `dur` | `d` |
| `rate-limit` | `rate` | `r` |
| `schedule` | `sched` | `s` |
| `adaptive` | `adapt` | `a` |

### Common Options
| Full | Short | Example |
|------|-------|---------|
| `--every DURATION` | `-e DURATION` | `-e 30s` |
| `--times COUNT` | `-t COUNT` | `-t 100` |
| `--for DURATION` | `-f DURATION` | `-f 1h` |
| `--immediate` | `-n` | `-n` |
| `--quiet` | `-q` | `-q` |
| `--verbose` | `-v` | `-v` |
| `--continue-on-error` | `-k` | `-k` |
| `--jitter PERCENT` | `-j PERCENT` | `-j 20%` |
| `--parallel N` | `-p N` | `-p 5` |
| `--output-file FILE` | `-o FILE` | `-o log.txt` |
| `--timeout DURATION` | `-T DURATION` | `-T 30s` |