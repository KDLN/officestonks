# OfficeStonks CORS Proxy (v1.1.1)

A comprehensive CORS proxy service for OfficeStonks that eliminates CORS issues between the frontend and backend, with enhanced handling for all routes.

## Overview

**⚠️ IMPORTANT UPDATE (v1.1.1)**:
- This version includes enhanced CORS handling for all routes including login/auth endpoints
- Fixes the "No 'Access-Control-Allow-Origin' header" errors for login requests
- Adds special handling for preflight OPTIONS requests on all API routes
- Specifically targets the frontend URL: `https://officestonks-frontend-production.up.railway.app`

This proxy service sits between your frontend and backend, handling CORS headers properly and forwarding requests to the backend service. It supports both regular HTTP requests and WebSocket connections.

## Features

- Complete CORS header management
- WebSocket proxy support
- Admin API endpoint handling
- Bearer token forwarding
- Detailed request logging
- Health check endpoint
- IPv4 compatibility mode

## How It Works

1. Frontend makes requests to this proxy instead of directly to the backend
2. Proxy adds appropriate CORS headers to all responses
3. Proxy forwards requests to the actual backend service
4. WebSocket connections are also proxied properly

## Environment Configuration

All sensitive configuration should be provided via environment variables. The proxy uses the following environment variables:

### Required Environment Variables
- `BACKEND_URL`: The URL of the backend API service
- `JWT_SECRET`: A secure random string used for JWT token verification

### Optional Environment Variables
- `PORT`: The port to listen on (defaults to 3000)
- `DB_HOST`: Database hostname (if your proxy needs direct DB access)
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `MYSQL_TCP_PROTOCOL`: Set to "4" to force IPv4 connections
- `IPV6_DISABLED`: Set to "true" to disable IPv6
- `GODEBUG`: Set to "netdns=go" to use Go's native DNS resolver

## Setup

1. Copy `.env.example` to `.env` and fill in the appropriate values
2. Install dependencies: `npm install`
3. Start the server: `npm start`

For security, generate a proper JWT secret with:
```
node generate-jwt-secret.js
```

## Deployment

### Railway Deployment

1. Make sure you have the Railway CLI installed:
   ```
   npm i -g @railway/cli
   ```

2. Login to your Railway account:
   ```
   railway login
   ```

3. Navigate to the CORS proxy directory:
   ```
   cd /path/to/officestonks/cors-proxy
   ```

4. Deploy directly using our script:
   ```
   ./deploy-to-railway.sh
   ```

   Or deploy manually:
   ```
   railway up
   ```

5. After deployment, set or verify these environment variables in the Railway dashboard:
   - `BACKEND_URL`: URL of your backend API service
   - `JWT_SECRET`: A secure random string for JWT verification
   - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: If database access is needed
   - `MYSQL_TCP_PROTOCOL=4`, `IPV6_DISABLED=true`, `GODEBUG=netdns=go`: For IPv4 compatibility

### Testing The Deployment

Once deployed, test the CORS proxy:

1. Test the health endpoint:
   ```
   curl https://your-cors-proxy-url.up.railway.app/health
   ```

2. Test admin routes preflight:
   ```
   curl -X OPTIONS -i https://your-cors-proxy-url.up.railway.app/api/admin/stocks/reset -H "Origin: https://officestonks-frontend-production.up.railway.app"
   ```

3. Test auth routes preflight (new in v1.1.1):
   ```
   curl -X OPTIONS -i https://your-cors-proxy-url.up.railway.app/api/auth/login -H "Origin: https://officestonks-frontend-production.up.railway.app"
   ```

   Both should return HTTP 204 with proper CORS headers.

### Usage in Frontend

Update your frontend code to use the proxy URL:

```javascript
// Before
const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';

// After
const apiUrl = process.env.REACT_APP_API_URL || 'https://officestonks-cors-proxy.up.railway.app';
```

WebSocket connections will automatically be proxied through the `/ws` endpoint.

## Endpoints

- `/api/*` - Proxy for REST API calls
- `/api/auth/*` - Authentication endpoints with special CORS handling (new in v1.1.1)
- `/api/admin/*` - Admin endpoints with special CORS handling
- `/ws/*` - Proxy for WebSocket connections
- `/health` - Health check endpoint
- `/emergency/*` - Emergency admin endpoints
- `/debug_admin_status` - Debug endpoint for admin status
- `/debug/headers` - Debug endpoint to check request headers

## Special Route Handling

This version (v1.1.1) includes enhanced CORS handling for all API routes with special focus on:

### Auth API Routes (New in v1.1.1)

- Enhanced support for login and authentication endpoints
- Dedicated middleware for authentication routes
- Fixed CORS issues with the login process
- Auth requests use the path `/api/auth/*` (Example: `/api/auth/login`)

### Admin API Routes

- Special handling for preflight OPTIONS requests to admin endpoints
- Dedicated middleware for admin API routes with proper CORS headers
- Fixed the "No 'Access-Control-Allow-Origin' header" error for admin routes
- Admin requests use the path `/api/admin/*` (Example: `/api/admin/stocks/reset`)

### Global Improvements

- Enhanced preflight OPTIONS handling for all endpoints
- Extra CORS headers in all responses to ensure compatibility
- Explicit frontend URL allowance in CORS configuration
- Detailed logging for troubleshooting CORS issues


## Local Development

```bash
# Start with environment variables
npm install
BACKEND_URL=http://localhost:8080 JWT_SECRET=your-secret-key npm start
```

## Testing Database Connectivity

To test database connectivity:
```bash
# With environment variables set
DB_HOST=turntable.proxy.rlwy.net DB_NAME=railway DB_PASSWORD=your-password DB_PORT=28889 DB_USER=root node db-test.js
```

## Security Notes

1. Never hardcode credentials in scripts or source code
2. Do not log or expose sensitive information
3. Use environment variables for all sensitive configuration
4. Keep your JWT_SECRET value secure and never commit it to source control
5. Use the `.env.example` template as a reference, but create your own `.env` file with actual values

## Benefits

- Eliminates CORS issues permanently
- Keeps your API secure
- Easy to deploy on Railway
- Works with both HTTP and WebSocket connections
- Handles authentication token forwarding
- Detailed logging for debugging
- Minimal configuration required
