import React, { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const AuthRedirect = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, loading } = useAuth();

  useEffect(() => {
    const handleRedirect = async () => {
      // Wait for auth to load
      if (loading) return;

      console.log('AuthRedirect: Current state', {
        user: user?.email,
        loading,
        pathname: location.pathname,
        search: location.search,
        hostname: window.location.hostname
      });

      // Check if we have a user (successful auth)
      if (user) {
        // Check URL params for beta redirect
        const urlParams = new URLSearchParams(location.search);
        const shouldRedirectToBeta = urlParams.get('beta') === 'true';
        
        console.log('AuthRedirect: Beta param', shouldRedirectToBeta);

        if (shouldRedirectToBeta && window.location.hostname === 'officestonks.com') {
          console.log('AuthRedirect: Redirecting to beta.officestonks.com');
          // Force redirect to beta site
          window.location.replace('https://beta.officestonks.com/dashboard');
          return;
        }

        // Otherwise, navigate to dashboard normally
        navigate('/dashboard');
      } else {
        // No user, redirect to login
        console.log('AuthRedirect: No user, redirecting to login');
        navigate('/login');
      }
    };

    handleRedirect();
  }, [user, loading, navigate, location]);

  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      height: '100vh',
      flexDirection: 'column'
    }}>
      <h2>Authenticating...</h2>
      <p>Please wait while we complete your sign in.</p>
    </div>
  );
};

export default AuthRedirect;