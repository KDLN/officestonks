import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './App.css';

// Import components
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import StockList from './pages/StockList';
import StockDetail from './pages/StockDetail';
import ProtectedRoute from './components/ProtectedRoute';
import { isAuthenticated } from './services/auth';

// Placeholder components for pages we haven't implemented yet
const Portfolio = () => (
  <div style={{ padding: '50px', textAlign: 'center' }}>
    <h1>Portfolio View</h1>
    <p>This page will be implemented soon.</p>
  </div>
);

const Leaderboard = () => (
  <div style={{ padding: '50px', textAlign: 'center' }}>
    <h1>Leaderboard</h1>
    <p>This page will be implemented soon.</p>
  </div>
);

const Transactions = () => (
  <div style={{ padding: '50px', textAlign: 'center' }}>
    <h1>Transaction History</h1>
    <p>This page will be implemented soon.</p>
  </div>
);

function App() {
  return (
    <Router>
      <div className="App">
        <Routes>
          <Route path="/login" element={!isAuthenticated() ? <Login /> : <Navigate to="/dashboard" />} />
          <Route path="/register" element={!isAuthenticated() ? <Register /> : <Navigate to="/dashboard" />} />
          
          {/* Protected routes */}
          <Route path="/dashboard" element={<ProtectedRoute element={<Dashboard />} />} />
          <Route path="/stocks" element={<ProtectedRoute element={<StockList />} />} />
          <Route path="/stock/:id" element={<ProtectedRoute element={<StockDetail />} />} />
          <Route path="/portfolio" element={<ProtectedRoute element={<Portfolio />} />} />
          <Route path="/leaderboard" element={<ProtectedRoute element={<Leaderboard />} />} />
          <Route path="/transactions" element={<ProtectedRoute element={<Transactions />} />} />
          
          {/* Default redirect */}
          <Route path="*" element={<Navigate to={isAuthenticated() ? "/dashboard" : "/login"} />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;