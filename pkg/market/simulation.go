package market

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// StockUpdate represents a price update for a stock
type StockUpdate struct {
	StockID int
	Symbol  string
	Price   float64
}

// NewsServiceInterface defines methods for news generation
type NewsServiceInterface interface {
	GenerateCrisisNews(stockID int, stockSymbol, stockName, sectorName string) error
	GenerateBankruptcyNews(stockID int, stockSymbol, stockName, sectorName string) error
	GenerateRecoveryNews(stockID int, stockSymbol, stockName, sectorName string, newPrice float64) error
	GenerateSectorContagionNews(sectorName string, eventType string, affectedCount int) error
}

// BankruptcyHandlerInterface defines methods for processing bankruptcies
type BankruptcyHandlerInterface interface {
	ProcessStockBankruptcy(stockID int) error
}

// MarketSimulator handles the stock price simulation
type MarketSimulator struct {
	stocksInfo        map[int]StockInfo
	sectorsInfo       map[int]SectorInfo
	updateInterval    time.Duration
	volatility        float64
	mu                sync.RWMutex
	updateChan        chan StockUpdate
	stopChan          chan struct{}
	pauseChan         chan bool
	isPaused          bool
	newsService       NewsServiceInterface
	bankruptcyHandler BankruptcyHandlerInterface
}

// StockInfo contains information about a stock for simulation
type StockInfo struct {
	ID           int
	Symbol       string
	Name         string // Company name for news generation
	BasePrice    float64
	Sector       string
	SectorID     int
	Trend        float64 // Bias for price movement: positive means upward trend, negative means downward
	TrendCounter int     // Counter to track trend duration
	LockedUntil  time.Time // Admin lock - prevents automatic updates until this time
	LockedPrice  float64   // The price set by admin during lock period
}

// SectorInfo tracks sector-wide trends and volatility
type SectorInfo struct {
	ID                 int
	Name               string
	Trend              float64 // Sector-wide trend
	VolatilityModifier float64 // Multiplier for sector volatility
	StockCount         int     // Number of stocks in this sector
}

// NewMarketSimulator creates a new market simulator
func NewMarketSimulator(updateInterval time.Duration, volatility float64) *MarketSimulator {
	return &MarketSimulator{
		stocksInfo:        make(map[int]StockInfo),
		sectorsInfo:       make(map[int]SectorInfo),
		updateInterval:    updateInterval,
		volatility:        volatility,
		updateChan:        make(chan StockUpdate, 100),
		stopChan:          make(chan struct{}),
		pauseChan:         make(chan bool, 1),
		isPaused:          false,
		newsService:       nil, // Will be set later
		bankruptcyHandler: nil, // Will be set later
	}
}

// SetNewsService connects a news service for automated news generation
func (s *MarketSimulator) SetNewsService(newsService NewsServiceInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.newsService = newsService
	log.Printf("📰 News service connected to market simulator")
}

// SetBankruptcyHandler connects a bankruptcy handler for portfolio processing
func (s *MarketSimulator) SetBankruptcyHandler(handler BankruptcyHandlerInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bankruptcyHandler = handler
	log.Printf("💀 Bankruptcy handler connected to market simulator")
}

// AddSector adds a sector to the simulator
func (s *MarketSimulator) AddSector(id int, name string, volatilityModifier float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize with a small random trend
	initialTrend := (rand.Float64() * 0.04) - 0.02 // Range: -0.02 to 0.02

	s.sectorsInfo[id] = SectorInfo{
		ID:                 id,
		Name:               name,
		Trend:              initialTrend,
		VolatilityModifier: volatilityModifier,
		StockCount:         0, // Will be updated as stocks are added
	}
}

