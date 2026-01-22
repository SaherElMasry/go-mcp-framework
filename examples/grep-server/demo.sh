#!/bin/bash

# Function to run search and measure time
run_search() {
    echo -e "\n\033[1;34m🔍 Searching $1 for '$2'...\033[0m"

    # We use sed to remove the "data: " prefix so jq can read it
    # We use 'time' to show the millisecond precision
    time (curl -sN -X POST "http://localhost:8080/stream?tool=search_csv" \
         -H "Content-Type: application/json" \
         -d "{\"file_path\":\"demo-data/info-records.csv\",\"search_type\":\"$1\",\"search_value\":\"$2\"}" \
         | sed 's/^data: //' | jq .)
}

# 1. Search for high salary
run_search "salary" ">120000"

sleep 2 # Pause so viewers can read

# 2. Search for Department
run_search "department" "Engineering"

sleep 2

# 3. Search for Age
run_search "age" ">40"
