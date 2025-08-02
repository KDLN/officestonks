import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getUserPortfolio, getTransactionHistory } from '../services/stock';
import { initWebSocket, addWebSocketListener, removeWebSocketListener } from '../services/websocket';
import { initSSE, addSSEListener, removeSSEListener, isSSEConnected } from '../services/sse';
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
  const [isWebSocketConnected, setIsWebSocketConnected] = useState(false);
  const [pollingInterval, setPollingInterval] = useState(null);

  useEffect(() => {
    loadPortfolioData();
    
    // Initialize WebSocket for chat and notifications (keeping existing functionality)
    const wsTimeout = setTimeout(() => {
      console.log('🔌 Initializing WebSocket for Portfolio chat/notifications...');
      initWebSocket().catch(err => {
        console.error('❌ Failed to initialize WebSocket:', err);
      });
    }, 500);

    // Initialize SSE for stock price updates
    const sseTimeout = setTimeout(() => {
      console.log('📡 Initializing SSE for Portfolio stock updates...');
      initSSE().catch(err => {
        console.error('❌ Failed to initialize SSE:', err);
        // Start polling fallback after SSE fails
        setTimeout(() => {
          if (!isSSEConnected()) {
            startPolling();
          }
        }, 2000);
      });
    }, 750);

    // Handle page visibility changes to pause/resume polling
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && !isSSEConnected() && !pollingInterval) {
        console.log('📱 Page became visible - resuming polling');
        startPolling();
      } else if (document.visibilityState === 'hidden') {
        console.log('📱 Page became hidden - pausing polling');
        stopPolling();
      }
    };
    
    document.addEventListener('visibilitychange', handleVisibilityChange);

    // Listen for SSE connection state changes
    const handleSSEConnectionState = (state) => {
      if (state.state === 'connected') {
        console.log('✅ SSE connected - stopping polling fallback');
        setIsWebSocketConnected(true); // Reusing this state for SSE connection
        stopPolling();
      } else if (state.state === 'disconnected' || state.state === 'failed') {
        console.log('❌ SSE disconnected - will start polling fallback');
        setIsWebSocketConnected(false);
        // Start polling fallback after a short delay
        setTimeout(() => {
          if (!isSSEConnected() && document.visibilityState === 'visible') {
            startPolling();
          }
        }, 3000);
      }
    };

    // Listen for SSE stock updates to update portfolio in real-time
    const handleStockUpdate = (message) => {
      console.log('Portfolio received SSE stock update:', message);
      
      const stock_id = message.stock_id;
      const price = message.price;
      
      if (!stock_id || !price) {
        console.log('Missing required fields in SSE stock update:', message);
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
              console.log(`SSE update: ${item.stock.symbol} price from $${oldPrice} to $${price}`);
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
        
        // Calculate new gains/losses for comparison
        const newPortfolioTemp = {
          ...prevPortfolio,
          portfolio_items: updatedItems,
          stock_value: newStockValue,
          total_value: newTotalValue
        };
        
        // Calculate old and new gains/losses
        const oldGainLoss = calculateTotalGainLossForPortfolio(prevPortfolio);
        const newGainLoss = calculateTotalGainLossForPortfolio(newPortfolioTemp);
        
        // Track total value and stock value change direction
        const oldStockValue = prevPortfolio.stock_value || 0;
        if (newTotalValue !== oldTotalValue || newStockValue !== oldStockValue) {
          setValueChanges(prev => ({
            ...prev,
            totalValue: newTotalValue > oldTotalValue ? 'up' : 'down',
            stockValue: newStockValue > oldStockValue ? 'up' : 'down',
            totalGainLoss: newGainLoss.amount > oldGainLoss.amount ? 'up' : 'down'
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
    };
    
    // Add SSE listeners
    addSSEListener('stockUpdate', handleStockUpdate);
    addSSEListener('connectionState', handleSSEConnectionState);
    
    // Cleanup on unmount
    return () => {
      clearTimeout(wsTimeout);
      clearTimeout(sseTimeout);
      removeSSEListener('stockUpdate', handleStockUpdate);
      removeSSEListener('connectionState', handleSSEConnectionState);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      stopPolling();
    };
  }, []);

  const loadPortfolioData = async (showLoading = true) => {
    try {
      if (showLoading) {
        setLoading(true);
        setError(null);
      }

      // Load portfolio and transactions in parallel
      const [portfolioData, transactionsData] = await Promise.all([
        getUserPortfolio(),
        showLoading ? getTransactionHistory(50, 0) : Promise.resolve(transactions)
      ]);

      // Compare with previous portfolio for animations
      if (!showLoading && portfolio) {
        const oldTotalValue = portfolio.total_value || 0;
        const newTotalValue = portfolioData.total_value || 0;
        const oldStockValue = portfolio.stock_value || 0;
        const newStockValue = portfolioData.stock_value || 0;

        // Check if any individual stock prices changed
        const stockChanges = {};
        if (portfolio.portfolio_items && portfolioData.portfolio_items) {
          portfolioData.portfolio_items.forEach(newItem => {
            const oldItem = portfolio.portfolio_items.find(item => item.stock_id === newItem.stock_id);
            if (oldItem && oldItem.stock.current_price !== newItem.stock.current_price) {
              const direction = newItem.stock.current_price > oldItem.stock.current_price ? 'up' : 'down';
              stockChanges[`stock_${newItem.stock_id}`] = direction;
              console.log(`Polling update: ${newItem.stock.symbol} price from $${oldItem.stock.current_price} to $${newItem.stock.current_price}`);
            }
          });
        }

        // Set animations for value changes
        if (newTotalValue !== oldTotalValue || newStockValue !== oldStockValue || Object.keys(stockChanges).length > 0) {
          setValueChanges(prev => ({
            ...prev,
            ...stockChanges,
            totalValue: newTotalValue > oldTotalValue ? 'up' : (newTotalValue < oldTotalValue ? 'down' : ''),
            stockValue: newStockValue > oldStockValue ? 'up' : (newStockValue < oldStockValue ? 'down' : ''),
            totalGainLoss: newTotalValue > oldTotalValue ? 'up' : (newTotalValue < oldTotalValue ? 'down' : '')
          }));

          // Clear animations after 2 seconds
          setTimeout(() => {
            setValueChanges(prev => {
              const cleared = { ...prev };
              cleared.totalValue = '';
              cleared.stockValue = '';
              cleared.totalGainLoss = '';
              Object.keys(cleared).forEach(key => {
                if (key.startsWith('stock_')) {
                  cleared[key] = '';
                }
              });
              return cleared;
            });
          }, 2000);
        }
      }

      setPortfolio(portfolioData);
      if (showLoading) {
        setTransactions(transactionsData);
      }
    } catch (err) {
      console.error('Error loading portfolio data:', err);
      if (showLoading) {
        setError('Failed to load portfolio data. Please try again.');
      }
    } finally {
      if (showLoading) {
        setLoading(false);
      }
    }
  };

  // Start polling as fallback when WebSocket fails
  const startPolling = () => {
    console.log('🔄 Starting polling fallback for real-time updates (every 10 seconds, only when visible)');
    const interval = setInterval(() => {
      // Only poll if the page is visible and WebSocket is not connected
      if (!isWebSocketConnected && document.visibilityState === 'visible') {
        loadPortfolioData(false); // Don't show loading spinner for polling updates
      }
    }, 10000); // Poll every 10 seconds to avoid rate limiting
    
    setPollingInterval(interval);
    return interval;
  };

  // Stop polling when WebSocket connects
  const stopPolling = () => {
    if (pollingInterval) {
      console.log('🛑 Stopping polling fallback - WebSocket connected');
      clearInterval(pollingInterval);
      setPollingInterval(null);
    }
  };

  const calculateTotalGainLoss = React.useCallback(() => {
    if (!portfolio || !portfolio.portfolio_items || !transactions) return { amount: 0, percentage: 0 };
    
    let totalCost = 0;
    let totalValue = 0;
    
    portfolio.portfolio_items.forEach(item => {
      const currentValue = item.quantity * item.stock.current_price;
      
      // Calculate average cost from transaction history
      const itemTransactions = transactions.filter(t => t.stock_id === item.stock_id);
      const buyTransactions = itemTransactions.filter(t => t.transaction_type === 'buy');
      
      let avgCost = item.stock.current_price; // fallback
      if (buyTransactions.length > 0) {
        const totalQuantityBought = buyTransactions.reduce((sum, t) => sum + t.quantity, 0);
        const totalCostBought = buyTransactions.reduce((sum, t) => sum + (t.quantity * t.price), 0);
        avgCost = totalQuantityBought > 0 ? totalCostBought / totalQuantityBought : item.stock.current_price;
      }
      
      const cost = item.quantity * avgCost;
      
      totalCost += cost;
      totalValue += currentValue;
    });
    
    const gainLoss = totalValue - totalCost;
    const percentage = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;
    
    return { amount: gainLoss, percentage };
  }, [portfolio, transactions]);

  // Helper function to calculate gains/losses for any portfolio object (used for real-time comparisons)
  const calculateTotalGainLossForPortfolio = (portfolioData) => {
    if (!portfolioData || !portfolioData.portfolio_items || portfolioData.portfolio_items.length === 0 || !transactions) {
      return { amount: 0, percentage: 0 };
    }

    let totalCost = 0;
    let totalValue = 0;

    portfolioData.portfolio_items.forEach(item => {
      if (!item || !item.stock || !item.quantity) return;

      const currentPrice = item.stock.current_price || 0;
      const currentValue = item.quantity * currentPrice;
      
      // Calculate average cost from transaction history
      const itemTransactions = transactions.filter(t => t.stock_id === item.stock_id);
      const buyTransactions = itemTransactions.filter(t => t.transaction_type === 'buy');
      
      let avgCost = currentPrice; // fallback
      if (buyTransactions.length > 0) {
        const totalQuantityBought = buyTransactions.reduce((sum, t) => sum + t.quantity, 0);
        const totalCostBought = buyTransactions.reduce((sum, t) => sum + (t.quantity * t.price), 0);
        avgCost = totalQuantityBought > 0 ? totalCostBought / totalQuantityBought : currentPrice;
      }
      
      const cost = item.quantity * avgCost;
      
      totalCost += cost;
      totalValue += currentValue;
    });
    
    const gainLoss = totalValue - totalCost;
    const percentage = totalCost > 0 ? (gainLoss / totalCost) * 100 : 0;
    
    return { amount: gainLoss, percentage };
  };


  // Recalculate gains/losses whenever portfolio changes (for real-time updates)
  const { amount: totalGainLoss, percentage: totalGainLossPercentage } = React.useMemo(() => {
    return calculateTotalGainLoss();
  }, [portfolio, calculateTotalGainLoss]);

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
                      
                      // Calculate average cost from transaction history
                      const itemTransactions = transactions.filter(t => t.stock_id === item.stock_id);
                      const buyTransactions = itemTransactions.filter(t => t.transaction_type === 'buy');
                      
                      let avgCost = item.stock.current_price; // fallback
                      if (buyTransactions.length > 0) {
                        const totalQuantityBought = buyTransactions.reduce((sum, t) => sum + t.quantity, 0);
                        const totalCostBought = buyTransactions.reduce((sum, t) => sum + (t.quantity * t.price), 0);
                        avgCost = totalQuantityBought > 0 ? totalCostBought / totalQuantityBought : item.stock.current_price;
                      }
                      
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