// AddStock adds a stock to the simulator
func (s *MarketSimulator) AddStock(id int, symbol, name, sector string, sectorID int, basePrice float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate base price before adding
	if math.IsInf(basePrice, 0) || math.IsNaN(basePrice) || basePrice <= 0 {
		log.Printf("⚠️ Invalid base price for stock %s: %f, setting to $10.00", symbol, basePrice)
		basePrice = 10.00
	}

	// Initialize with a random trend (slightly biased upward for a bull market)
	initialTrend := (rand.Float64() * 0.1) - 0.03 // Range: -0.03 to 0.07, slightly positive bias

	s.stocksInfo[id] = StockInfo{
		ID:           id,
		Symbol:       symbol,
		Name:         name,
		BasePrice:    basePrice,
		Sector:       sector,
		SectorID:     sectorID,
		Trend:        initialTrend,
		TrendCounter: rand.Intn(10) + 5, // Random initial trend duration (5-15 updates)
	}

	// Update sector stock count
	if sector, exists := s.sectorsInfo[sectorID]; exists {
		sector.StockCount++
		s.sectorsInfo[sectorID] = sector
	}
}

// UpdateStockPrice updates a stock's base price in the simulator
func (s *MarketSimulator) UpdateStockPrice(id int, newPrice float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if info, exists := s.stocksInfo[id]; exists {
		info.BasePrice = newPrice
		// Lock the price for 30 seconds after admin update to prevent race conditions
		info.LockedUntil = time.Now().Add(30 * time.Second)
		info.LockedPrice = newPrice
		s.stocksInfo[id] = info
		log.Printf("🔒 Stock %s (ID: %d) price locked at $%.2f for 30 seconds", info.Symbol, id, newPrice)
	}
}

// GetUpdateChannel returns the channel for receiving stock updates
func (s *MarketSimulator) GetUpdateChannel() <-chan StockUpdate {
	return s.updateChan
}

// Start begins the market simulation
func (s *MarketSimulator) Start() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	log.Printf("🚀 MarketSimulator: Starting market simulation with %d stocks", len(s.stocksInfo))
	log.Printf("📊 MarketSimulator: Update interval: %v, Volatility: %.2f%%", s.updateInterval, s.volatility*100)

	// Start the simulation loop in a goroutine
	go s.simulationLoop()
}

// Stop halts the market simulation
func (s *MarketSimulator) Stop() {
	close(s.stopChan)
}

// simulationLoop runs the main simulation
func (s *MarketSimulator) simulationLoop() {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	log.Printf("📈 MarketSimulator: Simulation loop started with %v interval", s.updateInterval)
	updateCount := 0
	lastLogTime := time.Now()

	for {
		select {
		case <-ticker.C:
			// Only update prices if not paused
			s.mu.RLock()
			paused := s.isPaused
			stockCount := len(s.stocksInfo)
			s.mu.RUnlock()

			if !paused {
				updateCount++
				s.updatePrices()

				// Log progress every 30 seconds
				if time.Since(lastLogTime) >= 30*time.Second {
					log.Printf("📊 MarketSimulator: Generated %d updates for %d stocks in last 30s", updateCount, stockCount)
					updateCount = 0
					lastLogTime = time.Now()
				}
			} else {
				// Log pause status occasionally
				if updateCount == 0 || updateCount%15 == 0 { // Every 30 seconds when paused
					log.Printf("⏸️ MarketSimulator: Simulation paused, skipping price updates")
				}
			}
		case pauseState := <-s.pauseChan:
			s.mu.Lock()
			s.isPaused = pauseState
			s.mu.Unlock()
			if pauseState {
				log.Printf("⏸️ MarketSimulator: Simulation paused")
			} else {
				log.Printf("▶️ MarketSimulator: Simulation resumed")
			}
		case <-s.stopChan:
			log.Printf("🛑 MarketSimulator: Simulation stopped")
			close(s.updateChan)
			return
		}
	}
}

// updateSectorTrends updates sector-wide trends (called less frequently than stock updates)
func (s *MarketSimulator) updateSectorTrends() {
	for sectorID, sector := range s.sectorsInfo {
		// Sector trends change more slowly than individual stocks (10% chance per update)
		if rand.Float64() < 0.1 {
			adjustment := (rand.Float64() - 0.5) * 0.02 // ±1% change
			sector.Trend += adjustment

			// Cap sector trends to reasonable limits
			if sector.Trend > 0.05 {
				sector.Trend = 0.05
			} else if sector.Trend < -0.05 {
				sector.Trend = -0.05
			}

			s.sectorsInfo[sectorID] = sector
		}
	}
}

