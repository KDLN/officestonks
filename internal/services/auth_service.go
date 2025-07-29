package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"officestonks/internal/auth"
	"officestonks/internal/models"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo models.UserRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo models.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register creates a new user account
func (s *AuthService) Register(username, password string) (*models.AuthResponse, error) {
	// Check if username already exists
	_, err := s.userRepo.GetUserByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	
	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	
	// Create the user
	user, err := s.userRepo.CreateUser(username, hashedPassword)
	if err != nil {
		return nil, err
	}
	
	// Generate a JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	
	// Return the auth response
	return &models.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(username, password string) (*models.AuthResponse, error) {
	// Get the user by username
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}
	
	// Verify the password
	valid, err := auth.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, errors.New("invalid username or password")
	}
	
	// Generate a JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}
	
	// Return the auth response
	return &models.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	// Quick check if this looks like a Supabase token (has 3 dots indicating JWK format)
	// Office Stonks tokens are HMAC-signed and won't have a 'kid' header
	isLikelySupabaseToken := strings.Count(tokenString, ".") == 2 && 
		strings.Contains(tokenString, "ey") // JWT tokens start with 'ey'

	if isLikelySupabaseToken && auth.IsSupabaseEnabled() {
		// Try to parse the header to check for 'kid'
		parts := strings.Split(tokenString, ".")
		if len(parts) == 3 {
			// Decode header to check for 'kid'
			headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
			if err == nil {
				var header map[string]interface{}
				if json.Unmarshal(headerBytes, &header) == nil {
					if _, hasKid := header["kid"]; hasKid {
						// This looks like a Supabase token, try it first
						supabaseClaims, err := auth.ValidateSupabaseToken(tokenString)
						if err == nil {
							userID, err := s.GetOrCreateSupabaseUser(supabaseClaims)
							if err != nil {
								log.Printf("Error getting/creating Supabase user: %v", err)
								return 0, err
							}
							return userID, nil
						}
						log.Printf("Supabase token validation failed: %v", err)
					}
				}
			}
		}
	}

	// Try custom JWT validation (Office Stonks tokens)
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		// If custom JWT fails and Supabase is enabled, try Supabase as fallback
		if auth.IsSupabaseEnabled() {
			supabaseClaims, err := auth.ValidateSupabaseToken(tokenString)
			if err == nil {
				userID, err := s.GetOrCreateSupabaseUser(supabaseClaims)
				if err != nil {
					log.Printf("Error getting/creating Supabase user: %v", err)
					return 0, err
				}
				return userID, nil
			}
			// Don't log Supabase errors if it's just not configured properly
			if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "JWKS endpoint") {
				log.Printf("Both custom JWT and Supabase validation failed: %v", err)
			}
		}
		return 0, err
	}

	// Check if the user exists
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return 0, errors.New("invalid token: user not found")
	}

	return user.ID, nil
}

// GetOrCreateSupabaseUser gets or creates a user based on Supabase claims
func (s *AuthService) GetOrCreateSupabaseUser(claims *auth.SupabaseClaims) (int, error) {
	// First try to find user by Supabase ID
	if claims.Sub != "" {
		user, err := s.userRepo.GetUserBySupabaseID(claims.Sub)
		if err == nil {
			// User exists, return their ID
			return user.ID, nil
		}
	}

	// Extract Discord username from user metadata, fallback to email
	username := claims.Email
	if claims.User != nil {
		if discordUsername, ok := claims.User["preferred_username"].(string); ok && discordUsername != "" {
			username = discordUsername
		} else if fullName, ok := claims.User["full_name"].(string); ok && fullName != "" {
			username = fullName
		}
	}

	if username == "" {
		return 0, errors.New("no username or email in Supabase token")
	}

	// Generate a random password hash since Supabase handles auth
	hashedPassword, err := auth.HashPassword("supabase_managed_" + claims.Sub)
	if err != nil {
		return 0, err
	}

	// Create user with Discord username and Supabase ID
	user, err := s.userRepo.CreateUserWithSupabase(username, hashedPassword, claims.Sub)
	if err != nil {
		return 0, err
	}

	log.Printf("Created new user from Supabase: ID=%d, Username=%s, SupabaseID=%s", user.ID, username, claims.Sub)
	return user.ID, nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}