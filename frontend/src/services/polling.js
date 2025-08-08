// Simple HTTP Polling service for real-time stock updates
// Railway-compatible solution that avoids WebSocket and complex protocols
import { getAuthToken } from './authBridge';

let pollingInterval = null;
let listeners = {};
let connectionState = 'disconnected'; // disconnected, connecting, connected, failed
let pollCounter = 0;
let lastSuccessfulPoll = null;

// Check the current hostname to determine if we're running locally
const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

// Get backend URL
const getBackendURL = () => {
  if (process.env.REACT_APP_BACKEND_URL) {
    return process.env.REACT_APP_BACKEND_URL;
  }
  
  if (isLocalhost) {
    return `${window.location.protocol}//${window.location.hostname}:8080`;
  }
  
  // For production, use current domain
  return window.location.origin;
};

const BACKEND_URL = getBackendURL();
console.log('🔗 Polling service using backend URL:', BACKEND_URL);

// Polling configuration - optimized for rate limiting
const POLL_INTERVAL = 5000; // 5 seconds - reduced from 2s to avoid rate limits
const MAX_CONSECUTIVE_FAILURES = 3;
const ADAPTIVE_POLLING = true; // Enable adaptive polling based on activity
let consecutiveFailures = 0;

// Adaptive polling state
let isPageVisible = true;
let lastUserActivity = Date.now();
let currentPollInterval = POLL_INTERVAL;
let adaptiveTimer = null;

// Initialize HTTP polling
export const startPolling = async () => {
  if (pollingInterval) {
    console.log('📡 HTTP Polling already active');
    return;
  }

  connectionState = 'connecting';
  notifyListeners('connectionState', { state: connectionState });
  
  console.log('🚀 Starting HTTP polling for real-time updates');
  console.log(`📊 Poll interval: ${POLL_INTERVAL}ms`);
  console.log(`🎯 Backend URL: ${BACKEND_URL}`);

  // Start polling immediately
  await performPoll();
  
  // Set up adaptive polling
  if (ADAPTIVE_POLLING) {
    setupAdaptivePolling();
  } else {
    // Set up regular polling
    pollingInterval = setInterval(async () => {
      await performPoll();
    }, POLL_INTERVAL);
  }
  
  connectionState = 'connected';
  notifyListeners('connectionState', { state: connectionState });
  console.log('✅ HTTP polling started successfully');
};

// Stop HTTP polling
export const stopPolling = () => {
  if (pollingInterval) {
    clearInterval(pollingInterval);
    pollingInterval = null;
  }
  
  if (adaptiveTimer) {
    clearTimeout(adaptiveTimer);
    adaptiveTimer = null;
  }
  
  connectionState = 'disconnected';
  notifyListeners('connectionState', { state: connectionState });
  console.log('🛑 HTTP polling stopped');
};

// Perform a single poll request
const performPoll = async () => {
  try {
    pollCounter++;
    
    // Get auth token
    let token;
    try {
      token = await getAuthToken();
    } catch (error) {
      console.warn('⚠️ Could not get auth token for polling, continuing without auth');
    }

    // Construct poll URL
    const pollUrl = `${BACKEND_URL}/api/stock-updates/poll`;
    
    // Make the request
    const headers = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
      'X-Request-Type': 'polling', // Identify as polling request to bypass rate limiting
    };
    
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(pollUrl, {
      method: 'GET',
      headers: headers,
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    lastSuccessfulPoll = Date.now();
    consecutiveFailures = 0;
    
    // Log occasionally to avoid spam
    if (pollCounter % 30 === 1) { // Every minute at 2s intervals
      console.log(`📊 Poll #${pollCounter} successful - ${data.updates?.length || 0} updates received`);
    }
    
    // Process stock updates
    if (data.updates && Array.isArray(data.updates)) {
      data.updates.forEach(update => {
        if (update.type === 'stock_update') {
          notifyListeners('stockUpdate', update);
        } else if (update.type === 'connection') {
          notifyListeners('connection', update);
        } else {
          notifyListeners('message', update);
        }
      });
    }
    
    // Notify successful poll
    if (pollCounter % 30 === 1) {
      notifyListeners('poll', {
        success: true,
        counter: pollCounter,
        timestamp: Date.now(),
        updatesCount: data.updates?.length || 0
      });
    }
    
  } catch (error) {
    consecutiveFailures++;
    
    console.error(`❌ Polling error #${consecutiveFailures}:`, error.message);
    
    // Notify error
    notifyListeners('error', {
      error: error.message,
      consecutiveFailures,
      counter: pollCounter
    });
    
    // If too many failures, change state to failed
    if (consecutiveFailures >= MAX_CONSECUTIVE_FAILURES) {
      connectionState = 'failed';
      notifyListeners('connectionState', { 
        state: connectionState, 
        reason: 'too_many_failures',
        consecutiveFailures 
      });
      console.error(`🚫 HTTP polling failed after ${MAX_CONSECUTIVE_FAILURES} consecutive failures`);
    }
  }
};

// Add event listener
export const addPollingListener = (event, callback) => {
  if (!listeners[event]) {
    listeners[event] = [];
  }
  listeners[event].push(callback);
};

// Remove event listener
export const removePollingListener = (event, callback) => {
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
        console.error(`Error in polling listener for event ${event}:`, error);
      }
    });
  }
  
  // Notify wildcard listeners (*)
  if (listeners['*']) {
    listeners['*'].forEach(callback => {
      try {
        callback({ event, data });
      } catch (error) {
        console.error(`Error in polling wildcard listener:`, error);
      }
    });
  }
};

