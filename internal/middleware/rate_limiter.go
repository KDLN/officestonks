package middleware

import (
	"fmt"
	"log"
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

// isPollingRequest checks if a request is an automated polling request that should be excluded from rate limiting
func isPollingRequest(r *http.Request) bool {
	// Check for polling endpoints
	if strings.HasSuffix(r.URL.Path, "/stock-updates/poll") {
		return true
	}
	
	// Check for SSE endpoints (also automated)
	if strings.HasSuffix(r.URL.Path, "/sse/stock-updates") {
		return true
	}
	
	// Check for health checks and monitoring endpoints
	if strings.HasSuffix(r.URL.Path, "/health") || strings.HasSuffix(r.URL.Path, "/health-check") {
		return true
	}
	
	// Check for User-Agent indicating automated requests
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(strings.ToLower(userAgent), "polling") || 
	   strings.Contains(strings.ToLower(userAgent), "automated") {
		return true
	}
	
	// Check for custom header indicating polling request
	if r.Header.Get("X-Request-Type") == "polling" {
		return true
	}
	
	// Exclude admin endpoints from rate limiting - admins should have unlimited access
	if strings.Contains(r.URL.Path, "/api/admin/") {
		return true
	}
	
	return false
}

// RateLimit is a middleware that limits requests based on client IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a polling request that should be excluded from rate limiting
		if isPollingRequest(r) {
			// Skip rate limiting for polling requests but still call next handler
			reason := "unknown"
			if strings.Contains(r.URL.Path, "/api/admin/") {
				reason = "admin endpoint"
			} else if strings.Contains(r.URL.Path, "/stock-updates/poll") || strings.Contains(r.URL.Path, "/sse/") {
				reason = "polling/sse endpoint"
			} else if r.Header.Get("X-Request-Type") == "polling" {
				reason = "polling header"
			} else if strings.Contains(r.URL.Path, "/health") {
				reason = "health check"
			}
			log.Printf("⚡ BYPASS: Request bypassed rate limiting (%s): %s %s", reason, r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

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
			
			// Enhanced logging for rate limit blocks
			log.Printf("🚫 RATE LIMIT: Client %s blocked (requests: %d/%d) for %s %s", 
				clientIP, len(rl.clients[clientIP]), rl.maxRequests, r.Method, r.URL.Path)
			log.Printf("🚫 RATE LIMIT: User-Agent: %s", r.Header.Get("User-Agent"))
			log.Printf("🚫 RATE LIMIT: Origin: %s", r.Header.Get("Origin"))
			log.Printf("🚫 RATE LIMIT: X-Request-Type: %s", r.Header.Get("X-Request-Type"))
			
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