// updatePrices calculates new prices for all stocks
func (s *MarketSimulator) updatePrices() {
	s.mu.Lock() // Use write lock since we're updating the stocksInfo
	defer s.mu.Unlock()

	// Update sector trends first
	s.updateSectorTrends()

	// Update all stocks
	for id, info := range s.stocksInfo {
		// Check if stock is locked by admin update
		if !info.LockedUntil.IsZero() && time.Now().Before(info.LockedUntil) {
			// Stock is locked, use the locked price and skip automatic updates
			select {
			case s.updateChan <- StockUpdate{
				StockID: info.ID,
				Symbol:  info.Symbol,
				Price:   info.LockedPrice,
			}:
			default:
				// Channel is full, skip this update
			}
			continue // Skip to next stock
		}
		
		// If lock has expired, clear it
		if !info.LockedUntil.IsZero() && time.Now().After(info.LockedUntil) {
			info.LockedUntil = time.Time{} // Clear the lock
			info.LockedPrice = 0
			log.Printf("🔓 Stock %s (ID: %d) price lock expired, resuming automatic updates", info.Symbol, id)
		}
		
		// Validate base price before any calculations
		if math.IsInf(info.BasePrice, 0) || math.IsNaN(info.BasePrice) || info.BasePrice <= 0 {
			info.BasePrice = 0.01 // Reset to safe value
			s.stocksInfo[id] = info
		}
		// Check if we need to change the trend
		if info.TrendCounter <= 0 {
			// Time to reverse or modify the trend

			// Generate new trend - more likely to reverse direction
			if rand.Float64() < 0.7 { // 70% chance of trend reversal
				// Reverse the trend with a moderate factor to avoid runaway drops
				info.Trend = -info.Trend * (0.5 + rand.Float64()*0.5) // 50-100% reversal
			} else {
				// Modify current trend with dampening (regression to mean)
				dampening := 0.3 + rand.Float64()*0.4 // 30-70% of current trend
				info.Trend = info.Trend * dampening
			}

			// Set new duration for this trend
			info.TrendCounter = rand.Intn(15) + 5 // 5-20 updates
		} else {
			info.TrendCounter--
		}

		// Calculate price zone volatility based on current price
		priceZoneVolatility := s.getPriceZoneVolatility(info.BasePrice)

		// Calculate new price with sector correlation
		// Individual stock factors (70% weight)
		randomChange := (rand.Float64() - 0.5) * priceZoneVolatility
		individualChange := randomChange + info.Trend

		// Validate individual change
		if math.IsInf(individualChange, 0) || math.IsNaN(individualChange) {
			individualChange = randomChange // fallback to just random change
		}

		// Sector factors (30% weight)
		var sectorChange float64
		if info.SectorID > 0 {
			if sector, exists := s.sectorsInfo[info.SectorID]; exists {
				sectorVolatility := s.volatility * sector.VolatilityModifier * 0.3
				if sectorVolatility > 0 && !math.IsInf(sectorVolatility, 0) && !math.IsNaN(sectorVolatility) {
					sectorRandomChange := (rand.Float64() - 0.5) * sectorVolatility
					if !math.IsInf(sector.Trend, 0) && !math.IsNaN(sector.Trend) {
						sectorChange = sectorRandomChange + sector.Trend
					} else {
						sectorChange = sectorRandomChange
					}
				}
			}
		}

		// Validate sector change
		if math.IsInf(sectorChange, 0) || math.IsNaN(sectorChange) {
			sectorChange = 0
		}

		// Combined change: 70% individual, 30% sector
		totalChange := (individualChange * 0.7) + (sectorChange * 0.3)

		// Validate total change - cap extreme values
		if math.IsInf(totalChange, 0) || math.IsNaN(totalChange) {
			totalChange = 0 // Reset to no change
		} else if totalChange > 0.5 { // Cap at 50% change per update
			totalChange = 0.5
		} else if totalChange < -0.5 {
			totalChange = -0.5
		}

		// Calculate final price change
		newPrice := info.BasePrice * (1 + totalChange)

		// Validate the calculated price
		if math.IsInf(newPrice, 0) || math.IsNaN(newPrice) || newPrice <= 0 {
			// Reset to a safe price if calculation went wrong
			newPrice = 10.00
		}

		// Ensure price doesn't go below $0.01 (penny stock floor)
		if newPrice < 0.01 {
			newPrice = 0.01
		} else if newPrice > 1000000 { // Cap at $1M per share
			newPrice = 1000000
		}

		// Add some randomness to make prices jumpy sometimes (market surprises)
		if rand.Float64() < 0.05 { // 5% chance of a price jump
			jumpMultiplier := 1.0
			if rand.Float64() < 0.5 {
				// Positive jump
				jumpMultiplier = 1.0 + (rand.Float64() * 0.05) // 0-5% jump up
			} else {
				// Negative jump
				jumpMultiplier = 1.0 - (rand.Float64() * 0.05) // 0-5% jump down
			}
			newPrice *= jumpMultiplier

			// Validate after jump
			if math.IsInf(newPrice, 0) || math.IsNaN(newPrice) || newPrice <= 0 {
				newPrice = 0.01
			}
		}

		// Round to 2 decimal places
		newPrice = math.Round(newPrice*100) / 100

		// Final validation after rounding
		if math.IsInf(newPrice, 0) || math.IsNaN(newPrice) || newPrice <= 0 {
			newPrice = 0.01
		}

		// Update the base price for future calculations
		info.BasePrice = newPrice
		s.stocksInfo[id] = info

		// Check for crisis events if stock hits $0.01
		if newPrice <= 0.01 {
			s.processCrisisEvent(id, info)
		}

		// Send the update
		update := StockUpdate{
			StockID: id,
			Symbol:  info.Symbol,
			Price:   newPrice,
		}

		select {
		case s.updateChan <- update:
			// Successfully sent update
		default:
			// Channel is full, skip this update
			log.Printf("⚠️ MarketSimulator: Update channel full, skipping update for %s", info.Symbol)
		}
	}
}

