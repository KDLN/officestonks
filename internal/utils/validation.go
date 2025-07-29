package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Input validation constants
const (
	MaxUsernameLength = 50
	MinUsernameLength = 3
	MaxPasswordLength = 128
	MinPasswordLength = 8
	MaxMessageLength  = 1000
	MaxNewsTitle      = 200
	MaxNewsContent    = 5000
	MaxEmailLength    = 254
)

// Regex patterns
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	sqlPattern    = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script|javascript|<script|onload|onerror)`)
)

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if len(username) < MinUsernameLength {
		return ValidationError{"username", fmt.Sprintf("must be at least %d characters", MinUsernameLength)}
	}
	if len(username) > MaxUsernameLength {
		return ValidationError{"username", fmt.Sprintf("must be less than %d characters", MaxUsernameLength)}
	}
	if !usernameRegex.MatchString(username) {
		return ValidationError{"username", "can only contain letters, numbers, and underscores"}
	}
	
	// Check for reserved usernames
	reserved := []string{"admin", "root", "system", "null", "undefined", "api", "www", "ftp", "mail", "test"}
	lowerUsername := strings.ToLower(username)
	for _, r := range reserved {
		if lowerUsername == r {
			return ValidationError{"username", "username is reserved"}
		}
	}
	
	return nil
}

// ValidatePassword validates a password
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return ValidationError{"password", fmt.Sprintf("must be at least %d characters", MinPasswordLength)}
	}
	if len(password) > MaxPasswordLength {
		return ValidationError{"password", fmt.Sprintf("must be less than %d characters", MaxPasswordLength)}
	}
	
	// Check for complexity requirements
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if !hasUpper || !hasLower || !hasDigit {
		return ValidationError{"password", "must contain at least one uppercase letter, one lowercase letter, and one digit"}
	}
	
	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return ValidationError{"email", "is required"}
	}
	if len(email) > MaxEmailLength {
		return ValidationError{"email", fmt.Sprintf("must be less than %d characters", MaxEmailLength)}
	}
	if !emailRegex.MatchString(email) {
		return ValidationError{"email", "invalid email format"}
	}
	return nil
}

// ValidateMessage validates a chat message
func ValidateMessage(message string) error {
	if len(message) == 0 {
		return ValidationError{"message", "cannot be empty"}
	}
	if len(message) > MaxMessageLength {
		return ValidationError{"message", fmt.Sprintf("must be less than %d characters", MaxMessageLength)}
	}
	
	// Check for potential injection attempts
	if sqlPattern.MatchString(message) {
		return ValidationError{"message", "contains potentially unsafe content"}
	}
	
	return nil
}

// ValidateNewsTitle validates a news title
func ValidateNewsTitle(title string) error {
	if len(title) == 0 {
		return ValidationError{"title", "is required"}
	}
	if len(title) > MaxNewsTitle {
		return ValidationError{"title", fmt.Sprintf("must be less than %d characters", MaxNewsTitle)}
	}
	return nil
}

// ValidateNewsContent validates news content
func ValidateNewsContent(content string) error {
	if len(content) == 0 {
		return ValidationError{"content", "is required"}
	}
	if len(content) > MaxNewsContent {
		return ValidationError{"content", fmt.Sprintf("must be less than %d characters", MaxNewsContent)}
	}
	return nil
}

// ValidateTradeQuantity validates a trade quantity
func ValidateTradeQuantity(quantity int) error {
	if quantity <= 0 {
		return ValidationError{"quantity", "must be greater than 0"}
	}
	if quantity > 10000 {
		return ValidationError{"quantity", "cannot exceed 10,000 shares"}
	}
	return nil
}

// ValidateStockID validates a stock ID
func ValidateStockID(stockID int) error {
	if stockID <= 0 {
		return ValidationError{"stock_id", "must be a valid stock ID"}
	}
	return nil
}

// SanitizeString sanitizes a string for safe HTML output
func SanitizeString(input string) string {
	// Remove any HTML tags and escape HTML entities
	sanitized := html.EscapeString(input)
	
	// Remove any remaining potentially dangerous content
	sanitized = strings.ReplaceAll(sanitized, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "javascript:", "")
	sanitized = strings.ReplaceAll(sanitized, "data:", "")
	
	return sanitized
}

// ValidateAndSanitizeMessage validates and sanitizes a message
func ValidateAndSanitizeMessage(message string) (string, error) {
	if err := ValidateMessage(message); err != nil {
		return "", err
	}
	return SanitizeString(message), nil
}

// IsValidJSON checks if a string is valid JSON (basic check)
func IsValidJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		   (strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// ContainsSQLInjection checks for common SQL injection patterns
func ContainsSQLInjection(input string) bool {
	return sqlPattern.MatchString(input)
}