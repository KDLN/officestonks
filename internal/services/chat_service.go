package services

import (
	"officestonks/internal/models"
	"officestonks/internal/websocket"
)

// ChatService handles chat-related business logic
type ChatService struct {
	chatRepo  models.ChatRepository
	userRepo  models.UserRepository
	wsHub     *websocket.Hub
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo models.ChatRepository,
	userRepo models.UserRepository,
	wsHub *websocket.Hub,
) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
		wsHub:    wsHub,
	}
}

// SendMessage sends a new chat message
func (s *ChatService) SendMessage(userID int, messageText string) (*models.ChatMessage, error) {
	// Validate message
	if messageText == "" {
		return nil, nil // Ignore empty messages
	}
	
	// Save the message to the database
	message, err := s.chatRepo.SaveMessage(userID, messageText)
	if err != nil {
		return nil, err
	}
	
	// Broadcast the message to all clients
	s.broadcastMessage(message)
	
	return message, nil
}

// GetRecentMessages gets the most recent chat messages
func (s *ChatService) GetRecentMessages(limit int) ([]*models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // Default value
	}
	
	return s.chatRepo.GetRecentMessages(limit)
}

// broadcastMessage broadcasts a chat message to all connected websocket clients
func (s *ChatService) broadcastMessage(message *models.ChatMessage) {
	// Broadcast via WebSocket
	s.wsHub.BroadcastMessage("chat_message", message)
}