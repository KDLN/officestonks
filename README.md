# Office Stonks - Multiplayer Stock Market Game

A real-time multiplayer stock market simulation game where players can trade stocks, form investment groups, and compete for the highest portfolio value.

## Overview

Office Stonks is an online multiplayer stock market simulation that allows players to:
- Buy and sell virtual stocks based on real market dynamics
- See real-time price updates via WebSockets
- Compete on leaderboards with other players
- Chat with other players
- Manage their portfolios and view transaction history

## Tech Stack

- **Backend**: Go with standard library
- **Frontend**: React with a simple component library
- **Database**: MySQL
- **Hosting**: Railway
- **Real-time Updates**: WebSockets

## Project Structure

The repository follows standard Go project layout:

- `/cmd/api`: Application entry point
- `/internal`: Internal packages (models, handlers, etc.)
  - `/auth`: Authentication utilities
  - `/handlers`: HTTP route handlers
  - `/middleware`: HTTP middleware
  - `/models`: Data models
  - `/repository`: Database access
  - `/services`: Business logic
  - `/websocket`: WebSocket handling
- `/pkg`: Shared packages
  - `/market`: Market simulation logic
- `/frontend`: React frontend
  - `/src`: Source code
  - `/public`: Static assets
- `/cors-proxy`: CORS proxy for API requests
- `/docs`: Consolidated documentation

## Documentation

All project documentation is organized in the `/docs` directory:

- [Project Documentation](docs/PROJECT.md) - Overview, structure, and development guidelines
- [Deployment Guide](docs/DEPLOYMENT.md) - Complete deployment instructions
- [Cleanup Summary](docs/CLEANUP.md) - Repository cleanup and organization

## Getting Started

### Prerequisites
- Git
- Go 1.20+
- Node.js 18+
- MySQL database

### Local Development

1. Clone this repository
2. Start the backend:
   ```
   go run cmd/api/main.go
   ```
3. Start the frontend:
   ```
   cd frontend
   npm install
   npm start
   ```

### Docker Development Environment
```bash
cd docker
docker-compose up
```

## Deployment

This application is deployed on Railway. See [Deployment Guide](docs/DEPLOYMENT.md) for detailed deployment instructions.

## Features

### Stock Trading
- Buy and sell virtual stocks based on simulated market dynamics
- Real-time price updates via WebSockets
- Portfolio management with transaction history

### User System
- Registration and authentication
- JWT-based secure login
- Cash balance management
- Admin privileges for system management

### Community Features
- Real-time chat system
- Leaderboard showing top performing traders
- Portfolio comparisons

### Admin Panel
- User management
- System status monitoring
- Market reset capabilities

## Contributing

We welcome contributions! Please see [Contributing Guidelines](docs/PROJECT.md#contributing) for details.

## License

This project is licensed under the MIT License - see the LICENSE file for details.