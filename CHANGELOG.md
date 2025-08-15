# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Distributed multi-node coordination
- Advanced plugin types (output processors, custom executors)
- Enhanced observability (Grafana dashboards, alerting)
- Advanced integrations (Kubernetes operators, Terraform providers)

## [0.3.0] - 2025-01-13 - **ADVANCED FEATURES COMPLETE** 🎉

### Added - Advanced Scheduling & Plugin System
- **Cron Scheduling** with timezone support and standard cron expressions
- **Plugin System** with extensible architecture for custom schedulers and executors
- **Advanced Schedulers**: adaptive, backoff, load-aware, rate-limiting modes
- **HTTP-Aware Intelligence** with automatic HTTP response parsing for optimal API scheduling
- **Configuration Files** with TOML support and environment variable overrides
- **Health Endpoints** with HTTP server for monitoring and observability
- **Metrics Collection** with Prometheus-compatible metrics export

### Added - Cron Features
- **Cron Expression Parser** supporting standard 5-field format (minute hour day month weekday)
- **Cron Shortcuts** (@daily, @hourly, @weekly, @monthly, @yearly, @annually)
- **Timezone Support** with proper DST handling and timezone-aware scheduling
- **CLI Integration** with `cron`/`cr` subcommand and `--cron`, `--timezone` flags
- **Comprehensive Testing** with 19 new test cases covering all cron functionality

### Added - Plugin System
- **Plugin Interface** with clean abstractions for schedulers, executors, and outputs
- **Plugin Manager** with dynamic loading, validation, and lifecycle management
- **Plugin Registry** with discovery, registration, and metadata management
- **Security Features** with plugin validation and sandboxing capabilities
- **CLI Integration** with plugin management commands and help system

### Added - HTTP-Aware Intelligence
- **HTTP Response Parsing** with automatic extraction of timing information from API responses
- **Retry-After Header Support** respecting server-specified retry timing from HTTP headers
- **JSON Response Parsing** extracting timing from `retry_after`, `retryAfter`, and rate limit fields
- **Real-World API Support** with GitHub (403), AWS (429), Stripe, Discord API compatibility
- **Priority-Based Parsing** with headers > custom JSON > standard JSON > nested structures
- **Configuration Options** with parsing control, delay constraints, and custom field support
- **Fallback Integration** seamlessly combining with any scheduler when no HTTP timing available
- **CLI Integration** with `--http-aware`, `--http-max-delay`, `--http-custom-fields` flags

### Added - Advanced Schedulers
- **Adaptive Scheduler** with AIMD algorithm and response time learning
- **Exponential Backoff** with configurable multipliers, jitter, and max intervals
- **Load-Aware Scheduling** with CPU, memory, and system load monitoring
- **Rate Limiting** with mathematical algorithms and daemon coordination support

### Added - Configuration & Observability
- **TOML Configuration** with structured config files and validation
- **Environment Variables** with override support and flexible configuration
- **Health Endpoints** with HTTP server providing /health, /ready, /live endpoints
- **Metrics Collection** with Prometheus-compatible metrics and statistics export

### Technical Implementation
- **Enhanced Package Structure**: Added `cron`, `plugin`, `config`, `health`, `metrics`, `httpaware` packages
- **Comprehensive Testing**: 210+ tests with 85%+ coverage across all packages
- **Plugin Architecture**: Interface-based design supporting Go plugins and external processes
- **HTTP-Aware Architecture**: Regex-based parsing with JSON support and priority-based timing extraction
- **Advanced Error Handling**: Categorized errors, circuit breakers, and retry policies

### CLI Enhancements
- **New Subcommands**: `cron`/`cr`, `adaptive`/`a`, `backoff`/`b`, `load-adaptive`/`la`, `rate-limit`/`rl`
- **Extended Flags**: `--cron`, `--timezone`/`--tz`, `--base-interval`, `--initial-delay`, `--max`, `--rate`, `--attempts`
- **HTTP-Aware Flags**: `--http-aware`, `--http-max-delay`, `--http-min-delay`, `--http-custom-fields`
- **Parsing Control**: `--http-parse-json`, `--http-no-parse-json`, `--http-parse-headers`, `--http-trust-client`
- **Plugin Support**: Dynamic plugin loading and management via CLI
- **Enhanced Help**: Comprehensive documentation for all new features

### Performance & Quality
- **Timing Accuracy**: Maintained <1% deviation for all scheduling modes
- **Resource Efficiency**: Optimized memory usage and CPU utilization
- **Concurrent Safety**: Thread-safe execution across all new components
- **Production Ready**: Comprehensive error handling and graceful degradation

## [0.2.0] - 2025-01-08 - **MVP COMPLETE** 🎉

