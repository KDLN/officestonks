import React from 'react';
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
    // Log error details for debugging
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    // Update state with error details
    this.setState(prevState => ({
      error,
      errorInfo,
      errorCount: prevState.errorCount + 1
    }));

    // Log to external error reporting service if needed
    // Example: logErrorToService(error, errorInfo);
  }

  handleReset = () => {
    // Clear all localStorage data
    try {
      localStorage.clear();
      console.log('Cleared all localStorage data');
    } catch (err) {
      console.error('Failed to clear localStorage:', err);
    }

    // Clear sessionStorage too
    try {
      sessionStorage.clear();
      console.log('Cleared all sessionStorage data');
    } catch (err) {
      console.error('Failed to clear sessionStorage:', err);
    }

    // Reload the page to start fresh
    window.location.href = '/login';
  };

  handleSoftReset = () => {
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