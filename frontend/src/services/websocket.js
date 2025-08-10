// WebSocket service now using Socket.IO for real-time updates with Railway compatibility
// This maintains backward compatibility with existing code
import socketIOService from './socketio';
import websocketForensics from './websocketForensics';

let isInitialized = false;
let stockUpdateListeners = [];
let chatMessageListeners = [];
let connectionStateListeners = [];
let lastError = null;

// Initialize WebSocket connection (now using Socket.IO)
export const initWebSocket = async () => {
  if (isInitialized) {
    console.log('🔗 WebSocket already initialized (Socket.IO)');
    return;
  }

  try {
    console.log('🚀 Initializing WebSocket with Socket.IO...');
    
    // Connect using Socket.IO
    await socketIOService.connect();
    
    // Set up event listeners
    socketIOService.on('stockUpdate', (data) => {
      stockUpdateListeners.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error('Error in stock update listener:', error);
        }
      });
    });

    socketIOService.on('chatMessage', (data) => {
      chatMessageListeners.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error('Error in chat message listener:', error);
        }
      });
    });

    socketIOService.on('connectionStateChange', (state) => {
      connectionStateListeners.forEach(callback => {
        try {
          callback(state);
        } catch (error) {
          console.error('Error in connection state listener:', error);
        }
      });
    });

    isInitialized = true;
    lastError = null;
    
    // Log successful initialization for forensics
    websocketForensics.logConnectionAttempt({
      url: socketIOService.getConnectionStats().backendUrl,
      timestamp: Date.now(),
      success: true,
      transport: 'Socket.IO',
      error: null
    });
    
    console.log('✅ WebSocket initialized successfully with Socket.IO');
    
  } catch (error) {
    console.error('❌ WebSocket initialization failed:', error);
    lastError = error;
    
    // Log failed initialization for forensics
    websocketForensics.logConnectionAttempt({
      url: socketIOService.getConnectionStats().backendUrl,
      timestamp: Date.now(),
      success: false,
      transport: 'Socket.IO',
      error: error.message
    });
    
    throw error;
  }
};

// Add WebSocket listener (backward compatibility)
export const addWebSocketListener = (type, callback) => {
  switch (type) {
    case 'stock_update':
    case 'stockUpdate':
      stockUpdateListeners.push(callback);
      break;
    case 'chat_message':
    case 'chatMessage':
      chatMessageListeners.push(callback);
      break;
    case 'connection_state':
    case 'connectionState':
      connectionStateListeners.push(callback);
      break;
    default:
      console.warn('Unknown listener type:', type);
  }
};

// Remove WebSocket listener (backward compatibility)
export const removeWebSocketListener = (type, callback) => {
  switch (type) {
    case 'stock_update':
    case 'stockUpdate':
      stockUpdateListeners = stockUpdateListeners.filter(cb => cb !== callback);
      break;
    case 'chat_message':
    case 'chatMessage':
      chatMessageListeners = chatMessageListeners.filter(cb => cb !== callback);
      break;
    case 'connection_state':
    case 'connectionState':
      connectionStateListeners = connectionStateListeners.filter(cb => cb !== callback);
      break;
    default:
      console.warn('Unknown listener type:', type);
  }
};

// Close WebSocket connection
export const closeWebSocket = () => {
  console.log('🔌 Closing WebSocket connection (Socket.IO)');
  socketIOService.disconnect();
  isInitialized = false;
  stockUpdateListeners = [];
  chatMessageListeners = [];
  connectionStateListeners = [];
};

// Get WebSocket instance (compatibility)
export const getWebSocketInstance = () => {
  return {
    isConnected: () => socketIOService.getConnectionState().connected,
    getState: () => socketIOService.getConnectionState().state,
    getStats: () => socketIOService.getConnectionStats(),
    ping: () => socketIOService.ping(),
    emit: (event, data) => socketIOService.emit(event, data),
    testConnection: () => socketIOService.testConnection(),
  };
};

// Send chat message
export const sendChatMessage = (message) => {
  return socketIOService.sendChatMessage(message);
};

// Get connection diagnostics (for diagnostics component)
export const getConnectionDiagnostics = () => {
  const stats = socketIOService.getConnectionStats();
  const forensicsData = websocketForensics.getComprehensiveReport();
  
  return {
    isConnected: stats.connected,
    connectionState: stats.state,
    socketId: stats.socketId,
    transport: stats.transport,
    backendUrl: stats.backendUrl,
    reconnectAttempts: stats.reconnectAttempts,
    lastError: lastError,
    forensicsData: forensicsData,
    connectionHistory: forensicsData.connectionHistory || [],
    stats: {
      totalAttempts: forensicsData.totalAttempts || 0,
      successfulConnections: forensicsData.successfulConnections || 0,
      failedConnections: forensicsData.failedConnections || 0,
      successRate: forensicsData.successRate || 0,
    }
  };
};

// Test connection quality
export const testConnectionQuality = async () => {
  return await socketIOService.testConnection();
};

// Get connection status
export const isConnected = () => {
  return socketIOService.getConnectionState().connected;
};

// Get connection state
export const getConnectionState = () => {
  return socketIOService.getConnectionState();
};

// Ping test
export const ping = () => {
  return socketIOService.ping();
};

// Detect Railway deployment
export const isRailwayDeployment = () => {
  return window.location.hostname.includes('railway.app');
};

// Legacy WebSocket functions for compatibility
export const connectWebSocket = initWebSocket;
export const disconnectWebSocket = closeWebSocket;

// Export default for convenience
export default {
  initWebSocket,
  addWebSocketListener,
  removeWebSocketListener,
  closeWebSocket,
  getWebSocketInstance,
  sendChatMessage,
  getConnectionDiagnostics,
  testConnectionQuality,
  isConnected,
  getConnectionState,
  ping,
  isRailwayDeployment,
  connectWebSocket: initWebSocket,
  disconnectWebSocket: closeWebSocket,
};