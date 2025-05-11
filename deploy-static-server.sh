#!/bin/bash

# Deploy static server to Railway

echo "===== Deploying Static Server to Railway ====="
echo "This will create a new service in your Railway project that serves mock data"
echo "The frontend can use this service while the database issues are being resolved"

# Create a new service in Railway
echo "Creating new service in Railway..."
railway service create static-api

# Link the service to the current project
echo "Linking service to the current project..."
railway link

# Use the static configuration
echo "Using static server configuration..."
cp railway.static.json railway.json

# Deploy the service
echo "Deploying static server..."
railway up

# Output the service URL
echo "Static server deployed!"
echo "You can access it at: https://<your-service-url>"
echo "Update your frontend to use this URL instead of the main API URL"