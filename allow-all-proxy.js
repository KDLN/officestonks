// Allow ALL CORS proxy - absolutely no restrictions
import express from 'express';
import fetch from 'node-fetch';

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
app.all('*', async (req, res) => {
  try {
    // Construct target URL
    let targetPath = req.url;
    
    // Add /api prefix for admin endpoints
    if (targetPath.startsWith('/admin')) {
      targetPath = `/api${targetPath}`;
    }
    
    const targetUrl = `${API_URL}${targetPath}`;
    console.log(`Proxying ${req.method} ${req.url} to ${targetUrl}`);
    
    // Get token from query params
    const urlObj = new URL(req.url, `http://${req.headers.host}`);
    const token = urlObj.searchParams.get('token');
    
    // Copy request headers
    const headers = {};
    Object.entries(req.headers).forEach(([key, value]) => {
      if (key !== 'host') {
        headers[key] = value;
      }
    });
    
    // Add authorization header if token is present
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
      console.log('Added token to Authorization header');
    }
    
    // Forward the request
    const fetchOptions = {
      method: req.method,
      headers: headers,
    };
    
    // Add body for non-GET requests
    if (req.method !== 'GET' && req.method !== 'HEAD') {
      fetchOptions.body = JSON.stringify(req.body);
    }
    
    const response = await fetch(targetUrl, fetchOptions);
    
    // Copy status
    res.status(response.status);
    
    // Copy headers (except those that cause issues)
    for (const [key, value] of Object.entries(response.headers.raw())) {
      if (key.toLowerCase() !== 'content-length' && 
          key.toLowerCase() !== 'transfer-encoding') {
        res.set(key, value);
      }
    }
    
    // Set CORS headers again to be 100% sure
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', '*');
    res.header('Access-Control-Allow-Headers', '*');
    
    // Get response content
    const responseBody = await response.text();
    
    // Send response back to client
    res.send(responseBody);
    
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