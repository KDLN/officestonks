import React from 'react';
import { Link } from 'react-router-dom';
import './PortfolioCrisisAlert.css';

const PortfolioCrisisAlert = ({ portfolioItems }) => {
  if (!portfolioItems || portfolioItems.length === 0) {
    return null;
  }

  // Find stocks in crisis (at or near $0.01)
  const crisisStocks = portfolioItems.filter(item => 
    item.stock && item.stock.current_price <= 0.02
  );

  // Find stocks approaching crisis (under $1)
  const riskStocks = portfolioItems.filter(item => 
    item.stock && item.stock.current_price > 0.02 && item.stock.current_price <= 1.0
  );

  if (crisisStocks.length === 0 && riskStocks.length === 0) {
    return null;
  }

  const calculateHoldingValue = (item) => {
    return item.quantity * item.stock.current_price;
  };

  const calculatePotentialLoss = (item) => {
    const currentValue = calculateHoldingValue(item);
    return currentValue; // Full loss potential in bankruptcy
  };

  const totalCrisisValue = crisisStocks.reduce((sum, item) => sum + calculateHoldingValue(item), 0);
  const totalRiskValue = riskStocks.reduce((sum, item) => sum + calculateHoldingValue(item), 0);

  return (
    <div className="portfolio-crisis-alert">
      {crisisStocks.length > 0 && (
        <div className="crisis-section crisis-critical">
          <div className="crisis-header">
            <span className="crisis-icon">🚨</span>
            <h3>Critical Alert: Stocks in Crisis Mode</h3>
            <span className="crisis-count">{crisisStocks.length} stock{crisisStocks.length > 1 ? 's' : ''}</span>
          </div>
          
          <div className="crisis-description">
            <p>These stocks are at $0.01 and face potential bankruptcy (5% chance every 2 seconds) or recovery (3% chance).</p>
            <div className="crisis-stats">
              <span className="stat-item">
                <strong>Total Value at Risk:</strong> ${totalCrisisValue.toFixed(2)}
              </span>
              <span className="stat-item bankruptcy-risk">
                <strong>Bankruptcy Risk:</strong> Total loss possible
              </span>
              <span className="stat-item recovery-potential">
                <strong>Recovery Potential:</strong> 10-100x gains
              </span>
            </div>
          </div>

          <div className="crisis-stocks-list">
            {crisisStocks.map((item) => (
              <div key={item.stock.id} className="crisis-stock-item">
                <div className="stock-info">
                  <Link to={`/stocks/${item.stock.id}`} className="stock-symbol">
                    {item.stock.symbol}
                  </Link>
                  <span className="stock-name">{item.stock.name}</span>
                </div>
                <div className="stock-holding">
                  <span className="quantity">{item.quantity} shares</span>
                  <span className="price">${item.stock.current_price.toFixed(2)}</span>
                  <span className="value">${calculateHoldingValue(item).toFixed(2)}</span>
                </div>
                <div className="crisis-actions">
                  <span className="risk-badge bankruptcy">💀 Bankruptcy Risk</span>
                  <span className="risk-badge recovery">🚀 Recovery Potential</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {riskStocks.length > 0 && (
        <div className="crisis-section crisis-warning">
          <div className="crisis-header">
            <span className="crisis-icon">⚠️</span>
            <h3>Warning: Penny Stocks (High Risk)</h3>
            <span className="crisis-count">{riskStocks.length} stock{riskStocks.length > 1 ? 's' : ''}</span>
          </div>
          
          <div className="crisis-description">
            <p>These penny stocks (under $1) have high volatility and may enter crisis mode.</p>
            <div className="crisis-stats">
              <span className="stat-item">
                <strong>Total Value:</strong> ${totalRiskValue.toFixed(2)}
              </span>
              <span className="stat-item volatility">
                <strong>High Volatility:</strong> 10% price swings
              </span>
            </div>
          </div>

          <div className="crisis-stocks-list">
            {riskStocks.map((item) => (
              <div key={item.stock.id} className="crisis-stock-item warning">
                <div className="stock-info">
                  <Link to={`/stocks/${item.stock.id}`} className="stock-symbol">
                    {item.stock.symbol}
                  </Link>
                  <span className="stock-name">{item.stock.name}</span>
                </div>
                <div className="stock-holding">
                  <span className="quantity">{item.quantity} shares</span>
                  <span className="price">${item.stock.current_price.toFixed(2)}</span>
                  <span className="value">${calculateHoldingValue(item).toFixed(2)}</span>
                </div>
                <div className="crisis-actions">
                  <span className="risk-badge warning">⚠️ High Risk</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="crisis-footer">
        <div className="crisis-tips">
          <h4>💡 Crisis Trading Tips:</h4>
          <ul>
            <li><strong>Hold for Recovery:</strong> 3% chance every 2 seconds for 10-100x gains</li>
            <li><strong>Sell to Cut Losses:</strong> Avoid total loss from bankruptcy</li>
            <li><strong>Buy During Crisis:</strong> High risk, high reward opportunity</li>
            <li><strong>Diversify:</strong> Don't put all funds in penny stocks</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default PortfolioCrisisAlert;