// Stock market service for frontend
import { getToken } from './auth';
import { authenticatedFetch } from './authBridge';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? process.env.REACT_APP_API_URL || '/api'
  : `${BACKEND_URL}/api`;
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;
console.log("Stock service using API URL:", API_URL);

// Get all available stocks
export const getAllStocks = async () => {
  try {
    const response = await fetch(`${API_URL}/stocks`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch stocks');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching stocks:', error);
    throw error;
  }
};

// Get a specific stock by ID
export const getStockById = async (stockId) => {
  try {
    const response = await fetch(`${API_URL}/stocks/${stockId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch stock');
    }

    return await response.json();
  } catch (error) {
    console.error(`Error fetching stock ${stockId}:`, error);
    throw error;
  }
};

// Get the user's portfolio
export const getUserPortfolio = async () => {
  try {
    const response = await authenticatedFetch(`${API_URL}/portfolio`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch portfolio');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching portfolio:', error);
    throw error;
  }
};

// Execute a trade (buy or sell)
export const executeTrade = async (stockId, quantity, action) => {
  try {
    const response = await authenticatedFetch(`${API_URL}/trading`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        stock_id: stockId,
        quantity: quantity,
        action: action, // 'buy' or 'sell'
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to ${action} stock`);
    }

    return await response.json();
  } catch (error) {
    console.error(`Error executing ${action} trade:`, error);
    throw error;
  }
};

// Get transaction history
export const getTransactionHistory = async (limit = 50, offset = 0) => {
  try {
    const response = await authenticatedFetch(`${API_URL}/transactions?limit=${limit}&offset=${offset}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transaction history');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching transaction history:', error);
    throw error;
  }
};