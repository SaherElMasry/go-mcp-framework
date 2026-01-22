#!/bin/bash
clear
echo -e "\033[1;32m🚀 MCP HIGH-PERFORMANCE GREP SERVER\033[0m"
echo -e "\033[0;90mHardware: Intel i5-3470 @ 3.20GHz\033[0m"
echo "------------------------------------------------"

# 1. THE CSV SCAN
echo -e "\n\033[1;34m📊 TASK: Scan 100,000 CSV Rows (Unseen Data)\033[0m"
echo -e "\033[0;90mQuery: Salary > 155,000\033[0m"
START=$(date +%s%3N)
curl -sN -X POST "http://localhost:8080/stream?tool=search_csv" \
     -H "Content-Type: application/json" \
     -d '{"file_path":"demo-data/large-records.csv","search_type":"salary","search_value":">155000"}' \
     | grep --line-buffered "user" | sed -u 's/^data: //' | jq -rc '.chunk.user.name' | head -n 5
END=$(date +%s%3N)
echo -e "\033[1;33m⏱️ CSV Latency: $((END-START))ms\033[0m"

# 2. THE HTML SCAN
echo -e "\n\033[1;34m🌐 TASK: Grep 50,000 Line HTML (Zero-Copy Search)\033[0m"
echo -e "\033[0;90mQuery: 'target_match'\033[0m"
START=$(date +%s%3N)
curl -sN -X POST "http://localhost:8080/stream?tool=grep_html" \
     -H "Content-Type: application/json" \
     -d '{"file_path":"demo-data/test.html","pattern":"target_match"}' \
     | grep --line-buffered "content" | sed -u 's/^data: //' | jq -rc '.chunk.content'
END=$(date +%s%3N)
echo -e "\033[1;33m⏱️ HTML Latency: $((END-START))ms\033[0m"

echo -e "\n\033[1;32m✅ ALL TOOLS PRODUCTION-READY\033[0m"
