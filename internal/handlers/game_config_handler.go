package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"officestonks/internal/config"
)

// GameConfigHandler handles game configuration management
type GameConfigHandler struct {
	currentConfig *config.GameConfig
}

// NewGameConfigHandler creates a new game configuration handler
func NewGameConfigHandler() *GameConfigHandler {
	return &GameConfigHandler{
		currentConfig: config.DefaultGameConfig(),
	}
}

// GameConfigResponse represents the response format for game config
type GameConfigResponse struct {
	StartingCash         float64 `json:"starting_cash"`
	MarketUpdateInterval int     `json:"market_update_interval_seconds"`
	MarketVolatility     float64 `json:"market_volatility"`
	MinTradeQuantity     int     `json:"min_trade_quantity"`
	MaxTradeQuantity     int     `json:"max_trade_quantity"`
	TradeCooldownSeconds int     `json:"trade_cooldown_seconds"`
	MaxTradesPerHour     int     `json:"max_trades_per_hour"`
	BaseImpactFactor     float64 `json:"base_impact_factor"`
	LargeTradeThreshold  int     `json:"large_trade_threshold"`
	TrendAmplification   float64 `json:"trend_amplification"`
	MinStockPrice        float64 `json:"min_stock_price"`
	MaxStockPrice        float64 `json:"max_stock_price"`
}

// UpdateGameConfigRequest represents the request format for updating game config
type UpdateGameConfigRequest struct {
	StartingCash         *float64 `json:"starting_cash,omitempty"`
	MarketUpdateInterval *int     `json:"market_update_interval_seconds,omitempty"`
	MarketVolatility     *float64 `json:"market_volatility,omitempty"`
	MinTradeQuantity     *int     `json:"min_trade_quantity,omitempty"`
	MaxTradeQuantity     *int     `json:"max_trade_quantity,omitempty"`
	TradeCooldownSeconds *int     `json:"trade_cooldown_seconds,omitempty"`
	MaxTradesPerHour     *int     `json:"max_trades_per_hour,omitempty"`
	BaseImpactFactor     *float64 `json:"base_impact_factor,omitempty"`
	LargeTradeThreshold  *int     `json:"large_trade_threshold,omitempty"`
	TrendAmplification   *float64 `json:"trend_amplification,omitempty"`
	MinStockPrice        *float64 `json:"min_stock_price,omitempty"`
	MaxStockPrice        *float64 `json:"max_stock_price,omitempty"`
}

