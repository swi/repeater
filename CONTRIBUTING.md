# Contributing to Repeater

Thank you for your interest in contributing to Repeater! This document provides guidelines for contributing to the project.

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
```

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