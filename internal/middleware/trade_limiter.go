package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// TradeLimiter tracks and limits trading frequency per user
type TradeLimiter struct {
	mu               sync.RWMutex
	userLastTrade    map[int]time.Time
	userTradeCount   map[int][]time.Time
	cooldownSeconds  int
	maxTradesPerHour int
}

// NewTradeLimiter creates a new trade limiter
func NewTradeLimiter(cooldownSeconds, maxTradesPerHour int) *TradeLimiter {
	tl := &TradeLimiter{
		userLastTrade:    make(map[int]time.Time),
		userTradeCount:   make(map[int][]time.Time),
		cooldownSeconds:  cooldownSeconds,
		maxTradesPerHour: maxTradesPerHour,
	}
	
	// Start cleanup goroutine to remove old trade records
	go tl.cleanupOldRecords()
	
	return tl
}

// Middleware returns the HTTP middleware function
func (tl *TradeLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to trade endpoints
		if r.URL.Path != "/api/stocks/trade" {
			next.ServeHTTP(w, r)
			return
		}
		
		// Extract user ID from context
		userIDValue := r.Context().Value("userID")
		if userIDValue == nil {
			next.ServeHTTP(w, r)
			return
		}
		
		userID, err := strconv.Atoi(fmt.Sprintf("%v", userIDValue))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		
		// Check trade limits
		canTrade, reason := tl.CanUserTrade(userID)
		if !canTrade {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": reason,
			})
			return
		}
		
		// Record the trade attempt
		tl.RecordTrade(userID)
		
		next.ServeHTTP(w, r)
	})
}

// CanUserTrade checks if a user can make a trade
func (tl *TradeLimiter) CanUserTrade(userID int) (bool, string) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	
	now := time.Now()
	
	// Check cooldown
	if lastTrade, exists := tl.userLastTrade[userID]; exists {
		timeSinceLastTrade := now.Sub(lastTrade)
		if timeSinceLastTrade < time.Duration(tl.cooldownSeconds)*time.Second {
			remainingTime := time.Duration(tl.cooldownSeconds)*time.Second - timeSinceLastTrade
			return false, fmt.Sprintf("Please wait %.0f seconds before making another trade", remainingTime.Seconds())
		}
	}
	
	// Check hourly limit
	trades := tl.userTradeCount[userID]
	recentTrades := 0
	oneHourAgo := now.Add(-time.Hour)
	
	for _, tradeTime := range trades {
		if tradeTime.After(oneHourAgo) {
			recentTrades++
		}
	}
	
	if recentTrades >= tl.maxTradesPerHour {
		return false, fmt.Sprintf("You have reached the maximum of %d trades per hour", tl.maxTradesPerHour)
	}
	
	return true, ""
}

// RecordTrade records a trade for a user
func (tl *TradeLimiter) RecordTrade(userID int) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	
	now := time.Now()
	tl.userLastTrade[userID] = now
	
	if tl.userTradeCount[userID] == nil {
		tl.userTradeCount[userID] = []time.Time{}
	}
	tl.userTradeCount[userID] = append(tl.userTradeCount[userID], now)
}

// cleanupOldRecords periodically removes old trade records
func (tl *TradeLimiter) cleanupOldRecords() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		tl.mu.Lock()
		now := time.Now()
		oneHourAgo := now.Add(-time.Hour)
		
		// Clean up trade counts
		for userID, trades := range tl.userTradeCount {
			var recentTrades []time.Time
			for _, tradeTime := range trades {
				if tradeTime.After(oneHourAgo) {
					recentTrades = append(recentTrades, tradeTime)
				}
			}
			
			if len(recentTrades) == 0 {
				delete(tl.userTradeCount, userID)
			} else {
				tl.userTradeCount[userID] = recentTrades
			}
		}
		
		// Clean up last trade times older than 1 hour
		for userID, lastTrade := range tl.userLastTrade {
			if lastTrade.Before(oneHourAgo) {
				delete(tl.userLastTrade, userID)
			}
		}
		
		tl.mu.Unlock()
	}
}