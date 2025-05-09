package market

import (
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

// MarketSimulator handles the stock price simulation
type MarketSimulator struct {
	stocksInfo     map[int]StockInfo
	updateInterval time.Duration
	volatility     float64
	mu             sync.RWMutex
	updateChan     chan StockUpdate
	stopChan       chan struct{}
}

// StockInfo contains information about a stock for simulation
type StockInfo struct {
	ID           int
	Symbol       string
	BasePrice    float64
	Sector       string
	Trend        float64  // Bias for price movement: positive means upward trend, negative means downward
	TrendCounter int      // Counter to track trend duration
}

// NewMarketSimulator creates a new market simulator
func NewMarketSimulator(updateInterval time.Duration, volatility float64) *MarketSimulator {
	return &MarketSimulator{
		stocksInfo:     make(map[int]StockInfo),
		updateInterval: updateInterval,
		volatility:     volatility,
		updateChan:     make(chan StockUpdate, 100),
		stopChan:       make(chan struct{}),
	}
}

// AddStock adds a stock to the simulator
func (s *MarketSimulator) AddStock(id int, symbol, sector string, basePrice float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Initialize with a random trend (slightly biased upward for a bull market)
	initialTrend := (rand.Float64() * 0.1) - 0.03  // Range: -0.03 to 0.07, slightly positive bias

	s.stocksInfo[id] = StockInfo{
		ID:           id,
		Symbol:       symbol,
		BasePrice:    basePrice,
		Sector:       sector,
		Trend:        initialTrend,
		TrendCounter: rand.Intn(10) + 5, // Random initial trend duration (5-15 updates)
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
	
	for {
		select {
		case <-ticker.C:
			s.updatePrices()
		case <-s.stopChan:
			close(s.updateChan)
			return
		}
	}
}

// updatePrices calculates new prices for all stocks
func (s *MarketSimulator) updatePrices() {
	s.mu.Lock() // Use write lock since we're updating the stocksInfo
	defer s.mu.Unlock()

	// Update all stocks
	for id, info := range s.stocksInfo {
		// Check if we need to change the trend
		if info.TrendCounter <= 0 {
			// Time to reverse or modify the trend
			// Stronger reversal for extreme trends (mean reversion)
			reversalStrength := 1.0 + math.Abs(info.Trend)*5

			// Generate new trend - more likely to reverse direction
			if rand.Float64() < 0.7 { // 70% chance of trend reversal
				// Reverse the trend with some randomness, amplified by reversalStrength
				info.Trend = -info.Trend * (0.5 + rand.Float64()) * math.Min(reversalStrength, 3.0)
			} else {
				// Modify current trend with dampening (regression to mean)
				// Stronger dampening for extreme trends
				dampening := 0.3 + rand.Float64()*0.4 // 30-70% of current trend
				dampening /= math.Min(reversalStrength, 2.0) // More dampening for extreme trends
				info.Trend = info.Trend * dampening
			}

			// Set new duration for this trend
			info.TrendCounter = rand.Intn(15) + 5 // 5-20 updates
		} else {
			info.TrendCounter--
		}

		// Calculate new price with random fluctuation + trend bias
		// Base random change
		randomChange := (rand.Float64() - 0.5) * s.volatility

		// Add trend bias to the random change
		biasedChange := randomChange + info.Trend

		// Calculate final price change
		newPrice := info.BasePrice * (1 + biasedChange)

		// Ensure price doesn't go below 0.01
		if newPrice < 0.01 {
			newPrice = 0.01
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
		}

		// Round to 2 decimal places
		newPrice = math.Round(newPrice*100) / 100

		// Update the base price for future calculations
		info.BasePrice = newPrice
		s.stocksInfo[id] = info

		// Send the update
		select {
		case s.updateChan <- StockUpdate{
			StockID: id,
			Symbol:  info.Symbol,
			Price:   newPrice,
		}:
		default:
			// Channel is full, skip this update
		}
	}
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
	if newPrice < 0.01 {
		newPrice = 0.01
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