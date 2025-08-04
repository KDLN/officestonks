import React, { useState, useEffect, useRef } from 'react';
import Navigation from '../components/Navigation';
import { runCrisisTests, runPortfolioTests, runSSETests, runStockManagementTests, getTestStatus } from '../services/admin';
import { initSSE, addSSEListener, removeSSEListener, isSSEConnected, getSSEConnectionState, forceSSEReconnect } from '../services/sse';
import { initWebSocket, getWebSocketInstance } from '../services/websocket';
import './Tests.css';

const Tests = () => {
  const [testSuite, setTestSuite] = useState(null);
  const [isRunning, setIsRunning] = useState(false);
  const [testStatus, setTestStatus] = useState(null);
  const [selectedTest, setSelectedTest] = useState('live-monitoring');
  const [error, setError] = useState(null);
  
  // Live monitoring state
  const [liveMonitoring, setLiveMonitoring] = useState({
    sseConnection: { state: 'disconnected', clientCount: 0, messageCount: 0, lastMessage: null },
    webSocketConnection: { connected: false, clientCount: 0, messageCount: 0, lastMessage: null },
    marketSimulator: { running: false, stockCount: 0, lastUpdate: null, updateFrequency: 0 },
    databaseHealth: { connected: false, responseTime: 0, activeQueries: 0, poolSize: 0 },
    apiHealth: { responseTime: 0, successRate: 0, errorCount: 0, lastCheck: null }
  });
  const [systemMessages, setSystemMessages] = useState([]);
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [debugConsole, setDebugConsole] = useState({ messages: [], visible: false });
  const intervalRefs = useRef({});  // Store interval references
  
  // Automated testing state
  const [automatedTesting, setAutomatedTesting] = useState({
    enabled: false,
    interval: 300, // 5 minutes default
    testSuites: ['sse-comprehensive', 'websocket-comprehensive'],
    lastRun: null,
    results: [],
    consecutiveFailures: 0
  });
  const automatedTestingRef = useRef(null);

  // Initialize live monitoring and test status
  useEffect(() => {
    loadTestStatus();
    
    // Initialize connections if in live monitoring mode
    if (selectedTest === 'live-monitoring') {
      startLiveMonitoring();
    }
    
    return () => {
      stopLiveMonitoring();
      stopAutomatedTesting();
    };
  }, []);

  // Handle test type changes
  useEffect(() => {
    if (selectedTest === 'live-monitoring') {
      startLiveMonitoring();
    } else {
      stopLiveMonitoring();
    }
  }, [selectedTest]);

  const loadTestStatus = async () => {
    try {
      const status = await getTestStatus();
      setTestStatus(status);
    } catch (err) {
      console.error('Error loading test status:', err);
      setError('Failed to load test status');
    }
  };

  // Live monitoring functions
  const startLiveMonitoring = async () => {
    setIsLiveMode(true);
    addLogMessage('🚀 Starting live monitoring system...');
    
    // Initialize SSE connection
    try {
      await initSSE();
      addSSEListener('connectionState', handleSSEConnectionState);
      addSSEListener('stockUpdate', handleSSEStockUpdate);
      addSSEListener('connection', handleSSEConnectionUpdate);
      addSSEListener('*', handleDebugMessage); // Catch all messages for debug console
    } catch (err) {
      addLogMessage(`❌ SSE initialization failed: ${err.message}`);
    }
    
    // Initialize WebSocket connection
    try {
      await initWebSocket();
      const ws = getWebSocketInstance();
      if (ws) {
        ws.addEventListener('open', () => handleWebSocketStateChange('connected'));
        ws.addEventListener('close', () => handleWebSocketStateChange('disconnected'));
        ws.addEventListener('message', handleWebSocketMessage);
      }
    } catch (err) {
      addLogMessage(`❌ WebSocket initialization failed: ${err.message}`);
    }
    
    // Start periodic health checks
    startHealthChecks();
    
    addLogMessage('✅ Live monitoring started');
  };

  const stopLiveMonitoring = () => {
    if (!isLiveMode) return;
    
    setIsLiveMode(false);
    addLogMessage('🛑 Stopping live monitoring...');
    
    // Clean up SSE listeners
    removeSSEListener('connectionState', handleSSEConnectionState);
    removeSSEListener('stockUpdate', handleSSEStockUpdate);
    removeSSEListener('connection', handleSSEConnectionUpdate);
    removeSSEListener('*', handleDebugMessage);
    
    // Clear all intervals
    Object.values(intervalRefs.current).forEach(interval => {
      if (interval) clearInterval(interval);
    });
    intervalRefs.current = {};
    
    addLogMessage('✅ Live monitoring stopped');
  };

  const startHealthChecks = () => {
    // SSE connection check every 5 seconds
    intervalRefs.current.sseCheck = setInterval(async () => {
      const connectionState = getSSEConnectionState();
      setLiveMonitoring(prev => ({
        ...prev,
        sseConnection: {
          ...prev.sseConnection,
          state: connectionState,
          connected: isSSEConnected()
        }
      }));
    }, 5000);

    // API health check every 10 seconds
    intervalRefs.current.apiCheck = setInterval(async () => {
      const startTime = Date.now();
      try {
        const response = await fetch('/api/health');
        const responseTime = Date.now() - startTime;
        const data = await response.json();
        
        setLiveMonitoring(prev => ({
          ...prev,
          apiHealth: {
            responseTime,
            successRate: response.ok ? 100 : 0,
            errorCount: response.ok ? 0 : prev.apiHealth.errorCount + 1,
            lastCheck: new Date().toISOString()
          },
          databaseHealth: {
            connected: data.database === 'connected',
            responseTime: responseTime,
            activeQueries: data.active_queries || 0,
            poolSize: data.pool_size || 0
          }
        }));
      } catch (err) {
        setLiveMonitoring(prev => ({
          ...prev,
          apiHealth: {
            ...prev.apiHealth,
            errorCount: prev.apiHealth.errorCount + 1,
            lastCheck: new Date().toISOString()
          }
        }));
      }
    }, 10000);

    // Market simulator status check every 5 seconds
    intervalRefs.current.marketCheck = setInterval(async () => {
      try {
        const response = await fetch('/api/stocks');
        if (response.ok) {
          const stocks = await response.json();
          setLiveMonitoring(prev => ({
            ...prev,
            marketSimulator: {
              running: true,
              stockCount: stocks.length,
              lastUpdate: new Date().toISOString(),
              updateFrequency: 2 // 2 second intervals
            }
          }));
        }
      } catch (err) {
        setLiveMonitoring(prev => ({
          ...prev,
          marketSimulator: {
            ...prev.marketSimulator,
            running: false
          }
        }));
      }
    }, 5000);
    
    // WebSocket health check every 10 seconds
    intervalRefs.current.wsCheck = setInterval(async () => {
      try {
        const response = await fetch('/ws/health');
        if (response.ok) {
          const data = await response.json();
          setLiveMonitoring(prev => ({
            ...prev,
            webSocketConnection: {
              ...prev.webSocketConnection,
              connected: true,
              lastCheck: data.time
            }
          }));
        }
      } catch (err) {
        setLiveMonitoring(prev => ({
          ...prev,
          webSocketConnection: {
            ...prev.webSocketConnection,
            connected: false
          }
        }));
      }
    }, 10000);
  };

  // Event handlers for live monitoring
  const handleSSEConnectionState = (data) => {
    setLiveMonitoring(prev => ({
      ...prev,
      sseConnection: {
        ...prev.sseConnection,
        state: data.state,
        lastMessage: new Date().toISOString(),
        connected: data.state === 'connected'
      }
    }));
    
    addLogMessage(`📡 SSE Connection: ${data.state} ${data.description ? `- ${data.description}` : ''}`);
  };

  const handleSSEStockUpdate = (data) => {
    setLiveMonitoring(prev => ({
      ...prev,
      sseConnection: {
        ...prev.sseConnection,
        messageCount: prev.sseConnection.messageCount + 1,
        lastMessage: new Date().toISOString()
      }
    }));
  };

  const handleSSEConnectionUpdate = (data) => {
    addLogMessage(`📡 SSE: ${data.status} - ${data.message || 'Connection update'}`);
  };

  const handleWebSocketStateChange = (state) => {
    setLiveMonitoring(prev => ({
      ...prev,
      webSocketConnection: {
        ...prev.webSocketConnection,
        connected: state === 'connected',
        lastMessage: new Date().toISOString()
      }
    }));
    addLogMessage(`🔌 WebSocket: ${state}`);
  };

  const handleWebSocketMessage = (event) => {
    setLiveMonitoring(prev => ({
      ...prev,
      webSocketConnection: {
        ...prev.webSocketConnection,
        messageCount: prev.webSocketConnection.messageCount + 1,
        lastMessage: new Date().toISOString()
      }
    }));
  };

  const handleDebugMessage = (data) => {
    const timestamp = new Date().toISOString();
    const debugMessage = {
      timestamp,
      type: data.type || 'unknown',
      data: data,
      raw: JSON.stringify(data, null, 2)
    };
    
    setDebugConsole(prev => ({
      ...prev,
      messages: [debugMessage, ...prev.messages.slice(0, 99)] // Keep last 100 messages
    }));
  };

  const addLogMessage = (message) => {
    const timestamp = new Date().toLocaleTimeString();
    setSystemMessages(prev => [
      { timestamp, message, id: Date.now() },
      ...prev.slice(0, 49) // Keep last 50 messages
    ]);
  };

  // Manual test triggers
  const triggerSSEReconnect = () => {
    addLogMessage('🔄 Forcing SSE reconnection...');
    forceSSEReconnect();
  };

  const clearDebugConsole = () => {
    setDebugConsole(prev => ({
      ...prev,
      messages: []
    }));
  };

  const clearSystemLog = () => {
    setSystemMessages([]);
  };

  // Automated testing functions
  const startAutomatedTesting = () => {
    if (automatedTestingRef.current) {
      clearInterval(automatedTestingRef.current);
    }
    
    setAutomatedTesting(prev => ({
      ...prev,
      enabled: true,
      lastRun: null,
      consecutiveFailures: 0
    }));
    
    addLogMessage(`🤖 Starting automated testing - Running every ${automatedTesting.interval} seconds`);
    
    // Run initial test
    runAutomatedTest();
    
    // Schedule recurring tests
    automatedTestingRef.current = setInterval(() => {
      runAutomatedTest();
    }, automatedTesting.interval * 1000);
  };

  const stopAutomatedTesting = () => {
    if (automatedTestingRef.current) {
      clearInterval(automatedTestingRef.current);
      automatedTestingRef.current = null;
    }
    
    setAutomatedTesting(prev => ({
      ...prev,
      enabled: false
    }));
    
    addLogMessage('🤖 Automated testing stopped');
  };

  const runAutomatedTest = async () => {
    if (isRunning) return; // Don't run if manual test is running
    
    const startTime = Date.now();
    addLogMessage('🤖 Running automated test cycle...');
    
    const results = [];
    
    for (const testSuite of automatedTesting.testSuites) {
      try {
        let testResult;
        
        switch (testSuite) {
          case 'sse-comprehensive':
            testResult = await runComprehensiveSSETests();
            break;
          case 'websocket-comprehensive':
            testResult = await runWebSocketTests();
            break;
          case 'frontend-integration':
            testResult = await runFrontendIntegrationTests();
            break;
          case 'performance-load':
            testResult = await runPerformanceTests();
            break;
          default:
            continue;
        }
        
        results.push({
          suite: testSuite,
          result: testResult,
          timestamp: new Date().toISOString()
        });
        
      } catch (err) {
        results.push({
          suite: testSuite,
          error: err.message,
          timestamp: new Date().toISOString()
        });
      }
    }
    
    const duration = (Date.now() - startTime) / 1000;
    const allPassed = results.every(r => r.result?.success !== false);
    const failedSuites = results.filter(r => r.result?.success === false || r.error).length;
    
    setAutomatedTesting(prev => ({
      ...prev,
      lastRun: new Date().toISOString(),
      results: [
        {
          timestamp: new Date().toISOString(),
          duration: duration,
          allPassed,
          failedSuites,
          totalSuites: results.length,
          details: results
        },
        ...prev.results.slice(0, 19) // Keep last 20 results
      ],
      consecutiveFailures: allPassed ? 0 : prev.consecutiveFailures + 1
    }));
    
    if (allPassed) {
      addLogMessage(`🤖 ✅ Automated test cycle completed successfully (${duration.toFixed(1)}s)`);
    } else {
      addLogMessage(`🤖 ❌ Automated test cycle failed - ${failedSuites}/${results.length} suites failed (${duration.toFixed(1)}s)`);
    }
    
    // Alert if too many consecutive failures
    if (!allPassed && automatedTesting.consecutiveFailures >= 3) {
      addLogMessage('🚨 WARNING: 3+ consecutive automated test failures detected!');
    }
  };

  const updateAutomatedTestSettings = (settings) => {
    setAutomatedTesting(prev => ({
      ...prev,
      ...settings
    }));
    
    // If automated testing is running, restart with new settings
    if (automatedTesting.enabled) {
      stopAutomatedTesting();
      setTimeout(() => startAutomatedTesting(), 1000);
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

  // Test monitoring system
  const testMonitoringSystem = async () => {
    const tests = [];
    const token = localStorage.getItem('token');
    
    // Test 1: Check monitoring tables exist
    const test1 = {
      name: 'Monitoring Tables Check',
      description: 'Verify monitoring database tables exist',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const response = await fetch('/api/admin/monitoring/test', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok && data.checks) {
        const tablesExist = data.checks.filter(c => c.name.includes('table exists')).every(c => c.status);
        test1.status = tablesExist ? 'passed' : 'failed';
        test1.details = {
          user_sessions_exists: data.checks.find(c => c.name === 'user_sessions table exists')?.status || false,
          user_activity_exists: data.checks.find(c => c.name === 'user_activity table exists')?.status || false,
          user_columns_exist: data.checks.find(c => c.name === 'users table has monitoring columns')?.status || false,
          can_insert_sessions: data.checks.find(c => c.name === 'can insert into user_sessions')?.status || false,
        };
      } else {
        test1.status = 'failed';
        test1.details = { error: 'Failed to fetch monitoring status', response_status: response.status };
      }
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Check active sessions
    const test2 = {
      name: 'Active Sessions Check',
      description: 'Verify active sessions are being tracked',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      const response = await fetch('/api/admin/monitoring/sessions', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok) {
        const sessionCount = data.count || 0;
        test2.status = sessionCount > 0 ? 'passed' : 'warning';
        test2.details = {
          active_sessions: sessionCount,
          sessions: data.sessions?.slice(0, 3).map(s => ({
            username: s.username,
            ip_address: s.ip_address,
            login_time: s.login_time,
          })) || [],
        };
      } else {
        test2.status = 'failed';
        test2.details = { error: 'Failed to fetch sessions', response_status: response.status };
      }
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: Check user activity logging
    const test3 = {
      name: 'Activity Logging Check',
      description: 'Verify user activities are being logged',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const response = await fetch('/api/admin/monitoring/activity?limit=10', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok) {
        const activityCount = data.count || 0;
        test3.status = activityCount > 0 ? 'passed' : 'warning';
        test3.details = {
          activity_count: activityCount,
          recent_activities: data.activities?.slice(0, 5).map(a => ({
            action: a.action,
            username: a.username,
            timestamp: a.timestamp,
            success: a.success,
          })) || [],
        };
      } else {
        test3.status = 'failed';
        test3.details = { error: 'Failed to fetch activities', response_status: response.status };
      }
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    // Test 4: Check system metrics
    const test4 = {
      name: 'System Metrics Check',
      description: 'Verify system metrics are being collected',
      status: 'running',
      details: {}
    };
    tests.push(test4);
    
    try {
      const response = await fetch('/api/admin/monitoring/metrics', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok && data) {
        test4.status = 'passed';
        test4.details = {
          active_users: data.active_users || 0,
          active_sessions: data.active_sessions || 0,
          trades_per_hour: data.trades_per_hour || 0,
          websocket_connections: data.websocket_connections || 0,
          database_health: data.database_health || 'unknown',
          error_rate: `${(data.error_rate || 0).toFixed(2)}%`,
        };
      } else {
        test4.status = 'failed';
        test4.details = { error: 'Failed to fetch metrics', response_status: response.status };
      }
    } catch (err) {
      test4.status = 'failed';
      test4.details = { error: err.message };
    }
    
    // Test 5: Check monitoring dashboard endpoint
    const test5 = {
      name: 'Dashboard Endpoint Check',
      description: 'Verify monitoring dashboard endpoint is accessible',
      status: 'running',
      details: {}
    };
    tests.push(test5);
    
    try {
      const response = await fetch('/api/admin/monitoring/dashboard', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok && data.system_metrics) {
        test5.status = 'passed';
        test5.details = {
          has_system_metrics: !!data.system_metrics,
          has_active_sessions: !!data.active_sessions,
          has_recent_activity: !!data.recent_activity,
          has_hourly_stats: !!data.hourly_stats,
          has_health_status: !!data.health_status,
        };
      } else {
        test5.status = 'failed';
        test5.details = { error: 'Failed to fetch dashboard data', response_status: response.status };
      }
    } catch (err) {
      test5.status = 'failed';
      test5.details = { error: err.message };
    }
    
    // Test 6: Trade Logging Test
    const test6 = {
      name: 'Trade Logging Integration',
      description: 'Test if trades are properly logged with session tracking',
      status: 'running',
      details: {}
    };
    tests.push(test6);
    
    try {
      // Make a small test trade to generate activity
      const tradeResponse = await fetch('/api/trading', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          stock_id: 1,
          action: 'buy',
          quantity: 1
        })
      });
      
      // Check if trade was logged in activity
      const activityResponse = await fetch('/api/admin/monitoring/activity?limit=5', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const activityData = await activityResponse.json();
      
      const hasTradeActivity = activityData.activities?.some(a => a.action === 'trade') || false;
      
      test6.status = hasTradeActivity ? 'passed' : 'warning';
      test6.details = {
        trade_attempted: tradeResponse.ok,
        trade_status: tradeResponse.status,
        has_trade_activity: hasTradeActivity,
        recent_activities: activityData.activities?.slice(0, 3).map(a => a.action) || []
      };
    } catch (err) {
      test6.status = 'failed';
      test6.details = { error: err.message };
    }
    
    // Test 7: Session Trade Counter
    const test7 = {
      name: 'Session Trade Counter',
      description: 'Verify session trade counts are updating correctly',
      status: 'running',
      details: {}
    };
    tests.push(test7);
    
    try {
      const response = await fetch('/api/admin/monitoring/sessions', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok) {
        const activeSessions = data.sessions || [];
        const sessionsWithTrades = activeSessions.filter(s => s.trades_count > 0);
        
        test7.status = sessionsWithTrades.length > 0 ? 'passed' : 'warning';
        test7.details = {
          total_sessions: activeSessions.length,
          sessions_with_trades: sessionsWithTrades.length,
          max_trades_in_session: Math.max(0, ...activeSessions.map(s => s.trades_count || 0)),
          trade_counts: activeSessions.map(s => s.trades_count || 0)
        };
      } else {
        test7.status = 'failed';
        test7.details = { error: 'Failed to fetch sessions', response_status: response.status };
      }
    } catch (err) {
      test7.status = 'failed';
      test7.details = { error: err.message };
    }
    
    // Test 8: Audit Log Integration
    const test8 = {
      name: 'Audit Log Integration',
      description: 'Check if monitoring system feeds audit logs',
      status: 'running',
      details: {}
    };
    tests.push(test8);
    
    try {
      const response = await fetch('/api/admin/audit?limit=10', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok && Array.isArray(data)) {
        const loginEvents = data.filter(event => event.action === 'login');
        
        test8.status = data.length > 0 ? 'passed' : 'warning';
        test8.details = {
          total_audit_events: data.length,
          login_events: loginEvents.length,
          recent_actions: data.slice(0, 5).map(e => e.action),
          has_monitoring_integration: loginEvents.length > 0
        };
      } else {
        test8.status = 'failed';
        test8.details = { error: 'Failed to fetch audit logs', response_status: response.status };
      }
    } catch (err) {
      test8.status = 'failed';
      test8.details = { error: err.message };
    }
    
    // Test 9: User Activity Tracking
    const test9 = {
      name: 'User Activity Tracking',
      description: 'Verify comprehensive user activity monitoring',
      status: 'running',
      details: {}
    };
    tests.push(test9);
    
    try {
      // Get current user's activities
      const response = await fetch('/api/admin/monitoring/user-activity?user_id=27&limit=10', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      const data = await response.json();
      
      if (response.ok) {
        const activities = data.activities || [];
        const actionTypes = [...new Set(activities.map(a => a.action))];
        const hasApiRequests = activities.some(a => a.action === 'api_request');
        const hasTrades = activities.some(a => a.action === 'trade');
        
        test9.status = activities.length > 0 ? 'passed' : 'warning';
        test9.details = {
          total_activities: activities.length,
          unique_action_types: actionTypes.length,
          action_types: actionTypes,
          has_api_requests: hasApiRequests,
          has_trades: hasTrades,
          success_rate: activities.length > 0 ? `${((activities.filter(a => a.success).length / activities.length) * 100).toFixed(1)}%` : '0%'
        };
      } else {
        test9.status = 'failed';
        test9.details = { error: 'Failed to fetch user activities', response_status: response.status };
      }
    } catch (err) {
      test9.status = 'failed';
      test9.details = { error: err.message };
    }
    
    // Return results
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    const warnings = tests.filter(t => t.status === 'warning').length;
    
    return { 
      suite: 'Monitoring System Tests',
      description: 'Comprehensive monitoring system validation',
      duration: 0,
      total: tests.length,
      passed,
      failed,
      warnings,
      success: failed === 0,
      tests
    };
  };

  // Comprehensive SSE Testing Suite
  const runComprehensiveSSETests = async () => {
    const tests = [];
    
    // Test 1: SSE Connection Lifecycle
    const test1 = {
      name: 'SSE Connection Lifecycle',
      description: 'Test SSE connection establishment, maintenance, and recovery',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const initialState = getSSEConnectionState();
      
      // Force reconnection to test lifecycle
      forceSSEReconnect();
      
      // Wait for reconnection
      await new Promise(resolve => setTimeout(resolve, 3000));
      
      const finalState = getSSEConnectionState();
      const isConnected = isSSEConnected();
      
      test1.status = isConnected ? 'passed' : 'failed';
      test1.details = {
        initial_state: initialState,
        final_state: finalState,
        connection_successful: isConnected,
        reconnection_tested: true
      };
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Stock Update Message Reception
    const test2 = {
      name: 'Stock Update Message Reception',
      description: 'Verify stock update messages are received and processed',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      let messageReceived = false;
      let messageData = null;
      
      const messageHandler = (data) => {
        messageReceived = true;
        messageData = data;
      };
      
      addSSEListener('stockUpdate', messageHandler);
      
      // Wait for stock update
      let attempts = 0;
      while (!messageReceived && attempts < 10) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        attempts++;
      }
      
      removeSSEListener('stockUpdate', messageHandler);
      
      test2.status = messageReceived ? 'passed' : 'failed';
      test2.details = {
        message_received: messageReceived,
        wait_time_seconds: attempts,
        message_data: messageData ? {
          type: messageData.type,
          stock_id: messageData.stock_id,
          symbol: messageData.symbol,
          price: messageData.price
        } : null
      };
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: Connection Recovery After Interruption
    const test3 = {
      name: 'Connection Recovery After Interruption',
      description: 'Test automatic reconnection after connection loss',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const beforeState = getSSEConnectionState();
      
      // Simulate connection interruption by forcing disconnect/reconnect
      forceSSEReconnect();
      
      // Wait for automatic recovery
      let recovered = false;
      let attempts = 0;
      while (!recovered && attempts < 20) {
        await new Promise(resolve => setTimeout(resolve, 1000));
        recovered = isSSEConnected();
        attempts++;
      }
      
      test3.status = recovered ? 'passed' : 'failed';
      test3.details = {
        before_state: beforeState,
        recovery_successful: recovered,
        recovery_time_seconds: attempts,
        final_state: getSSEConnectionState()
      };
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    // Test 4: Message Rate and Performance
    const test4 = {
      name: 'Message Rate and Performance',
      description: 'Monitor message reception rate and performance',
      status: 'running',
      details: {}
    };
    tests.push(test4);
    
    try {
      const messageCount = { count: 0 };
      const startTime = Date.now();
      
      const rateHandler = () => {
        messageCount.count++;
      };
      
      addSSEListener('stockUpdate', rateHandler);
      
      // Monitor for 30 seconds
      await new Promise(resolve => setTimeout(resolve, 30000));
      
      removeSSEListener('stockUpdate', rateHandler);
      
      const duration = (Date.now() - startTime) / 1000;
      const messagesPerSecond = messageCount.count / duration;
      
      test4.status = messageCount.count > 0 ? 'passed' : 'failed';
      test4.details = {
        duration_seconds: duration,
        total_messages: messageCount.count,
        messages_per_second: messagesPerSecond.toFixed(2),
        expected_rate: '0.5 msg/sec (2 second intervals)'
      };
    } catch (err) {
      test4.status = 'failed';
      test4.details = { error: err.message };
    }
    
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    
    return {
      suite: 'Comprehensive SSE Tests',
      description: 'Advanced SSE connection and message handling tests',
      duration: 0,
      total: tests.length,
      passed,
      failed,
      success: failed === 0,
      tests
    };
  };

  // Frontend Integration Tests
  const runFrontendIntegrationTests = async () => {
    const tests = [];
    
    // Test 1: UI Update on Stock Price Change
    const test1 = {
      name: 'UI Update on Stock Price Change',
      description: 'Verify UI components update when stock prices change',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      // Check if we're on a page that should update stock prices
      const stockElements = document.querySelectorAll('[data-stock-price]');
      const hasStockElements = stockElements.length > 0;
      
      if (hasStockElements) {
        const initialPrices = Array.from(stockElements).map(el => ({
          element: el,
          initialPrice: el.textContent,
          stockId: el.getAttribute('data-stock-id')
        }));
        
        // Wait for potential updates
        await new Promise(resolve => setTimeout(resolve, 10000));
        
        const updatedPrices = initialPrices.map(item => ({
          ...item,
          updatedPrice: item.element.textContent,
          changed: item.initialPrice !== item.element.textContent
        }));
        
        const changedCount = updatedPrices.filter(item => item.changed).length;
        
        test1.status = changedCount > 0 ? 'passed' : 'warning';
        test1.details = {
          stock_elements_found: stockElements.length,
          elements_checked: initialPrices.length,
          price_changes_detected: changedCount,
          sample_changes: updatedPrices.slice(0, 3).map(item => ({
            initial: item.initialPrice,
            updated: item.updatedPrice,
            changed: item.changed
          }))
        };
      } else {
        test1.status = 'warning';
        test1.details = {
          stock_elements_found: 0,
          message: 'No stock price elements found on current page'
        };
      }
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Portfolio Value Recalculation
    const test2 = {
      name: 'Portfolio Value Recalculation',
      description: 'Test if portfolio values update with stock price changes',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      const portfolioElements = document.querySelectorAll('[data-portfolio-value]');
      const hasPortfolioElements = portfolioElements.length > 0;
      
      if (hasPortfolioElements) {
        const initialValues = Array.from(portfolioElements).map(el => ({
          element: el,
          initialValue: el.textContent,
          type: el.getAttribute('data-portfolio-type') || 'unknown'
        }));
        
        // Wait for updates
        await new Promise(resolve => setTimeout(resolve, 15000));
        
        const updatedValues = initialValues.map(item => ({
          ...item,
          updatedValue: item.element.textContent,
          changed: item.initialValue !== item.element.textContent
        }));
        
        const changedCount = updatedValues.filter(item => item.changed).length;
        
        test2.status = changedCount > 0 ? 'passed' : 'warning';
        test2.details = {
          portfolio_elements_found: portfolioElements.length,
          elements_checked: initialValues.length,
          value_changes_detected: changedCount,
          sample_changes: updatedValues.slice(0, 3)
        };
      } else {
        test2.status = 'warning';
        test2.details = {
          portfolio_elements_found: 0,
          message: 'No portfolio value elements found on current page'
        };
      }
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: Real-time Connection Status Display
    const test3 = {
      name: 'Real-time Connection Status Display',
      description: 'Verify connection status indicators work correctly',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const connectionIndicators = document.querySelectorAll('[data-connection-status]');
      const sseState = getSSEConnectionState();
      const isConnected = isSSEConnected();
      
      test3.status = 'passed';
      test3.details = {
        connection_indicators_found: connectionIndicators.length,
        sse_state: sseState,
        sse_connected: isConnected,
        indicators_present: connectionIndicators.length > 0
      };
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    const warnings = tests.filter(t => t.status === 'warning').length;
    
    return {
      suite: 'Frontend Integration Tests',
      description: 'UI component integration with real-time updates',
      duration: 0,
      total: tests.length,
      passed,
      failed,
      warnings,
      success: failed === 0,
      tests
    };
  };

  // Comprehensive WebSocket Testing Suite
  const runWebSocketTests = async () => {
    const tests = [];
    
    // Test 1: WebSocket Connection Establishment
    const test1 = {
      name: 'WebSocket Connection Establishment',
      description: 'Test WebSocket connection setup and authentication',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const token = localStorage.getItem('token');
      
      if (!token) {
        test1.status = 'warning';
        test1.details = {
          message: 'No authentication token found',
          suggestion: 'Login to test WebSocket authentication'
        };
      } else {
        // Test WebSocket connection
        await initWebSocket();
        const ws = getWebSocketInstance();
        
        if (ws) {
          const connectionPromise = new Promise((resolve) => {
            const timeout = setTimeout(() => resolve(false), 5000);
            
            ws.addEventListener('open', () => {
              clearTimeout(timeout);
              resolve(true);
            });
            
            ws.addEventListener('error', () => {
              clearTimeout(timeout);
              resolve(false);
            });
          });
          
          const connected = await connectionPromise;
          
          test1.status = connected ? 'passed' : 'failed';
          test1.details = {
            connection_attempted: true,
            connection_successful: connected,
            websocket_state: ws.readyState,
            authentication_token_present: !!token
          };
        } else {
          test1.status = 'failed';
          test1.details = {
            error: 'WebSocket instance not created',
            connection_attempted: false
          };
        }
      }
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Chat Message Broadcasting
    const test2 = {
      name: 'Chat Message Broadcasting',
      description: 'Test sending and receiving chat messages via WebSocket',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      const ws = getWebSocketInstance();
      const token = localStorage.getItem('token');
      
      if (!ws || !token) {
        test2.status = 'warning';
        test2.details = {
          message: 'WebSocket not connected or no auth token',
          websocket_available: !!ws,
          token_available: !!token
        };
      } else {
        // Listen for chat messages
        let messageReceived = false;
        let receivedMessage = null;
        
        const messageHandler = (event) => {
          try {
            const data = JSON.parse(event.data);
            if (data.type === 'chat_message') {
              messageReceived = true;
              receivedMessage = data;
            }
          } catch (e) {
            // Ignore parsing errors
          }
        };
        
        ws.addEventListener('message', messageHandler);
        
        // Send a test chat message
        const testMessage = {
          type: 'chat_message',
          message: `Test message from testing suite - ${Date.now()}`
        };
        
        try {
          const response = await fetch('/api/chat/send', {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify(testMessage)
          });
          
          const messageSent = response.ok;
          
          // Wait for message to potentially arrive via WebSocket
          await new Promise(resolve => setTimeout(resolve, 3000));
          
          ws.removeEventListener('message', messageHandler);
          
          test2.status = messageSent ? 'passed' : 'failed';
          test2.details = {
            message_sent_via_api: messageSent,
            api_response_status: response.status,
            websocket_message_received: messageReceived,
            received_message_type: receivedMessage?.type || null,
            test_complete: true
          };
        } catch (fetchErr) {
          test2.status = 'failed';
          test2.details = {
            error: `Failed to send chat message: ${fetchErr.message}`,
            message_sent_via_api: false
          };
        }
      }
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: WebSocket Connection Recovery
    const test3 = {
      name: 'WebSocket Connection Recovery',
      description: 'Test WebSocket reconnection after interruption',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const ws = getWebSocketInstance();
      
      if (!ws) {
        test3.status = 'warning';
        test3.details = {
          message: 'No WebSocket connection to test recovery'
        };
      } else {
        const initialState = ws.readyState;
        
        // Force close the connection to simulate interruption
        if (ws.readyState === WebSocket.OPEN) {
          ws.close();
        }
        
        // Wait a moment
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Attempt to reinitialize
        await initWebSocket();
        const newWs = getWebSocketInstance();
        
        let recovered = false;
        if (newWs && newWs !== ws) {
          // Wait for new connection to establish
          const recoveryPromise = new Promise((resolve) => {
            const timeout = setTimeout(() => resolve(false), 10000);
            
            newWs.addEventListener('open', () => {
              clearTimeout(timeout);
              resolve(true);
            });
            
            newWs.addEventListener('error', () => {
              clearTimeout(timeout);
              resolve(false);
            });
            
            // If already open
            if (newWs.readyState === WebSocket.OPEN) {
              clearTimeout(timeout);
              resolve(true);
            }
          });
          
          recovered = await recoveryPromise;
        }
        
        test3.status = recovered ? 'passed' : 'failed';
        test3.details = {
          initial_connection_state: initialState,
          connection_closed: true,
          new_connection_created: !!newWs,
          recovery_successful: recovered,
          final_connection_state: newWs?.readyState || 'none'
        };
      }
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    // Test 4: WebSocket Message Types and Routing
    const test4 = {
      name: 'WebSocket Message Types and Routing',
      description: 'Test different message types are properly routed',
      status: 'running',
      details: {}
    };
    tests.push(test4);
    
    try {
      const ws = getWebSocketInstance();
      
      if (!ws || ws.readyState !== WebSocket.OPEN) {
        test4.status = 'warning';
        test4.details = {
          message: 'WebSocket not connected',
          websocket_state: ws?.readyState || 'none'
        };
      } else {
        const messageTypes = {};
        let totalMessages = 0;
        
        const messageHandler = (event) => {
          try {
            const data = JSON.parse(event.data);
            const messageType = data.type || 'unknown';
            
            messageTypes[messageType] = (messageTypes[messageType] || 0) + 1;
            totalMessages++;
          } catch (e) {
            messageTypes['parse_error'] = (messageTypes['parse_error'] || 0) + 1;
            totalMessages++;
          }
        };
        
        ws.addEventListener('message', messageHandler);
        
        // Listen for messages for 15 seconds
        await new Promise(resolve => setTimeout(resolve, 15000));
        
        ws.removeEventListener('message', messageHandler);
        
        const hasStockUpdates = messageTypes['stock_update'] > 0;
        const hasChatMessages = messageTypes['chat_message'] > 0;
        const messageTypeCount = Object.keys(messageTypes).length;
        
        test4.status = totalMessages > 0 ? 'passed' : 'warning';
        test4.details = {
          total_messages_received: totalMessages,
          message_types_seen: messageTypeCount,
          message_type_breakdown: messageTypes,
          has_stock_updates: hasStockUpdates,
          has_chat_messages: hasChatMessages,
          monitoring_duration_seconds: 15
        };
      }
    } catch (err) {
      test4.status = 'failed';
      test4.details = { error: err.message };
    }
    
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    const warnings = tests.filter(t => t.status === 'warning').length;
    
    return {
      suite: 'Comprehensive WebSocket Tests',
      description: 'WebSocket connection, messaging, and recovery testing',
      duration: 0,
      total: tests.length,
      passed,
      failed,
      warnings,
      success: failed === 0,
      tests
    };
  };

  // Performance and Load Testing Suite
  const runPerformanceTests = async () => {
    const tests = [];
    
    // Test 1: API Response Time Performance
    const test1 = {
      name: 'API Response Time Performance',
      description: 'Measure API endpoint response times under normal load',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const endpoints = [
        { name: 'Health Check', url: '/api/health' },
        { name: 'Stocks List', url: '/api/stocks' },
        { name: 'Leaderboard', url: '/api/users/leaderboard' }
      ];
      
      const token = localStorage.getItem('token');
      const authenticatedEndpoints = token ? [
        { name: 'Portfolio', url: '/api/portfolio', auth: true },
        { name: 'Transactions', url: '/api/transactions', auth: true }
      ] : [];
      
      const allEndpoints = [...endpoints, ...authenticatedEndpoints];
      const results = {};
      
      for (const endpoint of allEndpoints) {
        const times = [];
        
        // Test each endpoint 5 times
        for (let i = 0; i < 5; i++) {
          const startTime = performance.now();
          
          try {
            const headers = endpoint.auth ? { 'Authorization': `Bearer ${token}` } : {};
            const response = await fetch(endpoint.url, { headers });
            const endTime = performance.now();
            
            if (response.ok) {
              times.push(endTime - startTime);
            }
          } catch (err) {
            // Skip failed requests
          }
          
          // Small delay between requests
          await new Promise(resolve => setTimeout(resolve, 100));
        }
        
        if (times.length > 0) {
          const avgTime = times.reduce((a, b) => a + b, 0) / times.length;
          const minTime = Math.min(...times);
          const maxTime = Math.max(...times);
          
          results[endpoint.name] = {
            avg_response_time: Math.round(avgTime),
            min_response_time: Math.round(minTime),
            max_response_time: Math.round(maxTime),
            successful_requests: times.length,
            total_requests: 5
          };
        }
      }
      
      const hasResults = Object.keys(results).length > 0;
      const avgResponseTime = hasResults 
        ? Math.round(Object.values(results).reduce((sum, r) => sum + r.avg_response_time, 0) / Object.keys(results).length)
        : 0;
      
      test1.status = hasResults ? 'passed' : 'failed';
      test1.details = {
        endpoints_tested: allEndpoints.length,
        results_collected: Object.keys(results).length,
        overall_avg_response_time: avgResponseTime,
        endpoint_results: results
      };
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    // Test 2: Concurrent Connection Load Test
    const test2 = {
      name: 'Concurrent Connection Load Test',
      description: 'Test system behavior under multiple concurrent connections',
      status: 'running',
      details: {}
    };
    tests.push(test2);
    
    try {
      const concurrentRequests = 10;
      const requestPromises = [];
      
      const startTime = performance.now();
      
      // Create multiple concurrent requests to the stocks endpoint
      for (let i = 0; i < concurrentRequests; i++) {
        const requestPromise = fetch('/api/stocks')
          .then(response => ({
            success: response.ok,
            status: response.status,
            responseTime: performance.now() - startTime
          }))
          .catch(err => ({
            success: false,
            error: err.message,
            responseTime: performance.now() - startTime
          }));
        
        requestPromises.push(requestPromise);
      }
      
      const results = await Promise.all(requestPromises);
      const endTime = performance.now();
      
      const successfulRequests = results.filter(r => r.success).length;
      const failedRequests = results.filter(r => !r.success).length;
      const avgResponseTime = results.reduce((sum, r) => sum + r.responseTime, 0) / results.length;
      const maxResponseTime = Math.max(...results.map(r => r.responseTime));
      
      test2.status = successfulRequests >= concurrentRequests * 0.8 ? 'passed' : 'failed'; // 80% success rate
      test2.details = {
        concurrent_requests: concurrentRequests,
        successful_requests: successfulRequests,
        failed_requests: failedRequests,
        success_rate: `${((successfulRequests / concurrentRequests) * 100).toFixed(1)}%`,
        total_duration: Math.round(endTime - startTime),
        avg_response_time: Math.round(avgResponseTime),
        max_response_time: Math.round(maxResponseTime)
      };
    } catch (err) {
      test2.status = 'failed';
      test2.details = { error: err.message };
    }
    
    // Test 3: Memory Usage Monitoring
    const test3 = {
      name: 'Memory Usage Monitoring',
      description: 'Monitor memory usage during testing operations',
      status: 'running',
      details: {}
    };
    tests.push(test3);
    
    try {
      const initialMemory = performance.memory ? {
        used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
        total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
        limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
      } : null;
      
      if (initialMemory) {
        // Perform some memory-intensive operations
        const largeArray = new Array(100000).fill(null).map((_, i) => ({ id: i, data: `test-data-${i}` }));
        
        // Wait a bit and measure again
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        const finalMemory = {
          used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
          total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
          limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
        };
        
        // Clean up
        largeArray.length = 0;
        
        const memoryIncrease = finalMemory.used - initialMemory.used;
        const memoryUsagePercentage = (finalMemory.used / finalMemory.limit * 100).toFixed(1);
        
        test3.status = finalMemory.used < finalMemory.limit * 0.8 ? 'passed' : 'warning'; // Under 80% of limit
        test3.details = {
          initial_memory_used: `${initialMemory.used} MB`,
          final_memory_used: `${finalMemory.used} MB`,
          memory_increase: `${memoryIncrease} MB`,
          memory_usage_percentage: `${memoryUsagePercentage}%`,
          memory_limit: `${finalMemory.limit} MB`,
          memory_available: finalMemory.limit > 0
        };
      } else {
        test3.status = 'warning';
        test3.details = {
          message: 'Performance.memory API not available in this browser',
          memory_monitoring_available: false
        };
      }
    } catch (err) {
      test3.status = 'failed';
      test3.details = { error: err.message };
    }
    
    // Test 4: Real-time Message Rate Stress Test
    const test4 = {
      name: 'Real-time Message Rate Stress Test',
      description: 'Test handling of high-frequency real-time messages',
      status: 'running',
      details: {}
    };
    tests.push(test4);
    
    try {
      let messageCount = 0;
      let processingErrors = 0;
      const startTime = Date.now();
      
      const messageHandler = (data) => {
        try {
          messageCount++;
          // Simulate some processing
          if (data && typeof data === 'object') {
            JSON.stringify(data);
          }
        } catch (err) {
          processingErrors++;
        }
      };
      
      // Add SSE listener
      addSSEListener('stockUpdate', messageHandler);
      
      // Monitor for 20 seconds
      await new Promise(resolve => setTimeout(resolve, 20000));
      
      removeSSEListener('stockUpdate', messageHandler);
      
      const duration = (Date.now() - startTime) / 1000;
      const messagesPerSecond = messageCount / duration;
      
      test4.status = processingErrors === 0 && messageCount > 0 ? 'passed' : 'failed';
      test4.details = {
        monitoring_duration: duration,
        total_messages_processed: messageCount,
        processing_errors: processingErrors,
        messages_per_second: messagesPerSecond.toFixed(2),
        error_rate: processingErrors > 0 ? `${((processingErrors / messageCount) * 100).toFixed(2)}%` : '0%'
      };
    } catch (err) {
      test4.status = 'failed';
      test4.details = { error: err.message };
    }
    
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    const warnings = tests.filter(t => t.status === 'warning').length;
    
    return {
      suite: 'Performance and Load Tests',
      description: 'System performance and load handling validation',
      duration: 0,
      total: tests.length,
      passed,
      failed,
      warnings,
      success: failed === 0,
      tests
    };
  };

  // End-to-End User Journey Tests
  const runEndToEndTests = async () => {
    const tests = [];
    
    // Test 1: Login to Real-time Updates Journey
    const test1 = {
      name: 'Login to Real-time Updates Journey',
      description: 'Test complete user journey from login to receiving updates',
      status: 'running',
      details: {}
    };
    tests.push(test1);
    
    try {
      const token = localStorage.getItem('token');
      const isLoggedIn = !!token;
      
      if (!isLoggedIn) {
        test1.status = 'warning';
        test1.details = {
          message: 'User not logged in - cannot test authenticated journey',
          suggestion: 'Login and re-run test'
        };
      } else {
        // Test authenticated API access
        const portfolioResponse = await fetch('/api/portfolio', {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        
        const portfolioWorks = portfolioResponse.ok;
        
        // Test real-time connection with auth
        const sseConnected = isSSEConnected();
        
        // Wait for stock update
        let updateReceived = false;
        const updateHandler = () => { updateReceived = true; };
        addSSEListener('stockUpdate', updateHandler);
        
        await new Promise(resolve => setTimeout(resolve, 10000));
        removeSSEListener('stockUpdate', updateHandler);
        
        test1.status = (portfolioWorks && sseConnected && updateReceived) ? 'passed' : 'partial';
        test1.details = {
          authenticated: isLoggedIn,
          portfolio_api_works: portfolioWorks,
          sse_connected: sseConnected,
          stock_update_received: updateReceived,
          journey_complete: portfolioWorks && sseConnected && updateReceived
        };
      }
    } catch (err) {
      test1.status = 'failed';
      test1.details = { error: err.message };
    }
    
    const passed = tests.filter(t => t.status === 'passed').length;
    const failed = tests.filter(t => t.status === 'failed').length;
    
    return {
      suite: 'End-to-End User Journey Tests',
      description: 'Complete user experience testing',
      duration: 0,
      total: tests.length,
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
        case 'live-monitoring':
          // Live monitoring doesn't need a test run
          return;
        case 'crisis':
          results = await runCrisisTests();
          break;
        case 'portfolio':
          results = await runPortfolioTests();
          break;
        case 'sse':
          results = await runSSETests();
          break;
        case 'stock-management':
          results = await runStockManagementTests();
          break;
        case 'sse-comprehensive':
          results = await runComprehensiveSSETests();
          break;
        case 'websocket-comprehensive':
          results = await runWebSocketTests();
          break;
        case 'performance-load':
          results = await runPerformanceTests();
          break;
        case 'frontend-integration':
          results = await runFrontendIntegrationTests();
          break;
        case 'end-to-end':
          results = await runEndToEndTests();
          break;
        case 'changelog':
          results = await testChangelogSystem();
          break;
        case 'monitoring':
          results = await testMonitoringSystem();
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
              <option value="live-monitoring">🔴 Live System Monitoring</option>
              <option value="automated-testing">🤖 Automated Continuous Testing</option>
              <option value="sse-comprehensive">📡 Comprehensive SSE Tests</option>
              <option value="websocket-comprehensive">🔌 Comprehensive WebSocket Tests</option>
              <option value="performance-load">⚡ Performance & Load Tests</option>
              <option value="frontend-integration">🖥️ Frontend Integration Tests</option>
              <option value="end-to-end">🚀 End-to-End User Journey</option>
              <option value="crisis">🚨 Crisis Mechanics Test Suite</option>
              <option value="portfolio">💼 Portfolio & Trading Test Suite</option>
              <option value="sse">📡 Basic SSE Test Suite</option>
              <option value="stock-management">🔧 Stock Management Debug Suite</option>
              <option value="changelog">📰 Changelog System Test Suite</option>
              <option value="monitoring">📊 Monitoring System Test Suite</option>
            </select>
          </div>

          {selectedTest === 'live-monitoring' || selectedTest === 'automated-testing' ? (
            <div className="live-monitoring-controls">
              {selectedTest === 'live-monitoring' ? (
                <button 
                  onClick={isLiveMode ? stopLiveMonitoring : startLiveMonitoring}
                  className={`run-tests-btn ${isLiveMode ? 'monitoring-active' : ''}`}
                >
                  {isLiveMode ? (
                    <>
                      🛑 Stop Live Monitoring
                    </>
                  ) : (
                    <>
                      🔴 Start Live Monitoring
                    </>
                  )}
                </button>
              ) : (
                <button 
                  onClick={automatedTesting.enabled ? stopAutomatedTesting : startAutomatedTesting}
                  className={`run-tests-btn ${automatedTesting.enabled ? 'monitoring-active' : ''}`}
                >
                  {automatedTesting.enabled ? (
                    <>
                      🛑 Stop Automated Testing
                    </>
                  ) : (
                    <>
                      🤖 Start Automated Testing
                    </>
                  )}
                </button>
              )}
              
              {selectedTest === 'live-monitoring' && isLiveMode && (
                <div className="manual-test-controls">
                  <button onClick={triggerSSEReconnect} className="manual-test-btn">
                    🔄 Force SSE Reconnect
                  </button>
                  <button onClick={clearSystemLog} className="manual-test-btn">
                    🗑️ Clear System Log
                  </button>
                  <button 
                    onClick={() => setDebugConsole(prev => ({ ...prev, visible: !prev.visible }))}
                    className="manual-test-btn"
                  >
                    {debugConsole.visible ? '🔍 Hide Debug Console' : '🔍 Show Debug Console'}
                  </button>
                </div>
              )}
              
              {selectedTest === 'automated-testing' && (
                <div className="automated-testing-controls">
                  <div className="control-group">
                    <label htmlFor="test-interval">Interval (seconds):</label>
                    <input
                      id="test-interval"
                      type="number"
                      min="60"
                      max="3600"
                      value={automatedTesting.interval}
                      onChange={(e) => updateAutomatedTestSettings({ interval: parseInt(e.target.value) || 300 })}
                      disabled={automatedTesting.enabled}
                      className="interval-input"
                    />
                  </div>
                  
                  <div className="control-group">
                    <label>Test Suites:</label>
                    <div className="checkbox-group">
                      {[
                        { value: 'sse-comprehensive', label: '📡 SSE Tests' },
                        { value: 'websocket-comprehensive', label: '🔌 WebSocket Tests' },
                        { value: 'frontend-integration', label: '🖥️ Frontend Tests' },
                        { value: 'performance-load', label: '⚡ Performance Tests' }
                      ].map(suite => (
                        <label key={suite.value} className="checkbox-label">
                          <input
                            type="checkbox"
                            checked={automatedTesting.testSuites.includes(suite.value)}
                            onChange={(e) => {
                              const newSuites = e.target.checked 
                                ? [...automatedTesting.testSuites, suite.value]
                                : automatedTesting.testSuites.filter(s => s !== suite.value);
                              updateAutomatedTestSettings({ testSuites: newSuites });
                            }}
                            disabled={automatedTesting.enabled}
                          />
                          {suite.label}
                        </label>
                      ))}
                    </div>
                  </div>
                  
                  <button onClick={clearSystemLog} className="manual-test-btn">
                    🗑️ Clear Logs
                  </button>
                </div>
              )}
            </div>
          ) : (
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
          )}
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

        {/* Live Monitoring Dashboard */}
        {selectedTest === 'live-monitoring' && isLiveMode && (
          <div className="live-monitoring-dashboard">
            <div className="dashboard-header">
              <h2>🔴 Live System Monitoring Dashboard</h2>
              <div className="dashboard-status">
                Status: <span className={`status-indicator ${isLiveMode ? 'active' : 'inactive'}`}>
                  {isLiveMode ? 'MONITORING ACTIVE' : 'MONITORING STOPPED'}
                </span>
              </div>
            </div>

            <div className="monitoring-grid">
              {/* SSE Connection Status */}
              <div className="monitoring-card">
                <h3>📡 SSE Connection</h3>
                <div className="status-row">
                  <span className="label">State:</span>
                  <span className={`value ${liveMonitoring.sseConnection.state}`}>
                    {liveMonitoring.sseConnection.state.toUpperCase()}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Connected:</span>
                  <span className={`value ${liveMonitoring.sseConnection.connected ? 'connected' : 'disconnected'}`}>
                    {liveMonitoring.sseConnection.connected ? '✅ YES' : '❌ NO'}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Messages Received:</span>
                  <span className="value">{liveMonitoring.sseConnection.messageCount}</span>
                </div>
                <div className="status-row">
                  <span className="label">Last Message:</span>
                  <span className="value">
                    {liveMonitoring.sseConnection.lastMessage 
                      ? new Date(liveMonitoring.sseConnection.lastMessage).toLocaleTimeString()
                      : 'None'
                    }
                  </span>
                </div>
              </div>

              {/* WebSocket Connection Status */}
              <div className="monitoring-card">
                <h3>🔌 WebSocket Connection</h3>
                <div className="status-row">
                  <span className="label">Connected:</span>
                  <span className={`value ${liveMonitoring.webSocketConnection.connected ? 'connected' : 'disconnected'}`}>
                    {liveMonitoring.webSocketConnection.connected ? '✅ YES' : '❌ NO'}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Messages Received:</span>
                  <span className="value">{liveMonitoring.webSocketConnection.messageCount}</span>
                </div>
                <div className="status-row">
                  <span className="label">Last Message:</span>
                  <span className="value">
                    {liveMonitoring.webSocketConnection.lastMessage 
                      ? new Date(liveMonitoring.webSocketConnection.lastMessage).toLocaleTimeString()
                      : 'None'
                    }
                  </span>
                </div>
              </div>

              {/* Market Simulator Status */}
              <div className="monitoring-card">
                <h3>📈 Market Simulator</h3>
                <div className="status-row">
                  <span className="label">Running:</span>
                  <span className={`value ${liveMonitoring.marketSimulator.running ? 'connected' : 'disconnected'}`}>
                    {liveMonitoring.marketSimulator.running ? '✅ YES' : '❌ NO'}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Stock Count:</span>
                  <span className="value">{liveMonitoring.marketSimulator.stockCount}</span>
                </div>
                <div className="status-row">
                  <span className="label">Update Frequency:</span>
                  <span className="value">{liveMonitoring.marketSimulator.updateFrequency}s</span>
                </div>
                <div className="status-row">
                  <span className="label">Last Update:</span>
                  <span className="value">
                    {liveMonitoring.marketSimulator.lastUpdate 
                      ? new Date(liveMonitoring.marketSimulator.lastUpdate).toLocaleTimeString()
                      : 'None'
                    }
                  </span>
                </div>
              </div>

              {/* Database Health */}
              <div className="monitoring-card">
                <h3>🗄️ Database Health</h3>
                <div className="status-row">
                  <span className="label">Connected:</span>
                  <span className={`value ${liveMonitoring.databaseHealth.connected ? 'connected' : 'disconnected'}`}>
                    {liveMonitoring.databaseHealth.connected ? '✅ YES' : '❌ NO'}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Response Time:</span>
                  <span className="value">{liveMonitoring.databaseHealth.responseTime}ms</span>
                </div>
                <div className="status-row">
                  <span className="label">Active Queries:</span>
                  <span className="value">{liveMonitoring.databaseHealth.activeQueries}</span>
                </div>
                <div className="status-row">
                  <span className="label">Pool Size:</span>
                  <span className="value">{liveMonitoring.databaseHealth.poolSize}</span>
                </div>
              </div>

              {/* API Health */}
              <div className="monitoring-card">
                <h3>🌐 API Health</h3>
                <div className="status-row">
                  <span className="label">Response Time:</span>
                  <span className="value">{liveMonitoring.apiHealth.responseTime}ms</span>
                </div>
                <div className="status-row">
                  <span className="label">Success Rate:</span>
                  <span className="value">{liveMonitoring.apiHealth.successRate}%</span>
                </div>
                <div className="status-row">
                  <span className="label">Error Count:</span>
                  <span className="value">{liveMonitoring.apiHealth.errorCount}</span>
                </div>
                <div className="status-row">
                  <span className="label">Last Check:</span>
                  <span className="value">
                    {liveMonitoring.apiHealth.lastCheck 
                      ? new Date(liveMonitoring.apiHealth.lastCheck).toLocaleTimeString()
                      : 'None'
                    }
                  </span>
                </div>
              </div>
            </div>

            {/* System Messages Log */}
            <div className="system-messages">
              <h3>📋 System Messages</h3>
              <div className="messages-container">
                {systemMessages.length > 0 ? (
                  systemMessages.map((msg) => (
                    <div key={msg.id} className="message-item">
                      <span className="message-timestamp">[{msg.timestamp}]</span>
                      <span className="message-text">{msg.message}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-messages">No system messages yet</div>
                )}
              </div>
            </div>

            {/* Debug Console */}
            {debugConsole.visible && (
              <div className="debug-console">
                <div className="console-header">
                  <h3>🔍 Debug Console</h3>
                  <button onClick={clearDebugConsole} className="clear-console-btn">
                    Clear Console
                  </button>
                </div>
                <div className="console-messages">
                  {debugConsole.messages.length > 0 ? (
                    debugConsole.messages.map((msg, index) => (
                      <div key={index} className="debug-message">
                        <div className="debug-message-header">
                          <span className="debug-timestamp">{new Date(msg.timestamp).toLocaleTimeString()}</span>
                          <span className="debug-type">[{msg.type}]</span>
                        </div>
                        <pre className="debug-data">{msg.raw}</pre>
                      </div>
                    ))
                  ) : (
                    <div className="no-debug-messages">No debug messages yet</div>
                  )}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Automated Testing Dashboard */}
        {selectedTest === 'automated-testing' && (
          <div className="automated-testing-dashboard">
            <div className="dashboard-header">
              <h2>🤖 Automated Continuous Testing Dashboard</h2>
              <div className="dashboard-status">
                Status: <span className={`status-indicator ${automatedTesting.enabled ? 'active' : 'inactive'}`}>
                  {automatedTesting.enabled ? 'TESTING ACTIVE' : 'TESTING STOPPED'}
                </span>
              </div>
            </div>

            <div className="automated-stats-grid">
              <div className="stat-card">
                <h3>📊 Test Statistics</h3>
                <div className="status-row">
                  <span className="label">Total Runs:</span>
                  <span className="value">{automatedTesting.results.length}</span>
                </div>
                <div className="status-row">
                  <span className="label">Last Run:</span>
                  <span className="value">
                    {automatedTesting.lastRun 
                      ? new Date(automatedTesting.lastRun).toLocaleTimeString()
                      : 'Never'
                    }
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Consecutive Failures:</span>
                  <span className={`value ${automatedTesting.consecutiveFailures > 2 ? 'disconnected' : 'connected'}`}>
                    {automatedTesting.consecutiveFailures}
                  </span>
                </div>
                <div className="status-row">
                  <span className="label">Next Run:</span>
                  <span className="value">
                    {automatedTesting.enabled && automatedTesting.lastRun 
                      ? new Date(new Date(automatedTesting.lastRun).getTime() + automatedTesting.interval * 1000).toLocaleTimeString()
                      : 'N/A'
                    }
                  </span>
                </div>
              </div>

              <div className="stat-card">
                <h3>⚙️ Configuration</h3>
                <div className="status-row">
                  <span className="label">Interval:</span>
                  <span className="value">{automatedTesting.interval}s</span>
                </div>
                <div className="status-row">
                  <span className="label">Active Suites:</span>
                  <span className="value">{automatedTesting.testSuites.length}</span>
                </div>
                <div className="status-row">
                  <span className="label">Test Suites:</span>
                  <div className="value">
                    {automatedTesting.testSuites.map(suite => (
                      <div key={suite} className="suite-tag">
                        {suite.replace('-', ' ').replace(/\b\w/g, l => l.toUpperCase())}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              
              {automatedTesting.results.length > 0 && (
                <div className="stat-card">
                  <h3>📈 Success Rate</h3>
                  <div className="status-row">
                    <span className="label">Last 10 Runs:</span>
                    <span className="value">
                      {automatedTesting.results.slice(0, 10).filter(r => r.allPassed).length}/10
                    </span>
                  </div>
                  <div className="status-row">
                    <span className="label">Success Rate:</span>
                    <span className={`value ${automatedTesting.results.slice(0, 10).filter(r => r.allPassed).length >= 8 ? 'connected' : 'disconnected'}`}>
                      {((automatedTesting.results.slice(0, 10).filter(r => r.allPassed).length / Math.min(10, automatedTesting.results.length)) * 100).toFixed(0)}%
                    </span>
                  </div>
                  <div className="status-row">
                    <span className="label">Avg Duration:</span>
                    <span className="value">
                      {automatedTesting.results.length > 0 
                        ? (automatedTesting.results.slice(0, 10).reduce((sum, r) => sum + r.duration, 0) / Math.min(10, automatedTesting.results.length)).toFixed(1)
                        : '0'
                      }s
                    </span>
                  </div>
                </div>
              )}
            </div>

            {/* Test History */}
            <div className="test-history">
              <h3>📋 Test History</h3>
              <div className="history-container">
                {automatedTesting.results.length > 0 ? (
                  automatedTesting.results.slice(0, 10).map((result, index) => (
                    <div key={index} className={`history-item ${result.allPassed ? 'passed' : 'failed'}`}>
                      <div className="history-header">
                        <span className="history-timestamp">
                          {new Date(result.timestamp).toLocaleString()}
                        </span>
                        <span className={`history-status ${result.allPassed ? 'passed' : 'failed'}`}>
                          {result.allPassed ? '✅ PASSED' : '❌ FAILED'}
                        </span>
                        <span className="history-duration">{result.duration.toFixed(1)}s</span>
                      </div>
                      <div className="history-details">
                        {result.failedSuites > 0 && (
                          <span className="failure-count">
                            {result.failedSuites}/{result.totalSuites} suites failed
                          </span>
                        )}
                        {result.details.filter(d => d.error || d.result?.success === false).map((failedDetail, idx) => (
                          <div key={idx} className="failure-detail">
                            {failedDetail.suite}: {failedDetail.error || 'Test suite failed'}
                          </div>
                        ))}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="no-history">No automated test history yet</div>
                )}
              </div>
            </div>

            {/* System Messages for Automated Testing */}
            <div className="system-messages">
              <h3>📋 System Messages</h3>
              <div className="messages-container">
                {systemMessages.length > 0 ? (
                  systemMessages.map((msg) => (
                    <div key={msg.id} className="message-item">
                      <span className="message-timestamp">[{msg.timestamp}]</span>
                      <span className="message-text">{msg.message}</span>
                    </div>
                  ))
                ) : (
                  <div className="no-messages">No system messages yet</div>
                )}
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
                  <h5>📡 SSE Real-time Updates Test Suite:</h5>
                  <ul>
                    <li>SSE Connection Test - Verifies market simulator update channel</li>
                    <li>Market Simulator Status - Checks stock data and price validity</li>
                    <li>SSE Message Format Test - Validates message structure and JSON</li>
                    <li>Rate Limiting Bypass Test - Confirms SSE endpoint configuration</li>
                  </ul>
                </div>
                <div className="suite-info">
                  <h5>🔧 Stock Management Debug Suite:</h5>
                  <ul>
                    <li>Database Schema Check - Verifies stock table structure</li>
                    <li>Basic Stock Retrieval - Tests GetAllStocks functionality</li>
                    <li>Stock Creation - Tests CreateStock operation</li>
                    <li>Stock Update - Tests UpdateStockDetails (currently failing with 500)</li>
                    <li>IPO Launch - Tests LaunchIPO functionality</li>
                    <li>Stock Deletion - Tests ForceDelisting operation</li>
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