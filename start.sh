#!/bin/bash

# Office Stonks - Start both Go backend and Socket.IO server in single service

echo "🚀 Starting Office Stonks with integrated Socket.IO..."

# Install Node.js dependencies for Socket.IO
echo "📦 Installing Socket.IO dependencies..."
cd /app/socketio-server && npm ci --omit=dev

# Start Socket.IO server in background
echo "📡 Starting Socket.IO server on port 3001..."
cd /app/socketio-server && npm start &
SOCKETIO_PID=$!

# Wait a moment for Socket.IO to start
sleep 3

# Start Go backend on main port
echo "⚡ Starting Go backend on port $PORT..."
cd /app && ./out

# If Go backend stops, cleanup Socket.IO
kill $SOCKETIO_PID 2>/dev/null || true
echo "🔄 Services stopped"