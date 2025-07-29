import { getCurrentSession } from './supabaseAuth'
import { getToken as getOfficeToken } from './auth'

const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com'

console.log('AuthBridge API_URL:', API_URL, 'env var:', process.env.REACT_APP_BACKEND_URL)

// Sync Supabase session with Office Stonks backend
export const syncAuthWithBackend = async () => {
  const session = await getCurrentSession()
  
  if (!session?.access_token) {
    return null
  }

  try {
    // Send Supabase token to backend for validation and user creation/sync
    const response = await fetch(`${API_URL}/api/auth/supabase`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${session.access_token}`
      },
      credentials: 'include'
    })

    if (!response.ok) {
      throw new Error('Failed to sync with backend')
    }

    const data = await response.json()
    
    // Store the Office Stonks token for game API calls
    localStorage.setItem('token', data.token)
    localStorage.setItem('userId', data.userID)
    localStorage.setItem('username', data.username)
    localStorage.setItem('isAdmin', data.isAdmin)
    
    return data
  } catch (error) {
    console.error('Auth sync error:', error)
    throw error
  }
}

// Get auth token (prefers Office Stonks token after sync, falls back to Supabase)
export const getAuthToken = async () => {
  // First check if we have an Office Stonks token from auth sync
  const officeToken = getOfficeToken()
  if (officeToken) {
    return officeToken
  }
  
  // Fall back to Supabase session token
  const session = await getCurrentSession()
  if (session?.access_token) {
    return session.access_token
  }
  
  return null
}

// Enhanced fetch that includes proper authentication
export const authenticatedFetch = async (url, options = {}) => {
  const token = await getAuthToken()
  
  if (!token) {
    throw new Error('No authentication token available')
  }

  const headers = {
    ...options.headers,
    'Authorization': `Bearer ${token}`
  }

  return fetch(url, {
    ...options,
    headers,
    credentials: 'include'
  })
}