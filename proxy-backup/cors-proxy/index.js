const express = require('express');
const cors = require('cors');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const port = process.env.PORT || 3000;

// Enable CORS for all requests
app.use(cors({
  origin: '*',
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin']
}));

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    service: 'CORS Proxy',
    timestamp: new Date().toISOString()
  });
});

// Configure proxy middleware
const backendUrl = process.env.BACKEND_URL || 'https://web-production-1e26.up.railway.app';
console.log(`Proxy configured to forward requests to: ${backendUrl}`);

// Log all requests
app.use((req, res, next) => {
  console.log(`${req.method} ${req.url} from ${req.headers.origin || 'unknown'}`);
  next();
});

// WebSocket proxy for /ws endpoints
app.use('/ws', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  ws: true, // Enable WebSocket proxy
  pathRewrite: {
    '^/ws': '/ws' // Keep the /ws path
  },
  // Special handling for WebSocket upgrade
  onProxyReq: (proxyReq, req, res) => {
    // Log WebSocket connection attempt
    console.log(`Proxying WebSocket initial request to: ${backendUrl}/ws`);
    console.log(`  Origin: ${req.headers.origin || 'unknown'}`);

    // Copy authentication token if present
    const token = req.query.token;
    if (token) {
      console.log('  Authentication token present');
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
  pathRewrite: {
    '^/api': '/api' // Keep the /api path
  },
  onProxyReq: (proxyReq, req, res) => {
    // Log each API request for debugging
    console.log(`Proxying API request: ${req.method} ${req.url}`);
    console.log(`  To: ${backendUrl}${req.url}`);
    console.log(`  From origin: ${req.headers.origin || 'unknown'}`);

    // Preserve original headers
    if (req.headers.authorization) {
      console.log('  Authorization header present');
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Ensure CORS headers are present in the response
    res.setHeader('Access-Control-Allow-Origin', req.headers.origin || '*');
    res.setHeader('Access-Control-Allow-Credentials', 'true');

    // Log response status for debugging
    console.log(`API response status: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
  },
  onError: (err, req, res) => {
    console.error('API proxy error:', err);
    console.error(`  Failed request: ${req.method} ${req.url}`);

    if (!res.headersSent) {
      res.status(502).json({
        error: 'API Proxy Error',
        message: err.message,
        code: 'API_PROXY_ERROR',
        path: req.url
      });
    }
  }
}));

// Handle unknown routes
app.use('*', (req, res) => {
  res.status(404).json({ 
    error: 'Not Found',
    message: 'Use /api/* to proxy API requests or /ws for WebSocket connections' 
  });
});

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy running on port ${port}`);
  console.log(`Proxying to backend: ${backendUrl}`);
});