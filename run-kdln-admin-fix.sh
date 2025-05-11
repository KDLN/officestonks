#!/bin/bash

# Script to ensure KDLN user has admin privileges in the database
# This connects to the database configured for the application
# and runs the SQL script to update the user

# Set bold text and colors for better readability
BOLD='\033[1m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}=== KDLN Admin Fix Script ===${NC}"

# Load environment variables from env file if it exists
if [ -f ".env" ]; then
  echo -e "${YELLOW}Loading environment variables from .env${NC}"
  export $(grep -v '^#' .env | xargs)
fi

# Use environment variables for database connection
DB_HOST=${MYSQLHOST:-"localhost"}
DB_PORT=${MYSQLPORT:-"3306"}
DB_USER=${MYSQLUSER:-"root"}
DB_PASS=${MYSQLPASSWORD:-"password"}
DB_NAME=${MYSQLDATABASE:-"officestonks"}

echo -e "${BLUE}Connecting to MySQL database...${NC}"
echo -e "Host: ${DB_HOST}:${DB_PORT}"
echo -e "Database: ${DB_NAME}"
echo -e "User: ${DB_USER}"

# Run the SQL script
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < ensure-kdln-admin.sql

if [ $? -eq 0 ]; then
  echo -e "${GREEN}KDLN admin fix script executed successfully!${NC}"
else
  echo -e "${RED}Error executing KDLN admin fix script.${NC}"
  exit 1
fi

echo -e "${BLUE}Deployment KDLN admin fix script...${NC}"

# For Railway deployments, also try using the Railway MySQL connection details
if [ -n "$RAILWAY_STATIC_URL" ]; then
  echo -e "${YELLOW}Detected Railway environment, using Railway MySQL connection...${NC}"
  
  # Extract connection details from RAILWAY_STATIC_URL if it exists
  DB_HOST=$(echo $RAILWAY_STATIC_URL | awk -F/ '{print $3}' | awk -F: '{print $1}')
  DB_PORT=$(echo $RAILWAY_STATIC_URL | awk -F/ '{print $3}' | awk -F: '{print $2}')
  DB_USER=$(echo $RAILWAY_STATIC_URL | awk -F/ '{print $3}' | awk -F@ '{print $1}')
  DB_PASS=$(echo $RAILWAY_STATIC_URL | awk -F/ '{print $3}' | awk -F: '{print $2}')
  DB_NAME=$(echo $RAILWAY_STATIC_URL | awk -F/ '{print $4}')
  
  echo -e "Railway Host: ${DB_HOST}:${DB_PORT}"
  echo -e "Railway Database: ${DB_NAME}"
  
  # Run the SQL script with Railway connection
  mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < ensure-kdln-admin.sql
  
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}Railway KDLN admin fix script executed successfully!${NC}"
  else
    echo -e "${RED}Error executing Railway KDLN admin fix script.${NC}"
  fi
fi

echo -e "${BOLD}${GREEN}KDLN admin fix complete.${NC}"