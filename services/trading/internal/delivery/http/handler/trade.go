package handler

import (
	"net/http"
	"strconv"

	"trading/internal/delivery/http/middleware"
	"trading/internal/domain"
)

type TradeHandler struct {
	tradeRepo domain.TradeRepository
}

func NewTradeHandler(tradeRepo domain.TradeRepository) *TradeHandler {
	return &TradeHandler{tradeRepo: tradeRepo}
}

type TradeResponse struct {
	ID         int64  `json:"id"`
	PositionID int64  `json:"position_id"`
	OrderID    int64  `json:"order_id"`
	Symbol     string `json:"symbol"`
	Side       string `json:"side"`
	Type       string `json:"type"`
	Quantity   string `json:"quantity"`
	Price      string `json:"price"`
	PnL        string `json:"pnl"`
	Fee        string `json:"fee"`
	CreatedAt  string `json:"created_at"`
}

func (h *TradeHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	trades, err := h.tradeRepo.GetByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, "failed to get trades", http.StatusInternalServerError)
		return
	}

	response := make([]TradeResponse, len(trades))
	for i, t := range trades {
		response[i] = TradeResponse{
			ID:         int64(t.ID),
			PositionID: int64(t.PositionID),
			OrderID:    int64(t.OrderID),
			Symbol:     t.Symbol,
			Side:       string(t.Side),
			Type:       string(t.Type),
			Quantity:   t.Quantity.String(),
			Price:      t.Price.String(),
			PnL:        t.PnL.String(),
			Fee:        t.Fee.String(),
			CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, response, http.StatusOK)
}
