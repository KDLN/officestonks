// User service for frontend
import { getToken } from './auth';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const CORS_PROXY_URL = process.env.REACT_APP_CORS_PROXY_URL || 'https://officestonks-proxy-production.up.railway.app';
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Make sure to include the correct API path
const BASE_URL = isLocalhost
  ? process.env.REACT_APP_API_URL || '/api'
  : CORS_PROXY_URL;
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;

// Get leaderboard data
export const getLeaderboard = async (limit = 10) => {
  try {
    const response = await fetch(`${API_URL}/users/leaderboard?limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch leaderboard');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching leaderboard:', error);
    throw error;
  }
};

// Get current user's profile
export const getUserProfile = async () => {
  try {
    const token = getToken();
    
    const response = await fetch(`${API_URL}/users/me`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch user profile');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching user profile:', error);
    throw error;
  }
};