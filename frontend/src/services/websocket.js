// WebSocket service for real-time updates
import { getAuthToken } from './authBridge';
import websocketForensics from './websocketForensics';

let socket = null;
let listeners = {};
let reconnectTimer = null;
let reconnectAttempts = 0;
let connectionState = 'disconnected'; // disconnected, connecting, connected, failed
let heartbeatInterval = null;
let heartbeatTimeout = null;
let missedHeartbeats = 0;
let pollingInterval = null;
let usePollingFallback = false;
const MAX_RECONNECT_ATTEMPTS = 10; // Increased from 5
const INITIAL_RECONNECT_DELAY = 1000; // 1 second
const MAX_RECONNECT_DELAY = 30000; // 30 seconds
const HEARTBEAT_INTERVAL = 10000; // 10 seconds
const HEARTBEAT_TIMEOUT = 5000; // 5 seconds to wait for pong
const MAX_MISSED_HEARTBEATS = 3; // Force reconnect after 3 missed heartbeats

// SSE connection for Railway compatibility
let sseEventSource = null;

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
console.log('WebSocket using backend URL:', BACKEND_URL);

// Get current connection state
export const getConnectionState = () => connectionState;

// Detect if we're on Railway deployment
const isRailwayDeployment = () => {
  return window.location.hostname.includes('railway.app') || 
         window.location.hostname.includes('officestonks.com') ||
         window.location.hostname.includes('beta.officestonks.com');
};

