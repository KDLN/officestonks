package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	supabaseJWKSURL string
	supabaseProjectRef string
	publicKeys map[string]*rsa.PublicKey
	lastKeyFetch time.Time
)

func init() {
	supabaseProjectRef = os.Getenv("SUPABASE_PROJECT_REF")
	if supabaseProjectRef != "" {
		supabaseJWKSURL = fmt.Sprintf("https://%s.supabase.co/.well-known/jwks", supabaseProjectRef)
		log.Printf("Supabase JWT validation enabled for project: %s", supabaseProjectRef)
	} else {
		log.Println("SUPABASE_PROJECT_REF not set - Supabase JWT validation disabled")
	}
	
	publicKeys = make(map[string]*rsa.PublicKey)
}

// SupabaseClaims represents Supabase JWT claims
type SupabaseClaims struct {
	Sub   string                 `json:"sub"`
	Email string                 `json:"email"`
	Role  string                 `json:"role"`
	App   map[string]interface{} `json:"app_metadata"`
	User  map[string]interface{} `json:"user_metadata"`
	jwt.StandardClaims
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// fetchSupabasePublicKeys fetches and caches Supabase public keys
func fetchSupabasePublicKeys() error {
	if supabaseJWKSURL == "" {
		return errors.New("Supabase JWKS URL not configured")
	}

	// Cache keys for 1 hour
	if time.Since(lastKeyFetch) < time.Hour && len(publicKeys) > 0 {
		return nil
	}

	resp, err := http.Get(supabaseJWKSURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %v", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			pubKey, err := jwkToPublicKey(key)
			if err != nil {
				log.Printf("Error converting JWK to public key: %v", err)
				continue
			}
			newKeys[key.Kid] = pubKey
		}
	}

	if len(newKeys) == 0 {
		return errors.New("no valid RSA signing keys found in JWKS")
	}

	publicKeys = newKeys
	lastKeyFetch = time.Now()
	log.Printf("Fetched %d Supabase public keys", len(publicKeys))
	
	return nil
}

// jwkToPublicKey converts a JWK to an RSA public key
func jwkToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %v", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %v", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	
	var e int
	if len(eBytes) == 3 {
		e = int(eBytes[0])<<16 + int(eBytes[1])<<8 + int(eBytes[2])
	} else if len(eBytes) == 4 {
		e = int(eBytes[0])<<24 + int(eBytes[1])<<16 + int(eBytes[2])<<8 + int(eBytes[3])
	} else {
		e = int(new(big.Int).SetBytes(eBytes).Int64())
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// ValidateSupabaseToken validates a Supabase JWT token
func ValidateSupabaseToken(tokenString string) (*SupabaseClaims, error) {
	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Parse token to get the key ID
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure we're using RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		// Fetch public keys if needed
		if err := fetchSupabasePublicKeys(); err != nil {
			return nil, fmt.Errorf("failed to fetch public keys: %v", err)
		}

		// Get the public key for this kid
		publicKey, exists := publicKeys[kid]
		if !exists {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		return publicKey, nil
	})

	if err != nil {
		log.Printf("Supabase JWT validation error: %v", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid Supabase token")
	}

	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	// Validate required fields
	if claims.Sub == "" {
		return nil, errors.New("missing sub claim")
	}

	// Check if token is expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// IsSupabaseEnabled returns true if Supabase JWT validation is configured
func IsSupabaseEnabled() bool {
	return supabaseProjectRef != ""
}

// ConvertSupabaseToCustomClaims converts Supabase claims to custom claims format
func ConvertSupabaseToCustomClaims(supabaseClaims *SupabaseClaims, userID int) *Claims {
	return &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: supabaseClaims.ExpiresAt,
			IssuedAt:  supabaseClaims.IssuedAt,
			Subject:   supabaseClaims.Sub,
			Issuer:    "supabase",
		},
	}
}