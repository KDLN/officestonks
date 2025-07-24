import React from 'react';
import './TradeConfirmationModal.css';

const TradeConfirmationModal = ({ 
  isOpen, 
  onClose, 
  onConfirm, 
  stock, 
  action, 
  quantity, 
  totalCost,
  loading 
}) => {
  if (!isOpen) return null;

  const handleConfirm = () => {
    onConfirm();
  };

  const handleCancel = () => {
    if (!loading) {
      onClose();
    }
  };

  return (
    <div className="modal-overlay" onClick={handleCancel}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Confirm Trade</h2>
          <button 
            className="close-button" 
            onClick={handleCancel}
            disabled={loading}
          >
            ×
          </button>
        </div>
        
        <div className="modal-body">
          <div className="trade-summary">
            <div className="summary-row">
              <span className="label">Action:</span>
              <span className={`value ${action}-action`}>
                {action.toUpperCase()}
              </span>
            </div>
            
            <div className="summary-row">
              <span className="label">Stock:</span>
              <span className="value">
                {stock.symbol} - {stock.name}
              </span>
            </div>
            
            <div className="summary-row">
              <span className="label">Quantity:</span>
              <span className="value">{quantity} shares</span>
            </div>
            
            <div className="summary-row">
              <span className="label">Price per share:</span>
              <span className="value">${stock.current_price.toFixed(2)}</span>
            </div>
            
            <div className="summary-row total-row">
              <span className="label">
                Total {action === 'buy' ? 'Cost' : 'Proceeds'}:
              </span>
              <span className="value total-value">
                ${totalCost.toFixed(2)}
              </span>
            </div>
          </div>
          
          <div className="confirmation-message">
            Are you sure you want to {action} {quantity} shares of {stock.symbol} for ${totalCost.toFixed(2)}?
          </div>
        </div>
        
        <div className="modal-footer">
          <button 
            className="cancel-button"
            onClick={handleCancel}
            disabled={loading}
          >
            Cancel
          </button>
          <button 
            className={`confirm-button ${action}-confirm`}
            onClick={handleConfirm}
            disabled={loading}
          >
            {loading ? 'Processing...' : `Confirm ${action.toUpperCase()}`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default TradeConfirmationModal;