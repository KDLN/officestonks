# CORS Issue Fix for OfficeStonks Admin Endpoints

This document explains the changes made to fix CORS issues in the OfficeStonks admin endpoints.

## Problems Identified

1. CORS headers weren't being applied properly to OPTIONS preflight requests
2. Proxy configuration had incorrect path rewriting for admin endpoints
3. Debugging information was insufficient to diagnose issues
4. Error handling was missing for proxy failures

## Changes Made

### 1. CORS Headers & Preflight Handling

- Added early CORS middleware with proper headers before all routes
- Set Access-Control-Allow-Origin to `*` to allow all origins
- Added explicit OPTIONS handler for all routes
- Made sure CORS headers are consistent across all responses

### 2. Proxy Configuration 

- Fixed path rewriting to correctly point to `/api/admin` on the backend
- Added environment variable for API base URL
- Added better error handling for proxy failures
- Supported all HTTP methods (GET, POST, PUT, DELETE, OPTIONS, PATCH)

### 3. Improved Debugging

- Added detailed logging for request/response cycles
- Added specific test endpoints for isolating issues:
  - `/health` - Basic server health check
  - `/cors-test` - Test CORS headers
  - `/admin-test` - Test admin routes directly
- Added version headers for tracking changes

### 4. Other Improvements

- Added proper package.json with all dependencies
- Added morgan for HTTP request logging
- Created a test script for verifying CORS configuration
- Added thorough startup logs

## Testing the Changes

Run `./test-cors.sh` to test the CORS configuration. This script will:

1. Check if basic endpoints are accessible
2. Verify OPTIONS preflight requests are handled correctly
3. Check that CORS headers are present in all responses
4. Test admin endpoint connectivity

## Deployment

1. Install dependencies: `npm install --prefix . --package-lock-only < cors-proxy-package.json`
2. Copy the package.json: `cp cors-proxy-package.json package.json`
3. Deploy the proxy to Railway

## Environment Variables

- `PORT` - Port to listen on (default: 3001)
- `API_BASE_URL` - Backend API URL (default: https://web-production-1e26.up.railway.app)
- `NODE_ENV` - Node environment (development/production)