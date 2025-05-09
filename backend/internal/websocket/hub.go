package websocket

import (
	"sync"

	"github.com/yourusername/officestonks/pkg/market"
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
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
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
	for client := range h.clients {
		client.Send(message)
	}
	h.mu.Unlock()
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
	for client := range h.clients {
		client.Send(message)
	}
	h.mu.Unlock()
}