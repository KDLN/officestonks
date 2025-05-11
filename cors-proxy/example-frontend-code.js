// Example frontend code showing how to use the CORS proxy
// This should replace the current websocket.js implementation

// Import authentication utilities
import { getToken } from './auth';

// Connection variables
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

  // *** IMPORTANT: USE THE CORS PROXY URL HERE ***
  // Replace with your actual CORS proxy URL from Railway
  const apiUrl = process.env.REACT_APP_API_URL || 'https://officestonks-cors-proxy.up.railway.app';
  
  // Check API health through the proxy
  fetch(`${apiUrl}/api/health`)
    .then(response => {
      if (!response.ok) {
        console.error(`Backend health check failed: ${response.status}`);
      } else {
        console.log('Backend health check passed');
        return response.json();
      }
    })
    .then(data => {
      if (data) console.log('Backend status:', data);
    })
    .catch(error => {
      console.error('Backend health check error:', error);
    });

  // Replace http/https with ws/wss for WebSocket connection
  const wsBase = apiUrl.replace(/^http/, 'ws');
  
  // Create the WebSocket URL through the proxy
  const wsUrl = `${wsBase}/ws?token=${token}`;
  
  console.log('Connecting to WebSocket:', wsUrl);
  socket = new WebSocket(wsUrl);
  
  // Connection opened
  socket.addEventListener('open', () => {
    console.log('WebSocket connection established successfully');
    reconnectAttempts = 0;
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  });
  
  // Listen for messages
  socket.addEventListener('message', (event) => {
    try {
      console.log('Received WebSocket message:', event.data);
      const message = JSON.parse(event.data);
      
      // Call listeners for this message type
      if (listeners[message.type]) {
        listeners[message.type].forEach(callback => callback(message));
      }
      
      // Call general listeners
      if (listeners['*']) {
        listeners['*'].forEach(callback => callback(message));
      }
    } catch (e) {
      console.error('Error processing WebSocket message:', e);
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