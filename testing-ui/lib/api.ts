// API client for OfficeStonks

// Base URL for the API
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Types
export interface User {
  id: number;
  username: string;
  cash_balance: number;
}

export interface Stock {
  id: number;
  symbol: string;
  name: string;
  sector: string;
  current_price: number;
  last_updated: string;
}

export interface PortfolioItem {
  id: number;
  user_id: number;
  stock_id: number;
  quantity: number;
  stock: Stock;
}

export interface Transaction {
  id: number;
  user_id: number;
  stock_id: number;
  quantity: number;
  price: number;
  transaction_type: 'buy' | 'sell';
  created_at: string;
  stock: Stock;
}

export interface TradeRequest {
  stock_id: number;
  quantity: number;
  transaction_type: 'buy' | 'sell';
}

// Helper to get auth token
const getToken = (): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('token');
  }
  return null;
};

// Helper for making authenticated API requests
async function fetchWithAuth(endpoint: string, options: RequestInit = {}) {
  const token = getToken();
  
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });
  
  if (!response.ok) {
    const error = await response.json().catch(() => ({
      message: 'An unknown error occurred',
    }));
    throw new Error(error.message || `Error ${response.status}: ${response.statusText}`);
  }
  
  return response.json();
}

// API functions
export const api = {
  // Auth
  register: async (username: string, password: string) => {
    return fetchWithAuth('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  },
  
  login: async (username: string, password: string) => {
    return fetchWithAuth('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  },
  
  // Stocks
  getAllStocks: async (): Promise<Stock[]> => {
    return fetchWithAuth('/api/stocks');
  },
  
  getStockById: async (id: number): Promise<Stock> => {
    return fetchWithAuth(`/api/stocks/${id}`);
  },
  
  // Portfolio
  getUserPortfolio: async (): Promise<PortfolioItem[]> => {
    return fetchWithAuth('/api/portfolio');
  },
  
  // Transactions
  getTransactionHistory: async (): Promise<Transaction[]> => {
    return fetchWithAuth('/api/transactions');
  },
  
  // Trading
  tradeStock: async (tradeData: TradeRequest) => {
    return fetchWithAuth('/api/trading', {
      method: 'POST',
      body: JSON.stringify(tradeData),
    });
  },
};