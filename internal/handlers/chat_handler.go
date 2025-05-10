package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"officestonks/internal/middleware"
	"officestonks/internal/services"
)

// ChatRequest represents a chat message request
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatHandler handles chat-related requests
type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// SendMessage handles sending a new chat message
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate input
	if req.Message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}
	
	// Send the message
	message, err := h.chatService.SendMessage(userID, req.Message)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}
	
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// GetRecentMessages handles retrieving recent chat messages
func (h *ChatHandler) GetRecentMessages(w http.ResponseWriter, r *http.Request) {
	// Get the limit parameter, default to 50
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default value
	
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	// Get the messages
	messages, err := h.chatService.GetRecentMessages(limit)
	if err != nil {
		http.Error(w, "Failed to retrieve messages", http.StatusInternalServerError)
		return
	}
	
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}