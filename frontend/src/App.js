import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './App.css';

// Import components
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import StockList from './pages/StockList';
import StockDetail from './pages/StockDetail';
import Leaderboard from './pages/Leaderboard';
import Transactions from './pages/Transactions';
import AdminPanel from './pages/AdminPanel';
import AuditLog from './pages/AuditLog';
import Portfolio from './pages/Portfolio';
import Tests from './pages/Tests';
import MonitoringDashboard from './pages/MonitoringDashboard';
import SSEDebug from './pages/SSEDebug';
import PollingDebug from './pages/PollingDebug';
import ProtectedRoute from './components/ProtectedRoute';
import AuthSync from './components/AuthSync';
import ChangelogModal from './components/ChangelogModal';
import ErrorBoundary from './components/ErrorBoundary';
import AdminToastManager from './components/AdminToastManager';
import { ThemeProvider } from './contexts/ThemeContext';
import { AuthProvider } from './contexts/AuthContext';
import { ChangelogProvider, useChangelog } from './contexts/ChangelogContext';

// Leaderboard component is now implemented

// Transactions component is now implemented

const AppContent = () => {
  const { manualTrigger, closeChangelog } = useChangelog();

  // Check for render issues
  React.useEffect(() => {
    const checkRenderHealth = () => {
      const root = document.getElementById('root');
      if (root && root.children.length === 0) {
        console.error('❌ App failed to render properly');
        window.location.reload();
      }
    };
    
    const timer = setTimeout(checkRenderHealth, 2000);
    return () => clearTimeout(timer);
  }, []);

  return (
    <Router>
          <div className="App">
            <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            
            {/* Protected routes */}
            <Route path="/dashboard" element={<ProtectedRoute element={<Dashboard />} />} />
            <Route path="/stocks" element={<ProtectedRoute element={<StockList />} />} />
            <Route path="/stock/:id" element={<ProtectedRoute element={<StockDetail />} />} />
            <Route path="/portfolio" element={<ProtectedRoute element={<Portfolio />} />} />
            <Route path="/leaderboard" element={<ProtectedRoute element={<Leaderboard />} />} />
            <Route path="/transactions" element={<ProtectedRoute element={<Transactions />} />} />
            <Route path="/admin" element={<ProtectedRoute element={<AdminPanel />} />} />
            <Route path="/audit" element={<ProtectedRoute element={<AuditLog />} />} />
            <Route path="/tests" element={<ProtectedRoute element={<Tests />} />} />
            <Route path="/monitoring" element={<ProtectedRoute element={<MonitoringDashboard />} />} />
            <Route path="/sse-debug" element={<ProtectedRoute element={<SSEDebug />} />} />
            <Route path="/polling-debug" element={<ProtectedRoute element={<PollingDebug />} />} />

            {/* Default redirect */}
            <Route path="*" element={<Navigate to="/login" />} />
          </Routes>
          
          {/* Global changelog modal - shows for authenticated users */}
          <ChangelogModal manualTrigger={manualTrigger} onManualClose={closeChangelog} />
          
          {/* Global admin toast manager */}
          <AdminToastManager />
          </div>
          </Router>
  );
};

function App() {
  // Helper function to log environment info
  const logEnvironmentInfo = () => {
    console.log('🌍 Current environment:', {
      hostname: window.location.hostname,
      pathname: window.location.pathname,
      search: window.location.search,
      protocol: window.location.protocol
    });
    console.log('⚙️ Environment variables:', {
      REACT_APP_BACKEND_URL: process.env.REACT_APP_BACKEND_URL,
      REACT_APP_API_URL: process.env.REACT_APP_API_URL,
      REACT_APP_SUPABASE_URL: process.env.REACT_APP_SUPABASE_URL?.substring(0, 20) + '...',
      NODE_ENV: process.env.NODE_ENV
    });
  };

  // Helper function to log localStorage info
  const logStorageInfo = () => {
    console.log('💾 LocalStorage contents:', {
      token: localStorage.getItem('token') ? 'present' : 'missing',
      userId: localStorage.getItem('userId') || 'missing',
      username: localStorage.getItem('username') || 'missing'
    });
  };

  // Log app initialization
  React.useEffect(() => {
    const initializeApp = async () => {
      console.log('🚀 Office Stonks app initializing...');
      logEnvironmentInfo();

      // Initialize storage manager before reading localStorage
      const { initializeStorage } = await import('./services/storageManager');
      await initializeStorage();

      logStorageInfo();
    };

    initializeApp();
  }, []);

  return (
    <ErrorBoundary>
      <AuthProvider>
        <ThemeProvider>
          <ChangelogProvider>
            <AuthSync>
              <AppContent />
            </AuthSync>
          </ChangelogProvider>
        </ThemeProvider>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App;