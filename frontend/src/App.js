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
import Portfolio from './pages/Portfolio';
import ProtectedRoute from './components/ProtectedRoute';
import { ThemeProvider } from './contexts/ThemeContext';
import { isAuthenticated } from './services/auth';

// Leaderboard component is now implemented

// Transactions component is now implemented

function App() {
  return (
    <ThemeProvider>
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
            <Route path="/admin" element={<ProtectedRoute element={<AdminPanel />} />} />

            {/* Default redirect */}
            <Route path="*" element={<Navigate to={isAuthenticated() ? "/dashboard" : "/login"} />} />
          </Routes>
        </div>
      </Router>
    </ThemeProvider>
  );
}

export default App;