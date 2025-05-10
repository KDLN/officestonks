#!/bin/bash

# Post-deployment verification script for OfficeStonks
# This script checks if the deployment was successful and provides next steps

echo "=== OFFICE STONKS POST-DEPLOYMENT VERIFICATION ==="

# Check Railway API status
echo "Checking deployed API status..."
API_URL="https://web-production-1e26.up.railway.app/api"
HEALTH_CHECK=$(curl -s "${API_URL}/health")

if [[ "$HEALTH_CHECK" == *"running"* ]]; then
  echo "✅ API is running"
else
  echo "❌ API health check failed"
  echo "Response: $HEALTH_CHECK"
fi

# Check stocks endpoint
echo "Checking stocks endpoint..."
STOCKS=$(curl -s "${API_URL}/stocks")
if [[ "$STOCKS" == *"["* ]]; then
  echo "✅ Stocks endpoint is returning data"
  echo "First few stocks: $(echo "$STOCKS" | head -c 100)..."
else
  echo "❌ Stocks endpoint is not returning data"
  echo "Response: $STOCKS"
fi

# Check CORS debug endpoint
echo "Checking CORS configuration..."
CORS_CHECK=$(curl -s "${API_URL}/debug/cors")
if [[ "$CORS_CHECK" == *"success"* ]]; then
  echo "✅ CORS configuration is working"
else
  echo "❌ CORS check failed"
  echo "Response: $CORS_CHECK"
fi

echo ""
echo "=== ADMIN PANEL USAGE INSTRUCTIONS ==="
echo "1. Log in to the frontend with:"
echo "   Username: admin"
echo "   Password: admin123"
echo ""
echo "2. Navigate to the Admin Panel"
echo "3. Check if you can see users list"
echo "4. Try resetting stock prices - this should work now"
echo "5. Try clearing chat messages - this should work now"
echo ""

echo "=== DATABASE FIX INSTRUCTIONS ==="
echo "If you still have issues, run the database fix script in DBeaver:"
echo "1. Open DBeaver and connect to your Railway MySQL database"
echo "2. Create a new SQL script"
echo "3. Copy the contents of fix-admin-user.sql into the script"
echo "4. Execute the script"
echo ""

echo "=== FIX SUMMARY ==="
echo "The deployed version includes these fixes:"
echo "1. Simplified stock price reset without transactions"
echo "2. Fixed CORS headers for admin endpoints"
echo "3. Improved error handling and logging"
echo "4. Simplified chat clearing without transactions"
echo "5. Added comprehensive database fix script"
echo ""

echo "For details, see ADMIN_FIX_SUMMARY.md"
echo "================================================="

exit 0