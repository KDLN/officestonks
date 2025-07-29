import React, { useState } from 'react';
import './GameConfigSection.css';

const GameConfigSection = ({ 
  gameConfig, 
  onUpdate, 
  onReset, 
  onLoadBalanced, 
  loading 
}) => {
  const [editedConfig, setEditedConfig] = useState({});
  const [activeSection, setActiveSection] = useState('user');

  if (!gameConfig) {
    return <div className="game-config-loading">Loading game configuration...</div>;
  }

  const handleInputChange = (field, value) => {
    setEditedConfig(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSave = () => {
    // Only send changed values
    const changedConfig = {};
    Object.keys(editedConfig).forEach(key => {
      if (editedConfig[key] !== gameConfig[key]) {
        changedConfig[key] = editedConfig[key];
      }
    });

    if (Object.keys(changedConfig).length > 0) {
      onUpdate(changedConfig);
      setEditedConfig({});
    }
  };

  const getCurrentValue = (field) => {
    return editedConfig[field] !== undefined ? editedConfig[field] : gameConfig[field];
  };

  const hasChanges = () => {
    return Object.keys(editedConfig).some(key => editedConfig[key] !== gameConfig[key]);
  };


  return (
    <div className="game-config-section">
      <div className="game-config-header">
        <h2>Game Configuration</h2>
        <div className="config-actions">
          <button 
            className="config-btn balanced" 
            onClick={onLoadBalanced}
            disabled={loading}
          >
            Load Balanced
          </button>
          <button 
            className="config-btn reset" 
            onClick={onReset}
            disabled={loading}
          >
            Reset to Defaults
          </button>
          {hasChanges() && (
            <button 
              className="config-btn save" 
              onClick={handleSave}
              disabled={loading}
            >
              Save Changes
            </button>
          )}
        </div>
      </div>

      <div className="config-tabs">
        <button 
          className={`tab ${activeSection === 'user' ? 'active' : ''}`}
          onClick={() => setActiveSection('user')}
        >
          User Settings
        </button>
        <button 
          className={`tab ${activeSection === 'market' ? 'active' : ''}`}
          onClick={() => setActiveSection('market')}
        >
          Market Settings
        </button>
        <button 
          className={`tab ${activeSection === 'trading' ? 'active' : ''}`}
          onClick={() => setActiveSection('trading')}
        >
          Trading Limits
        </button>
        <button 
          className={`tab ${activeSection === 'impact' ? 'active' : ''}`}
          onClick={() => setActiveSection('impact')}
        >
          Price Impact
        </button>
      </div>

      <div className="config-content">
        {activeSection === 'user' && (
          <div className="config-group">
            <h3>User Settings</h3>
            <div className="config-field">
              <label>Starting Cash ($)</label>
              <input
                type="number"
                min="1000"
                max="1000000"
                step="1000"
                value={getCurrentValue('starting_cash')}
                onChange={(e) => handleInputChange('starting_cash', parseFloat(e.target.value))}
              />
              <span className="field-description">
                Amount of cash new users start with ($1,000 - $1,000,000)
              </span>
            </div>
          </div>
        )}

        {activeSection === 'market' && (
          <div className="config-group">
            <h3>Market Settings</h3>
            <div className="config-field">
              <label>Update Interval (seconds)</label>
              <input
                type="number"
                min="1"
                max="60"
                value={getCurrentValue('market_update_interval_seconds')}
                onChange={(e) => handleInputChange('market_update_interval_seconds', parseInt(e.target.value))}
              />
              <span className="field-description">
                How often stock prices update (1-60 seconds)
              </span>
            </div>
            <div className="config-field">
              <label>Market Volatility (%)</label>
              <input
                type="number"
                min="0.1"
                max="20"
                step="0.1"
                value={(getCurrentValue('market_volatility') * 100).toFixed(1)}
                onChange={(e) => handleInputChange('market_volatility', parseFloat(e.target.value) / 100)}
              />
              <span className="field-description">
                How much prices can fluctuate (0.1% - 20%)
              </span>
            </div>
            <div className="config-field">
              <label>Min Stock Price ($)</label>
              <input
                type="number"
                min="0.01"
                max="10"
                step="0.01"
                value={getCurrentValue('min_stock_price')}
                onChange={(e) => handleInputChange('min_stock_price', parseFloat(e.target.value))}
              />
              <span className="field-description">
                Minimum allowed stock price ($0.01 - $10)
              </span>
            </div>
            <div className="config-field">
              <label>Max Stock Price ($)</label>
              <input
                type="number"
                min="100"
                max="1000000"
                step="100"
                value={getCurrentValue('max_stock_price')}
                onChange={(e) => handleInputChange('max_stock_price', parseFloat(e.target.value))}
              />
              <span className="field-description">
                Maximum allowed stock price ($100 - $1,000,000)
              </span>
            </div>
          </div>
        )}

        {activeSection === 'trading' && (
          <div className="config-group">
            <h3>Trading Limits</h3>
            <div className="config-field">
              <label>Min Trade Quantity</label>
              <input
                type="number"
                min="1"
                max="100"
                value={getCurrentValue('min_trade_quantity')}
                onChange={(e) => handleInputChange('min_trade_quantity', parseInt(e.target.value))}
              />
              <span className="field-description">
                Minimum shares per trade (1-100)
              </span>
            </div>
            <div className="config-field">
              <label>Max Trade Quantity</label>
              <input
                type="number"
                min="1"
                max="100000"
                value={getCurrentValue('max_trade_quantity')}
                onChange={(e) => handleInputChange('max_trade_quantity', parseInt(e.target.value))}
              />
              <span className="field-description">
                Maximum shares per trade (1-100,000)
              </span>
            </div>
            <div className="config-field">
              <label>Trade Cooldown (seconds)</label>
              <input
                type="number"
                min="0"
                max="300"
                value={getCurrentValue('trade_cooldown_seconds')}
                onChange={(e) => handleInputChange('trade_cooldown_seconds', parseInt(e.target.value))}
              />
              <span className="field-description">
                Minimum time between trades (0-300 seconds)
              </span>
            </div>
            <div className="config-field">
              <label>Max Trades Per Hour</label>
              <input
                type="number"
                min="1"
                max="1000"
                value={getCurrentValue('max_trades_per_hour')}
                onChange={(e) => handleInputChange('max_trades_per_hour', parseInt(e.target.value))}
              />
              <span className="field-description">
                Maximum trades per user per hour (1-1,000)
              </span>
            </div>
          </div>
        )}

        {activeSection === 'impact' && (
          <div className="config-group">
            <h3>Price Impact Settings</h3>
            <div className="config-field">
              <label>Base Impact Factor (%)</label>
              <input
                type="number"
                min="0"
                max="1"
                step="0.01"
                value={(getCurrentValue('base_impact_factor') * 100).toFixed(2)}
                onChange={(e) => handleInputChange('base_impact_factor', parseFloat(e.target.value) / 100)}
              />
              <span className="field-description">
                Price impact per share traded (0% - 1%)
              </span>
            </div>
            <div className="config-field">
              <label>Large Trade Threshold</label>
              <input
                type="number"
                min="10"
                max="10000"
                value={getCurrentValue('large_trade_threshold')}
                onChange={(e) => handleInputChange('large_trade_threshold', parseInt(e.target.value))}
              />
              <span className="field-description">
                Shares needed to be considered a large trade (10-10,000)
              </span>
            </div>
            <div className="config-field">
              <label>Trend Amplification</label>
              <input
                type="number"
                min="1"
                max="5"
                step="0.1"
                value={getCurrentValue('trend_amplification')}
                onChange={(e) => handleInputChange('trend_amplification', parseFloat(e.target.value))}
              />
              <span className="field-description">
                How much trends amplify price impact (1.0 - 5.0)
              </span>
            </div>
          </div>
        )}
      </div>

      {hasChanges() && (
        <div className="config-changes-notice">
          <span>⚠️ You have unsaved changes</span>
        </div>
      )}
    </div>
  );
};

export default GameConfigSection;