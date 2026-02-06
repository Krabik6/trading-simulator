package handler

import (
	"net/http"

	"trading/internal/auth"
	"trading/internal/delivery/ws"
	"trading/internal/domain"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub        *ws.Hub
	jwtService *auth.JWTService
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(hub *ws.Hub, jwtService *auth.JWTService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:        hub,
		jwtService: jwtService,
	}
}

// HandleWs handles WebSocket upgrade requests
// Clients can connect with ?token=<jwt> for authenticated connections
// or without token for anonymous price updates only
func (h *WebSocketHandler) HandleWs(w http.ResponseWriter, r *http.Request) {
	var userID domain.UserID

	// Try to authenticate from query parameter
	token := r.URL.Query().Get("token")
	if token != "" {
		if id, err := h.jwtService.GetUserID(token); err == nil {
			userID = id
		}
	}

	ws.ServeWs(h.hub, w, r, userID)
}
