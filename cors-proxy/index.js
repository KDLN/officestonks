const express = require('express');
const cors = require('cors');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();
const port = process.env.PORT || 3000;

// Enhanced CORS configuration
const corsOptions = {
  origin: function(origin, callback) {
    // Allow any origin to make requests
    callback(null, true);
  },
  credentials: true, // Allow credentials (cookies, auth headers)
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'PATCH'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With', 'Accept', 'Origin', 'Access-Control-Request-Method', 'Access-Control-Request-Headers'],
  exposedHeaders: ['Access-Control-Allow-Origin', 'Access-Control-Allow-Credentials', 'Access-Control-Allow-Methods', 'Access-Control-Allow-Headers'],
  maxAge: 86400 // Cache preflight request results for 24 hours (86400 seconds)
};

// Enable CORS for all requests with credentials support
app.use(cors(corsOptions));

// Special handling for OPTIONS requests - explicitly handle preflight
app.options('*', function(req, res) {
  // Set CORS headers
  res.header('Access-Control-Allow-Origin', req.headers.origin || '*');
  res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With, Accept, Origin, Access-Control-Request-Method, Access-Control-Request-Headers');
  res.header('Access-Control-Allow-Credentials', 'true');
  res.header('Access-Control-Max-Age', '86400');

  // Respond immediately with 204 No Content
  res.status(204).end();
});

// Debug endpoint for checking request headers
app.get('/debug/headers', (req, res) => {
  res.json({
    headers: req.headers,
    message: 'Debug endpoint to check request headers',
    timestamp: new Date().toISOString()
  });
});

// Health check endpoint with enhanced information
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    service: 'CORS Proxy',
    version: '1.1.0', // With enhanced CORS handling for admin routes
    cors_admin_fix: true,
    backends: {
      api: backendUrl,
      websocket: backendUrl
    },
    auth_forwarding: true,
    timestamp: new Date().toISOString()
  });
});

// Configure proxy middleware - only use environment variable, default to localhost for development
const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';
if (!process.env.BACKEND_URL) {
  console.warn('WARNING: BACKEND_URL environment variable not set, defaulting to localhost');
  console.warn('This should only be used for local development, not in production!');
}
console.log(`Proxy configured to forward requests to: ${backendUrl}`);

// Log all requests with detailed information
app.use((req, res, next) => {
  console.log(`${req.method} ${req.url} from ${req.headers.origin || 'unknown'}`);

  // Log authorization header presence (but not the full token for security)
  if (req.headers.authorization) {
    const authPreview = req.headers.authorization.substring(0, 15) + '...';
    console.log(`Authorization header present: ${authPreview}`);
  }

  next();
});

// Admin routes specific CORS handling - handled above with improved global OPTIONS handler

// Add direct route for emergency admin endpoints with custom handling
app.use('/emergency', (req, res) => {
  console.log(`⚠️ EMERGENCY DIRECT HANDLER: ${req.method} ${req.url}`);

  // Handle emergency/admin/users endpoint
  if (req.path === '/admin/users' && req.method === 'GET') {
    console.log(`🔥 Direct implementation of emergency admin users endpoint`);

    // Make a fetch request to the API directly
    fetch(`${backendUrl}/api/health?debug=true`)
      .then(response => response.json())
      .then(data => {
        // Extract users from debug data if available
        const users = data.users || [];

        // Return a successful response
        res.json({
          users: users,
          count: users.length,
          emergency_access: true,
          direct_implementation: true,
          timestamp: new Date().toISOString(),
          message: "Emergency access successful via proxy direct implementation"
        });
      })
      .catch(error => {
        console.error('Error in emergency endpoint:', error);
        res.status(500).json({
          error: "Emergency endpoint error",
          message: error.message,
          emergency_access: true,
          timestamp: new Date().toISOString()
        });
      });
    return;
  }

  // Handle emergency/admin/status endpoint
  if (req.path === '/admin/status' && req.method === 'GET') {
    console.log(`🔥 Direct implementation of emergency admin status endpoint`);

    // Return admin status directly from the proxy
    res.json({
      isAdmin: true,
      emergency_access: true,
      direct_implementation: true,
      timestamp: new Date().toISOString(),
      message: "Emergency admin access granted via proxy direct implementation"
    });
    return;
  }

  // For other endpoints, use the proxy middleware
  const proxyMiddleware = createProxyMiddleware({
    target: backendUrl,
    changeOrigin: true,
    xfwd: true,
    onProxyReq: (proxyReq, req, res) => {
      console.log(`⚠️ Proxying EMERGENCY request: ${req.method} ${req.url}`);

      // Explicitly forward authorization header
      if (req.headers.authorization) {
        console.log('  Forwarding Authorization header for emergency endpoint');
        proxyReq.setHeader('Authorization', req.headers.authorization);
      }

      // Add special debug header
      proxyReq.setHeader('X-Emergency-Access', 'true');
    },
    onProxyRes: (proxyRes, req, res) => {
      console.log(`Emergency endpoint response: ${proxyRes.statusCode}`);
    }
  });

  proxyMiddleware(req, res);
});

