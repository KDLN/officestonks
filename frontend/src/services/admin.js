// Admin service for frontend
import { getToken } from './auth';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : `${BACKEND_URL}/api`;  // Direct connection to backend in production

const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;
const ADMIN_URL = `${API_URL}/admin`;

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
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to check admin status');
    }

    const data = await response.json();
    return data.isAdmin || false;
  } catch (error) {
    console.error('Error checking admin status:', error);
    throw error;
  }
};

// Get all users (admin only)
export const getAllUsers = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/users`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to fetch users');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching users:', error);
    throw error;
  }
};

// Update user (admin only)
export const updateUser = async (userId, updates) => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/users/${userId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to update user');
    }

    return await response.json();
  } catch (error) {
    console.error('Error updating user:', error);
    throw error;
  }
};

// Delete user (admin only)
export const deleteUser = async (userId) => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/users/${userId}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to delete user');
    }

    return await response.json();
  } catch (error) {
    console.error('Error deleting user:', error);
    throw error;
  }
};

// Reset stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/stocks/reset`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to reset stock prices');
    }

    return await response.json();
  } catch (error) {
    console.error('Error resetting stock prices:', error);
    throw error;
  }
};

// Clear all chats (admin only)
export const clearAllChats = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${ADMIN_URL}/chat/clear`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to clear chats');
    }

    return await response.json();
  } catch (error) {
    console.error('Error clearing chats:', error);
    throw error;
  }
};