#!/bin/bash

# Proxy Cleanup Script for OfficeStonks
# This script removes redundant proxy implementations and keeps the cors-proxy

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== OfficeStonks Proxy Cleanup ===${NC}"
echo -e "${YELLOW}This script will remove redundant proxy implementations and keep the cors-proxy.${NC}"
read -p "Do you want to continue? (y/n): " -n 1 -r
echo # Move to a new line

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Cleanup aborted.${NC}"
    exit 1
fi

echo -e "${BLUE}Creating backup of all proxy implementations...${NC}"
mkdir -p proxy-backup
cp -r cors-proxy minimal-proxy reverse-proxy proxy-backup/

echo -e "${BLUE}Examining proxy implementations...${NC}"
echo -e "${YELLOW}cors-proxy: $(wc -l ./cors-proxy/index.js | awk '{print $1}') lines${NC}"
echo -e "${YELLOW}minimal-proxy: $(wc -l ./minimal-proxy/index.js | awk '{print $1}') lines${NC}"
echo -e "${YELLOW}reverse-proxy: $(wc -l ./reverse-proxy/index.js | awk '{print $1}') lines${NC}"

echo -e "${BLUE}Decision: Keep cors-proxy as it has the most complete implementation with WebSocket support${NC}"

echo -e "${BLUE}Updating README files...${NC}"
# Create a new consolidated README in the cors-proxy folder
cat > cors-proxy/README-new.md << EOF
# OfficeStonks CORS Proxy

A comprehensive CORS proxy service for OfficeStonks that eliminates CORS issues between the frontend and backend.

## Overview

This proxy service sits between your frontend and backend, handling CORS headers properly and forwarding requests to the backend service. It supports both regular HTTP requests and WebSocket connections.

## Features

- Complete CORS header management
- WebSocket proxy support
- Admin API endpoint handling
- Bearer token forwarding
- Detailed request logging
- Health check endpoint

## How It Works

1. Frontend makes requests to this proxy instead of directly to the backend
2. Proxy adds appropriate CORS headers to all responses
3. Proxy forwards requests to the actual backend service
4. WebSocket connections are also proxied properly

## Deployment

### Railway Deployment

1. Create a new service in your Railway project
2. Connect to this repository directory \`/cors-proxy\`
3. Set the following environment variables:
   - \`BACKEND_URL\`: URL of your backend service (e.g., \`https://web-production-1e26.up.railway.app\`)
   - \`PORT\`: Optional, defaults to 3000

### Usage in Frontend

Update your frontend code to use the proxy URL:

\`\`\`javascript
// Before
const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';

// After 
const apiUrl = process.env.REACT_APP_API_URL || 'https://your-cors-proxy-url.up.railway.app';
\`\`\`

WebSocket connections will automatically be proxied through the \`/ws\` endpoint.

## Endpoints

- \`/api/*\` - Proxy for REST API calls
- \`/ws/*\` - Proxy for WebSocket connections
- \`/health\` - Health check endpoint
- \`/admin/*\` - Admin API endpoints (automatically adds /api prefix)

## Local Development

\`\`\`bash
npm install
npm run dev
\`\`\`

## Benefits

- Eliminates CORS issues permanently
- Keeps your API secure
- Easy to deploy on Railway
- Works with both HTTP and WebSocket connections
- Handles authentication token forwarding
- Detailed logging for debugging
- Minimal configuration required
EOF

# Rename the new README
mv cors-proxy/README-new.md cors-proxy/README.md

echo -e "${BLUE}Removing redundant proxy implementations...${NC}"
rm -rf minimal-proxy reverse-proxy

echo -e "${GREEN}Proxy cleanup completed!${NC}"
echo -e "${BLUE}The cors-proxy implementation has been kept as the primary proxy solution.${NC}"
echo -e "${YELLOW}A backup of all proxy implementations has been created in the proxy-backup directory.${NC}"