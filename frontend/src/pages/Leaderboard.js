import React, { useState, useEffect } from 'react';
import Navigation from '../components/Navigation';
import { getLeaderboard } from '../services/user';
import { getUserId } from '../services/auth';
import './Leaderboard.css';

const Leaderboard = () => {
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const currentUserId = parseInt(getUserId());

  const fetchLeaderboard = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getLeaderboard(20); // Get top 20 users
      setLeaderboard(data);
      setLoading(false);
    } catch (err) {
      setError('Failed to load leaderboard data. Please try again later.');
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const formatCurrency = (value) => {
    return `$${value.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,')}`;
  };

  const handleRefresh = () => {
    fetchLeaderboard();
  };

  return (
    <div className="leaderboard-page">
      <Navigation />
      <div className="leaderboard-container">
        <div className="leaderboard-header">
          <h1>Investor Leaderboard</h1>
          <button className="refresh-button" onClick={handleRefresh} disabled={loading}>
            <span>Refresh</span>
            <svg className="refresh-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
        
        {loading ? (
          <div className="loading">Loading leaderboard data...</div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : leaderboard.length === 0 ? (
          <div className="no-data">No leaderboard data available.</div>
        ) : (
          <table className="leaderboard-table">
            <thead>
              <tr>
                <th>Rank</th>
                <th>Investor</th>
                <th className="value">Cash Balance</th>
                <th className="value">Stock Value</th>
                <th className="value">Total Value</th>
              </tr>
            </thead>
            <tbody>
              {leaderboard.map(entry => (
                <tr 
                  key={entry.user_id}
                  className={entry.user_id === currentUserId ? 'current-user' : ''}
                >
                  <td className={`rank rank-${entry.rank <= 3 ? entry.rank : ''}`}>
                    {entry.rank}
                  </td>
                  <td className="username">
                    {entry.username}
                    {entry.user_id === currentUserId ? ' (You)' : ''}
                  </td>
                  <td className="value">{formatCurrency(entry.cash_balance)}</td>
                  <td className="value">{formatCurrency(entry.stock_value)}</td>
                  <td className="value">{formatCurrency(entry.total_value)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default Leaderboard;