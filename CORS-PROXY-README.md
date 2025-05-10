# CORS Proxy for OfficeStonks Admin API

A simplified Express server that proxies requests to the OfficeStonks backend API while handling CORS headers correctly.

## Overview

This proxy server solves CORS issues with admin API endpoints by:

1. Correctly handling OPTIONS preflight requests
2. Adding proper CORS headers to all responses
3. Forwarding API requests to the backend
4. Handling authentication token passing

## Files

- `simple-cors-proxy.js`: Main proxy server implementation
- `simple-package.json`: Dependencies for the proxy
- `Dockerfile.simple`: Docker configuration for containerized deployment
- `cors-debug.html`: Browser-based debugging tool for CORS
- `cors-test.html`: Simple test page for verifying CORS works
- `test-cors.sh`: Bash script for testing endpoints from command line

## Setup

1. Copy `simple-package.json` to `package.json`:
   ```
   cp simple-package.json package.json
   ```

2. Install dependencies:
   ```
   npm install
   ```

3. Run the proxy:
   ```
   node simple-cors-proxy.js
   ```

## Configuration

The proxy can be configured using environment variables:

- `PORT`: Port to listen on (default: 3001)
- `API_BASE_URL`: URL of the backend API (default: https://web-production-1e26.up.railway.app)

## Endpoints

The proxy provides the following endpoints:

- `/health`: Health check endpoint
- `/cors-test`: CORS testing endpoint
- `/admin-test`: Admin endpoint test
- `/admin/*`: All admin API endpoints, proxied to the backend

## Making Requests

Admin API requests should include a JWT token:

```
GET https://proxy-url/admin/users?token=your-jwt-token
```

The proxy will:
1. Add the token to the Authorization header
2. Forward the request to the backend
3. Ensure CORS headers are set in the response

## Testing

Use the included testing tools to verify CORS is working:

- Open `cors-debug.html` in a browser to test endpoints visually
- Run `test-cors.sh` to test from the command line
- Check the browser console for CORS-related errors

## Deployment

See `RAILWAY_DEPLOYMENT.md` for instructions on deploying to Railway.

## Troubleshooting

If you encounter CORS issues:

1. Check that the OPTIONS preflight request returns 200 OK
2. Verify that the Access-Control-Allow-Origin header is present in responses
3. Ensure the backend API is accessible from the proxy
4. Check that token authentication is working correctly