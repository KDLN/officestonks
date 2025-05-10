#!/bin/sh
# Print some debug info about the environment
echo "Starting server from start-server.sh..."
echo "Current directory: $(pwd)"
echo "Files in current directory: $(ls -la)"
echo "Environment: $(env | grep -v PASSWORD)"

# Initialize the database if mysql-client is available
if command -v mysql >/dev/null 2>&1; then
  echo "MySQL client found, initializing database..."
  if [ -f "/app/init-db.sh" ]; then
    sh /app/init-db.sh
  elif [ -f "./init-db.sh" ]; then
    sh ./init-db.sh
  else
    echo "❌ WARNING: init-db.sh not found. Using SQL script directly."
    # Try to use the schema.sql directly
    if [ -f "./schema.sql" ]; then
      echo "Found schema.sql, checking MySQL connection..."
      # Try a simple connection test before running the script
      MYSQL_HOST="${DB_HOST:-localhost}"
      MYSQL_PORT="${DB_PORT:-3306}"
      MYSQL_USER="${DB_USER:-officestonks}"
      MYSQL_PASS="${DB_PASSWORD:-officestonks}"
      MYSQL_DB="${DB_NAME:-officestonks}"

      echo "Connecting to MySQL: ${MYSQL_HOST}:${MYSQL_PORT}..."
      mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASS" \
        -e "SELECT 'MySQL connection successful!'" || echo "MySQL connection failed"

      echo "Initializing database from schema.sql..."
      mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASS" "$MYSQL_DB" < ./schema.sql \
        && echo "Database initialized successfully" || echo "Failed to initialize database"
    fi
  fi
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