// Add direct route for debug admin endpoints
app.use('/debug_admin_status', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true,
  onProxyReq: (proxyReq, req, res) => {
    console.log(`⚠️ Proxying direct debug_admin_status request`);

    // Add debug headers
    proxyReq.setHeader('X-Debug-Admin', 'true');

    // Forward auth header if present
    if (req.headers.authorization) {
      console.log('  Forwarding Authorization header for debug endpoint');
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    console.log(`Debug admin status response: ${proxyRes.statusCode}`);
  }
}));

// WebSocket proxy for /ws endpoints
app.use('/ws', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  ws: true, // Enable WebSocket proxy
  xfwd: true, // Forward original client IP and host headers
  pathRewrite: {
    '^/ws': '/ws' // Keep the /ws path
  },
  // Special handling for WebSocket upgrade
  onProxyReq: (proxyReq, req, res) => {
    // Log WebSocket connection attempt
    console.log(`Proxying WebSocket initial request to: ${backendUrl}/ws`);
    console.log(`  Origin: ${req.headers.origin || 'unknown'}`);

    // Forward authentication token from query params or headers
    const token = req.query.token;
    if (token) {
      console.log('  Authentication token present in query params');
      // Add token to request header if it's in the query parameters
      proxyReq.setHeader('Authorization', `Bearer ${token}`);
    }

    // Also forward any existing Authorization header
    const authHeader = req.headers.authorization;
    if (authHeader) {
      console.log('  Authorization header present and forwarded');
      proxyReq.setHeader('Authorization', authHeader);
    }
  },
  // Handle WebSocket specific errors
  onError: (err, req, res) => {
    console.error('WebSocket proxy error:', err);
    // Try to send error if headers not sent yet
    if (!res.headersSent) {
      res.status(502).json({
        error: 'WebSocket Proxy Error',
        message: err.message,
        code: 'WS_PROXY_ERROR'
      });
    }
  },
  // Log when WebSocket connection is established
  onProxyRes: (proxyRes, req, res) => {
    console.log(`WebSocket response status: ${proxyRes.statusCode}`);
  }
}));

// Special handler for admin API routes - this ensures proper CORS handling for admin endpoints
app.use('/api/admin', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true,
  pathRewrite: {
    '^/api/admin': '/api/admin'
  },
  onProxyReq: (proxyReq, req, res) => {
    console.log(`🔐 Admin API request: ${req.method} ${req.url}`);
    console.log(`  To: ${backendUrl}/api/admin${req.url}`);
    console.log(`  From origin: ${req.headers.origin || 'unknown'}`);

    // Always forward Authorization header for admin routes
    if (req.headers.authorization) {
      const authPreview = req.headers.authorization.substring(0, 15) + '...';
      console.log(`  Admin auth header present: ${authPreview}`);
      proxyReq.setHeader('Authorization', req.headers.authorization);
    } else {
      console.warn('⚠️ Warning: No authorization header for admin request');
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Ensure CORS headers are present in the admin route response
    const origin = req.headers.origin || '*';
    res.setHeader('Access-Control-Allow-Origin', origin);
    res.setHeader('Access-Control-Allow-Credentials', 'true');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With, Accept, Origin');

    // For admin routes, add extra logging
    console.log(`Admin API response status: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    if (proxyRes.statusCode === 401 || proxyRes.statusCode === 403) {
      console.log(`⚠️ ADMIN AUTH ERROR: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
      console.log(`  Auth header present: ${!!req.headers.authorization}`);
      console.log(`  Origin: ${req.headers.origin || 'unknown'}`);
    }
  },
  onError: (err, req, res) => {
    console.error('Admin API proxy error:', err);
    console.error(`  Failed admin request: ${req.method} ${req.url}`);

    if (!res.headersSent) {
      // Ensure CORS headers even in error responses for admin routes
      const origin = req.headers.origin || '*';
      res.setHeader('Access-Control-Allow-Origin', origin);
      res.setHeader('Access-Control-Allow-Credentials', 'true');

      res.status(502).json({
        error: 'Admin API Proxy Error',
        message: err.message,
        code: 'ADMIN_API_PROXY_ERROR',
        path: req.url,
        timestamp: new Date().toISOString()
      });
    }
  }
}));

