package services

import (
	"errors"
	"log"
	"strconv"
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
	// Try Supabase JWT first if enabled
	if auth.IsSupabaseEnabled() {
		supabaseClaims, err := auth.ValidateSupabaseToken(tokenString)
		if err == nil {
			// Get or create user based on Supabase claims
			userID, err := s.getOrCreateSupabaseUser(supabaseClaims)
			if err != nil {
				log.Printf("Error getting/creating Supabase user: %v", err)
				return 0, err
			}
			return userID, nil
		}
		log.Printf("Supabase token validation failed, trying custom JWT: %v", err)
	}

	// Fallback to custom JWT validation
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	// Check if the user exists
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return 0, errors.New("invalid token: user not found")
	}

	return user.ID, nil
}

// getOrCreateSupabaseUser gets or creates a user based on Supabase claims
func (s *AuthService) getOrCreateSupabaseUser(claims *auth.SupabaseClaims) (int, error) {
	// First try to find user by email or Supabase ID
	// For now, we'll use email as the primary identifier
	email := claims.Email
	if email == "" {
		return 0, errors.New("no email in Supabase token")
	}

	// Try to find user by username (we'll use email as username for Supabase users)
	user, err := s.userRepo.GetUserByUsername(email)
	if err == nil {
		// User exists, return their ID
		return user.ID, nil
	}

	// User doesn't exist, create them
	// Generate a random password hash since Supabase handles auth
	hashedPassword, err := auth.HashPassword("supabase_managed_" + claims.Sub)
	if err != nil {
		return 0, err
	}

	// Create user with email as username
	user, err = s.userRepo.CreateUser(email, hashedPassword)
	if err != nil {
		return 0, err
	}

	log.Printf("Created new user from Supabase: ID=%d, Email=%s", user.ID, email)
	return user.ID, nil
}