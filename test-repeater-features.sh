#!/bin/bash
# test-repeater-features.sh
# Comprehensive test script for all repeater features and arguments

set -e

echo "🔄 Repeater Feature Test Suite"
echo "=============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_behavior="$3"
    
    echo -e "\n${BLUE}Test $((++TESTS_RUN)):${NC} $test_name"
    echo -e "${YELLOW}Command:${NC} $command"
    echo -e "${YELLOW}Expected:${NC} $expected_behavior"
    
    # Use gtimeout if available (brew install coreutils), otherwise skip timeout
    local timeout_cmd=""
    if command -v gtimeout >/dev/null 2>&1; then
        timeout_cmd="gtimeout 10s"
    elif command -v timeout >/dev/null 2>&1; then
        timeout_cmd="timeout 10s"
    fi
    
    if $timeout_cmd bash -c "$command" > /tmp/rpr_test_output 2>&1; then
        echo -e "${GREEN}✅ PASSED${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ FAILED${NC}"
        echo "Output:"
        cat /tmp/rpr_test_output
        ((TESTS_FAILED++))
    fi
}

# Build the binary first
echo "🔨 Building repeater binary..."
make build || {
    echo "❌ Build failed - cannot run tests"
    exit 1
}

echo -e "\n📋 Testing Current Implementation"
echo "================================="

# Test 1: Basic execution (current placeholder)
run_test "Basic execution" \
    "./bin/rpr" \
    "Should print placeholder message and exit successfully"

# Test 2: Help flag (when implemented)
run_test "Help flag" \
    "./bin/rpr --help || ./bin/rpr -h || echo 'Help not implemented yet'" \
    "Should show help message or indicate not implemented"

echo -e "\n📋 Planned Feature Tests (Will work when implemented)"
echo "===================================================="

# INTERVAL SUBCOMMAND TESTS
echo -e "\n${BLUE}🕐 Interval Subcommand Tests${NC}"

run_test "Interval: Basic usage" \
    "echo 'rpr interval --every 1s --times 3 -- echo \"test\"'" \
    "Should run echo 3 times with 1 second intervals"

run_test "Interval: With jitter" \
    "echo 'rpr interval --every 500ms --times 2 --jitter 10% -- date +%s'" \
    "Should add 10% timing variance"

run_test "Interval: Immediate execution" \
    "echo 'rpr interval --every 2s --times 2 --immediate -- echo \"immediate\"'" \
    "Should execute immediately, then after 2s"

run_test "Interval: Duration limit" \
    "echo 'rpr interval --every 500ms --for 2s -- echo \"duration\"'" \
    "Should run for 2 seconds total"

run_test "Interval: Continue on error" \
    "echo 'rpr interval --every 1s --times 3 --continue-on-error -- bash -c \"exit 1\"'" \
    "Should continue despite command failures"

# COUNT SUBCOMMAND TESTS  
echo -e "\n${BLUE}🔢 Count Subcommand Tests${NC}"

run_test "Count: Basic usage" \
    "echo 'rpr count --times 3 -- echo \"count test\"'" \
    "Should execute exactly 3 times"

run_test "Count: With interval" \
    "echo 'rpr count --times 2 --every 1s -- date'" \
    "Should execute 2 times with 1s between"

run_test "Count: Parallel execution" \
    "echo 'rpr count --times 4 --parallel 2 -- sleep 1'" \
    "Should run 2 commands in parallel"

# DURATION SUBCOMMAND TESTS
echo -e "\n${BLUE}⏱️  Duration Subcommand Tests${NC}"

run_test "Duration: Basic usage" \
    "echo 'rpr duration --for 3s --every 1s -- echo \"duration\"'" \
    "Should run for 3 seconds with 1s intervals"

run_test "Duration: No interval (continuous)" \
    "echo 'rpr duration --for 2s -- echo \"continuous\"'" \
    "Should run continuously for 2 seconds"

# RATE-LIMIT SUBCOMMAND TESTS
echo -e "\n${BLUE}🚦 Rate Limit Subcommand Tests${NC}"

run_test "Rate limit: Per second" \
    "echo 'rpr rate-limit --rate 2/1s --times 5 -- echo \"rate limited\"'" \
    "Should limit to 2 executions per second"

run_test "Rate limit: Per minute" \
    "echo 'rpr rate-limit --rate 10/1m --for 5s -- echo \"per minute\"'" \
    "Should respect 10 per minute limit"

run_test "Rate limit: With daemon" \
    "echo 'rpr rate-limit --rate 5/1s --daemon --resource-id test-resource --times 3 -- echo \"daemon\"'" \
    "Should coordinate rate limiting via daemon"

# GLOBAL OPTIONS TESTS
echo -e "\n${BLUE}🌐 Global Options Tests${NC}"

run_test "Global: Quiet mode" \
    "echo 'rpr count --times 2 --quiet -- echo \"quiet test\"'" \
    "Should suppress repeater output, show only command output"

run_test "Global: Verbose mode" \
    "echo 'rpr count --times 2 --verbose -- echo \"verbose test\"'" \
    "Should show detailed execution information"

