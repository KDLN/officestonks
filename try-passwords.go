package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Define host and port
	host := "turntable.proxy.rlwy.net"
	port := "28889"
	user := "root"
	dbname := "railway"
	
	// Different passwords to try
	passwords := []string{
		"EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY",
		"DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE",
		os.Getenv("MY_DB_PASSWORD"), // Try from environment variable if set
	}
	
	// Try each password
	for i, password := range passwords {
		// Skip empty passwords
		if password == "" {
			continue
		}
		
		fmt.Printf("\n========= Trying password %d =========\n", i+1)
		
		// Create DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s", user, password, host, port, dbname)
		fmt.Printf("Connecting to %s@%s:%s/%s\n", user, host, port, dbname)
		
		// Try to connect
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Printf("Error opening connection: %v\n", err)
			continue
		}
		
		// Set short timeout
		db.SetConnMaxLifetime(10 * time.Second)
		
		// Ping
		fmt.Println("Pinging...")
		err = db.Ping()
		if err != nil {
			fmt.Printf("❌ Connection failed: %v\n", err)
			db.Close()
			continue
		}
		
		// Try a query
		fmt.Println("Running test query...")
		var version string
		err = db.QueryRow("SELECT VERSION()").Scan(&version)
		if err != nil {
			fmt.Printf("❌ Query failed: %v\n", err)
			db.Close()
			continue
		}
		
		fmt.Printf("✅ SUCCESS! MySQL version: %s\n", version)
		
		// Try to list tables
		fmt.Println("Listing tables...")
		rows, err := db.Query("SHOW TABLES")
		if err != nil {
			fmt.Printf("❌ Table listing failed: %v\n", err)
			db.Close()
			continue
		}
		
		// Collect tables
		tables := []string{}
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				fmt.Printf("Error scanning table name: %v\n", err)
				continue
			}
			tables = append(tables, tableName)
		}
		rows.Close()
		
		fmt.Printf("Found %d tables: %v\n", len(tables), tables)
		
		// Close DB
		db.Close()
		
		// If we got here, the password works - save it to a file
		workingPassword := fmt.Sprintf("DB_PASSWORD=%s\n", password)
		os.WriteFile("working-password.env", []byte(workingPassword), 0644)
		
		fmt.Println("========= Working password found and saved to 'working-password.env' =========")
		return
	}
	
	fmt.Println("❌ All password attempts failed")
}