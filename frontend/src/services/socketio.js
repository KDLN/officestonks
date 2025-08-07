// Socket.IO service for real-time updates with Railway compatibility
import io from 'socket.io-client';
import { getAuthToken } from './authBridge';

let socket = null;
let listeners = {};
let connectionState = 'disconnected'; // disconnected, connecting, connected, failed
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 10;

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const getCurrentBackendURL = () => {
  // If explicitly set, use that
  if (process.env.REACT_APP_BACKEND_URL) {
    return process.env.REACT_APP_BACKEND_URL;
  }
  
  // For localhost development
  if (isLocalhost) {
    return `${window.location.protocol}//${window.location.hostname}:8080`;
  }
  
  // For production, try to use the current domain
  // This handles Railway deployments automatically
  if (window.location.hostname.includes('railway.app')) {
    // If we're on a Railway frontend URL, try the backend
    const currentDomain = window.location.hostname;
    // Look for pattern like officestonks-frontend-production.up.railway.app
    // and convert to officestonks-production.up.railway.app (backend)
    const backendDomain = currentDomain.replace('-frontend', '');
    return `${window.location.protocol}//${backendDomain}`;
  }
  
  // For other production deployments, assume same domain
  return window.location.origin;
};

const BACKEND_URL = getCurrentBackendURL();

// Railway-optimized Socket.IO configuration
// Railway supports WebSocket over same PORT - this is the key to success!
const SOCKET_OPTIONS = {
  transports: ['websocket', 'polling'], // WebSocket first, polling fallback
  upgrade: true,
  rememberUpgrade: true,
  timeout: 20000,
  forceNew: false,
  reconnection: true,
  reconnectionAttempts: MAX_RECONNECT_ATTEMPTS,
  reconnectionDelay: 1000,
  reconnectionDelayMax: 30000,
  maxReconnectionAttempts: MAX_RECONNECT_ATTEMPTS,
  randomizationFactor: 0.5,
  // Railway-specific: ensure proper WebSocket URL formation
  autoConnect: true,
  forceBase64: false,
};

// Initialize Socket.IO connection
export const connect = () => {
  if (socket && socket.connected) {
    console.log('🔗 Socket.IO already connected');
    return Promise.resolve();
  }

  return new Promise((resolve, reject) => {
    connectionState = 'connecting';
    notifyConnectionStateChange();
    
    console.log(`🚀 Connecting to Socket.IO at ${BACKEND_URL}`);
    
    // Get auth token for connection
    const token = getAuthToken();
    const options = {
      ...SOCKET_OPTIONS,
      auth: {
        token: token
      },
      query: {
        token: token
      }
    };

    // Create Socket.IO connection
    socket = io(BACKEND_URL, options);

    // Connection successful
    socket.on('connect', () => {
      console.log('✅ Socket.IO connected successfully');
      console.log(`📡 Transport: ${socket.io.engine.transport.name}`);
      console.log(`🆔 Socket ID: ${socket.id}`);
      
      connectionState = 'connected';
      reconnectAttempts = 0;
      notifyConnectionStateChange();
      
      // Subscribe to stock updates
      socket.emit('subscribe_stocks');
      socket.emit('join_chat');
      
      resolve();
    });

    // Connection confirmation from server
    socket.on('connected', (data) => {
      console.log('🎯 Server connection confirmed:', data);
    });

    // Handle connection errors
    socket.on('connect_error', (error) => {
      console.error('❌ Socket.IO connection error:', error);
      connectionState = 'failed';
      reconnectAttempts++;
      notifyConnectionStateChange();
      
      if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        console.error('🚫 Max reconnection attempts reached');
        reject(new Error(`Connection failed after ${MAX_RECONNECT_ATTEMPTS} attempts`));
      }
    });

    // Handle disconnection
    socket.on('disconnect', (reason) => {
      console.warn('⚠️ Socket.IO disconnected:', reason);
      connectionState = 'disconnected';
      notifyConnectionStateChange();
      
      // Auto-reconnect for certain disconnect reasons
      if (reason === 'io server disconnect' || reason === 'transport close') {
        console.log('🔄 Attempting to reconnect...');
      }
    });

    // Handle transport upgrade (WebSocket fallback to polling)
    socket.io.on('upgrade', (transport) => {
      console.log('📈 Transport upgraded to:', transport.name);
    });

    socket.io.on('upgradeError', (error) => {
      console.warn('📉 Transport upgrade failed:', error);
    });

    // Handle stock updates
    socket.on('stock_update', (data) => {
      if (listeners.stockUpdate) {
        listeners.stockUpdate(data);
      }
    });

    // Handle subscription confirmations
    socket.on('subscription_confirmed', (data) => {
      console.log('📊 Subscription confirmed:', data);
    });

    socket.on('chat_joined', (data) => {
      console.log('💬 Chat joined:', data);
    });

    // Handle chat messages
    socket.on('chat_message', (data) => {
      if (listeners.chatMessage) {
        listeners.chatMessage(data);
      }
    });

    // Handle ping/pong for connection quality testing
    socket.on('pong', (data) => {
      console.log('🏓 Pong received:', data);
    });

    // Set a connection timeout
    setTimeout(() => {
      if (connectionState === 'connecting') {
        console.error('⏰ Socket.IO connection timeout');
        connectionState = 'failed';
        notifyConnectionStateChange();
        reject(new Error('Connection timeout'));
      }
    }, SOCKET_OPTIONS.timeout);
  });
};

