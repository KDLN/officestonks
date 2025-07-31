import React, { useState, useEffect } from 'react';
import { fetchActiveNews } from '../services/news';
import './NewsDisplay.css';

const NewsDisplay = () => {
  const [news, setNews] = useState([]);
  const [filteredNews, setFilteredNews] = useState([]);
  const [isVisible, setIsVisible] = useState(() => {
    const saved = localStorage.getItem('newsVisible');
    return saved !== null ? JSON.parse(saved) : true;
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState('all'); // all, crisis, bankruptcy, recovery, sector

  // Save visibility preference to localStorage
  const toggleVisibility = () => {
    const newVisibility = !isVisible;
    setIsVisible(newVisibility);
    localStorage.setItem('newsVisible', JSON.stringify(newVisibility));
  };

  useEffect(() => {
    const fetchNews = async () => {
      try {
        setLoading(true);
        const activeNews = await fetchActiveNews();
        
        // Filter out expired news
        const now = new Date();
        const validNews = (activeNews || []).filter(item => {
          if (!item.expires_at) return true; // No expiration date means always show
          const expirationDate = new Date(item.expires_at);
          return expirationDate > now;
        });
        
        setNews(validNews);
        setError(null);
      } catch (err) {
        console.error('Error fetching news:', err);
        setError('Failed to load news');
        setNews([]);
      } finally {
        setLoading(false);
      }
    };

    fetchNews();
  }, []);

  // Filter news based on selected filter
  useEffect(() => {
    if (filter === 'all') {
      setFilteredNews(news);
    } else {
      const filtered = news.filter(item => item.type === filter);
      setFilteredNews(filtered);
    }
  }, [news, filter]);

  const getNewsTypeIcon = (type) => {
    switch (type) {
      case 'crisis': return '🚨';
      case 'bankruptcy': return '💀';
      case 'recovery': return '🚀';
      case 'sector': return '🏭';
      case 'market': return '📈';
      default: return '📰';
    }
  };

  const getNewsTypeColor = (type) => {
    switch (type) {
      case 'crisis': return '#ff4444';
      case 'bankruptcy': return '#cc0000';
      case 'recovery': return '#00cc44';
      case 'sector': return '#4488ff';
      case 'market': return '#888888';
      default: return '#333333';
    }
  };

  if (loading) {
    return (
      <div className="news-display loading">
        <div className="news-header">
          <h3>📰 Office Stonks News</h3>
        </div>
        <p>Loading news...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="news-display error">
        <div className="news-header">
          <h3>📰 Office Stonks News</h3>
        </div>
        <p className="error-message">{error}</p>
      </div>
    );
  }

  if (!news || news.length === 0) {
    return (
      <div className="news-display empty">
        <div className="news-header">
          <h3>📰 Office Stonks News</h3>
        </div>
        <p className="no-news">No active news at this time.</p>
      </div>
    );
  }

  return (
    <div className="news-display">
      <div className="news-header">
        <h3>📰 Office Stonks News</h3>
        <button 
          className="toggle-button"
          onClick={toggleVisibility}
          aria-label={isVisible ? "Hide news" : "Show news"}
        >
          {isVisible ? '🔼' : '🔽'}
        </button>
      </div>
      
      {isVisible && (
        <div className="news-content">
          <div className="news-filters">
            <button 
              className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
              onClick={() => setFilter('all')}
            >
              All
            </button>
            <button 
              className={`filter-btn ${filter === 'crisis' ? 'active' : ''}`}
              onClick={() => setFilter('crisis')}
            >
              🚨 Crisis
            </button>
            <button 
              className={`filter-btn ${filter === 'bankruptcy' ? 'active' : ''}`}
              onClick={() => setFilter('bankruptcy')}
            >
              💀 Bankruptcy
            </button>
            <button 
              className={`filter-btn ${filter === 'recovery' ? 'active' : ''}`}
              onClick={() => setFilter('recovery')}
            >
              🚀 Recovery
            </button>
            <button 
              className={`filter-btn ${filter === 'sector' ? 'active' : ''}`}
              onClick={() => setFilter('sector')}
            >
              🏭 Sector
            </button>
          </div>
          
          <div className="news-items">
            {filteredNews.map((item, index) => (
              <div 
                key={item.id || index} 
                className={`news-item news-${item.type || 'default'}`}
                style={{ borderLeftColor: getNewsTypeColor(item.type) }}
              >
                <div className="news-item-header">
                  <span className="news-type-icon">{getNewsTypeIcon(item.type)}</span>
                  <h4 className="news-title">{item.title}</h4>
                  {item.stock_symbol && (
                    <span className="news-stock-symbol">{item.stock_symbol}</span>
                  )}
                </div>
                <p className="news-text">{item.content}</p>
                <div className="news-meta">
                  <span className="news-time">
                    {new Date(item.created_at).toLocaleString()}
                  </span>
                  {item.expires_at && (
                    <span className="expiry-info">
                      Expires: {new Date(item.expires_at).toLocaleDateString()}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
          
          {filteredNews.length === 0 && (
            <p className="no-news">No {filter === 'all' ? '' : filter} news at this time.</p>
          )}
        </div>
      )}
    </div>
  );
};

export default NewsDisplay;