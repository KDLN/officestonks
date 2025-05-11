# Repository Cleanup

This document details the cleanup efforts for the OfficeStonks repository.

## Table of Contents

1. [Cleanup Summary](#cleanup-summary)
2. [Cleanup Process](#cleanup-process)
3. [Final State](#final-state)

## Cleanup Summary



After a comprehensive audit of the OfficeStonks repository, we've identified several areas that need attention:

1. **Repository Bloat**: The repository contains numerous temporary files, mock implementations, and debugging tools that were created during previous troubleshooting efforts. These files are no longer needed and create confusion.

2. **Database Connection Issues**: The `db.go` file contains hardcoded database connection details instead of properly using environment variables, which is likely causing the connection issues in Railway.

3. **CORS Implementation**: The current CORS implementation has been modified multiple times with different approaches, leading to an overly complex solution that doesn't work consistently.

4. **Redundant Dockerfiles**: There are multiple Dockerfile variations with different configurations, making it unclear which one should be used.

5. **Excessive Documentation**: Several markdown files contain overlapping information about CORS fixes and admin panel issues.


We've prepared three key files to address these issues:

1. **`REPOSITORY_CLEANUP_FINAL.md`**: A comprehensive cleanup plan detailing all the issues and recommended solutions.

2. **`cleanup.sh`**: An executable script that automates the removal of unnecessary files and replaces key files with their clean versions.

3. **`README.md.new`**: An updated README file that reflects the current project structure and provides clear information for developers.


The cleanup process makes the following improvements:

1. **Simplified Codebase**: Removing ~50 unnecessary files makes the repository significantly cleaner and easier to navigate.

2. **Fixed Database Connection**: Reverted to a known-working database connection implementation that properly uses environment variables.

3. **Improved CORS Handling**: Implemented a simpler, more reliable CORS middleware that works with all required origins.

4. **Clear Documentation**: Updated documentation that properly reflects the current state of the project.


Beyond the immediate cleanup, we recommend:

1. **Implement Proper Git Workflow**: Adopt a branching strategy like GitFlow to prevent accumulation of temporary fixes in the main branch.

2. **Environment Variable Management**: Create a `.env.example` file to document all required environment variables.

3. **Automated Testing**: Implement unit and integration tests to catch issues early.

4. **CI/CD Pipeline**: Set up automated build and deployment to ensure consistency between environments.

5. **Structured Logging**: Replace the current ad-hoc logging with structured logging using appropriate log levels.


1. Run the `cleanup.sh` script to perform the cleanup
2. Replace the current README.md with README.md.new
3. Test the application locally to ensure everything works
4. Commit the changes and deploy to Railway
5. Verify functionality in the production environment

---

This cleanup will significantly improve code quality, maintainability, and deployment reliability for the OfficeStonks project.

## Cleanup Process


This document describes the cleanup performed on the repository structure to fix duplication issues.


The repository previously had a nested structure with code duplicated across multiple directories:
- `/officestonks/backend/` (older version)
- `/officestonks/backend/backend/` (newer version with chat and leaderboard functionality)

This structure was causing issues with Docker builds and Railway deployments.


1. Consolidated the Go codebase by:
   - Moving all code from `/backend/backend/` to the root level
   - Using the most up-to-date version of each file
   - Removing the nested directory structure

2. Updated the Dockerfile to:
   - Use the simplified directory structure
   - Build the binary in the correct location
   - Create a streamlined container image

3. Added missing features:
   - Chat system functionality
   - Leaderboard system
   - User profiles


- Simplified directory structure
- Cleaner build process
- More maintainable codebase
- Improved deployment reliability


Test the deployment to ensure all functionality works correctly with the new structure.

This document outlines a comprehensive audit of the OfficeStonks repository, identifying issues, unnecessary files, and recommendations for cleanup.


1. **Mock API and Static Server Files**:
   - `/mock-api/` directory
   - `static-server.go`
   - `cors-mock-server.js`
   - `server.fix.go`
   - `deploy-mock-api.sh`
   - `deploy-static-server.sh`
   - `build-static-server.sh`
   - `railway.static.json`
   - `Dockerfile.static`
   - `package-mock.json`

2. **Backup Directories**:
   - `/backup/` directory

3. **Temporary and Deprecated Files**:
   - Any `.bak` files
   - `ADMIN_FIX_SUMMARY.md`
   - `cors-proxy.js`
   - `admin.js.patch`
   - Unused Dockerfiles (`Dockerfile.railway`, `Dockerfile.scratch`, `Dockerfile.simple`)



1. **Environment Variable Handling**:
   - Inconsistent handling of environment variables
   - Hardcoded database connection details in recent commits

2. **Default Fallbacks**:
   - Code is falling back to `mysql.railway.internal` despite environment variables
   - Recent changes have overridden the Railway-provided environment


1. **Database Connection**:
   - `/internal/repository/db.go`: Needs consistent environment variable handling
   - Recent changes should be reverted to the version that previously worked

2. **Configuration**:
   - Ensure all environment variables are properly passed through in Railway
   - Double-check that Railway is providing the correct database URL


1. **CORS Handling**:
   - `/cmd/api/cors.go`: Overly complicated CORS handling
   - Multiple versions with inconsistent implementations
   - Needs simpler, more reliable implementation

2. **Error Handling**:
   - Inconsistent error handling throughout the codebase
   - Many errors are logged but not propagated properly

3. **Logging**:
   - Excessive logging makes it hard to find important information
   - Need more structured logging with appropriate log levels


1. **Cleanup Repository**:
   ```bash
   # Remove mock and temporary files
   git rm -rf mock-api/
   git rm static-server.go server.fix.go
   git rm cors-mock-server.js package-mock.json
   git rm deploy-mock-api.sh deploy-static-server.sh build-static-server.sh
   git rm railway.static.json Dockerfile.static
   git rm ADMIN_FIX_SUMMARY.md

   # Remove backup directory
   git rm -rf backup/

   # Remove deprecated files
   git rm Dockerfile.railway Dockerfile.scratch Dockerfile.simple
   git rm cors-proxy.js admin.js.patch
   ```

2. **Revert to Working Database Connection**:
   - Find the last known working version of `db.go` and restore it
   - Remove all hardcoded values added in recent commits
   - Ensure proper environment variable handling
   - Simplified error handling and logging

3. **Simplify CORS Handling**:
   - Implement a clean CORS policy that works with all origins
   - Use a single approach for all endpoints
   - Make sure OPTIONS preflight requests are handled properly

4. **Fix Docker Configuration**:
   - Ensure Docker entrypoint works correctly
   - Verify environment variables are passed through


Based on repository history, here's what has worked previously:

```go
// Get database details from environment variables
username := os.Getenv("MYSQLUSER")
if username == "" {
    username = os.Getenv("DB_USER")
    if username == "" {
        username = "root"
    }
}

password := os.Getenv("MYSQLPASSWORD")
if password == "" {
    password = os.Getenv("DB_PASSWORD")
    if password == "" {
        password = "your-default-password"
    }
}

host := os.Getenv("MYSQLHOST")
if host == "" {
    host = os.Getenv("DB_HOST")
    if host == "" {
        host = "localhost"
    }
}

port := os.Getenv("MYSQLPORT")
if port == "" {
    port = os.Getenv("DB_PORT")
    if port == "" {
        port = "3306"
    }
}

dbname := os.Getenv("MYSQLDATABASE")
if dbname == "" {
    dbname = os.Getenv("DB_NAME")
    if dbname == "" {
        dbname = "railway"
    }
}

// Simple DSN without excessive parameters
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
    username, password, host, port, dbname)
```


After cleanup, perform the following tests:

1. **Local Testing**:
   - Test database connection locally
   - Verify all API endpoints work
   - Confirm frontend can connect to backend

2. **Railway Testing**:
   - Deploy with clean configuration
   - Verify database connection works
   - Test all endpoints through the frontend


1. Complete the repository cleanup
2. Revert database connection to last known working version
3. Implement a simpler CORS policy
4. Test thoroughly before final deployment

## Final State


This document provides a complete plan for cleaning up the OfficeStonks repository, addressing outstanding issues, and ensuring a clean, maintainable codebase.


The repository contains many temporary files, mock implementations, and debugging tools that should be removed:


```bash
rm -f /home/kdln/code/officestonks/server.fix.go
rm -f /home/kdln/code/officestonks/static-server.go
rm -f /home/kdln/code/officestonks/static-server
rm -f /home/kdln/code/officestonks/static-mode.go
rm -f /home/kdln/code/officestonks/build-static-server.sh
rm -f /home/kdln/code/officestonks/deploy-static-server.sh

rm -rf /home/kdln/code/officestonks/mock-api
rm -f /home/kdln/code/officestonks/package-mock.json
rm -f /home/kdln/code/officestonks/cors-mock-server.js
rm -f /home/kdln/code/officestonks/deploy-mock-api.sh
rm -f /home/kdln/code/officestonks/deploy-ultra-simple.sh

rm -f /home/kdln/code/officestonks/Dockerfile.static
rm -f /home/kdln/code/officestonks/Dockerfile.scratch
rm -f /home/kdln/code/officestonks/Dockerfile.simple
rm -f /home/kdln/code/officestonks/railway.static.json

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

rm -f /home/kdln/code/officestonks/admin.js.patch
rm -f /home/kdln/code/officestonks/frontend-fix.js
rm -f /home/kdln/code/officestonks/quick-frontend-fix.js
rm -f /home/kdln/code/officestonks/final-frontend-fix.js
rm -f /home/kdln/code/officestonks/stock-reset-fix.js
rm -f /home/kdln/code/officestonks/test-mysql-connection.js

rm -f /home/kdln/code/officestonks/ADMIN_FIX_SUMMARY.md
rm -f /home/kdln/code/officestonks/CORS-FIX-FOR-FRONTEND.md
rm -f /home/kdln/code/officestonks/CORS-FIX-README.md
rm -f /home/kdln/code/officestonks/CORS-PROXY-README.md
rm -f /home/kdln/code/officestonks/cors-allowed-backend.md

```



The current database connection code in `/internal/repository/db.go` has hardcoded values that are causing connection issues. Replace it with the clean version in `db.go.clean`:

```bash
cp /home/kdln/code/officestonks/internal/repository/db.go.clean /home/kdln/code/officestonks/internal/repository/db.go
```

Key improvements in the clean version:
- Uses environment variables properly with Railway's standard variables
- Has simpler, more maintainable fallback logic
- Includes appropriate logging without excessive output
- Maintains the same connection parameters that previously worked


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


The README.md should be updated to reflect the current project structure. The backend code is now at the root level, not in a `/backend` directory as mentioned in the README.

Create an updated version of README.md:

```md

A real-time multiplayer stock market simulation game where players can trade stocks, form investment groups, and compete for the highest portfolio value.


Office Stonks is an online multiplayer stock market simulation that allows players to:
- Buy and sell virtual stocks based on real market dynamics
- See real-time price updates via WebSockets
- Compete on leaderboards with other players
- Chat with other players
- Manage their portfolios


- **Backend**: Go with standard library
- **Frontend**: React with a simple component library
- **Database**: MySQL
- **Hosting**: Railway
- **Real-time Updates**: WebSockets


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


- Git
- Go 1.20+
- Node.js 18+
- MySQL database


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

```bash
cd docker
docker-compose up
```


This application is deployed on Railway. See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.


We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.


This project is licensed under the MIT License - see the LICENSE file for details.
```


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


1. **Simplified Codebase**: Removing unnecessary files makes the codebase easier to understand and maintain.
2. **Improved Reliability**: The clean db.go and cors.go implementations have fewer edge cases and better handling.
3. **Reduced Confusion**: Eliminating temporary solutions prevents confusion about which files are actually used.
4. **Better Documentation**: Updated documentation ensures everyone understands the current state of the project.


1. **Environment Variables**: Create a `.env.example` file in the root directory showing all required environment variables.
2. **Branching Strategy**: Implement a clear Git workflow (e.g., GitFlow) for future development.
3. **CI/CD Pipeline**: Set up automated tests and deployments to prevent regression.
4. **Error Handling**: Review error handling throughout the codebase to ensure consistency.
5. **Logging**: Implement structured logging with appropriate log levels throughout the application.


This cleanup plan addresses the immediate issues with the repository structure, database connectivity, and CORS handling. By implementing these changes, the application will be more maintainable, reliable, and easier to deploy.
