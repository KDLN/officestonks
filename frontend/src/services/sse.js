// Server-Sent Events service for real-time stock updates
import { getAuthToken } from './authBridge';

let eventSource = null;
let listeners = {};
let reconnectTimer = null;
let reconnectAttempts = 0;
let connectionState = 'disconnected'; // disconnected, connecting, connected, failed
const MAX_RECONNECT_ATTEMPTS = 10;
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get configuration from environment variables with fallbacks
// Try to detect the current deployment URL automatically
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
  
  // Default fallback
  return 'https://officestonks.com';
};

const BACKEND_URL = getCurrentBackendURL();
console.log('SSE using backend URL:', BACKEND_URL);

// Polling fallback variables
let pollingInterval = null;
let isPolling = false;

// Get current connection state
export const getSSEConnectionState = () => connectionState;

// Initialize SSE connection
export const initSSE = async () => {
  // Don't attempt if already connecting or connected
  if (connectionState === 'connecting' || (eventSource && eventSource.readyState === EventSource.OPEN)) {
    console.log('SSE already connecting or connected');
    return;
  }

  // Don't attempt if we've failed too many times
  if (connectionState === 'failed') {
    console.error('SSE connection has failed permanently');
    return;
  }

  connectionState = 'connecting';
  notifyListeners('connectionState', { state: 'connecting' });

  // Determine SSE URL based on environment
  let sseUrl;
  if (isLocalhost) {
    // For local development
    const protocol = window.location.protocol;
    const host = window.location.hostname;
    const port = process.env.REACT_APP_BACKEND_PORT || '8080';
    sseUrl = `${protocol}//${host}:${port}/api/sse/stock-updates`;
  } else {
    // For production
    sseUrl = `${BACKEND_URL}/api/sse/stock-updates`;
  }

  console.log('🔗 Attempting SSE connection to:', sseUrl);
  console.log('🌐 Current location:', window.location.href);
  console.log('🔧 Is localhost:', isLocalhost);
  
  // Test basic connectivity first
  try {
    console.log('🔍 Testing basic connectivity to backend...');
    fetch(`${BACKEND_URL}/health`, { method: 'GET' })
      .then(response => {
        console.log(`✅ Backend health check: ${response.status} ${response.statusText}`);
      })
      .catch(error => {
        console.log(`❌ Backend health check failed:`, error);
      });
  } catch (error) {
    console.log(`❌ Backend connectivity test error:`, error);
  }

  try {
    // Create EventSource instance
    eventSource = new EventSource(sseUrl);
    
    // Set a connection timeout
    const connectionTimeout = setTimeout(() => {
      if (eventSource && eventSource.readyState === EventSource.CONNECTING) {
        console.error('SSE connection timeout');
        eventSource.close();
        connectionState = 'disconnected';
        
        // If we've tried too many times, start polling fallback
        if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS - 1) {
          console.log('❌ SSE timeout after max attempts, starting polling fallback');
          console.log('🔄 Attempted URL was:', sseUrl);
          startPollingFallback();
          notifyListeners('connectionState', { 
            state: 'failed', 
            reason: 'connection_timeout',
            description: 'Connection timed out, using polling fallback',
            attempted_url: sseUrl
          });
        } else {
          console.log('⏳ SSE connection timeout, will retry...');
          console.log('🔄 Attempted URL was:', sseUrl);
          notifyListeners('connectionState', { 
            state: 'disconnected', 
            reason: 'connection_timeout',
            description: 'Connection timed out',
            attempted_url: sseUrl
          });
          scheduleReconnect();
        }
      }
    }, 20000); // 20 second timeout for Railway compatibility
    
    // Handle successful connection
    eventSource.onopen = () => {
      clearTimeout(connectionTimeout);
      console.log('SSE connected successfully');
      reconnectAttempts = 0;
      connectionState = 'connected';
      notifyListeners('connection', { status: 'connected' });
      notifyListeners('connectionState', { state: 'connected' });
    };

    // Handle messages
    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Sanitize data to prevent infinity/NaN issues
        const sanitizedData = sanitizeSSEData(data);
        console.log('SSE message received:', sanitizedData);
        
        // Handle different message types
        if (sanitizedData.type === 'stock_update') {
          notifyListeners('stockUpdate', sanitizedData);
        } else if (sanitizedData.type === 'connection') {
          notifyListeners('connection', sanitizedData);
        } else if (sanitizedData.type === 'heartbeat') {
          // Just log heartbeats, don't notify listeners
          console.log('SSE heartbeat received');
        } else {
          notifyListeners('message', sanitizedData);
        }
      } catch (error) {
        console.error('Error parsing SSE message:', error, 'Raw data:', event.data);
        // Don't crash, just skip the message
      }
    };

    // Handle errors
    eventSource.onerror = (error) => {
      console.error('SSE error:', error);
      
      // Check if the connection is still open
      if (eventSource.readyState === EventSource.CLOSED) {
        console.log('SSE connection closed');
        connectionState = 'disconnected';
        notifyListeners('connection', { status: 'disconnected' });
        notifyListeners('connectionState', { state: 'disconnected', reason: 'connection_closed' });
        
        // Schedule reconnection if we haven't exceeded max attempts
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
          scheduleReconnect();
        } else {
          connectionState = 'failed';
          console.log('SSE failed permanently, starting polling fallback');
          startPollingFallback();
          notifyListeners('connectionState', { 
            state: 'failed', 
            reason: 'max_attempts',
            description: 'Maximum reconnection attempts exceeded, using polling fallback'
          });
        }
      }
      
      notifyListeners('error', error);
    };

  } catch (error) {
    console.error('Error creating SSE connection:', error);
    connectionState = 'failed';
    notifyListeners('error', error);
    notifyListeners('connectionState', { state: 'failed', error });
  }
};

