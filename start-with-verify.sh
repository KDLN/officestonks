#!/bin/bash

# Script to run database verification before starting the application

echo "Starting verification and application launch process..."

# Run comprehensive database diagnostics
if [ -f /app/diagnose-db.sh ]; then
    echo "Running comprehensive database diagnostics..."
    /bin/bash /app/diagnose-db.sh > /app/db-diagnostics.log 2>&1
    echo "Diagnostics saved to /app/db-diagnostics.log"
    cat /app/db-diagnostics.log
else
    echo "Database diagnostics script not found. Falling back to basic verification."

    # Run basic database verification
    if [ -f /app/verify-db.sh ]; then
        echo "Running database verification..."
        /bin/bash /app/verify-db.sh
    else
        echo "Database verification script not found. Skipping verification."
    fi
fi

# Print environment variables (excluding sensitive ones)
echo "Environment variables:"
env | grep -v -E 'PASSWORD|SECRET|KEY' | sort

# Start the application
echo "Starting application..."
exec /app/server