import React, { useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { logout } from "../services/auth";
import { checkAdminStatus } from "../services/admin";
import { fetchActiveNews } from "../services/news";
import NewsModal from "./NewsModal";
import ThemeToggle from "./ThemeToggle";
import "./Navigation.css";

const Navigation = () => {
  const navigate = useNavigate();
  const [isAdmin, setIsAdmin] = useState(false);
  const [newsItem, setNewsItem] = useState(null);

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

    const loadNews = async () => {
      try {
        const news = await fetchActiveNews();
        if (Array.isArray(news) && news.length > 0) {
          setNewsItem(news[0]);
        }
      } catch (err) {
        console.error("Error fetching news:", err);
      }
    };

    checkAdmin();
    loadNews();
  }, []);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <>
      <nav className="navigation">
        <div className="nav-logo">
          <Link to="/dashboard">Office Stonks</Link>
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
      <NewsModal news={newsItem} onClose={() => setNewsItem(null)} />
    </>
  );
};

export default Navigation;
