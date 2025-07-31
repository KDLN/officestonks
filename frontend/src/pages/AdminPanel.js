import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Navigation from "../components/Navigation";
import GameConfigSection from "../components/GameConfigSection";
import {
  checkAdminStatus,
  getAllUsers,
  resetStockPrices,
  clearAllChats,
  updateUser,
  deleteUser,
  createNews,
  getGameConfig,
  updateGameConfig,
  resetGameConfig,
  loadBalancedConfig,
  forceCrisisEvent,
  forceBankruptcy,
  forceRecovery,
  getSimulatorStatus,
} from "../services/admin";
import "./AdminPanel.css";

const AdminPanel = () => {
  const navigate = useNavigate();
  const [isAdmin, setIsAdmin] = useState(false);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [statusMessage, setStatusMessage] = useState(null);
  const [newsTitle, setNewsTitle] = useState("");
  const [newsContent, setNewsContent] = useState("");
  const [newsExpiry, setNewsExpiry] = useState("");
  const [gameConfig, setGameConfig] = useState(null);
  const [configLoading, setConfigLoading] = useState(false);
  const [crisisStockId, setCrisisStockId] = useState("");
  const [simulatorStatus, setSimulatorStatus] = useState(null);
  const [crisisLoading, setCrisisLoading] = useState(false);

  // Check if user is admin and load users data
  useEffect(() => {
    const checkAccess = async () => {
      try {
        const adminStatus = await checkAdminStatus();
        console.log("Admin status check result:", adminStatus);

        if (!adminStatus) {
          // Redirect non-admin users
          navigate("/dashboard");
          return;
        }

        setIsAdmin(true);
        await fetchUsers();
        await fetchGameConfig();
      } catch (err) {
        console.error("Admin access check error:", err);
        setError("Failed to check admin access. Please try again later.");
        setLoading(false);
      }
    };

    checkAccess();
  }, [navigate]);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      setError(null);
      console.log("Fetching users...");
      const usersData = await getAllUsers();
      console.log("Fetched users data:", usersData);

      // Ensure we have an array, even if empty
      setUsers(Array.isArray(usersData) ? usersData : []);
      setLoading(false);
    } catch (err) {
      console.error("Error fetching users:", err);
      setError(`Failed to load users: ${err.message}`);
      setUsers([]); // Set empty array to avoid undefined errors
      setLoading(false);
    }
  };

  const fetchGameConfig = async () => {
    try {
      setConfigLoading(true);
      const config = await getGameConfig();
      setGameConfig(config);
    } catch (err) {
      console.error("Error fetching game config:", err);
      setError(`Failed to load game configuration: ${err.message}`);
    } finally {
      setConfigLoading(false);
    }
  };

  const handleUpdateGameConfig = async (updatedConfig) => {
    try {
      setError(null);
      setStatusMessage("Updating game configuration...");
      const result = await updateGameConfig(updatedConfig);
      setGameConfig(result);
      setStatusMessage("Game configuration updated successfully!");
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error updating game config:", err);
      setError(`Failed to update game configuration: ${err.message}`);
    }
  };

  const handleResetGameConfig = async () => {
    if (!window.confirm("Are you sure you want to reset game configuration to defaults?")) {
      return;
    }
    
    try {
      setError(null);
      setStatusMessage("Resetting game configuration to defaults...");
      const result = await resetGameConfig();
      setGameConfig(result);
      setStatusMessage("Game configuration reset to defaults!");
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error resetting game config:", err);
      setError(`Failed to reset game configuration: ${err.message}`);
    }
  };

  const handleLoadBalancedConfig = async () => {
    if (!window.confirm("Are you sure you want to load the balanced game configuration?")) {
      return;
    }
    
    try {
      setError(null);
      setStatusMessage("Loading balanced game configuration...");
      const result = await loadBalancedConfig();
      setGameConfig(result);
      setStatusMessage("Balanced game configuration loaded!");
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error loading balanced config:", err);
      setError(`Failed to load balanced configuration: ${err.message}`);
    }
  };

  const handleResetStockPrices = async () => {
    try {
      setError(null);
      setStatusMessage("Resetting stock prices...");
      console.log("Resetting stock prices...");

      const result = await resetStockPrices();
      console.log("Reset stock prices result:", result);

      setStatusMessage("Stock prices reset successfully!");

      // Clear status message after 3 seconds
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error resetting stock prices:", err);
      setStatusMessage(null);
      setError("Failed to reset stock prices: " + err.message);

      // Clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    }
  };

  const handleClearChat = async () => {
    try {
      setError(null);
      setStatusMessage("Clearing chat messages...");
      console.log("Clearing chat messages...");

      const result = await clearAllChats();
      console.log("Clear chat result:", result);

      setStatusMessage("Chat messages cleared successfully!");

      // Clear status message after 3 seconds
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error clearing chat messages:", err);
      setStatusMessage(null);
      setError("Failed to clear chat messages: " + err.message);

      // Clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    }
  };

  const handleCreateNews = async () => {
    try {
      setError(null);
      setStatusMessage("Posting news...");
      await createNews(newsTitle, newsContent, newsExpiry);
      setStatusMessage("News posted successfully");
      setNewsTitle("");
      setNewsContent("");
      setNewsExpiry("");
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error creating news:", err);
      setStatusMessage(null);
      setError("Failed to create news: " + err.message);
      setTimeout(() => setError(null), 5000);
    }
  };

  const handleToggleAdmin = async (user) => {
    try {
      setError(null);
      setStatusMessage(`Updating admin status for ${user.username}...`);
      console.log(
        `Toggling admin status for user ${user.id}: ${user.username}`,
      );

      const userData = {
        username: user.username,
        cash_balance: user.cash_balance,
        is_admin: !user.is_admin,
      };

      const result = await updateUser(user.id, userData);
      console.log("Update user result:", result);

      // Refresh user list
      await fetchUsers();
      setStatusMessage(`Updated admin status for ${user.username}`);

      // Clear status message after 3 seconds
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error updating user:", err);
      setStatusMessage(null);
      setError("Failed to update user: " + err.message);

      // Clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    }
  };

  const handleDeleteUser = async (userId, username) => {
    // Ask for confirmation
    if (!window.confirm(`Are you sure you want to delete user ${username}?`)) {
      return;
    }

    try {
      setError(null);
      setStatusMessage(`Deleting user ${username}...`);
      console.log(`Deleting user ${userId}: ${username}`);

      const result = await deleteUser(userId);
      console.log("Delete user result:", result);

      // Refresh user list
      await fetchUsers();
      setStatusMessage(`User ${username} deleted successfully`);

      // Clear status message after 3 seconds
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error deleting user:", err);
      setStatusMessage(null);
      setError("Failed to delete user: " + err.message);

      // Clear error after 5 seconds
      setTimeout(() => setError(null), 5000);
    }
  };

  // Crisis testing functions
  const handleForceCrisis = async () => {
    if (!crisisStockId) {
      setError("Please enter a stock ID");
      return;
    }

    try {
      setCrisisLoading(true);
      setError(null);
      setStatusMessage("Forcing crisis event...");
      
      const result = await forceCrisisEvent(parseInt(crisisStockId));
      setStatusMessage("Crisis event triggered successfully!");
      console.log("Crisis event result:", result);
      
      // Refresh simulator status
      await fetchSimulatorStatus();
      
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error forcing crisis:", err);
      setError("Failed to force crisis: " + err.message);
      setTimeout(() => setError(null), 5000);
    } finally {
      setCrisisLoading(false);
    }
  };

  const handleForceBankruptcy = async () => {
    if (!crisisStockId) {
      setError("Please enter a stock ID");
      return;
    }

    try {
      setCrisisLoading(true);
      setError(null);
      setStatusMessage("Forcing bankruptcy...");
      
      const result = await forceBankruptcy(parseInt(crisisStockId));
      setStatusMessage("Bankruptcy triggered successfully!");
      console.log("Bankruptcy result:", result);
      
      await fetchSimulatorStatus();
      
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error forcing bankruptcy:", err);
      setError("Failed to force bankruptcy: " + err.message);
      setTimeout(() => setError(null), 5000);
    } finally {
      setCrisisLoading(false);
    }
  };

  const handleForceRecovery = async () => {
    if (!crisisStockId) {
      setError("Please enter a stock ID");
      return;
    }

    try {
      setCrisisLoading(true);
      setError(null);
      setStatusMessage("Forcing recovery...");
      
      const result = await forceRecovery(parseInt(crisisStockId));
      setStatusMessage("Recovery triggered successfully!");
      console.log("Recovery result:", result);
      
      await fetchSimulatorStatus();
      
      setTimeout(() => setStatusMessage(null), 3000);
    } catch (err) {
      console.error("Error forcing recovery:", err);
      setError("Failed to force recovery: " + err.message);
      setTimeout(() => setError(null), 5000);
    } finally {
      setCrisisLoading(false);
    }
  };

  const fetchSimulatorStatus = async () => {
    try {
      const status = await getSimulatorStatus();
      setSimulatorStatus(status);
    } catch (err) {
      console.error("Error fetching simulator status:", err);
    }
  };

  if (loading) {
    return <div className="loading">Loading admin panel...</div>;
  }

  if (!isAdmin) {
    return (
      <div className="error">Access denied. Admin privileges required.</div>
    );
  }

  // Ensure users is always an array
  const safeUsers = Array.isArray(users) ? users : [];

  return (
    <div className="admin-panel-page">
      <Navigation />
      <div className="admin-panel-container">
        <div className="admin-panel-header">
          <h1>Admin Panel</h1>
        </div>

        <div className="news-section">
          <h2>Post News</h2>
          <div className="news-form">
            <input
              type="text"
              placeholder="Title"
              value={newsTitle}
              onChange={(e) => setNewsTitle(e.target.value)}
            />
            <textarea
              placeholder="Content"
              value={newsContent}
              onChange={(e) => setNewsContent(e.target.value)}
            />
            <input
              type="datetime-local"
              value={newsExpiry}
              onChange={(e) => setNewsExpiry(e.target.value)}
            />
            <button onClick={handleCreateNews} className="action-button">
              Post News
            </button>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        {statusMessage && <div className="status-message">{statusMessage}</div>}

        <GameConfigSection
          gameConfig={gameConfig}
          onUpdate={handleUpdateGameConfig}
          onReset={handleResetGameConfig}
          onLoadBalanced={handleLoadBalancedConfig}
          loading={configLoading}
        />

        <div className="admin-actions">
          <h2>System Actions</h2>
          <div className="action-buttons">
            <button onClick={handleResetStockPrices} className="action-button">
              Reset Stock Prices
            </button>
            <button onClick={handleClearChat} className="action-button">
              Clear All Chat Messages
            </button>
          </div>
        </div>

        <div className="crisis-testing-section">
          <h2>🧪 Crisis Testing</h2>
          <p className="section-description">
            Test crisis events, bankruptcies, and recoveries for debugging and demonstration purposes.
          </p>
          
          <div className="crisis-controls">
            <div className="stock-input-group">
              <label htmlFor="crisisStockId">Stock ID:</label>
              <input
                id="crisisStockId"
                type="number"
                value={crisisStockId}
                onChange={(e) => setCrisisStockId(e.target.value)}
                placeholder="Enter stock ID (e.g., 1)"
                min="1"
              />
            </div>
            
            <div className="crisis-action-buttons">
              <button 
                onClick={handleForceCrisis}
                disabled={crisisLoading || !crisisStockId}
                className="crisis-button crisis-button-crisis"
              >
                🚨 Force Crisis Event
              </button>
              <button 
                onClick={handleForceBankruptcy}
                disabled={crisisLoading || !crisisStockId}
                className="crisis-button crisis-button-bankruptcy"
              >
                💀 Force Bankruptcy
              </button>
              <button 
                onClick={handleForceRecovery}
                disabled={crisisLoading || !crisisStockId}
                className="crisis-button crisis-button-recovery"
              >
                🚀 Force Recovery
              </button>
              <button 
                onClick={fetchSimulatorStatus}
                disabled={crisisLoading}
                className="crisis-button crisis-button-status"
              >
                📊 Get Simulator Status
              </button>
            </div>
          </div>

          {simulatorStatus && (
            <div className="simulator-status">
              <h3>Simulator Status</h3>
              <div className="status-info">
                <p><strong>Total Stocks:</strong> {simulatorStatus.stock_count}</p>
                <p><strong>Last Updated:</strong> {new Date(simulatorStatus.timestamp).toLocaleString()}</p>
              </div>
              
              {simulatorStatus.stocks && Object.keys(simulatorStatus.stocks).length > 0 && (
                <div className="stocks-status">
                  <h4>Stock Details:</h4>
                  <div className="stocks-grid">
                    {Object.entries(simulatorStatus.stocks).slice(0, 10).map(([id, stock]) => (
                      <div key={id} className="stock-status-card">
                        <div className="stock-header">
                          <span className="stock-symbol">{stock.symbol}</span>
                          <span className="stock-id">ID: {id}</span>
                        </div>
                        <div className="stock-info">
                          <div className="stock-price">${parseFloat(stock.base_price || 0).toFixed(2)}</div>
                          <div className="stock-trend">
                            Trend: {parseFloat(stock.trend || 0).toFixed(4)}
                          </div>
                          <div className="stock-sector">{stock.sector}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                  {Object.keys(simulatorStatus.stocks).length > 10 && (
                    <p className="more-stocks">
                      ... and {Object.keys(simulatorStatus.stocks).length - 10} more stocks
                    </p>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        <div className="users-section">
          <h2>User Management</h2>

          {safeUsers.length > 0 ? (
            <table className="users-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Username</th>
                  <th>Cash Balance</th>
                  <th>Admin</th>
                  <th>Created At</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {safeUsers.map(
                  (user) =>
                    user && (
                      <tr key={user.id || Math.random()}>
                        <td>{user.id || "N/A"}</td>
                        <td>{user.username || "Unknown"}</td>
                        <td>${(user.cash_balance || 0).toFixed(2)}</td>
                        <td>
                          <span
                            className={
                              user.is_admin ? "admin-badge" : "user-badge"
                            }
                          >
                            {user.is_admin ? "Admin" : "User"}
                          </span>
                        </td>
                        <td>
                          {user.created_at
                            ? new Date(user.created_at).toLocaleDateString()
                            : "N/A"}
                        </td>
                        <td className="action-cell">
                          <button
                            onClick={() => handleToggleAdmin(user)}
                            className="admin-toggle-btn"
                          >
                            {user.is_admin ? "Remove Admin" : "Make Admin"}
                          </button>
                          <button
                            onClick={() =>
                              handleDeleteUser(user.id, user.username)
                            }
                            className="delete-user-btn"
                          >
                            Delete
                          </button>
                        </td>
                      </tr>
                    ),
                )}
              </tbody>
            </table>
          ) : (
            <div className="no-users">
              {error ? (
                <div className="error-message">{error}</div>
              ) : (
                <p>No users found. {loading ? "Loading..." : ""}</p>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AdminPanel;
