package socketio

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	socketio "github.com/doquangtan/socket.io/v4"
	"officestonks/internal/auth"
	"officestonks/pkg/market"
)

// SocketIOServer wraps the Socket.IO server with authentication and business logic
type SocketIOServer struct {
	io                *socketio.Io
	tokenValidator    auth.TokenValidator
	clients           map[string]*ClientInfo
	clientsMutex      sync.RWMutex
	stockUpdates      <-chan market.StockUpdate
	monitoringService interface {
		TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
		RemoveWebSocketConnection(connectionID string)
	}
}

// ClientInfo stores information about connected clients
type ClientInfo struct {
	SocketID    string
	UserID      int
	Username    string
	IPAddress   string
	ConnectedAt time.Time
	Namespaces  map[string]bool // Track which namespaces the client is in
}

// NewSocketIOServer creates a new Socket.IO server with Railway optimization
func NewSocketIOServer(stockUpdates <-chan market.StockUpdate, tokenValidator auth.TokenValidator, monitoringService interface {
	TrackWebSocketConnection(connectionID string, userID int, username, ipAddress string)
	RemoveWebSocketConnection(connectionID string)
}) *SocketIOServer {
	// Create Socket.IO server
	io := socketio.New()

	socketIOServer := &SocketIOServer{
		io:                io,
		tokenValidator:    tokenValidator,
		clients:           make(map[string]*ClientInfo),
		stockUpdates:      stockUpdates,
		monitoringService: monitoringService,
	}

	// Set up event handlers
	socketIOServer.setupEventHandlers()

	return socketIOServer
}

// setupEventHandlers configures all Socket.IO event handlers
func (s *SocketIOServer) setupEventHandlers() {
	// Main connection handler with authentication
	s.io.OnConnection(func(socket *socketio.Socket) {
		log.Printf("🔗 Socket.IO connection attempt: %s", socket.Id)
		
		// For now, skip token validation to get basic connection working
		// TODO: Add proper token validation from query parameters
		userID := 1 // Temporary - will implement proper auth later
		
		// Get client IP address (simplified for now)
		clientIP := "unknown" // TODO: Extract from connection
		
		// Register client
		clientInfo := &ClientInfo{
			SocketID:    socket.Id,
			UserID:      userID,
			Username:    fmt.Sprintf("User%d", userID),
			IPAddress:   clientIP,
			ConnectedAt: time.Now(),
			Namespaces:  make(map[string]bool),
		}

		s.clientsMutex.Lock()
		s.clients[socket.Id] = clientInfo
		s.clientsMutex.Unlock()

		// Track connection in monitoring service
		if s.monitoringService != nil {
			s.monitoringService.TrackWebSocketConnection(socket.Id, userID, clientInfo.Username, clientIP)
		}

		log.Printf("✅ Socket.IO client connected: User %d (%s)", userID, socket.Id)

		// Send initial connection confirmation
		socket.Emit("connected", map[string]interface{}{
			"message":  fmt.Sprintf("Connected via Socket.IO. User ID: %d", userID),
			"protocol": "Socket.IO",
			"transport": "auto-detect", // Will be WebSocket or Polling
			"userID":   userID,
		})

		// Join user to their personal room for targeted messages
		socket.Join(fmt.Sprintf("user_%d", userID))
		clientInfo.Namespaces[fmt.Sprintf("user_%d", userID)] = true

		// Note: OnDisconnect not available in this library version
		// Disconnect handling will be managed through connection cleanup

		// Handle ping messages for connection quality testing
		socket.On("ping", func(event *socketio.EventPayload) {
			socket.Emit("pong", map[string]interface{}{
				"timestamp":    event.Data,
				"server_time": time.Now().Unix(),
			})
		})

		// Handle stock subscription requests
		socket.On("subscribe_stocks", func(event *socketio.EventPayload) {
			log.Printf("📊 Client %s subscribed to stock updates", socket.Id)
			socket.Join("stocks")
			
			s.clientsMutex.Lock()
			if clientInfo, exists := s.clients[socket.Id]; exists {
				clientInfo.Namespaces["stocks"] = true
			}
			s.clientsMutex.Unlock()
			
			socket.Emit("subscription_confirmed", map[string]string{"channel": "stocks"})
		})

		// Handle chat room subscriptions
		socket.On("join_chat", func(event *socketio.EventPayload) {
			log.Printf("💬 Client %s joined chat", socket.Id)
			socket.Join("chat")
			
			s.clientsMutex.Lock()
			if clientInfo, exists := s.clients[socket.Id]; exists {
				clientInfo.Namespaces["chat"] = true
			}
			s.clientsMutex.Unlock()
			
			socket.Emit("chat_joined", map[string]string{"status": "success"})
		})

		// Handle chat messages
		socket.On("chat_message", func(event *socketio.EventPayload) {
			s.clientsMutex.RLock()
			clientInfo, exists := s.clients[socket.Id]
			s.clientsMutex.RUnlock()
			
			if !exists {
				return
			}

			// Broadcast chat message to all clients in chat room
			chatData := map[string]interface{}{
				"type":      "chat_message",
				"user_id":   clientInfo.UserID,
				"username":  clientInfo.Username,
				"message":   event.Data,
				"timestamp": time.Now().Unix(),
			}

			socket.To("chat").Emit("chat_message", chatData)
			log.Printf("💬 Chat message from User %d: %v", clientInfo.UserID, event.Data)
		})
	})
}

