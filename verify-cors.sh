#!/bin/bash
# Verify CORS configuration for OfficeStonks API

# Configuration
API_URL="https://web-production-1e26.up.railway.app"
TOKEN=$1 # Pass JWT token as first argument

if [ -z "$TOKEN" ]; then
  echo "Please provide a JWT token as an argument"
  echo "Usage: ./verify-cors.sh eyJhbGc..."
  exit 1
fi

# Text formatting
BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test a specific endpoint
test_endpoint() {
  local endpoint=$1
  local method=$2
  local description=$3
  
  echo -e "\n${BOLD}Testing $description${NC}"
  echo "Endpoint: $endpoint"
  echo "Method: $method"
  
  # First test OPTIONS preflight
  echo -e "\n${BOLD}OPTIONS preflight request${NC}"
  curl -s -I -X OPTIONS \
    -H "Origin: https://officestonks-frontend-production.up.railway.app" \
    -H "Access-Control-Request-Method: $method" \
    -H "Access-Control-Request-Headers: Content-Type, Authorization" \
    "$API_URL$endpoint" | grep -i "access-control"
  
  # Then test actual request
  echo -e "\n${BOLD}$method request${NC}"
  response=$(curl -s -X $method \
    -H "Origin: https://officestonks-frontend-production.up.railway.app" \
    -H "Authorization: Bearer $TOKEN" \
    "$API_URL$endpoint")
  
  echo "Response: $response"
  
  if [[ "$response" == *"error"* ]]; then
    echo -e "${RED}Failed${NC}"
  else
    echo -e "${GREEN}Success${NC}"
  fi
}

echo -e "${BOLD}CORS Verification for OfficeStonks API${NC}"
echo "API URL: $API_URL"
echo "Testing with origin: https://officestonks-frontend-production.up.railway.app"

# Test various endpoints
test_endpoint "/api/health" "GET" "Health Check"
test_endpoint "/api/admin/users" "GET" "Admin Users"
test_endpoint "/api/admin/stocks/reset" "GET" "Reset Stocks (GET)"
test_endpoint "/api/admin/stocks/reset" "POST" "Reset Stocks (POST)"
test_endpoint "/api/admin/chat/clear" "POST" "Clear Chat"

echo -e "\n${BOLD}CORS Verification Complete${NC}"
echo "If you see 'Access-Control-Allow-Origin: *' in the responses, CORS is configured correctly."