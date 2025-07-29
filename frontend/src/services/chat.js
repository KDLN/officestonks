// Chat service for frontend
import { authenticatedFetch } from './authBridge';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? process.env.REACT_APP_API_URL || '/api'
  : `${BACKEND_URL}/api`;
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;

// Get recent chat messages
export const getRecentMessages = async (limit = 50) => {
  try {
    const response = await authenticatedFetch(`${API_URL}/chat/messages?limit=${limit}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
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
    const response = await authenticatedFetch(`${API_URL}/chat/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
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