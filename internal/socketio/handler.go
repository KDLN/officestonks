package socketio

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"officestonks/internal/auth"
	"officestonks/pkg/market"
)

// SocketIOHandler handles Socket.IO v4 protocol connections
type SocketIOHandler struct {
	upgrader         websocket.Upgrader
	clients          map[string]*Client
	clientsMu        sync.RWMutex
	stockUpdates     <-chan market.StockUpdate
	tokenValidator   auth.TokenValidator
	monitoringService interface {
		TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
		RemoveWebSocketConnection(connectionID string)
	}
}

// Client represents a connected Socket.IO client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	UserID   int
	Username string
	Send     chan []byte
	Rooms    map[string]bool
}

// Message types for Socket.IO protocol
const (
	MessageTypeOpen    = "0"
	MessageTypeClose   = "1"
	MessageTypePing    = "2"
	MessageTypePong    = "3"
	MessageTypeMessage = "4"
	MessageTypeUpgrade = "5"
	MessageTypeNoop    = "6"
)

// NewSocketIOHandler creates a new Socket.IO handler
func NewSocketIOHandler(stockUpdates <-chan market.StockUpdate, tokenValidator auth.TokenValidator, monitoringService interface {
	TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
	RemoveWebSocketConnection(connectionID string)
}) *SocketIOHandler {
	return &SocketIOHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for Socket.IO compatibility
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:           make(map[string]*Client),
		stockUpdates:      stockUpdates,
		tokenValidator:    tokenValidator,
		monitoringService: monitoringService,
	}
}

// ServeHTTP handles Socket.IO requests
func (h *SocketIOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract transport from query
	transport := r.URL.Query().Get("transport")
	
	// Handle based on transport type
	switch transport {
	case "websocket":
		h.handleWebSocket(w, r)
	case "polling":
		h.handlePolling(w, r)
	default:
		// Default to polling if no transport specified
		if r.Header.Get("Upgrade") == "websocket" {
			h.handleWebSocket(w, r)
		} else {
			h.handlePolling(w, r)
		}
	}
}

// handleWebSocket handles WebSocket transport
func (h *SocketIOHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get token from query
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("❌ Socket.IO: No token provided")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate token
	userID, username, err := h.validateToken(token)
	if err != nil {
		log.Printf("❌ Socket.IO: Invalid token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Socket.IO: Failed to upgrade: %v", err)
		return
	}

	// Create client
	clientID := fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())
	client := &Client{
		ID:       clientID,
		Conn:     conn,
		UserID:   userID,
		Username: username,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
	}

	// Register client
	h.clientsMu.Lock()
	h.clients[clientID] = client
	h.clientsMu.Unlock()

	// Track connection
	if h.monitoringService != nil {
		h.monitoringService.TrackWebSocketConnection(clientID, userID, username, r.RemoteAddr)
	}

	log.Printf("✅ Socket.IO WebSocket connected: %s (User: %s)", clientID, username)

	// Send handshake
	handshake := map[string]interface{}{
		"sid":          clientID,
		"upgrades":     []string{},
		"pingInterval": 25000,
		"pingTimeout":  60000,
	}
	handshakeData, _ := json.Marshal(handshake)
	client.Send <- []byte(MessageTypeOpen + string(handshakeData))

	// Handle client
	go client.writePump()
	go client.readPump(h)
	go h.handleStockUpdates(client)
}

// handlePolling handles HTTP polling transport
func (h *SocketIOHandler) handlePolling(w http.ResponseWriter, r *http.Request) {
	// For MVP, return a simple handshake response
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "GET" {
		// Send handshake for new connections
		sid := fmt.Sprintf("polling_%d", time.Now().Unix())
		handshake := map[string]interface{}{
			"sid":          sid,
			"upgrades":     []string{"websocket"},
			"pingInterval": 25000,
			"pingTimeout":  60000,
		}
		handshakeData, _ := json.Marshal(handshake)
		response := MessageTypeOpen + string(handshakeData)
		
		// Socket.IO polling format: length:type+data
		formatted := fmt.Sprintf("%d:%s", len(response), response)
		w.Write([]byte(formatted))
	} else {
		// Handle POST (client sending data)
		w.WriteHeader(http.StatusOK)
	}
}

// validateToken validates the JWT token
func (h *SocketIOHandler) validateToken(token string) (int, string, error) {
	if h.tokenValidator == nil {
		// For testing, allow any token
		return 1, "test_user", nil
	}

	// Validate token using the auth service
	userID, err := h.tokenValidator.ValidateToken(token)
	if err != nil {
		return 0, "", err
	}

	// For now, generate username from userID
	// In a real implementation, you'd fetch this from the database
	username := fmt.Sprintf("User%d", userID)

	return userID, username, nil
}

// handleStockUpdates sends stock updates to the client
func (h *SocketIOHandler) handleStockUpdates(client *Client) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update, ok := <-h.stockUpdates:
			if !ok {
				return
			}
			
			// Format stock update as Socket.IO message
			data := map[string]interface{}{
				"type":     "stock_update",
				"stock_id": update.StockID,
				"symbol":   update.Symbol,
				"price":    update.Price,
			}
			
			message, _ := json.Marshal(data)
			// Socket.IO message format: 42["event", data]
			socketIOMessage := `42["stock_update",` + string(message) + `]`
			
			select {
			case client.Send <- []byte(MessageTypeMessage + socketIOMessage):
			default:
				// Client send channel is full, close connection
				client.Conn.Close()
				return
			}
			
		case <-ticker.C:
			// Send ping to keep connection alive
			select {
			case client.Send <- []byte(MessageTypePing):
			default:
				client.Conn.Close()
				return
			}
		}
	}
}

// Client pump methods
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(h *SocketIOHandler) {
	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, c.ID)
		h.clientsMu.Unlock()
		
		if h.monitoringService != nil {
			h.monitoringService.RemoveWebSocketConnection(c.ID)
		}
		
		c.Conn.Close()
		log.Printf("👋 Socket.IO client disconnected: %s", c.ID)
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Socket.IO read error: %v", err)
			}
			break
		}

		// Parse Socket.IO message
		msgStr := string(message)
		if len(msgStr) > 0 {
			messageType := msgStr[:1]
			
			switch messageType {
			case MessageTypePing:
				// Respond with pong
				c.Send <- []byte(MessageTypePong)
				
			case MessageTypePong:
				// Client ponged, update deadline
				c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				
			case MessageTypeMessage:
				// Handle Socket.IO events
				if strings.Contains(msgStr, "subscribe_stocks") {
					c.Rooms["stocks"] = true
					log.Printf("📊 %s subscribed to stocks", c.Username)
					
					// Send confirmation
					confirmation := `42["subscription_confirmed",{"channel":"stocks"}]`
					c.Send <- []byte(MessageTypeMessage + confirmation)
				} else if strings.Contains(msgStr, "join_chat") {
					c.Rooms["chat"] = true
					log.Printf("💬 %s joined chat", c.Username)
					
					// Send confirmation
					confirmation := `42["chat_joined",{"status":"success"}]`
					c.Send <- []byte(MessageTypeMessage + confirmation)
				}
			}
		}
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (h *SocketIOHandler) BroadcastToRoom(room string, message []byte) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for _, client := range h.clients {
		if client.Rooms[room] {
			select {
			case client.Send <- message:
			default:
				// Client send channel is full, skip
			}
		}
	}
}