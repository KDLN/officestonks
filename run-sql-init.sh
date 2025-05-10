#!/bin/bash

# This script helps you initialize the database with an admin user and stocks
# It connects to your public MySQL endpoint and runs the init-admin-user.sql script

# Database credentials
DB_HOST="caboose.proxy.rlwy.net"
DB_PORT="40558"
DB_USER="root"
DB_PASS="DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE"
DB_NAME="railway"

echo "Initializing database tables and admin user..."
echo "Connecting to MySQL at $DB_HOST:$DB_PORT..."

# Connect to MySQL and run the script
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < init-admin-user.sql

if [ $? -eq 0 ]; then
    echo "Database initialization completed successfully!"
    echo "You should now be able to log in with:"
    echo "  Username: admin"
    echo "  Password: admin123"
else
    echo "Error: Database initialization failed."
    echo "Please check your connection details and try again."
fi