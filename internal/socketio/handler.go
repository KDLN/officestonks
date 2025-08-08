package socketio

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	nhooyrws "nhooyr.io/websocket"
	"officestonks/internal/auth"
	"officestonks/pkg/market"
)

// SocketIOHandler handles Socket.IO v4 protocol connections
type SocketIOHandler struct {
	upgrader         websocket.Upgrader
	clients          map[string]*Client
	clientsMu        sync.RWMutex
	pollingSessions  map[string]*PollingSession
	pollingMu        sync.RWMutex
	stockUpdates     <-chan market.StockUpdate
	tokenValidator   auth.TokenValidator
	monitoringService interface {
		TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
		RemoveWebSocketConnection(connectionID string)
	}
}

// PollingSession represents a polling transport session
type PollingSession struct {
	ID           string
	UserID       int
	Username     string
	CreatedAt    time.Time
	LastActivity time.Time
	MessageQueue []string
	mu           sync.Mutex
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
	handler := &SocketIOHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for Socket.IO compatibility
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:           make(map[string]*Client),
		pollingSessions:   make(map[string]*PollingSession),
		stockUpdates:      stockUpdates,
		tokenValidator:    tokenValidator,
		monitoringService: monitoringService,
	}
	
	// Start cleanup routine for expired polling sessions
	go handler.cleanupExpiredSessions()
	
	return handler
}

// ServeHTTP handles Socket.IO requests
func (h *SocketIOHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Debug logging
	log.Printf("🔍 Socket.IO Request: %s %s", r.Method, r.URL.String())
	log.Printf("🔍 Headers: %v", r.Header)
	log.Printf("🔍 Query params: %v", r.URL.Query())
	
	// Extract transport from query
	transport := r.URL.Query().Get("transport")
	log.Printf("🔍 Transport: '%s'", transport)
	
	// Handle based on transport type
	switch transport {
	case "websocket":
		log.Printf("🔄 Routing to WebSocket handler")
		h.handleWebSocket(w, r)
	case "polling":
		log.Printf("🔄 Routing to Polling handler")
		h.handlePolling(w, r)
	default:
		// Default to polling if no transport specified
		if r.Header.Get("Upgrade") == "websocket" {
			log.Printf("🔄 No transport specified, but has WebSocket upgrade header - routing to WebSocket")
			h.handleWebSocket(w, r)
		} else {
			log.Printf("🔄 No transport specified, defaulting to polling")
			h.handlePolling(w, r)
		}
	}
}

