package config

import "time"

// GameConfig holds all game balance parameters
type GameConfig struct {
	// User settings
	StartingCash float64 `json:"starting_cash"`
	
	// Market settings
	MarketUpdateInterval time.Duration `json:"market_update_interval"`
	MarketVolatility     float64       `json:"market_volatility"`
	
	// Trading settings
	MinTradeQuantity      int           `json:"min_trade_quantity"`
	MaxTradeQuantity      int           `json:"max_trade_quantity"`
	TradeCooldownSeconds  int           `json:"trade_cooldown_seconds"`
	MaxTradesPerHour      int           `json:"max_trades_per_hour"`
	
	// Price impact settings
	BaseImpactFactor      float64       `json:"base_impact_factor"`
	LargeTradeThreshold   int           `json:"large_trade_threshold"`
	TrendAmplification    float64       `json:"trend_amplification"`
	
	// Price bounds
	MinStockPrice         float64       `json:"min_stock_price"`
	MaxStockPrice         float64       `json:"max_stock_price"`
}

// DefaultGameConfig returns the default game configuration
func DefaultGameConfig() *GameConfig {
	return &GameConfig{
		// User settings
		StartingCash: 10000.00,
		
		// Market settings - currently 2 seconds, 5% volatility
		MarketUpdateInterval: 2 * time.Second,
		MarketVolatility:     0.05,
		
		// Trading settings
		MinTradeQuantity:     1,
		MaxTradeQuantity:     10000,
		TradeCooldownSeconds: 2,      // Minimum 2 seconds between trades
		MaxTradesPerHour:     100,     // Maximum 100 trades per hour per user
		
		// Price impact settings - currently 0.05% per share
		BaseImpactFactor:    0.0005,
		LargeTradeThreshold: 100,
		TrendAmplification:  1.5,
		
		// Price bounds
		MinStockPrice: 0.01,
		MaxStockPrice: 100000.00,
	}
}

// BalancedGameConfig returns a more balanced configuration for better gameplay
func BalancedGameConfig() *GameConfig {
	return &GameConfig{
		// User settings
		StartingCash: 10000.00,
		
		// Market settings - slower updates, less volatility
		MarketUpdateInterval: 5 * time.Second,
		MarketVolatility:     0.02, // 2% volatility (was 5%)
		
		// Trading settings
		MinTradeQuantity:     1,
		MaxTradeQuantity:     1000,    // Limit large trades
		TradeCooldownSeconds: 5,       // 5 seconds between trades
		MaxTradesPerHour:     50,      // 50 trades per hour max
		
		// Price impact settings - reduced impact
		BaseImpactFactor:    0.0002,   // 0.02% per share (was 0.05%)
		LargeTradeThreshold: 100,
		TrendAmplification:  1.2,      // Less trend amplification
		
		// Price bounds
		MinStockPrice: 0.01,
		MaxStockPrice: 10000.00,
	}
}