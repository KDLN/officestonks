// Ultra simple CORS proxy with Express
const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

// Add CORS headers to all requests
app.use((req, res, next) => {
  // Set wildcard CORS headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', '*');
  
  // Respond immediately to OPTIONS requests
  if (req.method === 'OPTIONS') {
    console.log('OPTIONS request received - responding with 200');
    return res.status(200).end();
  }
  
  next();
});

// Forward requests to the API
app.all('*', async (req, res) => {
  try {
    console.log(`Proxying request to: ${req.url}`);
    
    // Build the target URL
    const apiUrl = 'https://web-production-1e26.up.railway.app';
    let targetPath = req.url;
    
    // Adjust path for admin routes
    if (targetPath.startsWith('/admin')) {
      targetPath = `/api${targetPath}`;
    }
    
    const targetUrl = `${apiUrl}${targetPath}`;
    console.log(`Forwarding to: ${targetUrl}`);
    
    // Get token from query string if present
    const urlObj = new URL(req.url, `http://${req.headers.host}`);
    const token = urlObj.searchParams.get('token');
    
    // Prepare headers
    const headers = { ...req.headers };
    delete headers.host;
    
    // Add Authorization header if token is present
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
      console.log('Added Authorization header with token');
    }
    
    // Forward the request
    const response = await fetch(targetUrl, {
      method: req.method,
      headers: headers,
      body: req.method !== 'GET' && req.method !== 'HEAD' ? JSON.stringify(req.body) : undefined
    });
    
    // Copy status code
    res.status(response.status);
    
    // Copy all headers except content-encoding and content-length
    for (const [key, value] of response.headers.entries()) {
      if (key !== 'content-encoding' && key !== 'content-length') {
        res.setHeader(key, value);
      }
    }
    
    // Force CORS headers
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', '*');
    
    // Send the response
    const data = await response.text();
    res.send(data);
    
  } catch (error) {
    console.error('Proxy error:', error);
    res.status(500).json({ error: 'Proxy server error', message: error.message });
  }
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: 'Ultra simple CORS proxy is running' });
});

// Start server
app.listen(port, () => {
  console.log(`Ultra simple CORS proxy running on port ${port}`);
});