// processCrisisEvent handles bankruptcy/recovery events for stocks at $0.01
func (s *MarketSimulator) processCrisisEvent(stockID int, stock StockInfo) {
	log.Printf("🚨 CRISIS EVENT: %s at $0.01 - Rolling for fate...", stock.Symbol)

	// Generate crisis news if this is the first time hitting $0.01
	// We'll track this better later, for now just generate news occasionally
	if s.newsService != nil && rand.Float64() < 0.3 { // 30% chance to generate crisis news
		s.newsService.GenerateCrisisNews(stock.ID, stock.Symbol, stock.Name, stock.Sector)
	}

	// Crisis event probabilities
	roll := rand.Float64()

	if roll < 0.05 { // 5% chance of bankruptcy
		log.Printf("💀 BANKRUPTCY: %s going bankrupt!", stock.Symbol)
		s.triggerBankruptcy(stockID, stock)
	} else if roll < 0.08 { // 3% chance of recovery (5% + 3% = 8% total)
		log.Printf("🚀 RECOVERY: %s staging dramatic comeback!", stock.Symbol)
		s.triggerRecovery(stockID, stock)
	} else {
		// 92% chance of stagnation - just stay at $0.01
		log.Printf("⏳ STAGNATION: %s remains in crisis at $0.01", stock.Symbol)
	}
}

// triggerBankruptcy processes a stock bankruptcy event
func (s *MarketSimulator) triggerBankruptcy(stockID int, stock StockInfo) {
	log.Printf("📰 NEWS: %s files for bankruptcy - stock delisted", stock.Symbol)

	// Generate bankruptcy news
	if s.newsService != nil {
		s.newsService.GenerateBankruptcyNews(stock.ID, stock.Symbol, stock.Name, stock.Sector)
	}

	// Process bankruptcy with portfolio handling
	if s.bankruptcyHandler != nil {
		err := s.bankruptcyHandler.ProcessStockBankruptcy(stockID)
		if err != nil {
			log.Printf("❌ Error processing bankruptcy for %s: %v", stock.Symbol, err)
		}
	} else {
		log.Printf("⚠️ No bankruptcy handler set - portfolio losses not recorded")
	}

	// Apply sector contagion after processing bankruptcy
	s.applySectorContagion(stockID, stock, "bankruptcy")

	// Replace the bankrupt company with a new one (simulate new IPO)
	stock.BasePrice = rand.Float64()*5 + 1 // $1-6 range for "new company"
	stock.Trend = 0                        // Reset trend
	s.stocksInfo[stockID] = stock

	log.Printf("📈 SIMULATION: %s replaced with new company at $%.2f", stock.Symbol, stock.BasePrice)
}

