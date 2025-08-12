package repository

import (
        "database/sql"
        "fmt"
        "log"
        "strings"
        "time"
)

// RetryOptions defines options for retrying database operations
type RetryOptions struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryOptions provides sensible defaults for retry operations
var DefaultRetryOptions = RetryOptions{
	MaxRetries:  3,
	InitialWait: 100 * time.Millisecond,
	MaxWait:     1 * time.Second,
}

// RetryQuery executes a database query with retries for transient errors
func RetryQuery(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return RetryQueryWithOptions(db, DefaultRetryOptions, query, args...)
}

// RetryQueryWithOptions executes a database query with configurable retry options
func RetryQueryWithOptions(db *sql.DB, options RetryOptions, query string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows
	wait := options.InitialWait

	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying database query (attempt %d/%d) after error: %v", 
				attempt, options.MaxRetries, err)
			time.Sleep(wait)
			// Exponential backoff with cap
			wait *= 2
			if wait > options.MaxWait {
				wait = options.MaxWait
			}
		}

		rows, err = db.Query(query, args...)
		if err == nil {
			return rows, nil
		}

		// If this is not a connection error, don't retry
		if !isRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", options.MaxRetries, err)
}

// RetryExec executes a database exec operation with retries for transient errors
func RetryExec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	return RetryExecWithOptions(db, DefaultRetryOptions, query, args...)
}

// RetryExecWithOptions executes a database exec operation with configurable retry options
func RetryExecWithOptions(db *sql.DB, options RetryOptions, query string, args ...interface{}) (sql.Result, error) {
	var err error
	var result sql.Result
	wait := options.InitialWait

	for attempt := 0; attempt <= options.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying database exec (attempt %d/%d) after error: %v", 
				attempt, options.MaxRetries, err)
			time.Sleep(wait)
			// Exponential backoff with cap
			wait *= 2
			if wait > options.MaxWait {
				wait = options.MaxWait
			}
		}

		result, err = db.Exec(query, args...)
		if err == nil {
			return result, nil
		}

		// If this is not a connection error, don't retry
		if !isRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", options.MaxRetries, err)
}

// RetryQueryRow executes a database QueryRow operation with retries for transient errors
// Note: This returns a specialized RowScanner rather than sql.Row because sql.Row cannot be retried directly
func RetryQueryRow(db *sql.DB, query string, args ...interface{}) *RowScanner {
	return RetryQueryRowWithOptions(db, DefaultRetryOptions, query, args...)
}

// RetryQueryRowWithOptions executes a database QueryRow operation with configurable retry options
func RetryQueryRowWithOptions(db *sql.DB, options RetryOptions, query string, args ...interface{}) *RowScanner {
	return &RowScanner{
		db:      db,
		options: options,
		query:   query,
		args:    args,
	}
}

// RowScanner is a wrapper around sql.Row that supports retrying scan operations
type RowScanner struct {
	db      *sql.DB
	options RetryOptions
	query   string
	args    []interface{}
	err     error
}

// Scan implements the same interface as sql.Row.Scan but with retry logic
func (rs *RowScanner) Scan(dest ...interface{}) error {
	wait := rs.options.InitialWait

	for attempt := 0; attempt <= rs.options.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying database row scan (attempt %d/%d) after error: %v",
				attempt, rs.options.MaxRetries, rs.err)
			time.Sleep(wait)
			// Exponential backoff with cap
			wait *= 2
			if wait > rs.options.MaxWait {
				wait = rs.options.MaxWait
			}
		}

		row := rs.db.QueryRow(rs.query, rs.args...)
		rs.err = row.Scan(dest...)
		
		if rs.err == nil {
			return nil
		}

		// If this is not a connection error, don't retry
		if !isRetryableError(rs.err) {
			return rs.err
		}
	}

	return fmt.Errorf("failed after %d retries: %w", rs.options.MaxRetries, rs.err)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	// Add specific error types that should be retried
	// For MySQL, common retryable errors include connection issues and deadlocks
	errorMsg := err.Error()
	
	// Check for common MySQL connection errors
	retryableErrors := []string{
		"connection reset by peer",
		"broken pipe",
		"lost connection",
		"EOF",
		"connection refused",
		"connection timed out",
		"no connection",
		"server has gone away",
		"bad connection",
		"cannot read response",
		"unexpected EOF",
		"write: broken pipe",
		"deadline exceeded",
	}

	for _, errPattern := range retryableErrors {
		if containsIgnoreCase(errorMsg, errPattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains a substring, case-insensitive
func containsIgnoreCase(s, substr string) bool {
        return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}