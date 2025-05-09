import React, { useState, useEffect } from 'react';
import Navigation from '../components/Navigation';
import { getTransactionHistory } from '../services/stock';
import './Transactions.css';

const Transactions = () => {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [typeFilter, setTypeFilter] = useState('all');
  const [limit, setLimit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  const fetchTransactions = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const data = await getTransactionHistory(limit, offset);
      
      setTransactions(data);
      setHasMore(data.length === limit);
      setLoading(false);
    } catch (err) {
      setError('Failed to load transaction history. Please try again later.');
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTransactions();
  }, [offset, limit]);

  const handleRefresh = () => {
    fetchTransactions();
  };

  const filteredTransactions = transactions.filter(transaction => {
    if (typeFilter === 'all') return true;
    return transaction.transaction_type === typeFilter;
  });

  const formatCurrency = (value) => {
    return `$${value.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,')}`;
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date);
  };

  const handlePrevPage = () => {
    if (offset - limit >= 0) {
      setOffset(offset - limit);
    }
  };

  const handleNextPage = () => {
    if (hasMore) {
      setOffset(offset + limit);
    }
  };

  const handleTypeFilterChange = (e) => {
    setTypeFilter(e.target.value);
  };

  return (
    <div className="transactions-page">
      <Navigation />
      <div className="transactions-container">
        <div className="transactions-header">
          <h1>Transaction History</h1>
          <button className="refresh-button" onClick={handleRefresh} disabled={loading}>
            <span>Refresh</span>
            <svg className="refresh-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="16" height="16">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
        
        <div className="filter-controls">
          <div className="filter-group">
            <label htmlFor="type-filter">Transaction Type:</label>
            <select
              id="type-filter"
              value={typeFilter}
              onChange={handleTypeFilterChange}
            >
              <option value="all">All Transactions</option>
              <option value="buy">Buy Orders</option>
              <option value="sell">Sell Orders</option>
            </select>
          </div>
          
          <div className="filter-group">
            <label htmlFor="limit-filter">Show:</label>
            <select
              id="limit-filter"
              value={limit}
              onChange={(e) => setLimit(parseInt(e.target.value))}
            >
              <option value="10">10 per page</option>
              <option value="20">20 per page</option>
              <option value="50">50 per page</option>
              <option value="100">100 per page</option>
            </select>
          </div>
        </div>
        
        {loading ? (
          <div className="loading">Loading transaction history...</div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : filteredTransactions.length === 0 ? (
          <div className="no-data">No transaction history available.</div>
        ) : (
          <>
            <table className="transactions-table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Stock</th>
                  <th>Type</th>
                  <th>Quantity</th>
                  <th className="value">Price</th>
                  <th className="value">Total</th>
                </tr>
              </thead>
              <tbody>
                {filteredTransactions.map(transaction => (
                  <tr key={transaction.id}>
                    <td>{formatDate(transaction.created_at)}</td>
                    <td className="stock-symbol">{transaction.stock.symbol}</td>
                    <td>
                      <span className={`transaction-type ${transaction.transaction_type}`}>
                        {transaction.transaction_type}
                      </span>
                    </td>
                    <td>{transaction.quantity}</td>
                    <td className="value">{formatCurrency(transaction.price)}</td>
                    <td className="value">{formatCurrency(transaction.quantity * transaction.price)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            
            <div className="pagination">
              <button 
                onClick={handlePrevPage} 
                disabled={offset === 0}
              >
                Previous
              </button>
              <button 
                onClick={handleNextPage} 
                disabled={!hasMore}
              >
                Next
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default Transactions;