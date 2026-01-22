#!/bin/bash

# GitHub MCP Server - Complete Test Suite
BASE_URL="http://localhost:8080/rpc"
TOTAL=0
PASSED=0
FAILED=0

test_tool() {
    local test_name=$1
    local tool_name=$2
    local args=$3

    TOTAL=$((TOTAL + 1))

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Test $TOTAL: $test_name"
    echo "Tool: $tool_name"

    response=$(curl -s -X POST "$BASE_URL" \
        -H 'Content-Type: application/json' \
        -d "{\"jsonrpc\":\"2.0\",\"id\":$TOTAL,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool_name\",\"arguments\":$args}}")

    if echo "$response" | jq -e '.result' > /dev/null 2>&1; then
        echo "✅ PASSED"
        PASSED=$((PASSED + 1))
        echo "$response" | jq -r '.result.content[0].text' | jq '.' 2>/dev/null | head -20
    else
        echo "❌ FAILED"
        FAILED=$((FAILED + 1))
        echo "$response" | jq '.'
    fi
    echo ""
}

echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║   🧪  GitHub MCP Server - Test Suite                             ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""

# USER TOOLS
echo "👤 USER TOOLS"
test_tool "Get user" "get_user" "{}"

# REPOSITORY TOOLS
echo "📚 REPOSITORY TOOLS"
test_tool "List repos" "list_repos" '{"per_page":3}'
test_tool "Get repo" "get_repo" '{"owner":"golang","repo":"go"}'
test_tool "Get README" "get_readme" '{"owner":"golang","repo":"go"}'

# ISSUE TOOLS
echo "🐛 ISSUE TOOLS"
test_tool "List issues" "list_issues" '{"owner":"golang","repo":"go","state":"open","per_page":3}'
test_tool "Get issue" "get_issue" '{"owner":"golang","repo":"go","number":1}'

# STAR TOOLS
echo "⭐ STAR TOOLS"
test_tool "Is starred" "is_starred" '{"owner":"golang","repo":"go"}'

# SEARCH TOOLS
echo "🔎 SEARCH TOOLS"
test_tool "Search repos" "search_repos" '{"query":"language:go stars:>10000","per_page":3}'

# META TOOLS
echo "📊 META TOOLS"
test_tool "Rate limit" "get_rate_limit" "{}"

# SUMMARY
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║   📊 TEST SUMMARY                                                 ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Total:  $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ ALL TESTS PASSED! 🎉"
else
    echo "❌ SOME TESTS FAILED"
fi
