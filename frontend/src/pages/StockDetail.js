import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getStockById, executeTrade, getUserPortfolio } from '../services/stock';
import { initWebSocket, addWebSocketListener, closeWebSocket } from '../services/websocket';
import { initSSE, addSSEListener, removeSSEListener } from '../services/sse';
import Navigation from '../components/Navigation';
import TradeConfirmationModal from '../components/TradeConfirmationModal';
import LoadingSpinner from '../components/LoadingSpinner';
import './StockDetail.css';

const StockDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const stockId = parseInt(id);

  const [stock, setStock] = useState(null);
  const [portfolio, setPortfolio] = useState(null);
  const [quantity, setQuantity] = useState(1);
  const [action, setAction] = useState('buy');
  const [loading, setLoading] = useState(true);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [priceChange, setPriceChange] = useState(null);
  const [showConfirmModal, setShowConfirmModal] = useState(false);

  // Fetch stock and portfolio data
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [stockData, portfolioData] = await Promise.all([
          getStockById(stockId),
          getUserPortfolio()
        ]);
        
        setStock(stockData);
        setPortfolio(portfolioData);
        setLoading(false);
      } catch (err) {
        setError('Failed to load stock data. Please try again later.');
        setLoading(false);
      }
    };

    fetchData();

    // Initialize WebSocket connection for chat (keeping existing functionality)
    initWebSocket();

    // Initialize SSE connection for stock updates
    const sseTimeout = setTimeout(() => {
      console.log('📡 Initializing SSE for StockDetail updates...');
      initSSE().catch(err => {
        console.error('❌ Failed to initialize SSE:', err);
      });
    }, 750);

    // Listen for SSE stock price updates
    const handleStockUpdate = (message) => {
      if (message.stock_id === stockId) {
        setStock(prevStock => {
          if (!prevStock) return null;
          
          // Determine price change direction
          const direction = prevStock.current_price < message.price ? 'up' : 'down';
          setPriceChange(direction);
          
          // Clear price change indicator after animation
          setTimeout(() => setPriceChange(null), 2000);
          
          return {
            ...prevStock,
            current_price: message.price
          };
        });
      }
    };

    addSSEListener('stockUpdate', handleStockUpdate);

    // Cleanup on unmount
    return () => {
      clearTimeout(sseTimeout);
      removeSSEListener('stockUpdate', handleStockUpdate);
      // Don't close WebSocket or SSE on component unmount as they're shared
    };
  }, [stockId]);

  // Calculate max quantity for sell action
  const maxSellQuantity = portfolio?.portfolio_items
    ?.find(item => item.stock_id === stockId)?.quantity || 0;

  // Calculate max buy quantity based on cash balance
  const maxBuyQuantity = stock && portfolio
    ? Math.floor(portfolio.cash_balance / stock.current_price)
    : 0;

  // Update quantity if current value exceeds max
  useEffect(() => {
    if (action === 'buy' && quantity > maxBuyQuantity && maxBuyQuantity > 0) {
      setQuantity(maxBuyQuantity);
    } else if (action === 'sell' && quantity > maxSellQuantity && maxSellQuantity > 0) {
      setQuantity(maxSellQuantity);
    }
  }, [action, maxBuyQuantity, maxSellQuantity, quantity]);

  // Handle trade form submission (show confirmation modal)
  const handleTrade = (e) => {
    e.preventDefault();
    
    if (quantity <= 0) {
      setError('Quantity must be greater than zero');
      return;
    }

    if (action === 'sell' && quantity > maxSellQuantity) {
      setError('You do not own enough shares to sell this quantity');
      return;
    }

    if (action === 'buy') {
      const totalCost = stock.current_price * quantity;
      if (totalCost > portfolio.cash_balance) {
        setError('You do not have enough cash for this purchase');
        return;
      }
    }
    
    setError(null);
    setShowConfirmModal(true);
  };

  // Execute the actual trade after confirmation
  const executeConfirmedTrade = async () => {
    setSuccess(null);
    setExecuting(true);
    
    try {
      await executeTrade(stockId, quantity, action);
      
      // Refresh portfolio data after successful trade
      const portfolioData = await getUserPortfolio();
      setPortfolio(portfolioData);
      
      setSuccess(`Successfully ${action === 'buy' ? 'bought' : 'sold'} ${quantity} shares of ${stock.symbol}`);
      setExecuting(false);
      setShowConfirmModal(false);
    } catch (err) {
      console.error('Trade execution error:', err);
      const errorMessage = err.response?.data?.error || err.message || 'Unable to execute trade. Please check your connection and try again.';
      setError(errorMessage);
      setExecuting(false);
      setShowConfirmModal(false);
    }
  };

  if (loading) {
    return (
      <div className="stock-detail-page">
        <Navigation />
        <LoadingSpinner message="Loading stock data..." />
      </div>
    );
  }

  if (!stock) {
    return <div className="error">Stock not found</div>;
  }

  // Calculate total cost for the current transaction
  const totalCost = stock.current_price * quantity;

  // Find if the user owns this stock
  const ownedStock = portfolio?.portfolio_items?.find(item => item.stock_id === stockId);

  return (
    <div className="stock-detail-page">
      <Navigation />
      <div className="stock-detail-container">
        <div className="stock-header">
          <h1>
            {stock.symbol} - {stock.name}
            <span className="sector-tag">{stock.sector}</span>
          </h1>
          <div className={`stock-price ${priceChange ? `price-${priceChange}` : ''}`}>
            ${stock.current_price.toFixed(2)}
          </div>
        </div>
        
        <div className="trade-container">
          <div className="user-portfolio-summary">
            <h2>Your Portfolio</h2>
            <p className="cash-balance">Cash: <b>${portfolio?.cash_balance.toFixed(2)}</b></p>
            
            {ownedStock && (
              <div className="owned-stock">
                <p>You own: <b>{ownedStock.quantity} shares</b></p>
                <p>Value: <b>${(ownedStock.quantity * stock.current_price).toFixed(2)}</b></p>
              </div>
            )}
          </div>
          
          <div className="trade-form-container">
            <h2>Trade {stock.symbol}</h2>
            {error && <div className="error-message">{error}</div>}
            {success && <div className="success-message">{success}</div>}
            
            <form onSubmit={handleTrade} className="trade-form">
              <div className="form-group">
                <label htmlFor="action">Action</label>
                <select 
                  id="action" 
                  value={action} 
                  onChange={(e) => setAction(e.target.value)}
                  disabled={executing}
                >
                  <option value="buy">Buy</option>
                  <option value="sell" disabled={!maxSellQuantity}>Sell</option>
                </select>
              </div>
              
              <div className="form-group">
                <label htmlFor="quantity">Quantity</label>
                <input 
                  id="quantity" 
                  type="number" 
                  min="1" 
                  max={action === 'buy' ? maxBuyQuantity : maxSellQuantity}
                  value={quantity} 
                  onChange={(e) => setQuantity(parseInt(e.target.value) || 1)}
                  disabled={executing}
                />
                <span className="max-quantity">
                  Max: {action === 'buy' ? maxBuyQuantity : maxSellQuantity}
                </span>
              </div>
              
              <div className="form-group">
                <label>Total {action === 'buy' ? 'Cost' : 'Proceeds'}</label>
                <div className="total-cost">${totalCost.toFixed(2)}</div>
              </div>
              
              <button 
                type="submit" 
                className={`trade-button ${action}-button`}
                disabled={
                  executing || 
                  (action === 'buy' && maxBuyQuantity === 0) || 
                  (action === 'sell' && maxSellQuantity === 0)
                }
              >
                {executing ? 'Processing...' : `${action === 'buy' ? 'Buy' : 'Sell'} ${stock.symbol}`}
              </button>
            </form>
          </div>
        </div>
        
        <div className="action-buttons">
          <button onClick={() => navigate('/stocks')} className="back-button">
            Back to Stocks
          </button>
          <button onClick={() => navigate('/portfolio')} className="portfolio-button">
            View Portfolio
          </button>
        </div>
      </div>

      <TradeConfirmationModal
        isOpen={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        onConfirm={executeConfirmedTrade}
        stock={stock}
        action={action}
        quantity={quantity}
        totalCost={totalCost}
        loading={executing}
      />
    </div>
  );
};

export default StockDetail;