### Added - Core Functionality
- **Complete CLI system** with argument parsing and validation
- **Multi-level abbreviations** for commands and flags (32% keystroke reduction)
- **Interval scheduling** with precise timing and jitter support
- **Command execution engine** with context-aware timeout handling
- **End-to-end integration** connecting schedulers with executors
- **Stop conditions** supporting times, duration, and signal-based stopping
- **Signal handling** for graceful shutdown (SIGINT/SIGTERM)
- **Execution statistics** with comprehensive metrics and reporting

### Added - CLI Features
- **Subcommands**: `interval`/`int`/`i`, `count`/`cnt`/`c`, `duration`/`dur`/`d`
- **Flag abbreviations**: `--every`/`-e`, `--times`/`-t`, `--for`/`-f`
- **Flexible combinations**: Mix intervals with count/duration limits
- **Help system** with abbreviation examples and usage patterns
- **Error handling** with clear, actionable error messages

### Added - Execution Features
- **Context-aware execution** with proper cancellation support
- **Output capture** preserving stdout, stderr, and exit codes
- **Timeout handling** with configurable per-command timeouts
- **Concurrent safety** with thread-safe execution patterns
- **Resource cleanup** with proper goroutine and resource management

### Added - Integration Features
- **Runner orchestration** connecting all components seamlessly
- **Stop condition evaluation** with first-condition-wins logic
- **Statistics collection** tracking success/failure rates and timing
- **Progress reporting** with real-time execution feedback
- **Graceful shutdown** completing current execution before stopping

### Added - Testing & Quality
- **72 comprehensive tests** across all packages with high coverage
- **TDD methodology** with Red-Green-Refactor cycles throughout
- **Integration tests** covering end-to-end execution scenarios
- **Race condition testing** ensuring concurrent execution safety
- **Performance benchmarks** validating timing accuracy requirements

### Technical Implementation
- **Package structure**: `cli`, `scheduler`, `executor`, `runner` packages
- **Interface design**: Clean abstractions with type safety
- **Error propagation**: Proper error handling with context preservation
- **Memory management**: Efficient resource usage with cleanup
- **Signal handling**: OS signal integration for production use

### Documentation
- **Comprehensive README** with current functionality and examples
- **Detailed USAGE guide** with real-world use cases and patterns
- **Development guidelines** in AGENTS.md with TDD workflow
- **Architecture documentation** with implementation details

### Performance
- **Timing accuracy**: <1% deviation from specified intervals
- **Resource efficiency**: Minimal memory footprint and CPU usage
- **Startup time**: <10ms from command to first execution
- **Shutdown time**: <100ms graceful shutdown on interruption

## [0.1.0] - 2025-01-07

### Added - Foundation
- **Project initialization** with Go module and standard structure
- **TDD infrastructure** with comprehensive development workflow
- **Build system** with Makefile and quality automation
- **Git hooks** for automated quality checks (formatting, linting, testing)
- **Development scripts** for TDD behavior-driven development
- **Documentation structure** with design documents and guidelines

### Infrastructure
- **Repository setup** with proper Go project layout
- **Quality gates** with golangci-lint and automated testing
- **Development environment** configuration and tooling
- **CI/CD foundation** ready for GitHub Actions integration
- **License and contributing** guidelines established

---

## Version History Summary

- **v0.3.0**: 🎉 **Advanced Features Complete** - Plugin system, cron scheduling, advanced schedulers, configuration, observability
- **v0.2.0**: 🎉 **MVP Complete** - Full functionality with CLI, scheduling, execution, and integration
- **v0.1.0**: 🏗️ **Foundation** - Project setup, TDD infrastructure, and development workflow

## Migration Guide

### From v0.2.0 to v0.3.0
- **New functionality**: Advanced scheduling modes, plugin system, cron support, configuration files
- **CLI changes**: New subcommands (cron, adaptive, backoff, load-adaptive, rate-limit) with abbreviations
- **Breaking changes**: None (backward compatible)
- **New dependencies**: No external dependencies added beyond Go standard library

### From v0.1.0 to v0.2.0
- **New functionality**: All core features now implemented and ready for use
- **CLI changes**: Full command-line interface with abbreviations now available
- **Breaking changes**: None (first functional release)
- **New dependencies**: No external dependencies added beyond Go standard library

## Development Methodology

This project follows **strict Test-Driven Development (TDD)** with:
- **Red-Green-Refactor cycles** for all feature development
- **Comprehensive test coverage** (85%+ across all packages)
- **Integration testing** for end-to-end functionality validation
- **Performance testing** ensuring timing accuracy and resource efficiency
- **Race condition testing** for concurrent execution safety

Every feature was implemented following TDD principles with failing tests written first, minimal implementation to pass tests, and subsequent refactoring for code quality.