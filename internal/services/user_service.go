package services

import (
	"officestonks/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo      models.UserRepository
	portfolioRepo models.PortfolioRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo models.UserRepository, portfolioRepo models.PortfolioRepository) *UserService {
	return &UserService{
		userRepo:      userRepo,
		portfolioRepo: portfolioRepo,
	}
}

// LeaderboardEntry represents a user in the leaderboard
type LeaderboardEntry struct {
	UserID      int     `json:"user_id"`
	Username    string  `json:"username"`
	CashBalance float64 `json:"cash_balance"`
	StockValue  float64 `json:"stock_value"`
	TotalValue  float64 `json:"total_value"`
	Rank        int     `json:"rank"`
}

// GetLeaderboard returns the top users by portfolio value
func (s *UserService) GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	// Get all users
	users, err := s.userRepo.GetTopUsers(limit)
	if err != nil {
		return nil, err
	}
	
	// Create leaderboard entries with portfolio values
	leaderboard := make([]LeaderboardEntry, 0, len(users))
	
	for i, user := range users {
		// Calculate portfolio value
		portfolioValue, err := s.portfolioRepo.CalculatePortfolioValue(user.ID)
		if err != nil {
			return nil, err
		}
		
		// Get stock value (portfolio value - cash balance)
		stockValue := portfolioValue - user.CashBalance
		
		// Create leaderboard entry
		entry := LeaderboardEntry{
			UserID:      user.ID,
			Username:    user.Username,
			CashBalance: user.CashBalance,
			StockValue:  stockValue,
			TotalValue:  portfolioValue,
			Rank:        i + 1, // 1-based ranking
		}
		
		leaderboard = append(leaderboard, entry)
	}
	
	return leaderboard, nil
}

// UserProfile represents a user's profile information
type UserProfile struct {
	UserID      int     `json:"user_id"`
	Username    string  `json:"username"`
	CashBalance float64 `json:"cash_balance"`
	StockValue  float64 `json:"stock_value"`
	TotalValue  float64 `json:"total_value"`
	JoinedDate  string  `json:"joined_date"`
}

// GetUserProfile returns the profile for a specific user
func (s *UserService) GetUserProfile(userID int) (*UserProfile, error) {
	// Get the user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	// Calculate portfolio value
	portfolioValue, err := s.portfolioRepo.CalculatePortfolioValue(userID)
	if err != nil {
		return nil, err
	}
	
	// Get stock value (portfolio value - cash balance)
	stockValue := portfolioValue - user.CashBalance
	
	// Create user profile
	profile := &UserProfile{
		UserID:      user.ID,
		Username:    user.Username,
		CashBalance: user.CashBalance,
		StockValue:  stockValue,
		TotalValue:  portfolioValue,
		JoinedDate:  user.CreatedAt.Format("January 2, 2006"),
	}
	
	return profile, nil
}