// Stock market service for frontend
import { getToken } from './auth';
import mockData from './mock-data';

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Make sure to include the correct API path
// Use correct Railway URL for production
const BASE_URL = isLocalhost
  ? process.env.REACT_APP_API_URL || '/api'
  : 'https://web-production-1e26.up.railway.app';
const API_URL = BASE_URL.endsWith('/api') ? BASE_URL : `${BASE_URL}/api`;
console.log("Stock service using API URL:", API_URL);

// Get all available stocks
export const getAllStocks = async () => {
  try {
    const token = getToken();
    const url = token ? `${API_URL}/stocks?token=${token}` : `${API_URL}/stocks`;

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch stocks');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching stocks:', error);
    console.warn('Returning mock stock data due to API error');
    // Return mock data when the API fails
    return mockData.stocks;
  }
};

// Get a specific stock by ID
export const getStockById = async (stockId) => {
  try {
    const token = getToken();
    const url = token ? `${API_URL}/stocks/${stockId}?token=${token}` : `${API_URL}/stocks/${stockId}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch stock');
    }

    return await response.json();
  } catch (error) {
    console.error(`Error fetching stock ${stockId}:`, error);
    console.warn('Returning mock stock data due to API error');
    // Return mock data for the specific stock
    const mockStock = mockData.stocks.find(stock => stock.id === parseInt(stockId));
    return mockStock || { id: stockId, symbol: 'MOCK', name: 'Mock Stock', current_price: 100.00 };
  }
};

// Get the user's portfolio
export const getUserPortfolio = async () => {
  try {
    const token = getToken();

    const response = await fetch(`${API_URL}/portfolio?token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch portfolio');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching portfolio:', error);
    console.warn('Returning mock portfolio data due to API error');
    return mockData.portfolio;
  }
};

// Execute a trade (buy or sell)
export const executeTrade = async (stockId, quantity, action) => {
  try {
    const token = getToken();
    
    const response = await fetch(`${API_URL}/trading`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
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
    const token = getToken();

    const response = await fetch(`${API_URL}/transactions?limit=${limit}&offset=${offset}&token=${token}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      mode: 'cors',
    });

    if (!response.ok) {
      throw new Error('Failed to fetch transaction history');
    }

    return await response.json();
  } catch (error) {
    console.error('Error fetching transaction history:', error);
    console.warn('Returning mock transaction data due to API error');
    return mockData.transactions;
  }
};