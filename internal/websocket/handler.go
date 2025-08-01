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

// Upgrader upgrades HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development and production
	CheckOrigin: func(r *http.Request) bool {
		// Log the origin attempting to connect
		origin := r.Header.Get("Origin")
		log.Printf("WebSocket CheckOrigin function called with origin: %s", origin)

		// Always return true to accept all origins
		return true
	},
	// Add error handling for Railway proxy issues
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Printf("WebSocket upgrade error: status=%d, reason=%v", status, reason)
		http.Error(w, reason.Error(), status)
	},
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

// HandleConnection handles a new websocket connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Log request details for debugging
	log.Printf("WebSocket request received: %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))
	
	// Detect Railway deployment
	isRailway := r.Header.Get("X-Railway-Edge") != "" || r.Header.Get("X-Forwarded-Host") != ""
	if isRailway {
		log.Printf("Railway deployment detected, applying Railway-specific WebSocket handling")
	}
	
	// Set CORS headers for WebSocket handshake
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
	
	// Railway-specific headers - these MUST be set before any writes
	if isRailway {
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Sec-WebSocket-Accept", "")  // Let Gorilla handle this
	}

	// Log the origin for debugging
	log.Printf("WebSocket connection attempted from origin: %s", origin)

	// Handle preflight requests for WebSockets
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract token from query parameter
	token := r.URL.Query().Get("token")
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

	// Railway-specific WebSocket upgrade handling
	var conn *websocket.Conn
	
	if isRailway {
		// For Railway, use a custom upgrader with different settings
		railwayUpgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Railway already handles CORS
			},
			// Don't set Error handler for Railway - let it fail gracefully
		}
		
		// Set additional Railway headers right before upgrade
		w.Header().Set("X-Railway-WebSocket", "upgrade")
		
		conn, err = railwayUpgrader.Upgrade(w, r, nil)
	} else {
		// Use standard upgrader for non-Railway deployments
		conn, err = upgrader.Upgrade(w, r, nil)
	}
	
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