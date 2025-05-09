// Chat service for frontend
import { getToken } from './auth';

// Make sure to include the correct API path
const BASE_URL = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';
const API_URL = `${BASE_URL}/api`;

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