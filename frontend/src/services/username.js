// Username management service
import { authenticatedFetch } from './authBridge';

// Check if a username is available
export const checkUsernameAvailability = async (username) => {
  const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';
  
  try {
    const response = await fetch(`${API_URL}/api/auth/check-username`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      throw new Error('Failed to check username availability');
    }

    return await response.json();
  } catch (error) {
    console.error('Error checking username availability:', error);
    throw error;
  }
};

// Set/update username (for logged-in users)
export const setUsername = async (username) => {
  const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';
  
  try {
    const response = await authenticatedFetch(`${API_URL}/api/auth/set-username`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to set username');
    }

    return await response.json();
  } catch (error) {
    console.error('Error setting username:', error);
    throw error;
  }
};

// Validate username format (client-side validation)
export const validateUsername = (username) => {
  if (!username) {
    return { valid: false, error: 'Username is required' };
  }
  
  if (username.length < 3) {
    return { valid: false, error: 'Username must be at least 3 characters long' };
  }
  
  if (username.length > 20) {
    return { valid: false, error: 'Username must be no more than 20 characters long' };
  }
  
  // Check for valid characters only
  const validUsernameRegex = /^[a-zA-Z0-9_]+$/;
  if (!validUsernameRegex.test(username)) {
    return { valid: false, error: 'Username can only contain letters, numbers, and underscores' };
  }
  
  return { valid: true };
};