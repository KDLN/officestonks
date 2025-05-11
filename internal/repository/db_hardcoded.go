package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// InitDBHardcoded initializes the database connection with hardcoded values
// This is a fallback mechanism for when environment variables are not properly set
func InitDBHardcoded() (*sql.DB, error) {
	log.Println("Using hardcoded database connection parameters as fallback")
	
	// Hardcoded connection details - only for this specific Railway deployment
	username := "root"
	password := "EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY"
	host := "turntable.proxy.rlwy.net"
	port := "28889"
	dbname := "railway"
	
	log.Printf("Connecting to MySQL at %s:%s as user %s to database %s", host, port, username, dbname)
	
	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&allowNativePasswords=true&multiStatements=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		username, password, host, port, dbname)
	
	// Open connection
	log.Println("Opening hardcoded database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening hardcoded database connection: %v", err)
		return nil, err
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)
	
	// Verify connection with ping
	log.Println("Pinging database with hardcoded connection...")
	if err := db.Ping(); err != nil {
		log.Printf("Error pinging database with hardcoded connection: %v", err)
		return nil, err
	}
	
	log.Println("Hardcoded database connection successful")
	
	// Set global DB variable
	DB = db
	
	// Initialize schema
	log.Println("Initializing database schema with hardcoded connection...")
	if err := InitSchema(); err != nil {
		log.Printf("Warning: Failed to initialize schema: %v", err)
		log.Println("Application will continue but may encounter issues until database schema is properly initialized")
	} else {
		log.Println("Database schema initialization completed successfully")
	}
	
	return db, nil
}