import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getUserPortfolio, getTransactionHistory } from '../services/stock';
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

  useEffect(() => {
    loadPortfolioData();
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
            <div className="summary-card">
              <h3>Total Value</h3>
              <div className="value">{formatCurrency(portfolio?.total_value || 0)}</div>
            </div>
            
            <div className="summary-card">
              <h3>Cash Balance</h3>
              <div className="value">{formatCurrency(portfolio?.cash_balance || 0)}</div>
            </div>
            
            <div className="summary-card">
              <h3>Stock Value</h3>
              <div className="value">{formatCurrency(portfolio?.stock_value || 0)}</div>
            </div>
            
            <div className="summary-card">
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

                      return (
                        <tr key={item.id}>
                          <td>
                            <div className="stock-info">
                              <span className="stock-symbol">{item.stock.symbol}</span>
                              <span className="stock-name">{item.stock.name}</span>
                            </div>
                          </td>
                          <td>{item.quantity}</td>
                          <td>{formatCurrency(item.stock.current_price)}</td>
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