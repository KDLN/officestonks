package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a simple rate limiting middleware
type RateLimiter struct {
	// Map to store client requests
	clients map[string][]time.Time
	// Maximum number of requests allowed in the time window
	maxRequests int
	// Time window for rate limiting (e.g., 1 minute)
	window time.Duration
	// Mutex for thread safety
	mu sync.Mutex
	// Stats for monitoring
	stats struct {
		totalRequests     int
		blockedRequests   int
		lastMinuteHits    int
		lastReset         time.Time
	}
}

// NewRateLimiter creates a new rate limiter with the specified parameters
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}

	// Initialize stats
	rl.stats.lastReset = time.Now()

	// Start a goroutine to reset statistics periodically
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			rl.stats.lastMinuteHits = 0
			rl.stats.lastReset = time.Now()
			rl.mu.Unlock()
		}
	}()

	return rl
}

// cleanupOldRequests removes requests that are outside the time window
func (rl *RateLimiter) cleanupOldRequests(clientIP string) {
	now := time.Now()
	// Keep only requests within the time window
	validRequests := []time.Time{}
	
	for _, timestamp := range rl.clients[clientIP] {
		if now.Sub(timestamp) <= rl.window {
			validRequests = append(validRequests, timestamp)
		}
	}
	
	// Update the client's requests
	rl.clients[clientIP] = validRequests
}

// getIPAddress extracts the client's IP address from the request
// It respects X-Forwarded-For and X-Real-IP headers for proxied requests
func getIPAddress(r *http.Request) string {
	// Check for X-Forwarded-For header (common with proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Use the leftmost IP in the chain (client's original IP)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}

	// Check for X-Real-IP header (used by some proxies)
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to RemoteAddr if no proxy headers are found
	// Strip port number if present
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there was an error (e.g., no port in the address), use RemoteAddr as is
		return r.RemoteAddr
	}

	return ip
}

// RateLimit is a middleware that limits requests based on client IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP address
		clientIP := getIPAddress(r)

		// Lock for thread safety
		rl.mu.Lock()

		// Update statistics
		rl.stats.totalRequests++
		rl.stats.lastMinuteHits++

		// Clean up old requests
		rl.cleanupOldRequests(clientIP)

		// Check if client has exceeded rate limit
		if len(rl.clients[clientIP]) >= rl.maxRequests {
			// Too many requests, return 429 status
			rl.stats.blockedRequests++
			rl.mu.Unlock() // Unlock before returning response

			w.Header().Set("Retry-After", rl.window.String())
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded. Please try again later."))
			return
		}

		// Add current request timestamp
		rl.clients[clientIP] = append(rl.clients[clientIP], time.Now())

		// Calculate remaining requests
		remaining := rl.maxRequests - len(rl.clients[clientIP])

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		rl.mu.Unlock() // Unlock before calling next handler

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// GetStats returns the current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return map[string]interface{}{
		"total_requests":   rl.stats.totalRequests,
		"blocked_requests": rl.stats.blockedRequests,
		"last_minute_hits": rl.stats.lastMinuteHits,
		"active_clients":   len(rl.clients),
		"last_reset":       rl.stats.lastReset.Format(time.RFC3339),
	}
}