// WebSocket service for real-time updates
import { getToken } from './auth';

let socket = null;
let listeners = {};
let reconnectTimer = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000; // 3 seconds

// Initialize WebSocket connection
export const initWebSocket = () => {
  if (socket) {
    // Close existing connection before creating a new one
    socket.close();
  }

  const token = getToken();
  if (!token) {
    console.error('No authentication token available for WebSocket connection');
    return;
  }

  // Create WebSocket connection with token
  // Get the backend URL from environment or default to Railway URL
  const apiUrl = process.env.REACT_APP_API_URL || 'https://web-production-1e26.up.railway.app';

  // Replace http/https with ws/wss
  const wsBase = apiUrl.replace(/^http/, 'ws');

  // Create the WebSocket URL
  const wsUrl = `${wsBase}/ws?token=${token}`;

  console.log('Connecting to WebSocket:', wsUrl);
  socket = new WebSocket(wsUrl);

  // Connection opened
  socket.addEventListener('open', () => {
    console.log('WebSocket connection established');
    // Reset reconnect attempts on successful connection
    reconnectAttempts = 0;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  });

  // Listen for messages
  socket.addEventListener('message', (event) => {
    try {
      // Log raw message for debugging
      console.log('Raw WebSocket message:', event.data);

      // Clean up the message if needed (in case there are control characters)
      let jsonStr = event.data;
      if (typeof jsonStr === 'string') {
        // Remove any BOM and control characters
        jsonStr = jsonStr.replace(/^\ufeff/, ''); // Remove byte order mark
        jsonStr = jsonStr.replace(/[\x00-\x1F\x7F-\x9F]/g, ''); // Remove control chars
      }

      const message = JSON.parse(jsonStr);

      // Call all listeners for this message type
      if (listeners[message.type]) {
        listeners[message.type].forEach(callback => callback(message));
      }

      // Call general listeners
      if (listeners['*']) {
        listeners['*'].forEach(callback => callback(message));
      }
    } catch (e) {
      console.error('Error parsing WebSocket message', e);
      console.error('Message content:', event.data);
      // Continue - don't let the error stop the socket
    }
  });

  // Connection closed
  socket.addEventListener('close', () => {
    console.log('WebSocket connection closed');
    // Attempt to reconnect
    reconnect();
  });

  // Connection error
  socket.addEventListener('error', (error) => {
    console.error('WebSocket error:', error);
    // Socket will automatically close after error
  });

  return socket;
};

// Add a listener for a specific message type
export const addListener = (type, callback) => {
  if (!listeners[type]) {
    listeners[type] = [];
  }
  listeners[type].push(callback);
  
  // Return function to remove this listener
  return () => {
    if (listeners[type]) {
      listeners[type] = listeners[type].filter(cb => cb !== callback);
    }
  };
};

// Remove a listener
export const removeListener = (type, callback) => {
  if (listeners[type]) {
    listeners[type] = listeners[type].filter(cb => cb !== callback);
  }
};

// Close the WebSocket connection
export const closeWebSocket = () => {
  if (socket) {
    socket.close();
    socket = null;
  }
  
  // Clear reconnect timer
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  
  // Clear all listeners
  listeners = {};
};

// Reconnect to WebSocket
const reconnect = () => {
  if (reconnectTimer || reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      console.error(`Failed to reconnect after ${MAX_RECONNECT_ATTEMPTS} attempts`);
    }
    return;
  }
  
  reconnectAttempts++;
  
  reconnectTimer = setTimeout(() => {
    console.log(`Attempting to reconnect (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`);
    reconnectTimer = null;
    initWebSocket();
  }, RECONNECT_DELAY);
};