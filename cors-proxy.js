const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');
const morgan = require('morgan');

// Create Express server
const app = express();
const port = process.env.PORT || 3001;

// Log startup information
console.log(`[STARTUP] Starting CORS proxy server on port ${port}`);
console.log(`[STARTUP] Node environment: ${process.env.NODE_ENV || 'development'}`);
console.log(`[STARTUP] Time: ${new Date().toISOString()}`);

// Add detailed request logging
app.use(morgan(':date[iso] :method :url :status :response-time ms - :res[content-length]'));

// Global error handler to prevent crashes
app.use((err, req, res, next) => {
  console.error('[ERROR] Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

// Set CORS headers for all responses (including preflight)
app.use((req, res, next) => {
  // Log detailed request info
  console.log(`[DEBUG] Request received: ${req.method} ${req.url}`);
  console.log(`[DEBUG] Request headers: ${JSON.stringify(req.headers)}`);
  console.log(`[DEBUG] Origin: ${req.headers.origin || 'Not specified'}`);

  // Always set CORS headers first - allow all origins
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With');
  res.setHeader('Access-Control-Max-Age', '86400'); // 24 hours

  // Debug the response headers
  console.log(`[DEBUG] Response headers set: ${JSON.stringify(res.getHeaders ? res.getHeaders() : 'Not available')}`);

  // Handle OPTIONS method immediately
  if (req.method === 'OPTIONS') {
    console.log(`[DEBUG] Handling OPTIONS preflight for: ${req.url}`);
    // Send 200 OK with headers already set
    console.log(`[DEBUG] Sending 200 OK for OPTIONS`);
    return res.status(200).end();
  }

  next();
});

// Additional CORS middleware for extra safety - enabled for all routes
app.use(cors({
  origin: '*', // Allow all origins
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With'],
  preflightContinue: false,
  optionsSuccessStatus: 200
}));

// Set a global preflight OPTIONS handler for all routes
app.options('*', cors(), (req, res) => {
  console.log(`[DEBUG] Global OPTIONS handler for: ${req.url}`);
  res.status(200).end();
});

// Add body parsing middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: 'CORS proxy server is running' });
});

// CORS test endpoint
app.get('/cors-test', (req, res) => {
  console.log('[TEST] Received request to /cors-test from origin:', req.headers.origin);

  // This will log the request headers for debugging
  const headers = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS, PATCH',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization, X-Requested-With'
    // Can't use credentials with wildcard origin
  };

  // Log what headers we're sending
  console.log('[TEST] Sending CORS headers:', headers);

  // Send response with CORS info
  res.json({
    message: 'CORS test endpoint',
    received_headers: req.headers,
    cors_status: 'Headers set correctly - allowing all origins',
    allowed_origin: '*',
    server_time: new Date().toISOString(),
    version: '2.0'
  });
});

// Admin test endpoint - for verifying admin routes are working
app.get('/admin-test', (req, res) => {
  console.log('[TEST] Received request to /admin-test from origin:', req.headers.origin);

  res.json({
    message: 'Admin API test endpoint',
    status: 'success',
    info: 'If you can see this response, direct admin requests are working',
    note: 'This is served directly from the proxy, not the backend'
  });
});

// OPTIONS requests are now handled by the early CORS middleware above

// We need to get these before the proxy setup
const API_BASE_URL = process.env.API_BASE_URL || 'https://web-production-1e26.up.railway.app';
console.log(`[CONFIG] Using API base URL: ${API_BASE_URL}`);

// Proxy admin API requests to the backend
app.use('/admin', createProxyMiddleware({
  target: `${API_BASE_URL}`,
  changeOrigin: true,
  pathRewrite: {
    '^/admin': '/api/admin' // Make sure we point to /api/admin on the backend
  },
  // Enable all HTTP methods to be proxied
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  // Correctly handle preflight requests
  onProxyReq: (proxyReq, req, res) => {
    // Log the proxy request with detailed info
    console.log(`Proxying ADMIN ${req.method} ${req.url} to ${proxyReq.path}`);
    console.log(`Request headers: ${JSON.stringify(req.headers)}`);

    // Always adjust content-type header for proper JSON handling
    if (req.method !== 'OPTIONS' && !proxyReq.getHeader('content-type') && req.body) {
      proxyReq.setHeader('content-type', 'application/json');
    }

    // Copy token from query parameter to authorization header if needed
    if (req.query.token && !req.headers.authorization) {
      proxyReq.setHeader('Authorization', `Bearer ${req.query.token}`);
      console.log('Added token from query parameters to authorization header');
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Log the proxy response status with more details
    console.log(`[PROXY] Response: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    console.log(`[PROXY] Response headers before: ${JSON.stringify(proxyRes.headers)}`);

    // Always ensure CORS headers on every response
    // These must match exactly what's in the preflight response
    proxyRes.headers['access-control-allow-origin'] = '*';
    proxyRes.headers['access-control-allow-methods'] = 'GET, POST, PUT, DELETE, OPTIONS, PATCH';
    proxyRes.headers['access-control-allow-headers'] = 'Content-Type, Authorization, X-Requested-With';
    proxyRes.headers['access-control-max-age'] = '86400';

    // Handle 404s with better error messaging
    if (proxyRes.statusCode === 404) {
      console.log(`[ERROR] 404 for ${req.url} - endpoint may not exist on the backend`);
    }

    // Add debug headers to help identify proxy responses
    proxyRes.headers['x-proxy-time'] = new Date().toISOString();
    proxyRes.headers['x-proxy-version'] = '2.0'; // Increment this when making changes
    proxyRes.headers['x-proxy-debug'] = 'true';

    // Log the entire response for debugging
    console.log(`[PROXY] Final headers: ${JSON.stringify(proxyRes.headers)}`);
  },
  // Add error handling for the proxy
  onError: (err, req, res) => {
    console.error(`[PROXY ERROR] ${err.message}`);
    console.error(err.stack);

    // Send a response with CORS headers
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With');

    res.status(500).json({
      error: 'Proxy error',
      message: err.message,
      timestamp: new Date().toISOString()
    });
  }
}));

// Verify package.json dependency on morgan is added
try {
  require('morgan');
} catch (e) {
  console.warn('[WARNING] Morgan package not found. Logging may be limited.');
  console.warn('[WARNING] Run: npm install --save morgan');
}

// Start the server
app.listen(port, () => {
  console.log(`[STARTUP] CORS Proxy server listening on port ${port}`);
  console.log(`[STARTUP] Proxy configured with API base URL: ${API_BASE_URL}`);
  console.log(`[STARTUP] Test endpoints available:`);
  console.log(`[STARTUP]   - GET /health        (Server health check)`);
  console.log(`[STARTUP]   - GET /cors-test     (CORS headers test)`);
  console.log(`[STARTUP]   - GET /admin-test    (Admin route test)`);
  console.log(`[STARTUP]   - GET /admin/users   (Proxied admin users endpoint)`);
});