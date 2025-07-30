import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const AuthRedirect = () => {
  const navigate = useNavigate();
  const { user, loading } = useAuth();

  useEffect(() => {
    const handleRedirect = async () => {
      // Wait for auth to load
      if (loading) return;

      // Check if we have a user (successful auth)
      if (user) {
        // Check if we should redirect to beta
        const betaRedirect = localStorage.getItem('oauth_beta_redirect');
        console.log('AuthRedirect: Checking beta redirect', {
          betaRedirect,
          currentHost: window.location.hostname,
          user: user.email
        });

        if (betaRedirect === 'true') {
          localStorage.removeItem('oauth_beta_redirect');
          
          // If we're on production, redirect to beta
          if (window.location.hostname === 'officestonks.com') {
            console.log('AuthRedirect: Redirecting to beta.officestonks.com');
            window.location.href = 'https://beta.officestonks.com/dashboard';
            return;
          }
        }

        // Otherwise, navigate to dashboard normally
        navigate('/dashboard');
      } else {
        // No user, redirect to login
        navigate('/login');
      }
    };

    handleRedirect();
  }, [user, loading, navigate]);

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