// Centralized utility functions for the Office Stonks application
// Consolidates common functions used across multiple components

// Import storage utilities for re-export
import { safeGetItem, safeSetItem, safeRemoveItem } from '../services/storageManager';

// =============================================================================
// CURRENCY AND NUMBER FORMATTING
// =============================================================================

/**
 * Format a number as currency with proper comma separators
 * @param {number} value - The numeric value to format
 * @param {number} decimals - Number of decimal places (default: 2)
 * @param {string} currency - Currency symbol (default: '$')
 * @returns {string} Formatted currency string
 */
export const formatCurrency = (value, decimals = 2, currency = '$') => {
  if (value === null || value === undefined || isNaN(value)) {
    return `${currency}0.00`;
  }
  
  const numValue = parseFloat(value);
  return `${currency}${numValue.toFixed(decimals).replace(/\d(?=(\d{3})+\.)/g, '$&,')}`;
};

/**
 * Format a percentage value with proper sign and styling
 * @param {number} value - The percentage value
 * @param {number} decimals - Number of decimal places (default: 2)
 * @returns {string} Formatted percentage string with sign
 */
export const formatPercentage = (value, decimals = 2) => {
  if (value === null || value === undefined || isNaN(value)) {
    return '0.00%';
  }
  
  const numValue = parseFloat(value);
  const sign = numValue >= 0 ? '+' : '';
  return `${sign}${numValue.toFixed(decimals)}%`;
};

/**
 * Format a large number with appropriate suffixes (K, M, B)
 * @param {number} value - The numeric value to format
 * @param {number} decimals - Number of decimal places (default: 1)
 * @returns {string} Formatted number with suffix
 */
export const formatLargeNumber = (value, decimals = 1) => {
  if (value === null || value === undefined || isNaN(value)) {
    return '0';
  }
  
  const numValue = Math.abs(parseFloat(value));
  const sign = parseFloat(value) < 0 ? '-' : '';
  
  if (numValue >= 1000000000) {
    return `${sign}${(numValue / 1000000000).toFixed(decimals)}B`;
  }
  if (numValue >= 1000000) {
    return `${sign}${(numValue / 1000000).toFixed(decimals)}M`;
  }
  if (numValue >= 1000) {
    return `${sign}${(numValue / 1000).toFixed(decimals)}K`;
  }
  
  return `${sign}${numValue.toFixed(decimals)}`;
};

// =============================================================================
// DATE AND TIME FORMATTING
// =============================================================================

/**
 * Format a timestamp to a readable date string
 * @param {string|number|Date} timestamp - The timestamp to format
 * @param {object} options - Intl.DateTimeFormat options
 * @returns {string} Formatted date string
 */
export const formatDate = (timestamp, options = {}) => {
  if (!timestamp) return 'N/A';
  
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return 'Invalid Date';
  
  const defaultOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...options
  };
  
  return new Intl.DateTimeFormat('en-US', defaultOptions).format(date);
};

/**
 * Format a timestamp to a readable time string
 * @param {string|number|Date} timestamp - The timestamp to format
 * @param {boolean} includeSeconds - Whether to include seconds (default: false)
 * @returns {string} Formatted time string
 */
export const formatTime = (timestamp, includeSeconds = false) => {
  if (!timestamp) return 'N/A';
  
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return 'Invalid Time';
  
  const options = {
    hour: '2-digit',
    minute: '2-digit',
    ...(includeSeconds && { second: '2-digit' })
  };
  
  return new Intl.DateTimeFormat('en-US', options).format(date);
};

/**
 * Format a timestamp to a relative time string (e.g., "2 minutes ago")
 * @param {string|number|Date} timestamp - The timestamp to format
 * @returns {string} Relative time string
 */
export const formatRelativeTime = (timestamp) => {
  if (!timestamp) return 'N/A';
  
  const now = new Date();
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return 'Invalid Date';
  
  const diffMs = now - date;
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffSecs < 60) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  
  return formatDate(timestamp);
};

// =============================================================================
// STRING UTILITIES
// =============================================================================

/**
 * Capitalize the first letter of a string
 * @param {string} str - The string to capitalize
 * @returns {string} Capitalized string
 */
export const capitalize = (str) => {
  if (!str || typeof str !== 'string') return '';
  return str.charAt(0).toUpperCase() + str.slice(1);
};

/**
 * Convert a string to title case
 * @param {string} str - The string to convert
 * @returns {string} Title case string
 */
export const toTitleCase = (str) => {
  if (!str || typeof str !== 'string') return '';
  return str.replace(/\w\S*/g, (txt) => 
    txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase()
  );
};

/**
 * Truncate a string to a specified length with ellipsis
 * @param {string} str - The string to truncate
 * @param {number} maxLength - Maximum length before truncation
 * @param {string} suffix - Suffix to add when truncated (default: '...')
 * @returns {string} Truncated string
 */
export const truncateString = (str, maxLength, suffix = '...') => {
  if (!str || typeof str !== 'string') return '';
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength - suffix.length) + suffix;
};

// =============================================================================
// VALIDATION UTILITIES
// =============================================================================

/**
 * Check if a value is a valid number
 * @param {any} value - The value to check
 * @returns {boolean} True if valid number
 */
export const isValidNumber = (value) => {
  return value !== null && value !== undefined && !isNaN(parseFloat(value)) && isFinite(value);
};

