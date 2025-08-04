import React, { useState, useEffect } from 'react';
import {
  getAllStocksDetailed,
  createStock,
  updateStock,
  deleteStock,
  launchIPO,
  triggerSectorEvent,
} from '../services/admin';
import './StockManagementSection.css';

const StockManagementSection = () => {
  const [stocks, setStocks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [editingStock, setEditingStock] = useState(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showIPOModal, setShowIPOModal] = useState(false);
  const [showSectorEventModal, setShowSectorEventModal] = useState(false);
  
  // Form states
  const [createForm, setCreateForm] = useState({
    symbol: '',
    name: '',
    sector: 'Technology',
    sector_id: 1,
    initial_price: 10.00,
    market_cap_category: 'mid',
    volatility_profile: 'normal',
    description: '',
  });
  
  const [ipoForm, setIPOForm] = useState({
    symbol: '',
    name: '',
    sector: 'Technology',
    sector_id: 1,
    ipo_price: 1.00,
    shares_available: 1000000,
  });
  
  const [sectorEventForm, setSectorEventForm] = useState({
    sector_id: 1,
    event_type: 'boom',
    impact_percentage: 10,
    duration_minutes: 60,
  });

  const sectors = [
    { id: 1, name: 'Technology' },
    { id: 2, name: 'Healthcare' },
    { id: 3, name: 'Finance' },
    { id: 4, name: 'Entertainment' },
    { id: 5, name: 'Energy' },
  ];

  useEffect(() => {
    fetchStocks();
  }, []);

  const fetchStocks = async () => {
    setLoading(true);
    try {
      const data = await getAllStocksDetailed();
      setStocks(data);
      setError(null);
    } catch (err) {
      setError('Failed to fetch stocks');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateStock = async (e) => {
    e.preventDefault();
    try {
      await createStock(createForm);
      setShowCreateModal(false);
      setCreateForm({
        symbol: '',
        name: '',
        sector: 'Technology',
        sector_id: 1,
        initial_price: 10.00,
        market_cap_category: 'mid',
        volatility_profile: 'normal',
        description: '',
      });
      fetchStocks();
    } catch (err) {
      setError('Failed to create stock');
    }
  };

  const handleUpdateStock = async (stockId, updates) => {
    try {
      console.log(`🔄 Updating stock ${stockId} with:`, updates);
      await updateStock(stockId, updates);
      console.log(`✅ Stock ${stockId} updated successfully`);
      setEditingStock(null);
      fetchStocks();
    } catch (err) {
      console.error(`❌ Failed to update stock ${stockId}:`, err);
      setError(`Failed to update stock: ${err.message}`);
    }
  };

  const handleDeleteStock = async (stockId) => {
    if (window.confirm('Are you sure you want to delist this stock?')) {
      try {
        await deleteStock(stockId);
        fetchStocks();
      } catch (err) {
        setError('Failed to delete stock');
      }
    }
  };

  const handleLaunchIPO = async (e) => {
    e.preventDefault();
    try {
      await launchIPO(ipoForm);
      setShowIPOModal(false);
      setIPOForm({
        symbol: '',
        name: '',
        sector: 'Technology',
        sector_id: 1,
        ipo_price: 1.00,
        shares_available: 1000000,
      });
      fetchStocks();
    } catch (err) {
      setError('Failed to launch IPO');
    }
  };

  const handleTriggerSectorEvent = async (e) => {
    e.preventDefault();
    try {
      await triggerSectorEvent(sectorEventForm);
      setShowSectorEventModal(false);
      setSectorEventForm({
        sector_id: 1,
        event_type: 'boom',
        impact_percentage: 10,
        duration_minutes: 60,
      });
      fetchStocks();
    } catch (err) {
      setError('Failed to trigger sector event');
    }
  };

  const handleInlineEdit = (stock, field, value) => {
    const updates = { [field]: value };
    console.log(`🔍 Inline edit triggered - Stock:`, stock);
    console.log(`🔍 Field: ${field}, Value:`, value);
    console.log(`🔍 Updates object:`, updates);
    handleUpdateStock(stock.id, updates);
  };

  return (
    <div className="stock-management-section">
      <div className="section-header">
        <h2>📊 Stock Management</h2>
        <div className="action-buttons">
          <button onClick={() => setShowCreateModal(true)} className="btn-primary">
            ➕ Create Stock
          </button>
          <button onClick={() => setShowIPOModal(true)} className="btn-ipo">
            🚀 Launch IPO
          </button>
          <button onClick={() => setShowSectorEventModal(true)} className="btn-event">
            💥 Sector Event
          </button>
          <button onClick={fetchStocks} className="btn-refresh">
            🔄 Refresh
          </button>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {loading ? (
        <div className="loading">Loading stocks...</div>
      ) : (
        <div className="stocks-table-container">
          <table className="stocks-table">
            <thead>
              <tr>
                <th>Symbol</th>
                <th>Name</th>
                <th>Sector</th>
                <th>Current Price</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {stocks.map((stock) => (
                <tr key={stock.id} className={stock.status === 'delisted' ? 'delisted' : ''}>
                  <td className="symbol">{stock.symbol}</td>
                  <td>
                    {editingStock === stock.id ? (
                      <input
                        type="text"
                        defaultValue={stock.name}
                        onBlur={(e) => handleInlineEdit(stock, 'name', e.target.value)}
                        autoFocus
                      />
                    ) : (
                      <span onClick={() => setEditingStock(stock.id)}>{stock.name}</span>
                    )}
                  </td>
                  <td>{stock.sector}</td>
                  <td className="price">
                    ${editingStock === stock.id ? (
                      <input
                        type="number"
                        step="0.01"
                        defaultValue={stock.current_price}
                        onBlur={(e) => handleInlineEdit(stock, 'current_price', parseFloat(e.target.value))}
                        style={{ width: '80px' }}
                      />
                    ) : (
                      <span onClick={() => setEditingStock(stock.id)}>
                        {stock.current_price.toFixed(2)}
                      </span>
                    )}
                  </td>
                  <td>
                    <span className={`status ${stock.status}`}>{stock.status}</span>
                  </td>
                  <td>
                    <button
                      onClick={() => handleDeleteStock(stock.id)}
                      className="btn-delete"
                      disabled={stock.status === 'delisted'}
                    >
                      🗑️
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create Stock Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Create New Stock</h3>
            <form onSubmit={handleCreateStock}>
              <div className="form-group">
                <label>Symbol:</label>
                <input
                  type="text"
                  value={createForm.symbol}
                  onChange={(e) => setCreateForm({ ...createForm, symbol: e.target.value.toUpperCase() })}
                  required
                  maxLength="10"
                />
              </div>
              <div className="form-group">
                <label>Company Name:</label>
                <input
                  type="text"
                  value={createForm.name}
                  onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Sector:</label>
                <select
                  value={createForm.sector_id}
                  onChange={(e) => {
                    const sectorId = parseInt(e.target.value);
                    const sector = sectors.find(s => s.id === sectorId);
                    setCreateForm({ 
                      ...createForm, 
                      sector_id: sectorId,
                      sector: sector.name 
                    });
                  }}
                >
                  {sectors.map((sector) => (
                    <option key={sector.id} value={sector.id}>
                      {sector.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Initial Price:</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  value={createForm.initial_price}
                  onChange={(e) => setCreateForm({ ...createForm, initial_price: parseFloat(e.target.value) })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Market Cap Category:</label>
                <select
                  value={createForm.market_cap_category}
                  onChange={(e) => setCreateForm({ ...createForm, market_cap_category: e.target.value })}
                >
                  <option value="penny">Penny Stock</option>
                  <option value="small">Small Cap</option>
                  <option value="mid">Mid Cap</option>
                  <option value="large">Large Cap</option>
                </select>
              </div>
              <div className="form-group">
                <label>Volatility Profile:</label>
                <select
                  value={createForm.volatility_profile}
                  onChange={(e) => setCreateForm({ ...createForm, volatility_profile: e.target.value })}
                >
                  <option value="stable">Stable</option>
                  <option value="normal">Normal</option>
                  <option value="volatile">Volatile</option>
                  <option value="extreme">Extreme</option>
                </select>
              </div>
              <div className="form-group">
                <label>Description:</label>
                <textarea
                  value={createForm.description}
                  onChange={(e) => setCreateForm({ ...createForm, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-primary">Create Stock</button>
                <button type="button" onClick={() => setShowCreateModal(false)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* IPO Modal */}
      {showIPOModal && (
        <div className="modal-overlay" onClick={() => setShowIPOModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>🚀 Launch IPO</h3>
            <form onSubmit={handleLaunchIPO}>
              <div className="form-group">
                <label>Symbol:</label>
                <input
                  type="text"
                  value={ipoForm.symbol}
                  onChange={(e) => setIPOForm({ ...ipoForm, symbol: e.target.value.toUpperCase() })}
                  required
                  maxLength="10"
                />
              </div>
              <div className="form-group">
                <label>Company Name:</label>
                <input
                  type="text"
                  value={ipoForm.name}
                  onChange={(e) => setIPOForm({ ...ipoForm, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Sector:</label>
                <select
                  value={ipoForm.sector_id}
                  onChange={(e) => {
                    const sectorId = parseInt(e.target.value);
                    const sector = sectors.find(s => s.id === sectorId);
                    setIPOForm({ 
                      ...ipoForm, 
                      sector_id: sectorId,
                      sector: sector.name 
                    });
                  }}
                >
                  {sectors.map((sector) => (
                    <option key={sector.id} value={sector.id}>
                      {sector.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>IPO Price (Penny Stock: $0.01-$5):</label>
                <input
                  type="number"
                  step="0.01"
                  min="0.01"
                  max="100"
                  value={ipoForm.ipo_price}
                  onChange={(e) => setIPOForm({ ...ipoForm, ipo_price: parseFloat(e.target.value) })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Shares Available:</label>
                <input
                  type="number"
                  min="1000"
                  step="1000"
                  value={ipoForm.shares_available}
                  onChange={(e) => setIPOForm({ ...ipoForm, shares_available: parseInt(e.target.value) })}
                  required
                />
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-ipo">Launch IPO</button>
                <button type="button" onClick={() => setShowIPOModal(false)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Sector Event Modal */}
      {showSectorEventModal && (
        <div className="modal-overlay" onClick={() => setShowSectorEventModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>💥 Trigger Sector Event</h3>
            <form onSubmit={handleTriggerSectorEvent}>
              <div className="form-group">
                <label>Sector:</label>
                <select
                  value={sectorEventForm.sector_id}
                  onChange={(e) => setSectorEventForm({ ...sectorEventForm, sector_id: parseInt(e.target.value) })}
                >
                  {sectors.map((sector) => (
                    <option key={sector.id} value={sector.id}>
                      {sector.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Event Type:</label>
                <select
                  value={sectorEventForm.event_type}
                  onChange={(e) => setSectorEventForm({ ...sectorEventForm, event_type: e.target.value })}
                >
                  <option value="boom">💹 Boom</option>
                  <option value="crash">💥 Crash</option>
                </select>
              </div>
              <div className="form-group">
                <label>Impact Percentage:</label>
                <input
                  type="number"
                  min="1"
                  max="50"
                  value={sectorEventForm.impact_percentage}
                  onChange={(e) => setSectorEventForm({ ...sectorEventForm, impact_percentage: parseInt(e.target.value) })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Duration (minutes):</label>
                <input
                  type="number"
                  min="1"
                  max="1440"
                  value={sectorEventForm.duration_minutes}
                  onChange={(e) => setSectorEventForm({ ...sectorEventForm, duration_minutes: parseInt(e.target.value) })}
                  required
                />
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-event">Trigger Event</button>
                <button type="button" onClick={() => setShowSectorEventModal(false)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default StockManagementSection;