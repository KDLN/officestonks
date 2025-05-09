import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Login from './Login';
import * as authService from '../services/auth';

// Mock the auth service
jest.mock('../services/auth', () => ({
  login: jest.fn(),
}));

// Mock the useNavigate hook
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
}));

describe('Login Component', () => {
  beforeEach(() => {
    // Reset mocks before each test
    jest.clearAllMocks();
  });

  test('renders login form', () => {
    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Check that the form elements are rendered
    expect(screen.getByText('Login to Office Stonks')).toBeInTheDocument();
    expect(screen.getByLabelText('Username')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument();
    expect(screen.getByText('Don\'t have an account?')).toBeInTheDocument();
    expect(screen.getByText('Register')).toBeInTheDocument();
  });

  test('submits form with username and password', async () => {
    // Mock successful login
    authService.login.mockResolvedValueOnce({ token: 'test-token', user_id: 1 });

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    // Submit the form
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Check that login was called with correct arguments
    expect(authService.login).toHaveBeenCalledWith('testuser', 'password123');

    // Wait for login to complete
    await waitFor(() => {
      expect(authService.login).toHaveBeenCalled();
    });
  });

  test('displays error message when login fails', async () => {
    // Mock failed login
    authService.login.mockRejectedValueOnce(new Error('Invalid credentials'));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'wrongpassword' } });

    // Submit the form
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Wait for error message to appear
    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    });
  });

  test('disables form submission while loading', async () => {
    // Mock login that takes time to resolve
    authService.login.mockImplementationOnce(() => new Promise(resolve => {
      setTimeout(() => resolve({ token: 'test-token', user_id: 1 }), 100);
    }));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Username'), { target: { value: 'testuser' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    // Submit the form
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Check that the button shows loading state
    expect(screen.getByRole('button', { name: 'Logging in...' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Logging in...' })).toBeDisabled();

    // Wait for login to complete
    await waitFor(() => {
      expect(authService.login).toHaveBeenCalled();
    });
  });
});