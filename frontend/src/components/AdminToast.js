import React, { useState, useEffect } from 'react';
import './AdminToast.css';

const AdminToast = ({ message, onClose, duration = 10000 }) => {
  const [isVisible, setIsVisible] = useState(true);
  const [isClosing, setIsClosing] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      handleClose();
    }, duration);

    return () => clearTimeout(timer);
  }, [duration]);

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsVisible(false);
      onClose();
    }, 300); // Match CSS animation duration
  };

  if (!isVisible) return null;

  return (
    <div className={`admin-toast ${isClosing ? 'closing' : ''}`}>
      <div className="admin-toast-content">
        <div className="admin-toast-icon">📢</div>
        <div className="admin-toast-message">
          <div className="admin-toast-title">Admin Announcement</div>
          <div className="admin-toast-text">{message}</div>
        </div>
        <button 
          className="admin-toast-close"
          onClick={handleClose}
          aria-label="Close announcement"
        >
          ✕
        </button>
      </div>
    </div>
  );
};

export default AdminToast;