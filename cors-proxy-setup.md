# CORS Proxy Setup for Admin API

To fix the CORS issues with admin API endpoints, we need to deploy a CORS proxy service. This proxy specifically allows requests from the production frontend domain `https://officestonks-frontend-production.up.railway.app` and sets appropriate CORS headers on all responses. Here's how to set it up:

## Step 1: Create the proxy files

Create a file called `cors-proxy.js` with the following content:

```javascript
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');

// Create Express server
const app = express();
const port = process.env.PORT || 3001;

// Enable CORS specifically for the production frontend
app.use(cors({
  origin: 'https://officestonks-frontend-production.up.railway.app', // Only allow the production frontend
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

// Handle OPTIONS preflight requests explicitly
app.options('*', (req, res) => {
  console.log(`Handling OPTIONS preflight request for: ${req.path}`);

  // Set CORS headers
  res.header('Access-Control-Allow-Origin', 'https://officestonks-frontend-production.up.railway.app');
  res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
  res.header('Access-Control-Allow-Credentials', 'true');

  // Respond with 200 OK
  res.status(200).send();
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
    if (req.query.token && \!req.headers.authorization) {
      proxyReq.setHeader('Authorization', `Bearer ${req.query.token}`);
      console.log('Added token from query parameters to authorization header');
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Log the proxy response status
    console.log(`Proxy ADMIN response: ${proxyRes.statusCode}`);
    
    // Ensure CORS headers on the response
    proxyRes.headers['Access-Control-Allow-Origin'] = 'https://officestonks-frontend-production.up.railway.app';
    proxyRes.headers['Access-Control-Allow-Methods'] = 'GET, POST, PUT, DELETE, OPTIONS';
    proxyRes.headers['Access-Control-Allow-Headers'] = 'Content-Type, Authorization';
    proxyRes.headers['Access-Control-Allow-Credentials'] = 'true';
  }
}));

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy server listening on port ${port}`);
});
```

Create a file called `package.json` for the proxy:

```json
{
  "name": "officestonks-cors-proxy",
  "version": "1.0.0",
  "description": "CORS proxy for OfficeStonks API",
  "main": "cors-proxy.js",
  "scripts": {
    "start": "node cors-proxy.js"
  },
  "dependencies": {
    "cors": "^2.8.5",
    "express": "^4.18.2",
    "http-proxy-middleware": "^2.0.6"
  }
}
```

## Step 2: Deploy to Railway

1. Create a new Railway project
2. Add these files to the project
3. Deploy the service
4. Note the URL of the deployed proxy service

## Step 3: Update Frontend Code

Apply the patch to your admin.js file to point to the proxy service.

## Step 4: Deploy Frontend

Deploy the updated frontend code to make the admin panel work with the proxy.
EOL < /dev/null
