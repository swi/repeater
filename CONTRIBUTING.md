# Contributing to Repeater

Thank you for your interest in contributing to Repeater! This document provides comprehensive guidelines for contributing to the project, including our Test-Driven Development (TDD) methodology and documentation standards.

## 🎉 Project Status: **PRODUCTION READY (v0.5.1)**

**Repeater is now a production-ready platform** with excellent code quality, comprehensive testing, and complete feature set! The project has achieved A- grade (91.7/100) quality metrics with all core features plus extensible architecture for custom functionality.

### ✅ **What's Working**
- Complete CLI with multi-level abbreviations (`rpr i -e 30s -t 5 -- curl api.com`)
- Multiple execution modes: interval, count, duration, cron, adaptive, backoff, load-aware, rate-limit
- Plugin system with extensible architecture for custom schedulers and executors
- HTTP-aware intelligence for automatic API response parsing
- Pattern matching for success/failure detection with precedence rules
- Configuration files with TOML support and environment variable overrides
- Health endpoints and Prometheus-compatible metrics export
- Signal handling and graceful shutdown
- Comprehensive statistics and reporting
- 210+ tests with 90%+ coverage

## Development Workflow

This project follows **Test-Driven Development (TDD)** methodology. **NEVER write implementation code without tests first.**

### Quick Start

1. **Fork and clone the repository**
2. **Install development tools**: `make install-tools`
3. **Create a feature branch**: `git checkout -b feature/your-feature`
4. **Follow TDD workflow**: Red-Green-Refactor cycles
5. **Use commit proposals**: All commits require manual approval

## Development Environment

### Required Tools

