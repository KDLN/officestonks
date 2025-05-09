# Office Stonks - Multiplayer Stock Market Game

A real-time multiplayer stock market simulation game where players can trade stocks, form investment groups, and compete for the highest portfolio value.

## Overview

Office Stonks is an online multiplayer stock market simulation that allows players to:
- Buy and sell virtual stocks based on real market dynamics
- See real-time price updates via WebSockets
- Compete on leaderboards with other players
- Chat with other players

## Tech Stack

- **Backend**: Go with standard library/Gorilla Mux
- **Frontend**: React with a simple component library
- **Database**: MySQL
- **Hosting**: Railway
- **Real-time Updates**: WebSockets

## Project Structure

- `/backend`: Go server code
  - `/cmd/api`: Application entry point
  - `/internal`: Internal packages (models, handlers, etc.)
  - `/pkg`: Shared packages
- `/frontend`: React frontend
  - `/src`: Source code
  - `/public`: Static assets
- `/docker`: Docker configuration files
  - `Dockerfile.backend`: Backend production container
  - `Dockerfile.frontend`: Frontend production container
  - `docker-compose.yml`: Local development setup
  - `docker-compose.test.yml`: Testing environment

## Getting Started

### Prerequisites
- Git
- GitHub account
- Railway account

### Deployment
See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

### Local Development

1. Clone this repository
2. Start backend:
   ```
   cd backend
   go run cmd/api/main.go
   ```
3. Start frontend:
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

## Testing

See [TESTING.md](TESTING.md) for information on running tests.

## MVP Plan

Our MVP approach includes:

1. **Core Features**:
   - Stock trading system
   - User portfolios
   - Real-time price updates
   - Simple leaderboard

2. **Development Phases**:
   - Backend foundation
   - Trading and real-time features
   - User experience and interface
   - Testing and deployment

For more details, see [STEP_BY_STEP_MVP.md](STEP_BY_STEP_MVP.md).

## Contributing

We welcome contributions! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.