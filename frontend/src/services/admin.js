// Admin service for frontend
import { getToken } from './auth';

// Make sure to include the correct API path
// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// For Railway deployment, check if a CORS proxy is available or use direct path
// If a proxy is available, admin endpoints will go through the proxy
const CORS_PROXY_URL = process.env.REACT_APP_CORS_PROXY_URL || 'https://officestonks-cors-proxy.up.railway.app';

// For Railway deployment, we might have different URLs for frontend and backend
// Use the API URL that matches the environment
const API_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : 'https://web-production-1e26.up.railway.app/api';  // Use absolute URL in production

// Admin specific URL that goes through the CORS proxy to avoid CORS issues
const ADMIN_URL = isLocalhost
  ? '/api/admin'  // Use relative URL when running locally
  : `${CORS_PROXY_URL}/admin`;  // Use CORS proxy in production for admin endpoints

console.log('Admin service using API URL:', API_URL);

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/status?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin', // Don't use 'include' with CORS proxy
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to check admin status');
    }

    const data = await response.json();
    return data.isAdmin;
  } catch (error) {
    console.error('Error checking admin status:', error);
    return false;
  }
};

// Get all users (admin only)
export const getAllUsers = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/users?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to fetch users: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for getAllUsers');
      return [];
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      // Return empty array instead of throwing to prevent UI errors
      return [];
    }
  } catch (error) {
    console.error('Error fetching users:', error);
    // Return empty array instead of re-throwing to prevent UI errors
    return [];
  }
};

// Reset all stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/stocks/reset?token=${token}`, {
      method: 'GET', // Use GET method for the CORS proxy
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
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

    const response = await fetch(`${ADMIN_URL}/chat/clear?token=${token}`, {
      method: 'GET', // Use GET for the CORS proxy instead of POST
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
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

    // For the PUT request, we'll still use the direct API because the CORS proxy currently only handles GET
    const response = await fetch(`${API_URL}/admin/users/${userId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      mode: 'cors',
      body: JSON.stringify(data),
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

    // Similar to updateUser, we'll keep using the direct API for DELETE
    const response = await fetch(`${API_URL}/admin/users/${userId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
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