package auth

import (
	"errors"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	// In production, this should be set as an environment variable
	// IMPORTANT: Using a hardcoded secret for debugging - MUST be replaced with proper environment variable
	// in production for security!
	jwtSecret = []byte("your-secret-key-for-development-only")
)

// Claims represents the JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID int) (string, error) {
	// Log token generation
	println("Generating new token for user ID:", userID)
	println("Using JWT secret length:", len(jwtSecret))

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
		println("Error signing token:", err.Error())
		return "", err
	}

	// Log token preview (first part only for security)
	tokenPreview := tokenString
	if len(tokenString) > 20 {
		tokenPreview = tokenString[:20] + "..."
	}
	println("Generated token (preview):", tokenPreview)

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	// Print token debug info (truncated for security)
	tokenPreview := tokenString
	if len(tokenString) > 20 {
		tokenPreview = tokenString[:20] + "..."
	}

	// Log the token we're trying to validate
	println("Validating token:", tokenPreview)
	println("Using JWT secret length:", len(jwtSecret))

	// Parse the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			println("Invalid signing method:", token.Method.Alg())
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	// Handle parsing errors
	if err != nil {
		println("Token validation error:", err.Error())
		return nil, err
	}

	// Validate token
	if !token.Valid {
		println("Token is invalid")
		return nil, errors.New("invalid token")
	}

	// Token is valid, print userID
	println("Token valid for user ID:", claims.UserID)
	return claims, nil
}

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}