// Initialize WebSocket connection
export const initWebSocket = async () => {
  // If on Railway, immediately use SSE instead of WebSocket
  if (isRailwayDeployment()) {
    console.log('🚂 Railway deployment detected - using SSE for real-time updates');
    console.log('Railway does not support WebSocket hijacking, switching to SSE');
    
    // Use SSE through the same /ws endpoint (Railway handler will serve SSE)
    const token = await getAuthToken();
    if (!token) {
      console.error('No authentication token available for SSE connection');
      connectionState = 'disconnected';
      notifyListeners('connectionState', { state: 'disconnected', reason: 'no_token' });
      return;
    }
    
    initSSEConnection(token);
    return;
  }

  // Start forensic tracking for this connection attempt
  const attemptId = websocketForensics.startConnectionAttempt({
    url: getCurrentWebSocketURL(),
    timestamp: Date.now(),
    userAgent: navigator.userAgent,
    connectionState: connectionState,
    reconnectAttempt: reconnectAttempts
  });

  // For non-Railway deployments, try WebSocket
  console.log('🔌 Attempting WebSocket connection (non-Railway environment)');
  
  // Reset any previous polling if WebSocket attempt is being made
  if (usePollingFallback) {
    usePollingFallback = false;
    stopPollingFallback();
  }

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

  console.log('🔗 Attempting WebSocket connection to:', wsUrl);
  console.log('🌐 Current location:', window.location.href);
  console.log('🔧 Is localhost:', isLocalhost);

  try {
    socket = new WebSocket(wsUrl);
    
    // Set a connection timeout for Railway proxy issues
    const connectionTimeout = setTimeout(() => {
      if (socket && socket.readyState === WebSocket.CONNECTING) {
        console.error('WebSocket connection timeout - likely Railway proxy issue');
        socket.close();
        connectionState = 'disconnected';
        notifyListeners('connectionState', { 
          state: 'disconnected', 
          reason: 'connection_timeout',
          description: 'Connection timed out - likely proxy issue'
        });
        scheduleReconnect();
      }
    }, 45000); // Increased to 45 seconds for Railway cold starts
    
    // Clear timeout when connection is established
    socket.onopen = () => {
      clearTimeout(connectionTimeout);
      console.log('WebSocket connected successfully');
      
      // Log forensic success
      websocketForensics.logConnectionStage(attemptId, 'websocket_open', {
        protocol: socket.protocol,
        readyState: socket.readyState
      });
      websocketForensics.completeConnectionAttempt(attemptId, 'success', {
        protocol: socket.protocol,
        readyState: socket.readyState
      });
      
      reconnectAttempts = 0;
      connectionState = 'connected';
      usePollingFallback = false;
      stopPollingFallback();
      startHeartbeat();
      notifyListeners('connection', { status: 'connected' });
      notifyListeners('connectionState', { state: 'connected' });
    };


    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Handle heartbeat responses
        if (data.type === 'pong') {
          handleHeartbeatResponse();
          return;
        }
        
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
      clearTimeout(connectionTimeout);
      console.log('WebSocket disconnected:', event.code, event.reason);
      
      // Log forensic data for the close event
      websocketForensics.logConnectionStage(attemptId, 'websocket_close', {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
        readyState: socket.readyState
      });
      
      // Only complete the attempt as failed if it wasn't already completed successfully
      if (connectionState !== 'connected' || event.code !== 1000) {
        websocketForensics.completeConnectionAttempt(attemptId, 'failed', {
          reason: 'websocket_close',
          code: event.code,
          closeReason: event.reason
        });
      }
      
      // Log detailed error information for debugging
      const errorInfo = {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
        timestamp: new Date().toISOString()
      };
      
      // Map error codes to human-readable messages
      const errorMessages = {
        1000: 'Normal closure',
        1001: 'Going away',
        1002: 'Protocol error',
        1003: 'Unsupported data',
        1004: 'Reserved',
        1005: 'No status received',
        1006: 'Abnormal closure - likely network or proxy issue',
        1007: 'Invalid frame payload data',
        1008: 'Policy violation',
        1009: 'Message too big',
        1010: 'Mandatory extension',
        1011: 'Internal server error',
        1015: 'TLS handshake failure'
      };
      
      console.log(`WebSocket error details:`, {
        ...errorInfo,
        description: errorMessages[event.code] || `Unknown error code: ${event.code}`
      });
      
      connectionState = 'disconnected';
      stopHeartbeat();
      notifyListeners('connection', { status: 'disconnected' });
      notifyListeners('connectionState', { state: 'disconnected', ...errorInfo });
      
      // More aggressive reconnection for 1006 errors (Railway proxy issues)
      const shouldReconnect = event.code !== 1000 && reconnectAttempts < MAX_RECONNECT_ATTEMPTS;
      const isProxyError = event.code === 1006; // Railway proxy timeouts
      
      if (shouldReconnect || isProxyError) {
        scheduleReconnect();
        // Start polling fallback after multiple failures
        if (reconnectAttempts > 3 && !usePollingFallback) {
          console.log('Starting HTTP polling fallback after multiple WebSocket failures');
          usePollingFallback = true;
          startPollingFallback();
        }
      } else if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        connectionState = 'failed';
        notifyListeners('connectionState', { state: 'failed', reason: 'max_attempts', lastError: errorInfo });
        // Ensure polling fallback is running
        if (!usePollingFallback) {
          console.log('WebSocket failed permanently, starting HTTP polling fallback');
          usePollingFallback = true;
          startPollingFallback();
        }
      }
    };
  } catch (error) {
    websocketForensics.logConnectionError(attemptId, error, {
      stage: 'websocket_creation_exception',
      url: wsUrl
    });
    websocketForensics.completeConnectionAttempt(attemptId, 'failed', {
      reason: 'creation_exception',
      error: error.message
    });
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
  
  reconnectTimer = setTimeout(async () => {
    console.log('Attempting to reconnect WebSocket...');
    
    // For 1006 errors, try to refresh the token before reconnecting
    try {
      const newToken = await getAuthToken();
      if (!newToken) {
        console.error('No token available for WebSocket reconnection');
        connectionState = 'failed';
        notifyListeners('connectionState', { state: 'failed', reason: 'no_token_on_reconnect' });
        return;
      }
      console.log('Token refreshed for WebSocket reconnection');
    } catch (error) {
      console.error('Failed to refresh token for WebSocket reconnection:', error);
    }
    
    initWebSocket();
  }, delay);
};

