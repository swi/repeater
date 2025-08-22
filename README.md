# Repeater (rpr) - A Command Execution Tool

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/swi/repeater)
[![Coverage](https://img.shields.io/badge/coverage-90%25-brightgreen)](https://github.com/swi/repeater)
[![Version](https://img.shields.io/badge/version-v0.5.1-blue)](https://github.com/swi/repeater/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

A CLI tool for continuous, scheduled command execution with intelligent timing, Unix pipeline integration, and advanced scheduling capabilities.

## Overview

Repeater executes commands repeatedly with helpful scheduling algorithms, making it good for monitoring, testing, data processing, and automation workflows. Unlike simpler tools, like `watch` or `cron`, repeater provides intelligent timing, rate limiting, and pattern matching.

## Key Features

- **8 Scheduling Modes**: interval, count, duration, cron, adaptive, backoff, load-aware, rate-limit
- **Pattern Matching**: Success/failure detection via regex patterns with precedence rules
- **HTTP-Aware Intelligence**: Automatic API response parsing for optimal scheduling
- **Plugin System**: Extensible architecture for custom schedulers and executors
- **Multi-level Abbreviations**: Power user shortcuts (`rpr i -e 30s -t 5 -- curl api.com`)
- **Unix Pipeline Integration**: Clean output, proper exit codes, real-time streaming

## Quick Start

### Installation
```bash
# Or install directly
go install github.com/swi/repeater/cmd/rpr@latest

# Build from source
git clone https://github.com/swi/repeater
cd repeater
go build -o rpr ./cmd/rpr
```

> 📚 **For detailed installation instructions and troubleshooting**, see [USAGE.md - Installation](USAGE.md#installation)

### Basic Usage
```bash
# Monitor API every 30 seconds for 10 times
rpr interval --every 30s --times 10 -- curl https://api.example.com/health

# Abbreviated form (same as above)
rpr i -e 30s -t 10 -- curl https://api.example.com/health

# Unix pipeline integration
rpr i -e 10s -t 5 -- curl -s https://api.com | jq -r '.status'

# Count successful responses
rpr i -e 5s -t 20 -- curl -s https://api.com | grep -c "success"
```

> 💡 **Want more examples?** See [USAGE.md](USAGE.md) for comprehensive CLI examples and real-world use cases

## Core Usage Examples

### Interval Execution
```bash
# Health monitoring with clean pipeline output
rpr interval --every 30s --for 1h -- curl -f https://api.example.com/health

# Abbreviated with multiple stop conditions
rpr i -e 30s -t 100 -f 1h -- ./health-check.sh
```

### Advanced Scheduling
```bash
# Adaptive scheduling based on response times
rpr adaptive --base-interval 1s --show-metrics -- curl https://api.example.com

# Cron-like scheduling with timezone support
rpr cron --cron "0 9 * * 1-5" --timezone "America/New_York" -- ./backup.sh

# Rate-limited API calls
rpr rate-limit --rate 100/1h -- curl https://api.github.com/user

# HTTP-aware intelligence (respects Retry-After headers)
rpr i -e 30s --http-aware -- curl -s https://api.example.com
```

> 🔧 **Advanced Configuration:** Learn about [HTTP-aware intelligence](USAGE.md#http-aware-intelligence), [pattern matching](USAGE.md#pattern-matching), and [configuration files](USAGE.md#configuration) in the Usage Guide

### Pattern Matching
```bash
# Success detection via output patterns
rpr i -e 30s -t 10 --success-pattern "healthy" -- ./service-check.sh

# Error detection with case-insensitive matching
rpr i -e 1m --failure-pattern "(?i)error|timeout" --case-insensitive -- ./monitor.sh
```

### Output Control
```bash
# Quiet mode (no command output)
rpr i -e 30s -t 5 --quiet -- curl -f https://api.com

# Verbose mode (detailed execution info)
rpr i -e 10s -t 3 --verbose -- ./debug-script.sh

# Stats-only mode (metrics without output)
rpr i -e 5s -t 10 --stats-only -- ./performance-test.sh
```

## Real-World Use Cases

- **DevOps Monitoring**: API health checks, service monitoring, uptime tracking
- **CI/CD Pipelines**: Build monitoring, deployment verification, test automation
- **Data Processing**: ETL pipelines, log analysis, metrics collection
- **Load Testing**: Sustained traffic generation, performance monitoring
- **System Administration**: Maintenance tasks, resource monitoring, cleanup jobs

## Documentation

### User Guides
- 📖 **[Usage Guide](USAGE.md)** - Complete CLI reference, examples, and real-world use cases
- ⚙️ **[Configuration Guide](USAGE.md#configuration)** - TOML files, environment variables, and advanced setup

### Technical Documentation  
- 🏗️ **[Architecture](ARCHITECTURE.md)** - System design, components, and performance characteristics
- 📋 **[Feature Roadmap](FEATURES.md)** - Implementation status and future enhancements

### Development
- 🤝 **[Contributing Guide](CONTRIBUTING.md)** - TDD workflow, code standards, and plugin development
- 📝 **[Changelog](CHANGELOG.md)** - Version history and migration guides

### Quick Links
- 🚀 [Basic Usage Examples](USAGE.md#core-usage-examples)
- 🧠 [Advanced Scheduling](USAGE.md#advanced-scheduling) 
- 🔌 [Plugin Development](CONTRIBUTING.md#plugin-development)
- 🐛 [Troubleshooting](USAGE.md#troubleshooting)

## Status:

**Current Version**: v0.5.1

### Fully Implemented & Tested
- ✅ **Complete CLI** with multi-level abbreviations and intuitive UX
- ✅ **8 Scheduler Types** including advanced adaptive and load-aware scheduling
- ✅ **HTTP-Aware Intelligence** with automatic API response parsing
- ✅ **Plugin System** for extensible custom schedulers and executors
- ✅ **Pattern Matching** with regex success/failure detection and precedence
- ✅ **Unix Pipeline Integration** with clean output and proper exit codes
- ✅ **Production Features** (metrics, health endpoints, signal handling, recovery)
- ✅ **Decent Testing** (210+ tests, 90%+ coverage, benchmarks, race testing)

### Quality Metrics
- **Test Coverage**: 90%+ across all packages
- **Performance**: <1% timing deviation, minimal resource usage
- **Reliability**: Graceful error handling, proper signal management
- **Usability**: Intuitive CLI with comprehensive help and clear error messages

## Integration Examples

### Monitoring Systems
```bash
# Prometheus metrics collection
rpr i -e 30s --enable-metrics --metrics-port 8080 -- ./collect-metrics.sh

# ELK Stack integration
rpr i -e 5m -- ./log-analysis.sh | jq . | curl -X POST "elasticsearch:9200/logs/_doc" -d @-
```

### CI/CD Pipelines
```bash
# GitHub Actions health check
rpr i -e 10s -t 30 --quiet -- curl -f $DEPLOYMENT_URL/health

# Kubernetes deployment verification
rpr adaptive --base-interval 30s --min-interval 10s -- kubectl get pods -l app=myapp
```

### Unix Pipeline Workflows
```bash
# Data processing pipeline
rpr i -e 1m -- curl -s https://api.com/data | jq -r '.items[]' | sort | uniq -c

# System monitoring with alerting
rpr i -e 30s -- df -h / | awk '{print $5}' | sed 's/%//' | awk '$1>80{exit 1}' || alert.sh
```

## Exit Codes

Repeater follows Unix conventions for scripting integration:
- **0**: All commands executed successfully
- **1**: Some commands failed during execution
- **2**: Usage error (invalid arguments, configuration issues)
- **130**: Interrupted by user (Ctrl+C, SIGINT, SIGTERM)

## Performance

- **Timing Accuracy**: <1% deviation from specified intervals
- **Resource Efficient**: Minimal memory footprint and CPU usage
- **Concurrent Safe**: Thread-safe execution with proper cleanup
- **Signal Responsive**: <100ms shutdown time on interruption

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
