# Contributing to Repeater

Thank you for your interest in contributing to Repeater! This document provides guidelines for contributing to the project.

## 🎉 Project Status: **ADVANCED FEATURES COMPLETE (v0.3.0)**

**Repeater is now a feature-complete platform** with advanced scheduling, plugin system, and comprehensive observability! The project includes all core features plus extensible architecture for custom functionality.

### ✅ **What's Working**
- Complete CLI with multi-level abbreviations (`rpr i -e 30s -t 5 -- curl api.com`)
- Multiple execution modes: interval, count, duration, cron, adaptive, backoff, load-aware, rate-limit
- Plugin system with extensible architecture for custom schedulers and executors
- Configuration files with TOML support and environment variable overrides
- Health endpoints and Prometheus-compatible metrics export
- Signal handling and graceful shutdown
- Comprehensive statistics and reporting
- 85+ tests with 90%+ coverage

## Development Workflow

This project follows **Test-Driven Development (TDD)** methodology. Please read [AGENTS.md](AGENTS.md) for comprehensive development guidelines.

### Quick Start

1. **Fork and clone the repository**
2. **Install development tools**: `make install-tools`
3. **Create a feature branch**: `git checkout -b feature/your-feature`
4. **Follow TDD workflow**: Red-Green-Refactor cycles
5. **Use commit proposals**: All commits require manual approval

### TDD Requirements

- **Write tests first** - No implementation without failing tests
- **Follow Red-Green-Refactor** - Complete cycles for each behavior
- **Use behavior branches** - `tdd/specific-behavior` for individual features
- **Maintain coverage** - Minimum 85% test coverage required
- **Quality gates** - All linting and formatting must pass

### Branch Strategy

```
main
├── feature/component-name
│   ├── tdd/behavior-1
│   ├── tdd/behavior-2
│   └── tdd/behavior-3
```

### Commit Guidelines

- Follow conventional commits: `type(scope): description`
- Include TDD metadata in commit messages
- All commits must be approved before execution
- Use provided scripts: `make tdd-helper`

### Code Standards

- **Go formatting**: `go fmt` and `goimports`
- **Linting**: `golangci-lint` must pass
- **Testing**: `go test -race ./...` must pass
- **Documentation**: All public APIs must be documented

### Pull Request Process

1. **Complete TDD cycles** for all behaviors
2. **Ensure quality gates pass**: `make quality-gate`
3. **Update documentation** if needed
4. **Add changelog entry** if applicable
5. **Request review** from maintainers

### Development Commands

```bash
# Build and test
make build
make test
make quality-gate

# TDD workflow
make tdd-behavior BEHAVIOR=behavior-name
make tdd-helper

# Quality checks
make lint
make coverage

# Plugin development
make plugin-example
make plugin-test
```

## Plugin Development

### Creating Custom Plugins

Repeater supports extensible plugins for schedulers, executors, and outputs. Here's how to develop plugins:

#### Plugin Interface Implementation

```go
// Example scheduler plugin
package main

import (
    "time"
    "github.com/swi/repeater/pkg/plugin"
    "github.com/swi/repeater/pkg/scheduler"
)

type MySchedulerPlugin struct{}

func (p *MySchedulerPlugin) Name() string { return "my-scheduler" }
func (p *MySchedulerPlugin) Version() string { return "1.0.0" }
func (p *MySchedulerPlugin) Description() string { 
    return "Custom scheduling algorithm" 
}

func (p *MySchedulerPlugin) NewScheduler(config map[string]interface{}) (scheduler.Scheduler, error) {
    // Create and return your custom scheduler
    return NewMyScheduler(config), nil
}

func (p *MySchedulerPlugin) ValidateConfig(config map[string]interface{}) error {
    // Validate plugin configuration
    return nil
}

func (p *MySchedulerPlugin) ConfigSchema() *plugin.ConfigSchema {
    // Return configuration schema
    return &plugin.ConfigSchema{
        Fields: []plugin.ConfigField{
            {
                Name:        "interval",
                Type:        "duration",
                Required:    true,
                Description: "Base interval for scheduling",
            },
        },
    }
}

// Plugin entry point
var Plugin MySchedulerPlugin
```

#### Plugin Development Guidelines

1. **Follow Interface Contracts**: Implement all required methods
2. **Validate Configuration**: Provide comprehensive config validation
3. **Handle Errors Gracefully**: Return meaningful error messages
4. **Document Configuration**: Provide clear config schema
5. **Test Thoroughly**: Include unit and integration tests

#### Plugin Testing

```bash
# Test plugin loading
go test ./pkg/plugin/

# Test plugin functionality
go test ./examples/plugins/my-scheduler/

# Integration testing
rpr plugin my-scheduler --interval 1s -- echo "test"
```

#### Plugin Distribution

- **Go Plugins**: Compile to `.so` files for dynamic loading
- **External Processes**: Implement plugin protocol for any language
- **Configuration**: Provide `plugin.toml` manifest file

### Getting Help

- **Read AGENTS.md** for detailed TDD workflow
- **Check existing issues** for similar problems
- **Ask questions** in GitHub discussions
- **Follow project conventions** established in codebase

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain professional communication

## License

By contributing, you agree that your contributions will be licensed under the MIT License.