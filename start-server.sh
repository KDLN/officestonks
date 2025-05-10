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
echo "Checking if server binary exists..."
if [ -f "/app/bin/server" ]; then
  echo "✅ Server binary found at /app/bin/server"
  exec /app/bin/server
elif [ -f "./bin/server" ]; then
  echo "✅ Server binary found at ./bin/server"
  exec ./bin/server
elif [ -f "./server" ]; then
  echo "✅ Server binary found at ./server"
  exec ./server
else
  echo "❌ ERROR: Server binary not found!"
  echo "Available files in current directory:"
  ls -la
  echo "Available files in /app:"
  ls -la /app || echo "Cannot access /app"
  echo "Available files in /app/bin:"
  ls -la /app/bin || echo "Cannot access /app/bin"
  exit 1
fi