// GetGameConfig returns the current game configuration
func (h *GameConfigHandler) GetGameConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := GameConfigResponse{
		StartingCash:         h.currentConfig.StartingCash,
		MarketUpdateInterval: int(h.currentConfig.MarketUpdateInterval.Seconds()),
		MarketVolatility:     h.currentConfig.MarketVolatility,
		MinTradeQuantity:     h.currentConfig.MinTradeQuantity,
		MaxTradeQuantity:     h.currentConfig.MaxTradeQuantity,
		TradeCooldownSeconds: h.currentConfig.TradeCooldownSeconds,
		MaxTradesPerHour:     h.currentConfig.MaxTradesPerHour,
		BaseImpactFactor:     h.currentConfig.BaseImpactFactor,
		LargeTradeThreshold:  h.currentConfig.LargeTradeThreshold,
		TrendAmplification:   h.currentConfig.TrendAmplification,
		MinStockPrice:        h.currentConfig.MinStockPrice,
		MaxStockPrice:        h.currentConfig.MaxStockPrice,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding game config response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// UpdateGameConfig updates the game configuration
func (h *GameConfigHandler) UpdateGameConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateGameConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate and update configuration
	newConfig := *h.currentConfig // Copy current config

	if req.StartingCash != nil {
		if *req.StartingCash < 1000 || *req.StartingCash > 1000000 {
			http.Error(w, "Starting cash must be between $1,000 and $1,000,000", http.StatusBadRequest)
			return
		}
		newConfig.StartingCash = *req.StartingCash
	}

	if req.MarketUpdateInterval != nil {
		if *req.MarketUpdateInterval < 1 || *req.MarketUpdateInterval > 60 {
			http.Error(w, "Market update interval must be between 1 and 60 seconds", http.StatusBadRequest)
			return
		}
		newConfig.MarketUpdateInterval = time.Duration(*req.MarketUpdateInterval) * time.Second
	}

	if req.MarketVolatility != nil {
		if *req.MarketVolatility < 0.001 || *req.MarketVolatility > 0.2 {
			http.Error(w, "Market volatility must be between 0.1% and 20%", http.StatusBadRequest)
			return
		}
		newConfig.MarketVolatility = *req.MarketVolatility
	}

	if req.MinTradeQuantity != nil {
		if *req.MinTradeQuantity < 1 || *req.MinTradeQuantity > 100 {
			http.Error(w, "Min trade quantity must be between 1 and 100", http.StatusBadRequest)
			return
		}
		newConfig.MinTradeQuantity = *req.MinTradeQuantity
	}

	if req.MaxTradeQuantity != nil {
		if *req.MaxTradeQuantity < 1 || *req.MaxTradeQuantity > 100000 {
			http.Error(w, "Max trade quantity must be between 1 and 100,000", http.StatusBadRequest)
			return
		}
		newConfig.MaxTradeQuantity = *req.MaxTradeQuantity
	}

	if req.TradeCooldownSeconds != nil {
		if *req.TradeCooldownSeconds < 0 || *req.TradeCooldownSeconds > 300 {
			http.Error(w, "Trade cooldown must be between 0 and 300 seconds", http.StatusBadRequest)
			return
		}
		newConfig.TradeCooldownSeconds = *req.TradeCooldownSeconds
	}

	if req.MaxTradesPerHour != nil {
		if *req.MaxTradesPerHour < 1 || *req.MaxTradesPerHour > 1000 {
			http.Error(w, "Max trades per hour must be between 1 and 1,000", http.StatusBadRequest)
			return
		}
		newConfig.MaxTradesPerHour = *req.MaxTradesPerHour
	}

	if req.BaseImpactFactor != nil {
		if *req.BaseImpactFactor < 0 || *req.BaseImpactFactor > 0.01 {
			http.Error(w, "Base impact factor must be between 0 and 1%", http.StatusBadRequest)
			return
		}
		newConfig.BaseImpactFactor = *req.BaseImpactFactor
	}

	if req.LargeTradeThreshold != nil {
		if *req.LargeTradeThreshold < 10 || *req.LargeTradeThreshold > 10000 {
			http.Error(w, "Large trade threshold must be between 10 and 10,000", http.StatusBadRequest)
			return
		}
		newConfig.LargeTradeThreshold = *req.LargeTradeThreshold
	}

	if req.TrendAmplification != nil {
		if *req.TrendAmplification < 1 || *req.TrendAmplification > 5 {
			http.Error(w, "Trend amplification must be between 1.0 and 5.0", http.StatusBadRequest)
			return
		}
		newConfig.TrendAmplification = *req.TrendAmplification
	}

	if req.MinStockPrice != nil {
		if *req.MinStockPrice < 0.01 || *req.MinStockPrice > 10 {
			http.Error(w, "Min stock price must be between $0.01 and $10", http.StatusBadRequest)
			return
		}
		newConfig.MinStockPrice = *req.MinStockPrice
	}

	if req.MaxStockPrice != nil {
		if *req.MaxStockPrice < 100 || *req.MaxStockPrice > 1000000 {
			http.Error(w, "Max stock price must be between $100 and $1,000,000", http.StatusBadRequest)
			return
		}
		newConfig.MaxStockPrice = *req.MaxStockPrice
	}

	// Update the configuration
	h.currentConfig = &newConfig

	log.Printf("Game configuration updated by admin")

	// Return the updated configuration
	h.GetGameConfig(w, r)
}

// ResetGameConfig resets the game configuration to defaults
func (h *GameConfigHandler) ResetGameConfig(w http.ResponseWriter, r *http.Request) {
	h.currentConfig = config.DefaultGameConfig()
	log.Printf("Game configuration reset to defaults by admin")
	
	h.GetGameConfig(w, r)
}

// LoadBalancedConfig loads the balanced game configuration
func (h *GameConfigHandler) LoadBalancedConfig(w http.ResponseWriter, r *http.Request) {
	h.currentConfig = config.BalancedGameConfig()
	log.Printf("Game configuration set to balanced preset by admin")
	
	h.GetGameConfig(w, r)
}

// GetCurrentConfig returns the current configuration (for use by other handlers)
func (h *GameConfigHandler) GetCurrentConfig() *config.GameConfig {
	return h.currentConfig
}