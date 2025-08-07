#!/bin/bash
# scripts/tdd-commit-helper.sh
# Helps with TDD commit process but doesn't auto-commit

echo "🔄 TDD Commit Helper"
echo "==================="

# 1. Run quality checks first
echo "Running pre-commit checks..."
if [ -f ".git/hooks/pre-commit" ]; then
    if ! .git/hooks/pre-commit; then
        echo "❌ Quality checks failed - fix issues first"
        exit 1
    fi
else
    echo "⚠️  No pre-commit hook found - running basic checks"
    go fmt ./...
    if ! go test ./...; then
        echo "❌ Tests failing"
    fi
fi

# 2. Detect TDD phase
if ! go test ./... > /dev/null 2>&1; then
    PHASE="RED"
    echo "🔴 Detected RED phase (tests failing)"
else
    PHASE="GREEN"  
    echo "🟢 Detected GREEN/REFACTOR phase (tests passing)"
fi

# 3. Show what's staged
echo ""
echo "📋 Staged changes:"
git diff --cached --name-only

# 4. Suggest commit message structure (don't auto-commit)
echo ""
echo "💡 Suggested commit message structure:"
echo "----------------------------------------"
if [ "$PHASE" = "RED" ]; then
    echo "test(red): add failing test for [BEHAVIOR]"
    echo ""
    echo "- Test [SPECIFIC_FUNCTIONALITY]"
    echo "- Currently fails - no implementation yet"
    echo "- Part of TDD cycle for [FEATURE]"
    echo ""
    echo "TDD-Phase: RED"
    echo "Behavior: [behavior-name]"
    echo "Tests-Added: [number]"
else
    echo "feat(green): implement [BEHAVIOR]"
    echo ""
    echo "- Add [IMPLEMENTATION_DETAILS]"  
    echo "- Tests now pass"
    echo "- Minimal implementation for TDD cycle"
    echo ""
    echo "TDD-Phase: GREEN"
    echo "Behavior: [behavior-name]"
    echo "Tests-Modified: [number]"
fi

echo ""
echo "✏️  Now run: git commit"
echo "📝 Write your own commit message based on the template above"