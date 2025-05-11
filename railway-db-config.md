# Railway Database Configuration

Based on our tests, the MySQL database at `turntable.proxy.rlwy.net:28889` is accessible and properly configured with the necessary tables.

## Recommended Configuration

To ensure stable database connections across all Railway deployments, the following configuration is recommended:

### 1. Environment Variables

Make sure the following environment variables are set in your Railway project:

```
MYSQLHOST=turntable.proxy.rlwy.net
MYSQLPORT=28889
MYSQLUSER=root
MYSQLPASSWORD=EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY
MYSQLDATABASE=railway
```

### 2. Connection String Parameters

In your application, use the following DSN parameters for optimal stability:

```go
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&allowNativePasswords=true&multiStatements=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
    username, password, host, port, dbname)
```

Key parameters:
- `parseTime=true` - Support for parsing time values
- `tls=false` - Disable SSL/TLS (Railway doesn't require it)
- `allowNativePasswords=true` - Support modern MySQL authentication
- `multiStatements=true` - Allow multiple SQL statements in one query
- `timeout=30s` - Connection timeout
- `readTimeout=30s` - Read timeout
- `writeTimeout=30s` - Write timeout

### 3. Connection Pool Settings

Optimize your connection pool settings:

```go
// Set conservative connection pool limits to avoid overwhelming the database
db.SetMaxOpenConns(10)         // Maximum number of open connections
db.SetMaxIdleConns(5)          // Maximum number of idle connections
db.SetConnMaxLifetime(1 * time.Minute) // Maximum lifetime of a connection
db.SetConnMaxIdleTime(30 * time.Second) // Maximum idle time of a connection
```

### 4. Error Handling and Reconnection

Make sure your application handles temporary connection failures gracefully:

- Implement connection retry logic
- Use backoff strategies for reconnection attempts
- Log detailed error information
- Have a fallback mechanism if the main connection fails

### 5. Verify Connection at Startup

Always verify the database connection at application startup:

```go
// Ping to verify connection
if err := db.Ping(); err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Run a simple query to further verify connection
var version string
if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
    log.Fatalf("Failed to query database: %v", err)
}
```

These recommendations should ensure stable database connectivity in your Railway deployment.