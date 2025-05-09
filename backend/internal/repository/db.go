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
	// For development, you can hardcode these values
	username := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "DucukmJTCFzGLzfgcxnDiNnlHxFZyNzE")
	host := getEnv("DB_HOST", "mysql.railway.internal")
	port := getEnv("DB_PORT", "3306")
	dbname := getEnv("DB_NAME", "railway")

	// Create connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
		username, password, host, port, dbname)

	// Open connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established")
	DB = db
	return db, nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}