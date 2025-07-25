// WebSocket service for real-time updates
import { getAuthToken } from './authBridge';

let socket = null;
let listeners = {};
let reconnectTimer = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000; // 3 seconds

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://web-production-1e26.up.railway.app';

// Initialize WebSocket connection
export const initWebSocket = async () => {
  if (socket) {
    // Close existing connection before creating a new one
    socket.close();
  }

  const token = await getAuthToken();
  if (!token) {
    console.error('No authentication token available for WebSocket connection');
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
      notifyListeners('connection', { status: 'connected' });
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log('WebSocket message received:', data);
        
        // Handle different message types
        if (data.type === 'stock_update') {
          notifyListeners('stockUpdate', data.data);
        } else if (data.type === 'chat_message') {
          notifyListeners('chatMessage', data.data);
        } else {
          notifyListeners('message', data);
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      notifyListeners('error', error);
    };

    socket.onclose = (event) => {
      console.log('WebSocket disconnected:', event.code, event.reason);
      notifyListeners('connection', { status: 'disconnected' });
      
      // Attempt to reconnect if not a normal closure
      if (event.code !== 1000 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        scheduleReconnect();
      }
    };
  } catch (error) {
    console.error('Error creating WebSocket connection:', error);
    notifyListeners('error', error);
  }
};

// Schedule a reconnection attempt
const scheduleReconnect = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }

  reconnectAttempts++;
  console.log(`Scheduling reconnect attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS} in ${RECONNECT_DELAY}ms`);
  
  reconnectTimer = setTimeout(() => {
    console.log('Attempting to reconnect WebSocket...');
    initWebSocket();
  }, RECONNECT_DELAY);
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
  
  listeners = {};
  reconnectAttempts = 0;
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