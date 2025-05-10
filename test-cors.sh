#!/bin/bash
# Script to test CORS configuration for OfficeStonks CORS proxy

# Set to the CORS proxy URL
PROXY_URL="https://officestonks-cors-proxy.up.railway.app"
# Set to the frontend origin
FRONTEND_ORIGIN="https://officestonks-frontend-production.up.railway.app"

echo "=== CORS Tests for OfficeStonks Proxy ==="
echo "Proxy URL: $PROXY_URL"
echo "Frontend Origin: $FRONTEND_ORIGIN"
echo ""

# Test 1: Health check endpoint
echo "1. Testing health check endpoint..."
curl -s $PROXY_URL/health | jq .
echo ""

# Test 2: OPTIONS preflight to /cors-test
echo "2. Testing OPTIONS preflight to /cors-test..."
curl -s -I -X OPTIONS \
  -H "Origin: $FRONTEND_ORIGIN" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  $PROXY_URL/cors-test
echo ""

# Test 3: GET request to /cors-test
echo "3. Testing GET to /cors-test..."
curl -s \
  -H "Origin: $FRONTEND_ORIGIN" \
  $PROXY_URL/cors-test | jq .
echo ""

# Test 4: OPTIONS preflight to /admin-test
echo "4. Testing OPTIONS preflight to /admin-test..."
curl -s -I -X OPTIONS \
  -H "Origin: $FRONTEND_ORIGIN" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  $PROXY_URL/admin-test
echo ""

# Test 5: GET request to /admin-test
echo "5. Testing GET to /admin-test..."
curl -s \
  -H "Origin: $FRONTEND_ORIGIN" \
  $PROXY_URL/admin-test | jq .
echo ""

# Test 6: OPTIONS preflight to /admin/users
echo "6. Testing OPTIONS preflight to /admin/users..."
curl -s -I -X OPTIONS \
  -H "Origin: $FRONTEND_ORIGIN" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization" \
  $PROXY_URL/admin/users
echo ""

# Test 7: GET to /admin/users (without token)
echo "7. Testing GET to /admin/users (should return 401 if no token)..."
curl -s -i \
  -H "Origin: $FRONTEND_ORIGIN" \
  $PROXY_URL/admin/users
echo ""

echo "=== Tests Complete ==="
echo "If you see Access-Control-Allow-Origin: * in the headers, CORS is working correctly."
echo "Check for 200 OK responses to OPTIONS requests."
echo "If you get 404 errors for admin endpoints, the API route may not exist or may have a different path."