// Close WebSocket connection
export const closeWebSocket = () => {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  
  stopHeartbeat();
  stopPollingFallback();
  
  if (socket) {
    socket.close(1000, 'Client closing connection');
    socket = null;
  }
  
  connectionState = 'disconnected';
  listeners = {};
  reconnectAttempts = 0;
  usePollingFallback = false;
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
  // Notify specific event listeners
  if (listeners[event]) {
    listeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Error in WebSocket listener for event ${event}:`, error);
      }
    });
  }
  
  // Notify wildcard listeners (*)
  if (listeners['*']) {
    listeners['*'].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Error in WebSocket wildcard listener:`, error);
      }
    });
  }
};

// Send message through WebSocket
export const sendWebSocketMessage = (message) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    try {
      socket.send(JSON.stringify(message));
    } catch (error) {
      console.error('Error sending WebSocket message:', error);
    }
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

// Application-level heartbeat mechanism
const startHeartbeat = () => {
  stopHeartbeat(); // Clear any existing heartbeat
  missedHeartbeats = 0;
  
  console.log('Starting WebSocket heartbeat mechanism');
  
  // Send heartbeat every 10 seconds
  heartbeatInterval = setInterval(() => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      sendWebSocketMessage({ type: 'ping', timestamp: Date.now() });
      
      // Check for response after 5 seconds
      heartbeatTimeout = setTimeout(() => {
        missedHeartbeats++;
        console.warn(`Missed heartbeat #${missedHeartbeats}`);
        
        // If we miss 3 heartbeats, assume connection is dead
        if (missedHeartbeats >= MAX_MISSED_HEARTBEATS) {
          console.error('Connection appears dead after missing heartbeats, forcing reconnect');
          if (socket) {
            socket.close(4000, 'Heartbeat timeout');
          }
        }
      }, HEARTBEAT_TIMEOUT);
    }
  }, HEARTBEAT_INTERVAL);
};

const stopHeartbeat = () => {
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval);
    heartbeatInterval = null;
  }
  if (heartbeatTimeout) {
    clearTimeout(heartbeatTimeout);
    heartbeatTimeout = null;
  }
  missedHeartbeats = 0;
};

const handleHeartbeatResponse = () => {
  missedHeartbeats = 0;
  if (heartbeatTimeout) {
    clearTimeout(heartbeatTimeout);
    heartbeatTimeout = null;
  }
};

// HTTP polling fallback mechanism
const startPollingFallback = async () => {
  if (pollingInterval) return;
  
  console.log('Starting HTTP polling fallback for real-time updates');
  
  // Import the auth service for token
  const token = await getAuthToken();
  if (!token) {
    console.error('No token available for polling fallback');
    return;
  }
  
  // Poll every 2 seconds to match WebSocket update frequency
  pollingInterval = setInterval(async () => {
    try {
      const currentToken = await getAuthToken();
      if (!currentToken) {
        console.error('Token lost during polling');
        stopPollingFallback();
        return;
      }
      
      const response = await fetch(`${BACKEND_URL}/api/stocks`, {
        headers: {
          'Authorization': `Bearer ${currentToken}`
        }
      });
      
      if (response.ok) {
        const stocks = await response.json();
        // Convert to WebSocket message format and notify listeners
        if (Array.isArray(stocks)) {
          stocks.forEach(stock => {
            const sanitizedStock = sanitizeWebSocketData({
              type: 'stock_update',
              data: {
                stock_id: stock.id,
                symbol: stock.symbol,
                price: stock.current_price || stock.price
              }
            });
            notifyListeners('stockUpdate', sanitizedStock.data);
          });
        }
      } else if (response.status === 401) {
        console.error('Authentication failed during polling');
        stopPollingFallback();
      }
    } catch (error) {
      console.error('Polling fallback error:', error);
      // Don't stop polling on network errors, just log them
    }
  }, 2000);
};

const stopPollingFallback = () => {
  if (pollingInterval) {
    clearInterval(pollingInterval);
    pollingInterval = null;
    console.log('Stopped HTTP polling fallback');
  }
};

