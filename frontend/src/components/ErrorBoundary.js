import React from 'react';
import logger from '../services/logger';
import './ErrorBoundary.css';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { 
      hasError: false, 
      error: null,
      errorInfo: null,
      errorCount: 0
    };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    // Log error with full context using our logger
    logger.error('React Error Boundary triggered', {
      error: {
        name: error.name,
        message: error.message,
        stack: error.stack
      },
      componentStack: errorInfo.componentStack,
      component: this.props.component || 'Unknown',
      errorCount: this.state.errorCount + 1,
      url: window.location.href,
      timestamp: new Date().toISOString()
    });
    
    // Update state with error details
    this.setState(prevState => ({
      error,
      errorInfo,
      errorCount: prevState.errorCount + 1
    }));

    // Log user action that led to error
    logger.userAction('error_boundary_triggered', {
      errorMessage: error.message,
      component: this.props.component || 'Unknown',
      errorCount: this.state.errorCount + 1
    });
  }

  handleReset = () => {
    logger.userAction('error_boundary_reset_clicked', {
      errorCount: this.state.errorCount,
      errorMessage: this.state.error?.message
    });

    // Clear all localStorage data
    try {
      localStorage.clear();
      logger.info('Cleared all localStorage data');
    } catch (err) {
      logger.error('Failed to clear localStorage', { error: err.message });
    }

    // Clear sessionStorage too
    try {
      sessionStorage.clear();
      logger.info('Cleared all sessionStorage data');
    } catch (err) {
      logger.error('Failed to clear sessionStorage', { error: err.message });
    }

    // Reload the page to start fresh
    window.location.href = '/login';
  };

  handleSoftReset = () => {
    logger.userAction('error_boundary_retry_clicked', {
      errorCount: this.state.errorCount,
      errorMessage: this.state.error?.message
    });

    // Try to recover without clearing data
    this.setState({ 
      hasError: false, 
      error: null, 
      errorInfo: null 
    });
  };

  render() {
    if (this.state.hasError) {
      const isDevelopment = process.env.NODE_ENV === 'development';
      
      return (
        <div className="error-boundary-container">
          <div className="error-boundary-content">
            <div className="error-icon">⚠️</div>
            <h1>Oops! Something went wrong</h1>
            <p className="error-message">
              We're sorry, but Office Stonks encountered an unexpected error.
            </p>
            
            {/* Show error details in development */}
            {isDevelopment && this.state.error && (
              <details className="error-details">
                <summary>Error Details (Development Only)</summary>
                <pre className="error-stack">
                  {this.state.error.toString()}
                  {this.state.errorInfo?.componentStack}
                </pre>
              </details>
            )}
            
            <div className="error-actions">
              <button 
                onClick={this.handleSoftReset} 
                className="btn-secondary"
                disabled={this.state.errorCount > 2}
              >
                Try Again
              </button>
              <button 
                onClick={this.handleReset} 
                className="btn-primary"
              >
                Reset Application
              </button>
            </div>
            
            <p className="error-help">
              If the problem persists, please try:
            </p>
            <ul className="error-help-list">
              <li>Refreshing the page</li>
              <li>Clearing your browser cache</li>
              <li>Using a different browser</li>
              <li>Contacting support if the issue continues</li>
            </ul>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;