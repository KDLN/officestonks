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

### CORS Proxy (Deprecated)
The CORS proxy is no longer needed. The backend now handles CORS headers directly.

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
- **Pages**: `/frontend/src/pages/` - Main views:
  - Dashboard - Portfolio overview and top stocks
  - StockList - Browse all available stocks
  - StockDetail - Individual stock trading page
  - Portfolio - Detailed holdings and transaction history
  - Leaderboard - Top players by portfolio value
  - Transactions - Full transaction history
  - AdminPanel - Admin controls for game management
- **Components**: `/frontend/src/components/` - Reusable UI components
- **Services**: `/frontend/src/services/` - API client layers connecting directly to backend
- **Real-time**: Direct WebSocket connection to backend for live updates
- **Auth**: JWT-based with ProtectedRoute component
- **Build**: Production build served by backend from `/frontend/build/`

## Key Architectural Patterns

### Authentication Flow
1. User login via `/api/auth/login` returns JWT token
2. Token stored in localStorage and included in Authorization header
3. Backend validates token in auth middleware
4. Admin users have additional permissions checked in handlers

### Real-time Updates
1. WebSocket connection established on login at `/ws?token=<jwt>`
2. Hub pattern manages client connections in `internal/websocket/`
3. Stock price updates broadcast every 2 seconds to all connected clients
4. Chat messages routed through same WebSocket connection

### Market Simulation
- Market simulator with atomic price reset capability
- 2-second update interval with 5% volatility
- Pause/Resume functionality for admin operations
- Transaction impact on stock prices
- Price floor at $0.01 to prevent negative values

### Database Resilience
- Repository layer includes automatic retry logic
- Connection pooling with configurable limits
- Graceful degradation for read-only operations
- Health check endpoints for monitoring

### API Structure
- RESTful endpoints under `/api/`
- Emergency admin endpoints with bypass authentication
- Static file serving for frontend build from `/frontend/build/`
- Rate limiting: 100 requests per minute per IP
- Health check endpoints: `/health` and `/health-check`

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
- Single consolidated service architecture:
  - Backend: Go application serves both API and frontend
  - Frontend: Production build embedded and served from backend
  - Database: Railway MySQL with environment variable configuration
- GitHub Actions for automated deployment on push to main branch
- Frontend build must be committed to repository for deployment

### Important Files
- `schema.sql` - Database schema with seed data
- `railway.json` - Deployment configuration
- `setupProxy.js` - Frontend development proxy config
- `CONTRIBUTING.md` - Detailed contribution guidelines
- `docs/PROJECT.md` - Comprehensive project documentation

## MVP Launch Plan for User Testing

### Current Feature Status ✅
1. **Core Trading**
   - User registration and authentication
   - Buy/sell stocks with real-time pricing
   - Portfolio management and tracking
   - Transaction history
   
2. **Market Simulation**
   - Automated price updates every 2 seconds
   - Transaction-based price impact
   - Market volatility and trends
   
3. **Social Features**
   - Real-time chat system
   - Leaderboard rankings
   
4. **Admin Controls**
   - Stock price reset
   - User management
   - Chat moderation

### Priority Features for MVP Launch 🚀

#### Phase 1: Core Stability (Before Launch)
1. **User Experience Polish**
   - Add loading states for all async operations
   - Improve error messages and user feedback
   - Add confirmation dialogs for trades
   - Mobile responsive design verification

2. **Game Balance**
   - Review starting cash amount ($10,000)
   - Adjust market volatility parameters
   - Set appropriate price impact for trades
   - Configure leaderboard update frequency

3. **Security & Rate Limiting**
   - Verify JWT token expiration (currently 24h)
   - Test rate limiting (100 req/min)
   - Add trade frequency limits per user
   - Validate all user inputs

#### Phase 2: Launch Features (Week 1-2)
1. **Onboarding**
   - Welcome tutorial or guide
   - Tooltips for first-time users
   - Sample trades or paper trading mode

2. **Performance Monitoring**
   - User activity tracking
   - Performance metrics dashboard
   - Error logging and monitoring
   - Database query optimization

3. **User Engagement**
   - Email notifications (optional)
   - Daily/weekly challenges
   - Achievement system
   - Portfolio performance graphs

#### Phase 3: Post-Launch Enhancements (Based on Feedback)
1. **Advanced Trading**
   - Limit orders
   - Stop-loss orders
   - Market orders vs limit orders
   - Short selling

2. **Social Enhancements**
   - Private messaging
   - Follow other traders
   - Share trades/strategies
   - Trading groups/clubs

3. **Game Modes**
   - Time-limited competitions
   - Themed challenges
   - Different starting conditions
   - Seasonal events

### Testing Checklist Before Launch
- [ ] Load testing with multiple concurrent users
- [ ] WebSocket connection stability over time
- [ ] Database backup and recovery procedures
- [ ] Admin panel functionality verification
- [ ] Cross-browser compatibility testing
- [ ] Mobile device testing
- [ ] Security penetration testing
- [ ] Performance benchmarking

### Success Metrics
- User retention (1-day, 7-day, 30-day)
- Average session duration
- Trades per user per day
- Chat engagement rate
- Bug report frequency
- User feedback sentiment