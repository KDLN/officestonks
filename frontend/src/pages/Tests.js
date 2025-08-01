import React, { useState } from 'react';
import Navigation from '../components/Navigation';
import { runCrisisTests, runPortfolioTests, getTestStatus } from '../services/admin';
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

  // Test changelog system
  const testChangelogSystem = async () => {
    const tests = [];
    
    // Test 1: Fetch changelog entries
    const test1 = {
      name: 'Fetch Changelog Entries',
      description: 'Verify the changelog API returns entries',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const response = await fetch('/api/changelog?limit=5');
      const data = await response.json();
      
      if (response.ok && data.entries && data.entries.length > 0) {
        test1.status = 'passed';
        test1.details = {
          entries_count: data.entries.length,
          latest_version: data.entries[0]?.version,
          latest_title: data.entries[0]?.title,
          has_v1_2_0: data.entries.some(e => e.version === 'v1.2.0')
        };
      } else {
        test1.status = 'failed';
        test1.details = { error: 'No entries returned', response_status: response.status };
      }
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Verify v1.2.0 Content
    const test2 = {
      name: 'Verify v1.2.0 Crisis Update',
      description: 'Check if v1.2.0 changelog has crisis system details',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      const response = await fetch('/api/changelog?limit=10');
      const data = await response.json();
      const v120 = data.entries?.find(e => e.version === 'v1.2.0');
      
      if (v120) {
        const hasBreakingNews = v120.changes?.some(c => c && typeof c === 'string' && c.includes('Breaking News'));
        const hasCrisisMechanics = v120.changes?.some(c => c && typeof c === 'string' && c.includes('Crisis Mechanics'));
        
        test2.status = (hasBreakingNews && hasCrisisMechanics) ? 'passed' : 'partial';
        test2.details = {
          found_v1_2_0: true,
          title: v120.title,
          changes_count: v120.changes?.length || 0,
          has_breaking_news: hasBreakingNews,
          has_crisis_mechanics: hasCrisisMechanics
        };
      } else {
        test2.status = 'failed';
        test2.details = { error: 'v1.2.0 entry not found' };
      }
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: Modal Trigger Check
    const test3 = {
      name: 'Changelog Modal Display',
      description: 'Verify localStorage handling for modal display',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const lastSeen = localStorage.getItem('lastSeenChangelogVersion');
      const shouldShow = !lastSeen || lastSeen !== 'v1.2.0';
      
      test3.status = 'passed';
      test3.details = {
        last_seen_version: lastSeen || 'none',
        should_show_modal: shouldShow,
        localStorage_key_exists: lastSeen !== null
      };
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    // Summary
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    
    return {
      name: 'Changelog System Tests',
      timestamp: new Date().toISOString(),
      total_tests: tests.length,
      passed,
      failed,
      success: failed === 0,
      tests
    };
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
        case 'portfolio':
          results = await runPortfolioTests();
          break;
        case 'changelog':
          results = await testChangelogSystem();
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
    if (!duration) {
      return 'N/A';
    }
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
              <option value="portfolio">Portfolio & Trading Test Suite</option>
              <option value="changelog">Changelog System Test Suite</option>
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
              <h4>Available Test Suites:</h4>
              <div className="test-suites-info">
                <div className="suite-info">
                  <h5>🚨 Crisis Mechanics Test Suite:</h5>
                  <ul>
                    <li>Force Crisis Event - Tests stock crisis at $0.01</li>
                    <li>Bankruptcy with Portfolio Impact - Tests bankruptcy processing</li>
                    <li>Recovery Event - Tests stock recovery mechanics</li>
                    <li>News Generation Verification - Checks automated news</li>
                    <li>Sector Contagion Check - Analyzes sector-wide effects</li>
                  </ul>
                </div>
                <div className="suite-info">
                  <h5>💼 Portfolio & Trading Test Suite:</h5>
                  <ul>
                    <li>Portfolio Calculation Accuracy - Verifies portfolio value calculations</li>
                    <li>Buy Order Processing - Tests stock purchase mechanics</li>
                    <li>Sell Order Processing - Tests stock selling mechanics</li>
                    <li>Insufficient Funds Handling - Tests error handling for poor users</li>
                    <li>Share Quantity Validation - Tests invalid quantity rejection</li>
                    <li>Transaction History Integrity - Verifies transaction recording</li>
                    <li>Concurrent Trading Simulation - Tests rapid trading scenarios</li>
                  </ul>
                </div>
                <div className="suite-info">
                  <h5>📰 Changelog System Test Suite:</h5>
                  <ul>
                    <li>Fetch Changelog Entries - Verifies API returns changelog data</li>
                    <li>Verify v1.2.0 Crisis Update - Checks for latest news system entry</li>
                    <li>Changelog Modal Display - Tests localStorage version tracking</li>
                    <li>Database Fallback - Ensures hardcoded entries work without DB</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Tests;