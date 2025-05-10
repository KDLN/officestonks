#!/bin/bash
# Test script for the simple CORS proxy

# Install dependencies if needed
echo "Checking for node_modules..."
if [ ! -d "node_modules" ]; then
  echo "Installing dependencies from simple-package.json..."
  cp simple-package.json package.json
  npm install
fi

# Start the proxy server
echo "Starting simple CORS proxy on port 3001..."
node simple-cors-proxy.js