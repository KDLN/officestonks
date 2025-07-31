import React, { useState, useEffect } from 'react';
import logger, { getRecentLogs, exportLogs } from '../services/logger';
import './LogsPanel.css';

const LogsPanel = () => {
  const [logs, setLogs] = useState([]);
  const [selectedLevel, setSelectedLevel] = useState('ALL');
  const [searchTerm, setSearchTerm] = useState('');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [expanded, setExpanded] = useState(false);

  const LOG_LEVELS = ['ALL', 'ERROR', 'WARN', 'INFO', 'DEBUG', 'TRACE'];

  useEffect(() => {
    const updateLogs = () => {
      const recentLogs = getRecentLogs(100);
      setLogs(recentLogs);
    };

    updateLogs();

    if (autoRefresh) {
      const interval = setInterval(updateLogs, 2000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  const filteredLogs = logs.filter(log => {
    const levelMatch = selectedLevel === 'ALL' || log.level === selectedLevel;
    const searchMatch = searchTerm === '' || 
      log.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
      JSON.stringify(log.context).toLowerCase().includes(searchTerm.toLowerCase());
    
    return levelMatch && searchMatch;
  });

  const getLogLevelColor = (level) => {
    const colors = {
      ERROR: '#f44336',
      WARN: '#ff9800',
      INFO: '#2196f3',
      DEBUG: '#9c27b0',
      TRACE: '#607d8b'
    };
    return colors[level] || '#333';
  };

  const getLogLevelEmoji = (level) => {
    const emojis = {
      ERROR: '❌',
      WARN: '⚠️',
      INFO: 'ℹ️',
      DEBUG: '🔍',
      TRACE: '📍'
    };
    return emojis[level] || '📝';
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString();
  };

  const handleExportLogs = () => {
    const logsData = exportLogs();
    const blob = new Blob([logsData], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `office-stonks-logs-${new Date().toISOString().slice(0, 10)}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    logger.userAction('logs_exported', { logCount: logs.length });
  };

  const handleClearLogs = () => {
    if (window.confirm('Are you sure you want to clear all logs?')) {
      logger.clearLogs();
      setLogs([]);
      logger.userAction('logs_cleared');
    }
  };

  const copyLogToClipboard = (log) => {
    const logText = JSON.stringify(log, null, 2);
    navigator.clipboard.writeText(logText);
    logger.userAction('log_copied', { logLevel: log.level });
  };

  return (
    <div className="logs-panel">
      <div 
        className="logs-panel-header"
        onClick={() => setExpanded(!expanded)}
        style={{ cursor: 'pointer' }}
      >
        <span className="logs-panel-icon">📊</span>
        <span className="logs-panel-title">System Logs</span>
        <span className="logs-panel-count">({filteredLogs.length})</span>
        <span className={`logs-panel-arrow ${expanded ? 'expanded' : ''}`}>
          ▼
        </span>
      </div>

      {expanded && (
        <div className="logs-panel-content">
          <div className="logs-panel-controls">
            <div className="logs-control-group">
              <label htmlFor="log-level-filter">Level:</label>
              <select 
                id="log-level-filter"
                value={selectedLevel} 
                onChange={(e) => setSelectedLevel(e.target.value)}
                className="logs-select"
              >
                {LOG_LEVELS.map(level => (
                  <option key={level} value={level}>{level}</option>
                ))}
              </select>
            </div>

            <div className="logs-control-group">
              <label htmlFor="log-search">Search:</label>
              <input
                id="log-search"
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search logs..."
                className="logs-input"
              />
            </div>

            <div className="logs-control-group">
              <label>
                <input
                  type="checkbox"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                />
                Auto refresh
              </label>
            </div>
          </div>

          <div className="logs-panel-actions">
            <button onClick={handleExportLogs} className="btn-secondary">
              Export Logs
            </button>
            <button onClick={handleClearLogs} className="btn-secondary">
              Clear Logs
            </button>
            <button 
              onClick={() => {
                const recentLogs = getRecentLogs(100);
                setLogs(recentLogs);
              }} 
              className="btn-secondary"
            >
              Refresh
            </button>
          </div>

          <div className="logs-panel-list">
            {filteredLogs.length === 0 ? (
              <div className="logs-empty">
                No logs found matching the current filters.
              </div>
            ) : (
              filteredLogs.slice().reverse().map((log, index) => (
                <div 
                  key={`${log.timestamp}-${index}`} 
                  className="log-entry"
                  onClick={() => copyLogToClipboard(log)}
                  title="Click to copy log details"
                >
                  <div className="log-entry-header">
                    <span 
                      className="log-level-badge"
                      style={{ backgroundColor: getLogLevelColor(log.level) }}
                    >
                      {getLogLevelEmoji(log.level)} {log.level}
                    </span>
                    <span className="log-timestamp">
                      {formatTimestamp(log.timestamp)}
                    </span>
                    <span className="log-session-time">
                      +{Math.round(log.sessionTime / 1000)}s
                    </span>
                  </div>
                  
                  <div className="log-message">
                    {log.message}
                  </div>

                  {Object.keys(log.context).length > 0 && (
                    <details className="log-context">
                      <summary>Context ({Object.keys(log.context).length} items)</summary>
                      <pre className="log-context-content">
                        {JSON.stringify(log.context, null, 2)}
                      </pre>
                    </details>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default LogsPanel;