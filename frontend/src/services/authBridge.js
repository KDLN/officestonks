import { getCurrentSession } from './supabaseAuth'
import { getToken as getOfficeToken } from './auth'
import { validateToken } from './tokenManager'
import logger from './logger'

const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com'

// Sync Supabase session with Office Stonks backend
export const syncAuthWithBackend = async () => {
  const apiLogger = logger.apiCall('POST', '/api/auth/supabase');
  
  logger.info('Starting auth sync with backend');
  const session = await getCurrentSession()
  
  if (!session?.access_token) {
    logger.warn('No Supabase session found for backend sync');
    return null
  }

  logger.debug('Supabase session found, syncing with backend');
  try {
    const startTime = Date.now();
    // Send Supabase token to backend for validation and user creation/sync
    const response = await fetch(`${API_URL}/api/auth/supabase`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${session.access_token}`
      },
      credentials: 'include'
    })

    const syncTime = Date.now() - startTime;
    
    if (!response.ok) {
      const errorText = await response.text();
      apiLogger.error(new Error('Backend sync failed'), {
        status: response.status,
        statusText: response.statusText,
        errorText,
        syncTime
      });
      throw new Error('Failed to sync with backend')
    }

    const data = await response.json()
    apiLogger.success(response, {
      hasToken: !!data.token,
      userID: data.userID,
      username: data.username,
      isAdmin: data.isAdmin,
      syncTime
    });
    
    // Store the Office Stonks token for game API calls
    const { safeSetItem } = await import('./storageManager');
    safeSetItem('token', data.token);
    safeSetItem('userId', data.userID);
    safeSetItem('username', data.username);
    safeSetItem('isAdmin', data.isAdmin);
    
    return data
  } catch (error) {
    logger.error('Auth sync error', { 
      error: error.message,
      stack: error.stack
    });
    throw error
  }
}

// Get auth token (prefers Office Stonks token after sync, falls back to Supabase)
export const getAuthToken = async () => {
  // First check if we have an Office Stonks token from auth sync
  const officeToken = getOfficeToken()
  if (officeToken) {
    logger.trace('Using Office Stonks token for authentication');
    return officeToken
  }
  
  // Fall back to Supabase session token
  const session = await getCurrentSession()
  if (session?.access_token) {
    logger.trace('Using Supabase session token for authentication');
    return session.access_token
  }
  
  logger.warn('No authentication token available');
  return null
}

// Enhanced fetch that includes proper authentication
let syncPromise = null

// Helper function to clear stored auth credentials
const clearStoredCredentials = async () => {
  const { safeRemoveItem } = await import('./storageManager');
  safeRemoveItem('token');
  safeRemoveItem('userId');
  safeRemoveItem('username');
  safeRemoveItem('isAdmin');
};

// Helper function to create headers with authentication
const createAuthHeaders = (token, options = {}) => ({
  ...options.headers,
  'Authorization': `Bearer ${token}`
});

// Helper function to make authenticated request
const makeRequest = async (url, options, token) => {
  const headers = createAuthHeaders(token, options);
  
  logger.debug('Making authenticated request', { url, method: options.method || 'GET' });
  
  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include'
  });

  logger.debug('Response received', {
    url,
    status: response.status,
    statusText: response.statusText,
    ok: response.ok
  });

  return response;
};

// Helper function to handle auth sync and retry
const handleAuthFailureAndRetry = async (url, options) => {
  logger.warn('Auth failed, attempting token refresh');
  
  // Clear stored credentials
  await clearStoredCredentials();

  try {
    // Attempt to resync auth from Supabase session (deduplicated)
    if (!syncPromise) {
      logger.debug('Starting new auth sync');
      syncPromise = syncAuthWithBackend().finally(() => { syncPromise = null });
    } else {
      logger.debug('Auth sync already in progress, waiting');
    }
    
    await syncPromise;

    const newToken = await getAuthToken();
    if (newToken) {
      logger.debug('Retrying request with new token');
      const retryResponse = await makeRequest(url, options, newToken);
      
      if (retryResponse.ok) {
        return retryResponse;
      }
    }
  } catch (err) {
    logger.error('Auth resync failed', { error: err.message });
  }

  // Redirect to login if we still have no valid token
  logger.warn('Redirecting to login due to auth failure');
  if (!window.location.pathname.includes('/login')) {
    window.location.href = '/login';
  }
  
  throw new Error('Authentication failed - redirecting to login');
};

// Main authenticated fetch function (simplified)
export const authenticatedFetch = async (url, options = {}) => {
  const apiLogger = logger.apiCall(options.method || 'GET', url);
  
  try {
    // Validate token before making request
    const tokenValid = await validateToken();
    if (!tokenValid) {
      apiLogger.error(new Error('Token validation failed before request'));
      throw new Error('Token validation failed');
    }
    
    const token = await getAuthToken();
    if (!token) {
      apiLogger.error(new Error('No authentication token available'));
      throw new Error('No authentication token available');
    }

    // Make the authenticated request
    const response = await makeRequest(url, options, token);

    // Handle auth failures with retry logic
    if (response.status === 401 || response.status === 403) {
      return await handleAuthFailureAndRetry(url, options);
    }

    // Success case
    apiLogger.success(response);
    return response;
    
  } catch (error) {
    apiLogger.error(error);
    throw error;
  }
}
