#!/bin/bash

# Crisis Mechanics Testing Script
# This script tests the crisis/bankruptcy/recovery system

echo "🧪 Crisis Mechanics Testing Script"
echo "================================="

# Configuration
API_URL="${API_URL:-http://localhost:8080/api}"
ADMIN_TOKEN="${ADMIN_TOKEN}"

if [ -z "$ADMIN_TOKEN" ]; then
    echo "❌ Error: ADMIN_TOKEN environment variable not set"
    echo "Please set: export ADMIN_TOKEN=your_admin_jwt_token"
    exit 1
fi

# Helper function for API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$data" ]; then
        curl -s -X "$method" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            "$API_URL$endpoint"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $ADMIN_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint"
    fi
}

# Test 1: Get current simulator status
echo -e "\n📊 Test 1: Getting simulator status..."
api_call GET "/admin/crisis/status" | jq '.'

# Test 2: Force crisis on a specific stock
echo -e "\n🚨 Test 2: Forcing crisis event on stock ID 1..."
api_call POST "/admin/crisis/force" '{"stock_id": 1}' | jq '.'
sleep 2

# Test 3: Check for crisis news
echo -e "\n📰 Test 3: Checking for crisis news..."
api_call GET "/news?limit=5" | jq '.[] | select(.type == "crisis")'

# Test 4: Force bankruptcy
echo -e "\n💀 Test 4: Forcing bankruptcy on stock ID 2..."
api_call POST "/admin/crisis/bankruptcy" '{"stock_id": 2}' | jq '.'
sleep 2

# Test 5: Check portfolio losses (requires user endpoint)
echo -e "\n💸 Test 5: Checking portfolio losses..."
# This would need a user endpoint to check losses

# Test 6: Force recovery
echo -e "\n🚀 Test 6: Forcing recovery on stock ID 3..."
api_call POST "/admin/crisis/recovery" '{"stock_id": 3}' | jq '.'
sleep 2

# Test 7: Check sector contagion by forcing multiple events
echo -e "\n🔥 Test 7: Testing sector contagion..."
echo "Forcing crisis on multiple stocks in same sector..."
for i in 4 5 6; do
    echo "  - Stock ID $i"
    api_call POST "/admin/crisis/force" "{\"stock_id\": $i}" > /dev/null
    sleep 1
done

# Final status check
echo -e "\n📊 Final Status Check:"
api_call GET "/admin/crisis/status" | jq '.'

echo -e "\n✅ Crisis mechanics testing completed!"
echo "Check the logs and database for detailed results."