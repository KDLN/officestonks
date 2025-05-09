package auth

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	// In production, this should be set as an environment variable
	jwtSecret = []byte(getEnv("JWT_SECRET", "your-secret-key-for-development-only"))
)

// Claims represents the JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int) (string, error) {
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
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})
	
	// Handle parsing errors
	if err != nil {
		return nil, err
	}
	
	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}