/**
 * Check if a value is a valid positive number
 * @param {any} value - The value to check
 * @returns {boolean} True if valid positive number
 */
export const isValidPositiveNumber = (value) => {
  return isValidNumber(value) && parseFloat(value) > 0;
};

/**
 * Check if a string is a valid email address
 * @param {string} email - The email to validate
 * @returns {boolean} True if valid email
 */
export const isValidEmail = (email) => {
  if (!email || typeof email !== 'string') return false;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

// =============================================================================
// ARRAY AND OBJECT UTILITIES
// =============================================================================

/**
 * Deep clone an object or array
 * @param {any} obj - The object to clone
 * @returns {any} Deep cloned object
 */
export const deepClone = (obj) => {
  if (obj === null || typeof obj !== 'object') return obj;
  if (obj instanceof Date) return new Date(obj.getTime());
  if (obj instanceof Array) return obj.map(item => deepClone(item));
  if (typeof obj === 'object') {
    const cloned = {};
    Object.keys(obj).forEach(key => {
      cloned[key] = deepClone(obj[key]);
    });
    return cloned;
  }
  return obj;
};

/**
 * Sort an array of objects by a specified key
 * @param {Array} array - The array to sort
 * @param {string} key - The key to sort by
 * @param {string} direction - Sort direction ('asc' or 'desc')
 * @returns {Array} Sorted array
 */
export const sortByKey = (array, key, direction = 'asc') => {
  if (!Array.isArray(array)) return [];
  
  return [...array].sort((a, b) => {
    const aVal = a[key];
    const bVal = b[key];
    
    if (aVal < bVal) return direction === 'asc' ? -1 : 1;
    if (aVal > bVal) return direction === 'asc' ? 1 : -1;
    return 0;
  });
};

/**
 * Group an array of objects by a specified key
 * @param {Array} array - The array to group
 * @param {string} key - The key to group by
 * @returns {Object} Object with grouped arrays
 */
export const groupByKey = (array, key) => {
  if (!Array.isArray(array)) return {};
  
  return array.reduce((groups, item) => {
    const groupKey = item[key];
    if (!groups[groupKey]) groups[groupKey] = [];
    groups[groupKey].push(item);
    return groups;
  }, {});
};

// =============================================================================
// URL AND NAVIGATION UTILITIES
// =============================================================================

/**
 * Get URL parameters as an object
 * @param {string} url - The URL to parse (default: current URL)
 * @returns {Object} Object containing URL parameters
 */
export const getUrlParams = (url = window.location.search) => {
  const params = new URLSearchParams(url);
  const result = {};
  for (const [key, value] of params.entries()) {
    result[key] = value;
  }
  return result;
};

/**
 * Build a URL with parameters
 * @param {string} baseUrl - The base URL
 * @param {Object} params - Parameters to add
 * @returns {string} Complete URL with parameters
 */
export const buildUrl = (baseUrl, params = {}) => {
  const url = new URL(baseUrl, window.location.origin);
  Object.entries(params).forEach(([key, value]) => {
    if (value !== null && value !== undefined) {
      url.searchParams.set(key, value);
    }
  });
  return url.toString();
};

// =============================================================================
// PERFORMANCE UTILITIES
// =============================================================================

/**
 * Debounce a function call
 * @param {Function} func - The function to debounce
 * @param {number} delay - Delay in milliseconds
 * @returns {Function} Debounced function
 */
export const debounce = (func, delay) => {
  let timeoutId;
  return (...args) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func.apply(null, args), delay);
  };
};

/**
 * Throttle a function call
 * @param {Function} func - The function to throttle
 * @param {number} limit - Time limit in milliseconds
 * @returns {Function} Throttled function
 */
export const throttle = (func, limit) => {
  let inThrottle;
  return (...args) => {
    if (!inThrottle) {
      func.apply(null, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
};

// =============================================================================
// STORAGE UTILITIES (re-export from storageManager for convenience)
// =============================================================================

export { safeGetItem, safeSetItem, safeRemoveItem };

// =============================================================================
// CONSTANTS AND DEFAULTS
// =============================================================================

export const CONSTANTS = {
  // Currency formatting
  DEFAULT_CURRENCY_SYMBOL: '$',
  DEFAULT_DECIMAL_PLACES: 2,
  
  // Date formatting
  DEFAULT_DATE_FORMAT: 'MMM DD, YYYY',
  DEFAULT_TIME_FORMAT: 'HH:mm',
  
  // Validation
  MIN_PASSWORD_LENGTH: 6,
  MAX_USERNAME_LENGTH: 20,
  
  // UI
  DEFAULT_PAGE_SIZE: 25,
  MOBILE_BREAKPOINT: 768,
  
  // Performance
  DEFAULT_DEBOUNCE_DELAY: 300,
  DEFAULT_THROTTLE_LIMIT: 1000
};

// Export all utilities as default for easy importing
const utils = {
  formatCurrency,
  formatPercentage,
  formatLargeNumber,
  formatDate,
  formatTime,
  formatRelativeTime,
  capitalize,
  toTitleCase,
  truncateString,
  isValidNumber,
  isValidPositiveNumber,
  isValidEmail,
  deepClone,
  sortByKey,
  groupByKey,
  getUrlParams,
  buildUrl,
  debounce,
  throttle,
  safeGetItem,
  safeSetItem,
  safeRemoveItem,
  CONSTANTS
};

export default utils;