// Get connection state
export const getPollingConnectionState = () => connectionState;

// Check if polling is active
export const isPollingActive = () => !!pollingInterval;

// Get polling statistics
export const getPollingStats = () => {
  return {
    active: isPollingActive(),
    state: connectionState,
    pollCounter,
    consecutiveFailures,
    lastSuccessfulPoll,
    interval: POLL_INTERVAL,
    backendUrl: BACKEND_URL,
    listeners: Object.keys(listeners).length
  };
};

// Force a manual poll (useful for debugging)
export const forcePoll = async () => {
  console.log('🔄 Forcing manual poll...');
  await performPoll();
};

// Reset polling state (useful for debugging)
export const resetPolling = () => {
  stopPolling();
  consecutiveFailures = 0;
  pollCounter = 0;
  lastSuccessfulPoll = null;
  lastUserActivity = Date.now();
  currentPollInterval = POLL_INTERVAL;
  setTimeout(startPolling, 1000);
};

// Adaptive polling setup to reduce rate limiting
const setupAdaptivePolling = () => {
  // Track page visibility
  document.addEventListener('visibilitychange', () => {
    isPageVisible = !document.hidden;
    console.log(`📱 Page visibility changed: ${isPageVisible ? 'visible' : 'hidden'}`);
    scheduleNextPoll();
  });

  // Track user activity
  const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart', 'click'];
  const updateActivity = () => {
    lastUserActivity = Date.now();
  };
  
  activityEvents.forEach(event => {
    document.addEventListener(event, updateActivity, true);
  });

  // Start adaptive polling
  scheduleNextPoll();
};

// Calculate adaptive poll interval based on user activity and page visibility
const getAdaptivePollInterval = () => {
  if (!isPageVisible) {
    return POLL_INTERVAL * 4; // 20 seconds when page hidden
  }
  
  const timeSinceActivity = Date.now() - lastUserActivity;
  
  if (timeSinceActivity < 30000) { // Active in last 30 seconds
    return POLL_INTERVAL; // 5 seconds
  } else if (timeSinceActivity < 300000) { // Active in last 5 minutes
    return POLL_INTERVAL * 2; // 10 seconds
  } else {
    return POLL_INTERVAL * 3; // 15 seconds for idle users
  }
};

// Schedule next poll with adaptive timing
const scheduleNextPoll = () => {
  if (adaptiveTimer) {
    clearTimeout(adaptiveTimer);
  }

  currentPollInterval = getAdaptivePollInterval();
  
  adaptiveTimer = setTimeout(async () => {
    await performPoll();
    if (connectionState !== 'disconnected') {
      scheduleNextPoll(); // Schedule next poll
    }
  }, currentPollInterval);
};

// Get enhanced polling statistics including adaptive info
export const getEnhancedPollingStats = () => {
  return {
    ...getPollingStats(),
    adaptive: ADAPTIVE_POLLING,
    currentInterval: currentPollInterval,
    isPageVisible,
    lastUserActivity: new Date(lastUserActivity).toLocaleTimeString(),
    timeSinceActivity: Date.now() - lastUserActivity
  };
};

// Default export
export default {
  startPolling,
  stopPolling,
  addPollingListener,
  removePollingListener,
  getPollingConnectionState,
  isPollingActive,
  getPollingStats,
  getEnhancedPollingStats,
  forcePoll,
  resetPolling
};