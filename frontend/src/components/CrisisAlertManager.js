import React, { useState, useEffect, useCallback } from 'react';
import CrisisAlert from './CrisisAlert';

const CrisisAlertManager = ({ socket }) => {
  const [alerts, setAlerts] = useState([]);

  // Generate unique ID for alerts
  const generateAlertId = () => Date.now() + Math.random();

  // Helper function to create alert object
  const createAlert = useCallback((type, stockSymbol, stockName, message, details, options = {}) => ({
    id: generateAlertId(),
    type,
    stockSymbol,
    stockName,
    message,
    details,
    timestamp: new Date().toISOString(),
    ...options
  }), []);

  // Process stock price updates to detect crisis events
  const processStockUpdate = useCallback((data) => {
    if (!data || !data.price) return;

    const price = parseFloat(data.price);
    const stockName = data.name || `Stock ${data.symbol}`;
    
    // Detect crisis - stock hits $0.01
    if (price <= 0.01) {
      const alert = createAlert(
        'crisis',
        data.symbol,
        stockName,
        `${data.symbol} has crashed to penny stock levels!`,
        'This stock is now in crisis territory and may face bankruptcy or recovery.',
        { price, sector: data.sector, impact: -95, duration: 10000 }
      );
      addAlert(alert);
      return;
    }

    if (!data.previousPrice || data.previousPrice <= 0) return;

    const priceChange = ((price - data.previousPrice) / data.previousPrice) * 100;
    
    // Detect major price drops (>20%)
    if (priceChange < -20) {
      const alert = createAlert(
        'sector',
        data.symbol,
        stockName,
        `${data.symbol} drops ${Math.abs(priceChange).toFixed(1)}% in market turmoil!`,
        'Significant price movement detected. Monitor for potential sector impact.',
        { price, sector: data.sector, impact: priceChange, duration: 6000 }
      );
      addAlert(alert);
    }
    // Detect major price jumps (>100% recovery)
    else if (priceChange > 100) {
      const alert = createAlert(
        'recovery',
        data.symbol,
        stockName,
        `${data.symbol} rockets ${priceChange.toFixed(0)}% higher!`,
        'Dramatic recovery detected. This could be a turnaround story.',
        { price, sector: data.sector, impact: priceChange, duration: 8000 }
      );
      addAlert(alert);
    }
  }, [createAlert]);

  // Helper function to get news alert duration
  const getNewsAlertDuration = (type) => {
    const durations = {
      bankruptcy: 12000,
      recovery: 10000,
      sector: 6000,
      crisis: 8000
    };
    return durations[type] || durations.crisis;
  };

  // Process news items to create alerts
  const processNewsUpdate = useCallback((newsItem) => {
    if (!newsItem) return;

    const alertType = newsItem.type || 'crisis';
    const duration = getNewsAlertDuration(alertType);
    const details = newsItem.content ? newsItem.content.substring(0, 150) + '...' : null;

    const alert = createAlert(
      alertType,
      newsItem.stock_symbol,
      newsItem.stock_name,
      newsItem.title,
      details,
      {
        sector: newsItem.sector_name,
        impact: newsItem.impact_score,
        timestamp: newsItem.created_at || new Date().toISOString(),
        duration
      }
    );

    addAlert(alert);
  }, [createAlert]);

  // Add alert to the queue
  const addAlert = (alert) => {
    setAlerts(prev => {
      // Limit to 3 alerts at once to avoid overwhelming the user
      const newAlerts = [alert, ...prev.slice(0, 2)];
      return newAlerts;
    });
  };

  // Remove alert
  const removeAlert = (alertId) => {
    setAlerts(prev => prev.filter(alert => alert.id !== alertId));
  };

  // Listen to WebSocket events
  useEffect(() => {
    if (!socket) return;

    const handleStockUpdate = (data) => {
      try {
        processStockUpdate(data);
      } catch (error) {
        console.error('Error processing stock update for crisis detection:', error);
      }
    };

    const handleNewsUpdate = (data) => {
      try {
        if (data.type === 'news' && data.news) {
          processNewsUpdate(data.news);
        }
      } catch (error) {
        console.error('Error processing news update for crisis alerts:', error);
      }
    };

    const handleCrisisEvent = (data) => {
      try {
        if (data.type === 'crisis_event') {
          const alert = createAlert(
            data.event_type || 'crisis',
            data.stock_symbol,
            data.stock_name,
            data.message || 'Crisis event detected',
            data.details,
            {
              price: data.price,
              sector: data.sector,
              impact: data.impact,
              timestamp: data.timestamp || new Date().toISOString(),
              duration: 10000
            }
          );
          
          addAlert(alert);
        }
      } catch (error) {
        console.error('Error processing crisis event:', error);
      }
    };

    // Register event listeners
    socket.addEventListener('message', (event) => {
      try {
        const data = JSON.parse(event.data);
        
        if (data.type === 'stock_update') {
          handleStockUpdate(data);
        } else if (data.type === 'news') {
          handleNewsUpdate(data);
        } else if (data.type === 'crisis_event') {
          handleCrisisEvent(data);
        }
      } catch (error) {
        // Ignore JSON parse errors for non-JSON messages
      }
    });

    // Custom event listener for direct crisis alerts
    const handleCustomCrisisEvent = (event) => {
      addAlert(event.detail);
    };

    window.addEventListener('crisis-alert', handleCustomCrisisEvent);

    return () => {
      window.removeEventListener('crisis-alert', handleCustomCrisisEvent);
    };
  }, [socket, processStockUpdate, processNewsUpdate, createAlert]);

  return (
    <div className="crisis-alert-manager">
      {alerts.map((alert) => (
        <CrisisAlert
          key={alert.id}
          alert={alert}
          onClose={() => removeAlert(alert.id)}
        />
      ))}
    </div>
  );
};

// Utility function to manually trigger crisis alerts
export const triggerCrisisAlert = (alertData) => {
  const event = new CustomEvent('crisis-alert', {
    detail: {
      id: Date.now() + Math.random(),
      timestamp: new Date().toISOString(),
      duration: 8000,
      ...alertData
    }
  });
  window.dispatchEvent(event);
};

export default CrisisAlertManager;