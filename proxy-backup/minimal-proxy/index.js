const express = require('express');
const http = require('http');
const https = require('https');

const app = express();
const port = process.env.PORT || 3000;
const API_URL = process.env.API_URL || 'https://web-production-1e26.up.railway.app';

// Set CORS headers for all responses
app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', '*');
  res.header('Access-Control-Allow-Headers', '*');
  
  // Respond immediately to OPTIONS requests
  if (req.method === 'OPTIONS') {
    return res.status(200).end();
  }
  
  next();
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

// Proxy all other requests
app.all('*', (req, res) => {
  // Prepare URL with /api prefix for admin routes
  let targetPath = req.url;
  if (targetPath.startsWith('/admin')) {
    targetPath = `/api${targetPath}`;
  }
  
  const targetUrl = new URL(targetPath, API_URL);
  
  // Get token from query params
  const token = new URL(req.url, `http://${req.headers.host}`).searchParams.get('token');
  
  // Set up request options
  const options = {
    method: req.method,
    headers: { ...req.headers, host: targetUrl.host }
  };
  
  // Add Authorization header if token is present
  if (token) {
    options.headers.authorization = `Bearer ${token}`;
  }
  
  // Make the request
  const proxyReq = (targetUrl.protocol === 'https:' ? https : http).request(
    targetUrl,
    options,
    (proxyRes) => {
      // Set status code
      res.status(proxyRes.statusCode);
      
      // Copy headers (except problematic ones)
      Object.entries(proxyRes.headers).forEach(([key, value]) => {
        if (key !== 'content-length' && key !== 'transfer-encoding') {
          res.setHeader(key, value);
        }
      });
      
      // Ensure CORS headers
      res.header('Access-Control-Allow-Origin', '*');
      
      // Pipe the response
      proxyRes.pipe(res);
    }
  );
  
  // Handle errors
  proxyReq.on('error', (error) => {
    console.error(error);
    res.status(500).send('Proxy Error');
  });
  
  // Forward request body for non-GET requests
  if (req.method !== 'GET' && req.method !== 'HEAD' && req.body) {
    proxyReq.write(JSON.stringify(req.body));
  }
  
  proxyReq.end();
});

// Start server
app.listen(port, () => {
  console.log(`Proxy server running on port ${port}`);
});