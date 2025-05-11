# Frontend Updates to Use CORS Proxy

## Problem

The frontend is configured with both a direct backend URL and a CORS proxy URL, but it's not actually using the CORS proxy for its connections. This is causing CORS errors.

From the browser console, we can see:

```javascript
API Config: {
  isLocalhost: false,
  BACKEND_URL: 'https://web-production-1e26.up.railway.app',
  API_URL: 'https://web-production-1e26.up.railway.app/api',
  CORS_PROXY_URL: 'https://officestonks-cors-proxy.up.railway.app',
  WS_URL: 'wss://web-production-1e26.up.railway.app/ws'
}
```

Even though `CORS_PROXY_URL` is defined, it's not being used for API or WebSocket connections.

## Required Changes

### 1. Update API Configuration

Look for the file that defines these URLs (likely called `api.js` or `config.js`) and update it to use the CORS proxy:

```javascript
// INCORRECT (current setup):
const API_URL = BACKEND_URL + '/api';
const WS_URL = 'wss://' + BACKEND_URL.replace(/^https?:\/\//, '') + '/ws';

// CORRECT (updated setup):
const API_URL = CORS_PROXY_URL + '/api';
const WS_URL = 'wss://' + CORS_PROXY_URL.replace(/^https?:\/\//, '') + '/ws';
```

### 2. Update HTTP Service

Make sure any HTTP service or fetch wrapper is using the proper API_URL:

```javascript
// In your http.js or api-service.js
const fetchData = async (endpoint, options = {}) => {
  const url = `${API_URL}/${endpoint}`;
  // Rest of the function...
}
```

### 3. Update WebSocket Service

Ensure the WebSocket connection uses the WS_URL from your config:

```javascript
// In your websocket.js
// Use the WS_URL from your config, not a hardcoded URL
const wsUrl = `${WS_URL}?token=${token}`;
```

## Example Implementation

Here's a complete example for reference:

```javascript
// config.js or api.js
const isLocalhost = window.location.hostname === 'localhost' || 
                    window.location.hostname === '127.0.0.1';

// Base URLs
const BACKEND_URL = isLocalhost 
  ? 'http://localhost:8080' 
  : 'https://web-production-1e26.up.railway.app';

const CORS_PROXY_URL = isLocalhost
  ? 'http://localhost:3000'
  : 'https://officestonks-cors-proxy.up.railway.app';

// Derived URLs - USE THE PROXY for all external connections
const API_URL = CORS_PROXY_URL + '/api';
const WS_URL = 'wss://' + CORS_PROXY_URL.replace(/^https?:\/\//, '') + '/ws';

export {
  isLocalhost,
  BACKEND_URL,
  CORS_PROXY_URL,
  API_URL,
  WS_URL
};
```

## Testing the Changes

After making these changes:

1. Deploy the updated frontend
2. Open the browser console
3. Verify there are no CORS errors
4. Confirm successful API requests and WebSocket connections

## Common Pitfalls

- Make sure **all** API requests go through the proxy, not just some
- Ensure WebSocket connections are also routed through the proxy
- If using authentication, the proxy will forward all headers automatically
- When running locally for development, make sure to run the proxy locally too

If you continue to experience issues, the CORS proxy logs will provide additional diagnostics to help identify the problem.