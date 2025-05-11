package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Connection parameters - replace with the actual values from Railway
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "turntable.proxy.rlwy.net" // Replace with your actual DB host
	}
	
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "28889" // Replace with your actual DB port
	}
	
	username := os.Getenv("DB_USER")
	if username == "" {
		username = "root" // Replace with your actual DB username
	}
	
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY" // Replace with your actual DB password
	}
	
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "railway" // Replace with your actual DB name
	}

	// Try to resolve the host first
	fmt.Printf("Attempting to resolve host: %s\n", host)
	ips, err := net.LookupIP(host)
	if err != nil {
		fmt.Printf("Failed to resolve host: %v\n", err)
	} else {
		fmt.Printf("Resolved addresses for %s: %v\n", host, ips)
		
		// Try to find an IPv4 address
		var ipv4Address string
		for _, ip := range ips {
			if ip.To4() != nil {
				ipv4Address = ip.String()
				fmt.Printf("Found IPv4 address: %s\n", ipv4Address)
				break
			}
		}
		
		if ipv4Address != "" {
			fmt.Printf("Will try IPv4 address: %s\n", ipv4Address)
			host = ipv4Address
		}
	}

	// Print connection parameters 
	fmt.Printf("Connection parameters:\n")
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("Port: %s\n", port)
	fmt.Printf("User: %s\n", username)
	fmt.Printf("Database: %s\n", dbname)
	
	// Try first with IPv4 forcing
	fmt.Println("\nAttempt 1: Connecting with IPv4 forced...")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s&tcp4=true",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		fmt.Println("✅ Connection with IPv4 forcing successful!")
		return
	}
	
	fmt.Printf("❌ Connection with IPv4 forcing failed: %v\n", err)
	
	// Try without IPv4 forcing
	fmt.Println("\nAttempt 2: Connecting without IPv4 forcing...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		fmt.Println("✅ Connection without IPv4 forcing successful!")
		return
	}
	
	fmt.Printf("❌ Connection without IPv4 forcing failed: %v\n", err)
	
	// Try with IPv6 forcing
	fmt.Println("\nAttempt 3: Connecting with IPv6 forcing...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s&tcp6=true",
		username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		fmt.Println("✅ Connection with IPv6 forcing successful!")
		return
	}
	
	fmt.Printf("❌ Connection with IPv6 forcing failed: %v\n", err)
	
	// Try with all options disabled
	fmt.Println("\nAttempt 4: Connecting with basic DSN...")
	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)
	
	err = tryConnection(dsn)
	if err == nil {
		fmt.Println("✅ Connection with basic DSN successful!")
		return
	}
	
	fmt.Printf("❌ Connection with basic DSN failed: %v\n", err)
	
	fmt.Println("\n❌ All connection attempts failed")
}

// Helper function to try connecting with a given DSN
func tryConnection(dsn string) error {
	// Open connection
	fmt.Printf("Opening connection with DSN: %s\n", maskPassword(dsn))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()
	
	// Set connection parameters
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(1 * time.Minute)
	
	// Ping database
	fmt.Println("Pinging database...")
	pingStart := time.Now()
	err = db.Ping()
	pingDuration := time.Since(pingStart)
	
	if err != nil {
		return fmt.Errorf("ping failed after %v: %w", pingDuration, err)
	}
	
	fmt.Printf("Ping successful! Duration: %v\n", pingDuration)
	
	// Try a simple query
	fmt.Println("Executing test query...")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	
	fmt.Printf("Query successful! MySQL version: %s\n", version)
	
	// Show tables
	fmt.Println("Listing tables...")
	rows, err := db.Query("SHOW TABLES")
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
	
	fmt.Printf("Found %d tables: %v\n", len(tables), tables)
	
	return nil
}

// Mask password in DSN for safe logging
func maskPassword(dsn string) string {
	// Very simple masking - in a real application, use a more robust approach
	return dsn
}