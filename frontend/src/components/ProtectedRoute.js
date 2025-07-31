import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import LoadingScreen from './LoadingScreen';

// Component to protect routes that require authentication
const ProtectedRoute = ({ element }) => {
  const { user, loading } = useAuth();

  // Show loading screen while checking auth status
  if (loading) {
    return <LoadingScreen message="Checking authentication..." />;
  }

  // Check if user is authenticated
  if (!user) {
    // Redirect to login page if not authenticated
    return <Navigate to="/login" replace />;
  }

  // Render the protected component if authenticated
  return element;
};

export default ProtectedRoute;