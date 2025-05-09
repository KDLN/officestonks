package services

import (
	"errors"
	"time"

	"officestonks/internal/models"
	"officestonks/pkg/market"
)

// MarketService handles stock market operations
type MarketService struct {
	stockRepo      models.StockRepository
	userRepo       models.UserRepository
	portfolioRepo  models.PortfolioRepository
	transactionRepo models.TransactionRepository
	simulator      *market.MarketSimulator
}

// NewMarketService creates a new market service
func NewMarketService(
	stockRepo models.StockRepository,
	userRepo models.UserRepository,
	portfolioRepo models.PortfolioRepository,
	transactionRepo models.TransactionRepository,
) *MarketService {
	// Create a new market simulator with 5-second updates and 2% volatility
	simulator := market.NewMarketSimulator(5*time.Second, 0.02)
	
	// Return the service
	return &MarketService{
		stockRepo:      stockRepo,
		userRepo:       userRepo,
		portfolioRepo:  portfolioRepo,
		transactionRepo: transactionRepo,
		simulator:      simulator,
	}
}

// InitializeSimulator loads stocks and starts the simulation
func (s *MarketService) InitializeSimulator() error {
	// Load all stocks from the database
	stocks, err := s.stockRepo.LoadStocksForSimulation()
	if err != nil {
		return err
	}
	
	// Add stocks to the simulator
	for id, stock := range stocks {
		s.simulator.AddStock(id, stock.Symbol, stock.Sector, stock.Price)
	}
	
	// Start the simulator
	s.simulator.Start()
	
	// Start a goroutine to update stock prices in the database
	go s.updateStockPrices()
	
	return nil
}

// updateStockPrices handles updates from the simulator
func (s *MarketService) updateStockPrices() {
	updateChan := s.simulator.GetUpdateChannel()
	
	for update := range updateChan {
		// Update the stock price in the database
		if err := s.stockRepo.UpdateStockPrice(update.StockID, update.Price); err != nil {
			// Log the error but continue processing updates
			// In a real application, you'd want better error handling
			continue
		}
	}
}

// GetAllStocks returns all available stocks
func (s *MarketService) GetAllStocks() ([]*models.Stock, error) {
	return s.stockRepo.GetAllStocks()
}

// GetStockByID returns a stock by ID
func (s *MarketService) GetStockByID(id int) (*models.Stock, error) {
	return s.stockRepo.GetStockByID(id)
}

// GetUserPortfolio returns a user's portfolio
func (s *MarketService) GetUserPortfolio(userID int) (*models.PortfolioSummary, error) {
	// Get the user's portfolio items
	items, err := s.portfolioRepo.GetUserPortfolio(userID)
	if err != nil {
		return nil, err
	}
	
	// Get the user's cash balance
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	// Calculate total stock value
	var stockValue float64
	for _, item := range items {
		stockValue += float64(item.Quantity) * item.Stock.CurrentPrice
	}
	
	// Create the portfolio summary
	summary := &models.PortfolioSummary{
		CashBalance:    user.CashBalance,
		StockValue:     stockValue,
		TotalValue:     user.CashBalance + stockValue,
		PortfolioItems: items,
	}
	
	return summary, nil
}

// BuyStock handles a stock purchase
func (s *MarketService) BuyStock(userID, stockID, quantity int) error {
	// Input validation
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	
	// Get the stock
	stock, err := s.stockRepo.GetStockByID(stockID)
	if err != nil {
		return err
	}
	
	// Get the user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}
	
	// Calculate total cost
	totalCost := stock.CurrentPrice * float64(quantity)
	
	// Check if user has enough cash
	if user.CashBalance < totalCost {
		return errors.New("insufficient funds")
	}
	
	// Begin transaction (in a real app, you'd use a database transaction here)
	
	// Update user's cash balance
	newBalance := user.CashBalance - totalCost
	if err := s.userRepo.UpdateUserBalance(userID, newBalance); err != nil {
		return err
	}
	
	// Update user's portfolio
	if err := s.portfolioRepo.AddStockToPortfolio(userID, stockID, quantity); err != nil {
		// In a real app, you'd roll back the balance change on error
		return err
	}
	
	// Record the transaction
	_, err = s.transactionRepo.CreateTransaction(userID, stockID, quantity, stock.CurrentPrice, models.Buy)
	if err != nil {
		// In a real app, you'd roll back the previous changes on error
		return err
	}
	
	// Update market simulation
	s.simulator.ProcessTransaction(stockID, quantity, true)
	
	return nil
}

// SellStock handles a stock sale
func (s *MarketService) SellStock(userID, stockID, quantity int) error {
	// Input validation
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	
	// Get the stock
	stock, err := s.stockRepo.GetStockByID(stockID)
	if err != nil {
		return err
	}
	
	// Get the user's holding for this stock
	holding, err := s.portfolioRepo.GetUserStockHolding(userID, stockID)
	if err != nil {
		return err
	}
	
	// Check if user owns the stock and has enough shares
	if holding == nil || holding.Quantity < quantity {
		return errors.New("insufficient shares")
	}
	
	// Get the user
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}
	
	// Calculate total proceeds
	totalProceeds := stock.CurrentPrice * float64(quantity)
	
	// Begin transaction (in a real app, you'd use a database transaction here)
	
	// Update user's cash balance
	newBalance := user.CashBalance + totalProceeds
	if err := s.userRepo.UpdateUserBalance(userID, newBalance); err != nil {
		return err
	}
	
	// Update user's portfolio
	newQuantity := holding.Quantity - quantity
	if err := s.portfolioRepo.UpdateStockQuantity(holding.ID, newQuantity); err != nil {
		// In a real app, you'd roll back the balance change on error
		return err
	}
	
	// Record the transaction
	_, err = s.transactionRepo.CreateTransaction(userID, stockID, quantity, stock.CurrentPrice, models.Sell)
	if err != nil {
		// In a real app, you'd roll back the previous changes on error
		return err
	}
	
	// Update market simulation
	s.simulator.ProcessTransaction(stockID, quantity, false)
	
	return nil
}

// GetUserTransactions returns a user's transaction history
func (s *MarketService) GetUserTransactions(userID, limit, offset int) ([]*models.Transaction, error) {
	return s.transactionRepo.GetUserTransactions(userID, limit, offset)
}

// GetSimulatorUpdates returns the channel for stock price updates
func (s *MarketService) GetSimulatorUpdates() <-chan market.StockUpdate {
	return s.simulator.GetUpdateChannel()
}