// Official Socket.IO Server for Office Stonks
// This provides proper Socket.IO v4 implementation for Railway deployment

const express = require('express');
const { createServer } = require('http');
const { Server } = require('socket.io');
const cors = require('cors');
const jwt = require('jsonwebtoken');
const axios = require('axios');
require('dotenv').config();

const app = express();
app.use(cors());
app.use(express.json());

const httpServer = createServer(app);

// Socket.IO server with Railway-optimized configuration
const io = new Server(httpServer, {
  cors: {
    origin: process.env.CORS_ORIGIN || '*',
    methods: ['GET', 'POST'],
    credentials: true
  },
  transports: ['websocket', 'polling'],
  allowEIO3: true, // Allow different engine.io versions
  pingInterval: 25000,
  pingTimeout: 60000,
  upgradeTimeout: 30000,
  maxHttpBufferSize: 1e6
});

// JWT secret from environment or default for development
const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key-change-in-production';
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || 'http://localhost:8080';
// Always use port 3001 for Socket.IO server (internal)
const PORT = 3001;

// Track connected clients
const clients = new Map();

// Verify JWT token
function verifyToken(token) {
  try {
    const decoded = jwt.verify(token, JWT_SECRET);
    return decoded;
  } catch (error) {
    console.error('Token verification failed:', error.message);
    return null;
  }
}

// Socket.IO middleware for authentication
io.use((socket, next) => {
  const token = socket.handshake.auth.token || socket.handshake.query.token;
  
  if (!token) {
    return next(new Error('Authentication error: No token provided'));
  }

  const decoded = verifyToken(token);
  if (!decoded) {
    return next(new Error('Authentication error: Invalid token'));
  }

  // Attach user info to socket
  socket.userId = decoded.user_id || decoded.userId;
  socket.username = decoded.username || `User${socket.userId}`;
  next();
});

// Handle Socket.IO connections
io.on('connection', (socket) => {
  console.log(`✅ Client connected: ${socket.id} (User: ${socket.username})`);
  
  // Store client info
  const clientInfo = {
    socketId: socket.id,
    userId: socket.userId,
    username: socket.username,
    connectedAt: new Date(),
    rooms: new Set(['all'])
  };
  clients.set(socket.id, clientInfo);

  // Send connection confirmation
  socket.emit('connected', {
    message: 'Connected to Socket.IO server',
    protocol: 'Socket.IO v4',
    transport: socket.conn.transport.name,
    userId: socket.userId,
    username: socket.username
  });

  // Join user-specific room
  socket.join(`user_${socket.userId}`);
  clientInfo.rooms.add(`user_${socket.userId}`);

  // Handle stock subscription
  socket.on('subscribe_stocks', () => {
    socket.join('stocks');
    clientInfo.rooms.add('stocks');
    socket.emit('subscription_confirmed', { channel: 'stocks' });
    console.log(`📊 ${socket.username} subscribed to stock updates`);
  });

  // Handle chat room join
  socket.on('join_chat', () => {
    socket.join('chat');
    clientInfo.rooms.add('chat');
    socket.emit('chat_joined', { status: 'success' });
    console.log(`💬 ${socket.username} joined chat`);
  });

  // Handle chat messages
  socket.on('chat_message', async (message) => {
    const chatData = {
      type: 'chat_message',
      userId: socket.userId,
      username: socket.username,
      message: message,
      timestamp: Date.now()
    };

    // Broadcast to all clients in chat room
    io.to('chat').emit('chat_message', chatData);
    
    // Forward to Go backend for persistence
    try {
      await axios.post(`${GO_BACKEND_URL}/api/chat/message`, chatData, {
        headers: {
          'Authorization': `Bearer ${socket.handshake.auth.token}`
        }
      });
    } catch (error) {
      console.error('Failed to forward chat message to backend:', error.message);
    }
  });

  // Handle ping for connection testing
  socket.on('ping', (timestamp) => {
    socket.emit('pong', {
      timestamp: timestamp,
      serverTime: Date.now()
    });
  });

  // Handle disconnection
  socket.on('disconnect', (reason) => {
    console.log(`⚠️ Client disconnected: ${socket.id} (${reason})`);
    clients.delete(socket.id);
  });

  // Handle errors
  socket.on('error', (error) => {
    console.error(`❌ Socket error for ${socket.id}:`, error);
  });
});

// Connect to Go backend for stock updates
async function connectToGoBackend() {
  try {
    // Poll the Go backend for stock updates
    setInterval(async () => {
      try {
        const response = await axios.get(`${GO_BACKEND_URL}/api/stocks/current-prices`);
        const stocks = response.data;
        
        // Broadcast stock updates to subscribed clients
        if (stocks && stocks.length > 0) {
          stocks.forEach(stock => {
            const stockData = {
              type: 'stock_update',
              stock_id: stock.id,
              symbol: stock.symbol,
              price: stock.current_price,
              change: stock.price_change_24h
            };
            io.to('stocks').emit('stock_update', stockData);
          });
        }
      } catch (error) {
        // Silently handle errors (backend might be updating)
      }
    }, 2000); // Poll every 2 seconds for stock updates

    console.log('📈 Connected to Go backend for stock updates');
  } catch (error) {
    console.error('Failed to connect to Go backend:', error.message);
    // Retry connection after 5 seconds
    setTimeout(connectToGoBackend, 5000);
  }
}

// Admin endpoint for monitoring
app.get('/admin/socketio/stats', (req, res) => {
  const stats = {
    connected_clients: clients.size,
    clients: Array.from(clients.values()).map(client => ({
      socketId: client.socketId,
      userId: client.userId,
      username: client.username,
      connectedAt: client.connectedAt,
      rooms: Array.from(client.rooms)
    })),
    server_info: {
      protocol: 'Socket.IO v4',
      transports: ['websocket', 'polling'],
      cors_origin: process.env.CORS_ORIGIN || '*'
    }
  };
  res.json(stats);
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    connected_clients: clients.size,
    uptime: process.uptime()
  });
});

// Start server
httpServer.listen(PORT, () => {
  console.log(`🚀 Socket.IO server running on port ${PORT}`);
  console.log(`📡 Accepting connections from: ${process.env.CORS_ORIGIN || '*'}`);
  console.log(`🔗 Go backend URL: ${GO_BACKEND_URL}`);
  
  // Connect to Go backend for stock updates
  connectToGoBackend();
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, closing Socket.IO server...');
  io.close(() => {
    console.log('Socket.IO server closed');
    process.exit(0);
  });
});