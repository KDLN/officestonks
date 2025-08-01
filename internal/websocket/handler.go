package websocket

import (
	"log"
	"net/http"
	"strconv"
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
	
	// Add Railway-specific headers for WebSocket proxying
	w.Header().Set("Connection", "Upgrade")
	w.Header().Set("Upgrade", "websocket")

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

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Log the error details
		log.Printf("WebSocket upgrade failed: %v", err)
		log.Printf("Request headers: %v", r.Header)
		// Don't call http.Error after upgrade failure - connection is already invalid
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