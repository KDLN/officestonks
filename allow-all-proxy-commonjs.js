// Allow ALL CORS proxy - absolutely no restrictions (CommonJS version)
const express = require('express');
const http = require('http');
const https = require('https');

const app = express();

// Port configuration
const port = process.env.PORT || 3000;

// Backend API URL
const API_URL = process.env.API_URL || 'https://web-production-1e26.up.railway.app';

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

// Add CORS headers to every single response
app.use((req, res, next) => {
  // Set wildcard CORS headers - allow EVERYTHING
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Methods', '*');
  res.header('Access-Control-Allow-Headers', '*');
  res.header('Access-Control-Max-Age', '86400'); // 24 hours
  
  // For OPTIONS method, immediately return 200 OK with CORS headers
  if (req.method === 'OPTIONS') {
    console.log('OPTIONS request - returning 200 OK immediately');
    return res.status(200).end();
  }
  
  next();
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: 'Allow-all CORS proxy is running' });
});

// Proxy all requests to the backend API
app.all('*', (req, res) => {
  try {
    // Construct target URL
    let targetPath = req.url;
    
    // Add /api prefix for admin endpoints
    if (targetPath.startsWith('/admin')) {
      targetPath = `/api${targetPath}`;
    }
    
    const targetUrl = `${API_URL}${targetPath}`;
    console.log(`Proxying ${req.method} ${req.url} to ${targetUrl}`);
    
    // Parse target URL
    const targetUrlObj = new URL(targetUrl);
    
    // Get token from query params
    const urlObj = new URL(req.url, `http://${req.headers.host}`);
    const token = urlObj.searchParams.get('token');
    
    // Prepare request options
    const options = {
      hostname: targetUrlObj.hostname,
      port: targetUrlObj.port || (targetUrlObj.protocol === 'https:' ? 443 : 80),
      path: targetUrlObj.pathname + targetUrlObj.search,
      method: req.method,
      headers: {}
    };
    
    // Copy request headers (except host)
    for (const [key, value] of Object.entries(req.headers)) {
      if (key.toLowerCase() !== 'host') {
        options.headers[key] = value;
      }
    }
    
    // Add authorization header if token is present
    if (token) {
      options.headers['Authorization'] = `Bearer ${token}`;
      console.log('Added token to Authorization header');
    }
    
    // Choose http or https module
    const requestModule = targetUrlObj.protocol === 'https:' ? https : http;
    
    // Forward the request
    const proxyReq = requestModule.request(options, (proxyRes) => {
      // Copy status code
      res.status(proxyRes.statusCode);
      
      // Copy headers (except those that cause issues)
      for (const [key, value] of Object.entries(proxyRes.headers)) {
        if (key.toLowerCase() !== 'content-length' && 
            key.toLowerCase() !== 'transfer-encoding') {
          res.set(key, value);
        }
      }
      
      // Set CORS headers again to be 100% sure
      res.header('Access-Control-Allow-Origin', '*');
      res.header('Access-Control-Allow-Methods', '*');
      res.header('Access-Control-Allow-Headers', '*');
      
      // Pipe response to client
      proxyRes.pipe(res);
    });
    
    // Handle errors
    proxyReq.on('error', (error) => {
      console.error('Proxy request error:', error);
      
      // Set CORS headers even for errors
      res.header('Access-Control-Allow-Origin', '*');
      res.header('Access-Control-Allow-Methods', '*');
      res.header('Access-Control-Allow-Headers', '*');
      
      res.status(500).json({
        error: 'Proxy error',
        message: error.message
      });
    });
    
    // End the request
    if (req.method !== 'GET' && req.method !== 'HEAD' && req.body) {
      proxyReq.write(JSON.stringify(req.body));
    }
    proxyReq.end();
    
  } catch (error) {
    console.error('Proxy error:', error);
    
    // Set CORS headers even for errors
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', '*');
    res.header('Access-Control-Allow-Headers', '*');
    
    res.status(500).json({
      error: 'Proxy error',
      message: error.message
    });
  }
});

// Start the server
app.listen(port, () => {
  console.log(`Allow-all CORS proxy running on port ${port}`);
  console.log(`Proxying to backend API: ${API_URL}`);
});