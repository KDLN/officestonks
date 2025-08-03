package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"officestonks/internal/services"
)

// SSEHandler handles Server-Sent Events for real-time stock updates
type SSEHandler struct {
	marketService       *services.MarketService
	clients             map[chan []byte]bool
	clientsMutex        sync.Mutex
	noClientLogCounter  int
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(marketService *services.MarketService) *SSEHandler {
	handler := &SSEHandler{
		marketService: marketService,
		clients:       make(map[chan []byte]bool),
	}
	
	// Start listening for stock updates
	go handler.listenForStockUpdates()
	
	return handler
}

// StockUpdateMessage represents the SSE message format for stock updates
type StockUpdateMessage struct {
	Type    string  `json:"type"`
	StockID int     `json:"stock_id"`
	Symbol  string  `json:"symbol"`
	Price   float64 `json:"price"`
}

// HandleStockUpdates handles SSE connections for stock price updates
func (h *SSEHandler) HandleStockUpdates(w http.ResponseWriter, r *http.Request) {
	// Log connection attempt with more details
	origin := r.Header.Get("Origin")
	userAgent := r.Header.Get("User-Agent")
	remoteAddr := r.RemoteAddr
	log.Printf("🌟 SSE Handler: Connection attempt - Origin: %s, RemoteAddr: %s", origin, remoteAddr)
	log.Printf("🔍 SSE Handler: User-Agent: %s", userAgent)
	
	// Set SSE headers with Railway-specific settings (some may be duplicated from main.go but that's safe)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control, Accept, Accept-Encoding")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.Header().Set("Pragma", "no-cache")      // HTTP/1.0 cache control
	
	// Flush headers immediately to help with Railway proxy
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Create a channel for this client
	clientChan := make(chan []byte, 10) // Buffer to prevent blocking

	// Add client to the map
	h.clientsMutex.Lock()
	h.clients[clientChan] = true
	clientCount := len(h.clients)
	h.clientsMutex.Unlock()

	log.Printf("SSE client connected. Total clients: %d", clientCount)

	// Send initial connection confirmation immediately - CRITICAL for Railway proxy
	log.Printf("🚀 SSE Handler: Sending immediate connection confirmation...")
	initialMsg := map[string]interface{}{
		"type":      "connection",
		"status":    "connected",
		"timestamp": time.Now().Unix(),
		"message":   "SSE connection established",
		"client_id": fmt.Sprintf("client_%d", time.Now().UnixNano()),
	}
	if data, err := json.Marshal(initialMsg); err == nil {
		// Send multiple formats to ensure Railway proxy compatibility
		fmt.Fprintf(w, "data: %s\n\n", data)
		fmt.Fprintf(w, ": SSE heartbeat\n\n") // Comment line for keep-alive
		
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		log.Printf("✅ SSE Handler: Initial message sent to client successfully")
	} else {
		log.Printf("❌ SSE Handler: Error marshaling initial message: %v", err)
		return // Exit if we can't send initial message
	}

	// Send current stock prices immediately
	h.sendCurrentStockPrices(w)

	// Handle client disconnection
	defer func() {
		h.clientsMutex.Lock()
		delete(h.clients, clientChan)
		clientCount := len(h.clients)
		h.clientsMutex.Unlock()
		close(clientChan)
		log.Printf("SSE client disconnected. Total clients: %d", clientCount)
	}()

	// Set up a ticker to send periodic heartbeats (Railway-friendly interval)
	heartbeatTicker := time.NewTicker(15 * time.Second) // More frequent for Railway proxy
	defer heartbeatTicker.Stop()
	
	log.Printf("🔄 SSE Handler: Starting heartbeat timer (15s intervals)")

	// Listen for messages and client disconnection
	for {
		select {
		case data := <-clientChan:
			// Send data to client
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			heartbeat := map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Unix(),
			}
			if data, err := json.Marshal(heartbeat); err == nil {
				fmt.Fprintf(w, "data: %s\n\n", data)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}

		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// sendCurrentStockPrices sends all current stock prices to a newly connected client
func (h *SSEHandler) sendCurrentStockPrices(w http.ResponseWriter) {
	log.Printf("📊 SSE Handler: Sending current stock prices to client...")
	
	stocks, err := h.marketService.GetAllStocks()
	if err != nil {
		log.Printf("❌ SSE Handler: Error fetching stocks for initial data: %v", err)
		return
	}

	stockCount := 0
	for _, stock := range stocks {
		message := StockUpdateMessage{
			Type:    "stock_update",
			StockID: stock.ID,
			Symbol:  stock.Symbol,
			Price:   stock.CurrentPrice,
		}

		if data, err := json.Marshal(message); err == nil {
			fmt.Fprintf(w, "data: %s\n\n", data)
			stockCount++
		} else {
			log.Printf("⚠️ SSE Handler: Error marshaling stock %s: %v", stock.Symbol, err)
		}
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	
	log.Printf("✅ SSE Handler: Sent %d stock prices to client", stockCount)
}

// listenForStockUpdates listens for stock updates from the market service and broadcasts to all SSE clients
func (h *SSEHandler) listenForStockUpdates() {
	stockUpdates := h.marketService.GetSimulatorUpdates()
	
	log.Printf("📡 SSEHandler: Starting to listen for stock updates from market simulator")
	updatesReceived := 0
	lastLogTime := time.Now()
	
	for update := range stockUpdates {
		updatesReceived++
		
		// Log first few updates to verify flow
		if updatesReceived <= 5 {
			log.Printf("📈 SSEHandler: Received update #%d - %s: $%.2f", updatesReceived, update.Symbol, update.Price)
		}
		
		// Validate price to prevent infinity/NaN issues
		if !isValidPrice(update.Price) {
			log.Printf("⚠️ SSEHandler: Skipping invalid price for stock %s: %f", update.Symbol, update.Price)
			continue
		}

		// Create SSE message
		message := StockUpdateMessage{
			Type:    "stock_update",
			StockID: update.StockID,
			Symbol:  update.Symbol,
			Price:   update.Price,
		}

		// Marshal to JSON
		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("❌ SSEHandler: Error marshaling SSE message: %v", err)
			continue
		}

		// Broadcast to all connected clients
		h.broadcastToClients(data)
		
		// Log progress every 60 seconds
		if time.Since(lastLogTime) >= 60*time.Second {
			h.clientsMutex.Lock()
			clientCount := len(h.clients)
			h.clientsMutex.Unlock()
			
			log.Printf("📊 SSEHandler: Processed %d updates in last 60s, broadcasting to %d clients", updatesReceived, clientCount)
			updatesReceived = 0
			lastLogTime = time.Now()
		}
	}
	
	log.Printf("🛑 SSEHandler: Stock update listener stopped")
}

// broadcastToClients sends data to all connected SSE clients
func (h *SSEHandler) broadcastToClients(data []byte) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()

	clientCount := len(h.clients)
	if clientCount == 0 {
		// Log occasionally if no clients to help debug
		if h.noClientLogCounter%100 == 0 { // Every 100 updates
			log.Printf("📭 SSEHandler: No clients connected, skipping broadcast")
		}
		h.noClientLogCounter++
		return
	}

	// Track clients to remove if they're disconnected
	var disconnectedClients []chan []byte
	successfulSends := 0

	for clientChan := range h.clients {
		select {
		case clientChan <- data:
			// Successfully sent
			successfulSends++
		default:
			// Client channel is full or closed, mark for removal
			disconnectedClients = append(disconnectedClients, clientChan)
		}
	}

	// Remove disconnected clients
	for _, clientChan := range disconnectedClients {
		delete(h.clients, clientChan)
		close(clientChan)
	}

	if len(disconnectedClients) > 0 {
		log.Printf("🔌 SSEHandler: Removed %d disconnected clients. Active clients: %d", 
			len(disconnectedClients), len(h.clients))
	}
}

// isValidPrice checks if a price is valid (not infinity, NaN, or negative)
func isValidPrice(price float64) bool {
	return !isInf(price) && !isNaN(price) && price >= 0.01
}

// isInf checks if a float64 is infinite
func isInf(f float64) bool {
	return f > 1e308 || f < -1e308
}

// isNaN checks if a float64 is NaN
func isNaN(f float64) bool {
	return f != f
}

// GetClientCount returns the number of connected SSE clients
func (h *SSEHandler) GetClientCount() int {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	return len(h.clients)
}