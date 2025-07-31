// Centralized logging service with consistent formatting and error handling
// Provides structured logging with context, timestamps, and log levels

// Log levels with priorities
const LOG_LEVELS = {
  ERROR: { priority: 0, emoji: '❌', color: '#f44336' },
  WARN: { priority: 1, emoji: '⚠️', color: '#ff9800' },
  INFO: { priority: 2, emoji: 'ℹ️', color: '#2196f3' },
  DEBUG: { priority: 3, emoji: '🔍', color: '#9c27b0' },
  TRACE: { priority: 4, emoji: '📍', color: '#607d8b' }
};

// Current log level (can be configured via environment)
const CURRENT_LOG_LEVEL = process.env.NODE_ENV === 'production' ? 'WARN' : 'DEBUG';

class Logger {
  constructor() {
    this.sessionId = this.generateSessionId();
    this.startTime = Date.now();
    this.logs = [];
    this.maxLogsInMemory = 1000;
  }

  // Generate unique session ID
  generateSessionId() {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // Check if log level should be shown
  shouldLog(level) {
    const currentPriority = LOG_LEVELS[CURRENT_LOG_LEVEL]?.priority ?? 2;
    const logPriority = LOG_LEVELS[level]?.priority ?? 2;
    return logPriority <= currentPriority;
  }

  // Format log message with consistent structure
  formatMessage(level, message, context = {}) {
    const timestamp = new Date().toISOString();
    const emoji = LOG_LEVELS[level]?.emoji || '📝';
    const sessionTime = Date.now() - this.startTime;

    const logEntry = {
      timestamp,
      sessionTime,
      sessionId: this.sessionId,
      level,
      message,
      context,
      url: window.location.href,
      userAgent: navigator.userAgent.split(' ')[0] // Just browser name
    };

    // Store in memory for debugging
    this.logs.push(logEntry);
    if (this.logs.length > this.maxLogsInMemory) {
      this.logs.shift();
    }

    return {
      logEntry,
      consoleMessage: `${emoji} [${level}] ${message}`,
      consoleStyle: `color: ${LOG_LEVELS[level]?.color || '#333'}; font-weight: bold;`
    };
  }

  // Core logging method
  log(level, message, context = {}) {
    if (!this.shouldLog(level)) return;

    const { logEntry, consoleMessage, consoleStyle } = this.formatMessage(level, message, context);

    // Console output with styling
    if (Object.keys(context).length > 0) {
      console.groupCollapsed(`%c${consoleMessage}`, consoleStyle);
      console.table(context);
      console.trace('Stack trace:');
      console.groupEnd();
    } else {
      console.log(`%c${consoleMessage}`, consoleStyle);
    }

    // Send critical errors to backend (if available)
    if (level === 'ERROR') {
      this.reportError(logEntry);
    }

    return logEntry;
  }

  // Convenience methods
  error(message, context = {}) {
    return this.log('ERROR', message, context);
  }

  warn(message, context = {}) {
    return this.log('WARN', message, context);
  }

  info(message, context = {}) {
    return this.log('INFO', message, context);
  }

  debug(message, context = {}) {
    return this.log('DEBUG', message, context);
  }

  trace(message, context = {}) {
    return this.log('TRACE', message, context);
  }

  // API call logging with timing
  apiCall(method, url, context = {}) {
    const startTime = performance.now();
    const apiContext = {
      method,
      url,
      timestamp: new Date().toISOString(),
      ...context
    };

    this.debug(`API ${method} ${url}`, apiContext);

    return {
      success: (response, additionalContext = {}) => {
        const duration = Math.round(performance.now() - startTime);
        this.info(`API ${method} ${url} - Success (${duration}ms)`, {
          ...apiContext,
          duration,
          status: response?.status,
          ...additionalContext
        });
      },
      error: (error, additionalContext = {}) => {
        const duration = Math.round(performance.now() - startTime);
        this.error(`API ${method} ${url} - Failed (${duration}ms)`, {
          ...apiContext,
          duration,
          error: error.message,
          stack: error.stack,
          ...additionalContext
        });
      }
    };
  }

  // User action logging
  userAction(action, context = {}) {
    this.info(`User action: ${action}`, {
      action,
      userId: localStorage.getItem('userId'),
      username: localStorage.getItem('username'),
      timestamp: new Date().toISOString(),
      ...context
    });
  }

  // Component lifecycle logging
  componentLifecycle(component, event, context = {}) {
    this.debug(`Component ${component} - ${event}`, {
      component,
      event,
      ...context
    });
  }

  // Performance measurement
  performance(label, fn) {
    const start = performance.now();
    this.trace(`Performance start: ${label}`);
    
    try {
      const result = fn();
      const duration = Math.round(performance.now() - start);
      
      if (result && typeof result.then === 'function') {
        // Handle async functions
        return result
          .then(asyncResult => {
            const asyncDuration = Math.round(performance.now() - start);
            this.info(`Performance: ${label} completed (${asyncDuration}ms)`);
            return asyncResult;
          })
          .catch(error => {
            const errorDuration = Math.round(performance.now() - start);
            this.error(`Performance: ${label} failed (${errorDuration}ms)`, { error: error.message });
            throw error;
          });
      } else {
        // Handle sync functions
        this.info(`Performance: ${label} completed (${duration}ms)`);
        return result;
      }
    } catch (error) {
      const errorDuration = Math.round(performance.now() - start);
      this.error(`Performance: ${label} failed (${errorDuration}ms)`, { error: error.message });
      throw error;
    }
  }

  // Report critical errors to backend
  async reportError(logEntry) {
    try {
      // Only report in production and if we have auth
      if (process.env.NODE_ENV !== 'production' || !localStorage.getItem('token')) {
        return;
      }

      const errorReport = {
        ...logEntry,
        userAgent: navigator.userAgent,
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight
        },
        memory: performance.memory ? {
          used: performance.memory.usedJSHeapSize,
          total: performance.memory.totalJSHeapSize,
          limit: performance.memory.jsHeapSizeLimit
        } : null
      };

      // Send to backend error reporting endpoint (if exists)
      await fetch('/api/errors/report', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(errorReport)
      }).catch(() => {
        // Silently fail if error reporting endpoint doesn't exist
      });
    } catch (error) {
      // Don't log errors about error reporting to avoid infinite loops
      console.warn('Failed to report error to backend:', error);
    }
  }

  // Get recent logs for debugging
  getRecentLogs(count = 50) {
    return this.logs.slice(-count);
  }

  // Get logs by level
  getLogsByLevel(level) {
    return this.logs.filter(log => log.level === level);
  }

  // Export logs for debugging
  exportLogs() {
    const logsData = {
      sessionId: this.sessionId,
      sessionDuration: Date.now() - this.startTime,
      totalLogs: this.logs.length,
      logLevels: Object.keys(LOG_LEVELS).reduce((acc, level) => {
        acc[level] = this.logs.filter(log => log.level === level).length;
        return acc;
      }, {}),
      logs: this.logs,
      systemInfo: {
        userAgent: navigator.userAgent,
        url: window.location.href,
        viewport: `${window.innerWidth}x${window.innerHeight}`,
        timestamp: new Date().toISOString()
      }
    };

    return JSON.stringify(logsData, null, 2);
  }

  // Clear logs
  clearLogs() {
    this.logs = [];
    this.info('Logs cleared');
  }
}

// Create singleton instance
const logger = new Logger();

// Global error handler
window.addEventListener('error', (event) => {
  logger.error('Uncaught JavaScript error', {
    message: event.message,
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    stack: event.error?.stack
  });
});

// Global promise rejection handler
window.addEventListener('unhandledrejection', (event) => {
  logger.error('Unhandled promise rejection', {
    message: event.reason?.message || String(event.reason),
    stack: event.reason?.stack
  });
});

export default logger;

// Export convenience functions
export const logError = (message, context) => logger.error(message, context);
export const logWarn = (message, context) => logger.warn(message, context);
export const logInfo = (message, context) => logger.info(message, context);
export const logDebug = (message, context) => logger.debug(message, context);
export const logApiCall = (method, url, context) => logger.apiCall(method, url, context);
export const logUserAction = (action, context) => logger.userAction(action, context);
export const logPerformance = (label, fn) => logger.performance(label, fn);
export const getRecentLogs = (count) => logger.getRecentLogs(count);
export const exportLogs = () => logger.exportLogs();