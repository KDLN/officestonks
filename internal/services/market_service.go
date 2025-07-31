package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"officestonks/internal/models"
	"officestonks/pkg/market"
)

// MarketService handles stock market operations
type MarketService struct {
	stockRepo           models.StockRepository
	userRepo            models.UserRepository
	portfolioRepo       models.PortfolioRepository
	transactionRepo     models.TransactionRepository
	sectorRepo          models.SectorRepository
	delistedStockRepo   models.DelistedStockRepository
	portfolioLossRepo   models.PortfolioLossRepository
	newsService         *NewsService
	simulator           *market.MarketSimulator
}

// NewMarketService creates a new market service
func NewMarketService(
	stockRepo models.StockRepository,
	userRepo models.UserRepository,
	portfolioRepo models.PortfolioRepository,
	transactionRepo models.TransactionRepository,
	sectorRepo models.SectorRepository,
	delistedStockRepo models.DelistedStockRepository,
	portfolioLossRepo models.PortfolioLossRepository,
	newsService *NewsService,
) *MarketService {
	// Create a market simulator with faster updates and higher volatility for more dynamic price movements
	// 2-second updates and 5% volatility
	simulator := market.NewMarketSimulator(2*time.Second, 0.05)
	
	// Create the service first
	service := &MarketService{
		stockRepo:           stockRepo,
		userRepo:            userRepo,
		portfolioRepo:       portfolioRepo,
		transactionRepo:     transactionRepo,
		sectorRepo:          sectorRepo,
		delistedStockRepo:   delistedStockRepo,
		portfolioLossRepo:   portfolioLossRepo,
		newsService:         newsService,
		simulator:           simulator,
	}
	
	// Connect the news service to the simulator for automated news generation
	if newsService != nil {
		simulator.SetNewsService(newsService)
	}
	
	// Connect the bankruptcy handler (the service itself implements the interface)
	simulator.SetBankruptcyHandler(service)
	
	return service
}

