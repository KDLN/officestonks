const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');
const path = require('path');

// Create Express server
const app = express();
const port = process.env.PORT || 3000;

// Get environment variables or use defaults
const BACKEND_URL = process.env.BACKEND_URL || 'https://web-copy-production-5b48.up.railway.app';
const FRONTEND_URL = process.env.FRONTEND_URL || 'https://officestonks-frontend.vercel.app';

// Enable CORS
app.use(cors());

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    message: 'Reverse proxy is running',
    backend: BACKEND_URL,
    frontend: FRONTEND_URL
  });
});

// Proxy API requests to the backend
app.use('/api', createProxyMiddleware({
  target: BACKEND_URL,
  changeOrigin: true,
  pathRewrite: {
    '^/api': '/api' // Keep the /api prefix
  },
  onProxyReq: (proxyReq, req, res) => {
    // Log the proxy request
    console.log(`Proxying API request: ${req.method} ${req.url} to ${BACKEND_URL}${proxyReq.path}`);
  }
}));

// Proxy WebSocket requests to the backend
app.use('/ws', createProxyMiddleware({
  target: BACKEND_URL.replace(/^http/, 'ws'),
  ws: true,
  changeOrigin: true
}));

// Proxy all other requests to the frontend
app.use('/', createProxyMiddleware({
  target: FRONTEND_URL,
  changeOrigin: true,
  onProxyReq: (proxyReq, req, res) => {
    // Log the proxy request
    console.log(`Proxying frontend request: ${req.method} ${req.url} to ${FRONTEND_URL}${proxyReq.path}`);
  }
}));

// Start the server
app.listen(port, () => {
  console.log(`Reverse proxy server listening on port ${port}`);
  console.log(`Routing /api/* to ${BACKEND_URL}`);
  console.log(`Routing all other requests to ${FRONTEND_URL}`);
});