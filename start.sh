#!/bin/bash

# Office Stonks - Start both Go backend and Socket.IO server in single service

echo "🚀 Starting Office Stonks with integrated Socket.IO..."

# Debug: Show environment
echo "📍 Current directory: $(pwd)"
echo "📁 Directory contents:"
ls -la

# Check if Socket.IO server directory exists
if [ ! -d "/app/socketio-server" ]; then
    echo "❌ ERROR: Socket.IO server directory not found at /app/socketio-server"
    echo "📁 Available directories:"
    find /app -type d -maxdepth 2
    exit 1
fi

# Install Node.js dependencies for Socket.IO
echo "📦 Installing Socket.IO dependencies..."
cd /app/socketio-server
if [ ! -f "package.json" ]; then
    echo "❌ ERROR: package.json not found in Socket.IO directory"
    echo "📁 Socket.IO directory contents:"
    ls -la
    exit 1
fi

# Install dependencies
npm ci --omit=dev || npm install --omit=dev

# Set Socket.IO environment variables
export NODE_ENV=production
export JWT_SECRET=${JWT_SECRET:-"your-secret-key-change-in-production"}

# Start Socket.IO server in background with logging
echo "📡 Starting Socket.IO server on port 3001..."
node server.js 2>&1 | sed 's/^/[Socket.IO] /' &
SOCKETIO_PID=$!

# Wait for Socket.IO to start and verify it's running
echo "⏳ Waiting for Socket.IO server to start..."
for i in {1..10}; do
    if nc -z localhost 3001 2>/dev/null; then
        echo "✅ Socket.IO server is running on port 3001"
        break
    fi
    if [ $i -eq 10 ]; then
        echo "❌ ERROR: Socket.IO server failed to start on port 3001"
        echo "📋 Socket.IO process status:"
        ps aux | grep node
        exit 1
    fi
    sleep 1
done

# Start Go backend on main port
echo "⚡ Starting Go backend on port $PORT..."
cd /app
./out &
GO_PID=$!

# Monitor both processes
echo "✅ Both services started. Monitoring..."
while true; do
    # Check if Go backend is still running
    if ! kill -0 $GO_PID 2>/dev/null; then
        echo "❌ Go backend stopped"
        kill $SOCKETIO_PID 2>/dev/null || true
        exit 1
    fi
    
    # Check if Socket.IO is still running
    if ! kill -0 $SOCKETIO_PID 2>/dev/null; then
        echo "❌ Socket.IO server stopped, restarting..."
        cd /app/socketio-server
        node server.js 2>&1 | sed 's/^/[Socket.IO] /' &
        SOCKETIO_PID=$!
        sleep 2
    fi
    
    sleep 5
done