import React, { useState, useEffect, useCallback } from 'react';
import CrisisAlert from './CrisisAlert';

const CrisisAlertManager = ({ socket }) => {
  const [alerts, setAlerts] = useState([]);

  // Generate unique ID for alerts
  const generateAlertId = () => Date.now() + Math.random();

  // Process stock price updates to detect crisis events
  const processStockUpdate = useCallback((data) => {
    if (!data || !data.price) return;

    const price = parseFloat(data.price);
    
    // Detect crisis - stock hits $0.01
    if (price <= 0.01) {
      const alert = {
        id: generateAlertId(),
        type: 'crisis',
        stockSymbol: data.symbol,
        stockName: data.name || `Stock ${data.symbol}`,
        message: `${data.symbol} has crashed to penny stock levels!`,
        details: 'This stock is now in crisis territory and may face bankruptcy or recovery.',
        price: price,
        sector: data.sector,
        impact: -95, // Rough impact percentage
        timestamp: new Date().toISOString(),
        duration: 10000 // 10 seconds for crisis alerts
      };
      
      addAlert(alert);
    }
    // Detect major price drops (>20% in one update)
    else if (data.previousPrice && data.previousPrice > 0) {
      const priceDrop = ((data.previousPrice - price) / data.previousPrice) * 100;
      
      if (priceDrop > 20) {
        const alert = {
          id: generateAlertId(),
          type: 'sector',
          stockSymbol: data.symbol,
          stockName: data.name || `Stock ${data.symbol}`,
          message: `${data.symbol} drops ${priceDrop.toFixed(1)}% in market turmoil!`,
          details: 'Significant price movement detected. Monitor for potential sector impact.',
          price: price,
          sector: data.sector,
          impact: -priceDrop,
          timestamp: new Date().toISOString(),
          duration: 6000
        };
        
        addAlert(alert);
      }
    }
    // Detect major price jumps (recovery-like behavior)
    else if (data.previousPrice && data.previousPrice > 0) {
      const priceJump = ((price - data.previousPrice) / data.previousPrice) * 100;
      
      if (priceJump > 100) { // 100%+ jump
        const alert = {
          id: generateAlertId(),
          type: 'recovery',
          stockSymbol: data.symbol,
          stockName: data.name || `Stock ${data.symbol}`,
          message: `${data.symbol} rockets ${priceJump.toFixed(0)}% higher!`,
          details: 'Dramatic recovery detected. This could be a turnaround story.',
          price: price,
          sector: data.sector,
          impact: priceJump,
          timestamp: new Date().toISOString(),
          duration: 8000
        };
        
        addAlert(alert);
      }
    }
  }, []);

  // Process news items to create alerts
  const processNewsUpdate = useCallback((newsItem) => {
    if (!newsItem) return;

    let alertType = 'crisis';
    let icon = '📰';
    let duration = 8000;

    // Map news types to alert types
    switch (newsItem.type) {
      case 'bankruptcy':
        alertType = 'bankruptcy';
        duration = 12000; // Bankruptcy news stays longer
        break;
      case 'recovery':
        alertType = 'recovery';
        duration = 10000;
        break;
      case 'sector':
        alertType = 'sector';
        duration = 6000;
        break;
      case 'crisis':
      default:
        alertType = 'crisis';
        break;
    }

    const alert = {
      id: generateAlertId(),
      type: alertType,
      stockSymbol: newsItem.stock_symbol,
      stockName: newsItem.stock_name,
      message: newsItem.title,
      details: newsItem.content ? newsItem.content.substring(0, 150) + '...' : null,
      sector: newsItem.sector_name,
      impact: newsItem.impact_score,
      timestamp: newsItem.created_at || new Date().toISOString(),
      duration
    };

    addAlert(alert);
  }, []);

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
        // Direct crisis event from admin testing or automated system
        if (data.type === 'crisis_event') {
          const alert = {
            id: generateAlertId(),
            type: data.event_type || 'crisis',
            stockSymbol: data.stock_symbol,
            stockName: data.stock_name,
            message: data.message || 'Crisis event detected',
            details: data.details,
            price: data.price,
            sector: data.sector,
            impact: data.impact,
            timestamp: data.timestamp || new Date().toISOString(),
            duration: 10000
          };
          
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
  }, [socket, processStockUpdate, processNewsUpdate]);

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