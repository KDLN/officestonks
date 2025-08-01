import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { checkAdminStatus } from "../services/admin";
import { useChangelog } from "../contexts/ChangelogContext";
import { useAuth } from "../contexts/AuthContext";
import ThemeToggle from "./ThemeToggle";
import "./Navigation.css";

const Navigation = () => {
  const navigate = useNavigate();
  const [isAdmin, setIsAdmin] = useState(false);
  const { openChangelog } = useChangelog();
  const { signOut } = useAuth();

  useEffect(() => {
    // Check if the user is an admin
    const checkAdmin = async () => {
      try {
        const adminStatus = await checkAdminStatus();
        setIsAdmin(adminStatus);
      } catch (error) {
        console.error("Error checking admin status:", error);
        setIsAdmin(false);
      }
    };

    checkAdmin();
  }, []);

  const handleLogout = async () => {
    try {
      await signOut();
      // Clear any remaining localStorage items
      localStorage.removeItem('token');
      localStorage.removeItem('userId');
      // Navigate to login page
      navigate("/login");
    } catch (error) {
      console.error('Error during logout:', error);
      // Force redirect even if signOut fails
      navigate("/login");
    }
  };

  return (
    <>
      <nav className="navigation">
        <div className="nav-logo">
          <Link to="/dashboard">Office Stonks</Link>
          {process.env.NODE_ENV === 'development' || window.location.hostname.includes('beta') ? (
            <span className="env-badge">BETA</span>
          ) : null}
        </div>
        <ul className="nav-links">
          <li>
            <Link to="/dashboard">Dashboard</Link>
          </li>
          <li>
            <Link to="/stocks">Stocks</Link>
          </li>
          <li>
            <Link to="/portfolio">Portfolio</Link>
          </li>
          <li>
            <Link to="/leaderboard">Leaderboard</Link>
          </li>
          {isAdmin && (
            <li>
              <Link to="/admin" className="admin-link">
                Admin
              </Link>
            </li>
          )}
          {isAdmin && (
            <li>
              <Link to="/audit" className="admin-link">Audit Log</Link>
            </li>
          )}
          {isAdmin && (
            <li>
              <Link to="/tests" className="admin-link">🧪 Tests</Link>
            </li>
          )}
          {isAdmin && (
            <li>
              <Link to="/monitoring" className="admin-link">📊 Monitoring</Link>
            </li>
          )}
          <li>
            <button onClick={() => {
              console.log('Changelog button clicked');
              openChangelog();
            }} className="changelog-button">
              📋 Changelog
            </button>
          </li>
          <li>
            <ThemeToggle />
          </li>
          <li>
            <button onClick={handleLogout} className="logout-button">
              Logout
            </button>
          </li>
        </ul>
      </nav>
    </>
  );
};

export default Navigation;
