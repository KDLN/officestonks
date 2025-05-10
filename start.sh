#!/bin/sh
# Simple startup script - redirects to the main start-server.sh
echo "Starting container with start.sh"
# Use the current directory instead of /app
ls -la ./
exec ./start-server.sh