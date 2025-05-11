// Mock data for when the backend API is unavailable
const mockData = {
  // Mock stocks data
  stocks: [
    { id: 1, symbol: 'AAPL', name: 'Apple Inc.', sector: 'Technology', current_price: 150.25 },
    { id: 2, symbol: 'MSFT', name: 'Microsoft Corporation', sector: 'Technology', current_price: 252.75 },
    { id: 3, symbol: 'GOOGL', name: 'Alphabet Inc.', sector: 'Technology', current_price: 2530.50 },
    { id: 4, symbol: 'AMZN', name: 'Amazon.com Inc.', sector: 'Technology', current_price: 3100.25 },
    { id: 5, symbol: 'META', name: 'Meta Platforms Inc.', sector: 'Technology', current_price: 298.50 },
    { id: 6, symbol: 'TSLA', name: 'Tesla, Inc.', sector: 'Automotive', current_price: 725.60 },
    { id: 7, symbol: 'NFLX', name: 'Netflix, Inc.', sector: 'Entertainment', current_price: 540.25 }
  ],
  
  // Mock portfolio data
  portfolio: {
    cash_balance: 10000,
    holdings: [
      { stock_id: 1, symbol: 'AAPL', name: 'Apple Inc.', quantity: 10, current_price: 150.25, value: 1502.50 },
      { stock_id: 2, symbol: 'MSFT', name: 'Microsoft Corporation', quantity: 5, current_price: 252.75, value: 1263.75 }
    ],
    total_value: 12766.25
  },
  
  // Mock transaction history
  transactions: [
    { id: 1, stock_id: 1, symbol: 'AAPL', quantity: 10, price: 145.25, transaction_type: 'buy', created_at: '2023-04-01T10:30:00Z' },
    { id: 2, stock_id: 2, symbol: 'MSFT', quantity: 5, price: 240.50, transaction_type: 'buy', created_at: '2023-04-02T14:20:00Z' },
    { id: 3, stock_id: 1, symbol: 'AAPL', quantity: 2, price: 150.25, transaction_type: 'sell', created_at: '2023-04-10T09:15:00Z' }
  ],
  
  // Mock users for admin panel
  users: [
    { id: 1, username: 'admin', cash_balance: 10000.00, is_admin: true, created_at: '2023-01-01T00:00:00Z' },
    { id: 2, username: 'user1', cash_balance: 5000.50, is_admin: false, created_at: '2023-01-02T10:30:00Z' },
    { id: 3, username: 'user2', cash_balance: 7500.25, is_admin: false, created_at: '2023-01-03T15:45:00Z' }
  ],
  
  // Mock chat messages
  messages: [
    { id: 1, user_id: 1, username: 'admin', message: 'Welcome to Office Stonks!', created_at: '2023-04-01T09:00:00Z' },
    { id: 2, user_id: 2, username: 'user1', message: 'Thanks for the welcome!', created_at: '2023-04-01T09:05:00Z' },
    { id: 3, user_id: 3, username: 'user2', message: 'How is everyone doing?', created_at: '2023-04-01T09:10:00Z' }
  ]
};

export default mockData;