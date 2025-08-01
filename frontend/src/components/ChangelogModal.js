import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './ChangelogModal.css';

const ChangelogModal = ({ manualTrigger, onManualClose }) => {
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

  // Handle manual trigger
  useEffect(() => {
    const fetchChangelogManual = async () => {
      try {
        setLoading(true);
        
        const token = localStorage.getItem('token');
        const response = await fetch('/api/changelog?limit=10', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });
        
        if (!response.ok) throw new Error('Failed to fetch changelog');
        
        const data = await response.json();
        const entries = data.entries || [];
        setChangelog(entries);
        
      } catch (error) {
        console.error('Error fetching changelog:', error);
        // Fallback content for testing
        setChangelog([{
          id: 1,
          version: 'v1.1.0',
          title: 'Market Sectors Foundation',
          description: 'Introduced market sectors with correlated stock movements for more realistic trading.',
          changes: [
            'Added 6 market sectors: Technology, Automotive, Financial Services, Retail, Entertainment, Healthcare',
            'Stock prices now influenced by both individual trends (70%) and sector trends (30%)',
            'Sector-wide correlations create realistic market behavior',
            'Enhanced market simulator with sector tracking',
            'Database schema updated to support sector relationships'
          ],
          change_type: 'feature',
          is_major: true,
          created_at: new Date().toISOString()
        }]);
      } finally {
        setLoading(false);
      }
    };

    if (manualTrigger && isAuthenticated) {
      console.log('Manual changelog trigger activated');
      setIsOpen(true); // Open modal immediately
      fetchChangelogManual(); // Fetch data after opening
    }
  }, [manualTrigger, isAuthenticated]);

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
      
      // Show modal if this is a new version or if no version is stored
      if (lastSeenVersion !== latestVersion) {
        console.log(`📰 Showing changelog modal: latest=${latestVersion}, lastSeen=${lastSeenVersion}`);
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
    if (Array.isArray(changelog) && changelog.length > 0) {
      // Mark the latest version as seen (only if not manually triggered)
      if (!manualTrigger) {
        localStorage.setItem('lastSeenChangelogVersion', changelog[0].version);
      }
    }
    setIsOpen(false);
    
    // Call the manual close callback if provided
    if (onManualClose) {
      onManualClose();
    }
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
  if (!isAuthenticated || !isOpen) return null;

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
          {loading ? (
            <div style={{ textAlign: 'center', padding: '2rem' }}>
              <div>Loading changelog...</div>
            </div>
          ) : !Array.isArray(changelog) || changelog.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2rem' }}>
              <div>No changelog entries available.</div>
              <div style={{ fontSize: '0.9rem', marginTop: '1rem', opacity: 0.7 }}>
                This might be due to API connectivity issues.
              </div>
            </div>
          ) : (
            changelog.map((entry, index) => (
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
              
              {entry.changes && Array.isArray(entry.changes) && entry.changes.length > 0 && (
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
            ))
          )}
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