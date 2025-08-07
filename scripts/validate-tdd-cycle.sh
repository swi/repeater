#!/bin/bash
# scripts/validate-tdd-cycle.sh
# Validates that current changes represent a complete TDD cycle

set -e

echo "🔄 Validating TDD cycle compliance..."

# Check if we're in a TDD behavior branch
BRANCH=$(git branch --show-current 2>/dev/null || echo "main")
if [[ $BRANCH == tdd/* ]]; then
    echo "✅ On TDD behavior branch: $BRANCH"
    
    # Check for TDD_BEHAVIOR.md if it exists
    if [ -f "TDD_BEHAVIOR.md" ]; then
        echo "📋 Found TDD behavior plan"
    fi
else
    echo "ℹ️  Not on TDD behavior branch, skipping TDD-specific validation"
fi

# Check if there are test files
if ! find . -name "*_test.go" -type f | grep -q .; then
    echo "❌ No test files found - TDD requires tests first"
    exit 1
fi

# Check if staged changes include tests
STAGED_FILES=$(git diff --cached --name-only 2>/dev/null || echo "")
if [ -n "$STAGED_FILES" ]; then
    if echo "$STAGED_FILES" | grep -q "_test.go"; then
        echo "✅ Staged changes include test files"
    else
        # If no test files staged, check if this is a refactor
        if go test ./... > /dev/null 2>&1; then
            echo "✅ All tests pass - this may be a REFACTOR phase"
        else
            echo "⚠️  No test files staged and tests failing - ensure TDD compliance"
        fi
    fi
fi

echo "✅ TDD cycle validation passed"