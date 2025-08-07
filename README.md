# Repeater (rpr) - Continuous Command Execution Tool

A Go-based CLI tool for continuous, scheduled, and rate-limited execution of commands with intelligent timing, rate limiting, and monitoring capabilities.

## Quick Start

```bash
# Build the project
make build

# Run tests
make test

# Install development tools
make install-tools

# See all available commands
make help
```

## Development Workflow

This project follows **Test-Driven Development (TDD)** methodology. See [AGENTS.md](AGENTS.md) for comprehensive development guidelines.

### TDD Quick Start

1. **Create a new behavior branch:**
   ```bash
   make tdd-behavior BEHAVIOR=scheduler-creation FEATURE=feature/scheduler-core
   ```

2. **Follow Red-Green-Refactor cycle:**
   - 🔴 **RED**: Write failing tests first
   - 🟢 **GREEN**: Implement minimal code to pass
   - 🔵 **REFACTOR**: Improve code while keeping tests green

3. **Use TDD helper for commits:**
   ```bash
   make tdd-helper
   ```

## Project Structure

```
├── cmd/rpr/              # Main application entry point
├── pkg/                  # Public packages
│   ├── scheduler/        # Scheduling algorithms
│   ├── executor/         # Command execution
│   └── config/           # Configuration management
├── internal/             # Private packages
│   └── daemon/           # Daemon integration
├── tests/                # Test suites
│   ├── integration/      # Integration tests
│   └── e2e/             # End-to-end tests
├── scripts/              # Development scripts
└── repeater-design/      # Design documentation
```

## Build Commands

```bash
# Build binary
make build

# Run all tests
make test                    # Unit tests
make test-integration        # Integration tests  
make test-e2e               # End-to-end tests
make benchmark              # Performance tests

# Quality checks
make quality-gate           # All quality checks
make lint                   # Run linter
make fmt                    # Format code
make coverage               # Generate coverage report

# Development
make install-tools          # Install dev tools
make tidy                   # Tidy dependencies
make clean                  # Clean build artifacts
```

## TDD Development

### Creating New Features

1. **Plan the feature** by breaking it into testable behaviors
2. **Create behavior branches** for each testable unit
3. **Follow TDD cycles** with commit proposals
4. **Merge completed behaviors** back to feature branch

### Example TDD Workflow

```bash
# 1. Create feature branch
git checkout -b feature/interval-scheduler

# 2. Create behavior branch
make tdd-behavior BEHAVIOR=scheduler-creation

# 3. Write failing test (RED)
# Edit pkg/scheduler/scheduler_test.go
go test -v ./pkg/scheduler/  # Should fail

# 4. Implement minimal code (GREEN)  
# Edit pkg/scheduler/scheduler.go
go test -v ./pkg/scheduler/  # Should pass

# 5. Refactor if needed (REFACTOR)
# Improve code while keeping tests green

# 6. Use commit helper
make tdd-helper
```

## Quality Standards

- **Test Coverage**: Minimum 85% for all packages
- **TDD Compliance**: All code must be test-driven
- **Code Quality**: Passes all linting and formatting checks
- **Performance**: Benchmarks must meet requirements

## Git Hooks

Pre-commit hooks automatically:
- Format code with `go fmt` and `goimports`
- Run linting with `golangci-lint`
- Execute all tests
- Validate TDD compliance
- **Never auto-commit** - always require manual approval

## Architecture

- **Scheduler Interface**: Pluggable scheduling algorithms
- **Executor Pattern**: Command execution with timeout and context support
- **Configuration**: TOML files with environment variable overrides
- **Daemon Integration**: Multi-instance coordination via patience daemon

## Contributing

1. Read [AGENTS.md](AGENTS.md) for development guidelines
2. Follow TDD methodology strictly
3. Use provided scripts and tools
4. Ensure all quality gates pass
5. Get commit proposals approved before execution

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.