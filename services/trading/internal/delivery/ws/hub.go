package ws

import (
	"encoding/json"
	"sync"
	"time"

	"trading/internal/domain"
	"trading/internal/logger"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypePrices        MessageType = "prices"
	MessageTypePosition      MessageType = "position"
	MessageTypePositionClose MessageType = "position_close"
	MessageTypeTrade         MessageType = "trade"
	MessageTypeError         MessageType = "error"
	MessageTypePing          MessageType = "ping"
	MessageTypePong          MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PriceUpdate represents a price update message
type PriceUpdate struct {
	Symbol string  `json:"symbol"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
	Mid    float64 `json:"mid"`
}

// PositionUpdate represents a position update message
type PositionUpdate struct {
	ID            int64  `json:"id"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Quantity      string `json:"quantity"`
	EntryPrice    string `json:"entry_price"`
	MarkPrice     string `json:"mark_price"`
	UnrealizedPnL string `json:"unrealized_pnl"`
	Leverage      int    `json:"leverage"`
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[domain.UserID]map[*Client]bool

	// All clients for broadcast (e.g., prices)
	allClients map[*Client]bool

	// Register requests
	register chan *Client

	// Unregister requests
	unregister chan *Client

	// Broadcast to all clients
	broadcast chan []byte

	// Broadcast to specific user
	userBroadcast chan userMessage

	mu sync.RWMutex
}

type userMessage struct {
	userID  domain.UserID
	message []byte
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[domain.UserID]map[*Client]bool),
		allClients:    make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte, 256),
		userBroadcast: make(chan userMessage, 256),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.allClients[client] = true
			if client.userID != 0 {
				if h.clients[client.userID] == nil {
					h.clients[client.userID] = make(map[*Client]bool)
				}
				h.clients[client.userID][client] = true
			}
			h.mu.Unlock()
			logger.Info("websocket client connected", "user_id", client.userID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.allClients[client]; ok {
				delete(h.allClients, client)
				if client.userID != 0 {
					delete(h.clients[client.userID], client)
					if len(h.clients[client.userID]) == 0 {
						delete(h.clients, client.userID)
					}
				}
				close(client.send)
			}
			h.mu.Unlock()
			logger.Info("websocket client disconnected", "user_id", client.userID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.allClients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.allClients, client)
				}
			}
			h.mu.RUnlock()

		case msg := <-h.userBroadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.userID]; ok {
				for client := range clients {
					select {
					case client.send <- msg.message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastPrices broadcasts price updates to all connected clients
func (h *Hub) BroadcastPrices(prices map[string]*domain.Price) {
	updates := make([]PriceUpdate, 0, len(prices))
	for _, p := range prices {
		updates = append(updates, PriceUpdate{
			Symbol: p.Symbol,
			Bid:    p.Bid,
			Ask:    p.Ask,
			Mid:    p.Mid(),
		})
	}

	msg := Message{
		Type:      MessageTypePrices,
		Data:      updates,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal price update", "error", err)
		return
	}

	select {
	case h.broadcast <- data:
	default:
		logger.Warn("broadcast channel full, dropping price update")
	}
}

// BroadcastPositionUpdate broadcasts position update to specific user
func (h *Hub) BroadcastPositionUpdate(userID domain.UserID, position *domain.Position) {
	update := PositionUpdate{
		ID:            int64(position.ID),
		Symbol:        position.Symbol,
		Side:          string(position.Side),
		Quantity:      position.Quantity.String(),
		EntryPrice:    position.EntryPrice.String(),
		MarkPrice:     position.MarkPrice.String(),
		UnrealizedPnL: position.UnrealizedPnL.String(),
		Leverage:      position.Leverage,
	}

	msg := Message{
		Type:      MessageTypePosition,
		Data:      update,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal position update", "error", err)
		return
	}

	select {
	case h.userBroadcast <- userMessage{userID: userID, message: data}:
	default:
		logger.Warn("user broadcast channel full", "user_id", userID)
	}
}

// BroadcastPositionClose broadcasts position close event to specific user
func (h *Hub) BroadcastPositionClose(userID domain.UserID, positionID domain.PositionID, pnl string) {
	msg := Message{
		Type: MessageTypePositionClose,
		Data: map[string]interface{}{
			"position_id":  int64(positionID),
			"realized_pnl": pnl,
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal position close", "error", err)
		return
	}

	select {
	case h.userBroadcast <- userMessage{userID: userID, message: data}:
	default:
		logger.Warn("user broadcast channel full", "user_id", userID)
	}
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.allClients)
}
