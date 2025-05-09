#!/bin/sh
# Start script for Railway deployment
echo "Starting frontend service..."

# Install curl for healthcheck if not already installed
if ! command -v curl > /dev/null; then
  echo "Installing curl for healthcheck..."
  apk add --no-cache curl
fi

# Start Nginx with detailed logging
echo "Starting Nginx..."
nginx -g 'daemon off;'