// Sanitize SSE data to prevent infinity/NaN issues
const sanitizeSSEData = (data) => {
  if (data === null || data === undefined) return data;
  
  if (typeof data === 'number') {
    if (!isFinite(data) || isNaN(data)) {
      console.warn('Sanitizing invalid number:', data);
      return 0;
    }
    return data;
  }
  
  if (Array.isArray(data)) {
    return data.map(item => sanitizeSSEData(item));
  }
  
  if (typeof data === 'object') {
    const sanitized = {};
    for (const key in data) {
      if (data.hasOwnProperty(key)) {
        sanitized[key] = sanitizeSSEData(data[key]);
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
  
  console.log(`Scheduling SSE reconnect attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS} in ${delay}ms`);
  
  reconnectTimer = setTimeout(() => {
    console.log('Attempting to reconnect SSE...');
    initSSE();
  }, delay);
};

// Close SSE connection
export const closeSSE = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  
  if (eventSource) {
    eventSource.close();
    eventSource = null;
  }
  
  stopPollingFallback();
  
  connectionState = 'disconnected';
  listeners = {};
  reconnectAttempts = 0;
};

// Reset SSE connection (useful for error recovery)
export const resetSSE = () => {
  closeSSE();
  connectionState = 'disconnected';
  setTimeout(() => {
    initSSE();
  }, 1000);
};

// Add event listener
export const addSSEListener = (event, callback) => {
  if (!listeners[event]) {
    listeners[event] = [];
  }
  listeners[event].push(callback);
};

// Remove event listener
export const removeSSEListener = (event, callback) => {
  if (listeners[event]) {
    listeners[event] = listeners[event].filter(cb => cb !== callback);
  }
};

// Notify all listeners for an event
const notifyListeners = (event, data) => {
  // Notify specific event listeners
  if (listeners[event]) {
    listeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Error in SSE listener for event ${event}:`, error);
      }
    });
  }
  
  // Notify wildcard listeners (*)
  if (listeners['*']) {
    listeners['*'].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Error in SSE wildcard listener:`, error);
      }
    });
  }
};

// Check if SSE is connected
export const isSSEConnected = () => {
  return eventSource && eventSource.readyState === EventSource.OPEN;
};

// Get the current EventSource instance (for components that need direct access)
export const getSSEInstance = () => {
  return eventSource;
};

// Start polling fallback when SSE fails
const startPollingFallback = () => {
  if (isPolling) return;
  
  isPolling = true;
  console.log('Starting HTTP polling fallback for stock updates');
  
  const pollStocks = async () => {
    try {
      let pollUrl;
      if (isLocalhost) {
        const protocol = window.location.protocol;
        const host = window.location.hostname;
        const port = process.env.REACT_APP_BACKEND_PORT || '8080';
        pollUrl = `${protocol}//${host}:${port}/api/stock-updates/poll`;
      } else {
        pollUrl = `${BACKEND_URL}/api/stock-updates/poll`;
      }
      
      const response = await fetch(pollUrl);
      if (response.ok) {
        const data = await response.json();
        
        // Process each stock update
        if (data.updates && Array.isArray(data.updates)) {
          data.updates.forEach(update => {
            if (update.type === 'stock_update') {
              notifyListeners('stockUpdate', update);
            }
          });
        }
      }
    } catch (error) {
      console.error('Polling fallback error:', error);
    }
  };
  
  // Poll every 3 seconds (slightly slower than SSE's 2 seconds)
  pollingInterval = setInterval(pollStocks, 3000);
  
  // Do an immediate poll
  pollStocks();
};

// Stop polling fallback
const stopPollingFallback = () => {
  if (pollingInterval) {
    clearInterval(pollingInterval);
    pollingInterval = null;
  }
  isPolling = false;
  console.log('Stopped HTTP polling fallback');
};

// Force reconnection (useful for debugging)
export const forceSSEReconnect = () => {
  console.log('Forcing SSE reconnection...');
  stopPollingFallback();
  closeSSE();
  setTimeout(() => initSSE(), 1000);
};