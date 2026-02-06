package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"trading/internal/delivery/http/middleware"
	"trading/internal/domain"
	positionuc "trading/internal/usecase/position"
)

type PositionHandler struct {
	positionUC *positionuc.UseCase
}

func NewPositionHandler(positionUC *positionuc.UseCase) *PositionHandler {
	return &PositionHandler{positionUC: positionUC}
}

type PositionResponse struct {
	ID               int64   `json:"id"`
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	Status           string  `json:"status"`
	Quantity         string  `json:"quantity"`
	EntryPrice       string  `json:"entry_price"`
	MarkPrice        string  `json:"mark_price"`
	Leverage         int     `json:"leverage"`
	InitialMargin    string  `json:"initial_margin"`
	UnrealizedPnL    string  `json:"unrealized_pnl"`
	RealizedPnL      string  `json:"realized_pnl"`
	LiquidationPrice string  `json:"liquidation_price"`
	StopLoss         *string `json:"stop_loss,omitempty"`
	TakeProfit       *string `json:"take_profit,omitempty"`
	SLClosePercent   int     `json:"sl_close_percent"`
	TPClosePercent   int     `json:"tp_close_percent"`
	CreatedAt        string  `json:"created_at"`
}

func (h *PositionHandler) GetPositions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	positions, err := h.positionUC.GetPositions(r.Context(), userID)
	if err != nil {
		writeError(w, "failed to get positions", http.StatusInternalServerError)
		return
	}

	response := make([]PositionResponse, len(positions))
	for i, p := range positions {
		response[i] = positionToResponse(&p)
	}

	writeJSON(w, response, http.StatusOK)
}

func (h *PositionHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	positionIDStr := chi.URLParam(r, "id")
	positionID, err := strconv.ParseInt(positionIDStr, 10, 64)
	if err != nil {
		writeError(w, "invalid position id", http.StatusBadRequest)
		return
	}

	position, err := h.positionUC.GetPosition(r.Context(), userID, domain.PositionID(positionID))
	if err != nil {
		if errors.Is(err, domain.ErrPositionNotFound) {
			writeError(w, "position not found", http.StatusNotFound)
			return
		}
		writeError(w, "failed to get position", http.StatusInternalServerError)
		return
	}

	writeJSON(w, positionToResponse(position), http.StatusOK)
}

func (h *PositionHandler) ClosePosition(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	positionIDStr := chi.URLParam(r, "id")
	positionID, err := strconv.ParseInt(positionIDStr, 10, 64)
	if err != nil {
		writeError(w, "invalid position id", http.StatusBadRequest)
		return
	}

	// Parse optional quantity from body
	var closeReq struct {
		Quantity *string `json:"quantity"`
	}
	// Body is optional — ignore decode errors (empty body is fine)
	_ = json.NewDecoder(r.Body).Decode(&closeReq)

	input := positionuc.ClosePositionInput{
		UserID:     userID,
		PositionID: domain.PositionID(positionID),
	}

	if closeReq.Quantity != nil {
		qty, err := decimal.NewFromString(*closeReq.Quantity)
		if err != nil || !qty.IsPositive() {
			writeError(w, "invalid quantity", http.StatusBadRequest)
			return
		}
		input.Quantity = &qty
	}

	trade, err := h.positionUC.ClosePosition(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrPositionNotFound) {
			writeError(w, "position not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrPositionNotOpen) {
			writeError(w, "position is not open", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrPriceNotAvailable) {
			writeError(w, "price not available", http.StatusServiceUnavailable)
			return
		}
		writeError(w, "failed to close position", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"realized_pnl":    trade.PnL.String(),
		"closed_quantity": trade.Quantity.String(),
	}
	if input.Quantity != nil {
		resp["status"] = "partial"
	} else {
		resp["status"] = "closed"
	}
	writeJSON(w, resp, http.StatusOK)
}

type UpdateTPSLRequest struct {
	StopLoss       *string `json:"stop_loss"`
	TakeProfit     *string `json:"take_profit"`
	SLClosePercent *int    `json:"sl_close_percent,omitempty"`
	TPClosePercent *int    `json:"tp_close_percent,omitempty"`
}

func (h *PositionHandler) UpdateTPSL(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	positionIDStr := chi.URLParam(r, "id")
	positionID, err := strconv.ParseInt(positionIDStr, 10, 64)
	if err != nil {
		writeError(w, "invalid position id", http.StatusBadRequest)
		return
	}

	var req UpdateTPSLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var stopLoss, takeProfit *decimal.Decimal
	if req.StopLoss != nil {
		sl, err := decimal.NewFromString(*req.StopLoss)
		if err != nil {
			writeError(w, "invalid stop_loss", http.StatusBadRequest)
			return
		}
		stopLoss = &sl
	}
	if req.TakeProfit != nil {
		tp, err := decimal.NewFromString(*req.TakeProfit)
		if err != nil {
			writeError(w, "invalid take_profit", http.StatusBadRequest)
			return
		}
		takeProfit = &tp
	}

	position, err := h.positionUC.UpdateTPSL(r.Context(), positionuc.UpdateTPSLInput{
		UserID:         userID,
		PositionID:     domain.PositionID(positionID),
		StopLoss:       stopLoss,
		TakeProfit:     takeProfit,
		SLClosePercent: req.SLClosePercent,
		TPClosePercent: req.TPClosePercent,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPositionNotFound) {
			writeError(w, "position not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrPositionNotOpen) {
			writeError(w, "position is not open", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidStopLoss) {
			writeError(w, "invalid stop loss value", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidTakeProfit) {
			writeError(w, "invalid take profit value", http.StatusBadRequest)
			return
		}
		if errors.Is(err, domain.ErrInvalidClosePercent) {
			writeError(w, "close percent must be between 1 and 100", http.StatusBadRequest)
			return
		}
		writeError(w, "failed to update position", http.StatusInternalServerError)
		return
	}

	writeJSON(w, positionToResponse(position), http.StatusOK)
}

func positionToResponse(p *domain.Position) PositionResponse {
	resp := PositionResponse{
		ID:               int64(p.ID),
		Symbol:           p.Symbol,
		Side:             string(p.Side),
		Status:           string(p.Status),
		Quantity:         p.Quantity.String(),
		EntryPrice:       p.EntryPrice.String(),
		MarkPrice:        p.MarkPrice.String(),
		Leverage:         p.Leverage,
		InitialMargin:    p.InitialMargin.String(),
		UnrealizedPnL:    p.UnrealizedPnL.String(),
		RealizedPnL:      p.RealizedPnL.String(),
		LiquidationPrice: p.LiquidationPrice.String(),
		SLClosePercent:   p.SLClosePercent,
		TPClosePercent:   p.TPClosePercent,
		CreatedAt:        p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.StopLoss != nil {
		sl := p.StopLoss.String()
		resp.StopLoss = &sl
	}
	if p.TakeProfit != nil {
		tp := p.TakeProfit.String()
		resp.TakeProfit = &tp
	}
	return resp
}
