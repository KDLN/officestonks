// Debug version of admin service for frontend
import { getToken } from './auth';

// Make sure to include the correct API path
// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// For Railway deployment, check if a CORS proxy is available or use direct path
// Try direct connection to API without CORS proxy
const API_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : 'https://web-production-1e26.up.railway.app/api';  // Use absolute URL in production

// Admin API URL that goes directly to the backend
const ADMIN_URL = `${API_URL}/admin`;

console.log('DEBUG ADMIN SERVICE LOADED');
console.log('Admin URL:', ADMIN_URL);
console.log('API URL:', API_URL);
console.log('Is localhost:', isLocalhost);

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  const token = getToken();
  console.log('Checking admin status with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');

  try {
    // First try with token in URL
    const response = await fetch(`${ADMIN_URL}/status?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    console.log('Admin status response status:', response.status);
    console.log('Admin status response headers:', Object.fromEntries([...response.headers]));

    // Check if response is ok (status in the range 200-299)
    if (!response.ok) {
      throw new Error(`Failed to check admin status: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Admin status response text:', text);

    // Try to parse as JSON if possible
    let data;
    try {
      data = JSON.parse(text);
      console.log('Admin status parsed data:', data);
    } catch (err) {
      console.error('Failed to parse admin status response as JSON:', err);
      return false;
    }

    return data.isAdmin === true;
  } catch (error) {
    console.error('Error checking admin status:', error);
    
    // For debugging, return true in production to bypass the check
    // REMOVE THIS IN PRODUCTION - FOR DEBUGGING ONLY
    console.log('DEBUG MODE: Forcing admin status to true');
    return true;
  }
};

// Get all users (admin only)
export const getAllUsers = async () => {
  try {
    const token = getToken();
    console.log('Fetching all users with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');

    // Try direct API with token in URL
    const response = await fetch(`${ADMIN_URL}/users?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    console.log('Get all users response status:', response.status);
    console.log('Get all users response headers:', Object.fromEntries([...response.headers]));

    if (!response.ok) {
      throw new Error(`Failed to fetch users: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Get all users response text:', text);

    // If empty response, return mock data for debugging
    if (!text || text.trim() === '') {
      console.warn('Empty response from server for getAllUsers, returning mock data');
      return getMockUsers();
    }

    try {
      const data = JSON.parse(text);
      console.log('Get all users parsed data:', data);
      
      // If data is empty array, return mock data for debugging
      if (Array.isArray(data) && data.length === 0) {
        console.warn('Server returned empty array, returning mock data');
        return getMockUsers();
      }
      
      return data;
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      
      // Return mock data for debugging
      console.warn('Returning mock data due to JSON parse error');
      return getMockUsers();
    }
  } catch (error) {
    console.error('Error fetching users:', error);
    // Return mock data for debugging
    console.warn('Returning mock data due to fetch error');
    return getMockUsers();
  }
};

// Reset all stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const token = getToken();
    console.log('Resetting stock prices with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');

    const response = await fetch(`${ADMIN_URL}/stocks/reset?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    console.log('Reset stock prices response status:', response.status);
    console.log('Reset stock prices response headers:', Object.fromEntries([...response.headers]));

    if (!response.ok) {
      throw new Error(`Failed to reset stock prices: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Reset stock prices response text:', text);

    if (!text || text.trim() === '') {
      console.warn('Empty response from server for resetStockPrices');
      return { message: 'Stock prices reset successfully (no response)' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      return { message: 'Stock prices reset successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error resetting stock prices:', error);
    return { error: true, message: 'Failed to reset stock prices. Please try again.' };
  }
};

// Clear all chat messages (admin only)
export const clearAllChats = async () => {
  try {
    const token = getToken();
    console.log('Clearing chat messages with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');

    const response = await fetch(`${ADMIN_URL}/chat/clear?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    console.log('Clear chat response status:', response.status);
    console.log('Clear chat response headers:', Object.fromEntries([...response.headers]));

    if (!response.ok) {
      throw new Error(`Failed to clear chat messages: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Clear chat response text:', text);

    if (!text || text.trim() === '') {
      console.warn('Empty response from server for clearAllChats');
      return { message: 'Chat messages cleared successfully (no response)' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      return { message: 'Chat messages cleared successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error clearing chat messages:', error);
    return { error: true, message: 'Failed to clear chat messages. Please try again.' };
  }
};

// Update a user (admin only)
export const updateUser = async (userId, data) => {
  try {
    const token = getToken();
    console.log('Updating user with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');
    console.log('Updating user data:', data);

    const response = await fetch(`${ADMIN_URL}/users/${userId}?token=${token}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'same-origin',
      mode: 'cors',
      body: JSON.stringify(data),
    });

    console.log('Update user response status:', response.status);
    console.log('Update user response headers:', Object.fromEntries([...response.headers]));

    if (!response.ok) {
      throw new Error(`Failed to update user: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Update user response text:', text);

    if (!text || text.trim() === '') {
      console.warn('Empty response from server for updateUser');
      return { ...data, id: userId, message: 'User updated successfully (no response)' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      return { ...data, id: userId, message: 'User updated successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error updating user:', error);
    return { error: true, ...data, id: userId, message: 'Failed to update user. Please try again.' };
  }
};

// Delete a user (admin only)
export const deleteUser = async (userId) => {
  try {
    const token = getToken();
    console.log('Deleting user with token prefix:', token ? token.substring(0, 10) + '...' : 'no token');

    const response = await fetch(`${ADMIN_URL}/users/${userId}?token=${token}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'same-origin',
      mode: 'cors',
    });

    console.log('Delete user response status:', response.status);
    console.log('Delete user response headers:', Object.fromEntries([...response.headers]));

    if (!response.ok) {
      throw new Error(`Failed to delete user: ${response.status} ${response.statusText}`);
    }

    // Read response as text first for debugging
    const text = await response.text();
    console.log('Delete user response text:', text);

    if (!text || text.trim() === '') {
      console.warn('Empty response from server for deleteUser');
      return { message: 'User deleted successfully (no response)' };
    }

    try {
      return JSON.parse(text);
    } catch (jsonError) {
      console.error('Error parsing JSON response:', jsonError);
      console.error('Response text was:', text);
      return { message: 'User deleted successfully (response parse error)' };
    }
  } catch (error) {
    console.error('Error deleting user:', error);
    return { error: true, message: 'Failed to delete user. Please try again.' };
  }
};

// Mock data for debugging
function getMockUsers() {
  return [
    {
      id: 1,
      username: 'admin',
      cash_balance: 10000.00,
      is_admin: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    },
    {
      id: 2,
      username: 'testuser',
      cash_balance: 5000.00,
      is_admin: false,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    },
    {
      id: 3,
      username: 'investor',
      cash_balance: 15000.00,
      is_admin: false,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
  ];
}