// triggerRecovery processes a stock recovery event
func (s *MarketSimulator) triggerRecovery(stockID int, stock StockInfo) {
	log.Printf("📰 NEWS: Surprise acquisition saves %s!", stock.Symbol)

	// Recovery jump: 10x to 100x potential (1-5 dollar range)
	recoveryMultiplier := rand.Float64()*400 + 100 // 100x to 500x
	newPrice := 0.01 * recoveryMultiplier

	// Cap recovery to reasonable range
	if newPrice > 50 {
		newPrice = rand.Float64()*30 + 10 // $10-40 range for major recovery
	}

	// Generate recovery news before updating price
	if s.newsService != nil {
		s.newsService.GenerateRecoveryNews(stock.ID, stock.Symbol, stock.Name, stock.Sector, newPrice)
	}

	stock.BasePrice = newPrice
	stock.Trend = 0.02 + rand.Float64()*0.03 // Positive trend for a while
	stock.TrendCounter = rand.Intn(20) + 10  // Longer positive trend
	s.stocksInfo[stockID] = stock

	log.Printf("🚀 RECOVERY: %s jumps to $%.2f (%.0fx return!)", stock.Symbol, newPrice, newPrice/0.01)

	// Apply positive sector contagion
	s.applySectorContagion(stockID, stock, "recovery")
}

// applySectorContagion applies crisis effects to sector peers
func (s *MarketSimulator) applySectorContagion(stockID int, stock StockInfo, eventType string) {
	if stock.SectorID == 0 {
		return // No sector assigned
	}

	log.Printf("🔗 CONTAGION: Applying %s contagion to %s sector", eventType, stock.Sector)

	contagionCount := 0
	for id, peerStock := range s.stocksInfo {
		if id != stockID && peerStock.SectorID == stock.SectorID {
			switch eventType {
			case "bankruptcy":
				// Major negative impact on sector
				peerStock.Trend -= 0.05 // Push trend strongly negative

				// Small chance to trigger crisis in vulnerable stocks
				if peerStock.BasePrice < 10.0 && rand.Float64() < 0.1 {
					crashAmount := 0.3 + rand.Float64()*0.4 // 30-70% crash
					peerStock.BasePrice *= (1 - crashAmount)
					if peerStock.BasePrice < 0.01 {
						peerStock.BasePrice = 0.01
					}
					log.Printf("💥 CONTAGION CRASH: %s drops %.0f%% due to sector crisis", peerStock.Symbol, crashAmount*100)
				}
				contagionCount++

			case "recovery":
				// Positive sentiment for sector
				peerStock.Trend += 0.01 + rand.Float64()*0.02 // 1-3% positive trend
				peerStock.TrendCounter = rand.Intn(10) + 5    // Short-term boost
				log.Printf("📈 CONTAGION BOOST: %s gets positive sentiment", peerStock.Symbol)
				contagionCount++
			}

			s.stocksInfo[id] = peerStock
		}
	}

	log.Printf("🔗 CONTAGION: Affected %d stocks in %s sector", contagionCount, stock.Sector)

	// Generate sector contagion news if multiple stocks are affected
	if contagionCount >= 2 && s.newsService != nil {
		s.newsService.GenerateSectorContagionNews(stock.Sector, eventType, contagionCount)
	}
}

// Pause pauses the market simulation (prevents price updates)
func (s *MarketSimulator) Pause() {
	select {
	case s.pauseChan <- true:
		// Successfully sent pause signal
	default:
		// Channel is full, but that's okay - we're already paused or will be soon
	}
}

