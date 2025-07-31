import React, { useState } from 'react';
import Navigation from '../components/Navigation';
import { runCrisisTests, getTestStatus } from '../services/admin';
import './Tests.css';

const Tests = () => {
  const [testSuite, setTestSuite] = useState(null);
  const [isRunning, setIsRunning] = useState(false);
  const [testStatus, setTestStatus] = useState(null);
  const [selectedTest, setSelectedTest] = useState('crisis');
  const [error, setError] = useState(null);

  // Load test status on component mount
  React.useEffect(() => {
    loadTestStatus();
  }, []);

  const loadTestStatus = async () => {
    try {
      const status = await getTestStatus();
      setTestStatus(status);
    } catch (err) {
      console.error('Error loading test status:', err);
      setError('Failed to load test status');
    }
  };

  const runTests = async () => {
    if (isRunning) return;
    
    setIsRunning(true);
    setError(null);
    setTestSuite(null);

    try {
      let results;
      
      switch (selectedTest) {
        case 'crisis':
          results = await runCrisisTests();
          break;
        default:
          throw new Error('Unknown test type');
      }
      
      setTestSuite(results);
    } catch (err) {
      console.error('Error running tests:', err);
      setError(`Failed to run tests: ${err.message}`);
    } finally {
      setIsRunning(false);
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'passed': return '#4caf50';
      case 'failed': return '#f44336';
      case 'running': return '#ff9800';
      default: return '#666';
    }
  };

  const formatDuration = (duration) => {
    // Convert Go duration string to human readable
    if (duration.includes('ms')) {
      return duration;
    } else if (duration.includes('s')) {
      return duration;
    }
    return duration;
  };

  return (
    <div className="tests-page">
      <Navigation />
      <div className="tests-container">
        <div className="tests-header">
          <h1>🧪 Test Suite</h1>
          <p>Automated testing for crisis mechanics and system integrity</p>
        </div>

        {error && (
          <div className="error-message">
            {error}
          </div>
        )}

        <div className="test-controls">
          <div className="test-selector">
            <label htmlFor="test-type">Test Suite:</label>
            <select 
              id="test-type"
              value={selectedTest} 
              onChange={(e) => setSelectedTest(e.target.value)}
              disabled={isRunning}
            >
              <option value="crisis">Crisis Mechanics Test Suite</option>
              {/* Add more test suites here in the future */}
            </select>
          </div>

          <button 
            onClick={runTests} 
            disabled={isRunning}
            className={`run-tests-btn ${isRunning ? 'running' : ''}`}
          >
            {isRunning ? (
              <>
                <span className="spinner"></span>
                Running Tests...
              </>
            ) : (
              <>
                ▶️ Run Tests
              </>
            )}
          </button>
        </div>

        {testStatus && (
          <div className="test-status-card">
            <h3>Test Environment Status</h3>
            <div className="status-grid">
              <div className="status-item">
                <strong>Available Tests:</strong>
                <ul>
                  {testStatus.available_tests?.map((test, index) => (
                    <li key={index}>{test}</li>
                  ))}
                </ul>
              </div>
              <div className="status-item">
                <strong>Environment:</strong>
                <div>Simulator: {testStatus.environment?.simulator_running ? '✅ Running' : '❌ Stopped'}</div>
                <div>Test Mode: {testStatus.environment?.test_mode_enabled ? '✅ Enabled' : '❌ Disabled'}</div>
              </div>
            </div>
          </div>
        )}

        {testSuite && (
          <div className="test-results">
            <div className="test-summary">
              <h2>Test Results: {testSuite.suite_name}</h2>
              <div className="summary-stats">
                <div className="stat">
                  <span className="stat-number">{testSuite.total_tests}</span>
                  <span className="stat-label">Total Tests</span>
                </div>
                <div className="stat passed">
                  <span className="stat-number">{testSuite.passed}</span>
                  <span className="stat-label">Passed</span>
                </div>
                <div className="stat failed">
                  <span className="stat-number">{testSuite.failed}</span>
                  <span className="stat-label">Failed</span>
                </div>
                <div className="stat">
                  <span className="stat-number">
                    {((new Date(testSuite.end_time) - new Date(testSuite.start_time)) / 1000).toFixed(2)}s
                  </span>
                  <span className="stat-label">Total Time</span>
                </div>
              </div>
            </div>

            <div className="test-details">
              {testSuite.tests?.map((test, index) => (
                <div key={index} className={`test-result ${test.status}`}>
                  <div className="test-header">
                    <div className="test-name">
                      <span 
                        className="status-indicator"
                        style={{ backgroundColor: getStatusColor(test.status) }}
                      ></span>
                      {test.test_name}
                    </div>
                    <div className="test-meta">
                      <span className="test-status">{test.status.toUpperCase()}</span>
                      <span className="test-duration">{formatDuration(test.duration)}</span>
                    </div>
                  </div>
                  
                  <div className="test-message">
                    {test.message}
                  </div>

                  {test.error && (
                    <div className="test-error">
                      <strong>Error:</strong> {test.error}
                    </div>
                  )}

                  {test.details && (
                    <div className="test-details-section">
                      <h4>Details:</h4>
                      <pre className="test-details-data">
                        {JSON.stringify(test.details, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {!testSuite && !isRunning && (
          <div className="empty-state">
            <h3>Ready to Run Tests</h3>
            <p>Select a test suite and click "Run Tests" to begin automated testing.</p>
            <div className="test-info">
              <h4>Crisis Mechanics Test Suite includes:</h4>
              <ul>
                <li>Force Crisis Event - Tests stock crisis at $0.01</li>
                <li>Bankruptcy with Portfolio Impact - Tests bankruptcy processing</li>
                <li>Recovery Event - Tests stock recovery mechanics</li>
                <li>News Generation Verification - Checks automated news</li>
                <li>Sector Contagion Check - Analyzes sector-wide effects</li>
              </ul>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Tests;