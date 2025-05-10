const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  const backendUrl = process.env.REACT_APP_API_URL || 'https://web-copy-production-5b48.up.railway.app';
  
  // Proxy API requests to the backend
  app.use(
    '/api',
    createProxyMiddleware({
      target: backendUrl,
      changeOrigin: true,
      pathRewrite: { '^/api': '/api' },
      secure: false, // don't verify the SSL cert
      onProxyReq: function(proxyReq, req, res) {
        // Log the request
        console.log(`Proxying ${req.method} ${req.url} to ${backendUrl}${req.url}`);
      },
      onError: function(err, req, res) {
        console.error('Proxy error:', err);
        res.status(500).send('Proxy error occurred: ' + err.message);
      }
    })
  );

  // Proxy WebSocket requests
  app.use(
    '/ws',
    createProxyMiddleware({
      target: backendUrl.replace(/^http/, 'ws'),
      ws: true,
      changeOrigin: true,
      secure: false
    })
  );
};