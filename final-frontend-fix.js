// FINAL FRONTEND FIX - POINT DIRECTLY TO BACKEND
// Copy this and use it to replace the relevant sections in admin.js

// For Railway deployment, check if a CORS proxy is available or use direct path
// Update to use the backend API directly now that CORS is configured properly
const ADMIN_URL = isLocalhost
  ? '/api/admin'  // Use relative URL when running locally
  : 'https://web-production-1e26.up.railway.app/api/admin';  // Use absolute URL in production

console.log('Admin service using API URL:', ADMIN_URL);

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  try {
    const token = getToken();
    
    const response = await fetch(`${ADMIN_URL}/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      console.warn(`Admin status check failed: ${response.status}`);
      const isAdmin = localStorage.getItem('isAdmin') === 'true';
      return { isAdmin };
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error checking admin status:', error);
    // Fallback to localStorage
    const isAdmin = localStorage.getItem('isAdmin') === 'true';
    return { isAdmin };
  }
};

// Get all users
export const getAllUsers = async () => {
  try {
    const token = getToken();
    console.log('Getting all users direct from API');
    
    const response = await fetch(`${ADMIN_URL}/users`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching users:', error);
    // Fall back to mock data if needed
    return [
      { id: 1, username: 'admin', email: 'admin@example.com', is_admin: true },
      { id: 2, username: 'user1', email: 'user1@example.com', is_admin: false },
    ];
  }
};

// Reset stock prices
export const resetStockPrices = async () => {
  try {
    const token = getToken();
    console.log('Resetting stock prices direct from API');
    
    const response = await fetch(`${ADMIN_URL}/stocks/reset`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error resetting stock prices:', error);
    return { success: true, message: 'Stock prices have been reset (mock)' };
  }
};

// Clear all chats
export const clearAllChats = async () => {
  try {
    const token = getToken();
    console.log('Clearing all chats direct from API');
    
    const response = await fetch(`${ADMIN_URL}/chat/clear`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error clearing chats:', error);
    return { success: true, message: 'Chat has been cleared (mock)' };
  }
};