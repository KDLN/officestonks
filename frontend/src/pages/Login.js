import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { loginWithWorkaround } from '../services/supabaseWorkaround';
import { isAuthenticated } from '../services/auth';
import LoadingSpinner from '../components/LoadingSpinner';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { signInWithProvider, session, isAuthenticated: supabaseAuth } = useAuth();

  // Check if user is already authenticated and redirect
  useEffect(() => {
    if (session || supabaseAuth || isAuthenticated()) {
      console.log('🚀 User already authenticated, redirecting to dashboard');
      navigate('/dashboard');
    }
  }, [session, supabaseAuth, navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      console.log('🔑 Starting email login process...');
      
      // Use the workaround for noop mailer issues
      const result = await loginWithWorkaround(email, password);
      console.log('✅ Login successful:', result);
      
      // If we got a token from Office Stonks backend, navigate immediately
      if (result && (result.token || isAuthenticated())) {
        console.log('🎯 Office Stonks token detected, navigating to dashboard');
        navigate('/dashboard');
      } else if (result && result.session) {
        // Supabase login successful - let AuthContext handle the navigation
        console.log('🎯 Supabase session detected, waiting for AuthContext navigation');
      } else {
        // Force navigation after a short delay if nothing else happens
        setTimeout(() => {
          if (isAuthenticated()) {
            console.log('🔄 Fallback navigation after successful authentication');
            navigate('/dashboard');
          }
        }, 1000);
      }
    } catch (error) {
      console.error('❌ Login error:', error);
      setError(error.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleDiscordLogin = async () => {
    setError('');
    setLoading(true);
    try {
      await signInWithProvider('discord');
    } catch (error) {
      console.error('Discord login error:', error);
      setError(error.message || 'Login failed. Please try again.');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="auth-container">
        <h2>Login to Office Stonks</h2>
        <LoadingSpinner message="Logging you in..." />
      </div>
    );
  }

  return (
    <div className="auth-container">
      <h2>Login to Office Stonks</h2>
      {error && <div className="error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={loading}
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={loading}
          />
        </div>
        <button type="submit" className="btn" disabled={loading}>
          Login
        </button>
        <button type="button" className="btn" onClick={handleDiscordLogin} disabled={loading} style={{ marginTop: '10px' }}>
          Login with Discord
        </button>
      </form>
      <p>
        Don't have an account? <Link to="/register">Register</Link>
      </p>
    </div>
  );
};

export default Login;