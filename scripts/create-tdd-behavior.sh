#!/bin/bash
# scripts/create-tdd-behavior.sh
# Creates a new TDD behavior branch with planning template

BEHAVIOR_NAME=$1
FEATURE_BRANCH=${2:-$(git branch --show-current)}

if [ -z "$BEHAVIOR_NAME" ]; then
    echo "Usage: $0 <behavior-name> [feature-branch]"
    echo "Example: $0 scheduler-creation feature/scheduler-core"
    exit 1
fi

echo "🌿 Creating TDD behavior branch: tdd/$BEHAVIOR_NAME"

# Create and setup TDD behavior branch
git checkout $FEATURE_BRANCH 2>/dev/null || {
    echo "❌ Feature branch '$FEATURE_BRANCH' not found"
    echo "💡 Available branches:"
    git branch
    exit 1
}

git checkout -b tdd/$BEHAVIOR_NAME

# Create TDD behavior plan
cat > TDD_BEHAVIOR.md << EOF
# TDD Behavior: $BEHAVIOR_NAME

## Test Cases to Implement:
- [ ] Test case 1: [Description]
- [ ] Test case 2: [Description]  
- [ ] Test case 3: [Description]

## TDD Cycle Progress:
- [ ] RED: Failing tests written
- [ ] GREEN: Minimal implementation
- [ ] REFACTOR: Code improvement (if needed)

## Acceptance Criteria:
- All tests pass
- Coverage increased
- Behavior fully implemented
- Code follows project standards

## Notes:
[Add any implementation notes or considerations]
EOF

git add TDD_BEHAVIOR.md
git commit -m "docs: TDD plan for $BEHAVIOR_NAME behavior

- Created behavior-specific branch: tdd/$BEHAVIOR_NAME
- Added TDD planning template
- Ready to start RED phase"

echo "✅ TDD behavior branch created: tdd/$BEHAVIOR_NAME"
echo "📝 Edit TDD_BEHAVIOR.md to plan your test cases"
echo "🔴 Start with RED phase - write failing tests first"
echo ""
echo "Next steps:"
echo "1. Edit TDD_BEHAVIOR.md with specific test cases"
echo "2. Create test file: touch pkg/[component]/[component]_test.go"
echo "3. Write failing tests (RED phase)"
echo "4. Run: go test -v ./pkg/[component]/ (should fail)"
echo "5. Implement minimal code (GREEN phase)"
echo "6. Refactor if needed (REFACTOR phase)"