**Core Requirements:**
- **Go 1.22+**: Language runtime
- **golangci-lint v2.x**: Code linting (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- **goimports**: Import management (`go install golang.org/x/tools/cmd/goimports@latest`)

**Installation:**
```bash
# Install all development tools
make install-tools

# Manual installation if needed
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Tool Verification:**
```bash
# Check if tools are properly installed
golangci-lint version  # Should show v2.x
goimports -version     # Should show help text
go version            # Should show 1.22+
```

**Note**: If `goimports` shows "command not found", ensure `$(go env GOPATH)/bin` is in your PATH:
```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Quality Gates

**Pre-commit checks automatically run:**
- Code formatting via `go fmt` and `goimports`
- Linting via `golangci-lint run` (must pass with 0 issues)
- Full test suite via `go test ./...` (all tests must pass)
- TDD compliance validation

### TDD Requirements (MANDATORY)

**NEVER write implementation code without tests first. Follow this exact sequence:**

#### Red-Green-Refactor Cycle
1. **RED**: Write a failing test that describes the desired behavior
2. **GREEN**: Write minimal code to make the test pass
3. **REFACTOR**: Improve code while keeping tests green

#### Before Writing Any Code
- [ ] Create test file first (`*_test.go`)
- [ ] Write failing test with clear assertions
- [ ] Run test to confirm it fails: `go test -v ./pkg/...`
- [ ] Only then implement the minimal code to pass

#### TDD Task Scoping
**Break every task into TDD-sized chunks (1-3 test cases per commit)**
- **Micro-task**: 1 test case, 1 commit (15-30 min)
- **Small task**: 2-3 related test cases, 1-2 commits (1-2 hours)
- **Medium task**: Multiple behaviors, 3-5 commits (half day)
- **Large task**: Break down further - never start without breakdown

### LLM Commit Proposal Process

**After completing any TDD phase, LLM must propose commit for user approval:**

#### Commit Proposal Format
```
🔄 TDD Phase Complete - Commit Proposal
=====================================

📊 Changes Summary:
- Files modified: [list]
- TDD Phase: [RED/GREEN/REFACTOR]
- Tests status: [before] → [after]
- Behavior: [specific behavior implemented]

📝 Proposed Commit:
```bash
git add [files]
git commit -m "[type](phase): [description]

- [specific change 1]
- [specific change 2]
- [TDD context and next steps]

TDD-Phase: [RED/GREEN/REFACTOR]
Behavior: [behavior-name]
Tests-Added/Modified: [number]
Coverage-Change: [before]% -> [after]%"
```

❓ Approve this commit? (y/n/e/d)
```

#### User Approval Options
- **y/yes**: Execute the proposed commit exactly as shown
- **n/no**: Cancel commit, allow manual review/modification
- **e/edit**: Modify the commit message before executing
- **d/diff**: Show detailed changes before deciding

### Branch Strategy

```
main (production-ready, all tests pass)
├── develop (integration branch, all TDD cycles complete)
├── feature/component-name
│   ├── tdd/behavior-1
│   ├── tdd/behavior-2
│   └── tdd/behavior-3
```

#### Branch Naming Convention
- **Feature branches**: `feature/<component>-<capability>`
- **TDD behavior branches**: `tdd/<specific-behavior>`
- **Bug fix branches**: `fix/<issue-description>`

#### TDD Behavior Branch Workflow
```bash
# Create behavior-specific branch
git checkout feature/scheduler-core
git checkout -b tdd/scheduler-creation

# Complete TDD cycles on behavior branch
# RED → GREEN → REFACTOR (with LLM commit proposals)

# Merge completed behavior back
git checkout feature/scheduler-core
git merge --no-ff tdd/scheduler-creation
git branch -d tdd/scheduler-creation
```

### Build and Test Commands

```bash
# Build the binary
go build -o rpr ./cmd/rpr

# Run all tests
make test                    # Unit tests
make test-integration        # Integration tests  
make test-e2e               # End-to-end tests
make benchmark              # Performance tests
make quality-gate           # All quality checks

# Run single test
go test -v ./pkg/scheduler/  # Test specific package
go test -run TestSpecificFunction ./pkg/...  # Test specific function

# Linting and formatting
make lint                   # Run golangci-lint
go fmt ./...               # Format code
```

### TDD Quality Gates

#### Coverage Requirements
- Minimum 90% test coverage: `go test -cover ./...`
- No untested public functions
- All error paths must be tested

#### Test Quality Checks
```bash
# Verify tests fail without implementation
go test -v ./pkg/... # Should show failures initially

# Check coverage
go test -cover ./... | grep -E "coverage: [0-9]+\.[0-9]+%" 

# Race condition detection
go test -race ./...

# Benchmark performance
go test -bench=. ./...
```

#### TDD Anti-Patterns (FORBIDDEN)
- ❌ Writing implementation before tests
- ❌ Writing tests after implementation ("test-after")
- ❌ Skipping the "failing test" step
- ❌ Not running tests frequently during development
- ❌ Ignoring test failures or making tests pass by changing assertions

### Code Standards

#### Go Code Style
- **Go formatting**: `go fmt` and `goimports`
- **Linting**: `golangci-lint` must pass
- **Testing**: `go test -race ./...` must pass
- **Documentation**: All public APIs must be documented

#### Naming Conventions
- Use Go standard naming: PascalCase for exported, camelCase for unexported
- Interface names end with -er suffix (e.g., `Scheduler`, `Executor`)
- Package names are lowercase, single word when possible

#### Error Handling
- Return errors as the last return value
- Use `fmt.Errorf()` for error wrapping with context
- Handle all errors explicitly - no silent failures
- Use custom error types for domain-specific errors

#### Imports
- Group imports: standard library, third-party, local packages
- Use goimports for automatic formatting
- Avoid dot imports except in tests

#### Types and Interfaces
- Prefer small, focused interfaces
- Use context.Context for cancellation and timeouts
- Implement String() method for custom types used in logging

#### Concurrency
- Use channels for communication between goroutines
- Always handle context cancellation in long-running operations
- Protect shared state with mutexes or atomic operations
- Test for race conditions with `go test -race`

### Key Architecture Patterns
- **Scheduler Interface**: Pluggable scheduling algorithms (interval, count, duration, cron, adaptive, backoff, load-aware, rate-limited)
- **Plugin System**: Extensible architecture with interface-based plugins for schedulers, executors, and outputs
- **Executor Pattern**: Command execution with timeout, output capture, and context support
- **Configuration**: TOML files with environment variable overrides
- **Pattern Matching**: Regex-based success/failure detection with precedence rules

> 🏗️ **Deep Dive:** For detailed technical implementation of these patterns, see [Architecture Guide](ARCHITECTURE.md). For usage examples, see [Usage Guide](USAGE.md).

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

func (p *MySchedulerPlugin) NewScheduler(config map[string]interface{}) (Scheduler, error) {
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

### Pull Request Process

1. **Complete TDD cycles** for all behaviors
2. **Ensure quality gates pass**: `make quality-gate`
3. **Update documentation** if needed (see Documentation Standards below)
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

## Documentation Standards (MANDATORY)

**All documentation changes must follow these standards and be kept in sync.**

### Core Documentation Files

#### **README.md** - Project Landing Page (150-200 lines)
- **Purpose**: Concise project overview with quick start
- **Content**: Description, key features, basic examples, links to detailed docs
- **Updates Required**: When adding major features, changing core functionality, or updating status
- **Maintainer**: All contributors must keep this current

#### **USAGE.md** - Comprehensive User Guide (300-400 lines)  
- **Purpose**: Complete CLI reference and usage examples
- **Content**: Installation, all commands, configuration, integration patterns, troubleshooting
- **Updates Required**: When adding new commands, flags, or usage patterns
- **Testing**: All examples must be tested and functional

#### **ARCHITECTURE.md** - Technical Design
- **Purpose**: Technical architecture and implementation details
- **Content**: System design, component overview, data flow, performance characteristics
- **Updates Required**: When changing architecture, adding new components, or major refactoring
- **Audience**: Developers and contributors

#### **CONTRIBUTING.md** - This File
- **Purpose**: Development guidelines, TDD workflow, and documentation standards
- **Content**: TDD methodology, code standards, plugin development, documentation rules
- **Updates Required**: When changing development processes or adding new contribution types

#### **FEATURES.md** - Roadmap and Planning
- **Purpose**: Feature roadmap and implementation status
- **Content**: Implemented features, planned features, timelines, priorities
- **Updates Required**: When completing features or changing roadmap

#### **CHANGELOG.md** - Version History
- **Purpose**: Detailed version history and release notes
- **Content**: Version history, breaking changes, new features, bug fixes
- **Updates Required**: With every release, following semantic versioning

### Documentation Update Rules

#### **MANDATORY: Keep Documentation in Sync**

1. **When adding new features**:
   - [ ] Update README.md with feature summary
   - [ ] Add comprehensive examples to USAGE.md
   - [ ] Update ARCHITECTURE.md if design changes
   - [ ] Mark feature complete in FEATURES.md
   - [ ] Add changelog entry

2. **When modifying CLI**:
   - [ ] Update all command examples in USAGE.md
   - [ ] Test all examples for accuracy
   - [ ] Update help text and documentation
   - [ ] Update README.md quick start if needed

3. **When changing architecture**:
   - [ ] Update ARCHITECTURE.md with new design
   - [ ] Update component descriptions
   - [ ] Update data flow diagrams
   - [ ] Update performance characteristics

4. **When adding documentation**:
   - [ ] Follow established structure and style
   - [ ] Include working examples
   - [ ] Test all code examples
   - [ ] Link from appropriate locations

#### **Documentation Quality Standards**

1. **Accuracy**: All examples must be tested and functional
2. **Completeness**: Cover all features and use cases
3. **Clarity**: Write for the intended audience (users vs developers)
4. **Consistency**: Use consistent terminology and formatting
5. **Currency**: Keep information up-to-date with current implementation

#### **Prohibited Actions**

- ❌ **No duplicate content** across files
- ❌ **No outdated examples** or references
- ❌ **No conflicting information** between files
- ❌ **No template artifacts** or placeholder content
- ❌ **No version inconsistencies** across documentation

#### **Documentation Review Checklist**

Before submitting any changes:

- [ ] All affected documentation files updated
- [ ] All code examples tested and functional
- [ ] Version numbers consistent across all files
- [ ] No duplicate or conflicting information
- [ ] Links between documents work correctly
- [ ] Appropriate audience for content level
- [ ] Clear and accurate descriptions
- [ ] Examples follow current CLI patterns

### Automated Documentation Checks

```bash
# Check documentation consistency
make docs-check

# Test all documentation examples
make docs-test

# Validate links and references
make docs-validate

# Check for duplicated content
make docs-duplicate-check
```

## Git Hooks and Quality Automation

### Pre-commit Hook (Quality Checks Only - No Auto-commits)
```bash
#!/bin/bash
# .git/hooks/pre-commit
echo "🔍 Running automated quality checks..."

# Auto-format code (modify files, don't commit)
go fmt ./...
goimports -w .

# Run linting
if ! golangci-lint run; then
    echo "❌ Linting failed - fix issues before committing"
    exit 1
fi

# Run tests
if ! go test ./...; then
    echo "❌ Tests failing - fix before committing"
    exit 1
fi

# Check TDD compliance
if ! ./scripts/validate-tdd-cycle.sh; then
    echo "❌ TDD cycle incomplete"
    exit 1
fi

# Check documentation consistency
if ! make docs-check; then
    echo "❌ Documentation inconsistencies found"
    exit 1
fi

# If formatting changed files, require re-add
if ! git diff --quiet; then
    echo "✨ Code was auto-formatted - review and re-add files"
    exit 1
fi

echo "✅ All quality checks passed - commit approved"
```

## Manual Approval Safeguards
- **No Auto-Commits**: LLM proposes, user approves, then LLM executes
- **Quality Automation**: Formatting, linting, testing happen automatically
- **Manual Control**: User always writes/approves commit messages
- **Documentation Sync**: Automated checks ensure documentation consistency
- **Transparency**: User sees exact git commands before execution

## Getting Help

### Development Support
- **Read this guide completely** before starting development
- **Follow TDD methodology strictly** - no exceptions
- **Use the issue tracker** for bugs and feature requests
- **Join discussions** for questions and clarifications
- **Follow project conventions** established in codebase

### Code of Conduct
- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain professional communication
- Follow documentation standards rigorously

### Resources
- **[Architecture Guide](ARCHITECTURE.md)** for technical details
- **[Usage Guide](USAGE.md)** for CLI reference
- **[Features Roadmap](FEATURES.md)** for planned development
- **[Changelog](CHANGELOG.md)** for version history

## See Also

### Documentation for Contributors
- 📖 **[README.md](README.md)** - Project overview and contribution motivation
- 🏗️ **[ARCHITECTURE.md](ARCHITECTURE.md)** - Deep technical understanding for development
- 📚 **[USAGE.md](USAGE.md)** - User experience perspective for feature development
- 📋 **[FEATURES.md](FEATURES.md)** - Implementation status and priority guidance

### Development Resources
- 🌐 **[Project Repository](https://github.com/swi/repeater)** - Issues, discussions, and pull requests
- 🔧 **[Development Setup](CONTRIBUTING.md#development-environment)** - Complete environment configuration
- 🧪 **[TDD Methodology](CONTRIBUTING.md#tdd-workflow-mandatory)** - Test-driven development workflow
- 📝 **[Documentation Standards](CONTRIBUTING.md#documentation-standards-mandatory)** - Maintaining quality documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.