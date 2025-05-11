#!/bin/bash

# Script to verify database connection before starting the application

# Color codes for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Verifying database connection...${NC}"

# Check if mysql client is installed
if ! command -v mysql &> /dev/null; then
    echo -e "${RED}MySQL client not found. Skipping verification.${NC}"
    exit 0
fi

# Get database parameters from environment variables or use defaults
DB_HOST=${MYSQLHOST:-turntable.proxy.rlwy.net}
DB_PORT=${MYSQLPORT:-28889}
DB_USER=${MYSQLUSER:-root}
DB_PASS=${MYSQLPASSWORD:-EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY}
DB_NAME=${MYSQLDATABASE:-railway}

echo "Connecting to MySQL at ${DB_HOST}:${DB_PORT} as ${DB_USER} to database ${DB_NAME}..."

# First, check if port is open with timeout
echo "Checking if MySQL port is open..."
if nc -z -w 5 ${DB_HOST} ${DB_PORT}; then
    echo -e "${GREEN}MySQL port ${DB_PORT} is open on ${DB_HOST}${NC}"
else
    echo -e "${RED}MySQL port ${DB_PORT} is closed on ${DB_HOST}${NC}"
    echo "Will try connection anyway..."
fi

# Try MySQL connection with timeout
MYSQL_CMD="mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p${DB_PASS} -D ${DB_NAME} --connect-timeout=10 -e 'SELECT VERSION();'"
echo "Executing MySQL connection test..."

if eval ${MYSQL_CMD} &> /dev/null; then
    echo -e "${GREEN}Successfully connected to MySQL database!${NC}"
    echo "Checking database tables..."
    
    # Check tables
    TABLES=$(mysql -h ${DB_HOST} -P ${DB_PORT} -u ${DB_USER} -p${DB_PASS} -D ${DB_NAME} --connect-timeout=10 -e 'SHOW TABLES;' 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Tables found:${NC}"
        echo "$TABLES"
    else
        echo -e "${RED}Failed to retrieve tables.${NC}"
    fi
else
    echo -e "${RED}Failed to connect to MySQL database.${NC}"
    echo "Application will try to connect using Go database driver..."
fi

echo -e "${YELLOW}Database verification completed.${NC}"
exit 0