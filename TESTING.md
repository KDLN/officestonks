# Testing Guidelines for Office Stonks

This document outlines the testing strategy for the Office Stonks application, including how to run tests and write new ones.

## Testing Environment

We use Docker to create a consistent testing environment. This approach makes it easy to run tests on any machine without worrying about installing dependencies or configuring databases.

### Requirements

To run tests with Docker:
- Docker
- Docker Compose

To run tests locally without Docker:
- Go 1.20 or higher
- Node.js 18 or higher
- MySQL 8.0 or higher

## Running Tests

### With Docker (Recommended)

Run all tests in containerized environments:

```bash
./run_tests.sh
```

This script will:
1. Set up a test database in a Docker container
2. Run backend tests in a Docker container
3. Run frontend tests in a Docker container
4. Clean up all containers when finished

### Without Docker

For local development, you can run tests directly on your machine:

```bash
./run_local_tests.sh
```

Note: This requires Go, Node.js, and MySQL to be installed and properly configured on your machine.

## Test Structure

### Backend Tests

Backend tests are located in `backend/internal/tests/` and are organized by feature:

- `auth_test.go`: Tests for authentication endpoints
- `market_test.go`: Tests for market functionality
- `integration_test.go`: Database integration tests

The tests use a dedicated test database (`officestonks_test`) with isolated test data.

### Frontend Tests

Frontend tests are located alongside their components with the `.test.js` extension:

- `src/components/Navigation.test.js`: Tests for the Navigation component
- `src/pages/Login.test.js`: Tests for the Login page
- `src/services/stock.test.js`: Tests for the stock service

We use Jest and React Testing Library for frontend testing.

## Writing New Tests

### Backend

1. Add new test files to `backend/internal/tests/`
2. Use the provided test helper functions in `test_helpers.go`
3. Follow the pattern of existing tests

Example:

```go
func TestMyNewFeature(t *testing.T) {
    // Skip if no test database connection
    if TestDB == nil {
        t.Skip("No test database connection")
    }

    // Setup test router
    router := SetupTestRouter(TestDB)

    // Make request
    rr := MakeRequest("GET", "/api/my-endpoint", nil, router)

    // Check status code
    if rr.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
    }

    // Further assertions...
}
```

### Frontend

1. Create a `.test.js` file alongside the component or service you're testing
2. Use React Testing Library to test components
3. Use Jest to mock dependencies

Example:

```javascript
import { render, screen } from '@testing-library/react';
import MyComponent from './MyComponent';

test('renders my component correctly', () => {
  render(<MyComponent />);
  expect(screen.getByText('Expected Text')).toBeInTheDocument();
});
```

## Continuous Integration

In a real CI/CD setup, you would:
1. Run `./run_tests.sh` on every pull request
2. Block merging if tests fail
3. Generate test coverage reports

## Best Practices

1. Write tests for critical functionality first
2. Aim for high coverage in core services and data models
3. Test edge cases and error handling
4. Keep tests independent (no test should depend on another test)
5. Use setup and teardown functions to maintain a clean test environment
6. Mock external services and dependencies
7. Organize tests logically by feature or component