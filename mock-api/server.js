// Mock API server for OfficeStonks
const express = require('express');
const cors = require('cors');
const app = express();
const port = process.env.PORT || 8080;

// Enable CORS for all routes
app.use(cors({
  origin: '*',
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization']
}));

// Parse JSON bodies
app.use(express.json());

// Mock data
const mockData = {
  stocks: [
    { id: 1, symbol: 'AAPL', name: 'Apple Inc.', sector: 'Technology', current_price: 150.25 },
    { id: 2, symbol: 'MSFT', name: 'Microsoft Corporation', sector: 'Technology', current_price: 252.75 },
    { id: 3, symbol: 'GOOGL', name: 'Alphabet Inc.', sector: 'Technology', current_price: 2530.50 },
    { id: 4, symbol: 'AMZN', name: 'Amazon.com Inc.', sector: 'Technology', current_price: 3100.25 },
    { id: 5, symbol: 'META', name: 'Meta Platforms Inc.', sector: 'Technology', current_price: 298.50 },
    { id: 6, symbol: 'TSLA', name: 'Tesla, Inc.', sector: 'Automotive', current_price: 700.00 },
    { id: 7, symbol: 'NFLX', name: 'Netflix, Inc.', sector: 'Entertainment', current_price: 550.00 }
  ],
  portfolio: {
    cash_balance: 10000.00,
    holdings: [
      { stock_id: 1, symbol: 'AAPL', name: 'Apple Inc.', quantity: 10, current_price: 150.25, value: 1502.50 },
      { stock_id: 2, symbol: 'MSFT', name: 'Microsoft Corporation', quantity: 5, current_price: 252.75, value: 1263.75 }
    ],
    total_value: 12766.25
  },
  transactions: [
    { id: 1, stock_id: 1, symbol: 'AAPL', quantity: 10, price: 145.25, transaction_type: 'buy', created_at: new Date().toISOString() },
    { id: 2, stock_id: 2, symbol: 'MSFT', quantity: 5, price: 240.50, transaction_type: 'buy', created_at: new Date().toISOString() }
  ],
  users: [
    { id: 1, username: 'admin', cash_balance: 10000.00, is_admin: true, created_at: new Date().toISOString() },
    { id: 2, username: 'user1', cash_balance: 5000.00, is_admin: false, created_at: new Date().toISOString() },
    { id: 3, username: 'user2', cash_balance: 7500.00, is_admin: false, created_at: new Date().toISOString() }
  ],
  messages: [
    { id: 1, user_id: 1, username: 'admin', message: 'Welcome to Office Stonks!', created_at: new Date().toISOString() },
    { id: 2, user_id: 2, username: 'user1', message: 'Thanks for the welcome!', created_at: new Date().toISOString() }
  ]
};

// Log all requests
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
  next();
});

// API routes
// Health check
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', message: 'Mock API server running' });
});

// Stocks endpoints
app.get('/api/stocks', (req, res) => {
  res.json(mockData.stocks);
});

app.get('/api/stocks/:id', (req, res) => {
  const id = parseInt(req.params.id);
  const stock = mockData.stocks.find(s => s.id === id) || mockData.stocks[0];
  res.json(stock);
});

// User portfolio
app.get('/api/portfolio', (req, res) => {
  res.json(mockData.portfolio);
});

// Transactions history
app.get('/api/transactions', (req, res) => {
  res.json(mockData.transactions);
});

// Chat endpoints
app.get('/api/chat/messages', (req, res) => {
  res.json(mockData.messages);
});

app.post('/api/chat/send', (req, res) => {
  const newMessage = {
    id: mockData.messages.length + 1,
    user_id: 1,
    username: 'admin',
    message: req.body.message || 'New message',
    created_at: new Date().toISOString()
  };
  mockData.messages.push(newMessage);
  res.json(newMessage);
});

// Admin endpoints
app.get('/api/admin/status', (req, res) => {
  res.json({ isAdmin: true });
});

app.get('/api/admin/users', (req, res) => {
  res.json(mockData.users);
});

app.get('/api/admin/stocks/reset', (req, res) => {
  // Reset stock prices with random values
  mockData.stocks.forEach(stock => {
    stock.current_price = Math.round(Math.random() * 1000 * 100) / 100;
  });
  res.json({ message: 'Stock prices reset successfully' });
});

app.get('/api/admin/chat/clear', (req, res) => {
  mockData.messages = [
    { id: 1, user_id: 1, username: 'admin', message: 'Chat has been cleared', created_at: new Date().toISOString() }
  ];
  res.json({ message: 'Chat messages cleared successfully' });
});

// Catch-all for OPTIONS requests
app.options('*', cors());

// Start the server
app.listen(port, () => {
  console.log(`OfficeStonks Mock API server running on port ${port}`);
});