// Resume resumes the market simulation
func (s *MarketSimulator) Resume() {
	select {
	case s.pauseChan <- false:
		// Successfully sent resume signal
	default:
		// Channel is full, but that's okay - we're already resumed or will be soon
	}
}

// ReloadStock updates the base price for a stock (used after admin price resets)
func (s *MarketSimulator) ReloadStock(stockID int, newPrice float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stock, exists := s.stocksInfo[stockID]; exists {
		// Validate and set new price
		if math.IsInf(newPrice, 0) || math.IsNaN(newPrice) || newPrice <= 0 {
			newPrice = 0.01
		}
		stock.BasePrice = newPrice

		// Lock the price for 30 seconds after admin reset to prevent race conditions
		stock.LockedUntil = time.Now().Add(30 * time.Second)
		stock.LockedPrice = newPrice

		// Reset trend to prevent carrying over corrupted values
		stock.Trend = 0
		stock.TrendCounter = rand.Intn(10) + 5

		s.stocksInfo[stockID] = stock
		log.Printf("🔒 Stock %s (ID: %d) reloaded at $%.2f and locked for 30 seconds", stock.Symbol, stockID, newPrice)
	}
}

// ValidateAllStocks checks and fixes all stocks in the simulator
func (s *MarketSimulator) ValidateAllStocks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, stock := range s.stocksInfo {
		fixed := false

		// Fix invalid base prices
		if math.IsInf(stock.BasePrice, 0) || math.IsNaN(stock.BasePrice) || stock.BasePrice <= 0 {
			stock.BasePrice = 0.01
			fixed = true
		}

		// Fix invalid trends
		if math.IsInf(stock.Trend, 0) || math.IsNaN(stock.Trend) {
			stock.Trend = 0
			fixed = true
		}

		// Cap extreme trends
		if stock.Trend > 0.1 {
			stock.Trend = 0.1
			fixed = true
		} else if stock.Trend < -0.1 {
			stock.Trend = -0.1
			fixed = true
		}

		if fixed {
			s.stocksInfo[id] = stock
			log.Printf("Fixed corrupted stock data for %s: price=%.2f, trend=%.4f", stock.Symbol, stock.BasePrice, stock.Trend)
		}
	}
}

// Testing methods for crisis events

// ForceCrisisEvent forces a stock into crisis at $0.01 for testing
func (s *MarketSimulator) ForceCrisisEvent(stockID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stock, exists := s.stocksInfo[stockID]
	if !exists {
		return fmt.Errorf("stock with ID %d not found", stockID)
	}

	log.Printf("🧪 TESTING: Forcing crisis event for %s", stock.Symbol)

	// Set stock to crisis price
	stock.BasePrice = 0.01
	stock.Trend = -0.05 // Strong negative trend
	s.stocksInfo[stockID] = stock

	// Trigger crisis event processing
	s.processCrisisEvent(stockID, stock)

	// Send price update
	select {
	case s.updateChan <- StockUpdate{
		StockID: stockID,
		Symbol:  stock.Symbol,
		Price:   0.01,
	}:
	default:
		// Channel full, skip update
	}

	return nil
}

// ForceBankruptcy forces a stock into bankruptcy for testing
func (s *MarketSimulator) ForceBankruptcy(stockID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stock, exists := s.stocksInfo[stockID]
	if !exists {
		return fmt.Errorf("stock with ID %d not found", stockID)
	}

	log.Printf("🧪 TESTING: Forcing bankruptcy for %s", stock.Symbol)

	// Set to crisis price first
	stock.BasePrice = 0.01
	s.stocksInfo[stockID] = stock

	// Trigger bankruptcy directly
	s.triggerBankruptcy(stockID, stock)

	return nil
}

// ForceRecovery forces a stock recovery for testing
func (s *MarketSimulator) ForceRecovery(stockID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stock, exists := s.stocksInfo[stockID]
	if !exists {
		return fmt.Errorf("stock with ID %d not found", stockID)
	}

	log.Printf("🧪 TESTING: Forcing recovery for %s", stock.Symbol)

	// Set to crisis price first
	stock.BasePrice = 0.01
	s.stocksInfo[stockID] = stock

	// Trigger recovery directly
	s.triggerRecovery(stockID, stock)

	return nil
}

