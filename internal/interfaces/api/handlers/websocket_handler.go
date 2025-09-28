package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// WebSocketHandler manages WebSocket connections for real-time updates
type WebSocketHandler struct {
	// Map of customer ID to their WebSocket connections
	connections map[string][]*websocket.Conn
	mutex       sync.RWMutex
	upgrader    websocket.Upgrader
}

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	ClaimID   string      `json:"claim_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		connections: make(map[string][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking for production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket for real-time updates
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// TODO: Extract customer ID from authentication
	customerID := c.Query("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "customer_id is required",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Add connection to customer's connection pool
	h.addConnection(customerID, conn)
	defer h.removeConnection(customerID, conn)

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type:      "connection_established",
		Data:      map[string]string{"status": "connected", "customer_id": customerID},
		Timestamp: time.Now(),
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Handle incoming messages and keep connection alive
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle ping/pong for keep-alive
		if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
			pongMsg := WebSocketMessage{
				Type:      "pong",
				Data:      map[string]string{"status": "alive"},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
		}
	}
}

// addConnection adds a WebSocket connection for a customer
func (h *WebSocketHandler) addConnection(customerID string, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if h.connections[customerID] == nil {
		h.connections[customerID] = make([]*websocket.Conn, 0)
	}
	h.connections[customerID] = append(h.connections[customerID], conn)
	
	log.Printf("Added WebSocket connection for customer %s. Total connections: %d", 
		customerID, len(h.connections[customerID]))
}

// removeConnection removes a WebSocket connection for a customer
func (h *WebSocketHandler) removeConnection(customerID string, conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	connections := h.connections[customerID]
	for i, c := range connections {
		if c == conn {
			// Remove connection from slice
			h.connections[customerID] = append(connections[:i], connections[i+1:]...)
			break
		}
	}
	
	// Clean up empty connection lists
	if len(h.connections[customerID]) == 0 {
		delete(h.connections, customerID)
	}
	
	log.Printf("Removed WebSocket connection for customer %s. Remaining connections: %d", 
		customerID, len(h.connections[customerID]))
}

// BroadcastClaimUpdate sends a claim update to all connected customers for that claim
func (h *WebSocketHandler) BroadcastClaimUpdate(customerID string, update dto.CustomerClaimUpdate) {
	h.mutex.RLock()
	connections := h.connections[customerID]
	h.mutex.RUnlock()
	
	if len(connections) == 0 {
		return
	}
	
	message := WebSocketMessage{
		Type:      "claim_update",
		ClaimID:   update.ClaimID,
		Data:      update,
		Timestamp: time.Now(),
	}
	
	// Send to all connections for this customer
	for _, conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to send claim update to customer %s: %v", customerID, err)
			// Connection is likely dead, it will be cleaned up when the read loop exits
		}
	}
	
	log.Printf("Broadcasted claim update for customer %s to %d connections", 
		customerID, len(connections))
}

// BroadcastStatusChange sends a status change notification
func (h *WebSocketHandler) BroadcastStatusChange(customerID string, claimID string, oldStatus, newStatus string) {
	h.mutex.RLock()
	connections := h.connections[customerID]
	h.mutex.RUnlock()
	
	if len(connections) == 0 {
		return
	}
	
	message := WebSocketMessage{
		Type:    "status_change",
		ClaimID: claimID,
		Data: map[string]interface{}{
			"claim_id":       claimID,
			"previous_status": oldStatus,
			"new_status":     newStatus,
			"message":        "Your claim status has been updated",
		},
		Timestamp: time.Now(),
	}
	
	for _, conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to send status change to customer %s: %v", customerID, err)
		}
	}
	
	log.Printf("Broadcasted status change for claim %s to customer %s", claimID, customerID)
}

// BroadcastNewMessage sends a new message notification
func (h *WebSocketHandler) BroadcastNewMessage(customerID string, claimID string, message string) {
	h.mutex.RLock()
	connections := h.connections[customerID]
	h.mutex.RUnlock()
	
	if len(connections) == 0 {
		return
	}
	
	wsMessage := WebSocketMessage{
		Type:    "new_message",
		ClaimID: claimID,
		Data: map[string]interface{}{
			"claim_id": claimID,
			"message":  message,
			"title":    "New message from support team",
		},
		Timestamp: time.Now(),
	}
	
	for _, conn := range connections {
		if err := conn.WriteJSON(wsMessage); err != nil {
			log.Printf("Failed to send new message to customer %s: %v", customerID, err)
		}
	}
	
	log.Printf("Broadcasted new message for claim %s to customer %s", claimID, customerID)
}

// GetConnectionCount returns the number of active connections for a customer
func (h *WebSocketHandler) GetConnectionCount(customerID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return len(h.connections[customerID])
}

// GetTotalConnections returns the total number of active WebSocket connections
func (h *WebSocketHandler) GetTotalConnections() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	total := 0
	for _, connections := range h.connections {
		total += len(connections)
	}
	return total
}