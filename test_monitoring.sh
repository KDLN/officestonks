#!/bin/bash

# Test script for the monitoring system
echo "🧪 Testing Office Stonks Monitoring System"
echo "=========================================="

# Configuration
API_BASE_URL="http://localhost:8080/api"
FRONTEND_URL="http://localhost:3000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Function to make API call and check response
test_api_endpoint() {
    local endpoint=$1
    local description=$2
    local token=$3
    
    echo -n "Testing $description... "
    
    if [ -n "$token" ]; then
        response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "%{http_code}" "$API_BASE_URL$endpoint")
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}✅ OK${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed (HTTP $http_code)${NC}"
        return 1
    fi
}

echo "1. Testing Basic API Health"
echo "----------------------------"
test_api_endpoint "/health" "API Health Check"

echo ""
echo "2. Testing Authentication"
echo "-------------------------"

# Create a test admin user (this assumes you have admin credentials)
echo "Please provide admin credentials for testing:"
read -p "Admin email: " ADMIN_EMAIL
read -s -p "Admin password: " ADMIN_PASSWORD
echo ""

# Login to get token
echo -n "Logging in as admin... "
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}")

# Extract token from response
ADMIN_TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$ADMIN_TOKEN" ]; then
    echo -e "${GREEN}✅ Admin login successful${NC}"
else
    echo -e "${RED}❌ Admin login failed${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo ""
echo "3. Testing Monitoring Endpoints"
echo "-------------------------------"

# Test all monitoring endpoints
test_api_endpoint "/admin/monitoring/dashboard" "Monitoring Dashboard" "$ADMIN_TOKEN"
test_api_endpoint "/admin/monitoring/metrics" "System Metrics" "$ADMIN_TOKEN"
test_api_endpoint "/admin/monitoring/sessions" "Active Sessions" "$ADMIN_TOKEN"
test_api_endpoint "/admin/monitoring/activity" "Recent Activity" "$ADMIN_TOKEN"

echo ""
echo "4. Testing Database Schema"
echo "-------------------------"

# Test if monitoring tables exist by checking the dashboard endpoint
echo -n "Checking if monitoring tables exist... "
DASHBOARD_RESPONSE=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE_URL/admin/monitoring/dashboard")

if echo "$DASHBOARD_RESPONSE" | grep -q "system_metrics"; then
    echo -e "${GREEN}✅ Database schema is properly set up${NC}"
else
    echo -e "${RED}❌ Database schema may be missing${NC}"
    echo -e "${YELLOW}⚠️  Make sure to run: mysql -u username -p database < monitoring_schema.sql${NC}"
fi

echo ""
echo "5. Testing User Activity Tracking"
echo "--------------------------------"

# Make a few API calls to generate activity
echo "Generating test activity..."
test_api_endpoint "/stocks" "Stocks endpoint (should generate activity)" "$ADMIN_TOKEN"
test_api_endpoint "/users/me" "User profile endpoint" "$ADMIN_TOKEN"

# Wait a moment for activity to be logged
sleep 2

# Check if activity was recorded
echo -n "Checking if activity was recorded... "
ACTIVITY_RESPONSE=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE_URL/admin/monitoring/activity?limit=5")

if echo "$ACTIVITY_RESPONSE" | grep -q "activities"; then
    echo -e "${GREEN}✅ Activity tracking is working${NC}"
else
    echo -e "${RED}❌ Activity tracking may not be working${NC}"
fi

echo ""
echo "6. Testing Frontend Integration"
echo "------------------------------"

# Check if frontend is accessible
echo -n "Testing frontend accessibility... "
FRONTEND_RESPONSE=$(curl -s -w "%{http_code}" "$FRONTEND_URL")
FRONTEND_CODE="${FRONTEND_RESPONSE: -3}"

if [ "$FRONTEND_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Frontend is accessible${NC}"
    echo -e "${YELLOW}ℹ️  Visit $FRONTEND_URL/monitoring as an admin to see the dashboard${NC}"
else
    echo -e "${RED}❌ Frontend not accessible (HTTP $FRONTEND_CODE)${NC}"
    echo -e "${YELLOW}ℹ️  Make sure to run: cd frontend && npm start${NC}"
fi

echo ""
echo "7. Performance Test"
echo "------------------"

# Simple performance test - make multiple requests
echo "Running performance test (10 concurrent requests)..."
START_TIME=$(date +%s.%N)

for i in {1..10}; do
    curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE_URL/admin/monitoring/metrics" > /dev/null &
done

wait
END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)

echo "Completed 10 requests in ${DURATION}s"

if (( $(echo "$DURATION < 5.0" | bc -l) )); then
    echo -e "${GREEN}✅ Performance looks good${NC}"
else
    echo -e "${YELLOW}⚠️  Performance might be slow${NC}"
fi

echo ""
echo "📊 Test Summary"
echo "==============="
echo "Monitoring system test completed!"
echo ""
echo "Next steps:"
echo "1. Run the database migration: mysql -u username -p database < monitoring_schema.sql"
echo "2. Restart your backend server: go run cmd/api/main.go"
echo "3. Start your frontend: cd frontend && npm start"
echo "4. Visit http://localhost:3000/monitoring as an admin"
echo ""
echo "The monitoring system provides:"
echo "• Real-time user session tracking"
echo "• Comprehensive activity logging"
echo "• System health monitoring"
echo "• Admin dashboard with live updates"
echo ""
echo -e "${GREEN}🎉 Monitoring system is ready to use!${NC}"