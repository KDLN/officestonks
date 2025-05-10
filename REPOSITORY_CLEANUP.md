# Repository Cleanup

This document describes the cleanup performed on the repository structure to fix duplication issues.

## Background

The repository previously had a nested structure with code duplicated across multiple directories:
- `/officestonks/backend/` (older version)
- `/officestonks/backend/backend/` (newer version with chat and leaderboard functionality)

This structure was causing issues with Docker builds and Railway deployments.

## Cleanup Steps

1. **Relocated Core Files**
   - Moved all Go code from `/backend/backend/` to standard Go project directories:
     - `cmd/api/`: Main application entrypoint
     - `internal/`: Application internal packages
     - `pkg/`: Shared utilities and reusable components
   - Moved configuration files to project root:
     - `go.mod` and `go.sum`
     - `schema.sql` for database initialization

2. **Updated Startup Scripts**
   - Created `/start.sh`: Entry point that calls start-server.sh
   - Enhanced `/start-server.sh`: Main script that initializes the database and starts the server
   - Added `/init-db.sh`: Dedicated script for database initialization

3. **Updated Dockerfile**
   - Modified to use the new directory structure
   - Simplified build process to work with standard Go project layout
   - Ensured all necessary files are copied to the container
   - Added debugging output to help troubleshoot build issues

4. **Security Enhancements**
   - Added rate limiting middleware to protect against abuse
   - Implemented sliding window approach for IP-based rate limiting
   - Added IP extraction from various headers to work behind proxies

## Current Structure

The repository now follows a more standard Go project layout:

```
/officestonks/
├── cmd/
│   └── api/
│       ├── cors.go
│       └── main.go
├── internal/
│   ├── auth/
│   ├── handlers/
│   ├── middleware/
│   ├── models/
│   ├── repository/
│   ├── services/
│   ├── tests/
│   └── websocket/
├── pkg/
│   └── market/
│       └── simulation.go
├── go.mod
├── go.sum
├── schema.sql
├── start.sh
├── start-server.sh
├── init-db.sh
└── Dockerfile
```

## Verification

A verification script has been created at `/verify-deployment.sh` to ensure that all necessary files are in the correct locations and that the application can build successfully.

## Next Steps

1. **Remove Duplicate Code**
   - After confirming that the new structure works correctly, the old `/backend/backend/` directory should be removed

2. **Test Deployment**
   - Deploy the updated repository to Railway to verify that the build and startup process works correctly

3. **Update Documentation**
   - Update other documentation files to reflect the new repository structure
