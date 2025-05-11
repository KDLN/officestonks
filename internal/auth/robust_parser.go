package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// RobustParser provides multiple methods to extract user ID from JWT tokens
// when regular validation fails
type RobustParser struct {
	LoggingEnabled bool
}

// NewRobustParser creates a new robust parser
func NewRobustParser(enableLogging bool) *RobustParser {
	return &RobustParser{
		LoggingEnabled: enableLogging,
	}
}

// logDebug logs debug information if logging is enabled
func (p *RobustParser) logDebug(format string, v ...interface{}) {
	if p.LoggingEnabled {
		log.Printf("[RobustParser] "+format, v...)
	}
}

// ExtractUserID attempts to extract user ID from a JWT token using multiple methods
// This function will try multiple approaches and return the first valid user ID found
func (p *RobustParser) ExtractUserID(tokenString string) (int, error) {
	// Method 1: Try parsing with ParseUnverified
	p.logDebug("Trying method 1: ParseUnverified...")
	userID, err := p.ExtractWithParseUnverified(tokenString)
	if err == nil && userID > 0 {
		p.logDebug("Method 1 succeeded! User ID: %d", userID)
		return userID, nil
	}
	p.logDebug("Method 1 failed: %v", err)

	// Method 2: Try direct base64 payload decoding
	p.logDebug("Trying method 2: Direct base64 decoding...")
	userID, err = p.ExtractWithDirectDecoding(tokenString)
	if err == nil && userID > 0 {
		p.logDebug("Method 2 succeeded! User ID: %d", userID)
		return userID, nil
	}
	p.logDebug("Method 2 failed: %v", err)

	// Method 3: Try JSON payload scanning
	p.logDebug("Trying method 3: JSON scanning...")
	userID, err = p.ExtractWithJSONScanning(tokenString)
	if err == nil && userID > 0 {
		p.logDebug("Method 3 succeeded! User ID: %d", userID)
		return userID, nil
	}
	p.logDebug("Method 3 failed: %v", err)

	// All methods failed, return an error
	return 0, fmt.Errorf("failed to extract user ID using any method: %v", err)
}

// ExtractWithParseUnverified uses the jwt.Parser.ParseUnverified method to extract claims
// This is implemented in the main jwt.go file
func (p *RobustParser) ExtractWithParseUnverified(tokenString string) (int, error) {
	// We're not implementing this here as it's already in the main jwt.go file
	// This is just a placeholder for the method call in ExtractUserID
	// The actual implementation uses jwt.Parser.ParseUnverified
	
	// Simulate calling the existing code
	claims, err := Parser{}.ParseWithoutValidation(tokenString)
	if err != nil {
		return 0, err
	}
	
	return claims.UserID, nil
}

// ExtractWithDirectDecoding extracts user ID by directly decoding the base64 payload
func (p *RobustParser) ExtractWithDirectDecoding(tokenString string) (int, error) {
	// Split the token into its parts (header, payload, signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload from base64
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("error decoding payload: %v", err)
	}

	// Parse the JSON payload
	var claims struct {
		UserID int `json:"user_id"`
	}
	
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("error parsing claims: %v", err)
	}

	if claims.UserID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", claims.UserID)
	}

	return claims.UserID, nil
}

// ExtractWithJSONScanning extracts user ID by scanning for the user_id field in JSON
// This is a last-resort method that doesn't rely on proper base64 or JSON structure
func (p *RobustParser) ExtractWithJSONScanning(tokenString string) (int, error) {
	// Split the token into its parts (header, payload, signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid token format: insufficient parts")
	}

	// Add padding if needed
	padding := ""
	switch len(parts[1]) % 4 {
	case 2:
		padding = "=="
	case 3:
		padding = "="
	}
	
	// Try different encoding variants
	var decodedBytes []byte
	var err error
	
	// Try standard base64 with URL encoding
	decodedBytes, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		p.logDebug("Failed standard base64 URL decoding: %v", err)
		
		// Try with padding
		decodedBytes, err = base64.URLEncoding.DecodeString(parts[1] + padding)
		if err != nil {
			p.logDebug("Failed padded base64 URL decoding: %v", err)
			
			// Try standard base64
			decodedBytes, err = base64.RawStdEncoding.DecodeString(parts[1])
			if err != nil {
				p.logDebug("Failed standard base64 decoding: %v", err)
				
				// Try standard base64 with padding
				decodedBytes, err = base64.StdEncoding.DecodeString(parts[1] + padding)
				if err != nil {
					return 0, fmt.Errorf("all base64 decoding methods failed: %v", err)
				}
			}
		}
	}

	// Look for "user_id" in the payload
	payload := string(decodedBytes)
	userIDIndex := strings.Index(payload, "\"user_id\"")
	if userIDIndex == -1 {
		return 0, fmt.Errorf("user_id field not found in payload")
	}

	// Extract the value after "user_id"
	valueStart := userIDIndex + 10 // Length of "user_id" + 2 for the ", or ":
	
	// Find the first digit
	digitStart := -1
	for i := valueStart; i < len(payload); i++ {
		if payload[i] >= '0' && payload[i] <= '9' {
			digitStart = i
			break
		}
	}
	
	if digitStart == -1 {
		return 0, fmt.Errorf("couldn't find digit after user_id field")
	}
	
	// Find the end of the number
	digitEnd := digitStart
	for i := digitStart; i < len(payload); i++ {
		if payload[i] < '0' || payload[i] > '9' {
			digitEnd = i
			break
		}
		if i == len(payload)-1 {
			digitEnd = len(payload)
			break
		}
	}
	
	// Extract and parse the user ID
	userIDStr := payload[digitStart:digitEnd]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing user ID (%s) from token: %v", userIDStr, err)
	}
	
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", userID)
	}
	
	return userID, nil
}