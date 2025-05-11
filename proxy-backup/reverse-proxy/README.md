# OfficeStonks Reverse Proxy

This is a reverse proxy service for OfficeStonks that routes requests to the appropriate backend or frontend service.

## How it Works

1. **API Requests** (`/api/*`): Routed to the backend service
2. **WebSocket Connections** (`/ws`): Routed to the backend service
3. **All Other Requests**: Routed to the frontend service

## Deployment on Railway

1. Create a new service in your Railway project
2. Use this repository as the source
3. Set the following environment variables:
   - `BACKEND_URL` = URL of your backend service
   - `FRONTEND_URL` = URL of your frontend service
4. Deploy the service

## Configuration Options

You can configure the proxy through environment variables:

- `PORT`: The port to listen on (default: 3000)
- `BACKEND_URL`: The URL of your backend service
- `FRONTEND_URL`: The URL of your frontend service

## Local Development

To run the proxy locally:

```
npm install
npm start
```

## Benefits

- Eliminates CORS issues by serving everything from a single domain
- Allows separate deployment and scaling of frontend and backend
- Provides a unified entry point for your application

## Architecture

```
                  ┌─────────────────┐
                  │                 │
 User Request ───►│  Reverse Proxy  │
                  │                 │
                  └────────┬────────┘
                           │
                           ▼
         ┌─────────────────┴─────────────────┐
         │                                   │
         ▼                                   ▼
┌─────────────────┐                 ┌─────────────────┐
│                 │                 │                 │
│    Backend      │                 │    Frontend     │
│    Service      │                 │    Service      │
│                 │                 │                 │
└─────────────────┘                 └─────────────────┘
```