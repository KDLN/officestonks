package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Constants for alternative connection methods
const (
	ALT_HOST     = "turntable.proxy.rlwy.net"
	ALT_PORT     = "28889"
	ALT_USERNAME = "root"
	ALT_PASSWORD = "EJhRJRIwfkyeXofeEDCGLnlwFuhWAHAY"
	ALT_DATABASE = "railway"
)

// Attempt various connection methods to the database
func TryAlternativeConnections() (*sql.DB, error) {
	var db *sql.DB
	var err error
	
	// Method 1: Standard connection string
	log.Println("Trying alternative connection method 1: Standard connection")
	dsn1 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		ALT_USERNAME, ALT_PASSWORD, ALT_HOST, ALT_PORT, ALT_DATABASE)
	db, err = sql.Open("mysql", dsn1)
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Alternative connection method 1 successful")
			return configureConnection(db), nil
		}
		log.Printf("Alternative connection method 1 ping failed: %v", err)
		db.Close()
	} else {
		log.Printf("Alternative connection method 1 open failed: %v", err)
	}
	
	// Method 2: With no TLS
	log.Println("Trying alternative connection method 2: No TLS")
	dsn2 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=false",
		ALT_USERNAME, ALT_PASSWORD, ALT_HOST, ALT_PORT, ALT_DATABASE)
	db, err = sql.Open("mysql", dsn2)
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Alternative connection method 2 successful")
			return configureConnection(db), nil
		}
		log.Printf("Alternative connection method 2 ping failed: %v", err)
		db.Close()
	} else {
		log.Printf("Alternative connection method 2 open failed: %v", err)
	}
	
	// Method 3: With all options
	log.Println("Trying alternative connection method 3: All options")
	dsn3 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=false&timeout=30s&readTimeout=30s&writeTimeout=30s&allowNativePasswords=true&multiStatements=true",
		ALT_USERNAME, ALT_PASSWORD, ALT_HOST, ALT_PORT, ALT_DATABASE)
	db, err = sql.Open("mysql", dsn3)
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Alternative connection method 3 successful")
			return configureConnection(db), nil
		}
		log.Printf("Alternative connection method 3 ping failed: %v", err)
		db.Close()
	} else {
		log.Printf("Alternative connection method 3 open failed: %v", err)
	}
	
	// Method 4: Raw connection format
	log.Println("Trying alternative connection method 4: Raw connection format")
	dsn4 := fmt.Sprintf("%s:%s@(%s:%s)/%s",
		ALT_USERNAME, ALT_PASSWORD, ALT_HOST, ALT_PORT, ALT_DATABASE)
	db, err = sql.Open("mysql", dsn4)
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Alternative connection method 4 successful")
			return configureConnection(db), nil
		}
		log.Printf("Alternative connection method 4 ping failed: %v", err)
		db.Close()
	} else {
		log.Printf("Alternative connection method 4 open failed: %v", err)
	}
	
	// Method 5: Custom driver params
	log.Println("Trying alternative connection method 5: Custom driver params")
	dsn5 := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowAllFiles=true&clientFoundRows=true",
		ALT_USERNAME, ALT_PASSWORD, ALT_HOST, ALT_PORT, ALT_DATABASE)
	db, err = sql.Open("mysql", dsn5)
	if err == nil {
		if err = db.Ping(); err == nil {
			log.Println("Alternative connection method 5 successful")
			return configureConnection(db), nil
		}
		log.Printf("Alternative connection method 5 ping failed: %v", err)
		db.Close()
	} else {
		log.Printf("Alternative connection method 5 open failed: %v", err)
	}
	
	return nil, fmt.Errorf("all alternative connection methods failed")
}

// Helper to configure connection pool settings
func configureConnection(db *sql.DB) *sql.DB {
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)
	return db
}