// Disconnect from Socket.IO
export const disconnect = () => {
  if (socket) {
    console.log('🔌 Disconnecting Socket.IO');
    socket.disconnect();
    socket = null;
  }
  connectionState = 'disconnected';
  listeners = {};
  notifyConnectionStateChange();
};

// Subscribe to events
export const on = (event, callback) => {
  listeners[event] = callback;
  
  if (socket) {
    socket.on(event, callback);
  }
};

// Unsubscribe from events
export const off = (event) => {
  delete listeners[event];
  
  if (socket) {
    socket.off(event);
  }
};

// Send a message
export const emit = (event, data) => {
  if (socket && socket.connected) {
    socket.emit(event, data);
    return true;
  }
  console.warn('⚠️ Socket.IO not connected, cannot emit:', event);
  return false;
};

// Send ping for connection quality testing
export const ping = () => {
  if (socket && socket.connected) {
    socket.emit('ping', Date.now());
    return true;
  }
  return false;
};

// Get connection state
export const getConnectionState = () => {
  return {
    state: connectionState,
    connected: socket ? socket.connected : false,
    socketId: socket ? socket.id : null,
    transport: socket && socket.io && socket.io.engine ? socket.io.engine.transport.name : 'unknown',
    reconnectAttempts: reconnectAttempts,
  };
};

// Get connection statistics
export const getConnectionStats = () => {
  return {
    ...getConnectionState(),
    backendUrl: BACKEND_URL,
    options: SOCKET_OPTIONS,
    listeners: Object.keys(listeners),
  };
};

// Notify connection state changes
const notifyConnectionStateChange = () => {
  if (listeners.connectionStateChange) {
    listeners.connectionStateChange(getConnectionState());
  }
};

// Test connection quality
export const testConnection = async () => {
  return new Promise((resolve) => {
    if (!socket || !socket.connected) {
      resolve({
        success: false,
        error: 'Not connected',
        latency: null,
      });
      return;
    }

    const startTime = Date.now();
    const timeout = setTimeout(() => {
      resolve({
        success: false,
        error: 'Ping timeout',
        latency: null,
      });
    }, 5000);

    socket.once('pong', () => {
      clearTimeout(timeout);
      const latency = Date.now() - startTime;
      resolve({
        success: true,
        error: null,
        latency: latency,
      });
    });

    socket.emit('ping', startTime);
  });
};

// Handle chat messages specifically
export const sendChatMessage = (message) => {
  return emit('chat_message', message);
};

// Default export object
export default {
  connect,
  disconnect,
  on,
  off,
  emit,
  ping,
  getConnectionState,
  getConnectionStats,
  testConnection,
  sendChatMessage,
};