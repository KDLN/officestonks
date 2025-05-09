# Office Stonks - Project Structure

## Backend (Go)

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

## Frontend

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

## Database

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

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login a user

### Users
- `GET /api/users/me` - Get current user profile
- `GET /api/users/leaderboard` - Get top users by portfolio value

### Stocks
- `GET /api/stocks` - List all stocks
- `GET /api/stocks/{id}` - Get specific stock details

### Trading
- `POST /api/trading/buy` - Buy stocks
- `POST /api/trading/sell` - Sell stocks
- `GET /api/trading/history` - Get user's transaction history

### Portfolio
- `GET /api/portfolio` - Get user's portfolio

### WebSocket
- `/ws` - WebSocket connection for real-time updates

## Development Tools

- **Version Control**: Git/GitHub
- **API Testing**: Postman
- **Database Management**: MySQL Workbench
- **Local Development**: Docker Compose
- **CI/CD**: GitHub Actions