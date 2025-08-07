// Comprehensive WebSocket Testing Suite
// Runs automated tests to identify exact failure patterns and success rates

import websocketForensics from './websocketForensics';
import { getAuthToken } from './authBridge';

class WebSocketTester {
  constructor() {
    this.testResults = [];
    this.isRunning = false;
    this.listeners = [];
    this.testId = 0;
  }

  // Comprehensive reliability test - runs multiple connection attempts
  async runReliabilityTest(iterations = 100) {
    if (this.isRunning) {
      throw new Error('Test already running');
    }

    this.isRunning = true;
    const testSessionId = `reliability_${Date.now()}`;
    const results = {
      sessionId: testSessionId,
      startTime: Date.now(),
      totalAttempts: iterations,
      successful: 0,
      failed: 0,
      attempts: [],
      performance: {
        averageConnectionTime: 0,
        medianConnectionTime: 0,
        fastestConnection: Infinity,
        slowestConnection: 0
      },
      errorPatterns: {},
      recommendations: []
    };

    this.notifyListeners('testStarted', { sessionId: testSessionId, totalAttempts: iterations });

    for (let i = 0; i < iterations; i++) {
      const attemptResult = await this.runSingleConnectionTest(i + 1, iterations);
      results.attempts.push(attemptResult);
      
      if (attemptResult.success) {
        results.successful++;
        results.performance.fastestConnection = Math.min(results.performance.fastestConnection, attemptResult.duration);
        results.performance.slowestConnection = Math.max(results.performance.slowestConnection, attemptResult.duration);
      } else {
        results.failed++;
        // Track error patterns
        const errorCategory = attemptResult.error?.classification?.category || 'unknown';
        results.errorPatterns[errorCategory] = (results.errorPatterns[errorCategory] || 0) + 1;
      }

      // Notify progress every 10 attempts
      if (i % 10 === 9) {
        const progress = {
          completed: i + 1,
          total: iterations,
          successRate: (results.successful / (i + 1)) * 100,
          currentErrorPatterns: results.errorPatterns
        };
        this.notifyListeners('testProgress', progress);
        console.log(`🔬 WebSocket Test Progress: ${i + 1}/${iterations} - Success Rate: ${progress.successRate.toFixed(1)}%`);
      }

      // Small delay between attempts to avoid overwhelming the server
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    // Calculate final statistics
    results.endTime = Date.now();
    results.totalDuration = results.endTime - results.startTime;
    results.successRate = (results.successful / results.totalAttempts) * 100;
    
    if (results.successful > 0) {
      const connectionTimes = results.attempts
        .filter(a => a.success)
        .map(a => a.duration)
        .sort((a, b) => a - b);
      
      results.performance.averageConnectionTime = connectionTimes.reduce((a, b) => a + b, 0) / connectionTimes.length;
      results.performance.medianConnectionTime = connectionTimes[Math.floor(connectionTimes.length / 2)];
    }

    // Generate recommendations
    results.recommendations = this.generateTestRecommendations(results);

    this.testResults.push(results);
    this.isRunning = false;

    this.notifyListeners('testCompleted', results);
    console.log('🎯 WebSocket Reliability Test Completed:', results);
    
    return results;
  }

  // Single connection test with detailed timing
  async runSingleConnectionTest(attemptNumber, totalAttempts) {
    const testId = `test_${this.testId++}_${Date.now()}`;
    const startTime = Date.now();
    
    console.log(`🧪 Running connection test ${attemptNumber}/${totalAttempts} (ID: ${testId})`);

    try {
      const token = await getAuthToken();
      if (!token) {
        return {
          testId,
          attemptNumber,
          success: false,
          duration: Date.now() - startTime,
          error: { message: 'No authentication token', category: 'authentication' },
          stages: ['token_failure']
        };
      }

      const wsUrl = this.getTestWebSocketURL(token);
      const result = await this.performConnectionTest(wsUrl, testId, startTime);
      
      result.testId = testId;
      result.attemptNumber = attemptNumber;
      
      return result;
    } catch (error) {
      return {
        testId,
        attemptNumber,
        success: false,
        duration: Date.now() - startTime,
        error: {
          message: error.message,
          category: 'test_exception'
        },
        stages: ['exception']
      };
    }
  }

  // Perform the actual connection test with timing
  async performConnectionTest(wsUrl, testId, startTime) {
    return new Promise((resolve) => {
      const stages = [];
      const errors = [];
      let socket = null;
      let isResolved = false;

      const resolveTest = (success, finalData = {}) => {
        if (isResolved) return;
        isResolved = true;
        
        if (socket) {
          try {
            socket.close();
          } catch (e) {
            // Ignore close errors
          }
        }
        
        resolve({
          success,
          duration: Date.now() - startTime,
          stages,
          errors,
          ...finalData
        });
      };

      // Set a timeout for the test
      const timeout = setTimeout(() => {
        resolveTest(false, {
          error: {
            message: 'Connection test timeout',
            category: 'timeout'
          }
        });
      }, 30000); // 30 second timeout

      try {
        stages.push({ stage: 'websocket_creation', timestamp: Date.now() });
        socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          clearTimeout(timeout);
          stages.push({ stage: 'connection_opened', timestamp: Date.now() });
          resolveTest(true, {
            protocol: socket.protocol,
            readyState: socket.readyState
          });
        };

        socket.onerror = (error) => {
          stages.push({ stage: 'error_occurred', timestamp: Date.now() });
          errors.push({
            error: error,
            timestamp: Date.now(),
            classification: this.classifyError(error)
          });
        };

        socket.onclose = (event) => {
          clearTimeout(timeout);
          stages.push({ 
            stage: 'connection_closed', 
            timestamp: Date.now(),
            code: event.code,
            reason: event.reason 
          });
          
          if (!isResolved) {
            resolveTest(false, {
              error: {
                message: `Connection closed: ${event.code} - ${event.reason}`,
                category: this.categorizeCloseCode(event.code),
                code: event.code,
                reason: event.reason
              }
            });
          }
        };

      } catch (error) {
        clearTimeout(timeout);
        stages.push({ stage: 'creation_exception', timestamp: Date.now() });
        resolveTest(false, {
          error: {
            message: error.message,
            category: 'creation_exception'
          }
        });
      }
    });
  }

