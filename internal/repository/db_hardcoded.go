package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// InitDBHardcoded initializes the database connection with fallback values
// This is a fallback mechanism for when primary connection method fails
func InitDBHardcoded() (*sql.DB, error) {
	log.Println("Using fallback database connection parameters")

	// Get connection details from environment variables with fallbacks
	username := getEnv("FALLBACK_DB_USER", "root")
	password := getEnv("FALLBACK_DB_PASSWORD", "") // Empty default for security
	host := getEnv("FALLBACK_DB_HOST", "turntable.proxy.rlwy.net")
	port := getEnv("FALLBACK_DB_PORT", "28889")
	dbname := getEnv("FALLBACK_DB_NAME", "railway")
	
	log.Printf("Connecting to MySQL at %s:%s as user %s to database %s", host, port, username, dbname)

	// Check if password is empty
	if password == "" {
		log.Println("Error: No fallback database password provided")
		return nil, fmt.Errorf("fallback database password not provided, set FALLBACK_DB_PASSWORD environment variable")
	}

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&allowNativePasswords=true&multiStatements=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		username, password, host, port, dbname)

	// Open connection
	log.Println("Opening fallback database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening fallback database connection: %v", err)
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Verify connection with ping
	log.Println("Pinging database with fallback connection...")
	if err := db.Ping(); err != nil {
		log.Printf("Error pinging database with fallback connection: %v", err)
		return nil, err
	}

	log.Println("Fallback database connection successful")

	// Set global DB variable
	DB = db

	// Initialize schema
	log.Println("Initializing database schema with fallback connection...")
	if err := InitSchema(); err != nil {
		log.Printf("Warning: Failed to initialize schema: %v", err)
		log.Println("Application will continue but may encounter issues until database schema is properly initialized")
	} else {
		log.Println("Database schema initialization completed successfully")
	}

	return db, nil
}