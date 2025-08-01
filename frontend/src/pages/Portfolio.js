import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getUserPortfolio, getTransactionHistory } from '../services/stock';
import { initWebSocket, addWebSocketListener, removeWebSocketListener } from '../services/websocket';
import Navigation from '../components/Navigation';
import LoadingSpinner from '../components/LoadingSpinner';
import PortfolioCrisisAlert from '../components/PortfolioCrisisAlert';
import { formatCurrency, formatPercentage, formatDate } from '../utils';
import './Dashboard.css';  // Import Dashboard styles for consistent UI
import './Portfolio.css';

function Portfolio() {
  const [portfolio, setPortfolio] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('holdings');
  const [valueChanges, setValueChanges] = useState({}); // Track value changes for animations

  useEffect(() => {
    loadPortfolioData();
    
    // Initialize WebSocket for real-time updates
    const wsTimeout = setTimeout(() => {
      console.log('🔌 Initializing WebSocket for Portfolio real-time updates...');
      initWebSocket().catch(err => {
        console.error('❌ Failed to initialize WebSocket:', err);
      });
    }, 500);

    // Listen for stock updates to update portfolio in real-time
    const handleStockUpdate = (message) => {
      console.log('Portfolio received stock update:', message);
      
      // Process stock update message
      if (message.type === 'stock_update' || (message.id && message.current_price)) {
        const stock_id = message.stock_id || message.id;
        const price = message.price || message.current_price;
        
        if (!stock_id || !price) {
          console.log('Missing required fields in stock update:', message);
          return;
        }
        
        // Update portfolio with new stock prices
        setPortfolio(prevPortfolio => {
          if (!prevPortfolio || !prevPortfolio.portfolio_items) {
            return prevPortfolio;
          }
          
          const oldTotalValue = prevPortfolio.total_value || 0;
          let hasChanges = false;
          
          // Update portfolio items if the stock is in portfolio
          const updatedItems = prevPortfolio.portfolio_items.map(item => {
            if (!item || !item.stock) return item;
            
            if (item.stock_id === stock_id) {
              const oldPrice = item.stock.current_price;
              if (oldPrice !== price) {
                console.log(`Updating ${item.stock.symbol} price from $${oldPrice} to $${price}`);
                hasChanges = true;
                
                // Track the direction of change for this item
                setValueChanges(prev => ({
                  ...prev,
                  [`stock_${stock_id}`]: price > oldPrice ? 'up' : 'down'
                }));
                
                return {
                  ...item,
                  stock: { ...item.stock, current_price: price }
                };
              }
            }
            return item;
          });
          
          if (!hasChanges) return prevPortfolio;
          
          // Recalculate stock value
          const newStockValue = updatedItems.reduce(
            (total, item) => {
              if (!item || !item.stock) return total;
              return total + (item.quantity * item.stock.current_price);
            },
            0
          );
          
          const newTotalValue = newStockValue + (prevPortfolio.cash_balance || 0);
          
          // Track total value and stock value change direction
          const oldStockValue = prevPortfolio.stock_value || 0;
          if (newTotalValue !== oldTotalValue || newStockValue !== oldStockValue) {
            setValueChanges(prev => ({
              ...prev,
              totalValue: newTotalValue > oldTotalValue ? 'up' : 'down',
              stockValue: newStockValue > oldStockValue ? 'up' : 'down',
              totalGainLoss: newTotalValue > oldTotalValue ? 'up' : 'down'
            }));
            
            // Clear animation classes after animation completes
            setTimeout(() => {
              setValueChanges(prev => {
                const cleared = { ...prev };
                // Clear total values
                cleared.totalValue = '';
                cleared.stockValue = '';
                cleared.totalGainLoss = '';
                // Clear individual stock animations
                Object.keys(cleared).forEach(key => {
                  if (key.startsWith('stock_')) {
                    cleared[key] = '';
                  }
                });
                return cleared;
              });
            }, 2000);
          }
          
          return {
            ...prevPortfolio,
            portfolio_items: updatedItems,
            stock_value: newStockValue,
            total_value: newTotalValue
          };
        });
      }
    };
    
    addWebSocketListener('*', handleStockUpdate);
    
    // Cleanup on unmount
    return () => {
      clearTimeout(wsTimeout);
      removeWebSocketListener('*', handleStockUpdate);
    };
  }, []);

  const loadPortfolioData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Load portfolio and transactions in parallel
      const [portfolioData, transactionsData] = await Promise.all([
        getUserPortfolio(),
        getTransactionHistory(50, 0)
      ]);

      setPortfolio(portfolioData);
      setTransactions(transactionsData);
    } catch (err) {
      console.error('Error loading portfolio data:', err);
      setError('Failed to load portfolio data. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const calculateTotalGainLoss = () => {
    if (!portfolio || !portfolio.portfolio_items) return { amount: 0, percentage: 0 };
    
    let totalCost = 0;
    let totalValue = 0;
    
    portfolio.portfolio_items.forEach(item => {
      const currentValue = item.quantity * item.stock.current_price;
      const avgCost = item.average_cost || item.stock.current_price;
      const cost = item.quantity * avgCost;
      
      totalCost += cost;
      totalValue += currentValue;
    });
    
    const gainLoss = totalValue - totalCost;
    const percentage = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;
    
    return { amount: gainLoss, percentage };
  };


  if (loading) {
    return (
      <div className="portfolio-page">
        <Navigation />
        <LoadingSpinner message="Loading portfolio data..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className="portfolio-page">
        <Navigation />
        <div className="portfolio-container">
          <div className="error">{error}</div>
        </div>
      </div>
    );
  }

  const { amount: totalGainLoss, percentage: totalGainLossPercentage } = calculateTotalGainLoss();

  return (
    <div className="portfolio-page">
      <Navigation />
      <div className="portfolio-container">
        <div className="portfolio-header">
          <h1>My Portfolio</h1>
          
          {/* Crisis Alerts */}
          <PortfolioCrisisAlert portfolioItems={portfolio?.portfolio_items || []} />
          
          {/* Portfolio Summary */}
          <div className="portfolio-summary">
            <div className={`summary-card ${valueChanges.totalValue ? `value-${valueChanges.totalValue}` : ''}`}>
              <h3>Total Value</h3>
              <div className="value">{formatCurrency(portfolio?.total_value || 0)}</div>
            </div>
            
            <div className="summary-card">
              <h3>Cash Balance</h3>
              <div className="value">{formatCurrency(portfolio?.cash_balance || 0)}</div>
            </div>
            
            <div className={`summary-card ${valueChanges.stockValue ? `value-${valueChanges.stockValue}` : ''}`}>
              <h3>Stock Value</h3>
              <div className="value">{formatCurrency(portfolio?.stock_value || 0)}</div>
            </div>
            
            <div className={`summary-card ${valueChanges.totalGainLoss ? `value-${valueChanges.totalGainLoss}` : ''}`}>
              <h3>Total Gain/Loss</h3>
              <div className={`value ${totalGainLoss >= 0 ? 'positive' : 'negative'}`}>
                {formatCurrency(totalGainLoss)}
              </div>
              <div className={`percentage ${totalGainLoss >= 0 ? 'positive' : 'negative'}`}>
                {formatPercentage(totalGainLossPercentage)}
              </div>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="portfolio-tabs">
          <button
            className={`tab-button ${activeTab === 'holdings' ? 'active' : ''}`}
            onClick={() => setActiveTab('holdings')}
          >
            Holdings
          </button>
          <button
            className={`tab-button ${activeTab === 'transactions' ? 'active' : ''}`}
            onClick={() => setActiveTab('transactions')}
          >
            Transaction History
          </button>
        </div>

        {/* Tab Content */}
        <div className="portfolio-content">
          {activeTab === 'holdings' && (
            <>
              {portfolio?.portfolio_items && portfolio.portfolio_items.length > 0 ? (
                <table className="dashboard-table">
                  <thead>
                    <tr>
                      <th>Stock</th>
                      <th>Quantity</th>
                      <th>Current Price</th>
                      <th>Market Value</th>
                      <th>Avg Cost</th>
                      <th>Total Cost</th>
                      <th>Gain/Loss</th>
                      <th>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {portfolio.portfolio_items.map((item) => {
                      const marketValue = item.quantity * item.stock.current_price;
                      const avgCost = item.average_cost || item.stock.current_price;
                      const totalCost = item.quantity * avgCost;
                      const gainLoss = marketValue - totalCost;
                      const gainLossPercentage = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;

                      const changeClass = valueChanges[`stock_${item.stock_id}`] ? `value-${valueChanges[`stock_${item.stock_id}`]}` : '';
                      
                      return (
                        <tr key={item.id} className={changeClass}>
                          <td>
                            <div className="stock-info">
                              <span className="stock-symbol">{item.stock.symbol}</span>
                              <span className="stock-name">{item.stock.name}</span>
                            </div>
                          </td>
                          <td>{item.quantity}</td>
                          <td><strong>{formatCurrency(item.stock.current_price)}</strong></td>
                          <td><strong>{formatCurrency(marketValue)}</strong></td>
                          <td>{formatCurrency(avgCost)}</td>
                          <td>{formatCurrency(totalCost)}</td>
                          <td>
                            <div className={`gain-loss ${gainLoss >= 0 ? 'positive' : 'negative'}`}>
                              <span>{gainLoss >= 0 ? '▲' : '▼'}</span>
                              <span>{formatCurrency(Math.abs(gainLoss))}</span>
                              <span>({formatPercentage(gainLossPercentage)})</span>
                            </div>
                          </td>
                          <td>
                            <Link to={`/stock/${item.stock_id}`} className="trade-button">
                              Trade
                            </Link>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              ) : (
                <div className="empty-portfolio">
                  <p>You don't own any stocks yet.</p>
                  <Link to="/stocks" className="action-button">Start Trading</Link>
                </div>
              )}
            </>
          )}

          {activeTab === 'transactions' && (
            <>
              {transactions && transactions.length > 0 ? (
                <table className="dashboard-table">
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Type</th>
                      <th>Stock</th>
                      <th>Quantity</th>
                      <th>Price</th>
                      <th>Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactions.map((transaction) => {
                      const total = transaction.quantity * transaction.price;
                      const isBuy = transaction.transaction_type === 'buy';

                      return (
                        <tr key={transaction.id}>
                          <td>{formatDate(transaction.created_at)}</td>
                          <td>
                            <span className={`transaction-type ${isBuy ? 'buy' : 'sell'}`}>
                              {isBuy ? 'BUY' : 'SELL'}
                            </span>
                          </td>
                          <td>
                            <div className="stock-info">
                              <span className="stock-symbol">{transaction.stock.symbol}</span>
                              <span className="stock-name">{transaction.stock.name}</span>
                            </div>
                          </td>
                          <td>{transaction.quantity}</td>
                          <td>{formatCurrency(transaction.price)}</td>
                          <td><strong>{formatCurrency(total)}</strong></td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              ) : (
                <div className="empty-portfolio">
                  <p>No transactions yet.</p>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default Portfolio;