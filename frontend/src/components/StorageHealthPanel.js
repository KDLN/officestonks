import React, { useState, useEffect } from 'react';
import { getStorageHealth } from '../services/storageManager';
import './StorageHealthPanel.css';

const StorageHealthPanel = () => {
  const [health, setHealth] = useState(null);
  const [expanded, setExpanded] = useState(false);

  useEffect(() => {
    const updateHealth = () => {
      const healthData = getStorageHealth();
      setHealth(healthData);
    };

    updateHealth();
    
    // Update health status every 30 seconds
    const interval = setInterval(updateHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  if (!health) {
    return <div className="storage-health-loading">Loading storage health...</div>;
  }

  const getHealthIcon = () => {
    if (health.isHealthy) {
      return '✅';
    } else if (health.errors.length > 0) {
      return '❌';
    } else {
      return '⚠️';
    }
  };

  const getHealthColor = () => {
    if (health.isHealthy) {
      return 'var(--success-color, #4CAF50)';
    } else if (health.errors.length > 0) {
      return 'var(--error-color, #f44336)';
    } else {
      return 'var(--warning-color, #ff9800)';
    }
  };

  return (
    <div className="storage-health-panel">
      <div 
        className="storage-health-header"
        onClick={() => setExpanded(!expanded)}
        style={{ cursor: 'pointer' }}
      >
        <span className="storage-health-icon">{getHealthIcon()}</span>
        <span className="storage-health-title">Storage Health</span>
        <span className="storage-health-version">v{health.version}</span>
        <span className={`storage-health-arrow ${expanded ? 'expanded' : ''}`}>
          ▼
        </span>
      </div>

      {expanded && (
        <div className="storage-health-details">
          <div className="storage-health-summary">
            <div className="health-item">
              <strong>Status:</strong> 
              <span style={{ color: getHealthColor(), marginLeft: '8px' }}>
                {health.isHealthy ? 'Healthy' : 'Issues Detected'}
              </span>
            </div>
            <div className="health-item">
              <strong>Version:</strong> {health.version}
            </div>
            <div className="health-item">
              <strong>Backups:</strong> {health.backupCount}
            </div>
            <div className="health-item">
              <strong>Last Check:</strong> {new Date(health.lastValidation).toLocaleTimeString()}
            </div>
          </div>

          {health.errors.length > 0 && (
            <div className="storage-health-section">
              <h4 className="section-title error">Errors ({health.errors.length})</h4>
              <ul className="issue-list">
                {health.errors.map((error, index) => (
                  <li key={index} className="issue-item error">
                    {error}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {health.warnings.length > 0 && (
            <div className="storage-health-section">
              <h4 className="section-title warning">Warnings ({health.warnings.length})</h4>
              <ul className="issue-list">
                {health.warnings.map((warning, index) => (
                  <li key={index} className="issue-item warning">
                    {warning}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {health.migrationLog.length > 0 && (
            <div className="storage-health-section">
              <h4 className="section-title">Migration Log</h4>
              <div className="migration-log">
                {health.migrationLog.slice(-5).map((log, index) => (
                  <div key={index} className="migration-entry">
                    <span className="migration-type">{log.type}</span>
                    <span className="migration-time">
                      {new Date(log.timestamp).toLocaleString()}
                    </span>
                    <span className="migration-message">{log.message}</span>
                  </div>
                ))}
                {health.migrationLog.length > 5 && (
                  <div className="migration-more">
                    ... and {health.migrationLog.length - 5} more entries
                  </div>
                )}
              </div>
            </div>
          )}

          <div className="storage-health-actions">
            <button 
              className="btn-secondary"
              onClick={() => {
                const healthData = getStorageHealth();
                setHealth(healthData);
              }}
            >
              Refresh Status
            </button>
            <button 
              className="btn-secondary"
              onClick={() => {
                const healthData = JSON.stringify(health, null, 2);
                navigator.clipboard.writeText(healthData);
                alert('Storage health data copied to clipboard');
              }}
            >
              Copy Details
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default StorageHealthPanel;