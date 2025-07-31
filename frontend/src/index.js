import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import { initializeApp, emergencyRecovery } from './utils/startup';

// Run startup checks before rendering
const startApp = () => {
  try {
    // Initialize with safety checks
    const initialized = initializeApp();
    
    if (!initialized) {
      // App will reload after cleanup
      return;
    }
    
    const root = ReactDOM.createRoot(document.getElementById('root'));
    root.render(
      <React.StrictMode>
        <App />
      </React.StrictMode>
    );
  } catch (error) {
    console.error('❌ Critical startup error:', error);
    
    // Show emergency recovery UI
    const root = document.getElementById('root');
    root.innerHTML = `
      <div style="
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100vh;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        background: #0a0e27;
        color: #e0e6ed;
        text-align: center;
        padding: 20px;
      ">
        <h1 style="font-size: 3rem; margin-bottom: 1rem;">⚠️</h1>
        <h2 style="margin-bottom: 1rem;">Office Stonks Failed to Start</h2>
        <p style="margin-bottom: 2rem; opacity: 0.8;">
          The application encountered a critical error during startup.
        </p>
        <button onclick="window.emergencyRecovery()" style="
          background: #3b82f6;
          color: white;
          border: none;
          padding: 12px 24px;
          border-radius: 6px;
          font-size: 16px;
          cursor: pointer;
          margin-bottom: 1rem;
        ">
          Reset Application
        </button>
        <p style="font-size: 14px; opacity: 0.6;">
          This will clear all local data and restart the app
        </p>
      </div>
    `;
    
    // Make emergency recovery available globally
    window.emergencyRecovery = () => {
      emergencyRecovery();
      window.location.reload();
    };
  }
};

// Start the app
startApp();