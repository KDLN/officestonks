package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content Security Policy
			csp := strings.Join([]string{
				"default-src 'self'",
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Note: unsafe-inline/eval needed for React dev
				"style-src 'self' 'unsafe-inline'",
				"img-src 'self' data: https:",
				"font-src 'self'",
				"connect-src 'self' ws: wss:",
				"frame-ancestors 'none'",
				"base-uri 'self'",
				"form-action 'self'",
			}, "; ")
			w.Header().Set("Content-Security-Policy", csp)
			
			// X-Frame-Options to prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")
			
			// X-Content-Type-Options to prevent MIME sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")
			
			// X-XSS-Protection (legacy browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			
			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			
			// Remove server information
			w.Header().Set("Server", "") // Hide server info
			
			// For HTTPS deployments, add HSTS
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware with secure defaults
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// List of allowed origins
			allowedOrigins := []string{
				"http://localhost:3000",  // Development
				"http://localhost:8080",  // Alternative dev port
				"https://officestonks-frontend-production.up.railway.app", // Production (update as needed)
			}
			
			// Check if origin is allowed
			originAllowed := false
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					originAllowed = true
					break
				}
			}
			
			if originAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// LogSecurityEvents logs security-related events
func LogSecurityEvents() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log potentially suspicious requests
			userAgent := r.Header.Get("User-Agent")
			
			// Check for common attack patterns in User-Agent
			suspiciousPatterns := []string{
				"sqlmap", "nikto", "nmap", "dirb", "gobuster",
				"<script", "javascript:", "eval(",
			}
			
			for _, pattern := range suspiciousPatterns {
				if strings.Contains(strings.ToLower(userAgent), pattern) {
					// Log suspicious request (implement proper logging)
					// log.Printf("SECURITY: Suspicious User-Agent from %s: %s", r.RemoteAddr, userAgent)
					break
				}
			}
			
			// Check for suspicious headers
			if r.Header.Get("X-Forwarded-For") != "" {
				// Log forwarded requests for audit trail
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateContentType ensures proper content types for API endpoints
func ValidateContentType() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For POST/PUT requests to API endpoints, require JSON content type
			if (r.Method == "POST" || r.Method == "PUT") && 
			   strings.HasPrefix(r.URL.Path, "/api/") &&
			   !strings.HasPrefix(r.URL.Path, "/api/auth/") { // Auth endpoints might use form data
				
				contentType := r.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
					return
				}
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
				return
			}
			
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}