// Chat service for frontend
import { getToken } from './auth';
import mockData from './mock-data';

// Make sure to include the correct API path
const BASE_URL = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';
const API_URL = `${BASE_URL}/api`;

// Get recent chat messages
export const getRecentMessages = async (limit = 50) => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/chat/messages?limit=${limit}&token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch chat messages');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching chat messages:', error);
    console.warn('Returning mock chat messages due to API error');
    return mockData.messages;
  }
};

// Send a chat message
export const sendChatMessage = async (message) => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/chat/send?token=${token}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      mode: 'cors',
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
    console.warn('Creating mock message response due to API error');
    // Create a mock message response
    return {
      id: Math.floor(Math.random() * 1000),
      message: message,
      username: 'You (Mock)',
      created_at: new Date().toISOString()
    };
  }
};