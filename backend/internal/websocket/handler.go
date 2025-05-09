package websocket

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/yourusername/officestonks/internal/auth"
	"github.com/yourusername/officestonks/pkg/market"
)

// WebSocketHandler handles websocket connections
type WebSocketHandler struct {
	hub *Hub
}

// Upgrader upgrades HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewWebSocketHandler creates a new websocket handler
func NewWebSocketHandler(stockUpdates <-chan market.StockUpdate) *WebSocketHandler {
	hub := NewHub(stockUpdates)
	go hub.Run()
	
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleConnection handles a new websocket connection
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}
	
	// Validate token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}
	
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}
	
	// Create a new client
	client := NewClient(h.hub, conn, claims.UserID)
	
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