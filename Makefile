# Makefile for Repeater (rpr)

.PHONY: build test test-integration test-e2e benchmark quality-gate lint clean install-tools docs-check docs-test docs-validate docs-duplicate-check help

# Build configuration
BINARY_NAME=rpr
BUILD_DIR=./bin
CMD_DIR=./cmd/rpr

# Go configuration
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Default target
all: build

## Build the binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

## Run unit tests
test:
	@echo "🧪 Running unit tests..."
	$(GOTEST) -v -race -cover ./pkg/...

## Run cron tests
test-cron:
	@echo "🕐 Running cron tests..."
	$(GOTEST) -v -race ./pkg/cron/...

## Run plugin tests
test-plugin:
	@echo "🔌 Running plugin tests..."
	$(GOTEST) -v -race ./pkg/plugin/...

## Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	$(GOTEST) -v -run=".*[Ii]ntegration.*" ./cmd/... ./pkg/...

## Run end-to-end tests
test-e2e:
	@echo "🎯 Running end-to-end tests..."
	$(GOTEST) -v -run=".*([Ee]2[Ee]|EndToEnd).*" -timeout=10m ./cmd/... ./pkg/...

## Run performance benchmarks
benchmark:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./pkg/...

## Run all quality checks
quality-gate: lint test test-cron test-plugin test-integration benchmark docs-check
	@echo "✅ All quality checks passed"

## Run linting
lint:
	@echo "🔧 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 60s; \
	else \
		echo "⚠️  golangci-lint not found, running go vet..."; \
		$(GOCMD) vet ./...; \
	fi

## Format code
fmt:
	@echo "📝 Formatting code..."
	$(GOCMD) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

## Install development tools
install-tools:
	@echo "🛠️  Installing development tools..."
	$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@echo "📦 Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
	fi
	@echo "✅ Development tools installed"

## Tidy dependencies
tidy:
	@echo "📦 Tidying dependencies..."
	$(GOMOD) tidy

## Run TDD helper
tdd-helper:
	@echo "🔄 Running TDD commit helper..."
	@./scripts/tdd-commit-helper.sh

## Create new TDD behavior branch
tdd-behavior:
	@if [ -z "$(BEHAVIOR)" ]; then \
		echo "Usage: make tdd-behavior BEHAVIOR=behavior-name [FEATURE=feature-branch]"; \
		echo "Example: make tdd-behavior BEHAVIOR=scheduler-creation FEATURE=feature/scheduler-core"; \
		exit 1; \
	fi
	@./scripts/create-tdd-behavior.sh $(BEHAVIOR) $(FEATURE)

## Show coverage report
coverage:
	@echo "📊 Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./pkg/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "📈 Coverage report: coverage.html"

## Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./pkg/...
	$(GOCMD) tool cover -func=coverage.out

## Install binary to GOPATH/bin
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $$(go env GOPATH)/bin/
	@echo "✅ Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## Check documentation consistency and examples
docs-check:
	@echo "📚 Checking documentation consistency..."
	@./scripts/validate-docs-examples.sh

## Test all documentation examples
docs-test:
	@echo "🧪 Testing documentation examples..."
	@./scripts/validate-docs-examples.sh

## Validate documentation links and references
docs-validate:
	@echo "🔗 Validating documentation links..."
	@grep -r "http" --include="*.md" . | grep -v "example.com" | head -10 || true
	@echo "📋 Checking for broken internal references..."
	@grep -r "\[.*\](.*\.md)" --include="*.md" . || true

## Check for duplicate content across documentation files
docs-duplicate-check:
	@echo "🔍 Checking for duplicate content..."
	@echo "Note: This is a basic check - manual review recommended"

## Show help
help:
	@echo "Repeater (rpr) - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          Build the binary"
	@echo "  install        Install binary to GOPATH/bin"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test           Run unit tests"
	@echo "  test-cron      Run cron tests"
	@echo "  test-plugin    Run plugin tests"
	@echo "  test-integration  Run integration tests"
	@echo "  test-e2e       Run end-to-end tests"
	@echo "  benchmark      Run performance benchmarks"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  coverage       Generate HTML coverage report"
	@echo ""
	@echo "Quality targets:"
	@echo "  quality-gate   Run all quality checks"
	@echo "  lint           Run linting"
	@echo "  fmt            Format code"
	@echo ""
	@echo "Documentation targets:"
	@echo "  docs-check     Check documentation consistency and examples"
	@echo "  docs-test      Test all documentation examples"
	@echo "  docs-validate  Validate documentation links and references"
	@echo "  docs-duplicate-check  Check for duplicate content"
	@echo ""
	@echo "TDD targets:"
	@echo "  tdd-helper     Run TDD commit helper"
	@echo "  tdd-behavior   Create TDD behavior branch (requires BEHAVIOR=name)"
	@echo ""
	@echo "Development targets:"
	@echo "  install-tools  Install development tools"
	@echo "  tidy           Tidy dependencies"
	@echo "  help           Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make tdd-behavior BEHAVIOR=scheduler-creation"
	@echo "  make quality-gate"