package market

import (
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
	ID       int
	Symbol   string
	BasePrice float64
	Sector   string
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
	
	s.stocksInfo[id] = StockInfo{
		ID:       id,
		Symbol:   symbol,
		BasePrice: basePrice,
		Sector:   sector,
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Update all stocks
	for id, info := range s.stocksInfo {
		// Calculate new price with random fluctuation
		// Price changes are percentage-based and depend on volatility
		changePercent := (rand.Float64() - 0.5) * s.volatility
		newPrice := info.BasePrice * (1 + changePercent)
		
		// Ensure price doesn't go below 0.01
		if newPrice < 0.01 {
			newPrice = 0.01
		}
		
		// Round to 2 decimal places
		newPrice = float64(int(newPrice * 100)) / 100
		
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
	
	// Simplified market impact calculation
	// Buys push prices up, sells push prices down
	// Impact depends on quantity and current price
	impactFactor := 0.0001 * float64(quantity) // 0.01% per share
	if !isBuy {
		impactFactor = -impactFactor
	}
	
	// Calculate new price
	newPrice := stock.BasePrice * (1 + impactFactor)
	if newPrice < 0.01 {
		newPrice = 0.01
	}
	
	// Update the base price
	stock.BasePrice = newPrice
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