#!/bin/bash

echo "===== Deploying Mock API to Railway ====="
echo "This will create a new mock API service that doesn't need a database"

# Change to the mock API directory
cd /home/kdln/code/officestonks/mock-api

# Check if railway command exists
if ! command -v railway &> /dev/null; then
  echo "Railway CLI not found. Please install it first."
  exit 1
fi

# Initialize a git repository if not already
if [ ! -d ".git" ]; then
  git init
  git add .
  git commit -m "Initial mock API commit"
fi

# Run railway commands
echo "Creating new project..."
railway project create officestonks-mock

echo "Linking to project..."
railway link

echo "Deploying..."
railway up

echo "===== Mock API Deployed ====="
echo "Once deployed, update your frontend to use this new API URL"
echo "=============================="