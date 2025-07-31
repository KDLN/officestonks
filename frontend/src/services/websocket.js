// WebSocket service for real-time updates
import { getAuthToken } from './authBridge';

let socket = null;
let listeners = {};
let reconnectTimer = null;
let reconnectAttempts = 0;
let connectionState = 'disconnected'; // disconnected, connecting, connected, failed
const MAX_RECONNECT_ATTEMPTS = 5;
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';

// Get current connection state
export const getConnectionState = () => connectionState;

// Initialize WebSocket connection
export const initWebSocket = async () => {
  // Don't attempt if already connecting or connected
  if (connectionState === 'connecting' || (socket && socket.readyState === WebSocket.OPEN)) {
    console.log('WebSocket already connecting or connected');
    return;
  }

  // Don't attempt if we've failed too many times
  if (connectionState === 'failed') {
    console.error('WebSocket connection has failed permanently');
    return;
  }

  connectionState = 'connecting';
  notifyListeners('connectionState', { state: 'connecting' });

  const token = await getAuthToken();
  if (!token) {
    console.error('No authentication token available for WebSocket connection');
    connectionState = 'disconnected';
    notifyListeners('connectionState', { state: 'disconnected', reason: 'no_token' });
    return;
  }

  // Determine WebSocket URL based on environment
  let wsUrl;
  if (isLocalhost) {
    // For local development, use ws:// with localhost
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.hostname;
    const wsPort = process.env.REACT_APP_WS_PORT || '8080';
    wsUrl = `${wsProtocol}//${wsHost}:${wsPort}/ws?token=${token}`;
  } else {
    // For production, use wss:// with the backend URL
    const wsHost = BACKEND_URL.replace(/^https?:\/\//, '');
    wsUrl = `wss://${wsHost}/ws?token=${token}`;
  }

  console.log('Connecting to WebSocket:', wsUrl);

  try {
    socket = new WebSocket(wsUrl);

    socket.onopen = () => {
      console.log('WebSocket connected successfully');
      reconnectAttempts = 0;
      connectionState = 'connected';
      notifyListeners('connection', { status: 'connected' });
      notifyListeners('connectionState', { state: 'connected' });
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Sanitize data to prevent infinity/NaN issues
        const sanitizedData = sanitizeWebSocketData(data);
        console.log('WebSocket message received:', sanitizedData);
        
        // Handle different message types
        if (sanitizedData.type === 'stock_update') {
          notifyListeners('stockUpdate', sanitizedData.data);
        } else if (sanitizedData.type === 'chat_message') {
          notifyListeners('chatMessage', sanitizedData.data);
        } else {
          notifyListeners('message', sanitizedData);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error, 'Raw data:', event.data);
        // Don't crash, just skip the message
      }
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      notifyListeners('error', error);
    };

    socket.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      connectionState = 'disconnected';
      notifyListeners('connection', { status: 'disconnected' });
      notifyListeners('connectionState', { state: 'disconnected', code: event.code });
      
      // Attempt to reconnect if not a normal closure
      if (event.code !== 1000 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        scheduleReconnect();
      } else if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        connectionState = 'failed';
        notifyListeners('connectionState', { state: 'failed', reason: 'max_attempts' });
      }
    };
  } catch (error) {
    console.error('Error creating WebSocket connection:', error);
    connectionState = 'failed';
    notifyListeners('error', error);
    notifyListeners('connectionState', { state: 'failed', error });
  }
};

// Sanitize WebSocket data to prevent infinity/NaN issues
const sanitizeWebSocketData = (data) => {
  if (data === null || data === undefined) return data;
  
  if (typeof data === 'number') {
    if (!isFinite(data) || isNaN(data)) {
      console.warn('Sanitizing invalid number:', data);
      return 0;
    }
    return data;
  }
  
  if (Array.isArray(data)) {
    return data.map(item => sanitizeWebSocketData(item));
  }
  
  if (typeof data === 'object') {
    const sanitized = {};
    for (const key in data) {
      if (data.hasOwnProperty(key)) {
        sanitized[key] = sanitizeWebSocketData(data[key]);
      }
    }
    return sanitized;
  }
  
  return data;
};

// Schedule a reconnection attempt with exponential backoff
const scheduleReconnect = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }

  reconnectAttempts++;
  
  // Calculate delay with exponential backoff
  const delay = Math.min(
    INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts - 1),
    MAX_RECONNECT_DELAY
  );
  
  console.log(`Scheduling reconnect attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS} in ${delay}ms`);
  
  reconnectTimer = setTimeout(() => {
    console.log('Attempting to reconnect WebSocket...');
    initWebSocket();
  }, delay);
};

// Close WebSocket connection
export const closeWebSocket = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  
  if (socket) {
    socket.close(1000, 'Client closing connection');
    socket = null;
  }
  
  connectionState = 'disconnected';
  listeners = {};
  reconnectAttempts = 0;
};

// Reset WebSocket connection (useful for error recovery)
export const resetWebSocket = () => {
  closeWebSocket();
  connectionState = 'disconnected';
  setTimeout(() => {
    initWebSocket();
  }, 1000);
};

// Add event listener
export const addWebSocketListener = (event, callback) => {
  if (!listeners[event]) {
    listeners[event] = [];
  }
  listeners[event].push(callback);
};

// Remove event listener
export const removeWebSocketListener = (event, callback) => {
  if (listeners[event]) {
    listeners[event] = listeners[event].filter(cb => cb !== callback);
  }
};

// Notify all listeners for an event
const notifyListeners = (event, data) => {
  if (listeners[event]) {
    listeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Error in WebSocket listener for event ${event}:`, error);
      }
    });
  }
};

// Send message through WebSocket
export const sendWebSocketMessage = (message) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
  } else {
    console.error('WebSocket is not connected');
  }
};

// Check if WebSocket is connected
export const isWebSocketConnected = () => {
  return socket && socket.readyState === WebSocket.OPEN;
};

// Get the current WebSocket instance (for components that need direct access)
export const getWebSocketInstance = () => {
  return socket;
};