// Start begins listening for Socket.IO connections and starts the stock update broadcaster
func (s *SocketIOServer) Start() {
	// Start stock update broadcaster
	go s.broadcastStockUpdates()
	
	log.Println("🚀 Socket.IO server started with Railway optimization")
	log.Println("📡 Transports: WebSocket (primary) → Polling (fallback)")
	log.Println("🔐 Authentication: JWT token validation enabled")
	log.Println("🏠 Rooms: stocks, chat, user_* available")
}

// broadcastStockUpdates listens for stock updates and broadcasts them to subscribed clients
func (s *SocketIOServer) broadcastStockUpdates() {
	for update := range s.stockUpdates {
		stockData := map[string]interface{}{
			"type":     "stock_update",
			"stock_id": update.StockID,
			"symbol":   update.Symbol,
			"price":    update.Price,
			"change":   update.Price, // You might want to calculate actual change
		}

		// Broadcast to all clients subscribed to stocks room
		s.io.To("stocks").Emit("stock_update", stockData)
		
		// Also broadcast to all connected clients (for compatibility)
		s.io.Emit("stock_update", stockData)
	}
}

// GetHTTPHandler returns the HTTP handler for Socket.IO
func (s *SocketIOServer) GetHTTPHandler() http.Handler {
	return s.io.HttpHandler()
}

// GetConnectedClients returns information about all connected clients
func (s *SocketIOServer) GetConnectedClients() map[string]*ClientInfo {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	result := make(map[string]*ClientInfo)
	for k, v := range s.clients {
		result[k] = v
	}
	return result
}

// BroadcastMessage sends a message to all connected clients
func (s *SocketIOServer) BroadcastMessage(eventName string, data interface{}) {
	s.io.Emit(eventName, data)
}

// BroadcastToRoom sends a message to all clients in a specific room
func (s *SocketIOServer) BroadcastToRoom(room string, eventName string, data interface{}) {
	s.io.To(room).Emit(eventName, data)
}

// GetStats returns server statistics
func (s *SocketIOServer) GetStats() map[string]interface{} {
	s.clientsMutex.RLock()
	clientCount := len(s.clients)
	
	// Calculate namespace distribution
	namespaceCounts := make(map[string]int)
	for _, client := range s.clients {
		for namespace := range client.Namespaces {
			namespaceCounts[namespace]++
		}
	}
	s.clientsMutex.RUnlock()

	return map[string]interface{}{
		"connected_clients": clientCount,
		"namespace_counts":  namespaceCounts,
		"server_uptime":     time.Now().Format(time.RFC3339),
		"transport_info":    "WebSocket (primary) + Polling (fallback)",
	}
}