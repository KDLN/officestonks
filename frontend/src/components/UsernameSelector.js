import React, { useState, useEffect } from 'react';
import { checkUsernameAvailability, setUsername, validateUsername } from '../services/username';
import './UsernameSelector.css';

const UsernameSelector = ({ currentUsername, onUsernameSet, onCancel }) => {
  const [username, setUsernameInput] = useState(currentUsername || '');
  const [isChecking, setIsChecking] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [availability, setAvailability] = useState(null);
  const [validationError, setValidationError] = useState('');

  useEffect(() => {
    // Clear previous checks when username changes
    setAvailability(null);
    setValidationError('');

    if (!username) {
      return;
    }

    // Validate format first
    const validation = validateUsername(username);
    if (!validation.valid) {
      setValidationError(validation.error);
      return;
    }

    // Don't check if it's the current username
    if (username === currentUsername) {
      setAvailability({ available: true });
      return;
    }

    // Debounce username availability check
    const timeoutId = setTimeout(async () => {
      setIsChecking(true);
      try {
        const result = await checkUsernameAvailability(username);
        setAvailability(result);
      } catch (error) {
        setAvailability({ available: false, error: 'Error checking username' });
      } finally {
        setIsChecking(false);
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [username, currentUsername]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!availability?.available || validationError) {
      return;
    }

    setIsSubmitting(true);
    try {
      await setUsername(username);
      onUsernameSet(username);
    } catch (error) {
      setAvailability({ available: false, error: error.message });
    } finally {
      setIsSubmitting(false);
    }
  };

  const getStatusIcon = () => {
    if (isChecking) {
      return <span className="status-checking">⏳</span>;
    }
    
    if (validationError) {
      return <span className="status-error">❌</span>;
    }
    
    if (availability?.available) {
      return <span className="status-available">✅</span>;
    }
    
    if (availability?.error) {
      return <span className="status-error">❌</span>;
    }
    
    return null;
  };

  const getStatusMessage = () => {
    if (validationError) {
      return validationError;
    }
    
    if (availability?.error) {
      return availability.error;
    }
    
    if (availability?.available && username !== currentUsername) {
      return 'Username is available!';
    }
    
    return '';
  };

  const canSubmit = availability?.available && !validationError && !isChecking && username !== currentUsername;

  return (
    <div className="username-selector">
      <h3>Choose Your Username</h3>
      <p>Pick a unique username for Office Stonks (3-20 characters, letters, numbers, and underscores only)</p>
      
      <form onSubmit={handleSubmit}>
        <div className="username-input-group">
          <input
            type="text"
            value={username}
            onChange={(e) => setUsernameInput(e.target.value)}
            placeholder="Enter username..."
            className="username-input"
            autoFocus
          />
          {getStatusIcon()}
        </div>
        
        {getStatusMessage() && (
          <p className={`status-message ${availability?.available ? 'success' : 'error'}`}>
            {getStatusMessage()}
          </p>
        )}
        
        <div className="username-actions">
          <button
            type="submit"
            disabled={!canSubmit || isSubmitting}
            className="btn-primary"
          >
            {isSubmitting ? 'Setting Username...' : 'Set Username'}
          </button>
          
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="btn-secondary"
            >
              Cancel
            </button>
          )}
        </div>
      </form>
    </div>
  );
};

export default UsernameSelector;