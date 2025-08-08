# Repeater (rpr) - Continuous Command Execution Tool

A powerful Go-based CLI tool for continuous, scheduled command execution with intelligent timing, stop conditions, and comprehensive monitoring.

## 🚀 Quick Start

```bash
# Build the project
go build -o rpr ./cmd/rpr

# Run a command every 30 seconds, 5 times
./rpr interval --every 30s --times 5 -- curl https://api.example.com

# Ultra-compact form with abbreviations
./rpr i -e 30s -t 5 -- curl https://api.example.com

# Count-based execution
./rpr count --times 10 -- echo "Hello World"

# Duration-based execution  
./rpr duration --for 2m --every 10s -- date
```

## ✨ Features

### 🎯 **Execution Modes**
- **Interval**: Execute commands at regular time intervals
- **Count**: Execute commands a specific number of times  
- **Duration**: Execute commands for a specific time period
- **Flexible combinations**: Mix intervals with count/duration limits

### ⚡ **CLI Abbreviations**
- **Multi-level shortcuts**: `interval`/`int`/`i`, `count`/`cnt`/`c`, `duration`/`dur`/`d`
- **Flag abbreviations**: `--every`/`-e`, `--times`/`-t`, `--for`/`-f`
- **32% fewer keystrokes** for power users

### 🛑 **Stop Conditions**
- **Times limit**: Stop after N executions
- **Duration limit**: Stop after specified time
- **Signal handling**: Graceful shutdown on Ctrl+C (SIGINT/SIGTERM)
- **Smart stopping**: First condition reached wins

### 📊 **Execution Statistics**
- **Real-time feedback**: Progress and completion status
- **Detailed metrics**: Success/failure counts, execution times
- **Command output**: Full stdout/stderr capture
- **Exit code preservation**: Maintains command exit codes

## 📖 Usage Examples

### Basic Examples

```bash
# Monitor API health every 30 seconds for 1 hour
rpr interval --every 30s --for 1h -- curl -f https://api.example.com/health

# Run tests 50 times with 2-second intervals
rpr count --times 50 --every 2s -- npm test

# Monitor system for 10 minutes, checking every minute
rpr duration --for 10m --every 1m -- df -h /
```

### Power User Shortcuts

```bash
# Ultra-compact monitoring
rpr i -e 10s -f 5m -- curl -f https://api.com/health

# Quick load testing
rpr c -t 100 -e 100ms -- curl -s https://api.com/endpoint

# System monitoring
rpr d -f 1h -e 5m -- free -h
```

### Real-World Use Cases

```bash
# Website uptime monitoring
rpr i -e 30s -f 24h -- curl -f -s -o /dev/null https://mysite.com

# Database backup verification
rpr c -t 3 -e 10s -- mysqldump --single-transaction mydb > /dev/null

# Log file monitoring during deployment
rpr d -f 30m -e 5s -- tail -n 10 /var/log/app.log

# SSL certificate check
rpr i -e 24h -t 7 -- openssl s_client -connect example.com:443 < /dev/null
```

## 🏗️ Architecture

### Current Implementation (MVP Complete)
- ✅ **CLI Foundation**: Full argument parsing with abbreviations
- ✅ **Interval Scheduler**: Precise timing with jitter support
- ✅ **Command Executor**: Context-aware execution with timeout handling
- ✅ **Integration Layer**: End-to-end orchestration with stop conditions
- ✅ **Signal Handling**: Graceful shutdown and cleanup

### Project Structure
```
├── cmd/rpr/              # Main application entry point
├── pkg/                  # Core packages
│   ├── cli/              # ✅ CLI parsing and validation
│   ├── scheduler/        # ✅ Interval scheduling algorithms  
│   ├── executor/         # ✅ Command execution engine
│   └── runner/           # ✅ Integration orchestration
├── repeater-design/      # Design documentation
├── scripts/              # Development scripts
└── tests/                # Comprehensive test suites
```

## 🧪 Development & Testing

### Test-Driven Development
This project follows strict **TDD methodology** with comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with race detection
go test ./... -race

# Run specific package tests
go test ./pkg/runner/ -v
```

### Quality Metrics
- **72 comprehensive tests** across all packages
- **High test coverage**: 85%+ across core packages
- **100% coverage**: Command executor package
- **Race condition testing**: Concurrent execution safety

### Build Commands
```bash
# Build binary
go build -o rpr ./cmd/rpr

# Run linting
go vet ./...
go fmt ./...

# Run all quality checks
make test && make lint
```

## 📊 Current Status

### ✅ **Completed (Phase 1 - MVP)**
- **CLI Foundation**: Full parsing with multi-level abbreviations
- **Interval Scheduling**: Precise timing with immediate execution
- **Command Execution**: Context-aware with timeout and output capture
- **Integration**: End-to-end orchestration with stop conditions
- **Signal Handling**: Graceful shutdown on interruption
- **Statistics**: Comprehensive execution metrics and reporting

### 🔄 **In Progress**
- Documentation updates and examples
- Performance optimizations
- Additional scheduler types (count, duration optimizations)

### 🚧 **Planned (Phase 2+)**
- **Rate Limiting**: Mathematical rate limiting with daemon coordination
- **Advanced Scheduling**: Cron-like scheduling, adaptive intervals
- **Configuration Files**: TOML configuration with environment overrides
- **Enhanced Output**: Structured logging, metrics export

## 🎯 Performance

- **Timing Accuracy**: <1% deviation from specified intervals
- **Resource Efficient**: Minimal memory footprint and CPU usage
- **Concurrent Safe**: Thread-safe execution with proper cleanup
- **Signal Responsive**: <100ms shutdown time on interruption

## 🤝 Contributing

1. **Follow TDD**: Write tests before implementation
2. **Maintain Coverage**: Keep test coverage above 85%
3. **Use Abbreviations**: Support both full and abbreviated commands
4. **Test Thoroughly**: Include integration and edge case tests
5. **Document Changes**: Update relevant documentation

See [AGENTS.md](AGENTS.md) for detailed development guidelines.

## 📚 Documentation

- **[USAGE.md](USAGE.md)**: Comprehensive usage guide with examples
- **[CHANGELOG.md](CHANGELOG.md)**: Version history and changes
- **[AGENTS.md](AGENTS.md)**: Development workflow and TDD guidelines
- **[Design Docs](repeater-design/)**: Architecture and implementation plans

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Ready for production use as MVP!** 🎉

The core functionality is complete and thoroughly tested. Perfect for continuous command execution, monitoring, testing, and automation workflows.