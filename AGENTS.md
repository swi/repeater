# Repeater Development Guide - Agent Instructions

## 🎉 Project Status: **PRODUCTION READY (v0.5.0)**

This is a Go-based CLI tool for continuous command execution with intelligent scheduling and monitoring capabilities. **The project is feature-complete and production-ready** with all core features implemented and thoroughly tested.

## 📚 **CRITICAL: Documentation Standards (MANDATORY)**

**ALL AGENTS MUST FOLLOW THESE DOCUMENTATION RULES:**

### Core Documentation Structure

The project uses a **streamlined 6-file documentation structure**. Never create additional markdown files or duplicate content.

#### **Required Documentation Files**

1. **README.md** - Project landing page (150-200 lines)
   - Concise overview, key features, quick start, links to detailed docs
   - **Update when**: Adding major features, changing core functionality

2. **USAGE.md** - Comprehensive user guide (300-400 lines)
   - Complete CLI reference, examples, configuration, troubleshooting
   - **Update when**: Adding commands, flags, usage patterns

3. **ARCHITECTURE.md** - Technical design
   - System architecture, components, data flow, performance
   - **Update when**: Changing architecture, adding components, refactoring

4. **CONTRIBUTING.md** - Development guidelines
   - TDD methodology, code standards, plugin development, documentation rules
   - **Update when**: Changing development processes

5. **FEATURES.md** - Roadmap and planning
   - Implementation status, feature roadmap, priorities
   - **Update when**: Completing features, changing roadmap

6. **CHANGELOG.md** - Version history
   - Detailed release notes, breaking changes, new features
   - **Update when**: Making releases

### **MANDATORY Documentation Update Rules**

#### **When Adding New Features:**
- [ ] **README.md**: Add feature to summary list
- [ ] **USAGE.md**: Add comprehensive examples and usage patterns
- [ ] **ARCHITECTURE.md**: Update if design changes
- [ ] **FEATURES.md**: Mark feature as complete
- [ ] **CHANGELOG.md**: Add entry with details

#### **When Modifying CLI:**
- [ ] **USAGE.md**: Update ALL affected command examples
- [ ] **README.md**: Update quick start if needed
- [ ] Test ALL examples for accuracy
- [ ] Update help text and error messages

#### **When Changing Architecture:**
- [ ] **ARCHITECTURE.md**: Update design diagrams and descriptions
- [ ] **CONTRIBUTING.md**: Update if development process changes
- [ ] **README.md**: Update if core functionality changes

#### **Documentation Quality Requirements:**
1. **Accuracy**: All examples must be tested and functional
2. **Completeness**: Cover all features and use cases
3. **Consistency**: Use consistent terminology across files
4. **No Duplication**: Never duplicate content between files
5. **Version Consistency**: Keep version numbers synchronized

#### **PROHIBITED Actions:**
- ❌ **Creating new markdown files** (use existing 6-file structure)
- ❌ **Duplicating content** across files
- ❌ **Outdated examples** or references
- ❌ **Conflicting information** between files
- ❌ **Version inconsistencies** across documentation

### **Documentation Enforcement Commands**

```bash
# MANDATORY: Run before any commit
make docs-check           # Check consistency across all files
make docs-test           # Test all code examples
make docs-validate       # Validate links and references

# Use during development
make docs-lint           # Check formatting and style
make docs-examples-test  # Test CLI examples specifically
```

## 🔧 **Build/Test Commands**

```bash
# Build the binary
go build -o rpr ./cmd/rpr

# Development workflow
make test                    # Unit tests
make test-integration        # Integration tests  
make test-e2e               # End-to-end tests
make benchmark              # Performance tests
make quality-gate           # All quality checks (REQUIRED before commits)

# Testing specific areas
go test -v ./pkg/scheduler/  # Test specific package
go test -run TestSpecificFunction ./pkg/...  # Test specific function

# Quality assurance
make lint                   # Run golangci-lint
go fmt ./...               # Format code
make docs-check             # MANDATORY: Check documentation consistency
```

## 🧪 **TDD Workflow (MANDATORY)**

**NEVER write implementation code without tests first. This is strictly enforced.**

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

## 📝 **Commit Requirements**

### LLM Commit Proposal Process
**After completing any TDD phase, you MUST propose commits for user approval:**

```
🔄 TDD Phase Complete - Commit Proposal
=====================================

📊 Changes Summary:
- Files modified: [list with documentation updates]
- TDD Phase: [RED/GREEN/REFACTOR]
- Tests status: [before] → [after]
- Behavior: [specific behavior implemented]
- Documentation updated: [YES/NO with list]

📝 Proposed Commit:
```bash
git add [files including documentation]
git commit -m "[type](phase): [description]

- [specific change 1]
- [specific change 2]
- [documentation updates]

TDD-Phase: [RED/GREEN/REFACTOR]
Behavior: [behavior-name]
Tests-Added/Modified: [number]
Documentation-Updated: [files]"
```

❓ Approve this commit? (y/n/e/d)
```

### Git Commands for Documentation Restructuring

