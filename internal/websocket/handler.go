package websocket

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"officestonks/internal/auth"
)

// WebSocketHandler handles websocket connections
type WebSocketHandler struct {
	hub               *Hub
	tokenValidator    auth.TokenValidator
	monitoringService interface {
		TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
		RemoveWebSocketConnection(connectionID string)
	}
}

// Railway-specific WebSocket upgrader with enhanced compatibility
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096, // Larger buffer for Railway proxy
	WriteBufferSize: 4096, // Larger buffer for Railway proxy
	// Allow all origins - Railway doesn't preserve origin headers consistently
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		log.Printf("WebSocket CheckOrigin called with origin: %s", origin)
		
		// Railway proxy may strip or modify origin headers
		// Always allow for production compatibility
		return true
	},
	// Enhanced error handler for Railway proxy debugging
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		isRailway := r.Header.Get("X-Railway-Edge") != "" || r.Header.Get("X-Forwarded-Host") != ""
		log.Printf("WebSocket upgrade error (Railway: %t): status=%d, reason=%v", isRailway, status, reason)
		
		// For Railway, log all headers to debug proxy issues
		if isRailway {
			log.Printf("Railway headers: Connection=%s, Upgrade=%s, Sec-WebSocket-Key=%s",
				r.Header.Get("Connection"), r.Header.Get("Upgrade"), r.Header.Get("Sec-WebSocket-Key"))
		}
		
		// Don't call http.Error if we suspect connection state issues
		if !strings.Contains(reason.Error(), "hijacker") {
			http.Error(w, reason.Error(), status)
		}
	},
	// Longer timeout for Railway's proxy with cold starts
	HandshakeTimeout: 60 * time.Second,
	// Enable compression to work better with Railway proxy
	EnableCompression: true,
}

// NewWebSocketHandler creates a new websocket handler
func NewWebSocketHandler(hub *Hub, tokenValidator auth.TokenValidator, monitoringService interface {
	TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
	RemoveWebSocketConnection(connectionID string)
}) *WebSocketHandler {
	return &WebSocketHandler{
		hub:               hub,
		tokenValidator:    tokenValidator,
		monitoringService: monitoringService,
	}
}

// HandleConnection handles a new websocket connection with Railway proxy fixes
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Enhanced logging for Railway debugging
	log.Printf("WebSocket request: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

	// Detect Railway deployment with multiple indicators
	isRailway := h.detectRailwayEnvironment(r)
	if isRailway {
		log.Printf("Railway environment detected - applying proxy-compatible headers")
	}

	// Railway proxy fix: Set critical WebSocket headers BEFORE any response writes
	if isRailway {
		h.setRailwayWebSocketHeaders(w, r)
	}

	// Set CORS headers compatible with Railway proxy
	h.setCORSHeaders(w, r)

	// Log all incoming headers for Railway debugging
	if isRailway {
		h.logRailwayHeaders(r)
	}

	// Log the origin for debugging
	origin := r.Header.Get("Origin")
	log.Printf("WebSocket connection attempted from origin: %s", origin)

	// Handle preflight requests for WebSockets
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract token from query parameter or Authorization header
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Validate token using TokenValidator which handles both Supabase and custom tokens
	userID, err := h.tokenValidator.ValidateToken(token)
	if err != nil {
		log.Printf("JWT validation error: %v", err)
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Railway WebSocket upgrade with proxy compatibility
	log.Printf("Attempting WebSocket upgrade (Railway: %t)", isRailway)
	
	// Create Railway-compatible response headers
	var responseHeader http.Header
	if isRailway {
		responseHeader = h.createRailwayResponseHeaders(r)
	}
	
	conn, err := upgrader.Upgrade(w, r, responseHeader)

	if err != nil {
		// Log the error details with Railway context
		log.Printf("WebSocket upgrade failed (Railway: %t): %v", isRailway, err)
		log.Printf("Request headers: %v", r.Header)

		// For Railway hijacker issues, try to provide helpful error response
		if isRailway && (strings.Contains(err.Error(), "hijacker") || strings.Contains(err.Error(), "Hijacker")) {
			log.Printf("Railway hijacker interface issue detected - this may be a Railway proxy limitation")
			// Don't call http.Error as connection state is unknown
		}
		return
	}

	log.Printf("WebSocket connection successfully established for user ID: %d", userID)

	// Create a new client
	client := NewClient(h.hub, conn, userID)

	// Track the connection in monitoring service
	if h.monitoringService != nil {
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.Header.Get("X-Real-IP")
		}
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}

		// Generate a unique connection ID using timestamp
		connectionID := strconv.Itoa(userID) + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		h.monitoringService.TrackWebSocketConnection(connectionID, userID, "", clientIP)

		// Set up cleanup when connection closes
		client.connectionID = connectionID
		client.monitoringService = h.monitoringService
	}

	// Register the client
	h.hub.register <- client

	// Start the client's pumps
	go client.writePump()
	go client.readPump()

	// Send initial data to the client
	h.sendInitialData(client)
}

