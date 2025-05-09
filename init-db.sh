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

# MySQL/MariaDB command with SSL disabled
# Check mysql version and use compatible SSL options
echo "Detecting database client version..."
if mysql --version 2>&1 | grep -q -i "mariadb"; then
  echo "Detected MariaDB client, using compatible options"
  # MariaDB client options - use --ssl=0 instead of --ssl-mode
  MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\" --ssl=0"
else
  echo "Detected MySQL client, using compatible options"
  # MySQL client options
  MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\" --ssl=0"
fi

# Wait for MySQL to be ready
MAX_RETRIES=30
RETRY_INTERVAL=3
RETRIES=0

while [ $RETRIES -lt $MAX_RETRIES ]; do
  echo "Attempting to connect to MySQL (try $((RETRIES+1))/$MAX_RETRIES)..."
  if eval $MYSQL_CMD -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
    echo "✅ Database connection successful."
    break
  else
    CONNECTION_ERROR=$(eval $MYSQL_CMD -e "SELECT 1" "$DB_NAME" 2>&1 | grep -v "mysql: \[Warning\]")
    echo "❌ Failed to connect: $CONNECTION_ERROR"
    echo "Waiting for database to be ready... Attempt $((RETRIES+1))/$MAX_RETRIES"
    RETRIES=$((RETRIES+1))
    sleep $RETRY_INTERVAL
  fi
done

if [ $RETRIES -eq $MAX_RETRIES ]; then
  echo "❌ Failed to connect to database after $MAX_RETRIES attempts."
  echo "Last error: $CONNECTION_ERROR"

  # Try with different SSL options as a fallback
  echo "Attempting to connect with alternative SSL settings..."

  # Try without any SSL options
  if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
    echo "✅ Connection successful with no SSL options"
    MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\""
  # Try with --skip-ssl (might work on some MySQL versions)
  elif mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" --skip-ssl -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
    echo "✅ Connection successful with --skip-ssl"
    MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\" --skip-ssl"
  else
    echo "⚠️ All connection attempts failed, but will continue anyway"
    # Don't exit, let the application try to connect
    # It might have more success with the Go MySQL driver
    MYSQL_CMD="mysql -h \"$DB_HOST\" -P \"$DB_PORT\" -u \"$DB_USER\" -p\"$DB_PASSWORD\""
  fi
fi

# Check if tables already exist
echo "Checking if database tables already exist..."
TABLES=$(eval $MYSQL_CMD -e "SHOW TABLES" "$DB_NAME" 2>/dev/null)
echo "Current tables in database:"
echo "$TABLES"

TABLES_EXIST=$(echo "$TABLES" | grep -c "stocks")

if [ "$TABLES_EXIST" -gt 0 ]; then
  echo "✅ Database tables already exist, skipping initialization."
else
  echo "🔄 Initializing database schema..."
  eval $MYSQL_CMD "$DB_NAME" < /app/schema.sql
  if [ $? -eq 0 ]; then
    echo "✅ Database schema initialized successfully."
    echo "Tables created:"
    eval $MYSQL_CMD -e "SHOW TABLES" "$DB_NAME" 2>/dev/null
  else
    ERROR_MSG=$(eval $MYSQL_CMD "$DB_NAME" < /app/schema.sql 2>&1)
    echo "❌ Failed to initialize database schema."
    echo "Error: $ERROR_MSG"
    exit 1
  fi
fi

echo "✅ Database initialization complete."