```bash
# When restructuring documentation (like this cleanup)
git add README.md USAGE.md ARCHITECTURE.md CONTRIBUTING.md FEATURES.md CHANGELOG.md
git add .archive/docs-old/  # Include archived files
git commit -m "docs: restructure documentation for clarity and eliminate duplication

- Consolidate README.md to concise project overview (150 lines)
- Expand USAGE.md with comprehensive examples and patterns
- Create ARCHITECTURE.md from consolidated technical content
- Update CONTRIBUTING.md with complete development guidelines
- Add FEATURES.md roadmap and implementation status
- Archive duplicate files: PROJECT_STRUCTURE.md, IMPLEMENTATION_PLANNING.md, 
  ADVANCED_FEATURES_PLAN.md, examples/USAGE_EXAMPLES.md, project-status/
- Eliminate 80%+ content duplication across 20 files → 6 focused files
- Standardize version references to v0.5.0 throughout
- Implement mandatory documentation sync requirements

Documentation-Restructure: 20-files → 6-files
Content-Duplication: Eliminated
Version-Consistency: v0.5.0 standardized
Enforcement-Rules: Added to AGENTS.md and CONTRIBUTING.md"
```

## 🎯 **Available Features (v0.5.0)**

### **CLI Commands**
- **CLI with Abbreviations**: `rpr i -e 30s -t 5 -- curl api.com`
- **8 Execution Modes**: interval, count, duration, cron, adaptive, backoff, load-aware, rate-limit
- **Pattern Matching**: `--success-pattern`, `--failure-pattern`, `--case-insensitive`
- **HTTP-Aware Intelligence**: `--http-aware`, `--http-max-delay`, `--http-custom-fields`
- **Output Control**: `--quiet`, `--verbose`, `--stats-only`, `--stream`

### **Production Features**
- **Plugin System**: Extensible architecture for custom schedulers and executors
- **Signal Handling**: Graceful shutdown on Ctrl+C
- **Configuration**: TOML files with environment variable overrides
- **Health Endpoints**: HTTP server for monitoring
- **Metrics**: Prometheus-compatible metrics export
- **Statistics**: Comprehensive execution metrics and reporting

### **Quality Metrics**
- **210+ Tests**: Comprehensive test coverage (90%+)
- **Performance**: <1% timing deviation, minimal resource usage
- **Race Testing**: Concurrent execution safety verified
- **Integration Testing**: End-to-end workflow validation

## 🏗️ **Key Architecture Patterns**

- **Scheduler Interface**: Pluggable scheduling algorithms
- **Plugin System**: Interface-based plugins for schedulers, executors, outputs
- **Executor Pattern**: Command execution with timeout, output capture, context support
- **Configuration**: TOML files with environment variable overrides
- **Pattern Matching**: Regex-based success/failure detection with precedence
- **HTTP-Aware**: Automatic API response parsing for optimal scheduling

## 🚀 **Development Guidelines**

### **Code Standards**
- **Go Formatting**: `go fmt` and `goimports` (automated)
- **Linting**: `golangci-lint` must pass
- **Testing**: `go test -race ./...` must pass  
- **Documentation**: All public APIs documented
- **Coverage**: Minimum 90% test coverage

### **Quality Gates (MANDATORY)**
```bash
# MUST pass before any commit
make quality-gate

# Individual checks
make test              # All tests pass
make lint              # Linting passes
make docs-check        # Documentation consistency
make benchmark         # Performance benchmarks
```

### **TDD Anti-Patterns (FORBIDDEN)**
- ❌ Writing implementation before tests
- ❌ Writing tests after implementation ("test-after")
- ❌ Skipping the "failing test" step
- ❌ Ignoring test failures
- ❌ Making tests pass by changing assertions
- ❌ Creating documentation without updating ALL affected files

## 🔄 **Agent Workflow Summary**

### **For Any Code Changes:**
1. **TDD First**: Write failing tests, then minimal implementation
2. **Quality Gates**: Run `make quality-gate` before proposing commits
3. **Documentation Sync**: Update ALL affected documentation files
4. **Commit Proposals**: Always propose commits for user approval
5. **Validation**: Ensure all examples work and documentation is consistent

### **For Documentation Changes:**
1. **Update ALL affected files** according to documentation rules
2. **Test ALL examples** for accuracy
3. **Check consistency** across all 6 core files
4. **Validate links** and references
5. **Run documentation checks** before proposing commits

### **For Feature Additions:**
1. **Follow TDD methodology** strictly
2. **Update README.md** with feature summary
3. **Add comprehensive examples** to USAGE.md
4. **Update ARCHITECTURE.md** if design changes
5. **Mark complete** in FEATURES.md
6. **Add changelog entry** with details

## 🚨 **Critical Reminders**

- **NO NEW MARKDOWN FILES**: Use the 6-file structure only
- **NO CONTENT DUPLICATION**: Each piece of information lives in exactly one place
- **ALL EXAMPLES MUST WORK**: Test every code example before committing
- **DOCUMENTATION IS MANDATORY**: Never code without updating docs
- **TDD IS ENFORCED**: No implementation without tests first
- **USER APPROVAL REQUIRED**: All commits must be approved before execution

This development guide ensures consistent, high-quality contributions while maintaining the streamlined documentation structure and production-ready codebase.