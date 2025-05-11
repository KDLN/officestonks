# Frontend WebSocket Changes

This document outlines the changes needed in the frontend codebase to fix WebSocket connectivity issues.

## Changes Required in `/frontend/src/services/websocket.js`

### 1. Update WebSocket API URL

```javascript
// Change this line
const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';
```

The URL should match the actual URL of your backend service on Railway.

### 2. Add Health Check Before WebSocket Connection

```javascript
// Add this code after the apiUrl definition
// First check if the backend API is available
fetch(`${apiUrl}/api/health`, {
  method: 'GET',
  credentials: 'include',
  headers: {
    'Accept': 'application/json',
  }
})
  .then(response => {
    if (!response.ok) {
      console.error(`Backend health check failed: ${response.status} ${response.statusText}`);
    } else {
      console.log('Backend health check passed');
      return response.json();
    }
  })
  .then(data => {
    if (data) console.log('Backend API status:', data);
  })
  .catch(error => {
    console.error('Backend health check error:', error);
    console.error('Backend API server may be unreachable - check server status');
  });

// Also check WebSocket health endpoint
fetch(`${apiUrl}/ws/health`, {
  method: 'GET',
  credentials: 'include',
  headers: {
    'Accept': 'application/json',
  }
})
  .then(response => {
    if (!response.ok) {
      console.error(`WebSocket health check failed: ${response.status} ${response.statusText}`);
    } else {
      console.log('WebSocket health check passed');
      return response.json();
    }
  })
  .then(data => {
    if (data) console.log('WebSocket health data:', data);
  })
  .catch(error => {
    console.error('WebSocket server health check failed:', error);
    console.error('WebSocket server may be unreachable');
    
    // Recommend alternative approach if health check fails
    console.log('Trying to establish WebSocket connection anyway...');
  });
```

### 3. Enhance WebSocket Error Handling

```javascript
// Update the error handler
socket.addEventListener('error', (error) => {
  console.error('WebSocket error:', error);
  // Add more detailed error information
  console.error('WebSocket connection failed - possible CORS issue or server unavailable');
  console.error('If this is a CORS error, ensure the backend allows WebSocket connections from this origin');
  console.error('Current origin:', window.location.origin);
  // Socket will automatically close after error
});
```

## Troubleshooting WebSocket Issues

1. **Check the console logs** for detailed error messages about WebSocket connectivity
2. **Verify that the backend URL is correct** - it should match your Railway deployment URL
3. **Check that the backend service is running** using the health check endpoints
4. **Verify CORS settings** if you're seeing CORS-related errors
5. **Check authentication token validity** if you're seeing authentication errors

## Deployment

After making these changes, deploy the frontend application. The WebSocket connection should now work, or you'll get more detailed error messages to help diagnose the issue.

If WebSocket connectivity problems persist, check the backend logs for any error messages related to WebSocket connections or CORS issues.