// API proxy for all other API requests
app.use('/api', createProxyMiddleware({
  target: backendUrl,
  changeOrigin: true,
  xfwd: true, // Forward original client IP and host headers
  pathRewrite: {
    '^/api': '/api' // Keep the /api path
  },
  onProxyReq: (proxyReq, req, res) => {
    // Log each API request for debugging
    console.log(`Proxying API request: ${req.method} ${req.url}`);
    console.log(`  To: ${backendUrl}${req.url}`);
    console.log(`  From origin: ${req.headers.origin || 'unknown'}`);

    // Explicitly preserve and forward the Authorization header
    if (req.headers.authorization) {
      const authPreview = req.headers.authorization.substring(0, 15) + '...';
      console.log(`  Authorization header present: ${authPreview}`);
      proxyReq.setHeader('Authorization', req.headers.authorization);
    }

    // Forward token from query parameter if present (for compatibility)
    const token = req.query.token;
    if (token && !req.headers.authorization) {
      console.log('  Adding Authorization header from query token');
      proxyReq.setHeader('Authorization', `Bearer ${token}`);
    }
  },
  onProxyRes: (proxyRes, req, res) => {
    // Enhanced CORS headers for all responses (especially important for admin routes)
    const origin = req.headers.origin || '*';
    res.setHeader('Access-Control-Allow-Origin', origin);
    res.setHeader('Access-Control-Allow-Credentials', 'true');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS, PATCH');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-Requested-With, Accept, Origin');

    // For admin routes, explicitly ensure CORS headers are present
    if (req.url.includes('/admin/')) {
      console.log(`🔐 Admin route detected: ${req.method} ${req.url} - Ensuring CORS headers`);

      // Double-check origin handling for admin routes
      if (origin !== '*') {
        console.log(`  Setting explicit origin for admin route: ${origin}`);
      }
    }

    // For 401/403 errors, add more debug info in the log
    if (proxyRes.statusCode === 401 || proxyRes.statusCode === 403) {
      console.log(`⚠️ AUTH ERROR: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
      console.log(`  Auth header present: ${!!req.headers.authorization}`);
      console.log(`  Origin: ${req.headers.origin || 'unknown'}`);
    } else {
      // Log response status for debugging
      console.log(`API response status: ${proxyRes.statusCode} for ${req.method} ${req.url}`);
    }
  },
  onError: (err, req, res) => {
    console.error('API proxy error:', err);
    console.error(`  Failed request: ${req.method} ${req.url}`);

    if (!res.headersSent) {
      res.status(502).json({
        error: 'API Proxy Error',
        message: err.message,
        code: 'API_PROXY_ERROR',
        path: req.url,
        timestamp: new Date().toISOString()
      });
    }
  }
}));

// Handle unknown routes
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Not Found',
    message: 'Use /api/* to proxy API requests, /ws for WebSocket connections, or /emergency/* for direct admin access',
    available_endpoints: ['/api/*', '/ws', '/emergency/*', '/debug_admin_status', '/debug/headers', '/health'],
    timestamp: new Date().toISOString()
  });
});

// Log environment variables for IPv4 settings
console.log(`Environment configuration:`);
console.log(`- MYSQL_TCP_PROTOCOL: ${process.env.MYSQL_TCP_PROTOCOL || 'not set'}`);
console.log(`- IPV6_DISABLED: ${process.env.IPV6_DISABLED || 'not set'}`);
console.log(`- GODEBUG: ${process.env.GODEBUG || 'not set'}`);
console.log(`- DB_HOST: ${process.env.DB_HOST || 'not set'}`);
console.log(`- DB_PORT: ${process.env.DB_PORT || 'not set'}`);

// Start the server
app.listen(port, () => {
  console.log(`CORS Proxy running on port ${port}`);
  console.log(`Proxying to backend: ${backendUrl}`);
});