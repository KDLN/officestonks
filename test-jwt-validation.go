package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// Print colored text
func printColored(color, text string) {
	colorCodes := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"reset":  "\033[0m",
	}
	
	fmt.Print(colorCodes[color] + text + colorCodes["reset"])
}

// Generate a token using the specified secret
func generateToken(userID int, secret string) (string, error) {
	// Set token expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims with user ID and expiration time
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// Create the token using the claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Parse token without validation
func parseWithoutValidation(tokenString string) (*Claims, error) {
	// Split the token into its parts (header, payload, signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload from base64
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %v", err)
	}

	// Parse the JSON payload
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("error parsing claims: %v", err)
	}

	return &claims, nil
}

// Test strict token validation
func testValidateStrict(tokenString string, secretsToTry []string) (*Claims, error) {
	var lastError error

	for i, secret := range secretsToTry {
		claims := &Claims{} // Reset claims for each attempt
		
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", token.Method.Alg())
			}
			return []byte(secret), nil
		})

		// If this secret worked, return claims
		if err == nil && token.Valid {
			printColored("green", fmt.Sprintf("🔑 Token valid using secret #%d [%s] for user ID: %d\n", 
				i+1, secret, claims.UserID))
			return claims, nil
		}

		// This secret didn't work, save the error
		lastError = err
		printColored("yellow", fmt.Sprintf("⚠️ Secret #%d [%s] failed: %v\n", i+1, secret, err))
	}

	// None of the secrets worked
	printColored("red", "❌ All secrets failed to validate token\n")
	return nil, lastError
}

// Run JWT bypass test with the existing token
func testBypassMode(tokenString string) (*Claims, error) {
	printColored("blue", "\n--- TESTING BYPASS MODE ---\n")

	parser := jwt.Parser{
		SkipClaimsValidation: true,
	}
	claims := &Claims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err == nil && claims.UserID > 0 {
		printColored("green", fmt.Sprintf("✅ BYPASS MODE: Successfully extracted user ID without validation: %d\n", claims.UserID))
		return claims, nil
	} else if err != nil {
		printColored("red", fmt.Sprintf("❌ BYPASS MODE: Error parsing token: %v\n", err))
		return nil, err
	} else {
		printColored("red", fmt.Sprintf("❌ BYPASS MODE: Invalid user ID in token: %d\n", claims.UserID))
		return nil, fmt.Errorf("invalid user ID")
	}
}

// Run manual claims extraction test
func testManualExtraction(tokenString string) (*Claims, error) {
	printColored("blue", "\n--- TESTING MANUAL EXTRACTION ---\n")
	
	claims, err := parseWithoutValidation(tokenString)
	if err != nil {
		printColored("red", fmt.Sprintf("❌ MANUAL EXTRACTION: Error: %v\n", err))
		return nil, err
	}
	
	if claims.UserID > 0 {
		printColored("green", fmt.Sprintf("✅ MANUAL EXTRACTION: Successfully extracted user ID: %d\n", claims.UserID))
		return claims, nil
	} else {
		printColored("red", fmt.Sprintf("❌ MANUAL EXTRACTION: Invalid user ID: %d\n", claims.UserID))
		return nil, fmt.Errorf("invalid user ID")
	}
}

func main() {
	// Define the user ID to use in tokens
	userID := 3  // This is the KDLN user ID

	// Define secrets to try
	secrets := []string{
		"your-secret-key-for-development-only",
		"OfficeStonksSecret",
		"stonkstoken",
		"YourJwtSecretKey", 
		"yourjwtsecretkey",
		"your-jwt-secret-key",
		"admin123",
	}

	// Create section divider
	printColored("blue", "=============================\n")
	printColored("blue", "JWT TOKEN VALIDATION TEST\n")
	printColored("blue", "=============================\n\n")

	// Generate our own token with each secret
	printColored("blue", "--- TESTING TOKEN GENERATION ---\n")
	var savedToken string
	
	for i, secret := range secrets {
		token, err := generateToken(userID, secret)
		if err != nil {
			printColored("red", fmt.Sprintf("❌ Failed to generate token with secret #%d: %v\n", i+1, err))
			continue
		}
		
		tokenPreview := token
		if len(token) > 20 {
			tokenPreview = token[:20] + "..."
		}
		
		printColored("green", fmt.Sprintf("✅ Generated token #%d with secret [%s]: %s\n", 
			i+1, secret, tokenPreview))
		
		// Save the first token for later use
		if i == 0 {
			savedToken = token
		}
	}

	// Test validation with strict mode
	printColored("blue", "\n--- TESTING STRICT VALIDATION ---\n")
	_, err := testValidateStrict(savedToken, secrets)
	if err != nil {
		printColored("yellow", fmt.Sprintf("⚠️ Strict validation failed: %v\n", err))
	}

	// Test bypass mode
	testBypassMode(savedToken)
	
	// Test manual extraction
	testManualExtraction(savedToken)

	// Try from environment if available
	if len(os.Args) > 1 {
		externalToken := os.Args[1]
		printColored("blue", "\n=============================\n")
		printColored("blue", "TESTING PROVIDED TOKEN\n")
		printColored("blue", "=============================\n")
		
		tokenPreview := externalToken
		if len(externalToken) > 20 {
			tokenPreview = externalToken[:20] + "..."
		}
		printColored("blue", fmt.Sprintf("Token: %s\n", tokenPreview))
		
		// Test strict validation
		printColored("blue", "\n--- TESTING STRICT VALIDATION ---\n")
		_, err := testValidateStrict(externalToken, secrets)
		if err != nil {
			printColored("yellow", fmt.Sprintf("⚠️ Strict validation failed: %v\n", err))
		}

		// Test bypass mode
		testBypassMode(externalToken)
		
		// Test manual extraction
		testManualExtraction(externalToken)
	} else {
		log.Println("Tip: You can provide an external token to test by passing it as an argument")
		log.Println("Example: go run test-jwt-validation.go <your-token>")
	}
}