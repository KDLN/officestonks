import React, { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const AuthRedirect = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, loading } = useAuth();

  useEffect(() => {
    const handleRedirect = async () => {
      console.log('AuthRedirect: Current state', {
        user: user?.email,
        loading,
        pathname: location.pathname,
        search: location.search,
        hash: window.location.hash,
        hostname: window.location.hostname,
        fullURL: window.location.href
      });

      // Check if we have auth data in the URL hash (Supabase OAuth response)
      if (window.location.hash && window.location.hash.includes('access_token')) {
        console.log('AuthRedirect: Auth tokens in URL hash, waiting for Supabase to process...');
        // Supabase will automatically process the hash and trigger auth state change
        // Just wait for it to complete
        return;
      }

      // Wait for auth to load
      if (loading) return;

      // Check if we have a user (successful auth)
      if (user) {
        // Check URL params for beta redirect
        const urlParams = new URLSearchParams(location.search);
        const shouldRedirectToBeta = urlParams.get('beta') === 'true';
        
        console.log('AuthRedirect: User authenticated, beta param:', shouldRedirectToBeta);

        if (shouldRedirectToBeta && window.location.hostname === 'officestonks.com') {
          console.log('AuthRedirect: Redirecting to beta.officestonks.com');
          // Force redirect to beta site
          window.location.replace('https://beta.officestonks.com/dashboard');
          return;
        }

        // Otherwise, navigate to dashboard normally
        navigate('/dashboard');
      } else if (!window.location.hash) {
        // No user and no auth hash, redirect to login
        console.log('AuthRedirect: No user and no auth hash, redirecting to login');
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
      flexDirection: 'column',
      backgroundColor: '#f5f5f5',
      fontFamily: 'Arial, sans-serif'
    }}>
      <h2>Completing sign in...</h2>
      <p>Please wait while we authenticate your account.</p>
      {window.location.hash && window.location.hash.includes('access_token') && (
        <p style={{ fontSize: '14px', color: '#666', marginTop: '10px' }}>
          Processing authentication tokens...
        </p>
      )}
    </div>
  );
};

export default AuthRedirect;