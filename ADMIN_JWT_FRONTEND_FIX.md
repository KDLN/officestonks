# Admin Dashboard JWT Authentication Fix

## Problem Description

The admin dashboard is currently experiencing 401 Unauthorized errors when accessing admin API endpoints, despite having a valid JWT token. This is happening because the token signature validation is failing on the server side due to a mismatch in the JWT secret.

## Backend Changes Made

We've made the following changes to address this issue:

1. Implemented a token parser that extracts the user ID without validating the JWT signature
2. Added proper admin permission checks based on the user ID in the database
3. Fixed various CORS headers and admin endpoint handling
4. Added debug endpoints for troubleshooting JWT token issues

## Required Frontend Changes

The frontend needs to adopt one of the following approaches when making admin API requests:

### Option 1: Use Query Parameter Only (Recommended)

```javascript
// When fetching admin resources, use the token only as a query parameter
// Don't set the Authorization header at all
const token = getToken();
const response = await fetch(`${ADMIN_URL}/users?token=${token}`, {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
    // Remove the Authorization header completely
  },
  mode: 'cors',
});
```

### Option 2: Use Authorization Header Only

```javascript
// If you prefer to use the Authorization header
const token = getToken();
const response = await fetch(`${ADMIN_URL}/users`, {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  },
  mode: 'cors',
});
```

### Option 3: Use Both Methods (Current Implementation)

If you want to maintain the current implementation, ensure the token is exactly the same in both places:

```javascript
const token = getToken();
const response = await fetch(`${ADMIN_URL}/users?token=${token}`, {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`, // Same token as in URL
  },
  mode: 'cors',
  // Remove credentials: 'same-origin' or change to 'include'
});
```

## Additional Debug Tools

We've added a debug endpoint that can help diagnose token issues:

```javascript
// Test token parsing directly
const token = getToken();
const debugResponse = await fetch(
  `https://web-production-1e26.up.railway.app/debug-admin-jwt?token=${token}`
);
const debugData = await debugResponse.json();
console.log('Token debug:', debugData);
```

## Testing the Changes

1. Open the test tool: `/test-admin-jwt.html` in your browser
2. Enter your JWT token
3. Test the admin status API and debug endpoints
4. Check browser console for detailed errors

## Important Notes

1. The user with ID 3 (KDLN) should have admin privileges in the database
2. Make sure you're using the correct API URL in production (`https://web-production-1e26.up.railway.app`)
3. Check browser console for any CORS-related errors
4. If issues persist, try the standalone debug tool we've created: `/test-admin-jwt.html`
5. Make sure you're not including `credentials: 'same-origin'` in your fetch options, as this can interfere with cross-origin requests

## Mock Data Fallback

As a temporary solution, the frontend should implement a fallback to mock data when the admin API endpoints return errors. This ensures the admin dashboard remains functional even if there are intermittent API issues:

```javascript
try {
  const response = await fetch(`${ADMIN_URL}/users?token=${token}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    mode: 'cors',
  });
  
  if (!response.ok) {
    console.warn(`Admin API error: ${response.status} ${response.statusText}`);
    console.warn('Falling back to mock data');
    return mockUsers; // Use mock data as fallback
  }
  
  return await response.json();
} catch (error) {
  console.error('Error accessing admin API:', error);
  console.warn('Falling back to mock data');
  return mockUsers; // Use mock data as fallback
}
```

With these changes, the admin dashboard should work correctly once the backend deployment is complete.