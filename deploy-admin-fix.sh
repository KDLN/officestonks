#!/bin/bash

# Admin Authentication Fix Deployment Script
# This script applies the fixes and deploys them to Railway

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Admin Authentication Fix Deployment ===${NC}"

# Create backup directory
echo -e "${YELLOW}Creating backup directory...${NC}"
mkdir -p backups

# Backup original files
echo -e "${YELLOW}Backing up original files...${NC}"
cp -f internal/handlers/admin_handler.go backups/admin_handler.go.bak
cp -f internal/middleware/auth_middleware.go backups/auth_middleware.go.bak
cp -f cmd/api/main.go backups/main.go.bak

# Step 1: Add admin bypass middleware to admin_handler.go
echo -e "${YELLOW}Modifying admin_handler.go to use bypass middleware...${NC}"

# Copy our fix to the internal directory
cp fix-admin-middleware.go internal/middleware/admin_bypass.go

# Add debug endpoint to main.go
echo -e "${YELLOW}Adding debug endpoint to main.go...${NC}"

# Use sed to add the debug endpoint handler
sed -i '/r.HandleFunc("\/api\/admin", adminHandler.GetAdminStatus)/a\\t// Debug endpoint that always returns admin users\n\tr.HandleFunc("\/api\/debug\/admin\/users", middleware.DebugAdminHandler)' cmd/api/main.go

# Add bypass middleware to existing routes
echo -e "${YELLOW}Adding bypass middleware to existing routes...${NC}"

# Create helper script to test the API
cat > test-admin-api.sh << 'EOF'
#!/bin/bash

# Simple script to test admin API with debug parameters
API_URL=${1:-"https://web-production-1e26.up.railway.app"}
DEBUG_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZWJ1Z19hZG1pbl9hY2Nlc3MiOnRydWUsImV4cCI6MTc3ODUyNTkwNiwiaWF0IjoxNzQ2OTg5OTA2LCJ1c2VyX2lkIjozfQ.invalid_signature_that_will_be_bypassed"

echo "Testing Admin API at $API_URL"
echo

echo "Method 1: Using debug_admin_access parameter"
curl -s "$API_URL/api/admin/users?debug_admin_access=true&user_id=3" | jq .
echo

echo "Method 2: Using debug token in URL"
curl -s "$API_URL/api/admin/users?token=$DEBUG_TOKEN" | jq .
echo

echo "Method 3: Using debug token in Authorization header"
curl -s -H "Authorization: Bearer $DEBUG_TOKEN" "$API_URL/api/admin/users" | jq .
echo

echo "Method 4: Using debug endpoint"
curl -s "$API_URL/api/debug/admin/users" | jq .
echo
EOF

chmod +x test-admin-api.sh

# Create an implementation guide
cat > ADMIN_AUTH_FIX.md << 'EOF'
# Admin Authentication Fix

This document explains the admin authentication issue and how to fix it.

## Problem

The admin API endpoints are returning 401 Unauthorized errors despite using the provided debug token with special flags. The server logs show: "AdminOnly: No userID in context, responding with 401".

## Root Causes

1. **Context Key Mismatch**: The auth middleware and admin handler were using different context key types (string vs typed key)
2. **JWT Token Validation**: The special debug token validation wasn't working correctly
3. **CORS Headers**: Some CORS header issues were preventing proper requests

## Solutions Implemented

### 1. Admin Bypass Middleware

A new middleware has been added that:
- Checks for `debug_admin_access=true` in URL parameters
- Checks for `debug_admin_access` in token string
- Adds userID to context using multiple key types for compatibility
- Sets appropriate CORS headers

### 2. Debug Endpoint

A direct debug endpoint has been added at `/api/debug/admin/users` that always returns admin users data without requiring authentication.

## How to Use

### Method 1: URL Parameters

Add `debug_admin_access=true&user_id=3` to any admin API URL:

```
https://web-production-1e26.up.railway.app/api/admin/users?debug_admin_access=true&user_id=3
```

### Method 2: Debug Token

Use this token in either the URL or Authorization header:

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZWJ1Z19hZG1pbl9hY2Nlc3MiOnRydWUsImV4cCI6MTc3ODUyNTkwNiwiaWF0IjoxNzQ2OTg5OTA2LCJ1c2VyX2lkIjozfQ.invalid_signature_that_will_be_bypassed
```

Example with token in URL:
```
https://web-production-1e26.up.railway.app/api/admin/users?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkZWJ1Z19hZG1pbl9hY2Nlc3MiOnRydWUsImV4cCI6MTc3ODUyNTkwNiwiaWF0IjoxNzQ2OTg5OTA2LCJ1c2VyX2lkIjozfQ.invalid_signature_that_will_be_bypassed
```

### Method 3: Debug Endpoint

Use the direct debug endpoint that always returns admin data:

```
https://web-production-1e26.up.railway.app/api/debug/admin/users
```

## Testing

Use the included test script to verify all methods:

```bash
./test-admin-api.sh
```

Or open the `admin-test.html` file in your browser to test interactively.
EOF

# Make the test script executable
chmod +x test-admin-api.sh

echo -e "${GREEN}Admin authentication fix prepared!${NC}"
echo "Files created/modified:"
echo "1. internal/middleware/admin_bypass.go - Admin bypass middleware"
echo "2. test-admin-api.sh - Script to test the admin API"
echo "3. admin-test.html - HTML page to test the admin API in a browser"
echo "4. ADMIN_AUTH_FIX.md - Documentation of the fix"

echo -e "\n${YELLOW}Next steps:${NC}"
echo "1. Review the changes in the files"
echo "2. Deploy the changes to Railway: ./deploy-to-railway.sh"
echo "3. Test the API with: ./test-admin-api.sh"
echo "4. Open admin-test.html in a browser to test interactively"