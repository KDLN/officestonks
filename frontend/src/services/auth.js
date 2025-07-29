// Authentication service for the frontend

// API URL
// Make sure to include the correct API path
// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? '/api'  // Use relative URL when running locally
  : `${BACKEND_URL}/api`;  // Direct connection to backend in production

// Ensure we have the right URL format
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;
console.log("Auth service using API URL:", API_URL);

// Log full configuration for debugging
console.log("API Config:", {
  isLocalhost,
  BACKEND_URL,
  API_URL,
  WS_URL: `wss://${BACKEND_URL.replace(/^https?:\/\//, '')}/ws`
});

// Register a new user
export const register = async (username, password) => {
  try {
    const response = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Registration failed');
    }

    const data = await response.json();
    
    // Store token in localStorage
    localStorage.setItem('token', data.token);
    localStorage.setItem('userId', data.user_id);
    
    return data;
  } catch (error) {
    throw error;
  }
};

// Login an existing user
export const login = async (username, password) => {
  try {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Login failed');
    }

    const data = await response.json();
    
    // Store token in localStorage
    localStorage.setItem('token', data.token);
    localStorage.setItem('userId', data.user_id);
    
    return data;
  } catch (error) {
    throw error;
  }
};

// Logout
export const logout = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('userId');
  window.location.href = '/login';
};

// Check if user is authenticated
export const isAuthenticated = () => {
  return !!localStorage.getItem('token');
};

// Get authentication token
export const getToken = () => {
  return localStorage.getItem('token');
};

// Get user ID
export const getUserId = () => {
  return localStorage.getItem('userId');
};