import React, { useState, useEffect } from 'react';
import './LoadingScreen.css';

const LoadingScreen = ({ message = 'Loading Office Stonks...', timeout = 10000, onTimeout }) => {
  const [timedOut, setTimedOut] = useState(false);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    // Simulate progress
    const progressInterval = setInterval(() => {
      setProgress(prev => {
        if (prev >= 90) return prev;
        return prev + Math.random() * 15;
      });
    }, 500);

    // Set timeout
    const timeoutId = setTimeout(() => {
      setTimedOut(true);
      if (onTimeout) {
        onTimeout();
      }
    }, timeout);

    return () => {
      clearInterval(progressInterval);
      clearTimeout(timeoutId);
    };
  }, [timeout, onTimeout]);

  const handleRetry = () => {
    window.location.reload();
  };

  const handleReset = () => {
    // Clear all storage and reload
    try {
      localStorage.clear();
      sessionStorage.clear();
    } catch (err) {
      console.error('Failed to clear storage:', err);
    }
    window.location.href = '/login';
  };

  if (timedOut) {
    return (
      <div className="loading-screen loading-screen-error">
        <div className="loading-content">
          <div className="loading-icon-error">⚠️</div>
          <h2>Loading Taking Too Long</h2>
          <p>The application is taking longer than expected to load.</p>
          <div className="loading-actions">
            <button onClick={handleRetry} className="btn-retry">
              Retry
            </button>
            <button onClick={handleReset} className="btn-reset">
              Reset App
            </button>
          </div>
          <p className="loading-help">
            If this keeps happening, there might be a connection issue.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="loading-screen">
      <div className="loading-content">
        <div className="loading-logo">
          <div className="loading-icon">📈</div>
          <h1>Office Stonks</h1>
        </div>
        <div className="loading-spinner">
          <div className="spinner"></div>
        </div>
        <p className="loading-message">{message}</p>
        <div className="loading-progress">
          <div 
            className="loading-progress-bar" 
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>
    </div>
  );
};

export default LoadingScreen;