  // Stress test - multiple simultaneous connections
  async runStressTest(concurrentConnections = 10, iterations = 5) {
    const results = {
      testType: 'stress',
      startTime: Date.now(),
      concurrentConnections,
      iterations,
      results: [],
      overallStats: {
        totalAttempts: 0,
        successful: 0,
        failed: 0,
        successRate: 0
      }
    };

    this.notifyListeners('stressTestStarted', { concurrentConnections, iterations });

    for (let iteration = 0; iteration < iterations; iteration++) {
      console.log(`🚀 Stress Test Iteration ${iteration + 1}/${iterations} - ${concurrentConnections} concurrent connections`);
      
      const promises = Array.from({ length: concurrentConnections }, (_, index) => 
        this.runSingleConnectionTest(`${iteration + 1}.${index + 1}`, concurrentConnections * iterations)
      );

      const iterationResults = await Promise.all(promises);
      results.results.push({
        iteration: iteration + 1,
        attempts: iterationResults,
        successful: iterationResults.filter(r => r.success).length,
        failed: iterationResults.filter(r => !r.success).length
      });

      // Update overall stats
      results.overallStats.totalAttempts += iterationResults.length;
      results.overallStats.successful += iterationResults.filter(r => r.success).length;
      results.overallStats.failed += iterationResults.filter(r => !r.success).length;

      // Wait between iterations
      if (iteration < iterations - 1) {
        await new Promise(resolve => setTimeout(resolve, 5000));
      }
    }

    results.endTime = Date.now();
    results.totalDuration = results.endTime - results.startTime;
    results.overallStats.successRate = (results.overallStats.successful / results.overallStats.totalAttempts) * 100;

    this.notifyListeners('stressTestCompleted', results);
    console.log('💥 WebSocket Stress Test Completed:', results);
    
    return results;
  }

  // Recovery test - force disconnections and measure reconnection success
  async runRecoveryTest(disconnectionScenarios = 5) {
    const results = {
      testType: 'recovery',
      startTime: Date.now(),
      scenarios: [],
      overallStats: {
        totalScenarios: disconnectionScenarios,
        successfulRecoveries: 0,
        failedRecoveries: 0,
        averageRecoveryTime: 0
      }
    };

    this.notifyListeners('recoveryTestStarted', { scenarios: disconnectionScenarios });

    for (let scenario = 0; scenario < disconnectionScenarios; scenario++) {
      console.log(`🔄 Recovery Test Scenario ${scenario + 1}/${disconnectionScenarios}`);
      
      const scenarioResult = await this.runRecoveryScenario(scenario + 1);
      results.scenarios.push(scenarioResult);
      
      if (scenarioResult.recoverySuccessful) {
        results.overallStats.successfulRecoveries++;
      } else {
        results.overallStats.failedRecoveries++;
      }
    }

    results.endTime = Date.now();
    results.totalDuration = results.endTime - results.startTime;
    
    const successfulRecoveries = results.scenarios.filter(s => s.recoverySuccessful);
    if (successfulRecoveries.length > 0) {
      results.overallStats.averageRecoveryTime = successfulRecoveries
        .reduce((sum, s) => sum + s.recoveryTime, 0) / successfulRecoveries.length;
    }

    this.notifyListeners('recoveryTestCompleted', results);
    console.log('🔄 WebSocket Recovery Test Completed:', results);
    
    return results;
  }

