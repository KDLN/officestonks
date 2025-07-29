import React from 'react';
import './ErrorMessage.css';

const ErrorMessage = ({ 
  message = 'Something went wrong. Please try again.', 
  onRetry = null,
  type = 'error' // 'error', 'warning', 'info'
}) => {
  return (
    <div className={`error-message-container ${type}`}>
      <div className="error-icon">
        {type === 'error' && '⚠️'}
        {type === 'warning' && '⚡'}
        {type === 'info' && 'ℹ️'}
      </div>
      <div className="error-content">
        <p className="error-text">{message}</p>
        {onRetry && (
          <button className="retry-button" onClick={onRetry}>
            Try Again
          </button>
        )}
      </div>
    </div>
  );
};

export default ErrorMessage;