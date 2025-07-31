import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Navigation from './Navigation';
import { ChangelogProvider } from '../contexts/ChangelogContext';
import { ThemeProvider } from '../contexts/ThemeContext';
import * as authService from '../services/auth';

// Provide matchMedia mock for ThemeProvider
beforeAll(() => {
  if (!window.matchMedia) {
    window.matchMedia = () => ({
      matches: false,
      addListener: jest.fn(),
      removeListener: jest.fn(),
    });
  }
});

// Mock the auth service
jest.mock('../services/auth', () => ({
  logout: jest.fn(),
}));

// Mock the useNavigate hook
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
}));

describe('Navigation Component', () => {
  beforeEach(() => {
    // Reset mocks before each test
    jest.clearAllMocks();
  });

  test('renders all navigation links', () => {
    render(
      <BrowserRouter>
        <ThemeProvider>
          <ChangelogProvider>
            <Navigation />
          </ChangelogProvider>
        </ThemeProvider>
      </BrowserRouter>
    );

    // Check that all nav links are rendered
    expect(screen.getByText('Office Stonks')).toBeInTheDocument();
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Stocks')).toBeInTheDocument();
    expect(screen.getByText('Portfolio')).toBeInTheDocument();
    expect(screen.getByText('Leaderboard')).toBeInTheDocument();
    expect(screen.getByText('Logout')).toBeInTheDocument();
  });

  test('calls logout when logout button is clicked', () => {
    render(
      <BrowserRouter>
        <ThemeProvider>
          <ChangelogProvider>
            <Navigation />
          </ChangelogProvider>
        </ThemeProvider>
      </BrowserRouter>
    );

    // Click the logout button
    fireEvent.click(screen.getByText('Logout'));

    // Check that logout was called
    expect(authService.logout).toHaveBeenCalled();
  });
});