package websocket

import (
	"log"
	"math"
	"sync"

	"officestonks/pkg/market"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Stock updates channel
	stockUpdates <-chan market.StockUpdate

	// Message queue for when no clients are connected
	messageQueue []interface{}
	maxQueueSize int

	// Mutex for thread-safe operations
	mu sync.Mutex
}

// NewHub creates a new hub
func NewHub(stockUpdates <-chan market.StockUpdate) *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		stockUpdates: stockUpdates,
		messageQueue: make([]interface{}, 0, 100),
		maxQueueSize: 100, // Keep last 100 messages
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			log.Printf("Client registered. Total clients: %d", len(h.clients))
			
			// Send queued messages to new client
			if len(h.messageQueue) > 0 {
				log.Printf("Sending %d queued messages to new client", len(h.messageQueue))
				for _, msg := range h.messageQueue {
					client.Send(msg)
				}
				// Clear queue after sending
				h.messageQueue = h.messageQueue[:0]
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered. Total clients: %d", len(h.clients))
			}
			h.mu.Unlock()

		case update := <-h.stockUpdates:
			// Broadcast stock updates to all connected clients
			h.broadcastStockUpdate(update)
		}
	}
}

// broadcastStockUpdate sends a stock update to all connected clients
func (h *Hub) broadcastStockUpdate(update market.StockUpdate) {
	// Validate price for infinity or NaN
	if math.IsInf(update.Price, 0) || math.IsNaN(update.Price) {
		log.Printf("🚨 FOUND INFINITY SOURCE: Stock %s (%d) has invalid price %f", update.Symbol, update.StockID, update.Price)
		return
	}
	
	// Ensure price is within reasonable bounds
	if update.Price < 0.01 {
		update.Price = 0.01
	} else if update.Price > 1000000 { // Cap at $1M per share
		update.Price = 1000000
	}
	
	// Create a message for the update
	message := struct {
		Type    string  `json:"type"`
		StockID int     `json:"stock_id"`
		Symbol  string  `json:"symbol"`
		Price   float64 `json:"price"`
	}{
		Type:    "stock_update",
		StockID: update.StockID,
		Symbol:  update.Symbol,
		Price:   update.Price,
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	
	// If no clients connected, queue the message
	if len(h.clients) == 0 {
		h.queueMessage(message)
		return
	}
	
	for client := range h.clients {
		client.Send(message)
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(messageType string, data interface{}) {
	// Create a message
	message := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: messageType,
		Data: data,
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	
	// If no clients connected, queue the message
	if len(h.clients) == 0 {
		h.queueMessage(message)
		return
	}
	
	for client := range h.clients {
		client.Send(message)
	}
}

// queueMessage adds a message to the queue for later delivery
func (h *Hub) queueMessage(msg interface{}) {
	// Add to queue
	h.messageQueue = append(h.messageQueue, msg)
	
	// Trim queue if it exceeds max size
	if len(h.messageQueue) > h.maxQueueSize {
		// Keep only the most recent messages
		h.messageQueue = h.messageQueue[len(h.messageQueue)-h.maxQueueSize:]
	}
	
	log.Printf("Message queued (queue size: %d)", len(h.messageQueue))
}