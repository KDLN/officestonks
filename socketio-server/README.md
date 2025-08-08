# Office Stonks Socket.IO Server

This is the official Socket.IO v4 server for Office Stonks real-time communication.

## Why a Separate Socket.IO Server?

- **Full Compatibility**: Uses the official Socket.IO v4 implementation
- **Railway Optimized**: Works perfectly with Railway's infrastructure
- **Language Agnostic**: Separates real-time concerns from business logic
- **Scalable**: Can be scaled independently from the main Go backend

## Features

- ✅ WebSocket with automatic polling fallback
- ✅ JWT authentication
- ✅ Real-time stock price updates
- ✅ Chat messaging system
- ✅ Connection monitoring and admin stats
- ✅ Automatic reconnection handling

## Local Development

1. Install dependencies:
```bash
npm install
```

2. Create `.env` file:
```bash
cp .env.example .env
```

3. Update `.env` with your configuration:
```env
PORT=3001
CORS_ORIGIN=http://localhost:3000
JWT_SECRET=your-secret-key
GO_BACKEND_URL=http://localhost:8080
```

4. Start the server:
```bash
npm start
# or for development with auto-reload
npm run dev
```

## Railway Deployment

### Option 1: Deploy as Separate Service (Recommended)

1. Create a new service in Railway
2. Connect this `socketio-server` directory
3. Set environment variables:
   - `CORS_ORIGIN`: Your frontend URL
   - `JWT_SECRET`: Same as your Go backend
   - `GO_BACKEND_URL`: Internal Railway URL of your Go backend
4. Railway will automatically detect Node.js and deploy

### Option 2: Use with Docker

```bash
docker build -t officestonks-socketio .
docker run -p 3001:3001 --env-file .env officestonks-socketio
```

## Frontend Configuration

Update your frontend to connect to the Socket.IO server:

```javascript
// In your .env
REACT_APP_SOCKETIO_URL=https://your-socketio-service.railway.app
```

Or if deployed as subdomain:
```
https://socketio.officestonks.com
```

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Frontend  │────▶│  Socket.IO   │────▶│ Go Backend  │
│   (React)   │◀────│   Server     │◀────│   (API)     │
└─────────────┘     └──────────────┘     └─────────────┘
     WSS/HTTPS         Node.js              HTTP/Internal
```

## API Endpoints

- `/health` - Health check
- `/admin/socketio/stats` - Connection statistics

## Socket.IO Events

### Client → Server
- `subscribe_stocks` - Subscribe to stock updates
- `join_chat` - Join chat room
- `chat_message` - Send chat message
- `ping` - Connection test

### Server → Client
- `connected` - Connection confirmation
- `stock_update` - Stock price updates
- `chat_message` - Chat messages
- `pong` - Ping response

## Monitoring

Access connection stats:
```
GET /admin/socketio/stats
```

Returns:
```json
{
  "connected_clients": 5,
  "clients": [...],
  "server_info": {...}
}
```

## Troubleshooting

1. **Connection Issues**: Check CORS_ORIGIN matches your frontend URL
2. **Auth Errors**: Ensure JWT_SECRET matches your Go backend
3. **No Stock Updates**: Verify GO_BACKEND_URL is accessible
4. **Railway Issues**: Ensure PORT env var is not hardcoded

## Benefits Over Go Socket.IO Libraries

1. **Official Implementation**: Full Socket.IO v4 protocol support
2. **Better Compatibility**: Works with all Socket.IO clients
3. **Active Maintenance**: Regular updates and bug fixes
4. **Railway Native**: Node.js + Socket.IO is well-tested on Railway
5. **Simpler Integration**: No Go dependency conflicts