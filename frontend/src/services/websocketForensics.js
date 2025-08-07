// WebSocket Connection Forensics & Debugging System
// This system captures detailed data about every WebSocket connection attempt
// to identify exact failure patterns and root causes

class WebSocketForensics {
  constructor() {
    this.connectionAttempts = [];
    this.diagnosticData = {
      browserInfo: this.getBrowserFingerprint(),
      environmentInfo: this.getEnvironmentInfo(),
      networkInfo: this.getNetworkInfo(),
      railwayInfo: this.getRailwayInfo()
    };
    this.listeners = [];
    this.isRecording = true;
  }

  // Browser and environment fingerprinting
  getBrowserFingerprint() {
    return {
      userAgent: navigator.userAgent,
      browser: this.detectBrowser(),
      version: this.getBrowserVersion(),
      platform: navigator.platform,
      language: navigator.language,
      cookiesEnabled: navigator.cookieEnabled,
      webSocketSupport: typeof WebSocket !== 'undefined',
      connectionType: navigator.connection ? {
        type: navigator.connection.type,
        effectiveType: navigator.connection.effectiveType,
        downlink: navigator.connection.downlink,
        rtt: navigator.connection.rtt,
        saveData: navigator.connection.saveData
      } : null,
      hardwareConcurrency: navigator.hardwareConcurrency,
      maxTouchPoints: navigator.maxTouchPoints,
      timestamp: Date.now()
    };
  }

