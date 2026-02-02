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
	orderuc "trading/internal/usecase/order"
)

type OrderHandler struct {
	orderUC *orderuc.UseCase
}

func NewOrderHandler(orderUC *orderuc.UseCase) *OrderHandler {
	return &OrderHandler{orderUC: orderUC}
}

type PlaceOrderRequest struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`        // BUY or SELL
	Type       string  `json:"type"`        // MARKET or LIMIT
	Quantity   string  `json:"quantity"`    // decimal string
	Price      string  `json:"price"`       // for limit orders
	Leverage   int     `json:"leverage"`
	StopLoss   *string `json:"stop_loss"`   // optional
	TakeProfit *string `json:"take_profit"` // optional
}

type OrderResponse struct {
	ID         int64   `json:"id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	Quantity   string  `json:"quantity"`
	Price      string  `json:"price"`
	Leverage   int     `json:"leverage"`
	StopLoss   *string `json:"stop_loss,omitempty"`
	TakeProfit *string `json:"take_profit,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		writeError(w, "invalid quantity", http.StatusBadRequest)
		return
	}

	var price decimal.Decimal
	if req.Type == "LIMIT" {
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			writeError(w, "invalid price", http.StatusBadRequest)
			return
		}
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

	output, err := h.orderUC.PlaceOrder(r.Context(), orderuc.PlaceOrderInput{
		UserID:     userID,
		Symbol:     req.Symbol,
		Side:       domain.OrderSide(req.Side),
		Type:       domain.OrderType(req.Type),
		Quantity:   quantity,
		Price:      price,
		Leverage:   req.Leverage,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, domain.ErrInsufficientMargin) || errors.Is(err, domain.ErrInsufficientBalance) {
			status = http.StatusUnprocessableEntity
		} else if errors.Is(err, domain.ErrPriceNotAvailable) {
			status = http.StatusServiceUnavailable
		}
		writeError(w, err.Error(), status)
		return
	}

	writeJSON(w, orderToResponse(output.Order), http.StatusCreated)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}

	orders, err := h.orderUC.GetOrders(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, "failed to get orders", http.StatusInternalServerError)
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = orderToResponse(&o)
	}

	writeJSON(w, response, http.StatusOK)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		writeError(w, "invalid order id", http.StatusBadRequest)
		return
	}

	if err := h.orderUC.CancelOrder(r.Context(), userID, domain.OrderID(orderID)); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			writeError(w, "order not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrOrderNotPending) {
			writeError(w, "order cannot be cancelled", http.StatusBadRequest)
			return
		}
		writeError(w, "failed to cancel order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "cancelled"}, http.StatusOK)
}

func orderToResponse(o *domain.Order) OrderResponse {
	resp := OrderResponse{
		ID:        int64(o.ID),
		Symbol:    o.Symbol,
		Side:      string(o.Side),
		Type:      string(o.Type),
		Status:    string(o.Status),
		Quantity:  o.Quantity.String(),
		Price:     o.Price.String(),
		Leverage:  o.Leverage,
		CreatedAt: o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if o.StopLoss != nil {
		sl := o.StopLoss.String()
		resp.StopLoss = &sl
	}
	if o.TakeProfit != nil {
		tp := o.TakeProfit.String()
		resp.TakeProfit = &tp
	}
	return resp
}
