// Admin service for frontend
import { getToken } from './auth';

// Make sure to include the correct API path
// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// For Railway deployment, we might have different URLs for frontend and backend
// Use the API URL that matches the environment
const API_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : 'https://web-production-1e26.up.railway.app/api';  // Use absolute URL in production

console.log('Admin service using API URL:', API_URL);

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  try {
    const token = getToken();
    
    const response = await fetch(`${API_URL}/admin/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
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

    const response = await fetch(`${API_URL}/admin/users`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to fetch users: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text) {
      console.warn('Empty response from server for getAllUsers');
      return [];
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      throw new Error('Invalid JSON response from server');
    }
  } catch (error) {
    console.error('Error fetching users:', error);
    throw error;
  }
};

// Reset all stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/admin/stocks/reset`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to reset stock prices: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text) {
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
    throw error;
  }
};

// Clear all chat messages (admin only)
export const clearAllChats = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/admin/chat/clear`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      mode: 'cors',
    });

    if (!response.ok) {
      const statusText = response.statusText || 'Unknown error';
      throw new Error(`Failed to clear chat messages: ${response.status} ${statusText}`);
    }

    // Check if response has content before parsing JSON
    const text = await response.text();
    if (!text) {
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
    throw error;
  }
};

// Update a user (admin only)
export const updateUser = async (userId, data) => {
  try {
    const token = getToken();

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
    if (!text) {
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
    throw error;
  }
};

// Delete a user (admin only)
export const deleteUser = async (userId) => {
  try {
    const token = getToken();

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
    if (!text) {
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
    throw error;
  }
};