  // Single recovery scenario
  async runRecoveryScenario(scenarioNumber) {
    const startTime = Date.now();
    const scenario = {
      scenarioNumber,
      initialConnection: null,
      disconnectionTime: null,
      reconnectionAttempt: null,
      recoverySuccessful: false,
      recoveryTime: null,
      totalTime: null
    };

    try {
      // Step 1: Establish initial connection
      scenario.initialConnection = await this.runSingleConnectionTest(`recovery_${scenarioNumber}_initial`, 1);
      
      if (!scenario.initialConnection.success) {
        scenario.totalTime = Date.now() - startTime;
        return scenario;
      }

      // Step 2: Force disconnection (simulate by creating and immediately closing)
      await new Promise(resolve => setTimeout(resolve, 2000)); // Let connection stabilize
      scenario.disconnectionTime = Date.now();

      // Step 3: Attempt recovery
      const recoveryStartTime = Date.now();
      scenario.reconnectionAttempt = await this.runSingleConnectionTest(`recovery_${scenarioNumber}_recovery`, 1);
      
      if (scenario.reconnectionAttempt.success) {
        scenario.recoverySuccessful = true;
        scenario.recoveryTime = Date.now() - recoveryStartTime;
      }

      scenario.totalTime = Date.now() - startTime;
      return scenario;

    } catch (error) {
      scenario.error = error.message;
      scenario.totalTime = Date.now() - startTime;
      return scenario;
    }
  }

  // Browser compatibility test
  async runBrowserCompatibilityTest() {
    const browserInfo = this.getBrowserInfo();
    const results = {
      testType: 'browser_compatibility',
      browserInfo,
      webSocketSupport: typeof WebSocket !== 'undefined',
      protocolSupport: {
        ws: false,
        wss: false
      },
      features: {
        binaryType: false,
        extensions: false,
        protocol: false
      },
      connectionTest: null
    };

    // Test WebSocket feature support
    if (typeof WebSocket !== 'undefined') {
      try {
        const testSocket = new WebSocket('wss://echo.websocket.org');
        results.features.binaryType = typeof testSocket.binaryType !== 'undefined';
        results.features.extensions = typeof testSocket.extensions !== 'undefined';
        results.features.protocol = typeof testSocket.protocol !== 'undefined';
        testSocket.close();
      } catch (error) {
        // Feature detection failed
      }
    }

    // Test actual connection
    results.connectionTest = await this.runSingleConnectionTest('browser_compat', 1);

    this.notifyListeners('browserCompatibilityTestCompleted', results);
    return results;
  }

  // Network quality test
  async runNetworkQualityTest() {
    const results = {
      testType: 'network_quality',
      startTime: Date.now(),
      latency: null,
      bandwidth: null,
      stability: null,
      connectionTest: null
    };

    try {
      // Measure latency
      results.latency = await this.measureLatency();
      
      // Measure bandwidth
      results.bandwidth = await this.measureBandwidth();
      
      // Test connection stability
      results.connectionTest = await this.runSingleConnectionTest('network_quality', 1);
      
      results.endTime = Date.now();
      
    } catch (error) {
      results.error = error.message;
    }

    this.notifyListeners('networkQualityTestCompleted', results);
    return results;
  }

  // Utility methods
  getTestWebSocketURL(token) {
    const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    
    if (isLocalhost) {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsHost = window.location.hostname;
      const wsPort = process.env.REACT_APP_WS_PORT || '8080';
      return `${wsProtocol}//${wsHost}:${wsPort}/ws?token=${token}`;
    } else {
      const backendUrl = this.getBackendURL();
      const wsHost = backendUrl.replace(/^https?:\/\//, '');
      return `wss://${wsHost}/ws?token=${token}`;
    }
  }

  getBackendURL() {
    if (process.env.REACT_APP_BACKEND_URL) {
      return process.env.REACT_APP_BACKEND_URL;
    }
    
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      return `${window.location.protocol}//${window.location.hostname}:8080`;
    }
    
    if (window.location.hostname.includes('railway.app')) {
      const currentDomain = window.location.hostname;
      const backendDomain = currentDomain.replace('-frontend', '');
      return `${window.location.protocol}//${backendDomain}`;
    }
    
    return 'https://officestonks.com';
  }

