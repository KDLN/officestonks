# Repository Cleanup and Audit - May 2025 Update

This document outlines a comprehensive audit of the OfficeStonks repository, identifying issues, unnecessary files, and recommendations for cleanup.

## Unnecessary Files to Remove

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

## Database Connection Issues

### Current Issues

1. **Environment Variable Handling**:
   - Inconsistent handling of environment variables
   - Hardcoded database connection details in recent commits

2. **Default Fallbacks**:
   - Code is falling back to `mysql.railway.internal` despite environment variables
   - Recent changes have overridden the Railway-provided environment

### Key Files to Fix

1. **Database Connection**:
   - `/internal/repository/db.go`: Needs consistent environment variable handling
   - Recent changes should be reverted to the version that previously worked

2. **Configuration**:
   - Ensure all environment variables are properly passed through in Railway
   - Double-check that Railway is providing the correct database URL

## Codebase Issues

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

## Action Plan

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

## Database Connection Recommendations

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

## Deployment Testing

After cleanup, perform the following tests:

1. **Local Testing**:
   - Test database connection locally
   - Verify all API endpoints work
   - Confirm frontend can connect to backend

2. **Railway Testing**:
   - Deploy with clean configuration
   - Verify database connection works
   - Test all endpoints through the frontend

## Next Steps

1. Complete the repository cleanup
2. Revert database connection to last known working version
3. Implement a simpler CORS policy
4. Test thoroughly before final deployment