// handleWebSocket handles WebSocket transport using ResponseController
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

	log.Printf("🔄 Socket.IO: Attempting WebSocket upgrade using ResponseController")

	// Try ResponseController approach (Go 1.20+) for Railway compatibility
	ctrl := http.NewResponseController(w)
	
	// First try using gorilla/websocket directly - it might use ResponseController internally
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Socket.IO: gorilla/websocket upgrade failed: %v", err)
		
		// Try nhooyr.io/websocket as alternative (designed for cloud platforms)
		log.Printf("🔄 Socket.IO: Trying nhooyr.io/websocket library")
		nhooyrConn, nhooyrErr := nhooyrws.Accept(w, r, &nhooyrws.AcceptOptions{
			OriginPatterns: []string{"*"}, // Allow all origins
		})
		if nhooyrErr != nil {
			log.Printf("❌ Socket.IO: nhooyr.io/websocket upgrade failed: %v", nhooyrErr)
			
			// Try manual hijacking using ResponseController as last resort
			netConn, bufrw, hijackErr := ctrl.Hijack()
			if hijackErr != nil {
				log.Printf("❌ Socket.IO: ResponseController hijack failed: %v", hijackErr)
				log.Printf("⚠️ Socket.IO: WebSocket not available, client should use polling")
				h.handleWebSocketUpgradeFailure(w, r)
				return
			}
			
			log.Printf("✅ Socket.IO: Successfully hijacked connection using ResponseController")
			defer netConn.Close()
			
			// Handle the WebSocket upgrade manually
			h.handleManualWebSocketUpgrade(netConn, bufrw, r, userID, username)
			return
		}
		
		log.Printf("✅ Socket.IO: nhooyr.io/websocket upgrade successful!")
		
		// Handle nhooyr WebSocket connection
		h.handleNhooyrWebSocketConnection(nhooyrConn, r, userID, username)
		return
	}

	log.Printf("✅ Socket.IO: WebSocket upgrade successful!")

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
	// Get token from query
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("❌ Socket.IO Polling: No token provided")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate token
	userID, username, err := h.validateToken(token)
	if err != nil {
		log.Printf("❌ Socket.IO Polling: Invalid token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get or create session
	sid := r.URL.Query().Get("sid")
	
	log.Printf("🔍 Session ID from query: '%s'", sid)
	
	if r.Method == "GET" {
		if sid == "" {
			// Initial handshake - create new session
			log.Printf("📡 Socket.IO Polling: Initial handshake for %s (no sid)", username)
			h.handlePollingHandshake(w, r, userID, username)
		} else {
			// Existing session - send pending messages
			log.Printf("📡 Socket.IO Polling: GET request for existing session %s", sid)
			h.handlePollingGet(w, r, sid, userID, username)
		}
	} else if r.Method == "POST" {
		// Client sending data
		log.Printf("📡 Socket.IO Polling: POST request for session %s", sid)
		h.handlePollingPost(w, r, sid, userID, username)
	}
}

// handlePollingHandshake handles initial Socket.IO handshake
func (h *SocketIOHandler) handlePollingHandshake(w http.ResponseWriter, r *http.Request, userID int, username string) {
	// Create session ID
	sid := fmt.Sprintf("polling_%d_%d", userID, time.Now().UnixNano())
	
	// Create polling session
	session := &PollingSession{
		ID:           sid,
		UserID:       userID,
		Username:     username,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		MessageQueue: make([]string, 0),
	}
	
	// Store session
	h.pollingMu.Lock()
	h.pollingSessions[sid] = session
	h.pollingMu.Unlock()
	
	// Create handshake response
	handshake := map[string]interface{}{
		"sid":          sid,
		"upgrades":     []string{}, // No WebSocket upgrade available
		"pingInterval": 25000,
		"pingTimeout":  60000,
		"maxPayload":   1000000,   // Add maxPayload for Socket.IO v4
	}
	handshakeData, _ := json.Marshal(handshake)
	openMessage := MessageTypeOpen + string(handshakeData)
	
	// Socket.IO v4 uses record separator (\x1e) between packets, not length:message
	response := openMessage + "\x1e"
	
	log.Printf("✅ Socket.IO Polling: Created session %s for %s", sid, username)
	
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// handlePollingGet handles polling GET requests (client requesting messages)
func (h *SocketIOHandler) handlePollingGet(w http.ResponseWriter, r *http.Request, sid string, userID int, username string) {
	// Find session
	h.pollingMu.RLock()
	session, exists := h.pollingSessions[sid]
	h.pollingMu.RUnlock()
	
	if !exists {
		log.Printf("❌ Socket.IO Polling: Session %s not found for %s", sid, username)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	
	// Update last activity
	session.mu.Lock()
	session.LastActivity = time.Now()
	
	// Get queued messages
	messages := make([]string, len(session.MessageQueue))
	copy(messages, session.MessageQueue)
	session.MessageQueue = session.MessageQueue[:0] // Clear queue
	session.mu.Unlock()
	
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	
	var response string
	
	// If no messages, just send heartbeat (don't send stock updates until client connects)
	if len(messages) == 0 {
		// Send a simple pong message to keep connection alive
		pongMessage := MessageTypePong
		response = pongMessage + "\x1e"
		
		log.Printf("📊 Socket.IO Polling: Sent heartbeat to %s (session: %s)", username, sid)
	} else {
		// Send queued messages
		for _, msg := range messages {
			response += msg + "\x1e"
		}
		log.Printf("📊 Socket.IO Polling: Sent %d queued messages to %s (session: %s)", len(messages), username, sid)
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

// handlePollingPost handles polling POST requests (client sending messages)
func (h *SocketIOHandler) handlePollingPost(w http.ResponseWriter, r *http.Request, sid string, userID int, username string) {
	// Find session
	h.pollingMu.RLock()
	session, exists := h.pollingSessions[sid]
	h.pollingMu.RUnlock()
	
	if !exists {
		log.Printf("❌ Socket.IO Polling: Session %s not found for POST from %s", sid, username)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	
	// Update last activity
	session.mu.Lock()
	session.LastActivity = time.Now()
	session.mu.Unlock()
	
	// Read the request body
	body := make([]byte, r.ContentLength)
	r.Body.Read(body)
	
	log.Printf("📨 Socket.IO Polling: Received from %s (session: %s): %s", username, sid, string(body))
	
	// Parse Socket.IO polling format and handle messages
	messageStr := string(body)
	
	// Handle different message types
	if strings.Contains(messageStr, "subscribe_stocks") {
		log.Printf("📊 %s subscribed to stocks via polling (session: %s)", username, sid)
		// Queue confirmation message
		h.queueMessage(sid, `42["subscription_confirmed",{"channel":"stocks"}]`)
	} else if strings.Contains(messageStr, "join_chat") {
		log.Printf("💬 %s joined chat via polling (session: %s)", username, sid)
		// Queue confirmation message
		h.queueMessage(sid, `42["chat_joined",{"status":"success"}]`)
	} else if strings.Contains(messageStr, MessageTypePing) {
		log.Printf("🏓 Received ping from %s via polling (session: %s)", username, sid)
		// Queue pong response
		h.queueMessage(sid, MessageTypePong)
	} else if strings.HasPrefix(messageStr, "40") {
		// Socket.IO connect packet - client is connecting to Socket.IO namespace
		log.Printf("🔗 %s Socket.IO connected via polling (session: %s)", username, sid)
		// Send connect acknowledgment (packet type 4, message type 0)
		h.queueMessage(sid, "40") // Connect acknowledgment for default namespace
		// Also queue stock subscription confirmation after connect
		h.queueMessage(sid, `42["subscription_confirmed",{"channel":"stocks"}]`)
	}
	
	// Send OK response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// validateToken validates the JWT token
func (h *SocketIOHandler) validateToken(token string) (int, string, error) {
	// Special debug token for testing
	if token == "test_token_123" {
		log.Printf("🔧 Using debug token for testing")
		return 999, "debug_user", nil
	}
	
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

// handleWebSocketUpgradeFailure handles failed WebSocket upgrades
func (h *SocketIOHandler) handleWebSocketUpgradeFailure(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers first
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	// For Socket.IO WebSocket upgrade failures, we need to return a specific error
	// that tells the client to fall back to polling transport
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest) // 400 - Bad Request (Socket.IO expects this)
	
	// Socket.IO expects error code 3 for transport errors
	// Format: Socket.IO packet type (4) + namespace + error packet
	errorResponse := `4{"type":"transport error","description":"WebSocket upgrade failed"}`
	
	log.Printf("🚫 Socket.IO: WebSocket upgrade failed, sending transport error to client")
	w.Write([]byte(errorResponse))
}

// handleManualWebSocketUpgrade performs WebSocket handshake manually
func (h *SocketIOHandler) handleManualWebSocketUpgrade(netConn net.Conn, bufrw *bufio.ReadWriter, r *http.Request, userID int, username string) {
	// WebSocket handshake
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		log.Printf("❌ Socket.IO: Missing Sec-WebSocket-Key")
		netConn.Close()
		return
	}

	// Compute WebSocket accept key
	acceptKey := computeWebSocketAcceptKey(key)
	
	// Write WebSocket upgrade response
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: %s\r\n" +
		"\r\n", acceptKey)
	
	if _, err := netConn.Write([]byte(response)); err != nil {
		log.Printf("❌ Socket.IO: Failed to write WebSocket upgrade response: %v", err)
		netConn.Close()
		return
	}

	log.Printf("✅ Socket.IO: Manual WebSocket handshake completed")
	
	// Create a WebSocket connection from the net.Conn
	// For now, just log success and close - full implementation would need WebSocket framing
	log.Printf("🔄 Socket.IO: Manual WebSocket connection established for user %s", username)
	
	// TODO: Implement full WebSocket message framing and handling
	// This is complex and would require implementing the full WebSocket protocol
	
	time.Sleep(1 * time.Second) // Keep connection alive briefly for testing
	netConn.Close()
}

// computeWebSocketAcceptKey computes the WebSocket accept key
func computeWebSocketAcceptKey(key string) string {
	// WebSocket magic string as defined in RFC 6455
	const webSocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	
	h := sha1.New()
	h.Write([]byte(key + webSocketMagicString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// queueMessage adds a message to a polling session's queue
func (h *SocketIOHandler) queueMessage(sessionID, message string) {
	h.pollingMu.RLock()
	session, exists := h.pollingSessions[sessionID]
	h.pollingMu.RUnlock()
	
	if !exists {
		log.Printf("❌ Cannot queue message: session %s not found", sessionID)
		return
	}
	
	session.mu.Lock()
	session.MessageQueue = append(session.MessageQueue, message)
	session.mu.Unlock()
	
	log.Printf("📮 Queued message for session %s: %s", sessionID, message)
}

// cleanupExpiredSessions removes old polling sessions
func (h *SocketIOHandler) cleanupExpiredSessions() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		expiredSessions := make([]string, 0)
		
		h.pollingMu.RLock()
		for sid, session := range h.pollingSessions {
			session.mu.Lock()
			if now.Sub(session.LastActivity) > 2*time.Minute {
				expiredSessions = append(expiredSessions, sid)
			}
			session.mu.Unlock()
		}
		h.pollingMu.RUnlock()
		
		if len(expiredSessions) > 0 {
			h.pollingMu.Lock()
			for _, sid := range expiredSessions {
				delete(h.pollingSessions, sid)
				log.Printf("🧹 Cleaned up expired polling session: %s", sid)
			}
			h.pollingMu.Unlock()
		}
	}
}

// handleNhooyrWebSocketConnection handles nhooyr.io/websocket connections
func (h *SocketIOHandler) handleNhooyrWebSocketConnection(conn *nhooyrws.Conn, r *http.Request, userID int, username string) {
	defer conn.Close(nhooyrws.StatusInternalError, "Internal server error")

	// Create client
	clientID := fmt.Sprintf("nhooyr_user_%d_%d", userID, time.Now().Unix())
	
	log.Printf("✅ Socket.IO nhooyr connection established for user %s (ID: %s)", username, clientID)

	// Send initial Socket.IO handshake
	handshake := map[string]interface{}{
		"sid":          clientID,
		"upgrades":     []string{},
		"pingInterval": 25000,
		"pingTimeout":  60000,
	}
	handshakeData, _ := json.Marshal(handshake)
	handshakeMsg := "0" + string(handshakeData) // Socket.IO message type 0 = open

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Send handshake
	if err := conn.Write(ctx, nhooyrws.MessageText, []byte(handshakeMsg)); err != nil {
		log.Printf("❌ Socket.IO: Failed to send handshake: %v", err)
		return
	}

	log.Printf("📡 Socket.IO: Sent handshake to nhooyr client %s", clientID)

	// Handle messages
	for {
		select {
		case <-ctx.Done():
			log.Printf("👋 Socket.IO: nhooyr connection timeout for %s", username)
			return
		default:
			msgType, data, err := conn.Read(ctx)
			if err != nil {
				log.Printf("👋 Socket.IO: nhooyr client disconnected %s: %v", username, err)
				return
			}

			if msgType == nhooyrws.MessageText {
				msg := string(data)
				log.Printf("📨 Socket.IO: Received from %s: %s", username, msg)

				// Handle Socket.IO protocol messages
				if len(msg) > 0 {
					switch msg[:1] {
					case "2": // Ping
						// Respond with pong (message type 3)
						if err := conn.Write(ctx, nhooyrws.MessageText, []byte("3")); err != nil {
							log.Printf("❌ Socket.IO: Failed to send pong: %v", err)
							return
						}
					case "4": // Socket.IO message
						// Handle events like stock subscription
						if strings.Contains(msg, "subscribe_stocks") {
							log.Printf("📊 %s subscribed to stocks via nhooyr", username)
							// Send confirmation
							confirmation := `42["subscription_confirmed",{"channel":"stocks"}]`
							conn.Write(ctx, nhooyrws.MessageText, []byte(confirmation))
						}
					}
				}
			}

			// Send periodic stock updates (simplified)
			go func() {
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()
				
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						// Send test stock update
						testUpdate := `42["stock_update",{"type":"stock_update","stock_id":1,"symbol":"AAPL","price":150.00}]`
						if err := conn.Write(ctx, nhooyrws.MessageText, []byte(testUpdate)); err != nil {
							return
						}
					}
				}
			}()
		}
	}
}

// GetSessionStats returns current session statistics
func (h *SocketIOHandler) GetSessionStats() map[string]interface{} {
	h.pollingMu.RLock()
	pollingSessions := len(h.pollingSessions)
	h.pollingMu.RUnlock()
	
	h.clientsMu.RLock()
	websocketClients := len(h.clients)
	h.clientsMu.RUnlock()
	
	return map[string]interface{}{
		"polling_sessions":  pollingSessions,
		"websocket_clients": websocketClients,
		"total_connections": pollingSessions + websocketClients,
	}
}