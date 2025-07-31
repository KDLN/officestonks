import React, { useEffect, useState } from 'react';
import './CrisisAlert.css';

const CrisisAlert = ({ alert, onClose }) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (alert) {
      setIsVisible(true);
      // Auto-close after duration
      const timer = setTimeout(() => {
        handleClose();
      }, alert.duration || 8000); // Crisis alerts stay longer (8 seconds)

      return () => clearTimeout(timer);
    }
  }, [alert]);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(() => {
      if (onClose) onClose();
    }, 300); // Wait for fade-out animation
  };

  if (!alert) return null;

  const getAlertConfig = (type) => {
    const configs = {
      crisis: {
        icon: '🚨',
        className: 'crisis-alert-crisis',
        title: 'CRISIS ALERT',
        color: '#ff4444'
      },
      bankruptcy: {
        icon: '💀',
        className: 'crisis-alert-bankruptcy',
        title: 'BANKRUPTCY',
        color: '#8b0000'
      },
      recovery: {
        icon: '🚀',
        className: 'crisis-alert-recovery',
        title: 'RECOVERY',
        color: '#00aa44'
      },
      sector: {
        icon: '📉',
        className: 'crisis-alert-sector',
        title: 'SECTOR ALERT',
        color: '#ff8800'
      }
    };
    return configs[type] || configs.crisis;
  };

  const config = getAlertConfig(alert.type);

  return (
    <div className={`crisis-alert ${config.className} ${isVisible ? 'visible' : ''}`}>
      <div className="crisis-alert-header">
        <div className="crisis-alert-icon">{config.icon}</div>
        <div className="crisis-alert-title">{config.title}</div>
        <button className="crisis-alert-close" onClick={handleClose}>×</button>
      </div>
      
      <div className="crisis-alert-content">
        <div className="crisis-alert-stock">
          {alert.stockSymbol && (
            <span className="stock-symbol">{alert.stockSymbol}</span>
          )}
          {alert.stockName && (
            <span className="stock-name">{alert.stockName}</span>
          )}
        </div>
        
        <div className="crisis-alert-message">
          {alert.message}
        </div>
        
        {alert.details && (
          <div className="crisis-alert-details">
            {alert.details}
          </div>
        )}
        
        {alert.price && (
          <div className="crisis-alert-price">
            Current Price: ${alert.price.toFixed(2)}
          </div>
        )}
        
        {alert.sector && (
          <div className="crisis-alert-sector">
            Sector: {alert.sector}
          </div>
        )}
      </div>
      
      <div className="crisis-alert-footer">
        <div className="crisis-alert-timestamp">
          {new Date(alert.timestamp).toLocaleTimeString()}
        </div>
        {alert.impact && (
          <div className={`crisis-alert-impact impact-${alert.impact > 0 ? 'positive' : 'negative'}`}>
            Impact: {alert.impact > 0 ? '+' : ''}{alert.impact}%
          </div>
        )}
      </div>
    </div>
  );
};

export default CrisisAlert;