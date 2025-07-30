import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './ChangelogModal.css';

const ChangelogModal = () => {
  const { user, isAuthenticated } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const [changelog, setChangelog] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Only check for changes if user is authenticated
    if (isAuthenticated && user) {
      checkForNewChanges();
    }
  }, [isAuthenticated, user]);

  const checkForNewChanges = async () => {
    try {
      setLoading(true);
      
      // Fetch the latest changelog entries with auth token
      const token = localStorage.getItem('token');
      const response = await fetch('/api/changelog?limit=5', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });
      if (!response.ok) throw new Error('Failed to fetch changelog');
      
      const data = await response.json();
      const entries = data.entries || [];
      
      if (entries.length === 0) return;
      
      // Get the latest version
      const latestVersion = entries[0].version;
      
      // Check if user has seen this version
      const lastSeenVersion = localStorage.getItem('lastSeenChangelogVersion');
      
      // Show modal if this is a new version
      if (lastSeenVersion !== latestVersion) {
        setChangelog(entries);
        setIsOpen(true);
      }
      
    } catch (error) {
      console.error('Error checking changelog:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (changelog.length > 0) {
      // Mark the latest version as seen
      localStorage.setItem('lastSeenChangelogVersion', changelog[0].version);
    }
    setIsOpen(false);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const getChangeTypeIcon = (changeType) => {
    switch (changeType) {
      case 'feature': return '🆕';
      case 'improvement': return '⚡';
      case 'bugfix': return '🐛';
      case 'breaking': return '⚠️';
      default: return '📝';
    }
  };

  const getChangeTypeColor = (changeType) => {
    switch (changeType) {
      case 'feature': return '#10B981'; // Green
      case 'improvement': return '#3B82F6'; // Blue
      case 'bugfix': return '#F59E0B'; // Orange
      case 'breaking': return '#EF4444'; // Red
      default: return '#6B7280'; // Gray
    }
  };

  // Don't render anything if user is not authenticated or modal is closed
  if (!isAuthenticated || !isOpen || changelog.length === 0) return null;

  return (
    <div className="changelog-modal-overlay">
      <div className="changelog-modal">
        <div className="changelog-modal-header">
          <h2>🎉 What's New in Office Stonks</h2>
          <button 
            className="changelog-modal-close"
            onClick={handleClose}
            aria-label="Close changelog"
          >
            ×
          </button>
        </div>
        
        <div className="changelog-modal-content">
          {changelog.map((entry, index) => (
            <div key={entry.id} className="changelog-entry">
              <div className="changelog-entry-header">
                <div className="changelog-version-badge">
                  <span 
                    className="changelog-type-icon"
                    style={{ color: getChangeTypeColor(entry.change_type) }}
                  >
                    {getChangeTypeIcon(entry.change_type)}
                  </span>
                  <span className="changelog-version">{entry.version}</span>
                  {entry.is_major && <span className="changelog-major-badge">Major</span>}
                </div>
                <span className="changelog-date">
                  {formatDate(entry.created_at)}
                </span>
              </div>
              
              <h3 className="changelog-title">{entry.title}</h3>
              
              {entry.description && (
                <p className="changelog-description">{entry.description}</p>
              )}
              
              {entry.changes && entry.changes.length > 0 && (
                <ul className="changelog-changes">
                  {entry.changes.map((change, changeIndex) => (
                    <li key={changeIndex} className="changelog-change-item">
                      {change}
                    </li>
                  ))}
                </ul>
              )}
              
              {index < changelog.length - 1 && <hr className="changelog-separator" />}
            </div>
          ))}
        </div>
        
        <div className="changelog-modal-footer">
          <p className="changelog-footer-text">
            Stay updated with all the latest features and improvements!
          </p>
          <button 
            className="changelog-modal-button"
            onClick={handleClose}
          >
            Got it, thanks! 👍
          </button>
        </div>
      </div>
    </div>
  );
};

export default ChangelogModal;