// Helper functions for forensics integration
const getCurrentWebSocketURL = () => {
  const token = 'REDACTED'; // Don't log actual tokens
  let wsUrl;
  if (isLocalhost) {
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.hostname;
    const wsPort = process.env.REACT_APP_WS_PORT || '8080';
    wsUrl = `${wsProtocol}//${wsHost}:${wsPort}/ws?token=${token}`;
  } else {
    const wsHost = BACKEND_URL.replace(/^https?:\/\//, '');
    wsUrl = `wss://${wsHost}/ws?token=${token}`;
  }
  return wsUrl.replace(/token=[^&]+/, 'token=REDACTED');
};

const isRailwayOrProduction = () => {
  return window.location.hostname.includes('railway.app') || 
         window.location.hostname.includes('officestonks.com') ||
         window.location.hostname.includes('beta.officestonks.com');
};

// Export debugging functions
export const isPollingActive = () => {
  return usePollingFallback && pollingInterval !== null;
};

export const getWebSocketForensics = () => {
  return websocketForensics;
};

export const getConnectionDiagnostics = () => {
  return {
    connectionState,
    reconnectAttempts,
    usePollingFallback,
    isPollingActive: isPollingActive(),
    diagnostics: websocketForensics.generateReport()
  };
};

// Initialize SSE connection for Railway compatibility
const initSSEConnection = (token) => {
  if (sseEventSource && sseEventSource.readyState === EventSource.OPEN) {
    console.log('SSE already connected');
    return;
  }
  
  console.log('🔗 Initializing SSE connection for Railway compatibility');
  connectionState = 'connecting';
  notifyListeners('connectionState', { state: 'connecting', protocol: 'SSE' });
  
  // For Railway, use the working SSE endpoint since /ws is giving 502 errors
  // This is a temporary fix while the Railway handler deployment completes
  let sseUrl = `${BACKEND_URL}/api/sse/stock-updates`;
  
  console.log('SSE connecting to:', sseUrl);
  console.log('Railway SSE using dedicated endpoint (temporary fix)');
  
  sseEventSource = new EventSource(sseUrl);
  
  sseEventSource.onopen = () => {
    console.log('✅ SSE connection established');
    connectionState = 'connected';
    reconnectAttempts = 0;
    notifyListeners('connection', { status: 'connected', protocol: 'SSE' });
    notifyListeners('connectionState', { state: 'connected', protocol: 'SSE' });
  };
  
  sseEventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      console.log('📨 SSE message:', data);
      
      // Process the same way as WebSocket messages
      if (data.type === 'stock_update') {
        notifyListeners('stock_update', data);
      } else if (data.type === 'connected') {
        console.log('SSE connection confirmed:', data.message);
      } else if (data.type === 'heartbeat' || data.type === 'pong') {
        // Handle heartbeat/pong responses
        console.log('SSE heartbeat received');
      } else {
        // Notify all listeners for other message types
        notifyListeners(data.type || 'message', data);
      }
    } catch (error) {
      console.error('Error parsing SSE message:', error, event.data);
    }
  };
  
  sseEventSource.onerror = (error) => {
    console.error('❌ SSE connection error:', error);
    
    if (sseEventSource.readyState === EventSource.CLOSED) {
      console.log('SSE connection closed, attempting reconnect...');
      connectionState = 'disconnected';
      notifyListeners('connectionState', { state: 'disconnected', protocol: 'SSE' });
      
      // Attempt to reconnect after delay
      if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        scheduleSSEReconnect();
      } else {
        console.error('SSE max reconnection attempts reached');
        connectionState = 'failed';
        notifyListeners('connectionState', { state: 'failed', protocol: 'SSE' });
      }
    }
  };
};

// Schedule SSE reconnection
const scheduleSSEReconnect = () => {
  if (reconnectTimer) return;
  
  reconnectAttempts++;
  const delay = Math.min(INITIAL_RECONNECT_DELAY * Math.pow(2, reconnectAttempts - 1), MAX_RECONNECT_DELAY);
  
  console.log(`Scheduling SSE reconnect attempt ${reconnectAttempts} in ${delay}ms`);
  
  reconnectTimer = setTimeout(async () => {
    reconnectTimer = null;
    const token = await getAuthToken();
    if (token) {
      initSSEConnection(token);
    }
  }, delay);
};