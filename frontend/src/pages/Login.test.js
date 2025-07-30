import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Login from './Login';
import * as workaround from '../services/supabaseWorkaround';

// Mock the auth service
jest.mock('../services/supabaseWorkaround', () => ({
  loginWithWorkaround: jest.fn(),
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
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument();
    expect(screen.getByText('Don\'t have an account?')).toBeInTheDocument();
    expect(screen.getByText('Register')).toBeInTheDocument();
  });

  test('submits form with email and password', async () => {
    // Mock successful login
    workaround.loginWithWorkaround.mockResolvedValueOnce({ user: {}, session: {} });

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    // Submit the form
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Check that login was called with correct arguments
    expect(workaround.loginWithWorkaround).toHaveBeenCalledWith('test@example.com', 'password123');

    // Wait for login to complete
    await waitFor(() => {
      expect(workaround.loginWithWorkaround).toHaveBeenCalled();
    });
  });

  test('displays error message when login fails', async () => {
    // Mock failed login
    workaround.loginWithWorkaround.mockRejectedValueOnce(new Error('Invalid credentials'));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'test@example.com' } });
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
    workaround.loginWithWorkaround.mockImplementationOnce(() => new Promise(resolve => {
      setTimeout(() => resolve({ user: {}, session: {} }), 100);
    }));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    // Fill in form fields
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'test@example.com' } });
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } });

    // Submit the form
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    // Check that the button shows loading state
    expect(screen.getByRole('button', { name: 'Logging in...' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Logging in...' })).toBeDisabled();

    // Wait for login to complete
    await waitFor(() => {
      expect(workaround.loginWithWorkaround).toHaveBeenCalled();
    });
  });
});
