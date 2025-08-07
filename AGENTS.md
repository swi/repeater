# Repeater (rpr) - Agent Development Guide

## Project Overview
This is a Go-based CLI tool for continuous command execution with intelligent scheduling, rate limiting, and monitoring capabilities. The project is currently in design phase with comprehensive documentation but no implementation yet.

## Build/Test Commands
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

## TDD Workflow (MANDATORY)
**NEVER write implementation code without tests first. Follow this exact sequence:**

### Red-Green-Refactor Cycle
1. **RED**: Write a failing test that describes the desired behavior
2. **GREEN**: Write minimal code to make the test pass
3. **REFACTOR**: Improve code while keeping tests green

### Before Writing Any Code
- [ ] Create test file first (`*_test.go`)
- [ ] Write failing test with clear assertions
- [ ] Run test to confirm it fails: `go test -v ./pkg/...`
- [ ] Only then implement the minimal code to pass

### TDD Task Scoping
**Break every task into TDD-sized chunks (1-3 test cases per commit)**
- **Micro-task**: 1 test case, 1 commit (15-30 min)
- **Small task**: 2-3 related test cases, 1-2 commits (1-2 hours)
- **Medium task**: Multiple behaviors, 3-5 commits (half day)
- **Large task**: Break down further - never start without breakdown

## LLM Commit Proposal Process
**After completing any TDD phase, LLM must propose commit for user approval:**

### Commit Proposal Format
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

### User Approval Options
- **y/yes**: Execute the proposed commit exactly as shown
- **n/no**: Cancel commit, allow manual review/modification
- **e/edit**: Modify the commit message before executing
- **d/diff**: Show detailed changes before deciding

## TDD Branching Strategy

### Branch Types
```
main (production-ready, all tests pass)
├── develop (integration branch, all TDD cycles complete)
├── feature/scheduler-core (feature branch, multiple TDD cycles)
│   ├── tdd/scheduler-creation (single behavior, 1-3 TDD cycles)
│   ├── tdd/scheduler-timing (single behavior, 1-3 TDD cycles)
│   └── tdd/scheduler-cleanup (single behavior, 1-3 TDD cycles)
```

### Branch Naming Convention
- **Feature branches**: `feature/<component>-<capability>`
- **TDD behavior branches**: `tdd/<specific-behavior>`
- **Bug fix branches**: `fix/<issue-description>`

### TDD Behavior Branch Workflow
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

# If formatting changed files, require re-add
if ! git diff --quiet; then
    echo "✨ Code was auto-formatted - review and re-add files"
    exit 1
fi

echo "✅ All quality checks passed - commit approved"
```

## TDD Quality Gates

### Coverage Requirements
- Minimum 85% test coverage: `go test -cover ./...`
- No untested public functions
- All error paths must be tested

### Test Quality Checks
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

### TDD Anti-Patterns (FORBIDDEN)
- ❌ Writing implementation before tests
- ❌ Writing tests after implementation ("test-after")
- ❌ Skipping the "failing test" step
- ❌ Not running tests frequently during development
- ❌ Ignoring test failures or making tests pass by changing assertions

## Code Style Guidelines

### Project Structure
- Follow standard Go project layout: `cmd/`, `pkg/`, `internal/`, `tests/`
- Use Test-Driven Development (TDD) - write tests before implementation
- Maintain >85% test coverage for all packages

### Naming Conventions
- Use Go standard naming: PascalCase for exported, camelCase for unexported
- Interface names end with -er suffix (e.g., `Scheduler`, `Executor`)
- Package names are lowercase, single word when possible

### Error Handling
- Return errors as the last return value
- Use `fmt.Errorf()` for error wrapping with context
- Handle all errors explicitly - no silent failures
- Use custom error types for domain-specific errors

### Imports
- Group imports: standard library, third-party, local packages
- Use goimports for automatic formatting
- Avoid dot imports except in tests

### Types and Interfaces
- Prefer small, focused interfaces
- Use context.Context for cancellation and timeouts
- Implement String() method for custom types used in logging

### Concurrency
- Use channels for communication between goroutines
- Always handle context cancellation in long-running operations
- Protect shared state with mutexes or atomic operations
- Test for race conditions with `go test -race`

## Key Architecture Patterns
- **Scheduler Interface**: Pluggable scheduling algorithms (interval, count, duration, rate-limited)
- **Executor Pattern**: Command execution with timeout, output capture, and context support
- **Configuration**: TOML files with environment variable overrides
- **Daemon Integration**: Coordinate with patience daemon for multi-instance rate limiting

## Manual Approval Safeguards
- **No Auto-Commits**: LLM proposes, user approves, then LLM executes
- **Quality Automation**: Formatting, linting, testing happen automatically
- **Manual Control**: User always writes/approves commit messages
- **Transparency**: User sees exact git commands before execution