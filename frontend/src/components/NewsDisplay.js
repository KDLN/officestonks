import React, { useState, useEffect } from 'react';
import { fetchActiveNews } from '../services/news';
import './NewsDisplay.css';

const NewsDisplay = () => {
  const [news, setNews] = useState([]);
  const [isVisible, setIsVisible] = useState(() => {
    const saved = localStorage.getItem('newsVisible');
    return saved !== null ? JSON.parse(saved) : true;
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

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
          {news.map((item, index) => (
            <div key={item.id || index} className="news-item">
              <h4 className="news-title">{item.title}</h4>
              <p className="news-text">{item.content}</p>
              {item.expires_at && (
                <div className="news-meta">
                  <span className="expiry-info">
                    Expires: {new Date(item.expires_at).toLocaleDateString()}
                  </span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default NewsDisplay;