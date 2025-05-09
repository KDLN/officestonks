#!/bin/bash

# Exit on any error
set -e

echo "=== Starting Test Environment ==="
# Build and start the test environment
docker-compose -f docker/docker-compose.test.yml build
docker-compose -f docker/docker-compose.test.yml up -d mysql-test

# Wait for MySQL to be ready
echo "=== Waiting for MySQL to be ready ==="
sleep 10

echo "=== Running Backend Tests ==="
docker-compose -f docker/docker-compose.test.yml run --rm backend-test

echo "=== Running Frontend Tests ==="
docker-compose -f docker/docker-compose.test.yml run --rm frontend-test

echo "=== Cleaning Up ==="
docker-compose -f docker/docker-compose.test.yml down

echo "=== All tests completed! ==="