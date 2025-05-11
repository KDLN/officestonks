# Using the CORS Proxy for WebSocket Connections

## Overview

To solve the WebSocket connection issues, we've created a CORS proxy service that eliminates all CORS problems. This document explains how to use it.

## Setup Instructions

### 1. Deploy the CORS Proxy Service

1. Create a new service in your Railway project
2. Select "Deploy from GitHub repo"
3. Choose the `/cors-proxy` directory from the repository
4. Set the following environment variables:
   - `BACKEND_URL`: `https://web-production-1e26.up.railway.app` (your backend service URL)
5. Deploy the service
6. Note the URL of your deployed proxy service (e.g., `https://officestonks-cors-proxy.up.railway.app`)

### 2. Update Frontend Code

Make the following changes to the frontend code:

```javascript
// In /frontend/src/services/websocket.js

// CHANGE THIS LINE
// FROM:
const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';

// TO: (replace with your actual CORS proxy URL)
const apiUrl = process.env.REACT_APP_API_URL || 'https://officestonks-cors-proxy.up.railway.app';
```

That's it! No other changes are needed. The CORS proxy:
- Handles all CORS headers
- Proxies WebSocket connections
- Forwards API requests to the backend

### 3. How It Works

With this change:
- API requests will be sent to `/api/*` on the proxy
- WebSocket connections will go to `/ws` on the proxy
- The proxy ensures proper CORS headers on all responses
- The proxy forwards all requests to the actual backend

## Troubleshooting

If you continue to experience issues:

1. **Check proxy logs:** Look at the logs of the CORS proxy service in Railway
2. **Verify proxy URL:** Make sure you're using the correct URL for the proxy
3. **Check proxy health:** Visit `https://your-proxy-url/health` to verify the proxy is running
4. **Ensure backend URL is correct:** Verify the `BACKEND_URL` environment variable is set correctly

## Additional Notes

- The CORS proxy works for both development and production
- No changes needed to backend services
- The proxy handles both HTTP and WebSocket connections
- Deploying your own proxy gives you full control over CORS handling

## Security Considerations

- The CORS proxy should be deployed on the same Railway project as your other services
- HTTPS is automatically handled by Railway
- The proxy doesn't modify or store any of the data passing through it