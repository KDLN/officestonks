#!/bin/bash
# Execute the fix-admin-user.sql script to fix admin issues

# Database credentials
DB_HOST="caboose.proxy.rlwy.net"
DB_PORT="40558"
DB_USER="root"
DB_PASS="DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE"
DB_NAME="railway"

# Log start
echo "Starting database fixes for OfficeStonks..."
echo "Connecting to MySQL at $DB_HOST:$DB_PORT as $DB_USER..."

# Run the SQL script
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" "$DB_NAME" < fix-admin-user.sql

# Check the exit status
if [ $? -eq 0 ]; then
    echo "============================================"
    echo "Database fixes completed successfully!"
    echo "The following actions were performed:"
    echo "1. Admin user was created or updated"
    echo "2. Stock prices were reset"
    echo "3. Chat messages were cleared"
    echo "4. All required tables were created/verified"
    echo "============================================"
    echo "You should now be able to log in with:"
    echo "  Username: admin"
    echo "  Password: admin123"
    echo "============================================"
else
    echo "Error: Database fixes failed."
    echo "Please check your connection details and try again."
fi