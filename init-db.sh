#!/bin/sh

echo "=== DATABASE INITIALIZATION ==="
echo "Starting database initialization..."
echo "Using connection parameters:"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"
echo "  (DB_PASSWORD is hidden for security)"

# Wait for MySQL to be ready
MAX_RETRIES=30
RETRY_INTERVAL=3
RETRIES=0

while [ $RETRIES -lt $MAX_RETRIES ]; do
  echo "Attempting to connect to MySQL (try $((RETRIES+1))/$MAX_RETRIES)..."
  if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
    echo "✅ Database connection successful."
    break
  else
    CONNECTION_ERROR=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" "$DB_NAME" 2>&1 | grep -v "mysql: \[Warning\]")
    echo "❌ Failed to connect: $CONNECTION_ERROR"
    echo "Waiting for database to be ready... Attempt $((RETRIES+1))/$MAX_RETRIES"
    RETRIES=$((RETRIES+1))
    sleep $RETRY_INTERVAL
  fi
done

if [ $RETRIES -eq $MAX_RETRIES ]; then
  echo "❌ Failed to connect to database after $MAX_RETRIES attempts."
  echo "Last error: $CONNECTION_ERROR"
  exit 1
fi

# Check if tables already exist
echo "Checking if database tables already exist..."
TABLES=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SHOW TABLES" "$DB_NAME" 2>/dev/null)
echo "Current tables in database:"
echo "$TABLES"

TABLES_EXIST=$(echo "$TABLES" | grep -c "stocks")

if [ "$TABLES_EXIST" -gt 0 ]; then
  echo "✅ Database tables already exist, skipping initialization."
else
  echo "🔄 Initializing database schema..."
  mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < /app/schema.sql
  if [ $? -eq 0 ]; then
    echo "✅ Database schema initialized successfully."
    echo "Tables created:"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SHOW TABLES" "$DB_NAME" 2>/dev/null
  else
    ERROR_MSG=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < /app/schema.sql 2>&1)
    echo "❌ Failed to initialize database schema."
    echo "Error: $ERROR_MSG"
    exit 1
  fi
fi

echo "✅ Database initialization complete."