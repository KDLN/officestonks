#!/bin/sh
set -e

# Log environment variables (excluding passwords)
echo "Starting OfficeStonks server..."
echo "Environment:"
echo "  PORT=${PORT}"
echo "  DB_HOST=${DB_HOST}"
echo "  DB_PORT=${DB_PORT}"
echo "  DB_USER=${DB_USER}"
echo "  DB_NAME=${DB_NAME}"
echo "  FRONTEND_URL=${FRONTEND_URL}"

# Start the server
exec ./server