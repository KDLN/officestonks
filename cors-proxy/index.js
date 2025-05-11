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
const wsBackendUrl = backendUrl.replace(/^https?:\/\//, '');

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
  onProxyReq: (proxyReq, req, res) => {
    // Add any request modifications here if needed
    console.log(`Proxying WebSocket request to: ${backendUrl}/ws`);
  },
  onError: (err, req, res) => {
    console.error('Proxy error:', err);
    res.status(502).json({ error: 'Proxy error', message: err.message });
  }
}));

// API proxy for all other requests
app.use('/api', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  onProxyReq: (proxyReq, req, res) => {
    // Add any request modifications here if needed
    console.log(`Proxying API request to: ${backendUrl}${req.url}`);
  },
  onError: (err, req, res) => {
    console.error('Proxy error:', err);
    res.status(502).json({ error: 'Proxy error', message: err.message });
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