const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');

// Create Express server
const app = express();
const port = process.env.PORT || 3001;

// Enable CORS for all routes
app.use(cors({
  origin: '*', // Allow all origins
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization'],
  credentials: true
}));

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: 'CORS proxy server is running' });
});

// Proxy admin API requests to the backend
app.use('/admin', createProxyMiddleware({
  target: 'https://web-production-1e26.up.railway.app/api',
  changeOrigin: true,
  pathRewrite: {
    '^/admin': '/admin' // Keep the /admin path
  },
  onProxyReq: (proxyReq, req, res) => {
    // Log the proxy request
    console.log(`Proxying ADMIN ${req.method} ${req.url} to ${proxyReq.path}`);

    // Copy token from query parameter to authorization header if needed
    if (req.query.token && !req.headers.authorization) {
      proxyReq.setHeader('Authorization', `Bearer ${req.query.token}`);
      console.log('Added token from query parameters to authorization header');
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Log the proxy response status
    console.log(`Proxy ADMIN response: ${proxyRes.statusCode}`);

    // Ensure CORS headers on the response
    proxyRes.headers['Access-Control-Allow-Origin'] = '*';
    proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
    proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization';
    proxyRes.headers['Access-Control-Allow-Credentials'] = 'true';
  }
}));

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy server listening on port ${port}`);
});