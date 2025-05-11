package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
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
	
	// Create connection string with better timeouts
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s", 
		username, password, host, port, dbname)

	// Open connection
	fmt.Println("Opening database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Create a test table
	fmt.Println("Creating test table...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS connection_test (
			id INT AUTO_INCREMENT PRIMARY KEY,
			iteration INT NOT NULL,
			thread_id INT NOT NULL,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			data VARCHAR(100)
		)
	`)
	if err != nil {
		log.Fatalf("Error creating test table: %v", err)
	}

	// Get number of concurrent threads from command line
	threads := 5
	if len(os.Args) > 1 {
		if t, err := strconv.Atoi(os.Args[1]); err == nil && t > 0 {
			threads = t
		}
	}

	// Get number of iterations from command line
	iterations := 20
	if len(os.Args) > 2 {
		if i, err := strconv.Atoi(os.Args[2]); err == nil && i > 0 {
			iterations = i
		}
	}

	fmt.Printf("Starting stress test with %d threads, %d iterations each...\n", threads, iterations)

	// Create a wait group to wait for all threads
	var wg sync.WaitGroup
	wg.Add(threads)
	
	// Channel to collect errors
	errChan := make(chan error, threads*iterations)

	// Start the threads
	for t := 0; t < threads; t++ {
		go func(threadID int) {
			defer wg.Done()
			
			for i := 0; i < iterations; i++ {
				// Random sleep to simulate real-world usage patterns
				time.Sleep(time.Duration(50+i*10) * time.Millisecond)
				
				// Insert a record
				data := fmt.Sprintf("Thread %d, Iteration %d, Time %s", threadID, i, time.Now().Format(time.RFC3339))
				_, err := db.Exec("INSERT INTO connection_test (iteration, thread_id, data) VALUES (?, ?, ?)",
					i, threadID, data)
				
				if err != nil {
					fmt.Printf("Thread %d, Iteration %d: Error inserting record: %v\n", threadID, i, err)
					errChan <- err
					continue
				}
				
				// Query the record back
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM connection_test WHERE thread_id = ? AND iteration = ?", 
					threadID, i).Scan(&count)
				
				if err != nil {
					fmt.Printf("Thread %d, Iteration %d: Error querying record: %v\n", threadID, i, err)
					errChan <- err
					continue
				}
				
				fmt.Printf("Thread %d, Iteration %d: Success, found %d records\n", threadID, i, count)
			}
		}(t)
	}

	// Wait for all threads to complete
	wg.Wait()
	close(errChan)

	// Count errors
	errorCount := 0
	for err := range errChan {
		errorCount++
		fmt.Printf("Error: %v\n", err)
	}

	// Query the final count
	var totalCount int
	err = db.QueryRow("SELECT COUNT(*) FROM connection_test").Scan(&totalCount)
	if err != nil {
		log.Fatalf("Error getting final count: %v", err)
	}

	// Print the results
	fmt.Printf("\nTest completed with %d errors out of %d operations.\n", errorCount, threads*iterations*2)
	fmt.Printf("Expected %d records, found %d records.\n", threads*iterations, totalCount)
	
	if errorCount == 0 && totalCount == threads*iterations {
		fmt.Println("TEST PASSED: All operations succeeded.")
	} else {
		fmt.Println("TEST FAILED: Some operations failed.")
	}
}