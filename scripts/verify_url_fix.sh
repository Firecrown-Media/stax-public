#!/bin/bash
# Verification script for URL replacement fix

set -e

echo "=== URL Replacement Fix Verification ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
PASSED=0
FAILED=0

# Function to print test result
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((FAILED++))
}

warn() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

# 1. Check that files exist
echo "1. Checking modified files exist..."
FILES=(
    "pkg/wordpress/cli.go"
    "pkg/wordpress/search_replace.go"
    "cmd/db_helpers.go"
    "pkg/wordpress/cli_test.go"
    "pkg/wordpress/search_replace_test.go"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        pass "File exists: $file"
    else
        fail "File missing: $file"
    fi
done
echo ""

# 2. Check for new functions
echo "2. Checking new functions are present..."

if grep -q "func (c \*CLI) VerifySiteURL" pkg/wordpress/cli.go; then
    pass "VerifySiteURL function found in cli.go"
else
    fail "VerifySiteURL function not found in cli.go"
fi

if grep -q "func parseSearchReplaceOutput" pkg/wordpress/search_replace.go; then
    pass "parseSearchReplaceOutput function found in search_replace.go"
else
    fail "parseSearchReplaceOutput function not found in search_replace.go"
fi

if grep -q "type SearchReplaceResult struct" pkg/wordpress/search_replace.go; then
    pass "SearchReplaceResult type found in search_replace.go"
else
    fail "SearchReplaceResult type not found in search_replace.go"
fi
echo ""

# 3. Check for critical flags
echo "3. Checking for DNS workaround flags..."

if grep -q "\-\-skip-themes" pkg/wordpress/search_replace.go; then
    pass "--skip-themes flag present"
else
    fail "--skip-themes flag missing"
fi

if grep -q "\-\-skip-plugins" pkg/wordpress/search_replace.go; then
    pass "--skip-plugins flag present"
else
    fail "--skip-plugins flag missing"
fi

if grep -q "\-\-all-tables" pkg/wordpress/search_replace.go; then
    pass "--all-tables flag present"
else
    fail "--all-tables flag missing"
fi
echo ""

# 4. Check for verification logic
echo "4. Checking for verification logic in db_helpers.go..."

if grep -q "Verifying URL replacement" cmd/db_helpers.go; then
    pass "Verification message found"
else
    fail "Verification message not found"
fi

if grep -q "VerifySiteURL" cmd/db_helpers.go; then
    pass "VerifySiteURL call found in runSearchReplace"
else
    fail "VerifySiteURL call not found in runSearchReplace"
fi
echo ""

# 5. Run unit tests
echo "5. Running unit tests..."

if go test ./pkg/wordpress -v -run TestParseSearchReplaceOutput > /dev/null 2>&1; then
    pass "parseSearchReplaceOutput tests pass"
else
    fail "parseSearchReplaceOutput tests failed"
fi

if go test ./pkg/wordpress > /dev/null 2>&1; then
    pass "All WordPress package tests pass"
else
    fail "WordPress package tests failed"
fi

if go test ./cmd > /dev/null 2>&1; then
    pass "All cmd package tests pass"
else
    fail "cmd package tests failed"
fi
echo ""

# 6. Check code quality
echo "6. Running code quality checks..."

if go vet ./pkg/wordpress/... > /dev/null 2>&1; then
    pass "go vet: pkg/wordpress"
else
    fail "go vet failed for pkg/wordpress"
fi

if go vet ./cmd/... > /dev/null 2>&1; then
    pass "go vet: cmd"
else
    fail "go vet failed for cmd"
fi

# Check formatting
UNFORMATTED=$(gofmt -l pkg/wordpress/*.go cmd/db_helpers.go 2>/dev/null || true)
if [ -z "$UNFORMATTED" ]; then
    pass "gofmt: all files formatted"
else
    fail "gofmt: unformatted files: $UNFORMATTED"
fi
echo ""

# 7. Build binary
echo "7. Building binary..."

if go build -o /tmp/stax-verify . > /dev/null 2>&1; then
    pass "Binary builds successfully"
    rm -f /tmp/stax-verify
else
    fail "Binary build failed"
fi
echo ""

# Summary
echo "=== Summary ==="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All checks passed! ✓${NC}"
    echo ""
    echo "The URL replacement fix is ready for deployment."
    exit 0
else
    echo -e "${RED}Some checks failed. Please review.${NC}"
    exit 1
fi
