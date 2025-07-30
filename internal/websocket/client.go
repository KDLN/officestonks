package websocket

import (
	"encoding/json"
	"log"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a connected websocket client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	// User ID for authentication (would be extracted from token)
	userID int
}

// NewClient creates a new websocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID int) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		
		// Process incoming messages from client if needed
		// For now we're just handling server -> client communication
		_ = message
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// IMPORTANT: Send each message individually to prevent JSON parsing issues
			// DO NOT batch multiple JSON objects together
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			
			// Write a single message
			w.Write(message)
			
			// Close this message
			if err := w.Close(); err != nil {
				return
			}
			
			// Process any queued messages - each in its own write operation
			n := len(c.send)
			for i := 0; i < n; i++ {
				// Get next message
				queuedMessage := <-c.send
				
				// Create a new writer for each message
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				
				// Write the message
				w.Write(queuedMessage)
				
				// Close this message
				if err := w.Close(); err != nil {
					return
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sanitizeMessage recursively removes infinity and NaN values from a message
func sanitizeMessage(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		f := val.Float()
		if math.IsInf(f, 0) || math.IsNaN(f) {
			log.Printf("🚨 SANITIZING INVALID FLOAT: %f -> 0.01", f)
			return 0.01 // Replace with a safe default
		}
		return f
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = sanitizeMessage(val.Index(i).Interface())
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			keyStr := key.String()
			result[keyStr] = sanitizeMessage(val.MapIndex(key).Interface())
		}
		return result
	case reflect.Struct:
		// For structs, create a map representation
		result := make(map[string]interface{})
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			if field.IsExported() {
				fieldValue := val.Field(i)
				if fieldValue.CanInterface() {
					// Use JSON tag name if available, otherwise use field name
					name := field.Name
					if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
						// Handle comma-separated options like "price,omitempty"
						if commaIndex := strings.Index(tag, ","); commaIndex != -1 {
							name = tag[:commaIndex]
						} else {
							name = tag
						}
					}
					result[name] = sanitizeMessage(fieldValue.Interface())
				}
			}
		}
		return result
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return sanitizeMessage(val.Elem().Interface())
	default:
		return v
	}
}

// Send sends a message to the client
func (c *Client) Send(message interface{}) {
	// Sanitize the message to remove infinity and NaN values
	sanitizedMessage := sanitizeMessage(message)
	
	// Convert message to JSON
	jsonMessage, err := json.Marshal(sanitizedMessage)
	if err != nil {
		log.Printf("🚨 JSON MARSHAL ERROR after sanitization for user %d: %v", c.userID, err)
		log.Printf("🚨 Original message type: %T", message)
		log.Printf("🚨 Original message: %+v", message)
		log.Printf("🚨 Sanitized message: %+v", sanitizedMessage)
		return
	}
	
	// Send to client's channel with non-blocking behavior
	select {
	case c.send <- jsonMessage:
		// Message sent successfully
	default:
		// Channel buffer is full, log and drop message
		log.Printf("Client send buffer full, dropping message for user %d", c.userID)
	}
}