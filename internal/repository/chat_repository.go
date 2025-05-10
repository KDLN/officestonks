package repository

import (
	"database/sql"
	"time"

	"officestonks/internal/models"
)

// Helper function to get current time (makes testing easier)
func getNow() time.Time {
	return time.Now()
}

// ChatRepo implements the ChatRepository interface
type ChatRepo struct {
	db *sql.DB
}

// NewChatRepo creates a new chat repository
func NewChatRepo(db *sql.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// SaveMessage saves a new chat message to the database
func (r *ChatRepo) SaveMessage(userID int, message string) (*models.ChatMessage, error) {
	// SQL statement to insert a new message
	query := `
		INSERT INTO chat_messages (user_id, message)
		VALUES (?, ?)
	`
	
	// Execute the query
	result, err := r.db.Exec(query, userID, message)
	if err != nil {
		return nil, err
	}
	
	// Get the ID of the new message
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	// Get the username of the user
	var username string
	userQuery := `
		SELECT username FROM users WHERE id = ?
	`
	err = r.db.QueryRow(userQuery, userID).Scan(&username)
	if err != nil {
		return nil, err
	}
	
	// Return the new message
	return &models.ChatMessage{
		ID:        int(id),
		UserID:    userID,
		Username:  username,
		Message:   message,
		CreatedAt: getNow(), // Current time
	}, nil
}

// GetRecentMessages gets the most recent chat messages
func (r *ChatRepo) GetRecentMessages(limit int) ([]*models.ChatMessage, error) {
	query := `
		SELECT m.id, m.user_id, u.username, m.message, m.created_at
		FROM chat_messages m
		JOIN users u ON m.user_id = u.id
		ORDER BY m.created_at DESC
		LIMIT ?
	`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []*models.ChatMessage
	for rows.Next() {
		var message models.ChatMessage
		err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.Username,
			&message.Message,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	
	// Reverse the slice to get chronological order
	if len(messages) > 1 {
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}
	}
	
	return messages, nil
}
// ClearAllMessages clears all chat messages in the database
func (r *ChatRepo) ClearAllMessages() error {
	// Using a transaction for atomicity
	tx, err := r.db.Begin()
	if err \!= nil {
		return err
	}
	
	// Delete all messages
	query := `DELETE FROM chat_messages`
	
	// Execute the delete
	_, err = tx.Exec(query)
	if err \!= nil {
		tx.Rollback()
		return err
	}
	
	// Commit the transaction
	return tx.Commit()
}
EOF < /dev/null
