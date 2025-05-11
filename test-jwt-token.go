package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Claims struct {
	UserID int `json:"user_id"`
	Exp    int64 `json:"exp"`
	Iat    int64 `json:"iat"`
}

func parseJWTWithoutValidation(tokenString string) (*Claims, error) {
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

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <jwt-token>", os.Args[0])
	}

	token := os.Args[1]
	
	// Print token preview
	tokenPreview := token
	if len(token) > 20 {
		tokenPreview = token[:20] + "..."
	}
	fmt.Printf("Token preview: %s\n", tokenPreview)

	// Parse the token without validation
	claims, err := parseJWTWithoutValidation(token)
	if err != nil {
		log.Fatalf("Error parsing token: %v", err)
	}

	fmt.Printf("Successfully parsed token without validation!\n")
	fmt.Printf("User ID: %d\n", claims.UserID)
	fmt.Printf("Issued at: %d\n", claims.Iat)
	fmt.Printf("Expires at: %d\n", claims.Exp)
}