// UnlockStock removes the admin lock from a stock, allowing automatic updates to resume
func (s *MarketSimulator) UnlockStock(stockID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stock, exists := s.stocksInfo[stockID]; exists {
		stock.LockedUntil = time.Time{} // Clear the lock
		stock.LockedPrice = 0
		s.stocksInfo[stockID] = stock
		log.Printf("🔓 Stock %s (ID: %d) manually unlocked", stock.Symbol, stockID)
	}
}

// GetStockInfo returns current stock information for testing
func (s *MarketSimulator) GetStockInfo(stockID int) (StockInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stock, exists := s.stocksInfo[stockID]
	return stock, exists
}

// ListAllStocks returns all stocks in the simulator for testing
func (s *MarketSimulator) ListAllStocks() map[int]StockInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[int]StockInfo)
	for id, stock := range s.stocksInfo {
		result[id] = stock
	}
	return result
}

// ProcessTransaction simulates market impact of a transaction
func (s *MarketSimulator) ProcessTransaction(stockID int, quantity int, isBuy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stock, exists := s.stocksInfo[stockID]
	if !exists {
		return
	}

	// Enhanced market impact calculation
	// Buys push prices up, sells push prices down
	// Impact depends on quantity, current price, and is amplified by current trend

	// Base impact factor - higher than before for more dramatic changes
	impactFactor := 0.0005 * float64(quantity) // 0.05% per share

	// For larger transactions, apply diminishing returns (sqrt scaling)
	if quantity > 100 {
		impactFactor = 0.0005 * (100 + math.Sqrt(float64(quantity-100)))
	}

	// Reverse the direction for sells
	if !isBuy {
		impactFactor = -impactFactor
	}

	// Amplify impact if it aligns with current trend (momentum effect)
	// If buy during uptrend or sell during downtrend, amplify the effect
	if (isBuy && stock.Trend > 0) || (!isBuy && stock.Trend < 0) {
		impactFactor *= 1.5
	}

	// Calculate new price
	newPrice := stock.BasePrice * (1 + impactFactor)

	// Validate the calculated price
	if math.IsInf(newPrice, 0) || math.IsNaN(newPrice) || newPrice <= 0 {
		newPrice = 0.01
	}

	if newPrice < 0.01 {
		newPrice = 0.01
	} else if newPrice > 1000000 { // Cap at $1M per share
		newPrice = 1000000
	}

	// Round to 2 decimal places
	newPrice = math.Round(newPrice*100) / 100

	// Update the base price
	stock.BasePrice = newPrice

	// Transactions can influence the trend slightly
	// Large buys push trend up, large sells push trend down
	trendAdjustment := impactFactor * 0.2
	stock.Trend += trendAdjustment

	// Cap the trend to prevent extreme values
	const maxTrend = 0.1
	if stock.Trend > maxTrend {
		stock.Trend = maxTrend
	} else if stock.Trend < -maxTrend {
		stock.Trend = -maxTrend
	}

	s.stocksInfo[stockID] = stock

	// Send the update
	select {
	case s.updateChan <- StockUpdate{
		StockID: stockID,
		Symbol:  stock.Symbol,
		Price:   newPrice,
	}:
	default:
		// Channel is full, skip this update
	}
}

// getPriceZoneVolatility returns volatility based on stock price range
// Implements the price zone system from NEWS_UPDATE_PLAN.md
func (s *MarketSimulator) getPriceZoneVolatility(price float64) float64 {
	switch {
	case price <= 1.0:
		// Penny stocks: Reduced volatility to prevent death spiral (3%)
		return 0.03
	case price <= 10.0:
		// Low-cap: Medium volatility (5%)
		return 0.05
	case price <= 100.0:
		// Mid-cap: Normal volatility (5%)
		return s.volatility // Use base volatility (5%)
	default:
		// Large-cap: Low volatility (3%)
		return 0.03
	}
}
