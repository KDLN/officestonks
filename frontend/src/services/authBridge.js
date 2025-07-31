import { getCurrentSession } from './supabaseAuth'
import { getToken as getOfficeToken } from './auth'
import { validateToken } from './tokenManager'

const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com'

// Sync Supabase session with Office Stonks backend
export const syncAuthWithBackend = async () => {
  console.log('🔄 Starting auth sync with backend...');
  const session = await getCurrentSession()
  
  if (!session?.access_token) {
    console.log('❌ No Supabase session found for backend sync');
    return null
  }

  console.log('✅ Supabase session found, syncing with backend...');
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
    console.log(`🔄 Backend sync response in ${syncTime}ms:`, {
      status: response.status,
      statusText: response.statusText,
      ok: response.ok
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('🔄 Backend sync failed:', errorText);
      throw new Error('Failed to sync with backend')
    }

    const data = await response.json()
    console.log('✅ Backend sync successful:', {
      hasToken: !!data.token,
      userID: data.userID,
      username: data.username,
      isAdmin: data.isAdmin
    });
    
    // Store the Office Stonks token for game API calls
    localStorage.setItem('token', data.token)
    localStorage.setItem('userId', data.userID)
    localStorage.setItem('username', data.username)
    localStorage.setItem('isAdmin', data.isAdmin)
    
    return data
  } catch (error) {
    console.error('🔄 Auth sync error:', error)
    throw error
  }
}

// Get auth token (prefers Office Stonks token after sync, falls back to Supabase)
export const getAuthToken = async () => {
  // First check if we have an Office Stonks token from auth sync
  const officeToken = getOfficeToken()
  if (officeToken) {
    console.log('🔑 Using Office Stonks token for authentication');
    return officeToken
  }
  
  // Fall back to Supabase session token
  const session = await getCurrentSession()
  if (session?.access_token) {
    console.log('🔑 Using Supabase session token for authentication');
    return session.access_token
  }
  
  console.log('❌ No authentication token available');
  return null
}

// Enhanced fetch that includes proper authentication
let syncPromise = null

export const authenticatedFetch = async (url, options = {}) => {
  console.log('🌐 Making authenticated request to:', url);
  
  // Validate token before making request
  const tokenValid = await validateToken();
  if (!tokenValid) {
    console.error('🌐 Token validation failed before request');
    throw new Error('Token validation failed')
  }
  
  let token = await getAuthToken()

  if (!token) {
    console.error('🌐 No authentication token available for request');
    throw new Error('No authentication token available')
  }

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`
  }

  console.log('🌐 Sending request with auth headers...');
  let response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include'
  })

  console.log('🌐 Response received:', {
    url: url,
    status: response.status,
    statusText: response.statusText,
    ok: response.ok
  });

  if (response.status === 401 || response.status === 403) {
    console.log('🔄 Auth failed, attempting token refresh...');
    // Token might be invalid or expired - clear stored credentials
    localStorage.removeItem('token')
    localStorage.removeItem('userId')
    localStorage.removeItem('username')
    localStorage.removeItem('isAdmin')

    try {
      // Attempt to resync auth from Supabase session (deduplicated)
      if (!syncPromise) {
        console.log('🔄 Starting new auth sync...');
        syncPromise = syncAuthWithBackend().finally(() => { syncPromise = null })
      } else {
        console.log('🔄 Auth sync already in progress, waiting...');
      }
      await syncPromise

      token = await getAuthToken()
      if (token) {
        console.log('🔄 Retrying request with new token...');
        const retryHeaders = {
          ...options.headers,
          'Authorization': `Bearer ${token}`
        }
        response = await fetch(url, {
          ...options,
          headers: retryHeaders,
          credentials: 'include'
        })

        console.log('🔄 Retry response:', {
          status: response.status,
          statusText: response.statusText,
          ok: response.ok
        });

        if (response.ok) {
          return response
        }
      }
    } catch (err) {
      console.error('🔄 Auth resync failed:', err)
    }

    // Redirect to login if we still have no valid token
    console.log('🚪 Redirecting to login due to auth failure');
    if (!window.location.pathname.includes('/login')) {
      window.location.href = '/login'
    }
    return Promise.reject(new Error('Authentication failed - redirecting to login'))
  }

  return response
}
