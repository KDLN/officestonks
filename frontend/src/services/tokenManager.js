// Token management service for proactive token validation and refresh
import { safeGetItem, safeSetItem, safeRemoveItem } from '../utils';

// Token validation constants
const TOKEN_REFRESH_THRESHOLD = 2 * 60 * 60 * 1000; // Refresh if expires within 2 hours
const TOKEN_CHECK_INTERVAL = 10 * 60 * 1000; // Check token every 10 minutes

class TokenManager {
  constructor() {
    this.checkInterval = null;
    this.isRefreshing = false;
    this.refreshPromise = null;
  }

  // Initialize token monitoring
  startTokenMonitoring() {
    console.log('🔐 Starting token monitoring service...');
    
    // Initial check
    this.checkTokenExpiration();
    
    // Set up periodic checks
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
    }
    
    this.checkInterval = setInterval(() => {
      this.checkTokenExpiration();
    }, TOKEN_CHECK_INTERVAL);
  }

  // Stop token monitoring
  stopTokenMonitoring() {
    console.log('🔐 Stopping token monitoring service...');
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = null;
    }
  }


  // Parse JWT token to get expiration
  parseTokenExpiration(token) {
    if (!token) return null;
    
    try {
      // JWT tokens have three parts separated by dots
      const parts = token.split('.');
      if (parts.length !== 3) return null;
      
      // Decode the payload (second part)
      const payload = JSON.parse(atob(parts[1]));
      
      // Return expiration timestamp (convert from seconds to milliseconds)
      return payload.exp ? payload.exp * 1000 : null;
    } catch (error) {
      console.error('🔐 Error parsing token:', error);
      return null;
    }
  }

  // Get current token expiration status
  getTokenStatus() {
    const token = safeGetItem('token');
    if (!token) {
      return { valid: false, reason: 'no_token' };
    }

    const expirationTime = this.parseTokenExpiration(token);
    if (!expirationTime) {
      return { valid: false, reason: 'invalid_token' };
    }

    const now = Date.now();
    const timeUntilExpiry = expirationTime - now;

    return {
      valid: timeUntilExpiry > 0,
      expiresAt: expirationTime,
      timeUntilExpiry,
      needsRefresh: timeUntilExpiry < TOKEN_REFRESH_THRESHOLD,
      reason: timeUntilExpiry <= 0 ? 'expired' : 'valid'
    };
  }

  // Check token expiration and take appropriate action
  async checkTokenExpiration() {
    const status = this.getTokenStatus();
    
    if (!status.valid) {
      console.log('🔐 Token invalid or missing:', status.reason);
      if (status.reason === 'expired') {
        await this.handleTokenExpiration();
      }
      return;
    }

    const minutesUntilExpiry = Math.floor(status.timeUntilExpiry / (60 * 1000));
    
    // Silently refresh token if needed
    if (status.needsRefresh && !this.isRefreshing) {
      console.log(`🔐 Token refresh needed: expires in ${minutesUntilExpiry} minutes`);
      await this.refreshToken();
    }
  }

  // Attempt to refresh the token
  async refreshToken() {
    if (this.isRefreshing) {
      return this.refreshPromise;
    }

    this.isRefreshing = true;
    console.log('🔐 Attempting token refresh...');

    this.refreshPromise = this.performTokenRefresh();
    
    try {
      const result = await this.refreshPromise;
      this.isRefreshing = false;
      return result;
    } catch (error) {
      this.isRefreshing = false;
      throw error;
    }
  }

  // Perform the actual token refresh
  async performTokenRefresh() {
    const currentToken = safeGetItem('token');
    if (!currentToken) {
      throw new Error('No token to refresh');
    }

    try {
      const API_URL = process.env.REACT_APP_BACKEND_URL || 'https://officestonks.com';
      const response = await fetch(`${API_URL}/api/auth/refresh`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${currentToken}`,
          'Content-Type': 'application/json'
        },
        credentials: 'include'
      });

      if (!response.ok) {
        throw new Error(`Token refresh failed: ${response.status}`);
      }

      const data = await response.json();
      
      // Update stored token
      safeSetItem('token', data.token);
      
      console.log('🔐 Token refreshed successfully');
      
      return data.token;
    } catch (error) {
      console.error('🔐 Token refresh failed:', error);
      throw error;
    }
  }

  // Handle token expiration
  async handleTokenExpiration() {
    console.log('🔐 Handling token expiration...');
    
    // Try to refresh first
    try {
      await this.refreshToken();
      console.log('🔐 Token successfully refreshed after expiration');
      return;
    } catch (error) {
      console.log('🔐 Token refresh failed, logging out user');
    }
    
    // If refresh fails, logout
    this.stopTokenMonitoring();
    
    // Clear all auth data
    safeRemoveItem('token');
    safeRemoveItem('userId');
    safeRemoveItem('username');
    safeRemoveItem('isAdmin');
    
    // Redirect to login if not already there
    if (!window.location.pathname.includes('/login')) {
      console.log('🔐 Redirecting to login due to token expiration');
      window.location.href = '/login';
    }
  }

  // Manually validate token (for critical operations)
  async validateToken() {
    const status = this.getTokenStatus();
    
    if (!status.valid) {
      if (status.reason === 'expired') {
        await this.handleTokenExpiration();
      }
      return false;
    }
    
    // If token expires soon, try to refresh
    if (status.needsRefresh) {
      try {
        await this.refreshToken();
        return true;
      } catch (error) {
        console.error('🔐 Token validation refresh failed:', error);
        return false;
      }
    }
    
    return true;
  }

  // Get human-readable time until expiry
  getTimeUntilExpiryString() {
    const status = this.getTokenStatus();
    if (!status.valid) return 'Invalid token';
    
    const minutes = Math.floor(status.timeUntilExpiry / (60 * 1000));
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes % 60}m`;
    } else {
      return `${minutes}m`;
    }
  }
}

// Create singleton instance
const tokenManager = new TokenManager();

export default tokenManager;

// Export utility functions
export const startTokenMonitoring = () => tokenManager.startTokenMonitoring();
export const stopTokenMonitoring = () => tokenManager.stopTokenMonitoring();
export const getTokenStatus = () => tokenManager.getTokenStatus();
export const validateToken = () => tokenManager.validateToken();
export const refreshToken = () => tokenManager.refreshToken();