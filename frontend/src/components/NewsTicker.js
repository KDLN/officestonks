import React, { useState, useEffect } from 'react';
import { fetchActiveNews } from '../services/news';
import './NewsTicker.css';

const NewsTicker = () => {
  const [tickerNews, setTickerNews] = useState([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(true);

  useEffect(() => {
    const fetchTickerNews = async () => {
      try {
        const news = await fetchActiveNews();
        // Filter to show only crisis, bankruptcy, and recovery news for ticker
        const tickerItems = (news || []).filter(item => 
          ['crisis', 'bankruptcy', 'recovery'].includes(item.type)
        );
        
        setTickerNews(tickerItems);
      } catch (err) {
        console.error('Error fetching ticker news:', err);
        setTickerNews([]);
      }
    };

    fetchTickerNews();
    // Refresh ticker news every 30 seconds
    const interval = setInterval(fetchTickerNews, 30000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    if (tickerNews.length === 0 || !isPlaying) return;

    const interval = setInterval(() => {
      setCurrentIndex((prevIndex) => 
        prevIndex >= tickerNews.length - 1 ? 0 : prevIndex + 1
      );
    }, 4000); // Change news every 4 seconds

    return () => clearInterval(interval);
  }, [tickerNews.length, isPlaying]);

  const getTickerIcon = (type) => {
    switch (type) {
      case 'crisis': return '🚨';
      case 'bankruptcy': return '💀';
      case 'recovery': return '🚀';
      default: return '📰';
    }
  };

  const getTickerColor = (type) => {
    switch (type) {
      case 'crisis': return '#ff4444';
      case 'bankruptcy': return '#cc0000';
      case 'recovery': return '#00cc44';
      default: return '#333333';
    }
  };

  if (tickerNews.length === 0) {
    return null; // Don't show ticker if no breaking news
  }

  const currentNews = tickerNews[currentIndex];

  return (
    <div className="news-ticker">
      <div className="ticker-label">
        <span className="breaking-text">BREAKING</span>
      </div>
      
      <div className="ticker-content">
        <div 
          className="ticker-item"
          style={{ color: getTickerColor(currentNews.type) }}
        >
          <span className="ticker-icon">{getTickerIcon(currentNews.type)}</span>
          <span className="ticker-text">
            {currentNews.stock_symbol && (
              <strong>{currentNews.stock_symbol}:</strong>
            )} {currentNews.title}
          </span>
        </div>
      </div>

      <div className="ticker-controls">
        <button 
          className="ticker-control-btn"
          onClick={() => setIsPlaying(!isPlaying)}
          title={isPlaying ? 'Pause ticker' : 'Play ticker'}
        >
          {isPlaying ? '⏸️' : '▶️'}
        </button>
        
        <span className="ticker-counter">
          {currentIndex + 1}/{tickerNews.length}
        </span>
      </div>
    </div>
  );
};

export default NewsTicker;