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
2. Connect to this repository directory `/cors-proxy`
3. Set the following environment variables:
   - `BACKEND_URL`: URL of your backend service (e.g., `https://web-production-1e26.up.railway.app`)
   - `PORT`: Optional, defaults to 3000

### Usage in Frontend

Update your frontend code to use the proxy URL:

```javascript
// Before
const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';

// After 
const apiUrl = process.env.REACT_APP_API_URL || 'https://your-cors-proxy-url.up.railway.app';
```

WebSocket connections will automatically be proxied through the `/ws` endpoint.

## Endpoints

- `/api/*` - Proxy for REST API calls
- `/ws/*` - Proxy for WebSocket connections
- `/health` - Health check endpoint
- `/admin/*` - Admin API endpoints (automatically adds /api prefix)

## Local Development

```bash
npm install
npm run dev
```

## Benefits

- Eliminates CORS issues permanently
- Keeps your API secure
- Easy to deploy on Railway
- Works with both HTTP and WebSocket connections
- Handles authentication token forwarding
- Detailed logging for debugging
- Minimal configuration required
