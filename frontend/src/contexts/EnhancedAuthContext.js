import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { isAuthenticated, getToken, logout } from '../services/auth';
import LoadingScreen from '../components/LoadingScreen';

const EnhancedAuthContext = createContext({
  isAuthenticated: false,
  isLoading: true,
  token: null,
  error: null,
  validateToken: () => {},
  clearAuth: () => {},
  retryAuth: () => {}
});

export const useEnhancedAuth = () => {
  const context = useContext(EnhancedAuthContext);
  if (!context) {
    throw new Error('useEnhancedAuth must be used within an EnhancedAuthProvider');
  }
  return context;
};

// JWT token decoder (simple base64 decode)
const decodeToken = (token) => {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error('Failed to decode token:', error);
    return null;
  }
};

// Check if token is expired
const isTokenExpired = (token) => {
  const decoded = decodeToken(token);
  if (!decoded || !decoded.exp) return true;
  
  // Check if token expires in next 5 minutes
  const expiryTime = decoded.exp * 1000; // Convert to milliseconds
  const currentTime = Date.now();
  const bufferTime = 5 * 60 * 1000; // 5 minutes buffer
  
  return currentTime + bufferTime >= expiryTime;
};

export const EnhancedAuthProvider = ({ children }) => {
  const [state, setState] = useState({
    isAuthenticated: false,
    isLoading: true,
    token: null,
    error: null
  });

  const validateStoredData = useCallback(() => {
    try {
      // Check localStorage for corruption
      const token = getToken();
      
      if (!token) {
        return { isValid: false, reason: 'No token found' };
      }

      // Validate token format
      if (typeof token !== 'string' || token.split('.').length !== 3) {
        return { isValid: false, reason: 'Invalid token format' };
      }

      // Check token expiration
      if (isTokenExpired(token)) {
        return { isValid: false, reason: 'Token expired' };
      }

      return { isValid: true, token };
    } catch (error) {
      console.error('Error validating stored data:', error);
      return { isValid: false, reason: 'Validation error' };
    }
  }, []);

  const clearAuth = useCallback(() => {
    console.log('Clearing authentication data...');
    try {
      localStorage.removeItem('token');
      localStorage.removeItem('userId');
      // Clear any other auth-related data
      Object.keys(localStorage).forEach(key => {
        if (key.startsWith('auth_') || key.startsWith('user_')) {
          localStorage.removeItem(key);
        }
      });
    } catch (error) {
      console.error('Error clearing auth data:', error);
    }
    
    setState({
      isAuthenticated: false,
      isLoading: false,
      token: null,
      error: null
    });
  }, []);

  const validateToken = useCallback(async () => {
    setState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      const validation = validateStoredData();
      
      if (!validation.isValid) {
        console.log('Token validation failed:', validation.reason);
        clearAuth();
        return false;
      }

      // Optionally validate with backend
      // const response = await fetch('/api/auth/validate', {
      //   headers: { Authorization: `Bearer ${validation.token}` }
      // });
      // if (!response.ok) throw new Error('Token validation failed');

      setState({
        isAuthenticated: true,
        isLoading: false,
        token: validation.token,
        error: null
      });
      
      return true;
    } catch (error) {
      console.error('Token validation error:', error);
      clearAuth();
      setState(prev => ({
        ...prev,
        isLoading: false,
        error: 'Authentication failed'
      }));
      return false;
    }
  }, [validateStoredData, clearAuth]);

  const retryAuth = useCallback(() => {
    validateToken();
  }, [validateToken]);

  // Initial authentication check
  useEffect(() => {
    const initAuth = async () => {
      // Add small delay to prevent flash
      await new Promise(resolve => setTimeout(resolve, 100));
      
      const isAuth = isAuthenticated();
      if (isAuth) {
        await validateToken();
      } else {
        setState(prev => ({ ...prev, isLoading: false }));
      }
    };

    initAuth();
  }, [validateToken]);

  // Periodic token validation (every 5 minutes)
  useEffect(() => {
    if (!state.isAuthenticated) return;

    const interval = setInterval(() => {
      const token = getToken();
      if (token && isTokenExpired(token)) {
        console.log('Token expired, logging out...');
        logout();
      }
    }, 5 * 60 * 1000); // 5 minutes

    return () => clearInterval(interval);
  }, [state.isAuthenticated]);

  // Handle authentication timeout
  const handleAuthTimeout = useCallback(() => {
    setState(prev => ({
      ...prev,
      isLoading: false,
      error: 'Authentication check timed out'
    }));
  }, []);

  const value = {
    ...state,
    validateToken,
    clearAuth,
    retryAuth
  };

  // Show loading screen during initial auth check
  if (state.isLoading) {
    return (
      <EnhancedAuthContext.Provider value={value}>
        <LoadingScreen 
          message="Authenticating..." 
          timeout={10000}
          onTimeout={handleAuthTimeout}
        />
      </EnhancedAuthContext.Provider>
    );
  }

  return (
    <EnhancedAuthContext.Provider value={value}>
      {children}
    </EnhancedAuthContext.Provider>
  );
};