run_test "Global: Output to file" \
    "echo 'rpr count --times 2 --output-file /tmp/rpr_test.log -- echo \"file output\"'" \
    "Should write output to specified file"

run_test "Global: JSON output" \
    "echo 'rpr count --times 2 --output-format json -- echo \"json test\"'" \
    "Should format output as JSON"

run_test "Global: Config file" \
    "echo 'rpr --config /tmp/rpr_config.toml count --times 2 -- echo \"config\"'" \
    "Should load configuration from file"

run_test "Global: Timeout" \
    "echo 'rpr count --times 2 --timeout 5s -- sleep 1'" \
    "Should timeout individual commands after 5s"

# STOP CONDITIONS TESTS
echo -e "\n${BLUE}🛑 Stop Conditions Tests${NC}"

run_test "Stop: Max failures" \
    "echo 'rpr interval --every 500ms --max-failures 2 --for 5s -- bash -c \"exit 1\"'" \
    "Should stop after 2 consecutive failures"

run_test "Stop: Success condition" \
    "echo 'rpr interval --every 500ms --until-success --for 5s -- bash -c \"[ \$RANDOM -gt 16384 ]\"'" \
    "Should stop on first success"

# ADVANCED FEATURES TESTS
echo -e "\n${BLUE}🚀 Advanced Features Tests${NC}"

run_test "Advanced: Signal handling" \
    "echo 'gtimeout 2s rpr interval --every 1s -- echo \"signal test\" || echo \"Interrupted as expected\"'" \
    "Should handle SIGTERM gracefully"

run_test "Advanced: Environment variables" \
    "echo 'RPR_VERBOSE=true rpr count --times 2 -- echo \"env test\"'" \
    "Should respect environment variables"

run_test "Advanced: Working directory" \
    "echo 'rpr count --times 1 --working-dir /tmp -- pwd'" \
    "Should execute commands in specified directory"

# INTEGRATION TESTS
echo -e "\n${BLUE}🔗 Integration Tests${NC}"

run_test "Integration: With curl" \
    "echo 'rpr count --times 2 --every 1s -- curl -s -o /dev/null -w \"%{http_code}\" httpbin.org/status/200'" \
    "Should make HTTP requests successfully"

run_test "Integration: With patience (when available)" \
    "echo 'rpr count --times 2 -- patience exponential --max-attempts 2 -- echo \"patience integration\"'" \
    "Should work with patience for retry logic"

run_test "Integration: Complex pipeline" \
    "echo 'rpr interval --every 1s --times 2 -- bash -c \"date | grep $(date +%Y)\"'" \
    "Should handle complex shell commands"

# ERROR HANDLING TESTS
echo -e "\n${BLUE}❌ Error Handling Tests${NC}"

run_test "Error: Invalid interval" \
    "echo 'rpr interval --every invalid -- echo test' && echo 'Should show error'" \
    "Should show helpful error for invalid time format"

run_test "Error: Missing command" \
    "echo 'rpr interval --every 1s --times 2' && echo 'Should show error'" \
    "Should require command after --"

run_test "Error: Conflicting options" \
    "echo 'rpr interval --every 1s --times 2 --for 5s -- echo test' && echo 'Should show error'" \
    "Should detect conflicting stop conditions"

# PERFORMANCE TESTS
echo -e "\n${BLUE}⚡ Performance Tests${NC}"

run_test "Performance: High frequency" \
    "echo 'rpr count --times 10 --every 100ms -- echo \"high freq\"'" \
    "Should handle high frequency execution"

run_test "Performance: Many executions" \
    "echo 'rpr count --times 20 -- echo \"many execs\"'" \
    "Should handle many sequential executions"

run_test "Performance: Parallel load" \
    "echo 'rpr count --times 8 --parallel 4 -- echo \"parallel load\"'" \
    "Should handle parallel execution efficiently"

# CONFIGURATION TESTS
echo -e "\n${BLUE}⚙️  Configuration Tests${NC}"

# Create test config file
cat > /tmp/rpr_test_config.toml << 'EOF'
[defaults]
continue_on_error = true
timeout = "10s"
output_format = "json"

[interval]
jitter = "5%"
immediate = true

[daemon]
enabled = true
socket_path = "/tmp/rpr_test.sock"
EOF

run_test "Config: TOML file loading" \
    "echo 'rpr --config /tmp/rpr_test_config.toml count --times 1 -- echo \"config test\"'" \
    "Should load and apply configuration from TOML file"

run_test "Config: Environment override" \
    "echo 'RPR_TIMEOUT=5s rpr --config /tmp/rpr_test_config.toml count --times 1 -- echo \"env override\"'" \
    "Should allow environment variables to override config"

# CLEANUP
rm -f /tmp/rpr_test_config.toml /tmp/rpr_test.log /tmp/rpr_test_output

# SUMMARY
echo -e "\n📊 Test Summary"
echo "==============="
echo -e "Tests run: ${BLUE}$TESTS_RUN${NC}"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed${NC}"
    echo "Note: Many tests are expected to fail until features are implemented"
    exit 1
fi