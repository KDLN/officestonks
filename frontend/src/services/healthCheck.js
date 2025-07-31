// Health check service to verify server status
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';
const API_URL = `${BACKEND_URL}/api`;

let healthCheckInterval = null;
let lastHealthStatus = null;
let healthListeners = [];

// Health status object
const createHealthStatus = (isHealthy, details = {}) => ({
  isHealthy,
  timestamp: Date.now(),
  ...details
});

// Check server health
export const checkServerHealth = async () => {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

    const response = await fetch(`${API_URL}/health`, {
      method: 'GET',
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      throw new Error(`Health check failed: ${response.status}`);
    }

    const data = await response.json();
    const status = createHealthStatus(true, {
      response: data,
      responseTime: Date.now() - controller.signal.timestamp
    });

    lastHealthStatus = status;
    notifyListeners(status);
    
    return status;
  } catch (error) {
    console.error('Health check error:', error);
    
    const status = createHealthStatus(false, {
      error: error.message,
      errorType: error.name === 'AbortError' ? 'timeout' : 'network'
    });

    lastHealthStatus = status;
    notifyListeners(status);
    
    return status;
  }
};

// Start periodic health checks
export const startHealthChecks = (interval = 30000) => { // 30 seconds default
  if (healthCheckInterval) {
    clearInterval(healthCheckInterval);
  }

  // Initial check
  checkServerHealth();

  // Periodic checks
  healthCheckInterval = setInterval(() => {
    checkServerHealth();
  }, interval);
};

// Stop health checks
export const stopHealthChecks = () => {
  if (healthCheckInterval) {
    clearInterval(healthCheckInterval);
    healthCheckInterval = null;
  }
};

// Get last health status
export const getLastHealthStatus = () => lastHealthStatus;

// Add health status listener
export const addHealthListener = (callback) => {
  healthListeners.push(callback);
  
  // Return unsubscribe function
  return () => {
    healthListeners = healthListeners.filter(cb => cb !== callback);
  };
};

// Notify all listeners
const notifyListeners = (status) => {
  healthListeners.forEach(callback => {
    try {
      callback(status);
    } catch (error) {
      console.error('Error in health listener:', error);
    }
  });
};

// Check if we should attempt operations based on health
export const shouldAttemptOperation = () => {
  if (!lastHealthStatus) return true; // No status yet, allow attempt
  
  // If healthy, always allow
  if (lastHealthStatus.isHealthy) return true;
  
  // If unhealthy but recent (less than 1 minute), don't attempt
  const timeSinceCheck = Date.now() - lastHealthStatus.timestamp;
  if (timeSinceCheck < 60000) return false;
  
  // If unhealthy but old status, allow attempt (might be recovered)
  return true;
};

// Utility to wait for healthy status
export const waitForHealthy = async (timeout = 30000) => {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const status = await checkServerHealth();
    if (status.isHealthy) return true;
    
    // Wait 2 seconds before next check
    await new Promise(resolve => setTimeout(resolve, 2000));
  }
  
  return false;
};