// Chat service for frontend
import { getToken } from './auth';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? process.env.REACT_APP_API_URL || '/api'
  : `${BACKEND_URL}/api`;
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;

// Get recent chat messages
export const getRecentMessages = async (limit = 50) => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/chat/messages?limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch chat messages');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching chat messages:', error);
    throw error;
  }
};

// Send a chat message
export const sendChatMessage = async (message) => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/chat/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      credentials: 'include',
      body: JSON.stringify({
        message,
      }),
    });

    if (!response.ok) {
      throw new Error('Failed to send chat message');
    }

    return await response.json();
  } catch (error) {
    console.error('Error sending chat message:', error);
    throw error;
  }
};