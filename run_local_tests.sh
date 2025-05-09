#!/bin/bash

# This script is for running tests locally without Docker

echo "=== Running Backend Tests ==="
echo "Note: To run backend tests locally, you need Go installed and a MySQL database."
echo "If you don't have these, use 'run_tests.sh' with Docker instead."
cd backend
if command -v go &> /dev/null; then
    go test -v ./internal/tests/...
else
    echo "Go not found. Skipping backend tests."
fi

echo "=== Running Frontend Tests ==="
cd ../frontend
if command -v npm &> /dev/null; then
    npm run test:ci
else
    echo "npm not found. Skipping frontend tests."
fi

echo "=== All tests completed! ==="