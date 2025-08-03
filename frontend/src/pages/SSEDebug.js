import React, { useState, useEffect, useRef } from 'react';
import { initSSE, closeSSE, addSSEListener, removeSSEListener, getSSEConnectionState, forceSSEReconnect } from '../services/sse';

const SSEDebug = () => {
  const [connectionState, setConnectionState] = useState('disconnected');
  const [messages, setMessages] = useState([]);
  const [stockUpdates, setStockUpdates] = useState([]);
  const [stats, setStats] = useState({
    totalMessages: 0,
    stockUpdates: 0,
    heartbeats: 0,
    connectionMessages: 0,
    errors: 0
  });
  const messagesEndRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  useEffect(() => {
    console.log('🧪 SSE Debug Page: Starting SSE connection...');
    
    // Connection state listener
    const handleConnectionState = (data) => {
      console.log('🔗 SSE Debug: Connection state changed:', data);
      setConnectionState(data.state);
      addMessage('CONNECTION', `State: ${data.state}`, data);
    };

    // Generic message listener
    const handleMessage = (data) => {
      console.log('📝 SSE Debug: Generic message received:', data);
      addMessage('MESSAGE', `Type: ${data.type}`, data);
      updateStats('connectionMessages');
    };

    // Stock update listener
    const handleStockUpdate = (data) => {
      console.log('📈 SSE Debug: Stock update received:', data);
      addMessage('STOCK_UPDATE', `${data.symbol}: $${data.price}`, data);
      setStockUpdates(prev => {
        const newUpdates = [data, ...prev.slice(0, 19)]; // Keep last 20
        return newUpdates;
      });
      updateStats('stockUpdates');
    };

    // Connection status listener
    const handleConnection = (data) => {
      console.log('🔌 SSE Debug: Connection message received:', data);
      addMessage('CONNECTION', `Status: ${data.status}`, data);
      updateStats('connectionMessages');
    };

    // Error listener (if we add one)
    const handleError = (error) => {
      console.error('❌ SSE Debug: Error received:', error);
      addMessage('ERROR', `Error: ${error.message || 'Unknown error'}`, error);
      updateStats('errors');
    };

    // Add listeners
    addSSEListener('connectionState', handleConnectionState);
    addSSEListener('message', handleMessage);
    addSSEListener('stockUpdate', handleStockUpdate);
    addSSEListener('connection', handleConnection);
    addSSEListener('error', handleError);

    // Initialize SSE
    initSSE();

    // Cleanup
    return () => {
      removeSSEListener('connectionState', handleConnectionState);
      removeSSEListener('message', handleMessage);
      removeSSEListener('stockUpdate', handleStockUpdate);
      removeSSEListener('connection', handleConnection);
      removeSSEListener('error', handleError);
      closeSSE();
    };
  }, []);

  const addMessage = (type, summary, data) => {
    const timestamp = new Date().toLocaleTimeString();
    const message = {
      id: Date.now() + Math.random(),
      timestamp,
      type,
      summary,
      data: JSON.stringify(data, null, 2)
    };
    
    setMessages(prev => [message, ...prev.slice(0, 99)]); // Keep last 100 messages
    updateStats('totalMessages');
    
    if (type === 'STOCK_UPDATE') updateStats('stockUpdates');
    else if (data?.type === 'heartbeat') updateStats('heartbeats');
  };

  const updateStats = (statType) => {
    setStats(prev => ({
      ...prev,
      [statType]: prev[statType] + 1
    }));
  };

  const handleReconnect = () => {
    console.log('🔄 SSE Debug: Forcing reconnection...');
    forceSSEReconnect();
  };

  const testRawEventSource = () => {
    console.log('🧪 Testing raw EventSource directly...');
    const testES = new EventSource('https://beta.officestonks.com/api/sse/test');
    
    testES.onopen = () => {
      console.log('✅ Raw EventSource: onopen fired!');
      addMessage('RAW_TEST', 'Raw EventSource onopen fired', { success: true });
    };
    
    testES.onmessage = (event) => {
      console.log('🎯 Raw EventSource: onmessage fired!', event);
      addMessage('RAW_TEST', 'Raw EventSource message received', { data: event.data });
      testES.close();
    };
    
    testES.onerror = (error) => {
      console.log('❌ Raw EventSource: onerror fired!', error);
      addMessage('RAW_TEST', 'Raw EventSource error', error);
    };
    
    // Timeout test
    setTimeout(() => {
      if (testES.readyState === EventSource.CONNECTING) {
        console.log('⏰ Raw EventSource: Still connecting after 10s, closing...');
        addMessage('RAW_TEST', 'Raw EventSource timeout', { readyState: testES.readyState });
        testES.close();
      }
    }, 10000);
  };

  const testFetchPolling = () => {
    console.log('🔄 Testing fetch polling as alternative...');
    let pollCount = 0;
    
    const poll = async () => {
      try {
        pollCount++;
        const response = await fetch('https://beta.officestonks.com/api/stock-updates/poll');
        const data = await response.json();
        
        console.log(`📊 Fetch poll #${pollCount}:`, data);
        addMessage('FETCH_POLL', `Poll #${pollCount} - ${data.updates?.length || 0} updates`, data);
        
        if (pollCount < 5) {
          setTimeout(poll, 2000); // Poll every 2 seconds
        }
      } catch (error) {
        console.error('❌ Fetch poll error:', error);
        addMessage('FETCH_POLL', `Poll #${pollCount} failed`, error);
      }
    };
    
    poll();
  };

  const testMainSSEEndpoint = () => {
    console.log('🧪 Testing main SSE endpoint directly...');
    const mainES = new EventSource('https://beta.officestonks.com/api/sse/stock-updates');
    
    mainES.onopen = () => {
      console.log('✅ Main SSE: onopen fired!');
      addMessage('MAIN_SSE_TEST', 'Main SSE onopen fired', { success: true });
    };
    
    mainES.onmessage = (event) => {
      console.log('🎯 Main SSE: onmessage fired!', event);
      addMessage('MAIN_SSE_TEST', 'Main SSE message received', { data: event.data });
      if (event.data.includes('stock_update')) {
        mainES.close();
      }
    };
    
    mainES.onerror = (error) => {
      console.log('❌ Main SSE: onerror fired!', error);
      addMessage('MAIN_SSE_TEST', 'Main SSE error', error);
    };
    
    // Timeout test
    setTimeout(() => {
      if (mainES.readyState === EventSource.CONNECTING) {
        console.log('⏰ Main SSE: Still connecting after 15s, closing...');
        addMessage('MAIN_SSE_TEST', 'Main SSE timeout', { readyState: mainES.readyState });
        mainES.close();
      }
    }, 15000);
  };

  const handleClearMessages = () => {
    setMessages([]);
    setStockUpdates([]);
    setStats({
      totalMessages: 0,
      stockUpdates: 0,
      heartbeats: 0,
      connectionMessages: 0,
      errors: 0
    });
  };

  const getConnectionColor = () => {
    switch (connectionState) {
      case 'connected': return '#10b981'; // green
      case 'connecting': return '#f59e0b'; // yellow
      case 'failed': return '#ef4444'; // red
      default: return '#6b7280'; // gray
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto', fontFamily: 'monospace' }}>
      <h1 style={{ color: '#1f2937', marginBottom: '20px' }}>🧪 SSE Debug Console</h1>
      
      {/* Status Bar */}
      <div style={{ 
        display: 'flex', 
        gap: '20px', 
        marginBottom: '20px',
        padding: '15px',
        backgroundColor: '#f3f4f6',
        borderRadius: '8px',
        alignItems: 'center'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div style={{ 
            width: '12px', 
            height: '12px', 
            borderRadius: '50%', 
            backgroundColor: getConnectionColor()
          }}></div>
          <strong>Status: {connectionState}</strong>
        </div>
        
        <div>📊 Messages: {stats.totalMessages}</div>
        <div>📈 Stock Updates: {stats.stockUpdates}</div>
        <div>💓 Heartbeats: {stats.heartbeats}</div>
        <div>❌ Errors: {stats.errors}</div>
        
        <button 
          onClick={handleReconnect}
          style={{ 
            padding: '8px 16px', 
            backgroundColor: '#3b82f6', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          🔄 Reconnect
        </button>
        
        <button 
          onClick={testRawEventSource}
          style={{ 
            padding: '8px 16px', 
            backgroundColor: '#10b981', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          🧪 Test Raw SSE
        </button>
        
        <button 
          onClick={testFetchPolling}
          style={{ 
            padding: '8px 16px', 
            backgroundColor: '#f59e0b', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          📊 Test Polling
        </button>
        
        <button 
          onClick={testMainSSEEndpoint}
          style={{ 
            padding: '8px 16px', 
            backgroundColor: '#8b5cf6', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          🎯 Test Main SSE
        </button>
        
        <button 
          onClick={handleClearMessages}
          style={{ 
            padding: '8px 16px', 
            backgroundColor: '#6b7280', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px',
            cursor: 'pointer'
          }}
        >
          🗑️ Clear
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
        {/* Real-time Stock Updates */}
        <div>
          <h3 style={{ color: '#1f2937', marginBottom: '10px' }}>📈 Live Stock Updates</h3>
          <div style={{ 
            height: '300px', 
            overflow: 'auto', 
            backgroundColor: '#000', 
            color: '#00ff00', 
            padding: '10px',
            borderRadius: '4px',
            fontSize: '12px'
          }}>
            {stockUpdates.length === 0 ? (
              <div style={{ color: '#666' }}>Waiting for stock updates...</div>
            ) : (
              stockUpdates.map((update, index) => (
                <div key={index} style={{ marginBottom: '5px' }}>
                  <span style={{ color: '#00ffff' }}>{new Date().toLocaleTimeString()}</span> {' '}
                  <span style={{ color: '#ffff00' }}>{update.symbol}</span> {' '}
                  <span style={{ color: '#00ff00' }}>${update.price}</span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* All Messages Log */}
        <div>
          <h3 style={{ color: '#1f2937', marginBottom: '10px' }}>📝 All Messages</h3>
          <div style={{ 
            height: '300px', 
            overflow: 'auto', 
            backgroundColor: '#1f2937', 
            color: '#e5e7eb', 
            padding: '10px',
            borderRadius: '4px',
            fontSize: '11px'
          }}>
            {messages.length === 0 ? (
              <div style={{ color: '#9ca3af' }}>No messages yet...</div>
            ) : (
              messages.map((msg) => (
                <div key={msg.id} style={{ marginBottom: '8px', borderBottom: '1px solid #374151', paddingBottom: '4px' }}>
                  <div style={{ 
                    display: 'flex', 
                    justifyContent: 'space-between',
                    color: getMessageTypeColor(msg.type)
                  }}>
                    <span>{msg.type}</span>
                    <span style={{ color: '#9ca3af', fontSize: '10px' }}>{msg.timestamp}</span>
                  </div>
                  <div style={{ color: '#d1d5db', fontSize: '10px' }}>{msg.summary}</div>
                </div>
              ))
            )}
            <div ref={messagesEndRef} />
          </div>
        </div>
      </div>

      {/* Test URLs */}
      <div style={{ 
        marginTop: '20px', 
        padding: '15px', 
        backgroundColor: '#fef3c7', 
        borderRadius: '8px' 
      }}>
        <h4>🧪 Test URLs:</h4>
        <div style={{ fontSize: '12px', fontFamily: 'monospace' }}>
          <div>SSE Endpoint: <a href="https://beta.officestonks.com/api/sse/stock-updates" target="_blank" rel="noopener noreferrer">https://beta.officestonks.com/api/sse/stock-updates</a></div>
          <div>SSE Test: <a href="https://beta.officestonks.com/api/sse/test" target="_blank" rel="noopener noreferrer">https://beta.officestonks.com/api/sse/test</a></div>
          <div>Health Check: <a href="https://beta.officestonks.com/health" target="_blank" rel="noopener noreferrer">https://beta.officestonks.com/health</a></div>
        </div>
      </div>

      {/* Debug Console Output */}
      <div style={{ 
        marginTop: '20px', 
        padding: '15px', 
        backgroundColor: '#f9fafb', 
        borderRadius: '8px' 
      }}>
        <h4>💻 Console Instructions:</h4>
        <p style={{ fontSize: '12px', color: '#6b7280' }}>
          Open browser DevTools (F12) → Console tab to see detailed SSE debugging logs with 🔧, 🎯, ✅, and ❌ emojis.
          All SSE events are logged there in real-time.
        </p>
      </div>
    </div>
  );
};

const getMessageTypeColor = (type) => {
  switch (type) {
    case 'STOCK_UPDATE': return '#10b981';
    case 'CONNECTION': return '#3b82f6';
    case 'ERROR': return '#ef4444';
    default: return '#f59e0b';
  }
};

export default SSEDebug;