// GetHub returns the websocket hub
func (h *WebSocketHandler) GetHub() *Hub {
	return h.hub
}

// Railway environment detection with multiple indicators
func (h *WebSocketHandler) detectRailwayEnvironment(r *http.Request) bool {
	// Check for Railway-specific headers
	railwayHeaders := []string{
		"X-Railway-Edge",
		"X-Forwarded-Host",
		"X-Forwarded-Proto",
		"X-Railway-Request-ID",
	}
	
	for _, header := range railwayHeaders {
		if r.Header.Get(header) != "" {
			return true
		}
	}
	
	// Check host patterns
	host := r.Host
	if strings.Contains(host, "railway.app") || strings.Contains(host, "up.railway.app") {
		return true
	}
	
	return false
}

// Set Railway-compatible WebSocket headers
func (h *WebSocketHandler) setRailwayWebSocketHeaders(w http.ResponseWriter, r *http.Request) {
	// Critical: Railway proxy requires these headers to be set early
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Upgrade", "websocket")
	
	// Ensure proper protocol handling
	if r.Header.Get("Sec-WebSocket-Protocol") != "" {
		w.Header().Set("Sec-WebSocket-Protocol", r.Header.Get("Sec-WebSocket-Protocol"))
	}
	
	// Railway proxy compatibility
	w.Header().Set("Sec-WebSocket-Version", "13")
	
	log.Printf("Railway WebSocket headers set: Connection=Upgrade, Upgrade=websocket")
}

// Set CORS headers compatible with Railway proxy
func (h *WebSocketHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	
	// Railway may not preserve origin consistently
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		// Fallback for Railway proxy
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Protocol")
}

// Log Railway headers for debugging
func (h *WebSocketHandler) logRailwayHeaders(r *http.Request) {
	log.Printf("Railway WebSocket Debug Headers:")
	log.Printf("  Method: %s", r.Method)
	log.Printf("  Host: %s", r.Host)
	log.Printf("  Connection: %s", r.Header.Get("Connection"))
	log.Printf("  Upgrade: %s", r.Header.Get("Upgrade"))
	log.Printf("  Sec-WebSocket-Key: %s", r.Header.Get("Sec-WebSocket-Key"))
	log.Printf("  Sec-WebSocket-Version: %s", r.Header.Get("Sec-WebSocket-Version"))
	log.Printf("  Origin: %s", r.Header.Get("Origin"))
	log.Printf("  X-Forwarded-For: %s", r.Header.Get("X-Forwarded-For"))
	log.Printf("  X-Forwarded-Proto: %s", r.Header.Get("X-Forwarded-Proto"))
	log.Printf("  X-Railway-Edge: %s", r.Header.Get("X-Railway-Edge"))
}

// Create Railway-compatible response headers
func (h *WebSocketHandler) createRailwayResponseHeaders(r *http.Request) http.Header {
	responseHeader := make(http.Header)
	
	// Copy essential WebSocket headers for Railway proxy
	if protocol := r.Header.Get("Sec-WebSocket-Protocol"); protocol != "" {
		responseHeader.Set("Sec-WebSocket-Protocol", protocol)
	}
	
	// Ensure proper caching for Railway
	responseHeader.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	responseHeader.Set("Pragma", "no-cache")
	responseHeader.Set("Expires", "0")
	
	return responseHeader
}

// sendInitialData sends initial data to a new client
func (h *WebSocketHandler) sendInitialData(client *Client) {
	// This would be populated with real data in a full implementation
	initialData := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{
		Type:    "connected",
		Message: "Connected to Office Stonks real-time updates. User ID: " + strconv.Itoa(client.userID),
	}

	client.Send(initialData)
}
