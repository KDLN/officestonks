// WebSocket Diagnostics Dashboard Component
// Real-time monitoring and testing interface for WebSocket connections

import React, { useState, useEffect, useRef } from 'react';
import websocketForensics from '../services/websocketForensics';
import websocketTester from '../services/websocketTester';
import { getConnectionDiagnostics } from '../services/websocket';
import './WebSocketDiagnostics.css';

const WebSocketDiagnostics = () => {
  const [forensicsData, setForensicsData] = useState({});
  const [testResults, setTestResults] = useState(null);
  const [isTestRunning, setIsTestRunning] = useState(false);
  const [connectionStats, setConnectionStats] = useState({});
  const [liveUpdates, setLiveUpdates] = useState([]);
  const [selectedTab, setSelectedTab] = useState('overview');
  const [testConfig, setTestConfig] = useState({
    reliabilityIterations: 50,
    stressConcurrent: 5,
    stressIterations: 3
  });
  
  const liveUpdatesRef = useRef([]);
  const maxLiveUpdates = 100;

  useEffect(() => {
    // Set up forensics listener
    const forensicsListener = (event) => {
      const update = {
        id: Date.now() + Math.random(),
        timestamp: event.timestamp,
        type: event.event,
        data: event.data
      };
      
      liveUpdatesRef.current = [update, ...liveUpdatesRef.current.slice(0, maxLiveUpdates - 1)];
      setLiveUpdates([...liveUpdatesRef.current]);
      
      // Update forensics data
      if (event.event === 'patternAnalysis') {
        setForensicsData(event.data);
      }
    };

    // Set up tester listener
    const testerListener = (event) => {
      const update = {
        id: Date.now() + Math.random(),
        timestamp: event.timestamp,
        type: `test_${event.event}`,
        data: event.data
      };
      
      liveUpdatesRef.current = [update, ...liveUpdatesRef.current.slice(0, maxLiveUpdates - 1)];
      setLiveUpdates([...liveUpdatesRef.current]);
      
      if (event.event === 'testCompleted' || event.event === 'stressTestCompleted' || 
          event.event === 'recoveryTestCompleted') {
        setTestResults(event.data);
        setIsTestRunning(false);
      }
    };

    websocketForensics.addListener(forensicsListener);
    websocketTester.addListener(testerListener);

    // Initial data load
    loadInitialData();

    // Periodic updates
    const updateInterval = setInterval(() => {
      updateConnectionStats();
    }, 2000);

    return () => {
      websocketForensics.removeListener(forensicsListener);
      websocketTester.removeListener(testerListener);
      clearInterval(updateInterval);
    };
  }, []);

  const loadInitialData = () => {
    setForensicsData(websocketForensics.analyzePatterns());
    setConnectionStats(getConnectionDiagnostics());
  };

  const updateConnectionStats = () => {
    setConnectionStats(getConnectionDiagnostics());
  };

  const runReliabilityTest = async () => {
    if (isTestRunning) return;
    
    setIsTestRunning(true);
    setTestResults(null);
    
    try {
      await websocketTester.runReliabilityTest(testConfig.reliabilityIterations);
    } catch (error) {
      console.error('Reliability test failed:', error);
      setIsTestRunning(false);
    }
  };

  const runStressTest = async () => {
    if (isTestRunning) return;
    
    setIsTestRunning(true);
    setTestResults(null);
    
    try {
      await websocketTester.runStressTest(testConfig.stressConcurrent, testConfig.stressIterations);
    } catch (error) {
      console.error('Stress test failed:', error);
      setIsTestRunning(false);
    }
  };

  const runRecoveryTest = async () => {
    if (isTestRunning) return;
    
    setIsTestRunning(true);
    setTestResults(null);
    
    try {
      await websocketTester.runRecoveryTest(5);
    } catch (error) {
      console.error('Recovery test failed:', error);
      setIsTestRunning(false);
    }
  };

  const runBrowserCompatibilityTest = async () => {
    if (isTestRunning) return;
    
    setIsTestRunning(true);
    
    try {
      const result = await websocketTester.runBrowserCompatibilityTest();
      setTestResults(result);
    } catch (error) {
      console.error('Browser compatibility test failed:', error);
    } finally {
      setIsTestRunning(false);
    }
  };

  const clearForensicsData = () => {
    websocketForensics.clearData();
    setForensicsData({});
    setLiveUpdates([]);
    liveUpdatesRef.current = [];
  };

  const exportDiagnosticData = () => {
    const data = {
      forensicsData: websocketForensics.exportDiagnosticData(),
      testResults: websocketTester.getAllTestResults(),
      connectionStats,
      timestamp: Date.now()
    };
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `websocket-diagnostics-${new Date().toISOString()}.json`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const getSuccessRateColor = (rate) => {
    if (rate >= 90) return 'success';
    if (rate >= 70) return 'warning';
    return 'error';
  };

  const getConnectionStateColor = (state) => {
    switch (state) {
      case 'connected': return 'success';
      case 'connecting': return 'warning';
      case 'disconnected': return 'error';
      case 'failed': return 'error';
      default: return 'neutral';
    }
  };

  const renderOverviewTab = () => (
    <div className="diagnostics-overview">
      <div className="stats-grid">
        <div className="stat-card">
          <h3>Current Connection</h3>
          <div className={`status-indicator ${getConnectionStateColor(connectionStats.connectionState)}`}>
            {connectionStats.connectionState || 'Unknown'}
          </div>
          <div className="stat-details">
            <div>Reconnect Attempts: {connectionStats.reconnectAttempts || 0}</div>
            <div>Polling Active: {connectionStats.isPollingActive ? 'Yes' : 'No'}</div>
          </div>
        </div>

        <div className="stat-card">
          <h3>Success Rate</h3>
          <div className={`success-rate ${getSuccessRateColor(forensicsData.successRate || 0)}`}>
            {forensicsData.successRate ? `${forensicsData.successRate.toFixed(1)}%` : 'No data'}
          </div>
          <div className="stat-details">
            <div>Total Attempts: {forensicsData.totalAttempts || 0}</div>
            <div>Avg Connection: {forensicsData.averageConnectionTime ? `${forensicsData.averageConnectionTime}ms` : 'N/A'}</div>
          </div>
        </div>

        <div className="stat-card">
          <h3>Error Patterns</h3>
          {forensicsData.commonErrorPatterns && forensicsData.commonErrorPatterns.length > 0 ? (
            <div className="error-patterns">
              {forensicsData.commonErrorPatterns.slice(0, 3).map((pattern, index) => (
                <div key={index} className="error-pattern">
                  <span className="error-category">{pattern.category}</span>
                  <span className="error-count">{pattern.count}x</span>
                </div>
              ))}
            </div>
          ) : (
            <div className="no-data">No error data</div>
          )}
        </div>

        <div className="stat-card">
          <h3>Railway Analysis</h3>
          {forensicsData.railwaySpecificIssues ? (
            <div className="railway-stats">
              <div>Railway Attempts: {forensicsData.railwaySpecificIssues.totalRailwayAttempts}</div>
              <div>Hijacker Errors: {forensicsData.railwaySpecificIssues.hijackerErrorCount}</div>
              <div className={`error-rate ${getSuccessRateColor(100 - forensicsData.railwaySpecificIssues.hijackerErrorRate)}`}>
                Error Rate: {forensicsData.railwaySpecificIssues.hijackerErrorRate.toFixed(1)}%
              </div>
            </div>
          ) : (
            <div className="no-data">No Railway data</div>
          )}
        </div>
      </div>

      <div className="recommendations">
        <h3>Recommendations</h3>
        {forensicsData.recommendations && forensicsData.recommendations.length > 0 ? (
          <div className="recommendation-list">
            {forensicsData.recommendations.map((rec, index) => (
              <div key={index} className={`recommendation ${rec.priority}`}>
                <div className="rec-header">
                  <span className="rec-priority">{rec.priority.toUpperCase()}</span>
                  <span className="rec-category">{rec.category}</span>
                </div>
                <div className="rec-message">{rec.recommendation || rec.message}</div>
                {rec.action && <div className="rec-action">Action: {rec.action}</div>}
                {rec.evidence && <div className="rec-evidence">Evidence: {rec.evidence}</div>}
              </div>
            ))}
          </div>
        ) : (
          <div className="no-recommendations">No recommendations available. Run tests to generate insights.</div>
        )}
      </div>
    </div>
  );

  const renderTestingTab = () => (
    <div className="diagnostics-testing">
      <div className="test-controls">
        <h3>Automated Testing Suite</h3>
        
        <div className="test-config">
          <div className="config-group">
            <label>Reliability Test Iterations:</label>
            <input 
              type="number" 
              value={testConfig.reliabilityIterations}
              onChange={(e) => setTestConfig({...testConfig, reliabilityIterations: parseInt(e.target.value)})}
              min="10"
              max="200"
              disabled={isTestRunning}
            />
          </div>
          
          <div className="config-group">
            <label>Stress Test - Concurrent:</label>
            <input 
              type="number" 
              value={testConfig.stressConcurrent}
              onChange={(e) => setTestConfig({...testConfig, stressConcurrent: parseInt(e.target.value)})}
              min="2"
              max="20"
              disabled={isTestRunning}
            />
          </div>
          
          <div className="config-group">
            <label>Stress Test - Iterations:</label>
            <input 
              type="number" 
              value={testConfig.stressIterations}
              onChange={(e) => setTestConfig({...testConfig, stressIterations: parseInt(e.target.value)})}
              min="1"
              max="10"
              disabled={isTestRunning}
            />
          </div>
        </div>

        <div className="test-buttons">
          <button 
            onClick={runReliabilityTest}
            disabled={isTestRunning}
            className="test-button reliability"
          >
            {isTestRunning ? '🧪 Running...' : '🎯 Reliability Test'}
          </button>
          
          <button 
            onClick={runStressTest}
            disabled={isTestRunning}
            className="test-button stress"
          >
            {isTestRunning ? '🧪 Running...' : '🚀 Stress Test'}
          </button>
          
          <button 
            onClick={runRecoveryTest}
            disabled={isTestRunning}
            className="test-button recovery"
          >
            {isTestRunning ? '🧪 Running...' : '🔄 Recovery Test'}
          </button>
          
          <button 
            onClick={runBrowserCompatibilityTest}
            disabled={isTestRunning}
            className="test-button compatibility"
          >
            {isTestRunning ? '🧪 Running...' : '🌐 Browser Test'}
          </button>
        </div>
      </div>

      {testResults && (
        <div className="test-results">
          <h3>Test Results</h3>
          <div className="results-summary">
            <div className="result-stat">
              <label>Test Type:</label>
              <span>{testResults.testType || 'Reliability'}</span>
            </div>
            
            {testResults.successRate && (
              <div className="result-stat">
                <label>Success Rate:</label>
                <span className={getSuccessRateColor(testResults.successRate)}>
                  {testResults.successRate.toFixed(1)}%
                </span>
              </div>
            )}
            
            {testResults.totalAttempts && (
              <div className="result-stat">
                <label>Total Attempts:</label>
                <span>{testResults.totalAttempts}</span>
              </div>
            )}
            
            {testResults.performance && testResults.performance.averageConnectionTime && (
              <div className="result-stat">
                <label>Avg Connection Time:</label>
                <span>{Math.round(testResults.performance.averageConnectionTime)}ms</span>
              </div>
            )}
          </div>

          {testResults.recommendations && testResults.recommendations.length > 0 && (
            <div className="test-recommendations">
              <h4>Test-Based Recommendations:</h4>
              {testResults.recommendations.map((rec, index) => (
                <div key={index} className={`test-recommendation ${rec.priority}`}>
                  <strong>{rec.category}:</strong> {rec.message}
                  {rec.action && <div className="rec-action">→ {rec.action}</div>}
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );

  const renderLiveUpdatesTab = () => (
    <div className="diagnostics-live">
      <div className="live-header">
        <h3>Live Updates</h3>
        <div className="live-controls">
          <button onClick={() => setLiveUpdates([])}>Clear Updates</button>
          <span className="update-count">Updates: {liveUpdates.length}</span>
        </div>
      </div>
      
      <div className="live-updates-container">
        {liveUpdates.length === 0 ? (
          <div className="no-updates">No live updates yet. Make WebSocket connections to see diagnostic data.</div>
        ) : (
          <div className="live-updates-list">
            {liveUpdates.map((update) => (
              <div key={update.id} className={`live-update ${update.type}`}>
                <div className="update-header">
                  <span className="update-type">{update.type}</span>
                  <span className="update-time">
                    {new Date(update.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className="update-data">
                  <pre>{JSON.stringify(update.data, null, 2)}</pre>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );

  const renderRawDataTab = () => (
    <div className="diagnostics-raw">
      <div className="raw-header">
        <h3>Raw Diagnostic Data</h3>
        <div className="raw-controls">
          <button onClick={exportDiagnosticData}>Export All Data</button>
          <button onClick={clearForensicsData} className="danger">Clear Data</button>
        </div>
      </div>
      
      <div className="raw-data-sections">
        <div className="raw-section">
          <h4>Forensics Analysis</h4>
          <pre className="raw-data">{JSON.stringify(forensicsData, null, 2)}</pre>
        </div>
        
        <div className="raw-section">
          <h4>Connection Stats</h4>
          <pre className="raw-data">{JSON.stringify(connectionStats, null, 2)}</pre>
        </div>
        
        <div className="raw-section">
          <h4>Test Results</h4>
          <pre className="raw-data">{JSON.stringify(testResults, null, 2)}</pre>
        </div>
      </div>
    </div>
  );

  return (
    <div className="websocket-diagnostics">
      <div className="diagnostics-header">
        <h2>🔬 WebSocket Diagnostics Suite</h2>
        <div className="header-stats">
          <div className="header-stat">
            <label>Current State:</label>
            <span className={`state-badge ${getConnectionStateColor(connectionStats.connectionState)}`}>
              {connectionStats.connectionState || 'Unknown'}
            </span>
          </div>
          <div className="header-stat">
            <label>Success Rate:</label>
            <span className={`rate-badge ${getSuccessRateColor(forensicsData.successRate || 0)}`}>
              {forensicsData.successRate ? `${forensicsData.successRate.toFixed(1)}%` : 'No data'}
            </span>
          </div>
        </div>
      </div>

      <div className="diagnostics-tabs">
        <button 
          className={`tab ${selectedTab === 'overview' ? 'active' : ''}`}
          onClick={() => setSelectedTab('overview')}
        >
          📊 Overview
        </button>
        <button 
          className={`tab ${selectedTab === 'testing' ? 'active' : ''}`}
          onClick={() => setSelectedTab('testing')}
        >
          🧪 Testing
        </button>
        <button 
          className={`tab ${selectedTab === 'live' ? 'active' : ''}`}
          onClick={() => setSelectedTab('live')}
        >
          📡 Live Updates
        </button>
        <button 
          className={`tab ${selectedTab === 'raw' ? 'active' : ''}`}
          onClick={() => setSelectedTab('raw')}
        >
          📋 Raw Data
        </button>
      </div>

      <div className="diagnostics-content">
        {selectedTab === 'overview' && renderOverviewTab()}
        {selectedTab === 'testing' && renderTestingTab()}
        {selectedTab === 'live' && renderLiveUpdatesTab()}
        {selectedTab === 'raw' && renderRawDataTab()}
      </div>

      {isTestRunning && (
        <div className="test-overlay">
          <div className="test-spinner">
            <div className="spinner"></div>
            <div>Running comprehensive WebSocket tests...</div>
            <div className="test-notice">This may take several minutes</div>
          </div>
        </div>
      )}
    </div>
  );
};

export default WebSocketDiagnostics;