  classifyError(error) {
    // Use the same classification logic as forensics
    return websocketForensics.classifyError(error, {});
  }

  categorizeCloseCode(code) {
    const categories = {
      1000: 'normal_closure',
      1001: 'going_away',
      1002: 'protocol_error',
      1003: 'unsupported_data',
      1006: 'abnormal_closure',
      1007: 'invalid_payload',
      1008: 'policy_violation',
      1009: 'message_too_big',
      1011: 'internal_error',
      1015: 'tls_handshake_failure'
    };
    return categories[code] || 'unknown_close_code';
  }

  getBrowserInfo() {
    return {
      userAgent: navigator.userAgent,
      browser: this.detectBrowser(),
      version: this.getBrowserVersion(),
      platform: navigator.platform,
      webSocketSupport: typeof WebSocket !== 'undefined'
    };
  }

  detectBrowser() {
    const userAgent = navigator.userAgent;
    if (userAgent.includes('Chrome')) return 'Chrome';
    if (userAgent.includes('Firefox')) return 'Firefox';
    if (userAgent.includes('Safari')) return 'Safari';
    if (userAgent.includes('Edge')) return 'Edge';
    return 'Unknown';
  }

  getBrowserVersion() {
    const userAgent = navigator.userAgent;
    const match = userAgent.match(/(Chrome|Firefox|Safari|Edge)\/(\d+)/);
    return match ? match[2] : 'Unknown';
  }

  async measureLatency() {
    const start = performance.now();
    try {
      await fetch('/health', { method: 'HEAD' });
      return performance.now() - start;
    } catch {
      return null;
    }
  }

  async measureBandwidth() {
    // Simple bandwidth estimation
    const start = performance.now();
    const testSize = 100000; // 100KB
    
    try {
      const response = await fetch(`/favicon.ico?t=${Date.now()}`);
      await response.blob();
      const duration = performance.now() - start;
      const bitsLoaded = testSize * 8;
      const speedBps = bitsLoaded / (duration / 1000);
      return speedBps / 1024; // Convert to Kbps
    } catch {
      return null;
    }
  }

  generateTestRecommendations(testResults) {
    const recommendations = [];
    
    if (testResults.successRate < 50) {
      recommendations.push({
        priority: 'critical',
        category: 'reliability',
        message: `WebSocket success rate is critically low at ${testResults.successRate.toFixed(1)}%`,
        action: 'Implement immediate fallback to SSE or HTTP polling'
      });
    } else if (testResults.successRate < 80) {
      recommendations.push({
        priority: 'high',
        category: 'reliability',
        message: `WebSocket success rate is below acceptable threshold at ${testResults.successRate.toFixed(1)}%`,
        action: 'Investigate primary failure patterns and implement targeted fixes'
      });
    }

    // Analyze error patterns
    const topErrorPattern = Object.entries(testResults.errorPatterns)
      .sort((a, b) => b[1] - a[1])[0];
    
    if (topErrorPattern && topErrorPattern[1] > testResults.totalAttempts * 0.3) {
      recommendations.push({
        priority: 'high',
        category: 'error_pattern',
        message: `Primary failure mode: ${topErrorPattern[0]} (${((topErrorPattern[1] / testResults.totalAttempts) * 100).toFixed(1)}% of failures)`,
        action: 'Focus debugging efforts on this specific error pattern'
      });
    }

    if (testResults.performance.averageConnectionTime > 5000) {
      recommendations.push({
        priority: 'medium',
        category: 'performance',
        message: `Average connection time is slow at ${testResults.performance.averageConnectionTime.toFixed(0)}ms`,
        action: 'Optimize connection handshake and reduce timeouts'
      });
    }

    return recommendations;
  }

  // Event system
  addListener(callback) {
    this.listeners.push(callback);
  }

  removeListener(callback) {
    this.listeners = this.listeners.filter(cb => cb !== callback);
  }

  notifyListeners(event, data) {
    this.listeners.forEach(callback => {
      try {
        callback({ event, data, timestamp: Date.now() });
      } catch (error) {
        console.error('Error in WebSocket tester listener:', error);
      }
    });
  }

  // Get all test results
  getAllTestResults() {
    return this.testResults;
  }

  // Clear test results
  clearResults() {
    this.testResults = [];
  }

  // Get current status
  getStatus() {
    return {
      isRunning: this.isRunning,
      totalTestsRun: this.testResults.length,
      lastTestResult: this.testResults[this.testResults.length - 1] || null
    };
  }
}

// Create singleton instance
const websocketTester = new WebSocketTester();

export default websocketTester;
export { WebSocketTester };