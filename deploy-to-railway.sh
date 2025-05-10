#!/bin/bash
# Script to prepare and deploy to Railway

echo "Preparing for Railway deployment..."

# Build Go binary
echo "Building Go binary..."
go build -o server ./cmd/api

# Commit changes to git
echo "Committing changes to Git..."
git add cmd/api/cors.go cmd/api/main.go start.sh start-server.sh
git commit -m "Update CORS handling for cross-origin requests"

# Push to Railway
echo "Pushing to Railway..."
railway up

echo "Deployment to Railway complete. Check the Railway dashboard for deployment status."
echo "Test the new CORS configuration by visiting:"
echo "https://web-copy-production-5b48.up.railway.app/api/debug/cors"
echo "from your browser."

# Provide some help for testing
echo ""
echo "To test the CORS configuration from the frontend, run:"
echo "curl -I -X OPTIONS -H \"Origin: https://officestonks-frontend.vercel.app\" https://web-copy-production-5b48.up.railway.app/api/auth/login"
echo ""
echo "The response should include:"
echo "Access-Control-Allow-Origin: https://officestonks-frontend.vercel.app"
echo "Access-Control-Allow-Credentials: true"