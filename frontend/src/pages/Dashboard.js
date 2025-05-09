import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getUserPortfolio, getTransactionHistory, getAllStocks } from '../services/stock';
import { initWebSocket, addListener, closeWebSocket } from '../services/websocket';
import Navigation from '../components/Navigation';
import './Dashboard.css';

const Dashboard = () => {
  const [portfolio, setPortfolio] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [topStocks, setTopStocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch portfolio, transactions, and stocks data in parallel
        const [portfolioData, transactionsData, stocksData] = await Promise.all([
          getUserPortfolio(),
          getTransactionHistory(5), // Get 5 most recent transactions
          getAllStocks()
        ]);
        
        setPortfolio(portfolioData);
        setTransactions(transactionsData);
        
        // Get top 5 stocks by price
        const sortedStocks = [...stocksData].sort((a, b) => b.current_price - a.current_price);
        setTopStocks(sortedStocks.slice(0, 5));
        
        setLoading(false);
      } catch (err) {
        setError('Failed to load dashboard data. Please try again later.');
        setLoading(false);
      }
    };

    fetchData();

    // Initialize WebSocket connection
    initWebSocket();

    // Listen for stock updates to refresh data
    const removeListener = addListener('*', (message) => {
      // Log the message for debugging
      console.log('Received message on dashboard:', message);

      // Process stock update message
      if (message.type === 'stock_update' || (message.id && message.current_price)) {
        // Extract stock_id and price - handle different message formats
        const stock_id = message.stock_id || message.id;
        const price = message.price || message.current_price;

        if (!stock_id || !price) {
          console.log('Missing required fields in message:', message);
          return;
        }

        // Update portfolio stocks if affected
        setPortfolio(prevPortfolio => {
          if (!prevPortfolio || !prevPortfolio.portfolio_items) return prevPortfolio;

          // Update portfolio items if the stock is in portfolio
          const updatedItems = prevPortfolio.portfolio_items.map(item => {
            if (item.stock_id === stock_id) {
              const oldValue = item.quantity * item.stock.current_price;
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

          // Recalculate stock value
          const newStockValue = updatedItems.reduce(
            (total, item) => total + (item.quantity * item.stock.current_price),
            0
          );

          return {
            ...prevPortfolio,
            portfolio_items: updatedItems,
            stock_value: newStockValue,
            total_value: prevPortfolio.cash_balance + newStockValue
          };
        });

        // Update top stocks if affected
        setTopStocks(prevTopStocks => {
          return prevTopStocks.map(stock => {
            if (stock.id === stock_id) {
              return {
                ...stock,
                current_price: price,
                priceChange: stock.current_price < price ? 'up' : 'down'
              };
            }
            return stock;
          });
        });
      }
    });

    // Clean up on unmount
    return () => {
      removeListener();
      closeWebSocket();
    };
  }, []);

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="dashboard-page">
      <Navigation />
      <div className="dashboard-container">
        <div className="dashboard-header">
          <h1>Dashboard</h1>
          <div className="portfolio-value">
            <h2>Total Portfolio Value</h2>
            <div className="value">${portfolio?.total_value.toFixed(2)}</div>
            <div className="portfolio-breakdown">
              <div className="breakdown-item">
                <span>Cash:</span>
                <span>${portfolio?.cash_balance.toFixed(2)}</span>
              </div>
              <div className="breakdown-item">
                <span>Stocks:</span>
                <span>${portfolio?.stock_value.toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
        
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
                    {portfolio.portfolio_items.slice(0, 3).map(item => (
                      <tr 
                        key={item.stock_id}
                        className={item.valueChange ? `value-${item.valueChange}` : ''}
                      >
                        <td>{item.stock.symbol}</td>
                        <td>{item.quantity}</td>
                        <td>${item.stock.current_price.toFixed(2)}</td>
                        <td>${(item.quantity * item.stock.current_price).toFixed(2)}</td>
                        <td>
                          <Link to={`/stock/${item.stock_id}`} className="trade-button">
                            Trade
                          </Link>
                        </td>
                      </tr>
                    ))}
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
                  {topStocks.map(stock => (
                    <tr 
                      key={stock.id}
                      className={stock.priceChange ? `price-${stock.priceChange}` : ''}
                    >
                      <td>{stock.symbol}</td>
                      <td>{stock.name}</td>
                      <td>${stock.current_price.toFixed(2)}</td>
                      <td>
                        <Link to={`/stock/${stock.id}`} className="trade-button">
                          Trade
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          
          <div className="dashboard-section">
            <div className="section-header">
              <h2>Recent Transactions</h2>
              <Link to="/transactions" className="view-all">View All</Link>
            </div>
            
            {transactions.length > 0 ? (
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
                    {transactions.map(transaction => (
                      <tr key={transaction.id}>
                        <td>{new Date(transaction.created_at).toLocaleDateString()}</td>
                        <td>{transaction.stock.symbol}</td>
                        <td className={`transaction-type ${transaction.transaction_type}`}>
                          {transaction.transaction_type}
                        </td>
                        <td>{transaction.quantity}</td>
                        <td>${transaction.price.toFixed(2)}</td>
                        <td>${(transaction.quantity * transaction.price).toFixed(2)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="empty-list">
                <p>No recent transactions.</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;