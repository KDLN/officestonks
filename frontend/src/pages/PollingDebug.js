import React, { useState, useEffect } from 'react';
import { 
  startPolling, 
  stopPolling, 
  addPollingListener, 
  removePollingListener,
  getPollingConnectionState,
  isPollingActive,
  getPollingStats,
  getEnhancedPollingStats,
  forcePoll,
  resetPolling
} from '../services/polling';
import Navigation from '../components/Navigation';

const PollingDebug = () => {
  const [stats, setStats] = useState(null);
  const [messages, setMessages] = useState([]);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Update stats every second
  useEffect(() => {
    const updateStats = () => {
      setStats(getEnhancedPollingStats());
    };

    updateStats();
    
    if (autoRefresh) {
      const interval = setInterval(updateStats, 1000);
      return () => clearInterval(interval);
    }
  }, [autoRefresh]);

  // Listen for all polling events
  useEffect(() => {
    const handlePollingEvent = (data) => {
      const timestamp = new Date().toLocaleTimeString();
      setMessages(prev => [
        { timestamp, event: 'stockUpdate', data },
        ...prev.slice(0, 49) // Keep last 50 messages
      ]);
    };

    const handleConnectionState = (data) => {
      const timestamp = new Date().toLocaleTimeString();
      setMessages(prev => [
        { timestamp, event: 'connectionState', data },
        ...prev.slice(0, 49)
      ]);
    };

    const handlePollEvent = (data) => {
      const timestamp = new Date().toLocaleTimeString();
      setMessages(prev => [
        { timestamp, event: 'poll', data },
        ...prev.slice(0, 49)
      ]);
    };

    const handleError = (data) => {
      const timestamp = new Date().toLocaleTimeString();
      setMessages(prev => [
        { timestamp, event: 'error', data },
        ...prev.slice(0, 49)
      ]);
    };

    // Add listeners
    addPollingListener('stockUpdate', handlePollingEvent);
    addPollingListener('connectionState', handleConnectionState);
    addPollingListener('poll', handlePollEvent);
    addPollingListener('error', handleError);

    return () => {
      removePollingListener('stockUpdate', handlePollingEvent);
      removePollingListener('connectionState', handleConnectionState);
      removePollingListener('poll', handlePollEvent);
      removePollingListener('error', handleError);
    };
  }, []);

  const handleStartPolling = async () => {
    try {
      await startPolling();
    } catch (error) {
      console.error('Failed to start polling:', error);
    }
  };

  const handleStopPolling = () => {
    stopPolling();
  };

  const handleForcePoll = async () => {
    try {
      await forcePoll();
    } catch (error) {
      console.error('Failed to force poll:', error);
    }
  };

  const handleResetPolling = () => {
    resetPolling();
  };

  const handleClearMessages = () => {
    setMessages([]);
  };

  const getStatusColor = (state) => {
    switch (state) {
      case 'connected': return '#4CAF50';
      case 'connecting': return '#FF9800';
      case 'failed': return '#F44336';
      case 'disconnected': return '#9E9E9E';
      default: return '#9E9E9E';
    }
  };

  return (
    <div className="debug-page">
      <Navigation />
      <div className="debug-container" style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto' }}>
        <h1>🔄 HTTP Polling Debug Console</h1>
        <p>Debug and monitor the HTTP polling service for real-time stock updates.</p>

        {/* Controls */}
        <div className="debug-controls" style={{ marginBottom: '20px' }}>
          <button onClick={handleStartPolling} style={{ marginRight: '10px' }}>
            ▶️ Start Polling
          </button>
          <button onClick={handleStopPolling} style={{ marginRight: '10px' }}>
            ⏹️ Stop Polling
          </button>
          <button onClick={handleForcePoll} style={{ marginRight: '10px' }}>
            🔄 Force Poll
          </button>
          <button onClick={handleResetPolling} style={{ marginRight: '10px' }}>
            🔁 Reset Polling
          </button>
          <button onClick={handleClearMessages} style={{ marginRight: '10px' }}>
            🗑️ Clear Messages
          </button>
          <label style={{ marginLeft: '20px' }}>
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
            Auto-refresh stats
          </label>
        </div>

        {/* Status Overview */}
        {stats && (
          <div className="debug-status" style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
            gap: '15px',
            marginBottom: '20px'
          }}>
            <div className="status-card" style={{ 
              backgroundColor: '#f5f5f5', 
              padding: '15px', 
              borderRadius: '5px',
              border: `3px solid ${getStatusColor(stats.state)}`
            }}>
              <h3>Connection Status</h3>
              <p style={{ color: getStatusColor(stats.state), fontWeight: 'bold' }}>
                {stats.state.toUpperCase()}
              </p>
              <small>Active: {stats.active ? 'Yes' : 'No'}</small>
            </div>

            <div className="status-card" style={{ backgroundColor: '#f5f5f5', padding: '15px', borderRadius: '5px' }}>
              <h3>Poll Statistics</h3>
              <p>Count: <strong>{stats.pollCounter}</strong></p>
              <p>Failures: <strong>{stats.consecutiveFailures}</strong></p>
            </div>

            <div className="status-card" style={{ backgroundColor: '#f5f5f5', padding: '15px', borderRadius: '5px' }}>
              <h3>Configuration</h3>
              <p>Base Interval: <strong>{stats.interval}ms</strong></p>
              <p>Current Interval: <strong>{stats.currentInterval}ms</strong></p>
              <p>Adaptive: <strong>{stats.adaptive ? 'Yes' : 'No'}</strong></p>
              <p>Listeners: <strong>{stats.listeners}</strong></p>
            </div>

            <div className="status-card" style={{ backgroundColor: '#f5f5f5', padding: '15px', borderRadius: '5px' }}>
              <h3>User Activity</h3>
              <p>Page Visible: <strong>{stats.isPageVisible ? 'Yes' : 'No'}</strong></p>
              <p>Last Activity: <strong>{stats.lastUserActivity}</strong></p>
              <p>Time Since: <strong>{Math.round(stats.timeSinceActivity / 1000)}s</strong></p>
            </div>

            <div className="status-card" style={{ backgroundColor: '#f5f5f5', padding: '15px', borderRadius: '5px' }}>
              <h3>Backend</h3>
              <p style={{ wordBreak: 'break-all', fontSize: '0.9em' }}>
                <strong>{stats.backendUrl}</strong>
              </p>
              <p>
                Last Success: {stats.lastSuccessfulPoll ? 
                  new Date(stats.lastSuccessfulPoll).toLocaleTimeString() : 'Never'
                }
              </p>
            </div>
          </div>
        )}

        {/* Live Messages */}
        <div className="debug-messages">
          <h2>📡 Live Messages ({messages.length})</h2>
          <div style={{ 
            height: '400px', 
            overflowY: 'auto', 
            border: '1px solid #ccc', 
            backgroundColor: '#000',
            color: '#00ff00',
            fontFamily: 'monospace',
            fontSize: '12px',
            padding: '10px'
          }}>
            {messages.length === 0 ? (
              <div style={{ color: '#666' }}>No messages yet...</div>
            ) : (
              messages.map((message, index) => (
                <div key={index} style={{ marginBottom: '5px' }}>
                  <span style={{ color: '#888' }}>[{message.timestamp}]</span>
                  <span style={{ 
                    color: message.event === 'error' ? '#ff4444' : 
                           message.event === 'stockUpdate' ? '#44ff44' :
                           message.event === 'connectionState' ? '#4444ff' : '#ffff44',
                    marginLeft: '5px',
                    fontWeight: 'bold'
                  }}>
                    {message.event.toUpperCase()}
                  </span>
                  <span style={{ marginLeft: '10px' }}>
                    {JSON.stringify(message.data)}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Instructions */}
        <div className="debug-instructions" style={{ marginTop: '20px', padding: '15px', backgroundColor: '#e3f2fd', borderRadius: '5px' }}>
          <h3>🔧 Testing Instructions</h3>
          <ul>
            <li><strong>Start Polling:</strong> Begin adaptive HTTP polling for stock updates</li>
            <li><strong>Stop Polling:</strong> Stop the polling service</li>
            <li><strong>Force Poll:</strong> Trigger an immediate poll request</li>
            <li><strong>Reset Polling:</strong> Reset the service and restart</li>
            <li><strong>Monitor Messages:</strong> Watch for stock updates, connection states, and errors</li>
          </ul>
          <p><strong>Rate Limiting:</strong> Polling requests are excluded from user rate limits via X-Request-Type header.</p>
          <p><strong>Adaptive Behavior:</strong></p>
          <ul>
            <li>Active users (last 30s): 5 second intervals</li>
            <li>Recent users (last 5min): 10 second intervals</li>
            <li>Idle users: 15 second intervals</li>
            <li>Hidden page: 20 second intervals</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default PollingDebug;