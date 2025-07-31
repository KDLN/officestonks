import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { registerWithWorkaround } from '../services/supabaseWorkaround';
import LoadingSpinner from '../components/LoadingSpinner';

const Register = () => {
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { signInWithProvider } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    // Password validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }

    setLoading(true);

    try {
      // Use the workaround for noop mailer
      await registerWithWorkaround(email, password, username);
      
      // If successful, navigate to dashboard
      navigate('/dashboard');
    } catch (error) {
      console.error('Registration error:', error);
      setError(error.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleDiscordRegister = async () => {
    setError('');
    setLoading(true);
    try {
      await signInWithProvider('discord');
    } catch (error) {
      console.error('Discord registration error:', error);
      setError(error.message || 'Registration failed. Please try again.');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="auth-container">
        <h2>Create an Account</h2>
        <LoadingSpinner message="Creating your account..." />
      </div>
    );
  }

  return (
    <div className="auth-container">
      <h2>Create an Account</h2>
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
          <label htmlFor="username">Username</label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
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
        <div className="form-group">
          <label htmlFor="confirmPassword">Confirm Password</label>
          <input
            type="password"
            id="confirmPassword"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            disabled={loading}
          />
        </div>
        <button type="submit" className="btn" disabled={loading}>
          Register
        </button>
        <button type="button" className="btn" onClick={handleDiscordRegister} disabled={loading} style={{ marginTop: '10px' }}>
          Register with Discord
        </button>
      </form>
      <p>
        Already have an account? <Link to="/login">Login</Link>
      </p>
    </div>
  );
};

export default Register;