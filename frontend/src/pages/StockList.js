import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { getAllStocks } from '../services/stock';
import { initWebSocket, addListener, closeWebSocket } from '../services/websocket';
import Navigation from '../components/Navigation';
import './StockList.css';

const StockList = () => {
  const [stocks, setStocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortBy, setSortBy] = useState('symbol');
  const [sortDirection, setSortDirection] = useState('asc');

  // Fetch stocks on component mount
  useEffect(() => {
    const fetchStocks = async () => {
      try {
        const stocksData = await getAllStocks();
        setStocks(stocksData);
        setLoading(false);
      } catch (err) {
        setError('Failed to load stocks. Please try again later.');
        setLoading(false);
      }
    };

    fetchStocks();

    // Initialize WebSocket connection
    initWebSocket();

    // Listen for stock price updates
    const removeListener = addListener('stock_update', (message) => {
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
    });

    // Cleanup on unmount
    return () => {
      removeListener();
      closeWebSocket();
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
    return <div className="loading">Loading stocks...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
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
                <td>{stock.symbol}</td>
                <td>{stock.name}</td>
                <td>{stock.sector}</td>
                <td className="price-cell">${stock.current_price.toFixed(2)}</td>
                <td>
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