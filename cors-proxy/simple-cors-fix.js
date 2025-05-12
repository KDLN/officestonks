const express = require('express');
const cors = require('cors');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const port = process.env.PORT || 3000;

// Define target backend URL
const backendUrl = process.env.BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Frontend URL that needs CORS access
const frontendUrl = 'https://officestonks-frontend-production.up.railway.app';

// Very permissive CORS setup to fix issues
app.use(function(req, res, next) {
  // Set CORS headers directly on all responses
  res.header('Access-Control-Allow-Origin', req.headers.origin || frontendUrl);
  res.header('Access-Control-Allow-Credentials', 'true');
  res.header('Access-Control-Allow-Methods', 'GET,HEAD,PUT,PATCH,POST,DELETE,OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With, Accept, Origin');
  
  // Handle preflight OPTIONS requests
  if (req.method === 'OPTIONS') {
    console.log(`⚠️ OPTIONS request for: ${req.path} from ${req.headers.origin || 'unknown'}`);
    return res.status(204).send();
  }
  
  next();
});

// Additional CORS middleware for extra safety
app.use(cors({
  origin: function(origin, callback) {
    // Allow the frontend URL and undefined origins (like curl)
    callback(null, true);
  },
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin']
}));

// Debug endpoint
app.get('/cors-debug', (req, res) => {
  console.log('Debug headers:', req.headers);
  res.json({
    headers: req.headers,
    message: 'CORS headers check',
    cors_enabled: true,
    origin: req.headers.origin || 'unknown',
    timestamp: new Date().toISOString()
  });
});

// Health check
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    service: 'CORS Proxy (Simple Fix)',
    version: '1.2.0',
    frontend_url: frontendUrl,
    backend_url: backendUrl,
    timestamp: new Date().toISOString()
  });
});

// Create a proxy middleware for all requests
const proxyOptions = {
  target: backendUrl,
  changeOrigin: true,
  onProxyReq: (proxyReq, req, res) => {
    // Log request
    console.log(`Proxying ${req.method} ${req.url}`);
    
    // Forward auth headers
    if (req.headers.authorization) {
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Ensure CORS headers are present in response
    res.setHeader('Access-Control-Allow-Origin', req.headers.origin || frontendUrl);
    res.setHeader('Access-Control-Allow-Credentials', 'true');
    
    // Log response
    console.log(`Response: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    
    // Special logging for auth errors
    if (proxyRes.statusCode === 401 || proxyRes.statusCode === 403) {
      console.log(`⚠️ AUTH ERROR for ${req.url}`);
    }
  },
  onError: (err, req, res) => {
    console.error('Proxy error:', err);
    
    if (!res.headersSent) {
      // Set CORS headers even on error
      res.setHeader('Access-Control-Allow-Origin', req.headers.origin || frontendUrl);
      res.setHeader('Access-Control-Allow-Credentials', 'true');
      
      res.status(500).json({
        error: 'Proxy Error',
        message: err.message,
        path: req.url,
        timestamp: new Date().toISOString()
      });
    }
  }
};

// Apply proxy to all routes
app.use('/', createProxyMiddleware(proxyOptions));

// Start the server
app.listen(port, () => {
  console.log(`Simple CORS Fix proxy running on port ${port}`);
  console.log(`Proxying to backend: ${backendUrl}`);
  console.log(`Allowing frontend: ${frontendUrl}`);
});