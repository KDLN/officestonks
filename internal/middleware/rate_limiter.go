package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements rate limiting with a sliding window approach
type RateLimiter struct {
	rate      int           // Max requests per time window
	window    time.Duration // Time window for rate limiting
	clients   map[string]*clientWindow
	stats     statsData
	mu        sync.Mutex
	cleanupMu sync.Mutex
}

// clientWindow tracks request history for a single client
type clientWindow struct {
	hits       []time.Time // Timestamps of requests within current window
	totalHits  int         // Total hits since tracking began
	blocked    int         // Number of times this client was blocked
	lastUpdate time.Time   // Last time this client made a request
}

// statsData contains global rate limiter statistics
type statsData struct {
	totalRequests   int       // Total number of requests processed
	lastMinuteHits  int       // Hits in the last minute
	totalBlocks     int       // Total number of requests blocked
	highestClientIP string    // IP with the most requests
	highestHits     int       // Number of hits for that IP
	lastUpdate      time.Time // Last time stats were updated
}

// NewRateLimiter creates a new rate limiter with the specified rate and window
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:    rate,
		window:  window,
		clients: make(map[string]*clientWindow),
		stats: statsData{
			lastUpdate: time.Now(),
		},
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
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

		// Initialize client window if not exists
		client, exists := rl.clients[clientIP]
		if !exists {
			client = &clientWindow{
				hits:       make([]time.Time, 0, rl.rate),
				totalHits:  0,
				lastUpdate: time.Now(),
			}
			rl.clients[clientIP] = client
		}

		// Update client stats
		client.totalHits++
		client.lastUpdate = time.Now()

		// Update highest hits client if this one has more
		if client.totalHits > rl.stats.highestHits {
			rl.stats.highestHits = client.totalHits
			rl.stats.highestClientIP = clientIP
		}

		// Remove expired hits
		cutoff := time.Now().Add(-rl.window)
		validHits := make([]time.Time, 0, len(client.hits))
		for _, hit := range client.hits {
			if hit.After(cutoff) {
				validHits = append(validHits, hit)
			}
		}
		client.hits = validHits

		// Check if rate limit is exceeded
		if len(client.hits) >= rl.rate {
			// Rate limit exceeded
			client.blocked++
			rl.stats.totalBlocks++
			rl.mu.Unlock()

			// Return 429 Too Many Requests
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Add current request timestamp
		client.hits = append(client.hits, time.Now())

		// Unlock before calling next handler
		rl.mu.Unlock()

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// cleanupRoutine periodically cleans up stale clients to prevent memory leaks
func (rl *RateLimiter) cleanupRoutine() {
	for {
		time.Sleep(rl.window) // Run cleanup at the window interval

		// Cleanup inactive clients older than 3x the window duration
		rl.cleanupMu.Lock()
		staleTime := time.Now().Add(-3 * rl.window)

		rl.mu.Lock()
		for ip, client := range rl.clients {
			if client.lastUpdate.Before(staleTime) {
				delete(rl.clients, ip)
			}
		}

		// Reset last minute hits counter every minute
		if time.Since(rl.stats.lastUpdate) > time.Minute {
			rl.stats.lastMinuteHits = 0
			rl.stats.lastUpdate = time.Now()
		}

		rl.mu.Unlock()
		rl.cleanupMu.Unlock()
	}
}

// GetIPAddress extracts the client IP from various headers or the remote address
func getIPAddress(r *http.Request) string {
	// Check for X-Real-IP header (set by some proxies)
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Check for X-Forwarded-For header (set by proxies)
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs; take the first one
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}

	// Check for CF-Connecting-IP (set by Cloudflare)
	ip = r.Header.Get("CF-Connecting-IP")
	if ip != "" {
		return ip
	}

	// Check for True-Client-IP (set by Akamai and some CDNs)
	ip = r.Header.Get("True-Client-IP")
	if ip != "" {
		return ip
	}

	// Use remote address as fallback
	return r.RemoteAddr
}

// GetStats returns a copy of the current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return map[string]interface{}{
		"totalRequests":  rl.stats.totalRequests,
		"lastMinuteHits": rl.stats.lastMinuteHits,
		"totalBlocked":   rl.stats.totalBlocks,
		"highestIP":      rl.stats.highestClientIP,
		"highestHits":    rl.stats.highestHits,
		"activeClients":  len(rl.clients),
	}
}
