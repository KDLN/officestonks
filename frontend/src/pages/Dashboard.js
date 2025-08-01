import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getUserPortfolio, getTransactionHistory, getAllStocks } from '../services/stock';
import { initWebSocket, addWebSocketListener, getWebSocketInstance } from '../services/websocket';
import { initSSE, addSSEListener, removeSSEListener, isSSEConnected } from '../services/sse';
import { safeGetItem, safeSetItem } from '../utils';
import logger from '../services/logger';
import Navigation from '../components/Navigation';
import Chat from '../components/Chat';
import NewsDisplay from '../components/NewsDisplay';
import NewsTicker from '../components/NewsTicker';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import CrisisAlertManager from '../components/CrisisAlertManager';
import './Dashboard.css';

// Default empty states to prevent null references
const DEFAULT_PORTFOLIO = {
  portfolio_items: [],
  cash_balance: 0,
  stock_value: 0,
  total_value: 0
};

const Dashboard = () => {
  const [portfolio, setPortfolio] = useState(DEFAULT_PORTFOLIO);
  const [transactions, setTransactions] = useState([]);
  const [topStocks, setTopStocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [socket, setSocket] = useState(null);
  const [chatDrawerOpen, setChatDrawerOpen] = useState(() => {
    const saved = safeGetItem('chatDrawerOpen', 'false');
    return saved === 'true';
  });

  // Save chat drawer state to localStorage
  const toggleChatDrawer = () => {
    const newState = !chatDrawerOpen;
    setChatDrawerOpen(newState);
    safeSetItem('chatDrawerOpen', newState.toString());
  };

  // Helper function to fetch dashboard data with logging
  const fetchDashboardData = async () => {
    return logger.performance('Dashboard data fetch', async () => {
      const [portfolioData, transactionsData, stocksData] = await Promise.all([
        getUserPortfolio(),
        getTransactionHistory(5),
        getAllStocks()
      ]);

      logger.debug('Dashboard data fetched successfully', {
        portfolioItems: portfolioData?.portfolio_items?.length || 0,
        transactionCount: transactionsData?.length || 0,
        stockCount: stocksData?.length || 0
      });

      return { portfolioData, transactionsData, stocksData };
    });
  };

  // Helper function to process stocks data
  const processStocksData = (stocksData) => {
    const sortedStocks = [...stocksData].sort((a, b) => b.current_price - a.current_price);
    const topStocks = sortedStocks.slice(0, 5);
    
    logger.debug('Top stocks calculated', {
      topStocks: topStocks.map(s => ({ symbol: s.symbol, price: s.current_price }))
    });
    
    return topStocks;
  };

  // Simplified main fetch function
  const fetchData = async () => {
    logger.info('Starting dashboard data fetch');
    
    try {
      const { portfolioData, transactionsData, stocksData } = await fetchDashboardData();
      
      // Update state
      setPortfolio(portfolioData);
      setTransactions(transactionsData);
      setTopStocks(processStocksData(stocksData));
      
      setLoading(false);
      logger.info('Dashboard data loading completed successfully');
      
    } catch (err) {
      const errorMessage = err.response?.data?.error || 'Unable to load dashboard. Please check your connection and try again.';
      
      logger.error('Dashboard data fetch failed', {
        error: err.message,
        status: err.response?.status,
        statusText: err.response?.statusText
      });
      
      setError(errorMessage);
      setLoading(false);
    }
  };

  useEffect(() => {
    console.log('🚀 Dashboard component mounted');
    fetchData();

    // Initialize WebSocket connection for chat (with delay to ensure auth is ready)
    console.log('⏳ Setting up WebSocket initialization timer...');
    const wsTimeout = setTimeout(() => {
      console.log('🔌 Initializing WebSocket connection for chat...');
      initWebSocket().then(() => {
        console.log('✅ WebSocket initialized successfully');
        // Set the socket instance for the CrisisAlertManager
        setSocket(getWebSocketInstance());
      }).catch(err => {
        console.error('❌ Failed to initialize WebSocket:', err);
        // Don't block the UI, just log the error
      });
    }, 500);

    // Initialize SSE connection for stock updates
    console.log('⏳ Setting up SSE initialization timer...');
    const sseTimeout = setTimeout(() => {
      console.log('📡 Initializing SSE connection for stock updates...');
      initSSE().then(() => {
        console.log('✅ SSE initialized successfully');
      }).catch(err => {
        console.error('❌ Failed to initialize SSE:', err);
        // Don't block the UI, just log the error
      });
    }, 750);

    // Listen for stock updates via SSE
    const handleStockUpdate = (message) => {
      console.log('Received SSE stock update on dashboard:', message);

      // Extract stock_id and price from SSE message
      const stock_id = message.stock_id;
      const price = message.price;

      if (!stock_id || !price) {
        console.log('Missing required fields in SSE message:', message);
        return;
      }

      // Update portfolio stocks if affected
      setPortfolio(prevPortfolio => {
        // If portfolio is null or undefined, use the default empty portfolio
        const portfolio = prevPortfolio || DEFAULT_PORTFOLIO;
        if (!portfolio.portfolio_items) {
          return portfolio;
        }

        // Update portfolio items if the stock is in portfolio
        const updatedItems = portfolio.portfolio_items.map(item => {
          if (!item || !item.stock) return item;

          if (item.stock_id === stock_id) {
            const oldValue = item.quantity * (item.stock.current_price || 0);
            const newValue = item.quantity * price;
            const updatedStock = { ...item.stock, current_price: price };

            return {
              ...item,
              stock: updatedStock,
              valueChange: oldValue < newValue ? 'up' : 'down'
            };
          }
          return item;
        });

        // Recalculate stock value with null safety
        const newStockValue = updatedItems.reduce(
          (total, item) => {
            if (!item || !item.stock) return total;
            return total + (item.quantity * (item.stock.current_price || 0));
          },
          0
        );

        return {
          ...portfolio,
          portfolio_items: updatedItems,
          stock_value: newStockValue,
          total_value: (portfolio.cash_balance || 0) + newStockValue
        };
      });

      // Update top stocks if affected
      setTopStocks(prevTopStocks => {
        if (!prevTopStocks || !Array.isArray(prevTopStocks)) {
          return prevTopStocks || [];
        }

        return prevTopStocks.map(stock => {
          if (!stock) return stock;

          if (stock.id === stock_id) {
            return {
              ...stock,
              current_price: price,
              priceChange: (stock.current_price || 0) < price ? 'up' : 'down'
            };
          }
          return stock;
        });
      });
    };

    // Add SSE listener for stock updates
    addSSEListener('stockUpdate', handleStockUpdate);

    // Clean up on unmount
    return () => {
      clearTimeout(wsTimeout);
      clearTimeout(sseTimeout);
      removeSSEListener('stockUpdate', handleStockUpdate);
      // Don't close WebSocket or SSE on component unmount as they're shared
    };
  }, []);

  if (loading) {
    console.log('⏳ Dashboard showing loading state');
    return (
      <div className="dashboard-page">
        <Navigation />
        <LoadingSpinner message="Loading dashboard..." />
      </div>
    );
  }

  if (error) {
    console.log('❌ Dashboard showing error state:', error);
    return (
      <div className="dashboard-page">
        <Navigation />
        <div className="dashboard-container">
          <ErrorMessage 
            message={error} 
            onRetry={() => {
              console.log('🔄 Dashboard retry requested');
              setError(null);
              setLoading(true);
              fetchData();
            }}
          />
        </div>
      </div>
    );
  }

  console.log('🎉 Dashboard rendering main content with data:', {
    portfolioValue: portfolio?.total_value || 0,
    portfolioItems: portfolio?.portfolio_items?.length || 0,
    topStocks: topStocks?.length || 0,
    transactions: transactions?.length || 0
  });

  return (
    <div className="dashboard-page">
      <Navigation />
      <div className="dashboard-container">
        <div className="dashboard-header">
          <h1>Dashboard</h1>
          <button 
            className="chat-toggle-btn"
            onClick={toggleChatDrawer}
            aria-label="Toggle Chat"
          >
            💬 Chat {chatDrawerOpen ? '✕' : ''}
          </button>
        </div>
        
        {/* Breaking News Ticker */}
        <NewsTicker />
        
        <div className="dashboard-main">
          {/* Left Column - News */}
          <aside className="dashboard-left-column">
            <div className="portfolio-value">
              <h2>Total Portfolio Value</h2>
              <div className="value">${(portfolio?.total_value || 0).toFixed(2)}</div>
              <div className="portfolio-breakdown">
                <div className="breakdown-item">
                  <span>Cash:</span>
                  <span>${(portfolio?.cash_balance || 0).toFixed(2)}</span>
                </div>
                <div className="breakdown-item">
                  <span>Stocks:</span>
                  <span>${(portfolio?.stock_value || 0).toFixed(2)}</span>
                </div>
              </div>
            </div>
            
            <NewsDisplay />
          </aside>
          
          {/* Main Content - Portfolio and Stocks */}
          <main className="dashboard-main-content">
        
        <div className="dashboard-content">
          <div className="dashboard-section">
            <div className="section-header">
              <h2>Your Portfolio</h2>
              <Link to="/portfolio" className="view-all">View All</Link>
            </div>
            
            {portfolio?.portfolio_items && portfolio.portfolio_items.length > 0 ? (
              <div className="portfolio-list">
                <table className="dashboard-table">
                  <thead>
                    <tr>
                      <th>Symbol</th>
                      <th>Shares</th>
                      <th>Price</th>
                      <th>Value</th>
                      <th>Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {portfolio.portfolio_items?.slice(0, 3).map(item => item && item.stock ? (
                      <tr
                        key={item.stock_id}
                        className={item.valueChange ? `value-${item.valueChange}` : ''}
                      >
                        <td>{item.stock.symbol}</td>
                        <td>{item.quantity}</td>
                        <td>${(item.stock.current_price || 0).toFixed(2)}</td>
                        <td>${(item.quantity * (item.stock.current_price || 0)).toFixed(2)}</td>
                        <td>
                          <Link to={`/stock/${item.stock_id}`} className="trade-button">
                            Trade
                          </Link>
                        </td>
                      </tr>
                    ) : null)}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="empty-list">
                <p>You don't own any stocks yet.</p>
                <Link to="/stocks" className="action-button">Start Trading</Link>
              </div>
            )}
          </div>
          
          <div className="dashboard-section">
            <div className="section-header">
              <h2>Top Stocks</h2>
              <Link to="/stocks" className="view-all">View All</Link>
            </div>
            
            <div className="top-stocks-list">
              <table className="dashboard-table">
                <thead>
                  <tr>
                    <th>Symbol</th>
                    <th>Name</th>
                    <th>Price</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {topStocks && topStocks.length > 0 ? topStocks.map(stock => stock ? (
                    <tr
                      key={stock.id}
                      className={stock.priceChange ? `price-${stock.priceChange}` : ''}
                    >
                      <td>{stock.symbol}</td>
                      <td>{stock.name}</td>
                      <td>${(stock.current_price || 0).toFixed(2)}</td>
                      <td>
                        <Link to={`/stock/${stock.id}`} className="trade-button">
                          Trade
                        </Link>
                      </td>
                    </tr>
                  ) : null) : (
                    <tr>
                      <td colSpan="4" className="empty-message">No stocks available</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
          
          <div className="dashboard-section">
            <div className="section-header">
              <h2>Recent Transactions</h2>
              <Link to="/transactions" className="view-all">View All</Link>
            </div>
            
            <div className="transactions-list">
              <table className="dashboard-table">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Stock</th>
                    <th>Type</th>
                    <th>Quantity</th>
                    <th>Price</th>
                    <th>Total</th>
                  </tr>
                </thead>
                <tbody>
                  {transactions && transactions.length > 0 ? transactions.map(transaction => transaction && transaction.stock ? (
                    <tr key={transaction.id}>
                      <td>{new Date(transaction.created_at).toLocaleDateString()}</td>
                      <td>{transaction.stock.symbol}</td>
                      <td className={`transaction-type ${transaction.transaction_type}`}>
                        {transaction.transaction_type}
                      </td>
                      <td>{transaction.quantity}</td>
                      <td>${(transaction.price || 0).toFixed(2)}</td>
                      <td>${(transaction.quantity * (transaction.price || 0)).toFixed(2)}</td>
                    </tr>
                  ) : null) : (
                    <tr>
                      <td colSpan="6" className="empty-message">No recent transactions.</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
          </main>
        </div>
      </div>

      {/* Chat Drawer */}
      <div className={`chat-drawer ${chatDrawerOpen ? 'open' : ''}`}>
        <div className="chat-drawer-header">
          <h3>💬 Chat</h3>
          <button 
            className="drawer-close-btn"
            onClick={toggleChatDrawer}
            aria-label="Close Chat"
          >
            ✕
          </button>
        </div>
        <div className="chat-drawer-content">
          <Chat />
        </div>
      </div>


      {/* Crisis Alert Manager */}
      <CrisisAlertManager socket={socket} />
    </div>
  );
};

export default Dashboard;