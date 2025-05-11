#!/bin/bash
set -e

# Script to prepare and deploy to Railway after cleanup

echo "=================================="
echo "OfficeStonks Railway Deployment"
echo "=================================="

# Build Go binary
echo "Building Go binary..."
go build -o server ./cmd/api/main.go

# Check if build was successful
if [ ! -f "./server" ]; then
  echo "Server build failed! Aborting deployment."
  exit 1
fi

echo "Build successful!"

# Ensure all files are executable
# Check if files exist before making them executable
[ -f "docker-entrypoint.sh" ] && chmod +x docker-entrypoint.sh
[ -f "start.sh" ] && chmod +x start.sh
[ -f "start-server.sh" ] && chmod +x start-server.sh

# Check if the Railway CLI is installed
if ! command -v railway &> /dev/null; then
  echo "Railway CLI not found. Please install it with:"
  echo "npm i -g @railway/cli"
  exit 1
fi

# Verify Railway login status
echo "Checking Railway login status..."
railway whoami || {
  echo "Not logged in to Railway. Please run 'railway login' first."
  exit 1
}

# Commit the cleanup changes (optional - uncomment if you want to commit)
# echo "Committing changes to Git..."
# git add -A
# git commit -m "Clean up repository structure and fix CORS/database issues"

# Deploy to Railway
echo "Deploying to Railway..."
railway up

echo "=================================="
echo "Deployment completed!"
echo "=================================="
echo "Note: Your application should now be building on Railway."
echo "Check the Railway dashboard for build status and logs."
echo "To test the CORS configuration, run:"
echo "curl -I -X OPTIONS -H \"Origin: https://officestonks-frontend-production.up.railway.app\" https://web-production-1e26.up.railway.app/api/auth/login"
echo "=================================="