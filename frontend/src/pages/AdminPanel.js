import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import Navigation from "../components/Navigation";
import GameConfigSection from "../components/GameConfigSection";
import StorageHealthPanel from "../components/StorageHealthPanel";
import LogsPanel from "../components/LogsPanel";
import StockManagementSection from "../components/StockManagementSection";
import {
  checkAdminStatus,
  getAllUsers,
  resetStockPrices,
  clearAllChats,
  updateUser,
  deleteUser,
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
import "../components/announcement-styles.css";

const AdminPanel = () => {
  const navigate = useNavigate();
  const [isAdmin, setIsAdmin] = useState(false);
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [statusMessage, setStatusMessage] = useState(null);
  const [gameConfig, setGameConfig] = useState(null);
  const [configLoading, setConfigLoading] = useState(false);
  const [crisisStockId, setCrisisStockId] = useState("");
  const [simulatorStatus, setSimulatorStatus] = useState(null);
  const [crisisLoading, setCrisisLoading] = useState(false);
  const [announcementText, setAnnouncementText] = useState("");
  const [announcementType, setAnnouncementType] = useState("general");

  // Helper function for consistent status/error handling
  const setTemporaryStatus = (message, duration = 3000) => {
    setStatusMessage(message);
    setTimeout(() => setStatusMessage(null), duration);
  };

  const setTemporaryError = (message, duration = 5000) => {
    setError(message);
    setTimeout(() => setError(null), duration);
  };

  // Helper function for common admin action pattern
  const handleAdminAction = async (actionName, actionFn, successMessage) => {
    try {
      setError(null);
      setStatusMessage(`${actionName}...`);
      console.log(actionName);

      const result = await actionFn();
      console.log(`${actionName} result:`, result);

      setTemporaryStatus(successMessage);
    } catch (err) {
      console.error(`Error ${actionName.toLowerCase()}:`, err);
      setStatusMessage(null);
      setTemporaryError(`Failed to ${actionName.toLowerCase()}: ${err.message}`);
    }
  };

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
      setTemporaryStatus("Game configuration updated successfully!");
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
      setTemporaryStatus("Game configuration reset to defaults!");
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
      setTemporaryStatus("Balanced game configuration loaded!");
    } catch (err) {
      console.error("Error loading balanced config:", err);
      setError(`Failed to load balanced configuration: ${err.message}`);
    }
  };

  const handleResetStockPrices = async () => {
    await handleAdminAction(
      "Resetting stock prices",
      resetStockPrices,
      "Stock prices reset successfully!"
    );
  };

  const handleClearChat = async () => {
    await handleAdminAction(
      "Clearing chat messages",
      clearAllChats,
      "Chat messages cleared successfully!"
    );
  };

  const handleSendAnnouncement = () => {
    if (!announcementText.trim()) return;
    
    // Create a custom event to trigger the toast
    const event = new CustomEvent('adminAnnouncement', {
      detail: { 
        message: announcementText.trim(),
        type: announcementType
      }
    });
    window.dispatchEvent(event);
    
    setAnnouncementText("");
    setTemporaryStatus(`${announcementType} announcement sent!`);
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

      await fetchUsers();
      setTemporaryStatus(`Updated admin status for ${user.username}`);
    } catch (err) {
      console.error("Error updating user:", err);
      setStatusMessage(null);
      setTemporaryError("Failed to update user: " + err.message);
    }
  };

  const handleDeleteUser = async (userId, username) => {
    if (!window.confirm(`Are you sure you want to delete user ${username}?`)) {
      return;
    }

    try {
      setError(null);
      setStatusMessage(`Deleting user ${username}...`);
      console.log(`Deleting user ${userId}: ${username}`);

      const result = await deleteUser(userId);
      console.log("Delete user result:", result);

      await fetchUsers();
      setTemporaryStatus(`User ${username} deleted successfully`);
    } catch (err) {
      console.error("Error deleting user:", err);
      setStatusMessage(null);
      setTemporaryError("Failed to delete user: " + err.message);
    }
  };

  // Helper function for crisis actions
  const handleCrisisAction = async (actionName, actionFn, successMessage) => {
    if (!crisisStockId) {
      setError("Please enter a stock ID");
      return;
    }

    try {
      setCrisisLoading(true);
      setError(null);
      setStatusMessage(`${actionName}...`);
      
      const result = await actionFn(parseInt(crisisStockId));
      setTemporaryStatus(successMessage);
      console.log(`${actionName} result:`, result);
      
      await fetchSimulatorStatus();
    } catch (err) {
      console.error(`Error ${actionName.toLowerCase()}:`, err);
      setTemporaryError(`Failed to ${actionName.toLowerCase()}: ${err.message}`);
    } finally {
      setCrisisLoading(false);
    }
  };

  // Crisis testing functions
  const handleForceCrisis = () => handleCrisisAction(
    "Forcing crisis event",
    forceCrisisEvent,
    "Crisis event triggered successfully!"
  );

  const handleForceBankruptcy = () => handleCrisisAction(
    "Forcing bankruptcy",
    forceBankruptcy,
    "Bankruptcy triggered successfully!"
  );

  const handleForceRecovery = () => handleCrisisAction(
    "Forcing recovery",
    forceRecovery,
    "Recovery triggered successfully!"
  );

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

        <StorageHealthPanel />

        <LogsPanel />

        <div className="announcement-section">
          <h2>Admin Announcement</h2>
          <div className="announcement-form">
            <div className="announcement-type-selector">
              <label htmlFor="announcement-type">Type:</label>
              <select 
                id="announcement-type"
                value={announcementType} 
                onChange={(e) => setAnnouncementType(e.target.value)}
                className="announcement-type-select"
              >
                <option value="general">📢 General</option>
                <option value="notice">⚠️ Notice</option>
                <option value="shutdown">🚨 Server Shutdown</option>
              </select>
            </div>
            <textarea
              placeholder="Type announcement message..."
              value={announcementText}
              onChange={(e) => setAnnouncementText(e.target.value)}
              rows={3}
              className="announcement-textarea"
            />
            <button 
              onClick={handleSendAnnouncement} 
              className={`action-button announcement-btn ${announcementType}`}
              disabled={!announcementText.trim()}
            >
              Send {announcementType.charAt(0).toUpperCase() + announcementType.slice(1)} Announcement
            </button>
          </div>
        </div>

        {error && <div className="error-message">{error}</div>}

        {statusMessage && <div className="status-message">{statusMessage}</div>}

        <StockManagementSection />

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
