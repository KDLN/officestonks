package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Connection parameters
	username := "root"
	password := "EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY"
	host := "turntable.proxy.rlwy.net"
	port := "28889"
	dbname := "railway"

	// Log connection details (excluding password)
	fmt.Printf("Database connection details:\n")
	fmt.Printf("  Host: %s\n", host)
	fmt.Printf("  Port: %s\n", port)
	fmt.Printf("  User: %s\n", username)
	fmt.Printf("  Database: %s\n", dbname)
	fmt.Printf("  (Password hidden for security)\n")

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=10s", 
		username, password, host, port, dbname)

	// Open connection
	fmt.Println("Opening database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Try to ping the database
	fmt.Println("Pinging database to verify connection...")
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatalf("Error pinging database: %v", pingErr)
	}
	fmt.Println("Successfully pinged database!")

	// Test a simple query
	fmt.Println("Running a test query...")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Fatalf("Error running test query: %v", err)
	}
	fmt.Printf("MySQL version: %s\n", version)

	// Try to query tables
	fmt.Println("Checking database tables...")
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("Error querying tables: %v", err)
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("Error scanning table name: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	if len(tables) > 0 {
		fmt.Printf("Found %d tables: %v\n", len(tables), tables)
	} else {
		fmt.Println("No tables found in database")
	}

	// Test a query against users table
	fmt.Println("Checking users table...")
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Fatalf("Error querying users table: %v", err)
	}
	fmt.Printf("Found %d users in database\n", userCount)

	fmt.Println("Connection test completed successfully")
}