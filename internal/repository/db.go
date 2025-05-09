package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	// Get database connection details from environment variables
	// Check for Railway's MySQL environment variables first, then fallback to generic ones
	username := getEnv("MYSQLUSER", getEnv("DB_USER", "root"))
	password := getEnv("MYSQLPASSWORD", getEnv("DB_PASSWORD", "DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE"))

	// Try to use the public MySQL URL if available
	// Default to the public MySQL URL provided
	host := getEnv("MYSQLHOST", getEnv("DB_HOST", "caboose.proxy.rlwy.net"))
	port := getEnv("MYSQLPORT", getEnv("DB_PORT", "40558"))
	dbname := getEnv("MYSQLDATABASE", getEnv("MYSQL_DATABASE", getEnv("DB_NAME", "railway")))

	// Log database connection details (excluding password)
	log.Printf("Database connection details:")
	log.Printf("  Host: %s", host)
	log.Printf("  Port: %s", port)
	log.Printf("  User: %s", username)
	log.Printf("  Database: %s", dbname)
	log.Printf("  Environment variables found: %v",
		map[string]bool{
			"MYSQLUSER":     os.Getenv("MYSQLUSER") != "",
			"MYSQLPASSWORD": os.Getenv("MYSQLPASSWORD") != "",
			"MYSQLHOST":     os.Getenv("MYSQLHOST") != "",
			"MYSQLPORT":     os.Getenv("MYSQLPORT") != "",
			"MYSQLDATABASE": os.Getenv("MYSQLDATABASE") != "",
			"DB_USER":       os.Getenv("DB_USER") != "",
			"DB_PASSWORD":   os.Getenv("DB_PASSWORD") != "",
			"DB_HOST":       os.Getenv("DB_HOST") != "",
			"DB_PORT":       os.Getenv("DB_PORT") != "",
			"DB_NAME":       os.Getenv("DB_NAME") != "",
		})
	log.Printf("  (Password hidden for security)")

	// Create connection string with SSL disabled for Railway
	// Add allowNativePasswords=true and multiStatements=true for better compatibility
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&allowNativePasswords=true&multiStatements=true",
		username, password, host, port, dbname)

	// Open connection
	log.Println("Opening database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Check connection
	log.Println("Pinging database to verify connection...")
	if err := db.Ping(); err != nil {
		log.Printf("Error pinging database: %v", err)
		return nil, err
	}

	log.Println("Database connection established successfully")

	// Test query to verify connection
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Printf("Warning: Could not query database version: %v", err)
	} else {
		log.Printf("Connected to MySQL version: %s", version)
	}

	// Set global DB variable first
	DB = db

	// Initialize database schema
	log.Println("Initializing database schema...")
	err = InitSchema()
	if err != nil {
		log.Printf("Warning: Failed to initialize schema: %v", err)
		log.Println("Application will continue but may encounter issues until database schema is properly initialized")
	} else {
		log.Println("Database schema initialization completed successfully")
	}

	// Check if tables exist after initialization
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Printf("Warning: Could not query tables: %v", err)
	} else {
		defer rows.Close()

		tables := []string{}
		var tableName string

		for rows.Next() {
			if err := rows.Scan(&tableName); err != nil {
				log.Printf("Warning: Error scanning table name: %v", err)
				continue
			}
			tables = append(tables, tableName)
		}

		if len(tables) > 0 {
			log.Printf("Found %d tables in database: %v", len(tables), tables)
		} else {
			log.Printf("Warning: No tables found in database even after schema initialization.")
		}
	}

	return db, nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		log.Printf("Environment variable %s found with custom value", key)
		return value
	}
	log.Printf("Environment variable %s not found, using default value", key)
	return defaultValue
}