import { getAllStocks, getStockById, getUserPortfolio, executeTrade } from './stock';
import { getToken } from './auth';

// Mock fetch
global.fetch = jest.fn();

// Mock auth service
jest.mock('./auth', () => ({
  getToken: jest.fn(),
}));

describe('Stock Service', () => {
  beforeEach(() => {
    // Reset mocks before each test
    jest.clearAllMocks();
    fetch.mockClear();
  });

  test('getAllStocks fetches from the correct endpoint', async () => {
    // Mock response
    const mockStocks = [
      { id: 1, symbol: 'AAPL', name: 'Apple Inc.', current_price: 150.0 },
      { id: 2, symbol: 'MSFT', name: 'Microsoft Corporation', current_price: 300.0 },
    ];

    // Setup fetch mock to return mock data
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockStocks,
    });

    // Call the function
    const result = await getAllStocks();

    // Check that fetch was called with the correct URL
    expect(fetch).toHaveBeenCalledWith('/api/stocks', expect.any(Object));
    
    // Check that the function returns the mock data
    expect(result).toEqual(mockStocks);
  });

  test('getStockById fetches the correct stock', async () => {
    // Mock response
    const mockStock = { id: 1, symbol: 'AAPL', name: 'Apple Inc.', current_price: 150.0 };

    // Setup fetch mock to return mock data
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockStock,
    });

    // Call the function
    const result = await getStockById(1);

    // Check that fetch was called with the correct URL
    expect(fetch).toHaveBeenCalledWith('/api/stocks/1', expect.any(Object));
    
    // Check that the function returns the mock data
    expect(result).toEqual(mockStock);
  });

  test('getUserPortfolio includes auth token', async () => {
    // Mock token
    const mockToken = 'test-token';
    getToken.mockReturnValueOnce(mockToken);

    // Mock response
    const mockPortfolio = {
      cash_balance: 10000,
      stock_value: 5000,
      total_value: 15000,
      portfolio_items: [],
    };

    // Setup fetch mock to return mock data
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockPortfolio,
    });

    // Call the function
    const result = await getUserPortfolio();

    // Check that fetch was called with the correct URL and auth header
    expect(fetch).toHaveBeenCalledWith(
      '/api/portfolio',
      expect.objectContaining({
        headers: expect.objectContaining({
          'Authorization': `Bearer ${mockToken}`,
        }),
      })
    );
    
    // Check that the function returns the mock data
    expect(result).toEqual(mockPortfolio);
  });

  test('executeTrade sends the correct trade request', async () => {
    // Mock token
    const mockToken = 'test-token';
    getToken.mockReturnValueOnce(mockToken);

    // Mock response
    const mockResponse = { message: 'Trade executed successfully' };

    // Setup fetch mock to return mock data
    fetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    });

    // Call the function
    const result = await executeTrade(1, 10, 'buy');

    // Check that fetch was called with the correct URL, auth header, and body
    expect(fetch).toHaveBeenCalledWith(
      '/api/trading',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Authorization': `Bearer ${mockToken}`,
        }),
        body: expect.any(String),
      })
    );

    // Parse the body that was sent
    const body = JSON.parse(fetch.mock.calls[0][1].body);
    
    // Check that the body contains the correct trade details
    expect(body).toEqual({
      stock_id: 1,
      quantity: 10,
      action: 'buy',
    });
    
    // Check that the function returns the mock data
    expect(result).toEqual(mockResponse);
  });

  test('service functions handle errors correctly', async () => {
    // Setup fetch mock to return an error
    fetch.mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: async () => ({ error: 'Not found' }),
    });

    // Call the function and expect it to throw
    await expect(getStockById(999)).rejects.toThrow();

    // Check that fetch was called
    expect(fetch).toHaveBeenCalled();
  });
});