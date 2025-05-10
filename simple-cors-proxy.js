const express = require('express');
const cors = require('cors');
const { createProxyMiddleware } = require('http-proxy-middleware');

// Create the express app
const app = express();
const port = process.env.PORT || 3001;

// Basic request logging
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

// Always respond to OPTIONS requests immediately before any other middleware
app.options('*', (req, res) => {
  // Set CORS headers
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  res.header('Access-Control-Max-Age', '86400'); // 24 hours

  console.log('Responding to OPTIONS request with 200 OK');

  // Respond with 200 OK
  res.status(200).end();
});

// Simple CORS configuration for non-OPTIONS requests
app.use(cors({
  origin: '*', // Allow all origins
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization'],
  optionsSuccessStatus: 200
}));

// Basic middlewares
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: 'Simple CORS proxy is running' });
});

// CORS test endpoint
app.get('/cors-test', (req, res) => {
  res.json({
    message: 'CORS test endpoint',
    origin: req.headers.origin || 'No origin provided'
  });
});

// Backend API target URL
const BACKEND_URL = process.env.API_BASE_URL || 'https://web-production-1e26.up.railway.app';

// Simple proxy for admin endpoints
app.use('/admin', createProxyMiddleware({
  target: BACKEND_URL,
  changeOrigin: true,
  pathRewrite: {
    '^/admin': '/api/admin'
  },
  onProxyReq: (proxyReq, req, res) => {
    // Copy token from query parameter to authorization header
    if (req.query.token && !req.headers.authorization) {
      proxyReq.setHeader('Authorization', `Bearer ${req.query.token}`);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    console.log(`Proxy response: ${proxyRes.statusCode} ${req.method} ${req.url}`);

    // Set CORS headers in lowercase (important for some browsers)
    proxyRes.headers['access-control-allow-origin'] = '*';
    proxyRes.headers['access-control-allow-methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
    proxyRes.headers['access-control-allow-headers'] = 'Content-Type, Authorization';

    // Also set in standard case for compatibility
    proxyRes.headers['Access-Control-Allow-Origin'] = '*';
    proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
    proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization';

    // Add a custom header for debugging
    proxyRes.headers['x-cors-proxy-version'] = '1.0-simple';

    console.log('Proxy response headers:', proxyRes.headers);
  },
  // Handle proxy errors
  onError: (err, req, res) => {
    console.error('Proxy error:', err);

    // Set CORS headers
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');

    // Send error response
    res.status(500).json({ error: 'Proxy error', message: err.message });
  }
}));

// Start server
app.listen(port, () => {
  console.log(`Simple CORS proxy server listening on port ${port}`);
});