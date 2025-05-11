package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// Parser provides a simple way to parse JWT tokens without validation
type Parser struct{}

// ParseWithoutValidation parses a JWT token without any signature validation
func (p *Parser) ParseWithoutValidation(tokenString string) (*Claims, error) {
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