// InitializeSimulator loads stocks and starts the simulation
func (s *MarketService) InitializeSimulator() error {
	// Load all sectors from the database
	sectors, err := s.sectorRepo.GetAllSectors()
	if err != nil {
		return fmt.Errorf("failed to load sectors: %w", err)
	}
	
	// Add sectors to the simulator
	for _, sector := range sectors {
		s.simulator.AddSector(sector.ID, sector.Name, sector.VolatilityModifier)
	}
	
	// Load all stocks from the database
	stocks, err := s.stockRepo.LoadStocksForSimulation()
	if err != nil {
		return fmt.Errorf("failed to load stocks: %w", err)
	}
	
	// Add stocks to the simulator
	for id, stock := range stocks {
		s.simulator.AddStock(id, stock.Symbol, stock.Name, stock.Sector, stock.SectorID, stock.Price)
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
		// Validate price before updating database
		if math.IsInf(update.Price, 0) || math.IsNaN(update.Price) || update.Price <= 0 {
			log.Printf("Skipping database update for stock %s: invalid price %f", update.Symbol, update.Price)
			continue
		}
		
		// Ensure price is within reasonable bounds
		price := update.Price
		if price < 0.01 {
			price = 0.01
		} else if price > 1000000 {
			price = 1000000
		}
		
		// Update the stock price in the database
		if err := s.stockRepo.UpdateStockPrice(update.StockID, price); err != nil {
			log.Printf("Error updating stock price for %s: %v", update.Symbol, err)
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

// ValidateSimulator checks and fixes any corrupted data in the market simulator
func (s *MarketService) ValidateSimulator() {
	log.Println("Validating and fixing market simulator data...")
	s.simulator.ValidateAllStocks()
	log.Println("Market simulator validation completed")
}

// PauseSimulator pauses the market simulation
func (s *MarketService) PauseSimulator() {
	s.simulator.Pause()
}

// ResumeSimulator resumes the market simulation
func (s *MarketService) ResumeSimulator() {
	s.simulator.Resume()
}

// ReloadSimulatorPrices reloads all stock prices from database into the simulator
func (s *MarketService) ReloadSimulatorPrices() error {
	// Load current stock prices from database
	stocks, err := s.stockRepo.LoadStocksForSimulation()
	if err != nil {
		return err
	}
	
	// Update each stock's price in the simulator
	for id, stock := range stocks {
		s.simulator.ReloadStock(id, stock.Price)
	}
	
	return nil
}

// AtomicResetStockPrices performs an atomic reset of all stock prices
func (s *MarketService) AtomicResetStockPrices() error {
	// Step 1: Pause the simulator to prevent race conditions
	s.simulator.Pause()
	defer func() {
		// Always resume the simulator, even if there was an error
		s.simulator.Resume()
	}()
	
	// Small delay to ensure pause takes effect
	time.Sleep(100 * time.Millisecond)
	
	// Step 2: Reset prices in database
	err := s.stockRepo.ResetAllStockPrices()
	if err != nil {
		return fmt.Errorf("failed to reset stock prices in database: %w", err)
	}
	
	// Step 3: Reload simulator with new prices from database
	err = s.ReloadSimulatorPrices()
	if err != nil {
		return fmt.Errorf("failed to reload simulator prices: %w", err)
	}
	
	return nil
}

// Testing methods for crisis events

// ForceCrisisEvent forces a specific stock into crisis for testing
func (s *MarketService) ForceCrisisEvent(stockID int) error {
	return s.simulator.ForceCrisisEvent(stockID)
}

// ForceBankruptcy forces a specific stock into bankruptcy for testing
func (s *MarketService) ForceBankruptcy(stockID int) error {
	return s.simulator.ForceBankruptcy(stockID)
}

// ForceRecovery forces a specific stock recovery for testing
func (s *MarketService) ForceRecovery(stockID int) error {
	return s.simulator.ForceRecovery(stockID)
}

// GetSimulatorStockInfo returns current stock info from simulator for testing
func (s *MarketService) GetSimulatorStockInfo(stockID int) (map[string]interface{}, error) {
	stock, exists := s.simulator.GetStockInfo(stockID)
	if !exists {
		return nil, fmt.Errorf("stock with ID %d not found in simulator", stockID)
	}
	
	return map[string]interface{}{
		"id":         stock.ID,
		"symbol":     stock.Symbol,
		"name":       stock.Name,
		"sector":     stock.Sector,
		"sector_id":  stock.SectorID,
		"base_price": stock.BasePrice,
		"trend":      stock.Trend,
		"trend_counter": stock.TrendCounter,
	}, nil
}

// ListSimulatorStocks returns all stocks in the simulator for testing
func (s *MarketService) ListSimulatorStocks() map[int]map[string]interface{} {
	stocks := s.simulator.ListAllStocks()
	result := make(map[int]map[string]interface{})
	
	for id, stock := range stocks {
		result[id] = map[string]interface{}{
			"id":         stock.ID,
			"symbol":     stock.Symbol,
			"name":       stock.Name,
			"sector":     stock.Sector,
			"sector_id":  stock.SectorID,
			"base_price": stock.BasePrice,
			"trend":      stock.Trend,
			"trend_counter": stock.TrendCounter,
		}
	}
	
	return result
}

// ProcessStockBankruptcy handles the complete bankruptcy process for a stock
func (s *MarketService) ProcessStockBankruptcy(stockID int) error {
	// Get the stock information
	stock, err := s.stockRepo.GetStockByID(stockID)
	if err != nil {
		return fmt.Errorf("failed to get stock info: %w", err)
	}
	
	log.Printf("💀 Processing bankruptcy for %s (%s)", stock.Symbol, stock.Name)
	
	// Step 1: Get all users who own this stock
	holders, err := s.portfolioRepo.GetAllHoldersOfStock(stockID)
	if err != nil {
		return fmt.Errorf("failed to get stock holders: %w", err)
	}
	
	log.Printf("Found %d users holding %s stock", len(holders), stock.Symbol)
	
	// Step 2: Record portfolio losses for all holders
	for _, holding := range holders {
		lossAmount := float64(holding.Quantity) * stock.CurrentPrice
		err := s.portfolioLossRepo.CreatePortfolioLoss(
			holding.UserID,
			stockID,
			stock.Symbol,
			stock.Name,
			holding.Quantity,
			lossAmount,
		)
		if err != nil {
			log.Printf("Error recording portfolio loss for user %d: %v", holding.UserID, err)
			// Continue processing other users even if one fails
		} else {
			log.Printf("💸 Recorded $%.2f loss for user %d from %s bankruptcy", lossAmount, holding.UserID, stock.Symbol)
		}
	}
	
	// Step 3: Record the delisted stock
	err = s.delistedStockRepo.CreateDelistedStock(
		stockID,
		stock.Symbol,
		stock.Name,
		stock.Sector,
		stock.CurrentPrice,
		models.DelistingBankruptcy,
	)
	if err != nil {
		return fmt.Errorf("failed to record delisted stock: %w", err)
	}
	
	// Step 4: Remove stock from all portfolios
	err = s.portfolioRepo.RemoveStockFromAllPortfolios(stockID)
	if err != nil {
		return fmt.Errorf("failed to remove stock from portfolios: %w", err)
	}
	
	// Step 5: Mark stock as delisted in database
	err = s.stockRepo.DelistStock(stockID)
	if err != nil {
		return fmt.Errorf("failed to delist stock: %w", err)
	}
	
	log.Printf("✅ Bankruptcy processing completed for %s", stock.Symbol)
	return nil
}