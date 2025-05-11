# OfficeStonks Repository - Final Cleanup Plan

This document provides a complete plan for cleaning up the OfficeStonks repository, addressing outstanding issues, and ensuring a clean, maintainable codebase.

## 1. File Cleanup

The repository contains many temporary files, mock implementations, and debugging tools that should be removed:

### Files to Remove

```bash
# Remove temporary server implementations
rm -f /home/kdln/code/officestonks/server.fix.go
rm -f /home/kdln/code/officestonks/static-server.go
rm -f /home/kdln/code/officestonks/static-server
rm -f /home/kdln/code/officestonks/static-mode.go
rm -f /home/kdln/code/officestonks/build-static-server.sh
rm -f /home/kdln/code/officestonks/deploy-static-server.sh

# Remove mock API files
rm -rf /home/kdln/code/officestonks/mock-api
rm -f /home/kdln/code/officestonks/package-mock.json
rm -f /home/kdln/code/officestonks/cors-mock-server.js
rm -f /home/kdln/code/officestonks/deploy-mock-api.sh
rm -f /home/kdln/code/officestonks/deploy-ultra-simple.sh

# Remove redundant Dockerfiles
rm -f /home/kdln/code/officestonks/Dockerfile.static
rm -f /home/kdln/code/officestonks/Dockerfile.scratch
rm -f /home/kdln/code/officestonks/Dockerfile.simple
rm -f /home/kdln/code/officestonks/railway.static.json

# Remove temporary CORS fix files
rm -f /home/kdln/code/officestonks/cors-proxy.js
rm -f /home/kdln/code/officestonks/cors-proxy-package.json
rm -f /home/kdln/code/officestonks/cors-proxy-setup.md
rm -f /home/kdln/code/officestonks/allow-all-proxy.js
rm -f /home/kdln/code/officestonks/allow-all-proxy-commonjs.js
rm -f /home/kdln/code/officestonks/allow-all-package.json
rm -f /home/kdln/code/officestonks/allow-all-package-commonjs.json
rm -f /home/kdln/code/officestonks/no-cors-proxy.js
rm -f /home/kdln/code/officestonks/no-cors-package.json
rm -f /home/kdln/code/officestonks/simple-cors-proxy.js
rm -f /home/kdln/code/officestonks/simple-package.json
rm -f /home/kdln/code/officestonks/ultra-simple-proxy.js
rm -f /home/kdln/code/officestonks/ultra-simple-package.json
rm -f /home/kdln/code/officestonks/cors-debug.html
rm -f /home/kdln/code/officestonks/cors-test.html
rm -f /home/kdln/code/officestonks/cors-verify.html
rm -f /home/kdln/code/officestonks/test-cors.sh
rm -f /home/kdln/code/officestonks/verify-cors.sh
rm -f /home/kdln/code/officestonks/test-simple-proxy.sh

# Remove patches and frontend fixes
rm -f /home/kdln/code/officestonks/admin.js.patch
rm -f /home/kdln/code/officestonks/frontend-fix.js
rm -f /home/kdln/code/officestonks/quick-frontend-fix.js
rm -f /home/kdln/code/officestonks/final-frontend-fix.js
rm -f /home/kdln/code/officestonks/stock-reset-fix.js
rm -f /home/kdln/code/officestonks/test-mysql-connection.js

# Remove redundant documentation
rm -f /home/kdln/code/officestonks/ADMIN_FIX_SUMMARY.md
rm -f /home/kdln/code/officestonks/CORS-FIX-FOR-FRONTEND.md
rm -f /home/kdln/code/officestonks/CORS-FIX-README.md
rm -f /home/kdln/code/officestonks/CORS-PROXY-README.md
rm -f /home/kdln/code/officestonks/cors-allowed-backend.md

# Consider removing the backup directory (after verifying all necessary code is migrated)
# rm -rf /home/kdln/code/officestonks/backup
```

## 2. Fix Core Issues

### Database Connection (db.go)

The current database connection code in `/internal/repository/db.go` has hardcoded values that are causing connection issues. Replace it with the clean version in `db.go.clean`:

```bash
cp /home/kdln/code/officestonks/internal/repository/db.go.clean /home/kdln/code/officestonks/internal/repository/db.go
```

Key improvements in the clean version:
- Uses environment variables properly with Railway's standard variables
- Has simpler, more maintainable fallback logic
- Includes appropriate logging without excessive output
- Maintains the same connection parameters that previously worked

### CORS Handling (cors.go)

The current CORS implementation should be replaced with the cleaner version in `cors.go.clean`:

```bash
cp /home/kdln/code/officestonks/cmd/api/cors.go.clean /home/kdln/code/officestonks/cmd/api/cors.go
```

Key improvements in the clean version:
- Respects the Origin header for proper same-origin requests
- Falls back to wildcard when needed
- Handles preflight requests properly
- Includes credential support
- Simpler, more maintainable implementation

## 3. Update Documentation

The README.md should be updated to reflect the current project structure. The backend code is now at the root level, not in a `/backend` directory as mentioned in the README.

Create an updated version of README.md:

```md
# Office Stonks - Multiplayer Stock Market Game

A real-time multiplayer stock market simulation game where players can trade stocks, form investment groups, and compete for the highest portfolio value.

## Overview

Office Stonks is an online multiplayer stock market simulation that allows players to:
- Buy and sell virtual stocks based on real market dynamics
- See real-time price updates via WebSockets
- Compete on leaderboards with other players
- Chat with other players
- Manage their portfolios

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

This application is deployed on Railway. See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
```

## 4. Testing Plan

After completing the cleanup, testing is crucial to ensure everything works correctly:

1. **Local Testing**:
   ```bash
   # Build the backend
   go build -o server ./cmd/api/main.go
   
   # Run with required environment variables
   DB_HOST=localhost DB_PORT=3306 DB_USER=root DB_PASSWORD=yourpassword DB_NAME=railway ./server
   
   # In another terminal, start the frontend
   cd frontend
   npm start
   ```

2. **Railway Deployment Testing**:
   - Push the changes to the repository
   - Deploy to Railway
   - Verify all endpoints work correctly
   - Test with the frontend to ensure connectivity

3. **Database Connection Testing**:
   ```bash
   # Set the correct environment variables for Railway
   export DB_HOST=mysql.railway.internal
   export DB_PORT=3306
   export DB_USER=root
   export DB_PASSWORD=yourpassword
   export DB_NAME=railway
   
   # Run a simple connection test
   go run cmd/api/main.go
   ```

## 5. Benefits of This Cleanup

1. **Simplified Codebase**: Removing unnecessary files makes the codebase easier to understand and maintain.
2. **Improved Reliability**: The clean db.go and cors.go implementations have fewer edge cases and better handling.
3. **Reduced Confusion**: Eliminating temporary solutions prevents confusion about which files are actually used.
4. **Better Documentation**: Updated documentation ensures everyone understands the current state of the project.

## 6. Additional Recommendations

1. **Environment Variables**: Create a `.env.example` file in the root directory showing all required environment variables.
2. **Branching Strategy**: Implement a clear Git workflow (e.g., GitFlow) for future development.
3. **CI/CD Pipeline**: Set up automated tests and deployments to prevent regression.
4. **Error Handling**: Review error handling throughout the codebase to ensure consistency.
5. **Logging**: Implement structured logging with appropriate log levels throughout the application.

## 7. Conclusion

This cleanup plan addresses the immediate issues with the repository structure, database connectivity, and CORS handling. By implementing these changes, the application will be more maintainable, reliable, and easier to deploy.