#!/bin/sh

echo "=== DATABASE INITIALIZATION ==="
echo "Starting database initialization..."

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

# Set schema paths
SCHEMA_PATH="/app/schema.sql"
UPDATE_SCHEMA_PATH="/app/update-schema.sql"

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
  echo "✅ Database tables already exist, skipping full initialization."
  echo "🔄 Applying schema updates..."

  # Always run the update schema for existing databases
  if [ -f "$UPDATE_SCHEMA_PATH" ]; then
    echo "Found update schema at $UPDATE_SCHEMA_PATH"
    if eval $MYSQL_CMD "$DB_NAME" < $UPDATE_SCHEMA_PATH; then
      echo "✅ Schema updates applied successfully."
    else
      echo "⚠️ Schema updates encountered issues."
    fi
  elif [ -f "./update-schema.sql" ]; then
    echo "Found update schema in current directory"
    if eval $MYSQL_CMD "$DB_NAME" < ./update-schema.sql; then
      echo "✅ Schema updates applied successfully."
    else
      echo "⚠️ Schema updates encountered issues."
    fi
  else
    echo "⚠️ Update schema file not found."
  fi
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

# We now handle admin schema updates directly in the update-schema.sql file

echo "✅ Database initialization complete."