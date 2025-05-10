#!/bin/bash

# This script verifies that the necessary files for deployment are in the right place

echo "=== OFFICE STONKS DEPLOYMENT VERIFICATION ==="
echo "Checking repository structure..."

# Check Docker file
if [ -f /home/kdln/code/officestonks/Dockerfile ]; then
  echo "✅ Dockerfile exists"
else
  echo "❌ Dockerfile not found"
  exit 1
fi

# Check Go files in root structure
if [ -f /home/kdln/code/officestonks/cmd/api/main.go ]; then
  echo "✅ main.go exists"
else
  echo "❌ main.go not found"
  exit 1
fi

if [ -f /home/kdln/code/officestonks/go.mod ] && [ -f /home/kdln/code/officestonks/go.sum ]; then
  echo "✅ go.mod and go.sum exist"
else
  echo "❌ go.mod or go.sum not found"
  exit 1
fi

if [ -f /home/kdln/code/officestonks/schema.sql ]; then
  echo "✅ schema.sql exists"
else
  echo "❌ schema.sql not found"
  exit 1
fi

# Check if schema.sql has the chat_messages table
if grep -q "chat_messages" /home/kdln/code/officestonks/schema.sql; then
  echo "✅ schema.sql contains chat_messages table"
else
  echo "❌ schema.sql missing chat_messages table"
  exit 1
fi

# Check startup script
if [ -f /home/kdln/code/officestonks/start.sh ]; then
  echo "✅ Startup script exists"
else
  echo "❌ Startup script not found"
  exit 1
fi

echo ""
echo "=== VERIFICATION PASSED ==="
echo "All necessary files are in place for Railway deployment."
echo "The Docker file now uses the simplified repository structure."
echo ""
echo "To deploy to Railway:"
echo "1. Push these changes to your GitHub repository"
echo "2. Railway will automatically deploy from the updated Dockerfile"
echo "3. Monitor the build logs in Railway to ensure successful deployment"
echo ""
echo "Improvements made:"
echo "- Eliminated duplicated directory structure"
echo "- Simplified build process in Dockerfile"
echo "- Streamlined startup with single script"
echo "- Better organization for easier maintenance"
echo "================================================="

exit 0