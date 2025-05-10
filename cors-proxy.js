const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');

// Create Express server
const app = express();
const port = process.env.PORT || 3001;

// Set CORS headers for all responses (including preflight)
app.use((req, res, next) => {
  // Always set CORS headers first - allow all origins
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  // When using wildcard origin, we can't use credentials
  // res.header('Access-Control-Allow-Credentials', 'true');

  console.log(`CORS headers set for ${req.method} ${req.url} from origin: ${req.headers.origin}`);

  // Handle OPTIONS method immediately
  if (req.method === 'OPTIONS') {
    console.log('Responding to OPTIONS preflight request immediately');
    return res.status(200).send();
  }

  next();
});

// Additional CORS middleware for extra safety
app.use(cors({
  origin: '*', // Allow all origins
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization']
  // credentials: true - can't use with wildcard origin
}));

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
  console.log('Received request to /cors-test from origin:', req.headers.origin);

  // This will log the request headers for debugging
  const headers = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    // Can't use credentials with wildcard origin
    // 'Access-Control-Allow-Credentials': 'true'
  };

  // Log what headers we're sending
  console.log('Sending CORS headers:', headers);

  // Send response with CORS info
  res.json({
    message: 'CORS test endpoint',
    received_headers: req.headers,
    cors_status: 'Headers set correctly - allowing all origins',
    allowed_origin: '*'
  });
});

// OPTIONS requests are now handled by the early CORS middleware above

// Proxy admin API requests to the backend
app.use('/admin', createProxyMiddleware({
  target: 'https://web-production-1e26.up.railway.app/api',
  changeOrigin: true,
  pathRewrite: {
    '^/admin': '/admin' // Keep the /admin path
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
    console.log(`Proxy ADMIN response: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    console.log(`Response headers: ${JSON.stringify(proxyRes.headers)}`);

    // Double ensure CORS headers on every response
    proxyRes.headers['Access-Control-Allow-Origin'] = '*';
    proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
    proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization';
    // Can't use credentials with wildcard origin
    // proxyRes.headers['Access-Control-Allow-Credentials'] = 'true';

    // Add debug headers to help identify proxy responses
    proxyRes.headers['X-Proxy-Time'] = new Date().toISOString();
    proxyRes.headers['X-Proxy-Version'] = '1.1'; // Increment this when making changes

    // Log the entire response for debugging
    console.log(`Full proxy response: Status=${proxyRes.statusCode}, Headers=${JSON.stringify(proxyRes.headers)}`)
  }
}));

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy server listening on port ${port}`);
});