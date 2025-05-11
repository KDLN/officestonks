const express = require('express');
const cors = require('cors');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const port = process.env.PORT || 3000;

// Enable CORS for all requests with credentials support
app.use(cors({
  origin: true, // Reflect the request origin instead of '*' to support credentials
  credentials: true, // Allow credentials (cookies, auth headers)
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin']
}));

// Special handling for OPTIONS requests
app.options('*', cors({
  origin: true,
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin']
}));

// Debug endpoint for checking request headers
app.get('/debug/headers', (req, res) => {
  res.json({
    headers: req.headers,
    message: 'Debug endpoint to check request headers',
    timestamp: new Date().toISOString()
  });
});

// Health check endpoint with enhanced information
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    service: 'CORS Proxy',
    version: '1.1.0', // Added version after authentication improvements
    backends: {
      api: backendUrl,
      websocket: backendUrl
    },
    auth_forwarding: true,
    timestamp: new Date().toISOString()
  });
});

// Configure proxy middleware
const backendUrl = process.env.BACKEND_URL || 'https://web-production-1e26.up.railway.app';
console.log(`Proxy configured to forward requests to: ${backendUrl}`);

// Log all requests with detailed information
app.use((req, res, next) => {
  console.log(`${req.method} ${req.url} from ${req.headers.origin || 'unknown'}`);

  // Log authorization header presence (but not the full token for security)
  if (req.headers.authorization) {
    const authPreview = req.headers.authorization.substring(0, 15) + '...';
    console.log(`Authorization header present: ${authPreview}`);
  }

  next();
});

// Add special handling for OPTIONS requests
app.options('*', cors({
  origin: true,
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin']
}));

// Add direct route for emergency admin endpoints
app.use('/emergency', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true,
  onProxyReq: (proxyReq, req, res) => {
    console.log(`⚠️ Proxying EMERGENCY request: ${req.method} ${req.url}`);

    // Explicitly forward authorization header
    if (req.headers.authorization) {
      console.log('  Forwarding Authorization header for emergency endpoint');
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }

    // Add special debug header
    proxyReq.setHeader('X-Emergency-Access', 'true');
  },
  onProxyRes: (proxyRes, req, res) => {
    console.log(`Emergency endpoint response: ${proxyRes.statusCode}`);
  }
}));

// Add direct route for debug admin endpoints
app.use('/debug_admin_status', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true,
  onProxyReq: (proxyReq, req, res) => {
    console.log(`⚠️ Proxying direct debug_admin_status request`);

    // Add debug headers
    proxyReq.setHeader('X-Debug-Admin', 'true');

    // Forward auth header if present
    if (req.headers.authorization) {
      console.log('  Forwarding Authorization header for debug endpoint');
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    console.log(`Debug admin status response: ${proxyRes.statusCode}`);
  }
}));

// WebSocket proxy for /ws endpoints
app.use('/ws', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  ws: true, // Enable WebSocket proxy
  xfwd: true, // Forward original client IP and host headers
  pathRewrite: {
    '^/ws': '/ws' // Keep the /ws path
  },
  // Special handling for WebSocket upgrade
  onProxyReq: (proxyReq, req, res) => {
    // Log WebSocket connection attempt
    console.log(`Proxying WebSocket initial request to: ${backendUrl}/ws`);
    console.log(`  Origin: ${req.headers.origin || 'unknown'}`);

    // Forward authentication token from query params or headers
    const token = req.query.token;
    if (token) {
      console.log('  Authentication token present in query params');
      // Add token to request header if it's in the query parameters
      proxyReq.setHeader('Authorization', `Bearer ${token}`);
    }

    // Also forward any existing Authorization header
    const authHeader = req.headers.authorization;
    if (authHeader) {
      console.log('  Authorization header present and forwarded');
      proxyReq.setHeader('Authorization', authHeader);
    }
  },
  // Handle WebSocket specific errors
  onError: (err, req, res) => {
    console.error('WebSocket proxy error:', err);
    // Try to send error if headers not sent yet
    if (!res.headersSent) {
      res.status(502).json({
        error: 'WebSocket Proxy Error',
        message: err.message,
        code: 'WS_PROXY_ERROR'
      });
    }
  },
  // Log when WebSocket connection is established
  onProxyRes: (proxyRes, req, res) => {
    console.log(`WebSocket response status: ${proxyRes.statusCode}`);
  }
}));

// API proxy for all other requests
app.use('/api', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true, // Forward original client IP and host headers
  pathRewrite: {
    '^/api': '/api' // Keep the /api path
  },
  onProxyReq: (proxyReq, req, res) => {
    // Log each API request for debugging
    console.log(`Proxying API request: ${req.method} ${req.url}`);
    console.log(`  To: ${backendUrl}${req.url}`);
    console.log(`  From origin: ${req.headers.origin || 'unknown'}`);

    // Explicitly preserve and forward the Authorization header
    if (req.headers.authorization) {
      const authPreview = req.headers.authorization.substring(0, 15) + '...';
      console.log(`  Authorization header present: ${authPreview}`);
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }

    // Forward token from query parameter if present (for compatibility)
    const token = req.query.token;
    if (token && !req.headers.authorization) {
      console.log('  Adding Authorization header from query token');
      proxyReq.setHeader('Authorization', `Bearer ${token}`);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Ensure CORS headers are present in the response
    // Use the actual origin rather than wildcard to allow credentials
    const origin = req.headers.origin || '*';
    res.setHeader('Access-Control-Allow-Origin', origin);
    res.setHeader('Access-Control-Allow-Credentials', 'true');

    // For 401/403 errors, add more debug info in the log
    if (proxyRes.statusCode === 401 || proxyRes.statusCode === 403) {
      console.log(`⚠️ AUTH ERROR: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
      console.log(`  Auth header present: ${!!req.headers.authorization}`);
      console.log(`  Origin: ${req.headers.origin || 'unknown'}`);
    } else {
      // Log response status for debugging
      console.log(`API response status: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    }
  },
  onError: (err, req, res) => {
    console.error('API proxy error:', err);
    console.error(`  Failed request: ${req.method} ${req.url}`);

    if (!res.headersSent) {
      res.status(502).json({
        error: 'API Proxy Error',
        message: err.message,
        code: 'API_PROXY_ERROR',
        path: req.url,
        timestamp: new Date().toISOString()
      });
    }
  }
}));

// Handle unknown routes
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Not Found',
    message: 'Use /api/* to proxy API requests, /ws for WebSocket connections, or /emergency/* for direct admin access',
    available_endpoints: ['/api/*', '/ws', '/emergency/*', '/debug_admin_status', '/debug/headers', '/health'],
    timestamp: new Date().toISOString()
  });
});

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy running on port ${port}`);
  console.log(`Proxying to backend: ${backendUrl}`);
});