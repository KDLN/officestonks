#!/bin/sh
# Simple startup script - redirects to the main start-server.sh
echo "Starting container with start.sh"
ls -la /app/
exec /app/start-server.sh