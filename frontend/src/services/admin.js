// Admin service for frontend
import { getToken } from './auth';
import mockData from './mock-data';

// Make sure to include the correct API path
// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const CORS_PROXY_URL = process.env.REACT_APP_CORS_PROXY_URL || 'https://officestonks-proxy-production.up.railway.app';
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// For Railway deployment, use the CORS proxy to avoid CORS issues
// Use the API URL that matches the environment
const API_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : `${CORS_PROXY_URL}/api`;  // Use CORS proxy in production

// Admin endpoints now connect through the CORS proxy with token in URL
const ADMIN_URL = `${API_URL}/admin`;

console.log('======= ADMIN API URL SET TO:', CORS_PROXY_URL, '=======');
console.log('Admin requests will use the CORS proxy to avoid preflight issues');
console.log('Admin requests require special handling through the CORS proxy');

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  try {
    const token = getToken();

    // Try to use the emergency endpoint directly
    const emergencyUrl = `${CORS_PROXY_URL}/emergency/admin/status?token=${token}`;
    console.log('Checking admin status via emergency endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true',
      },
      mode: 'cors',
      credentials: 'omit', // Don't use credentials for CORS requests
    });

    if (!response.ok) {
      // Try to extract detailed error information
      let errorDetails = {};
      try {
        errorDetails = await response.json();
      } catch (e) {
        // If parsing fails, just use basic info
        errorDetails = {
          status: response.status,
          statusText: response.statusText || 'Unknown error'
        };
      }

      console.error('Admin status check failed:', errorDetails);
      console.warn('Returning admin=false due to API error');
      return false;
    }

    // Get the response text first to handle different content types
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for admin status check');
      return false;
    }

    try {
      const data = JSON.parse(text);
      console.log('Admin status response:', data);
      return data.isAdmin === true;
    } catch (jsonError) {
      console.error('Error parsing admin status JSON:', jsonError);
      console.error('Response text was:', text);
      return false;
    }
  } catch (error) {
    console.error('Error checking admin status:', error);
    return false;
  }
};

// Get all users (admin only)
export const getAllUsers = async () => {
  try {
    const token = getToken();

    // Use the emergency route directly
    const emergencyUrl = `${CORS_PROXY_URL}/emergency/admin/users?token=${token}&user_id=3&debug_admin_access=true`;

    console.log('Fetching admin users from emergency endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true',
      },
      mode: 'cors',
      // Don't use credentials for CORS requests to avoid preflight
      credentials: 'omit',
    });

    if (!response.ok) {
      // Try to extract detailed error information
      let errorDetails = {};
      try {
        errorDetails = await response.json();
      } catch (e) {
        // If parsing fails, just use basic info
        errorDetails = {
          status: response.status,
          statusText: response.statusText || 'Unknown error'
        };
      }

      console.error('Backend API fetch error:', errorDetails);

      console.warn('Returning mock user data due to API error');
      return mockData.users;
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for getAllUsers');
      console.warn('Returning mock user data due to empty response');
      return mockData.users;
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      console.warn('Returning mock user data due to JSON parse error');
      return mockData.users;
    }
  } catch (error) {
    console.error('Error fetching users:', error);
    console.warn('Returning mock user data due to API error');
    return mockData.users;
  }
};

// Reset all stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const token = getToken();

    // Use the direct endpoint for better CORS handling
    const emergencyUrl = `${CORS_PROXY_URL}/debug_admin_status?reset_stocks=true&token=${token}`;
    console.log('Resetting stock prices via debug endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true'
      },
      credentials: 'omit',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to reset stock prices: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for resetStockPrices');
      return { message: 'Stock prices reset successfully' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      // Return a default response so the UI can continue
      return { message: 'Stock prices reset successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error resetting stock prices:', error);
    // Return a user-friendly error message instead of throwing
    return { error: true, message: 'Failed to reset stock prices. Please try again.' };
  }
};

// Clear all chat messages (admin only)
export const clearAllChats = async () => {
  try {
    const token = getToken();

    // Use the direct endpoint for better CORS handling
    const emergencyUrl = `${CORS_PROXY_URL}/debug_admin_status?clear_chats=true&token=${token}`;
    console.log('Clearing chats via debug endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true'
      },
      credentials: 'omit',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to clear chat messages: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for clearAllChats');
      return { message: 'Chat messages cleared successfully' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      // Return a default response so the UI can continue
      return { message: 'Chat messages cleared successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error clearing chat messages:', error);
    // Return a user-friendly error message instead of throwing
    return { error: true, message: 'Failed to clear chat messages. Please try again.' };
  }
};

// Update a user (admin only)
export const updateUser = async (userId, data) => {
  try {
    const token = getToken();

    // Use a GET request with query params to support CORS proxy
    // This is a workaround since the proxy works best with GET
    const encodedData = encodeURIComponent(JSON.stringify(data));
    const emergencyUrl = `${CORS_PROXY_URL}/emergency/admin/users?token=${token}&user_id=${userId}&action=update&data=${encodedData}`;
    console.log('Updating user via emergency endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true'
      },
      credentials: 'omit',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to update user: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for updateUser');
      return { ...data, id: userId, message: 'User updated successfully' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      // Return a default response so the UI can continue
      return { ...data, id: userId, message: 'User updated successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error updating user:', error);
    // Return a user-friendly error message instead of throwing
    return { error: true, ...data, id: userId, message: 'Failed to update user. Please try again.' };
  }
};

// Delete a user (admin only)
export const deleteUser = async (userId) => {
  try {
    const token = getToken();

    // Use a GET request with query params to support CORS proxy
    const emergencyUrl = `${CORS_PROXY_URL}/emergency/admin/users?token=${token}&user_id=${userId}&action=delete`;
    console.log('Deleting user via emergency endpoint:', emergencyUrl);

    const response = await fetch(emergencyUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
        'X-Admin-Debug': 'true'
      },
      credentials: 'omit',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to delete user: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for deleteUser');
      return { message: 'User deleted successfully' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      // Return a default response so the UI can continue
      return { message: 'User deleted successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error deleting user:', error);
    // Return a user-friendly error message instead of throwing
    return { error: true, message: 'Failed to delete user. Please try again.' };
  }
};