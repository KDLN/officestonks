#!/bin/sh
# Print some debug info about the environment
echo "Starting server..."
echo "Environment: $(env | grep -v PASSWORD)"

# Execute the server binary
exec /app/bin/server