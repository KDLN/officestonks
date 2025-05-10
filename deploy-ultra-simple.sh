#!/bin/bash
# Deploy the ultra-simple CORS proxy

# Step 1: Setup
echo "Setting up the ultra-simple CORS proxy..."
cp ultra-simple-package.json package.json
npm install express

# Step 2: Test locally
echo "Starting the proxy locally for testing..."
node ultra-simple-proxy.js &
PID=$!
sleep 2

# Step 3: Test the health endpoint
echo "Testing the health endpoint..."
curl -s http://localhost:3000/health

# Step 4: Kill the local process
kill $PID

# Step 5: Instructions for Railway deployment
echo ""
echo "===== DEPLOYMENT INSTRUCTIONS ====="
echo "To deploy on Railway:"
echo "1. Create a new service in Railway"
echo "2. Use this repository"
echo "3. Set the build command to: cp ultra-simple-package.json package.json && npm install"
echo "4. Set the start command to: node ultra-simple-proxy.js"
echo "5. Deploy the service"
echo ""
echo "After deployment, update the frontend to use the new proxy URL for admin endpoints"
echo "===================================="