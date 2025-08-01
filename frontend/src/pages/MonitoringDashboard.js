import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './MonitoringDashboard.css';

const MonitoringDashboard = () => {
  const { user } = useAuth();
  const [dashboardData, setDashboardData] = useState(null);
  const [activeSessions, setActiveSessions] = useState([]);
  const [recentActivity, setRecentActivity] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [userDetails, setUserDetails] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const intervalRef = useRef(null);

  // Redirect if not admin
  useEffect(() => {
    if (user && !user.is_admin) {
      window.location.href = '/dashboard';
    }
  }, [user]);

  // Fetch dashboard data
  const fetchDashboardData = async () => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch('/api/admin/monitoring/dashboard', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch dashboard data');
      }

      const data = await response.json();
      setDashboardData(data);
      setActiveSessions(data.active_sessions?.sessions || []);
      setRecentActivity(data.recent_activity?.activities || []);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Fetch user details
  const fetchUserDetails = async (userId) => {
    try {
      const token = localStorage.getItem('token');
      const response = await fetch(`/api/admin/monitoring/user-activity?user_id=${userId}&limit=20`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user details');
      }

      const data = await response.json();
      setUserDetails(data);
    } catch (err) {
      console.error('Error fetching user details:', err);
    }
  };

  // Auto-refresh functionality
  useEffect(() => {
    fetchDashboardData();

    if (autoRefresh) {
      intervalRef.current = setInterval(fetchDashboardData, 30000); // Refresh every 30 seconds
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [autoRefresh]);

  // Handle user selection
  const handleUserSelect = (userId) => {
    setSelectedUser(userId);
    fetchUserDetails(userId);
  };

  // Format time
  const formatTime = (timestamp) => {
    return new Date(timestamp).toLocaleString();
  };

  // Format duration
  const formatDuration = (start, end) => {
    if (!end) return 'Active';
    const duration = new Date(end) - new Date(start);
    const hours = Math.floor(duration / (1000 * 60 * 60));
    const minutes = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60));
    return `${hours}h ${minutes}m`;
  };

  // Get health status color
  const getHealthColor = (status) => {
    switch (status) {
      case 'healthy': return '#28a745';
      case 'degraded': return '#ffc107';
      case 'down': return '#dc3545';
      default: return '#6c757d';
    }
  };

  // Get activity status color
  const getActivityColor = (success) => {
    return success ? '#28a745' : '#dc3545';
  };

  if (!user?.is_admin) {
    return <div className="monitoring-access-denied">Access denied. Admin privileges required.</div>;
  }

  if (loading) {
    return <div className="monitoring-loading">Loading monitoring dashboard...</div>;
  }

  if (error) {
    return (
      <div className="monitoring-error">
        <h3>Error loading dashboard</h3>
        <p>{error}</p>
        <button onClick={fetchDashboardData}>Retry</button>
      </div>
    );
  }

  const metrics = dashboardData?.system_metrics || {};
  const hourlyStats = dashboardData?.hourly_stats || {};
  const healthStatus = dashboardData?.health_status || {};

  return (
    <div className="monitoring-dashboard">
      <div className="monitoring-header">
        <h1>System Monitoring Dashboard</h1>
        <div className="monitoring-controls">
          <label className="auto-refresh-toggle">
            <input 
              type="checkbox" 
              checked={autoRefresh} 
              onChange={(e) => setAutoRefresh(e.target.checked)}
            />
            Auto-refresh (30s)
          </label>
          <button onClick={fetchDashboardData} className="refresh-btn">
            🔄 Refresh Now
          </button>
        </div>
      </div>

      {/* System Metrics Overview */}
      <div className="metrics-overview">
        <div className="metric-card">
          <h3>Active Users</h3>
          <div className="metric-value">{metrics.active_users || 0}</div>
        </div>
        <div className="metric-card">
          <h3>Active Sessions</h3>
          <div className="metric-value">{metrics.active_sessions || 0}</div>
        </div>
        <div className="metric-card">
          <h3>Trades/Hour</h3>
          <div className="metric-value">{metrics.trades_per_hour || 0}</div>
        </div>
        <div className="metric-card">
          <h3>WebSocket Conns</h3>
          <div className="metric-value">{metrics.websocket_connections || 0}</div>
        </div>
      </div>

      {/* Health Status */}
      <div className="health-status">
        <h2>System Health</h2>
        <div className="health-indicators">
          <div className="health-indicator">
            <span className="health-label">Database:</span>
            <span 
              className="health-value" 
              style={{ color: getHealthColor(healthStatus.database) }}
            >
              {healthStatus.database || 'unknown'}
            </span>
          </div>
          <div className="health-indicator">
            <span className="health-label">Error Rate:</span>
            <span className="health-value">
              {(healthStatus.error_rate || 0).toFixed(2)}%
            </span>
          </div>
          <div className="health-indicator">
            <span className="health-label">Avg Response:</span>
            <span className="health-value">
              {(healthStatus.avg_response_ms || 0).toFixed(1)}ms
            </span>
          </div>
        </div>
      </div>

      {/* Hourly Activity Summary */}
      <div className="hourly-stats">
        <h2>Last Hour Activity</h2>
        <div className="hourly-metrics">
          <div className="hourly-metric">
            <span className="hourly-label">Logins:</span>
            <span className="hourly-value">{hourlyStats.logins || 0}</span>
          </div>
          <div className="hourly-metric">
            <span className="hourly-label">Trades:</span>
            <span className="hourly-value">{hourlyStats.trades || 0}</span>
          </div>
          <div className="hourly-metric">
            <span className="hourly-label">Errors:</span>
            <span className="hourly-value error">{hourlyStats.errors || 0}</span>
          </div>
        </div>
      </div>

      <div className="monitoring-content">
        {/* Active Sessions */}
        <div className="monitoring-section">
          <h2>Active Sessions ({activeSessions.length})</h2>
          <div className="sessions-table">
            {activeSessions.length === 0 ? (
              <p>No active sessions</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>User</th>
                    <th>IP Address</th>
                    <th>Login Time</th>
                    <th>Last Activity</th>
                    <th>Trades</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {activeSessions.map((session) => (
                    <tr key={session.id}>
                      <td className="username-cell">{session.username}</td>
                      <td>{session.ip_address}</td>
                      <td>{formatTime(session.login_time)}</td>
                      <td>{formatTime(session.last_activity)}</td>
                      <td>{session.trades_count}</td>
                      <td>
                        <button 
                          onClick={() => handleUserSelect(session.user_id)}
                          className="view-user-btn"
                        >
                          View Details
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        {/* Recent Activity */}
        <div className="monitoring-section">
          <h2>Recent Activity ({recentActivity.length})</h2>
          <div className="activity-list">
            {recentActivity.length === 0 ? (
              <p>No recent activity</p>
            ) : (
              recentActivity.map((activity) => (
                <div key={activity.id} className="activity-item">
                  <div className="activity-header">
                    <span className="activity-user">{activity.username}</span>
                    <span className="activity-time">{formatTime(activity.timestamp)}</span>
                    <span 
                      className="activity-status"
                      style={{ color: getActivityColor(activity.success) }}
                    >
                      {activity.success ? '✓' : '✗'}
                    </span>
                  </div>
                  <div className="activity-details">
                    <span className="activity-action">{activity.action}</span>
                    {activity.details && (
                      <span className="activity-description">: {activity.details}</span>
                    )}
                    {activity.error_message && (
                      <span className="activity-error"> - {activity.error_message}</span>
                    )}
                  </div>
                  <div className="activity-meta">
                    <span className="activity-ip">{activity.ip_address}</span>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* User Details Modal */}
      {selectedUser && userDetails && (
        <div className="user-details-modal">
          <div className="modal-content">
            <div className="modal-header">
              <h2>User Details - ID: {selectedUser}</h2>
              <button 
                className="modal-close"
                onClick={() => {
                  setSelectedUser(null);
                  setUserDetails(null);
                }}
              >
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="user-activity-section">
                <h3>Recent Activities ({userDetails.activities?.length || 0})</h3>
                <div className="user-activity-list">
                  {userDetails.activities?.map((activity) => (
                    <div key={activity.id} className="user-activity-item">
                      <div className="activity-timestamp">{formatTime(activity.timestamp)}</div>
                      <div className="activity-info">
                        <span className="activity-action">{activity.action}</span>
                        {activity.details && <span>: {activity.details}</span>}
                      </div>
                      <div className="activity-status" style={{ color: getActivityColor(activity.success) }}>
                        {activity.success ? '✓' : '✗'}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
              <div className="user-sessions-section">
                <h3>Recent Sessions ({userDetails.recent_sessions?.length || 0})</h3>
                <div className="user-sessions-list">
                  {userDetails.recent_sessions?.map((session) => (
                    <div key={session.id} className="user-session-item">
                      <div className="session-info">
                        <span className="session-ip">{session.ip_address}</span>
                        <span className="session-time">{formatTime(session.login_time)}</span>
                        <span className="session-duration">
                          {formatDuration(session.login_time, session.logout_time)}
                        </span>
                        <span className="session-trades">{session.trades_count} trades</span>
                      </div>
                      <div className={`session-status ${session.is_active ? 'active' : 'inactive'}`}>
                        {session.is_active ? 'Active' : 'Ended'}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MonitoringDashboard;