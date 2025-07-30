// Stock market service for frontend
import { getToken } from './auth';
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
console.log("Stock service using API URL:", API_URL);

// Get all available stocks
export const getAllStocks = async () => {
  console.log('📈 Fetching all stocks from:', `${API_URL}/stocks`);
  try {
    const startTime = Date.now();
    const response = await fetch(`${API_URL}/stocks`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    const fetchTime = Date.now() - startTime;
    console.log(`📈 Stocks fetch response in ${fetchTime}ms:`, {
      status: response.status,
      statusText: response.statusText,
      ok: response.ok
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('📈 Stocks fetch failed:', errorText);
      throw new Error('Failed to fetch stocks');
    }

    const data = await response.json();
    console.log('📈 Stocks data parsed successfully:', data?.length || 0, 'stocks');
    return data;
  } catch (error) {
    console.error('📈 Error fetching stocks:', error);
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
  console.log('💼 Fetching user portfolio from:', `${API_URL}/portfolio`);
  try {
    const startTime = Date.now();
    const response = await authenticatedFetch(`${API_URL}/portfolio`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const fetchTime = Date.now() - startTime;
    console.log(`💼 Portfolio fetch response in ${fetchTime}ms:`, {
      status: response.status,
      statusText: response.statusText,
      ok: response.ok
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('💼 Portfolio fetch failed:', errorText);
      throw new Error('Failed to fetch portfolio');
    }

    const data = await response.json();
    console.log('💼 Portfolio data parsed successfully:', {
      total_value: data?.total_value || 0,
      cash_balance: data?.cash_balance || 0,
      stock_value: data?.stock_value || 0,
      items_count: data?.portfolio_items?.length || 0
    });
    return data;
  } catch (error) {
    console.error('💼 Error fetching portfolio:', error);
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
  console.log('📊 Fetching transaction history from:', `${API_URL}/transactions?limit=${limit}&offset=${offset}`);
  try {
    const startTime = Date.now();
    const response = await authenticatedFetch(`${API_URL}/transactions?limit=${limit}&offset=${offset}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const fetchTime = Date.now() - startTime;
    console.log(`📊 Transactions fetch response in ${fetchTime}ms:`, {
      status: response.status,
      statusText: response.statusText,
      ok: response.ok
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('📊 Transactions fetch failed:', errorText);
      throw new Error('Failed to fetch transaction history');
    }

    const data = await response.json();
    console.log('📊 Transactions data parsed successfully:', data?.length || 0, 'transactions');
    return data;
  } catch (error) {
    console.error('📊 Error fetching transaction history:', error);
    throw error;
  }
};