  getEnvironmentInfo() {
    return {
      url: window.location.href,
      protocol: window.location.protocol,
      hostname: window.location.hostname,
      port: window.location.port,
      pathname: window.location.pathname,
      search: window.location.search,
      isHTTPS: window.location.protocol === 'https:',
      isLocalhost: window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1',
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      },
      screen: {
        width: window.screen.width,
        height: window.screen.height,
        colorDepth: window.screen.colorDepth
      },
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      timestamp: Date.now()
    };
  }

  getNetworkInfo() {
    const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
    return {
      online: navigator.onLine,
      connection: connection ? {
        effectiveType: connection.effectiveType,
        type: connection.type,
        downlink: connection.downlink,
        downlinkMax: connection.downlinkMax,
        rtt: connection.rtt,
        saveData: connection.saveData
      } : null,
      estimatedBandwidth: this.estimateBandwidth(),
      timestamp: Date.now()
    };
  }

  getRailwayInfo() {
    // Detect Railway environment and proxy characteristics
    const isRailway = window.location.hostname.includes('railway.app') || 
                     window.location.hostname.includes('up.railway.app') ||
                     window.location.hostname.includes('officestonks.com');
    
    return {
      isRailway,
      hostname: window.location.hostname,
      subdomain: window.location.hostname.split('.')[0],
      isProduction: !window.location.hostname.includes('localhost'),
      detectedProxy: this.detectProxyType(),
      estimatedLocation: this.estimateServerLocation(),
      timestamp: Date.now()
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

  detectProxyType() {
    // Analyze headers and behavior patterns to detect proxy type
    const hostname = window.location.hostname;
    if (hostname.includes('railway.app')) {
      return 'Railway Proxy';
    }
    if (hostname.includes('herokuapp.com')) {
      return 'Heroku Router';
    }
    if (hostname.includes('vercel.app')) {
      return 'Vercel Edge Network';
    }
    return 'Unknown/Direct';
  }

  estimateServerLocation() {
    // Basic server location estimation based on hostname patterns
    const hostname = window.location.hostname;
    if (hostname.includes('us-east')) return 'US East';
    if (hostname.includes('us-west')) return 'US West';
    if (hostname.includes('eu-')) return 'Europe';
    if (hostname.includes('ap-')) return 'Asia Pacific';
    return 'Unknown';
  }

  estimateBandwidth() {
    // Simple bandwidth estimation using image loading
    return new Promise((resolve) => {
      const startTime = Date.now();
      const image = new Image();
      const testImageSize = 50000; // ~50KB test image
      
      image.onload = () => {
        const duration = (Date.now() - startTime) / 1000;
        const bitsLoaded = testImageSize * 8;
        const speedBps = bitsLoaded / duration;
        const speedKbps = speedBps / 1024;
        resolve(speedKbps);
      };
      
      image.onerror = () => resolve(null);
      
      // Use a small test image from the same domain
      image.src = '/favicon.ico?t=' + Date.now();
    });
  }

  // Connection attempt logging
  startConnectionAttempt(config = {}) {
    const attemptId = this.generateAttemptId();
    const attempt = {
      id: attemptId,
      startTime: Date.now(),
      config,
      environment: {
        ...this.diagnosticData.environmentInfo,
        timestamp: Date.now()
      },
      network: this.getNetworkInfo(),
      stages: [
        {
          stage: 'initiated',
          timestamp: Date.now(),
          data: config
        }
      ],
      errors: [],
      performance: {
        dnsLookup: null,
        tcpConnect: null,
        tlsHandshake: null,
        wsUpgrade: null,
        totalTime: null
      },
      outcome: 'pending'
    };

    this.connectionAttempts.push(attempt);
    this.notifyListeners('attemptStarted', attempt);
    
    console.log(`🔍 WebSocket Forensics: Started tracking attempt ${attemptId}`, attempt);
    return attemptId;
  }

  logConnectionStage(attemptId, stage, data = {}) {
    const attempt = this.connectionAttempts.find(a => a.id === attemptId);
    if (!attempt) {
      console.warn(`WebSocket Forensics: Attempt ${attemptId} not found for stage ${stage}`);
      return;
    }

    const stageData = {
      stage,
      timestamp: Date.now(),
      data,
      timeSinceStart: Date.now() - attempt.startTime
    };

    attempt.stages.push(stageData);
    console.log(`🔍 WebSocket Forensics [${attemptId}]: ${stage}`, stageData);
    
    this.notifyListeners('stageLogged', { attemptId, stage: stageData });
  }

  logConnectionError(attemptId, error, context = {}) {
    const attempt = this.connectionAttempts.find(a => a.id === attemptId);
    if (!attempt) {
      console.warn(`WebSocket Forensics: Attempt ${attemptId} not found for error logging`);
      return;
    }

    const errorData = {
      timestamp: Date.now(),
      timeSinceStart: Date.now() - attempt.startTime,
      error: {
        message: error.message || error,
        type: error.name || typeof error,
        stack: error.stack,
        code: error.code
      },
      context,
      classification: this.classifyError(error, context)
    };

    attempt.errors.push(errorData);
    console.error(`🚨 WebSocket Forensics [${attemptId}]: Error`, errorData);
    
    this.notifyListeners('errorLogged', { attemptId, error: errorData });
  }

  completeConnectionAttempt(attemptId, outcome, finalData = {}) {
    const attempt = this.connectionAttempts.find(a => a.id === attemptId);
    if (!attempt) {
      console.warn(`WebSocket Forensics: Attempt ${attemptId} not found for completion`);
      return;
    }

    attempt.outcome = outcome;
    attempt.endTime = Date.now();
    attempt.totalDuration = attempt.endTime - attempt.startTime;
    attempt.finalData = finalData;

    // Calculate performance metrics
    this.calculatePerformanceMetrics(attempt);
    
    console.log(`✅ WebSocket Forensics [${attemptId}]: Completed with outcome: ${outcome}`, {
      duration: attempt.totalDuration,
      outcome,
      errorCount: attempt.errors.length
    });
    
    this.notifyListeners('attemptCompleted', attempt);
    
    // Analyze patterns if we have enough data
    if (this.connectionAttempts.length % 10 === 0) {
      this.analyzePatterns();
    }
  }

  classifyError(error, context) {
    const errorMessage = error.message || error.toString();
    const errorType = error.name || typeof error;

    // Railway-specific error patterns
    if (errorMessage.includes('hijacker')) {
      return {
        category: 'railway_proxy_limitation',
        severity: 'high',
        description: 'Railway proxy does not support WebSocket hijacking',
        actionable: true,
        suggestion: 'Use Railway-compatible WebSocket upgrade method'
      };
    }

    if (errorMessage.includes('timeout') || errorMessage.includes('TIMEOUT')) {
      return {
        category: 'connection_timeout',
        severity: 'medium',
        description: 'Connection timed out during establishment',
        actionable: true,
        suggestion: 'Increase timeout or check network conditions'
      };
    }

    if (errorType === 'SecurityError') {
      return {
        category: 'security_restriction',
        severity: 'high',
        description: 'Browser security policy blocks WebSocket connection',
        actionable: true,
        suggestion: 'Check CORS settings and SSL configuration'
      };
    }

    if (errorMessage.includes('403') || errorMessage.includes('401')) {
      return {
        category: 'authentication_failure',
        severity: 'medium',
        description: 'Authentication or authorization failed',
        actionable: true,
        suggestion: 'Verify JWT token validity and permissions'
      };
    }

    if (errorMessage.includes('1006')) {
      return {
        category: 'abnormal_closure',
        severity: 'high',
        description: 'WebSocket closed abnormally, likely proxy/network issue',
        actionable: true,
        suggestion: 'Implement robust reconnection with exponential backoff'
      };
    }

    if (errorMessage.includes('network') || errorMessage.includes('NETWORK')) {
      return {
        category: 'network_failure',
        severity: 'medium',
        description: 'Network connectivity issue',
        actionable: false,
        suggestion: 'Enable automatic fallback to HTTP polling'
      };
    }

    return {
      category: 'unknown_error',
      severity: 'low',
      description: `Unclassified error: ${errorMessage}`,
      actionable: false,
      suggestion: 'Log for analysis and pattern detection'
    };
  }

  calculatePerformanceMetrics(attempt) {
    const stages = attempt.stages;
    let previousTime = attempt.startTime;

    stages.forEach((stage, index) => {
      const stageTime = stage.timestamp - previousTime;
      
      switch (stage.stage) {
        case 'dns_lookup_complete':
          attempt.performance.dnsLookup = stageTime;
          break;
        case 'tcp_connect_complete':
          attempt.performance.tcpConnect = stageTime;
          break;
        case 'tls_handshake_complete':
          attempt.performance.tlsHandshake = stageTime;
          break;
        case 'websocket_upgrade_complete':
          attempt.performance.wsUpgrade = stageTime;
          break;
      }
      
      previousTime = stage.timestamp;
    });

    attempt.performance.totalTime = attempt.totalDuration;
  }

  generateAttemptId() {
    return `ws_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // Pattern analysis and insights
  analyzePatterns() {
    const recent = this.connectionAttempts.slice(-50); // Analyze last 50 attempts
    const successful = recent.filter(a => a.outcome === 'success');
    const failed = recent.filter(a => a.outcome === 'failed');
    
    const analysis = {
      totalAttempts: recent.length,
      successRate: (successful.length / recent.length) * 100,
      failureRate: (failed.length / recent.length) * 100,
      averageConnectionTime: this.calculateAverageConnectionTime(successful),
      commonErrorPatterns: this.identifyCommonErrors(failed),
      railwaySpecificIssues: this.analyzeRailwayIssues(failed),
      browserCompatibility: this.analyzeBrowserPatterns(recent),
      networkCorrelations: this.analyzeNetworkCorrelations(recent),
      timeBasedPatterns: this.analyzeTimePatterns(recent),
      recommendations: this.generateRecommendations(recent)
    };

    console.log('📊 WebSocket Forensics Analysis:', analysis);
    this.notifyListeners('patternAnalysis', analysis);
    
    return analysis;
  }

  calculateAverageConnectionTime(successfulAttempts) {
    if (successfulAttempts.length === 0) return null;
    const totalTime = successfulAttempts.reduce((sum, attempt) => sum + attempt.totalDuration, 0);
    return Math.round(totalTime / successfulAttempts.length);
  }

  identifyCommonErrors(failedAttempts) {
    const errorCounts = {};
    
    failedAttempts.forEach(attempt => {
      attempt.errors.forEach(error => {
        const category = error.classification.category;
        errorCounts[category] = (errorCounts[category] || 0) + 1;
      });
    });
    
    return Object.entries(errorCounts)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 5)
      .map(([category, count]) => ({ category, count, percentage: (count / failedAttempts.length) * 100 }));
  }

  analyzeRailwayIssues(failedAttempts) {
    const railwayAttempts = failedAttempts.filter(attempt => 
      attempt.environment.hostname.includes('railway.app') || 
      attempt.environment.hostname.includes('officestonks.com')
    );
    
    if (railwayAttempts.length === 0) return null;
    
    const hijackerErrors = railwayAttempts.filter(attempt => 
      attempt.errors.some(error => error.classification.category === 'railway_proxy_limitation')
    );
    
    return {
      totalRailwayAttempts: railwayAttempts.length,
      hijackerErrorCount: hijackerErrors.length,
      hijackerErrorRate: (hijackerErrors.length / railwayAttempts.length) * 100,
      isRailwaySpecificIssue: hijackerErrors.length > railwayAttempts.length * 0.5
    };
  }

  analyzeBrowserPatterns(attempts) {
    const browserStats = {};
    
    attempts.forEach(attempt => {
      const browser = attempt.environment.browserInfo?.browser || 'Unknown';
      if (!browserStats[browser]) {
        browserStats[browser] = { total: 0, successful: 0 };
      }
      browserStats[browser].total++;
      if (attempt.outcome === 'success') {
        browserStats[browser].successful++;
      }
    });
    
    return Object.entries(browserStats).map(([browser, stats]) => ({
      browser,
      totalAttempts: stats.total,
      successRate: (stats.successful / stats.total) * 100
    }));
  }

  analyzeNetworkCorrelations(attempts) {
    const networkTypes = {};
    
    attempts.forEach(attempt => {
      const networkType = attempt.network?.connection?.effectiveType || 'Unknown';
      if (!networkTypes[networkType]) {
        networkTypes[networkType] = { total: 0, successful: 0 };
      }
      networkTypes[networkType].total++;
      if (attempt.outcome === 'success') {
        networkTypes[networkType].successful++;
      }
    });
    
    return Object.entries(networkTypes).map(([type, stats]) => ({
      networkType: type,
      totalAttempts: stats.total,
      successRate: (stats.successful / stats.total) * 100
    }));
  }

  analyzeTimePatterns(attempts) {
    const hourlyStats = {};
    
    attempts.forEach(attempt => {
      const hour = new Date(attempt.startTime).getHours();
      if (!hourlyStats[hour]) {
        hourlyStats[hour] = { total: 0, successful: 0 };
      }
      hourlyStats[hour].total++;
      if (attempt.outcome === 'success') {
        hourlyStats[hour].successful++;
      }
    });
    
    return Object.entries(hourlyStats).map(([hour, stats]) => ({
      hour: parseInt(hour),
      totalAttempts: stats.total,
      successRate: (stats.successful / stats.total) * 100
    }));
  }

  generateRecommendations(attempts) {
    const analysis = this.getBasicStats(attempts);
    const recommendations = [];
    
    if (analysis.successRate < 50) {
      recommendations.push({
        priority: 'critical',
        category: 'reliability',
        recommendation: 'WebSocket success rate is critically low. Implement immediate fallback to SSE/polling.',
        evidence: `Success rate: ${analysis.successRate.toFixed(1)}%`
      });
    }
    
    const railwayIssues = this.analyzeRailwayIssues(attempts.filter(a => a.outcome === 'failed'));
    if (railwayIssues && railwayIssues.hijackerErrorRate > 60) {
      recommendations.push({
        priority: 'high',
        category: 'railway_compatibility',
        recommendation: 'Implement Railway-specific WebSocket upgrade handler to bypass hijacker limitations.',
        evidence: `${railwayIssues.hijackerErrorRate.toFixed(1)}% of Railway connections fail due to hijacker errors`
      });
    }
    
    if (analysis.averageConnectionTime > 10000) {
      recommendations.push({
        priority: 'medium',
        category: 'performance',
        recommendation: 'Connection establishment is slow. Optimize handshake process and reduce timeouts.',
        evidence: `Average connection time: ${analysis.averageConnectionTime}ms`
      });
    }
    
    return recommendations;
  }

  getBasicStats(attempts) {
    const successful = attempts.filter(a => a.outcome === 'success');
    return {
      successRate: (successful.length / attempts.length) * 100,
      averageConnectionTime: this.calculateAverageConnectionTime(successful)
    };
  }

  // Event system for real-time monitoring
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
        console.error('Error in WebSocket forensics listener:', error);
      }
    });
  }

  // Data export and reporting
  exportDiagnosticData() {
    return {
      diagnosticData: this.diagnosticData,
      connectionAttempts: this.connectionAttempts,
      analysis: this.analyzePatterns(),
      exportTimestamp: Date.now()
    };
  }

  generateReport() {
    const analysis = this.analyzePatterns();
    const report = {
      summary: {
        totalAttempts: this.connectionAttempts.length,
        successRate: analysis.successRate,
        criticalIssues: analysis.recommendations.filter(r => r.priority === 'critical').length,
        railwayCompatibility: analysis.railwaySpecificIssues
      },
      details: analysis,
      rawData: this.connectionAttempts.slice(-20) // Last 20 attempts
    };
    
    console.log('📋 WebSocket Forensics Report:', report);
    return report;
  }

  // Reset and cleanup
  clearData() {
    this.connectionAttempts = [];
    console.log('🧹 WebSocket Forensics: Cleared all data');
  }

  startRecording() {
    this.isRecording = true;
    console.log('🎬 WebSocket Forensics: Started recording');
  }

  stopRecording() {
    this.isRecording = false;
    console.log('⏹️ WebSocket Forensics: Stopped recording');
  }
}

// Create singleton instance
const websocketForensics = new WebSocketForensics();

// Export for use in other modules
export default websocketForensics;
export { WebSocketForensics };