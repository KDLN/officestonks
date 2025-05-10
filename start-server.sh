#!/bin/sh
# Print some debug info about the environment
echo "Starting server from start-server.sh..."
echo "Current directory: $(pwd)"
echo "Files in current directory: $(ls -la)"
echo "Environment: $(env | grep -v PASSWORD)"

# Initialize the database if mysql-client is available
if command -v mysql >/dev/null 2>&1; then
  echo "MySQL client found, initializing database..."
  sh /app/init-db.sh
else
  echo "MySQL client not available, skipping database initialization."
  echo "IMPORTANT: You need to manually initialize the database using schema.sql"
fi

# Execute the server binary
echo "Starting the API server..."
exec /app/bin/server
