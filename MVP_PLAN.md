# Office Stonks - MVP Plan

## Core Features

### 1. User System
- Simple registration/login
- Portfolio tracking (cash balance + owned stocks)
- Basic profile information

### 2. Stock Market
- 10-15 fictional companies with stocks
- Dynamic pricing algorithm based on trading activity
- Regular price fluctuations (simulated market activity)

### 3. Trading Functionality
- Buy/sell interface
- Transaction history
- Real-time price updates

### 4. Simplified UI
- Trading dashboard
- Portfolio overview
- Basic leaderboard

### 5. Minimal Social Features
- Simple chat system
- News ticker for game events

## Technical Implementation

### Backend (Go)
- RESTful API endpoints for account management and trading
- WebSocket connections for real-time updates
- Simple algorithm for stock price fluctuations
- Database interactions with MySQL

### Frontend
- Clean, functional UI (no retro styling for MVP)
- Real-time data display
- Responsive design for basic mobile support

### Database Schema
- Users: ID, username, password, cash_balance
- Stocks: ID, name, current_price, last_updated
- Portfolios: UserID, StockID, quantity
- Transactions: ID, UserID, StockID, quantity, price, timestamp, type(buy/sell)

## Development Phases

### Phase 1: Foundation (1-2 weeks)
- Set up project structure and repositories
- Implement database models and connections
- Create authentication system
- Build basic API endpoints

### Phase 2: Core Gameplay (2-3 weeks)
- Implement stock market simulation
- Create trading functionality
- Develop portfolio management
- Build basic frontend for trading

### Phase 3: Real-time Features (1-2 weeks)
- Implement WebSockets
- Add live price updates
- Create simple chat system
- Build news ticker

### Phase 4: Polish & Deploy (1 week)
- Create leaderboard
- Add final UI improvements
- Test with 10-20 users
- Deploy to Railway

## Metrics for Success
- System stability with 10-20 concurrent users
- Average session length > 10 minutes
- Trading volume > 50 transactions per day
- User feedback score > 7/10

## Future Enhancements (Post-MVP)
- Investment groups
- Insider trading events
- Advanced market events
- Mobile app version
- Social media integration
