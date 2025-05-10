const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

// Create express app
const app = express();
const port = process.env.PORT || 3001;

// CORS wildcard - allow everything
app.use((req, res, next) => {
  // Set wide open CORS headers - allow everything from anywhere
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', '*');
  res.header('Access-Control-Allow-Headers', '*');
  
  // Handle OPTIONS immediately
  if (req.method === 'OPTIONS') {
    return res.status(200).end();
  }
  
  next();
});

// Backend API URL
const API_URL = process.env.API_URL || 'https://web-production-1e26.up.railway.app';

// Proxy everything to backend
app.use('/', createProxyMiddleware({
  target: API_URL,
  changeOrigin: true,
  pathRewrite: (path) => {
    if (path.startsWith('/admin')) {
      return `/api${path}`;
    }
    return path;
  },
  onProxyReq: (proxyReq, req, res) => {
    // Copy token from query param to header
    if (req.query.token) {
      proxyReq.setHeader('Authorization', `Bearer ${req.query.token}`);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Force CORS headers on all responses
    proxyRes.headers['access-control-allow-origin'] = '*';
    proxyRes.headers['access-control-allow-methods'] = '*';
    proxyRes.headers['access-control-allow-headers'] = '*';
  }
}));

// Start server
app.listen(port, () => {
  console.log(`No-CORS Proxy running on port ${port}`);
  console.log(`Proxying all requests to ${API_URL}`);
});