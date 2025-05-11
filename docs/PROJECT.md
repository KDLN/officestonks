# OfficeStonks Project Documentation

This document provides comprehensive information about the OfficeStonks project.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Project Structure](#project-structure)
3. [MVP Plan](#mvp-plan)
4. [MVP Accomplishments](#mvp-accomplishments)
5. [Next Steps](#next-steps)
6. [Testing](#testing)
7. [Contributing](#contributing)

## Project Overview

Project Overview
Game Title: Office Stonks

Genre: Multiplayer stock market simulation

Backend: Golang

Frontend: Open to any framework; suggestions provided

Database: MySQL

Hosting: Railway (initially)

Target Platforms: Desktop-focused with mobile support

Initial User Base: 10-20 concurrent users

Game Mechanics and Features
Player Actions
Buy and Sell Stocks: Players can trade stocks of various companies.

Investment Groups: Players can form groups to invest collectively.

Insider Trading Events:

Risk and Reward: Players can engage in insider trading events with a chance of significant gain.

Penalties: If caught, they risk losing a large amount of money.

Stock Market Simulation
Market Influence: Stock prices change based on:

Player Activity: The volume of buying and selling affects prices.

Random Events: Fictional events impact specific sectors.

Random Events:

Sector Impact: Events that affect entire industries (e.g., tech boom, oil spill).

Frequency: Occur at random intervals to keep the game dynamic.

Winning Conditions
Portfolio Value: Players progress and compete based on the total value of their portfolios.

Leaderboards:

Daily, Weekly, Monthly Rankings: Display top players globally based on their portfolio performance.

Social Features
Mock AOL Chat Room: A chat system reminiscent of classic AOL chat rooms for player interaction.

News Ticker: In-game alerts and notifications displayed in a ticker format.

Technical Stack
Backend
Language: Golang

Framework: Standard library or Gorilla Mux for routing

Database: MySQL

Real-Time Communication: WebSockets for live updates

Hosting: Railway (scalable as user base grows)

Frontend
Option 1: React.js

Advantages: Component-based architecture, reusable UI components

Libraries: Use Socket.IO for WebSocket communication

Option 2: Vanilla JavaScript with HTML/CSS

Advantages: Lightweight, simpler for a classic look

Implementation: Use native WebSocket API for live updates

Design Theme: Windows 95/AOL classic web feel

Styling: Use CSS to mimic classic UI elements (buttons, icons, fonts)

Architecture Overview
Client-Server Model: The frontend sends requests to the backend API server.

WebSockets: Bi-directional communication for real-time stock updates and chat messages.

Database Layer:

User Data: Authentication credentials, portfolio details

Stock Data: Current prices, historical data, event impacts

Authentication:

Method: Username and password

Security: Basic encryption, input validation, hashing passwords

Concurrency Handling:

Goroutines: Utilize Golang's lightweight threads for handling multiple connections

Mutexes/Channels: Ensure data consistency when multiple users interact simultaneously

Development Roadmap
1. Planning and Design
Define Data Models:

Users: ID, username, hashed password, portfolio

Stocks: ID, company name, current price, sector

Events: Event ID, description, affected sector, impact value

Design API Endpoints:

Authentication: Login, registration

Stock Operations: Buy, sell, get stock prices

Leaderboard: Fetch rankings

Chat System: Send and receive messages

2. Backend Development
Set Up Project Structure:

Organize folders for handlers, models, services, and routes.

Implement Authentication:

Use secure password hashing (e.g., bcrypt).

Create middleware for session management.

Develop Stock Market Logic:

Price Calculation:

Base on supply and demand from player transactions.

Apply random event modifiers.

Random Events Generator:

Schedule events at random intervals.

Update affected stock prices accordingly.

WebSocket Integration:

Implement real-time updates for stock prices and chat.

Handle connections and broadcast messages efficiently.

Database Integration:

Set up MySQL database.

Write functions for CRUD operations on user and stock data.

3. Frontend Development
Design the UI:

Create a mockup reflecting the Windows 95/AOL theme.

Focus on simplicity and usability.

Implement Core Features:

Stock Dashboard: Display current stock prices and player portfolio.

Trading Interface: Forms to buy and sell stocks.

Leaderboards: Display rankings with real-time updates.

Chat Room: Implement the mock AOL chat room.

WebSocket Communication:

Establish connections to receive live updates.

Handle incoming data to update the UI dynamically.

4. Testing
Unit Testing:

Write tests for backend functions (e.g., stock price calculations, event impacts).

Integration Testing:

Test API endpoints with tools like Postman.

Ensure seamless communication between frontend and backend.

User Acceptance Testing:

Have a small group of users test the game.

Gather feedback on gameplay and user experience.

5. Deployment
Set Up on Railway:

Configure the server environment.

Set up continuous deployment pipelines if possible.

Database Hosting:

Ensure the MySQL database is securely accessible by the backend.

Scaling Considerations:

Monitor resource usage.

Plan for scaling up as the user base grows.

Additional Considerations
Security
Basic Measures:

Validate all user inputs to prevent SQL injection and XSS attacks.

Use HTTPS to encrypt data in transit.

Password Security:

Store passwords securely using hashing algorithms.

Cheating Prevention:

Implement server-side checks for transactions.

Monitor for abnormal activities (e.g., rapid trading beyond normal limits).

Performance Optimization
Caching:

Use in-memory caching for frequently accessed data like stock prices.

Efficient Data Structures:

Optimize algorithms for calculating stock prices and handling events.

Load Testing:

Simulate multiple users to test server performance under load.

Future Enhancements
Mobile Optimization:

Improve responsiveness for better mobile support.

Advanced Features:

Add more complex financial instruments (e.g., options, futures).

Introduce achievements or badges for player engagement.

Social Media Integration:

Allow players to share achievements on platforms like Twitter or Facebook.

Conclusion
Building "Office Stonks" is an exciting project that combines real-time multiplayer interactions with a simulated stock market environment. By following this development plan, you'll create a solid foundation for your game, ensuring a fun and engaging experience for your players.

Next Steps:

Set Up Your Development Environment:

Install Golang, MySQL, and your chosen frontend framework.

Start Coding:

Begin with the backend APIs and database models.

Iterate and Test:

Regularly test each component as you develop.

Deploy and Gather Feedback:

Get the game running on Railway and invite initial users.

Use their feedback to improve the game.

## Project Structure



```
/backend
  /cmd
    /api
      main.go          # Application entry point
  /internal
    /auth              # Authentication logic
    /handlers          # HTTP handlers
    /middleware        # HTTP middleware
    /models            # Data models
    /repository        # Database interactions
    /services          # Business logic
    /websocket         # WebSocket handling
  /pkg
    /market            # Stock market simulation
    /utils             # Utility functions
  go.mod
  go.sum
```


```
/frontend
  /public
  /src
    /assets            # Images, icons, etc.
    /components        # UI components
    /contexts          # React contexts (if using React)
    /hooks             # Custom hooks (if using React)
    /pages             # Page components
    /services          # API service calls
    /utils             # Utility functions
    /websocket         # WebSocket client
```


```sql
-- Users Table
CREATE TABLE users (
  id INT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  cash_balance DECIMAL(15,2) DEFAULT 10000.00,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Stocks Table
CREATE TABLE stocks (
  id INT PRIMARY KEY AUTO_INCREMENT,
  symbol VARCHAR(10) UNIQUE NOT NULL,
  name VARCHAR(100) NOT NULL,
  sector VARCHAR(50),
  current_price DECIMAL(10,2) NOT NULL,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User Portfolios Table
CREATE TABLE portfolios (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (stock_id) REFERENCES stocks(id),
  UNIQUE KEY unique_user_stock (user_id, stock_id)
);

-- Transactions Table
CREATE TABLE transactions (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  stock_id INT NOT NULL,
  quantity INT NOT NULL,
  price DECIMAL(10,2) NOT NULL,
  transaction_type ENUM('buy', 'sell') NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (stock_id) REFERENCES stocks(id)
);

-- Chat Messages Table
CREATE TABLE chat_messages (
  id INT PRIMARY KEY AUTO_INCREMENT,
  user_id INT NOT NULL,
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);
```


- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login a user

- `GET /api/users/me` - Get current user profile
- `GET /api/users/leaderboard` - Get top users by portfolio value

- `GET /api/stocks` - List all stocks
- `GET /api/stocks/{id}` - Get specific stock details

- `POST /api/trading/buy` - Buy stocks
- `POST /api/trading/sell` - Sell stocks
- `GET /api/trading/history` - Get user's transaction history

- `GET /api/portfolio` - Get user's portfolio

- `/ws` - WebSocket connection for real-time updates


- **Version Control**: Git/GitHub
- **API Testing**: Postman
- **Database Management**: MySQL Workbench
- **Local Development**: Docker Compose
- **CI/CD**: GitHub Actions

## MVP Plan



- Simple registration/login
- Portfolio tracking (cash balance + owned stocks)
- Basic profile information

- 10-15 fictional companies with stocks
- Dynamic pricing algorithm based on trading activity
- Regular price fluctuations (simulated market activity)

- Buy/sell interface
- Transaction history
- Real-time price updates

- Trading dashboard
- Portfolio overview
- Basic leaderboard

- Simple chat system
- News ticker for game events


- RESTful API endpoints for account management and trading
- WebSocket connections for real-time updates
- Simple algorithm for stock price fluctuations
- Database interactions with MySQL

- Clean, functional UI (no retro styling for MVP)
- Real-time data display
- Responsive design for basic mobile support

- Users: ID, username, password, cash_balance
- Stocks: ID, name, current_price, last_updated
- Portfolios: UserID, StockID, quantity
- Transactions: ID, UserID, StockID, quantity, price, timestamp, type(buy/sell)


- Set up project structure and repositories
- Implement database models and connections
- Create authentication system
- Build basic API endpoints

- Implement stock market simulation
- Create trading functionality
- Develop portfolio management
- Build basic frontend for trading

- Implement WebSockets
- Add live price updates
- Create simple chat system
- Build news ticker

- Create leaderboard
- Add final UI improvements
- Test with 10-20 users
- Deploy to Railway

- System stability with 10-20 concurrent users
- Average session length > 10 minutes
- Trading volume > 50 transactions per day
- User feedback score > 7/10

- Investment groups
- Insider trading events
- Advanced market events
- Mobile app version
- Social media integration


1. **Create project repository structure**
   - Set up Git repository
   - Create folder structure following PROJECT_STRUCTURE.md
   - Add README and documentation

2. **Set up Go environment**
   - Initialize Go modules
   - Create main.go entry point
   - Set up basic HTTP server

3. **Database setup**
   - Create MySQL schema files
   - Write database migration scripts
   - Set up connection pool in Go

4. **User authentication**
   - Implement registration endpoint
   - Create login functionality
   - Set up JWT token system

5. **Stock data models**
   - Create stock model
   - Add seed data for initial companies
   - Implement API to list stocks

6. **Portfolio system**
   - Create portfolio and transaction models
   - Implement portfolio tracking
   - Set up portfolio view endpoint


7. **Stock price simulation**
   - Implement basic price fluctuation algorithm
   - Create scheduled tasks for price updates
   - Set up history tracking

8. **Trading functionality**
   - Create buy endpoint with validation
   - Implement sell endpoint with validation
   - Add transaction history endpoint

9. **WebSocket setup**
   - Implement WebSocket server
   - Create client connection handling
   - Set up real-time price broadcasting

10. **Frontend initialization**
    - Set up frontend project (React recommended)
    - Configure build system
    - Create basic layout components

11. **Authentication UI**
    - Build login screen
    - Create registration form
    - Implement JWT storage and management

12. **Stock listing and details**
    - Create stock listing page
    - Build stock detail view
    - Implement real-time price updates


13. **Trading interface**
    - Create buy/sell forms
    - Implement portfolio overview
    - Add transaction history view

14. **WebSocket integration**
    - Connect frontend to WebSockets
    - Implement real-time updates
    - Add notification system for price changes

15. **Leaderboard**
    - Create leaderboard API endpoint
    - Build leaderboard UI component
    - Implement periodic updates

16. **Chat system**
    - Implement simple chat backend
    - Create chat UI component
    - Connect to WebSockets

17. **Docker setup**
    - Create Dockerfiles for backend and frontend
    - Set up Docker Compose for local development
    - Test containerized application

18. **Deployment preparation**
    - Configure Railway deployment
    - Set up environment variables
    - Test deployment pipeline


19. **Testing**
    - Write basic API tests
    - Test critical trading functionality
    - Load test WebSockets with dummy clients

20. **Internal user testing**
    - Deploy to staging environment
    - Invite 3-5 test users
    - Collect feedback and fix critical issues

21. **Launch MVP**
    - Deploy to production
    - Monitor system performance
    - Collect initial user feedback


- **Project structure**: Keep it modular from the start to allow easy expansion
- **Authentication**: Use proven libraries rather than building from scratch
- **Database**: Include indexes for frequently queried fields
- **Stock simulation**: Start with a simple algorithm, can be refined later
- **WebSockets**: Implement heartbeats to maintain connections
- **Frontend**: Use component libraries to speed up development
- **Testing**: Focus on critical paths (authentication, trading)
- **Deployment**: Set up CI/CD from the beginning for smooth iterations


- Backend API endpoint tests passing
- WebSocket maintaining 20+ concurrent connections
- UI functioning on desktop and mobile browsers
- Trading system correctly updating portfolios and prices
- System handling 100+ trades per hour without performance issues

## MVP Accomplishments



- Simple registration/login with JWT authentication
- Portfolio tracking (cash balance + owned stocks)
- Basic profile information
- Admin user role with special permissions

- 10 fictional companies with stocks
- Dynamic pricing algorithm based on trading activity
- Regular price fluctuations (simulated market activity)

- Buy/sell interface
- Transaction history
- Real-time price updates via WebSockets

- Trading dashboard with stock listings
- Portfolio overview with cost basis and gain/loss calculations
- Leaderboard showing top users by portfolio value
- Admin panel for system management

- Live chat system with real-time updates
- WebSocket-based notifications


- RESTful API endpoints for all core functionality
- JWT-based authentication with role permissions
- WebSocket connections for real-time updates
- Stock price simulation algorithm
- MySQL database integration
- Rate limiting for API protection
- Admin middleware for protected endpoints

- Clean, functional UI with responsive design
- Real-time data display with WebSockets
- Mobile-responsive components
- Authentication with protected routes
- Admin access control
- Price caching for consistent UI experience

- Users: ID, username, password, cash_balance, is_admin
- Stocks: ID, name, current_price, last_updated
- Portfolios: UserID, StockID, quantity
- Transactions: ID, UserID, StockID, quantity, price, timestamp, type(buy/sell)
- Chat Messages: ID, UserID, message, timestamp


- Backend deployed on Railway with MySQL database
- Frontend deployed on Railway with connection to backend
- WebSocket connections functioning across domains
- CORS configuration for secure cross-origin requests

- Dockerized application with multi-stage builds
- Automated deployment via GitHub integration
- Database schema migrations and updates
- Error logging and monitoring


- Tracking of average purchase price per stock
- Real-time calculation of gains and losses
- History of all buy/sell transactions

- User management (view, edit, delete users)
- Admin-only actions (reset stock prices, clear chat)
- Role-based access control

- Live stock price updates
- Instant chat messages
- Portfolio value updates


1. **Repository Structure Cleanup**:
   - Reorganized code to follow standard Go project layout
   - Fixed deployment issues by restructuring directories

2. **Cross-Origin Resource Sharing (CORS)**:
   - Resolved complex CORS issues between frontend and backend
   - Configured proper headers and credentials for secure communication

3. **WebSocket Integration**:
   - Implemented secure WebSocket connections
   - Added fallback mechanisms for connection failures
   - Created protocol handling for different environments (ws vs wss)

4. **Price Caching**:
   - Solved price inconsistency issues across page navigation
   - Implemented global price cache with real-time updates

5. **Admin Authentication**:
   - Added proper role-based permissions
   - Created secure admin middleware
   - Built flexible authentication that supports admin users

6. **Database Schema Evolution**:
   - Enhanced schema to support new features
   - Created robust update scripts for existing databases
   - Maintained backward compatibility


- Advanced market events and simulations
- Mobile app version
- Social media integration
- Investment groups
- Insider trading events
- Analytics dashboard
- Performance optimizations

## Next Steps



The MVP has made significant progress with the following features implemented:

✅ **User Authentication**
- Registration and login functionality
- JWT-based authentication
- Protected routes

✅ **Core Market Functionality**
- Stock listing and detail views
- Real-time price updates via WebSocket
- Dynamic stock price simulation with trends

✅ **Trading System**
- Buy/sell functionality
- Portfolio tracking
- Transaction history

✅ **Frontend Dashboard**
- Portfolio overview
- Recent transactions
- Top stocks display

✅ **Deployment**
- Successfully deployed to Railway
- Frontend/backend communication working
- Database integration


- [ ] Create backend endpoint for leaderboard data
- [ ] Implement `GetTopUsers` method in user repository
- [ ] Add leaderboard handler in the API
- [ ] Develop leaderboard frontend page

- [ ] Create chat message repository
- [ ] Implement chat services in the backend
- [ ] Add WebSocket handlers for chat messages
- [ ] Build chat UI component for the frontend
- [ ] Connect chat to WebSockets for real-time updates

- [ ] Finish Portfolio page with detailed view
- [ ] Implement Transactions page with filters and sorting
- [ ] Enhance StockDetail page with charts and historical data

- [ ] Add loading indicators for data fetching
- [ ] Implement proper error handling and user feedback
- [ ] Add confirmation dialogs for trades
- [ ] Improve mobile responsiveness

- [ ] Implement frontend unit tests
- [ ] Create integration tests for critical flows
- [ ] Perform load testing for WebSockets
- [ ] Manual testing with 10-20 users

- [ ] Implement a news ticker for market events
- [ ] Add a notification system for price alerts
- [ ] Create a user settings page
- [ ] Enable avatar or profile customization


1. **Leaderboard and Placeholder Pages** - 1 week
2. **Chat System** - 1 week
3. **UX Improvements** - 1 week
4. **Testing and Validation** - 1 week

Total time to MVP completion: Approximately 4 weeks


```go
// Backend - Add to user repository
func (r *UserRepo) GetTopUsers(limit int) ([]*models.User, error) {
    // Calculate total portfolio value for each user and sort
    // Return top users by portfolio value
}

// Add endpoint to main.go
apiRouter.HandleFunc("/users/leaderboard", userHandler.GetLeaderboard).Methods("GET", "OPTIONS")
```

```go
// Add chat message model
type ChatMessage struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
}

// Add to websocket hub.go
func (h *Hub) BroadcastChatMessage(message ChatMessage) {
    // Broadcast chat message to all connected clients
}
```

```jsx
// Basic chat component structure
const Chat = () => {
    const [messages, setMessages] = useState([]);
    const [newMessage, setNewMessage] = useState('');
    
    // Handle incoming WebSocket messages
    // Send new messages to the server
    // Render chat interface
};
```


1. **User Engagement**
   - Average session time > 10 minutes
   - At least 5 trades per user per session
   - Daily active users > 50% of registered users

2. **Technical Performance**
   - WebSocket connections supporting 50+ concurrent users
   - Page load times < 2 seconds
   - API response times < 500ms

3. **User Satisfaction**
   - Positive feedback from beta testers
   - Low bounce rate (< 30%)
   - Feature request to bug report ratio > 2:1


1. **Enhanced Market Simulation**
   - Add market events (crashes, booms, etc.)
   - Implement sector-based trends
   - Create more sophisticated price algorithms

2. **Social Features**
   - Investment groups and teams
   - Friend system
   - Achievements and badges

3. **Advanced Trading**
   - Short selling
   - Options trading
   - Limit orders

4. **Mobile Experience**
   - PWA support
   - Responsive design optimization
   - Native app consideration

## Testing


This document outlines the testing strategy for the Office Stonks application, including how to run tests and write new ones.


We use Docker to create a consistent testing environment. This approach makes it easy to run tests on any machine without worrying about installing dependencies or configuring databases.


To run tests with Docker:
- Docker
- Docker Compose

To run tests locally without Docker:
- Go 1.20 or higher
- Node.js 18 or higher
- MySQL 8.0 or higher



Run all tests in containerized environments:

```bash
./run_tests.sh
```

This script will:
1. Set up a test database in a Docker container
2. Run backend tests in a Docker container
3. Run frontend tests in a Docker container
4. Clean up all containers when finished


For local development, you can run tests directly on your machine:

```bash
./run_local_tests.sh
```

Note: This requires Go, Node.js, and MySQL to be installed and properly configured on your machine.



Backend tests are located in `backend/internal/tests/` and are organized by feature:

- `auth_test.go`: Tests for authentication endpoints
- `market_test.go`: Tests for market functionality
- `integration_test.go`: Database integration tests

The tests use a dedicated test database (`officestonks_test`) with isolated test data.


Frontend tests are located alongside their components with the `.test.js` extension:

- `src/components/Navigation.test.js`: Tests for the Navigation component
- `src/pages/Login.test.js`: Tests for the Login page
- `src/services/stock.test.js`: Tests for the stock service

We use Jest and React Testing Library for frontend testing.



1. Add new test files to `backend/internal/tests/`
2. Use the provided test helper functions in `test_helpers.go`
3. Follow the pattern of existing tests

Example:

```go
func TestMyNewFeature(t *testing.T) {
    // Skip if no test database connection
    if TestDB == nil {
        t.Skip("No test database connection")
    }

    // Setup test router
    router := SetupTestRouter(TestDB)

    // Make request
    rr := MakeRequest("GET", "/api/my-endpoint", nil, router)

    // Check status code
    if rr.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
    }

    // Further assertions...
}
```


1. Create a `.test.js` file alongside the component or service you're testing
2. Use React Testing Library to test components
3. Use Jest to mock dependencies

Example:

```javascript
import { render, screen } from '@testing-library/react';
import MyComponent from './MyComponent';

test('renders my component correctly', () => {
  render(<MyComponent />);
  expect(screen.getByText('Expected Text')).toBeInTheDocument();
});
```


In a real CI/CD setup, you would:
1. Run `./run_tests.sh` on every pull request
2. Block merging if tests fail
3. Generate test coverage reports


1. Write tests for critical functionality first
2. Aim for high coverage in core services and data models
3. Test edge cases and error handling
4. Keep tests independent (no test should depend on another test)
5. Use setup and teardown functions to maintain a clean test environment
6. Mock external services and dependencies
7. Organize tests logically by feature or component

## Contributing


Thank you for considering contributing to Office Stonks! This document outlines the process for contributing to this project.


By participating in this project, you agree to abide by our Code of Conduct which expects respectful and inclusive behavior from all contributors.


1. **Fork the Repository**: Start by forking the repository to your GitHub account.

2. **Clone the Repository**: Clone your fork of the repository to your local machine.

3. **Create a Branch**: Create a new branch for your feature or bug fix.
   ```
   git checkout -b feature/your-feature-name
   ```

4. **Make Changes**: Make your changes to the codebase.

5. **Run Tests**: Ensure your changes pass the tests.
   ```
   ./run_tests.sh
   ```

6. **Commit Changes**: Commit your changes with a clear, descriptive commit message.
   ```
   git commit -m "Add feature: description of your changes"
   ```

7. **Push Changes**: Push your branch to your fork on GitHub.
   ```
   git push origin feature/your-feature-name
   ```

8. **Submit a Pull Request**: Open a pull request from your branch to our main repository.


1. Ensure your code passes all tests
2. Update documentation if needed
3. Describe what your changes do and why they should be included
4. Wait for review from maintainers


See the [README.md](README.md) for detailed instructions on setting up the development environment.


- Follow standard Go formatting rules
- Run `gofmt` before committing
- Follow Go best practices

- Follow standard ESLint rules
- Use proper indentation (2 spaces)
- Add JSDoc comments for functions


By contributing to Office Stonks, you agree that your contributions will be licensed under the project's MIT License.
