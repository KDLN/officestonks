import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getAllStocks } from '../services/stock';
import { initWebSocket, addWebSocketListener, closeWebSocket } from '../services/websocket';
import { initSSE, addSSEListener, removeSSEListener } from '../services/sse';
import Navigation from '../components/Navigation';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import './StockList.css';

const StockList = () => {
  const [stocks, setStocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState('symbol');
  const [sortDirection, setSortDirection] = useState('asc');

  const fetchStocks = async () => {
    try {
      const stocksData = await getAllStocks();
      setStocks(stocksData);
      setLoading(false);
    } catch (err) {
      console.error('Error loading stocks:', err);
      const errorMessage = err.response?.data?.error || 'Unable to load stocks. Please check your connection and try again.';
      setError(errorMessage);
      setLoading(false);
    }
  };

  // Fetch stocks on component mount
  useEffect(() => {
    fetchStocks();

    // Initialize WebSocket connection for chat (keeping existing functionality)
    initWebSocket();

    // Initialize SSE connection for stock updates
    const sseTimeout = setTimeout(() => {
      console.log('📡 Initializing SSE for StockList updates...');
      initSSE().catch(err => {
        console.error('❌ Failed to initialize SSE:', err);
      });
    }, 750);

    // Listen for SSE stock price updates
    const handleStockUpdate = (message) => {
      setStocks(prevStocks => 
        prevStocks.map(stock => 
          stock.id === message.stock_id 
            ? { 
                ...stock, 
                current_price: message.price,
                priceChange: stock.current_price < message.price ? 'up' : 'down'
              } 
            : stock
        )
      );
    };

    addSSEListener('stockUpdate', handleStockUpdate);

    // Cleanup on unmount
    return () => {
      clearTimeout(sseTimeout);
      removeSSEListener('stockUpdate', handleStockUpdate);
      // Don't close WebSocket or SSE on component unmount as they're shared
    };
  }, []);

  // Handle sort change
  const handleSortChange = (field) => {
    if (sortBy === field) {
      // If already sorting by this field, toggle direction
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // If sorting by a new field, default to ascending
      setSortBy(field);
      setSortDirection('asc');
    }
  };

  // Filter and sort stocks
  const filteredAndSortedStocks = stocks
    .filter(stock => 
      stock.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
      stock.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      stock.sector.toLowerCase().includes(searchTerm.toLowerCase())
    )
    .sort((a, b) => {
      let compareA, compareB;
      
      // Determine values to compare based on sort field
      switch (sortBy) {
        case 'price':
          compareA = a.current_price;
          compareB = b.current_price;
          break;
        case 'name':
          compareA = a.name;
          compareB = b.name;
          break;
        case 'sector':
          compareA = a.sector;
          compareB = b.sector;
          break;
        default: // 'symbol'
          compareA = a.symbol;
          compareB = b.symbol;
      }
      
      // Compare values with respect to sort direction
      if (sortDirection === 'asc') {
        return compareA > compareB ? 1 : -1;
      } else {
        return compareA < compareB ? 1 : -1;
      }
    });

  if (loading) {
    return (
      <div className="stock-list-page">
        <Navigation />
        <LoadingSpinner message="Loading stocks..." />
      </div>
    );
  }

  if (error) {
    return (
      <div className="stock-list-page">
        <Navigation />
        <div className="stock-list-container">
          <ErrorMessage 
            message={error} 
            onRetry={() => {
              setError(null);
              setLoading(true);
              fetchStocks();
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="stock-list-page">
      <Navigation />
      <div className="stock-list-container">
        <h1>Available Stocks</h1>
        
        {/* Search and filter */}
        <div className="stock-list-controls">
          <input 
            type="text" 
            placeholder="Search stocks..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="stock-search"
          />
        </div>
        
        {/* Stocks table */}
        <table className="stock-table">
          <thead>
            <tr>
              <th onClick={() => handleSortChange('symbol')}>
                Symbol {sortBy === 'symbol' && (sortDirection === 'asc' ? '▲' : '▼')}
              </th>
              <th onClick={() => handleSortChange('name')}>
                Name {sortBy === 'name' && (sortDirection === 'asc' ? '▲' : '▼')}
              </th>
              <th onClick={() => handleSortChange('sector')}>
                Sector {sortBy === 'sector' && (sortDirection === 'asc' ? '▲' : '▼')}
              </th>
              <th onClick={() => handleSortChange('price')}>
                Price {sortBy === 'price' && (sortDirection === 'asc' ? '▲' : '▼')}
              </th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {filteredAndSortedStocks.map(stock => (
              <tr 
                key={stock.id} 
                className={stock.priceChange ? `price-${stock.priceChange}` : ''}
              >
                <td data-label="Symbol">{stock.symbol}</td>
                <td data-label="Name">{stock.name}</td>
                <td data-label="Sector">{stock.sector}</td>
                <td data-label="Price" className="price-cell">${stock.current_price.toFixed(2)}</td>
                <td data-label="Action">
                  <Link to={`/stock/${stock.id}`} className="trade-button">
                    Trade
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        
        {filteredAndSortedStocks.length === 0 && (
          <div className="no-results">No stocks found matching your search.</div>
        )}
      </div>
    </div>
  );
};

export default StockList;