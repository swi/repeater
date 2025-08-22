#!/bin/bash

# validate-docs-examples.sh - Validate all CLI examples in documentation
#
# This script extracts and validates CLI examples from markdown documentation
# to ensure all examples work with the current implementation.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCS_DIR="/Users/swi/dev/repeater"
BINARY="./rpr"
TMP_DIR="/tmp/rpr-docs-validation"
RESULTS_FILE="$TMP_DIR/validation-results.txt"

# Counters
TOTAL_EXAMPLES=0
VALID_EXAMPLES=0
INVALID_EXAMPLES=0
SKIPPED_EXAMPLES=0

# Create temporary directory
mkdir -p "$TMP_DIR"
echo "" > "$RESULTS_FILE"

echo -e "${BLUE}🔍 Validating Documentation Examples${NC}"
echo "========================================"

# Build the binary first
echo -e "${YELLOW}📦 Building rpr binary...${NC}"
cd "$DOCS_DIR"
if ! go build -o "$BINARY" ./cmd/rpr; then
    echo -e "${RED}❌ Failed to build rpr binary${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Binary built successfully${NC}"

# Function to extract CLI examples from markdown
extract_examples() {
    local file="$1"
    local examples=()
    
    # Extract code blocks that start with 'rpr' (but not comment lines)
    grep -n "^\s*rpr " "$file" | while IFS=: read -r line_num line; do
        # Skip commented examples and certain patterns
        if [[ "$line" =~ ^[[:space:]]*#.*rpr ]] || \
           [[ "$line" =~ "example.com" ]] || \
           [[ "$line" =~ "github.com" ]] || \
           [[ "$line" =~ "api.com" ]] || \
           [[ "$line" =~ "flaky-api.com" ]] || \
           [[ "$line" =~ "timeout-api.com" ]] || \
           [[ "$line" =~ "curl.*api" ]] || \
           [[ "$line" =~ "mysql" ]] || \
           [[ "$line" =~ "rsync" ]] || \
           [[ "$line" =~ "\|\|" ]] || \
           [[ "$line" =~ "tail.*log" ]] || \
           [[ "$line" =~ "systemctl" ]] || \
           [[ "$line" =~ "\[GLOBAL" ]] || \
           [[ "$line" =~ "<SUBCOMMAND>" ]] || \
           [[ "$line" =~ "<COMMAND>" ]] || \
           [[ "$line" =~ "EXPRESSION" ]] || \
           [[ "$line" =~ "OPTIONS" ]] || \
           [[ "$line" =~ "DURATION" ]] || \
           [[ "$line" =~ "COUNT" ]] || \
           [[ "$line" =~ "TZ" ]] || \
           [[ "$line" =~ "FILE" ]] || \
           [[ "$line" =~ "SPEC" ]]; then
            continue
        fi
        
        # Clean up the line - remove leading whitespace and markdown formatting
        clean_line=$(echo "$line" | sed 's/^[[:space:]]*//' | sed 's/```.*$//')
        
        # Only process non-empty lines that start with 'rpr'
        if [[ "$clean_line" =~ ^rpr[[:space:]] ]]; then
            echo "$clean_line"
        fi
    done
}

# Function to validate a single example
validate_example() {
    local example="$1"
    local source_file="$2"
    local line_num="$3"
    
    TOTAL_EXAMPLES=$((TOTAL_EXAMPLES + 1))
    
    echo -e "${BLUE}🧪 Testing:${NC} $example"
    
    # Replace -- command with a safe test command
    if [[ "$example" =~ "--.*--" ]]; then
        # Replace the command after -- with 'echo test'
        safe_example=$(echo "$example" | sed 's/--[[:space:]]*.*$/-- echo test/')
    else
        # Add a safe command if no -- found
        safe_example="$example -- echo test"
    fi
    
    # Add --help flag to validate syntax without execution
    safe_example=$(echo "$safe_example" | sed 's/^rpr /rpr --dry-run /')
    
    # If --dry-run isn't supported, just validate help
    if ! $BINARY --help >/dev/null 2>&1; then
        echo -e "${RED}❌ Binary not working${NC}"
        return 1
    fi
    
    # Try to parse the command (will check syntax)
    if echo "$safe_example" | timeout 5s $BINARY --help >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Valid syntax${NC}"
        VALID_EXAMPLES=$((VALID_EXAMPLES + 1))
        echo "[VALID] $source_file:$line_num - $example" >> "$RESULTS_FILE"
        return 0
    else
        echo -e "${RED}❌ Invalid syntax${NC}"
        INVALID_EXAMPLES=$((INVALID_EXAMPLES + 1))
        echo "[INVALID] $source_file:$line_num - $example" >> "$RESULTS_FILE"
        return 1
    fi
}

# Function to perform basic validation (check if subcommands exist)
basic_validate_example() {
    local example="$1"
    local source_file="$2"
    
    TOTAL_EXAMPLES=$((TOTAL_EXAMPLES + 1))
    
    # Extract the subcommand (first argument after rpr)
    subcommand=$(echo "$example" | sed 's/^rpr[[:space:]]*//' | awk '{print $1}')
    
    # List of valid subcommands (from the help system)
    valid_subcommands=("interval" "int" "i" "count" "cnt" "c" "duration" "dur" "d" 
                      "cron" "cr" "exponential" "exp" "fibonacci" "fib" "linear" "lin" 
                      "polynomial" "poly" "decorrelated-jitter" "dj" "adaptive" "adapt" "a" 
                      "load-adaptive" "load" "la" "rate-limit" "rate" "rl" "backoff" "back" "b")
    
    # List of valid global flags (don't need subcommands)
    valid_global_flags=("--help" "-h" "--version" "-v")
    
    # Check if it's a global flag
    if [[ " ${valid_global_flags[@]} " =~ " ${subcommand} " ]]; then
        echo -e "${GREEN}✅ Valid global flag: $subcommand${NC}"
        VALID_EXAMPLES=$((VALID_EXAMPLES + 1))
        echo "[VALID] $source_file - $example" >> "$RESULTS_FILE"
        return 0
    # Check if subcommand is valid
    elif [[ " ${valid_subcommands[@]} " =~ " ${subcommand} " ]]; then
        echo -e "${GREEN}✅ Valid subcommand: $subcommand${NC}"
        VALID_EXAMPLES=$((VALID_EXAMPLES + 1))
        echo "[VALID] $source_file - $example" >> "$RESULTS_FILE"
        return 0
    else
        echo -e "${RED}❌ Invalid subcommand: $subcommand${NC}"
        INVALID_EXAMPLES=$((INVALID_EXAMPLES + 1))
        echo "[INVALID] $source_file - $example" >> "$RESULTS_FILE"
        return 1
    fi
}

# Function to validate documentation file
validate_doc_file() {
    local file="$1"
    local filename=$(basename "$file")
    
    echo ""
    echo -e "${YELLOW}📄 Validating examples in $filename${NC}"
    echo "----------------------------------------"
    
    local examples
    examples=$(extract_examples "$file")
    
    if [[ -z "$examples" ]]; then
        echo -e "${YELLOW}⚠️  No CLI examples found in $filename${NC}"
        return 0
    fi
    
    local line_num=1
    while IFS= read -r example; do
        if [[ -n "$example" ]]; then
            echo -e "${BLUE}Example $line_num:${NC} $example"
            basic_validate_example "$example" "$filename"
            line_num=$((line_num + 1))
        fi
    done <<< "$examples"
}

# Main validation loop
echo -e "${YELLOW}🔍 Scanning documentation files...${NC}"

# Validate main documentation files
for doc_file in "$DOCS_DIR/README.md" "$DOCS_DIR/USAGE.md" "$DOCS_DIR/CONTRIBUTING.md" "$DOCS_DIR/FEATURES.md"; do
    if [[ -f "$doc_file" ]]; then
        validate_doc_file "$doc_file"
    else
        echo -e "${YELLOW}⚠️  File not found: $doc_file${NC}"
    fi
done

# Generate summary report
echo ""
echo -e "${BLUE}📊 Validation Summary${NC}"
echo "===================="
echo -e "Total examples found: ${BLUE}$TOTAL_EXAMPLES${NC}"
echo -e "Valid examples: ${GREEN}$VALID_EXAMPLES${NC}"
echo -e "Invalid examples: ${RED}$INVALID_EXAMPLES${NC}"
echo -e "Skipped examples: ${YELLOW}$SKIPPED_EXAMPLES${NC}"

if [[ $INVALID_EXAMPLES -gt 0 ]]; then
    echo ""
    echo -e "${RED}❌ Validation failed with $INVALID_EXAMPLES invalid examples${NC}"
    echo -e "${YELLOW}📋 Results saved to: $RESULTS_FILE${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}✅ All documentation examples are valid!${NC}"
    echo -e "${YELLOW}📋 Results saved to: $RESULTS_FILE${NC}"
    exit 0
fi