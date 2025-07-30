// Startup utilities to prevent app freezing

// Check and clean corrupted localStorage
export const cleanCorruptedStorage = () => {
  console.log('🔍 Checking localStorage integrity...');
  
  const keysToCheck = ['token', 'userId', 'user'];
  const corruptedKeys = [];
  
  keysToCheck.forEach(key => {
    try {
      const value = localStorage.getItem(key);
      if (value) {
        // Try to parse if it looks like JSON
        if (value.startsWith('{') || value.startsWith('[')) {
          JSON.parse(value);
        }
      }
    } catch (error) {
      console.error(`❌ Corrupted localStorage key: ${key}`, error);
      corruptedKeys.push(key);
    }
  });
  
  // Remove corrupted keys
  corruptedKeys.forEach(key => {
    localStorage.removeItem(key);
    console.log(`🗑️ Removed corrupted key: ${key}`);
  });
  
  // Check for infinity values in all localStorage
  Object.keys(localStorage).forEach(key => {
    const value = localStorage.getItem(key);
    if (value && (value.includes('Infinity') || value.includes('NaN'))) {
      console.error(`❌ Found infinity/NaN in key: ${key}`);
      localStorage.removeItem(key);
    }
  });
  
  return corruptedKeys.length > 0;
};

// Version check for localStorage schema
export const checkStorageVersion = () => {
  const CURRENT_VERSION = '1.0';
  const storedVersion = localStorage.getItem('app_version');
  
  if (!storedVersion || storedVersion !== CURRENT_VERSION) {
    console.log('🔄 Storage version mismatch, clearing old data...');
    
    // Keep auth tokens but clear other data
    const token = localStorage.getItem('token');
    const userId = localStorage.getItem('userId');
    
    localStorage.clear();
    
    // Restore auth if valid
    if (token && userId) {
      localStorage.setItem('token', token);
      localStorage.setItem('userId', userId);
    }
    
    localStorage.setItem('app_version', CURRENT_VERSION);
    return true;
  }
  
  return false;
};

// Emergency recovery function
export const emergencyRecovery = () => {
  console.warn('🚨 Emergency recovery initiated');
  
  try {
    // Clear all storage
    localStorage.clear();
    sessionStorage.clear();
    
    // Clear all cookies
    document.cookie.split(";").forEach(c => {
      document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
    });
    
    // Remove any service workers
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.getRegistrations().then(registrations => {
        registrations.forEach(registration => {
          registration.unregister();
        });
      });
    }
    
    console.log('✅ Emergency recovery complete');
    return true;
  } catch (error) {
    console.error('❌ Emergency recovery failed:', error);
    return false;
  }
};

// Initialize app with safety checks
export const initializeApp = () => {
  console.log('🚀 Initializing Office Stonks...');
  
  // Add global error handler
  window.addEventListener('error', (event) => {
    console.error('Global error:', event.error);
    
    // Check for specific error patterns
    if (event.error?.message?.includes('Infinity') || 
        event.error?.message?.includes('NaN') ||
        event.error?.message?.includes('JSON')) {
      console.warn('🔧 Attempting automatic recovery...');
      cleanCorruptedStorage();
    }
  });
  
  // Add unhandled promise rejection handler
  window.addEventListener('unhandledrejection', (event) => {
    console.error('Unhandled promise rejection:', event.reason);
  });
  
  // Add keyboard shortcut for emergency recovery (Ctrl+Shift+R)
  let keys = {};
  window.addEventListener('keydown', (e) => {
    keys[e.key] = true;
    
    if (keys['Control'] && keys['Shift'] && keys['R']) {
      e.preventDefault();
      if (window.confirm('Emergency Recovery: This will clear all local data and restart the app. Continue?')) {
        emergencyRecovery();
        window.location.href = '/login';
      }
    }
  });
  
  window.addEventListener('keyup', (e) => {
    keys[e.key] = false;
  });
  
  // Check and clean storage
  const hadCorruption = cleanCorruptedStorage();
  const versionChanged = checkStorageVersion();
  
  if (hadCorruption || versionChanged) {
    console.log('🔄 Storage was cleaned, reloading app...');
    window.location.reload();
    return false;
  }
  
  return true;
};