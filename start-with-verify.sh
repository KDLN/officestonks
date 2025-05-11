#!/bin/bash

# Script to run database verification before starting the application

echo "Starting verification and application launch process..."

# Run database verification
if [ -f /app/verify-db.sh ]; then
    echo "Running database verification..."
    /bin/bash /app/verify-db.sh
else
    echo "Database verification script not found. Skipping verification."
fi

# Print environment variables (excluding sensitive ones)
echo "Environment variables:"
env | grep -v -E 'PASSWORD|SECRET|KEY' | sort

# Start the application
echo "Starting application..."
exec /app/server