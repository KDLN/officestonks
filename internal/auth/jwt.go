package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	// JWT secret loaded from environment variable
	jwtSecret []byte
)

// Initialize JWT secret from environment variable
func init() {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		// For development only - use a default but warn about it
		log.Println("WARNING: JWT_SECRET environment variable not set. Using default for development.")
		log.Println("SECURITY: Set JWT_SECRET environment variable in production!")
		secretKey = "development-jwt-secret-please-change-in-production"
	}
	jwtSecret = []byte(secretKey)
	
	if len(jwtSecret) < 32 {
		log.Printf("WARNING: JWT secret is only %d characters. Recommend at least 32 characters for security.", len(jwtSecret))
	}
}

// Claims represents the JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// TokenValidator interface for validating tokens and returning user IDs
type TokenValidator interface {
	ValidateToken(tokenString string) (int, error)
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int) (string, error) {
	if userID <= 0 {
		return "", errors.New("invalid user ID")
	}

	// Set token expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims with user ID and expiration time
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "office-stonks",
		},
	}

	// Create the token using the claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error signing token for user %d: %v", userID, err)
		return "", err
	}

	log.Printf("Generated JWT token for user %d", userID)
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	// Parse and validate the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		log.Printf("JWT validation error: %v", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Validate user ID
	if claims.UserID <= 0 {
		return nil, errors.New("invalid user ID in token")
	}

	// Check if token is expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken creates a new token with extended expiration for a valid token
func RefreshToken(tokenString string) (string, error) {
	// Validate the current token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate a new token with the same user ID
	return GenerateToken(claims.UserID)
}

// GetJWTSecretInfo returns information about the JWT secret (for debugging)
func GetJWTSecretInfo() map[string]interface{} {
	return map[string]interface{}{
		"length":     len(jwtSecret),
		"is_default": string(jwtSecret) == "development-jwt-secret-please-change-in-production",
		"from_env":   os.Getenv("JWT_SECRET") != "",
	}
}