package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RailwayCompatibleHandler provides real-time updates that work with Railway's proxy limitations
type RailwayCompatibleHandler struct {
	hub               *Hub
	wsHandler         *WebSocketHandler
	sseClients        map[string]chan []byte
	sseClientsMutex   sync.RWMutex
}

// NewRailwayCompatibleHandler creates a handler that works with Railway's proxy
func NewRailwayCompatibleHandler(hub *Hub, wsHandler *WebSocketHandler) *RailwayCompatibleHandler {
	return &RailwayCompatibleHandler{
		hub:        hub,
		wsHandler:  wsHandler,
		sseClients: make(map[string]chan []byte),
	}
}

// HandleRealTimeConnection intelligently chooses between WebSocket and SSE based on Railway detection
func (h *RailwayCompatibleHandler) HandleRealTimeConnection(w http.ResponseWriter, r *http.Request) {
	// Detect Railway environment
	isRailway := h.detectRailwayEnvironment(r)
	
	if isRailway {
		log.Printf("Railway detected - using SSE fallback for real-time updates")
		h.handleSSEConnection(w, r)
	} else {
		log.Printf("Non-Railway environment - attempting WebSocket connection")
		h.wsHandler.HandleConnection(w, r)
	}
}

// handleSSEConnection handles Server-Sent Events for Railway compatibility
func (h *RailwayCompatibleHandler) handleSSEConnection(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔗 Railway SSE connection request from %s", r.RemoteAddr)
	
	// Set SSE headers with Railway-specific optimizations
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx proxy buffering
	w.Header().Set("X-Railway-SSE", "enabled") // Railway-specific header
	
	// Extract and validate token
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
	
	// Validate token
	userID, err := h.wsHandler.tokenValidator.ValidateToken(token)
	if err != nil {
		log.Printf("SSE JWT validation error: %v", err)
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}
	
	log.Printf("SSE connection established for user %d", userID)
	
	// Create message channel for this client
	clientID := fmt.Sprintf("sse_%d_%d", userID, time.Now().UnixNano())
	messageChan := make(chan []byte, 256)
	
	// Register client
	h.sseClientsMutex.Lock()
	h.sseClients[clientID] = messageChan
	h.sseClientsMutex.Unlock()
	
	// Clean up on disconnect
	defer func() {
		h.sseClientsMutex.Lock()
		delete(h.sseClients, clientID)
		close(messageChan)
		h.sseClientsMutex.Unlock()
		log.Printf("SSE client %s disconnected", clientID)
	}()
	
	// Send initial connection message with Railway compatibility
	initialMsg := map[string]interface{}{
		"type":    "connected",
		"message": fmt.Sprintf("Connected via SSE (Railway compatible). User ID: %d", userID),
		"protocol": "SSE",
		"railway": true,
	}
	if data, err := json.Marshal(initialMsg); err == nil {
		fmt.Fprintf(w, "data: %s\r\n\r\n", data)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		log.Printf("✅ SSE initial message sent to user %d", userID)
	}
	
	// Create ticker for keep-alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	// Main event loop
	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				return
			}
			
			// Send message to client using proper SSE format
			fmt.Fprintf(w, "data: %s\r\n\r\n", message)
			
			// Flush the response writer (essential for Railway)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			} else {
				log.Printf("⚠️ SSE ResponseWriter does not support flushing")
			}
			
		case <-ticker.C:
			// Send keep-alive ping (Railway-compatible format)
			fmt.Fprintf(w, ": ping\r\n\r\n")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// BroadcastToSSEClients sends a message to all SSE clients
func (h *RailwayCompatibleHandler) BroadcastToSSEClients(message []byte) {
	h.sseClientsMutex.RLock()
	defer h.sseClientsMutex.RUnlock()
	
	for clientID, ch := range h.sseClients {
		select {
		case ch <- message:
			// Message sent successfully
		default:
			// Channel is full, skip this client
			log.Printf("SSE client %s buffer full, skipping message", clientID)
		}
	}
}

// detectRailwayEnvironment checks if running on Railway
func (h *RailwayCompatibleHandler) detectRailwayEnvironment(r *http.Request) bool {
	// Check for Railway-specific headers
	railwayHeaders := []string{
		"X-Railway-Edge",
		"X-Forwarded-Host",
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