// User service for frontend
import { getToken } from './auth';

// Make sure to include the correct API path
const BASE_URL = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';
const API_URL = `${BASE_URL}/api`;

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