package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Connection parameters - replace with actual values or use environment variables
	username := getEnv("MYSQLUSER", getEnv("DB_USER", "root"))
	password := getEnv("MYSQLPASSWORD", getEnv("DB_PASSWORD", "EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY"))
	host := getEnv("MYSQLHOST", getEnv("DB_HOST", "turntable.proxy.rlwy.net"))
	port := getEnv("MYSQLPORT", getEnv("DB_PORT", "28889"))
	dbname := getEnv("MYSQLDATABASE", getEnv("MYSQL_DATABASE", getEnv("DB_NAME", "railway")))

	// Log all environment variables for database connection
	log.Println("Environment variables:")
	for _, env := range os.Environ() {
		if strings.Contains(strings.ToLower(env), "db") || strings.Contains(strings.ToLower(env), "mysql") {
			log.Println("  " + env)
		}
	}

	// Print connection parameters 
	log.Printf("Connection parameters:")
	log.Printf("Host: %s", host)
	log.Printf("Port: %s", port)
	log.Printf("User: %s", username)
	log.Printf("Database: %s", dbname)
	
	// Try to resolve the host
	log.Printf("Attempting to resolve host: %s", host)
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Printf("Failed to resolve host: %v", err)
	} else {
		log.Printf("Resolved addresses for %s: %v", host, ips)
		
		// Show IP address types
		for _, ip := range ips {
			if ip.To4() != nil {
				log.Printf("IPv4 address: %s", ip.String())
			} else {
				log.Printf("IPv6 address: %s", ip.String())
			}
		}
	}
	
	// First try with regular connection
	log.Println("\nAttempt 1: Connecting with standard parameters...")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		log.Println("✅ Standard connection successful!")
		return
	}
	
	log.Printf("❌ Standard connection failed: %v", err)
	
	// Try with no SSL
	log.Println("\nAttempt 2: Connecting with SSL disabled...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s&tls=false",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		log.Println("✅ Connection with SSL disabled successful!")
		return
	}
	
	log.Printf("❌ Connection with SSL disabled failed: %v", err)
	
	// Try with all parameters
	log.Println("\nAttempt 3: Connecting with all parameters...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&allowNativePasswords=true&multiStatements=true&timeout=60s&readTimeout=60s&writeTimeout=60s",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		log.Println("✅ Connection with all parameters successful!")
		return
	}
	
	log.Printf("❌ Connection with all parameters failed: %v", err)
	
	// Try direct IP if we have it
	if len(ips) > 0 {
		var ipv4 string
		for _, ip := range ips {
			if ip.To4() != nil {
				ipv4 = ip.String()
				break
			}
		}
		
		if ipv4 != "" {
			log.Printf("\nAttempt 4: Connecting with direct IPv4 address: %s", ipv4)
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&timeout=60s",
				username, password, ipv4, port, dbname)
			
			err = tryConnection(dsn)
			if err == nil {
				log.Println("✅ Direct IPv4 connection successful!")
				return
			}
			
			log.Printf("❌ Direct IPv4 connection failed: %v", err)
		}
	}
	
	// Last attempt with minimal options
	log.Println("\nAttempt 5: Connecting with minimal options...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		log.Println("✅ Minimal options connection successful!")
		return
	}
	
	log.Printf("❌ Minimal options connection failed: %v", err)
	
	// Try checking if port is open
	log.Printf("\nChecking if port %s is open on host %s...", port, host)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 10*time.Second)
	if err != nil {
		log.Printf("❌ TCP connection failed: %v", err)
	} else {
		log.Printf("✅ TCP connection succeeded!")
		conn.Close()
	}
	
	log.Println("\n❌ All connection attempts failed")
}

// Helper function to try connecting with a given DSN
func tryConnection(dsn string) error {
	// Strip password for logging
	dsnForLogging := maskDSN(dsn)
	log.Printf("Opening connection with DSN: %s", dsnForLogging)
	
	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()
	
	// Set connection parameters
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(60 * time.Second)
	
	// Ping database with timeout
	log.Println("Pinging database...")
	
	// Create a timeout channel
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(15 * time.Second)
		timeout <- true
	}()
	
	// Create a result channel
	pingResult := make(chan error, 1)
	go func() {
		pingResult <- db.Ping()
	}()
	
	// Wait for either ping result or timeout
	select {
	case err := <-pingResult:
		if err != nil {
			return fmt.Errorf("ping failed: %w", err)
		}
	case <-timeout:
		return fmt.Errorf("ping timed out after 15 seconds")
	}
	
	log.Println("Ping successful!")
	
	// Try a simple query
	log.Println("Executing test query...")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	
	log.Printf("Query successful! MySQL version: %s", version)
	
	// Get connection details
	log.Println("Getting connection details...")
	var processInfo []string
	rows, err := db.Query("SHOW PROCESSLIST")
	if err != nil {
		log.Printf("Warning: Could not get process list: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id, user, host, db, command, time, state, info sql.NullString
			if err := rows.Scan(&id, &user, &host, &db, &command, &time, &state, &info); err != nil {
				log.Printf("Warning: Error scanning row: %v", err)
				continue
			}
			processInfo = append(processInfo, fmt.Sprintf("Connection id=%s, user=%s, host=%s", 
				orEmpty(id), orEmpty(user), orEmpty(host)))
		}
	}
	
	if len(processInfo) > 0 {
		log.Printf("Active connections: %d", len(processInfo))
		for _, info := range processInfo {
			log.Printf("  %s", info)
		}
	}
	
	// Show tables
	log.Println("Listing tables...")
	rows, err = db.Query("SHOW TABLES")
	if err != nil {
		return fmt.Errorf("listing tables failed: %w", err)
	}
	defer rows.Close()
	
	tables := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("scanning table name failed: %w", err)
		}
		tables = append(tables, tableName)
	}
	
	log.Printf("Found %d tables: %v", len(tables), tables)
	
	return nil
}

// Get environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Mask password in DSN for safe logging
func maskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		return dsn
	}
	
	authParts := strings.Split(parts[0], ":")
	if len(authParts) != 2 {
		return dsn
	}
	
	return authParts[0] + ":***@" + parts[1]
}

// Helper for null strings
func orEmpty(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "NULL"
}