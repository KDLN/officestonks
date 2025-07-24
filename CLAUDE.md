# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Backend (Go)
```bash
# Run the server locally
go run cmd/api/main.go

# Build the server binary
go build -o server ./cmd/api/main.go

# Run with Docker
cd docker && docker-compose up
```

### Frontend (React)
```bash
# Install dependencies and start development server
cd frontend && npm install && npm start

# Build for production
cd frontend && npm run build

# Run tests
cd frontend && npm test

# Run tests in CI mode (no watch)
cd frontend && npm test:ci
```

### CORS Proxy
```bash
# Start CORS proxy (required for local development)
cd cors-proxy && npm install && npm start

# Development mode with auto-reload
cd cors-proxy && npm run dev
```

### Testing
```bash
# Run full test suite with Docker
./run_tests.sh

# Run local tests
./run_local_tests.sh

# Frontend tests only
cd frontend && npm test
```

### Database
The application uses Railway MySQL with environment variables:
- `MYSQLHOST`, `MYSQLPORT`, `MYSQLUSER`, `MYSQLPASSWORD`, `MYSQLDATABASE`
- Database schema is automatically applied on startup from `schema.sql`

## High-Level Architecture

Office Stonks is a real-time multiplayer stock market simulation game with two main components:

### 1. Backend API (Go)
- **Entry Point**: `cmd/api/main.go`
- **Core Logic**: `/internal/` directory following Go conventions
  - `handlers/` - HTTP route handlers for all API endpoints
  - `services/` - Business logic layer (game, market, chat services)
  - `repository/` - Database access layer with environment variable configuration
  - `websocket/` - Real-time WebSocket hub for stock updates and chat
  - `middleware/` - Auth and rate limiting
  - `auth/` - JWT token handling and password utilities
- **Market Engine**: `/pkg/market/` - Stock price simulation logic
- **Database**: Railway MySQL with environment variable configuration
- **CORS**: Proper CORS headers configured in backend

### 2. Frontend (React SPA)
- **Pages**: `/frontend/src/pages/` - Main views (Dashboard, StockList, AdminPanel, etc.)
- **Components**: `/frontend/src/components/` - Reusable UI components
- **Services**: `/frontend/src/services/` - API client layers connecting directly to backend
- **Real-time**: Direct WebSocket connection to backend for live updates
- **Auth**: JWT-based with ProtectedRoute component

## Key Architectural Patterns

### Authentication Flow
1. User login via `/api/login` returns JWT token
2. Token stored in localStorage and included in Authorization header
3. Backend validates token in auth middleware
4. Admin users have additional permissions checked in handlers

### Real-time Updates
1. WebSocket connection established on login at `/ws`
2. Hub pattern manages client connections in `internal/websocket/`
3. Stock price updates broadcast to all connected clients
4. Chat messages routed through same WebSocket connection

### Database Resilience
- Repository layer includes automatic retry logic
- Connection pooling with configurable limits
- Graceful degradation for read-only operations
- Health check endpoints for monitoring

### API Structure
- RESTful endpoints under `/api/`
- Emergency admin endpoints under `/emergency/admin/`
- Static file serving for frontend build
- Rate limiting: 100 requests per minute per IP

## Development Guidelines

### Code Standards
- **Go**: Use `gofmt`, follow standard Go project layout
- **React**: 2-space indentation, functional components with hooks
- **Commit Messages**: Clear, descriptive messages required

### Testing Requirements
- Run tests before committing: `./run_tests.sh`
- Frontend: Jest with React Testing Library
- Backend: Go standard testing package

### Deployment
- Hosted on Railway.app with automatic deployments using Nixpacks
- Backend: Go application with direct MySQL connection
- Frontend: React SPA served separately
- Database: Railway MySQL with environment variable configuration
- Two services total (backend + frontend)

### Important Files
- `schema.sql` - Database schema with seed data
- `railway.json` - Deployment configuration
- `setupProxy.js` - Frontend development proxy config
- `CONTRIBUTING.md` - Detailed contribution guidelines
- `docs/PROJECT.md` - Comprehensive project documentation