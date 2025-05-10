#!/bin/sh
# Print some debug info about the environment
echo "Starting server..."
echo "Environment: $(env | grep -v PASSWORD)"

# Initialize the database if mysql-client is available
if command -v mysql >/dev/null 2>&1; then
  echo "MySQL client found, initializing database..."

  # Get database credentials from environment variables - prioritize Railway variables
  DB_HOST=${MYSQLHOST:-${DB_HOST:-"mysql.railway.internal"}}
  DB_PORT=${MYSQLPORT:-${DB_PORT:-"3306"}}
  DB_USER=${MYSQLUSER:-${DB_USER:-"root"}}
  DB_PASSWORD=${MYSQLPASSWORD:-${DB_PASSWORD:-""}}
  DB_NAME=${MYSQLDATABASE:-${MYSQL_DATABASE:-${DB_NAME:-"railway"}}}

  echo "Using connection parameters:"
  echo "  Host: $DB_HOST"
  echo "  Port: $DB_PORT"
  echo "  User: $DB_USER"
  echo "  Database: $DB_NAME"
  echo "  (Password hidden for security)"

  # List all available environment variables for debugging
  echo "Available environment variables:"
  env | grep -E "MYSQL|DB_" | grep -v PASSWORD | grep -v SECRET

  # Set schema path
  SCHEMA_PATH="/app/schema.sql"

  # For Railway, we need to use the most basic connection parameters
  echo "Using basic connection parameters for maximum compatibility"
  MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\""

  # Test the connection
  echo "Testing database connection..."
  if eval $MYSQL_CMD -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
    echo "✅ Database connection successful on first try."
  else
    echo "⚠️ Initial connection failed, proceeding with database setup anyway."
    echo "The application's built-in database driver may still connect successfully."
  fi

  # Check if tables already exist
  echo "Checking if database tables already exist..."
  TABLES=$(eval $MYSQL_CMD -e "SHOW TABLES" "$DB_NAME" 2>/dev/null || echo "")
  echo "Current tables in database: $TABLES"

  # Attempt to create tables if needed
  if echo "$TABLES" | grep -q "stocks"; then
    echo "✅ Database tables already exist, skipping initialization."
  else
    echo "🔄 Initializing database schema..."
    if eval $MYSQL_CMD "$DB_NAME" < $SCHEMA_PATH; then
      echo "✅ Database schema initialized successfully."
      echo "Tables created:"
      eval $MYSQL_CMD -e "SHOW TABLES" "$DB_NAME" 2>/dev/null || echo "Failed to show tables"
    else
      echo "⚠️ Schema initialization encountered issues."
      echo "The application will attempt to create necessary tables on startup."
    fi
  fi

  echo "✅ Database initialization complete."
else
  echo "MySQL client not available, skipping database initialization."
  echo "IMPORTANT: You need to manually initialize the database using schema.sql"
fi

# Execute the server binary
echo "Starting the API server..."
exec /app/bin/server