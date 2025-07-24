import React from 'react';
import { useTheme } from '../contexts/ThemeContext';
import './ThemeToggle.css';

const ThemeToggle = () => {
  const { isDarkMode, toggleTheme } = useTheme();

  return (
    <button 
      className="theme-toggle"
      onClick={toggleTheme}
      aria-label={`Switch to ${isDarkMode ? 'light' : 'dark'} mode`}
      title={`Switch to ${isDarkMode ? 'light' : 'dark'} mode`}
    >
      <div className="theme-toggle-track">
        <div className="theme-toggle-thumb">
          {isDarkMode ? (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
            </svg>
          ) : (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="5" />
              <path d="m12 1-1 3-1-3" />
              <path d="m12 23-1-3-1 3" />
              <path d="m20 12 3 1-3 1" />
              <path d="m1 12 3 1-3 1" />
              <path d="m16.5 7.5 2.1-2.1-2.1 2.1" />
              <path d="m5.5 16.5-2.1 2.1 2.1-2.1" />
              <path d="m16.5 16.5-2.1-2.1 2.1 2.1" />
              <path d="m5.5 7.5 2.1-2.1-2.1 2.1" />
            </svg>
          )}
        </div>
      </div>
    </button>
  );
};

export default ThemeToggle;