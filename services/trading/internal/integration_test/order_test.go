package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestPlaceOrder_MarketBuy(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_market_buy"), "password123")

	body := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var order OrderResponse
	err := json.NewDecoder(resp.Body).Decode(&order)
	require.NoError(t, err)

	assert.Greater(t, order.ID, int64(0))
	assert.Equal(t, "BTCUSDT", order.Symbol)
	assert.Equal(t, "BUY", order.Side)
	assert.Equal(t, "MARKET", order.Type)
	assert.Equal(t, "FILLED", order.Status) // Market orders are filled immediately
	assert.Equal(t, "0.1", order.Quantity)
	assert.Equal(t, 10, order.Leverage)

	// Verify position was created
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp.Body.Close()

	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)

	require.Len(t, positions, 1)
	assert.Equal(t, "BTCUSDT", positions[0].Symbol)
	assert.Equal(t, "LONG", positions[0].Side)
	assert.Equal(t, "OPEN", positions[0].Status)
}

func TestPlaceOrder_MarketSell(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_market_sell"), "password123")

	body := map[string]interface{}{
		"symbol":   "ETHUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "1.0",
		"leverage": 5,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var order OrderResponse
	err := json.NewDecoder(resp.Body).Decode(&order)
	require.NoError(t, err)

	assert.Equal(t, "ETHUSDT", order.Symbol)
	assert.Equal(t, "SELL", order.Side)
	assert.Equal(t, "FILLED", order.Status)

	// Verify SHORT position was created
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp.Body.Close()

	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)

	require.Len(t, positions, 1)
	assert.Equal(t, "SHORT", positions[0].Side)
}

func TestPlaceOrder_WithTPSL(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_tpsl"), "password123")

	sl := "49000"
	tp := "55000"
	body := map[string]interface{}{
		"symbol":      "BTCUSDT",
		"side":        "BUY",
		"type":        "MARKET",
		"quantity":    "0.1",
		"leverage":    10,
		"stop_loss":   sl,
		"take_profit": tp,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Verify position has TP/SL
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp.Body.Close()

	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)

	require.Len(t, positions, 1)
	require.NotNil(t, positions[0].StopLoss)
	require.NotNil(t, positions[0].TakeProfit)
	assert.Equal(t, sl, *positions[0].StopLoss)
	assert.Equal(t, tp, *positions[0].TakeProfit)
}

func TestPlaceOrder_InsufficientMargin(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_no_margin"), "password123")

	// Try to open a huge position that exceeds available margin
	// With 10000 balance and leverage 1, max position ~= 10000 USDT
	// BTC at 50000, qty = 1 would need 50000 margin
	body := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "1.0",
		"leverage": 1,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	errMsg := parseErrorResponse(t, resp)
	assert.Contains(t, errMsg, "margin")
}

func TestPlaceOrder_InvalidSymbol(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_invalid_symbol"), "password123")

	body := map[string]interface{}{
		"symbol":   "INVALID",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "1.0",
		"leverage": 10,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	errMsg := parseErrorResponse(t, resp)
	assert.Contains(t, errMsg, "not supported")
}

func TestPlaceOrder_InvalidLeverage(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_invalid_lev"), "password123")

	testCases := []struct {
		name     string
		leverage int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too_high", 150},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"symbol":   "BTCUSDT",
				"side":     "BUY",
				"type":     "MARKET",
				"quantity": "0.1",
				"leverage": tc.leverage,
			}

			resp := makeRequest(t, "POST", "/orders", body, user.Token)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestPlaceOrder_InvalidQuantity(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_invalid_qty"), "password123")

	testCases := []struct {
		name     string
		quantity string
	}{
		{"zero", "0"},
		{"negative", "-0.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"symbol":   "BTCUSDT",
				"side":     "BUY",
				"type":     "MARKET",
				"quantity": tc.quantity,
				"leverage": 10,
			}

			resp := makeRequest(t, "POST", "/orders", body, user.Token)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestPlaceOrder_LimitOrder(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("order_limit"), "password123")

	body := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "LIMIT",
		"quantity": "0.1",
		"price":    "45000",
		"leverage": 10,
	}

	resp := makeRequest(t, "POST", "/orders", body, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var order OrderResponse
	err := json.NewDecoder(resp.Body).Decode(&order)
	require.NoError(t, err)

	assert.Equal(t, "LIMIT", order.Type)
	assert.Equal(t, "PENDING", order.Status) // Limit orders stay pending
	assert.Equal(t, "45000", order.Price)
}

func TestCancelOrder_Pending(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("cancel_pending"), "password123")

	// Create a limit order (stays pending)
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "LIMIT",
		"quantity": "0.1",
		"price":    "45000",
		"leverage": 10,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	defer orderResp.Body.Close()
	require.Equal(t, http.StatusCreated, orderResp.StatusCode)

	var order OrderResponse
	json.NewDecoder(orderResp.Body).Decode(&order)

	// Cancel the order
	cancelResp := makeRequest(t, "DELETE", fmt.Sprintf("/orders/%d", order.ID), nil, user.Token)
	defer cancelResp.Body.Close()

	assert.Equal(t, http.StatusOK, cancelResp.StatusCode)

	var result map[string]string
	json.NewDecoder(cancelResp.Body).Decode(&result)
	assert.Equal(t, "cancelled", result["status"])
}

func TestCancelOrder_AlreadyFilled(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("cancel_filled"), "password123")

	// Create a market order (gets filled immediately)
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	defer orderResp.Body.Close()
	require.Equal(t, http.StatusCreated, orderResp.StatusCode)

	var order OrderResponse
	json.NewDecoder(orderResp.Body).Decode(&order)

	// Try to cancel the filled order
	cancelResp := makeRequest(t, "DELETE", fmt.Sprintf("/orders/%d", order.ID), nil, user.Token)
	defer cancelResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, cancelResp.StatusCode)
}

func TestCancelOrder_NotFound(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("cancel_notfound"), "password123")

	resp := makeRequest(t, "DELETE", "/orders/999999", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCancelOrder_WrongUser(t *testing.T) {
	cleanupDatabase(t)

	user1 := registerUser(t, uniqueEmail("cancel_user1"), "password123")
	user2 := registerUser(t, uniqueEmail("cancel_user2"), "password123")

	// User1 creates a limit order
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "LIMIT",
		"quantity": "0.1",
		"price":    "45000",
		"leverage": 10,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user1.Token)
	defer orderResp.Body.Close()

	var order OrderResponse
	json.NewDecoder(orderResp.Body).Decode(&order)

	// User2 tries to cancel user1's order
	cancelResp := makeRequest(t, "DELETE", fmt.Sprintf("/orders/%d", order.ID), nil, user2.Token)
	defer cancelResp.Body.Close()

	assert.Equal(t, http.StatusNotFound, cancelResp.StatusCode)
}

func TestGetOrders(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("get_orders"), "password123")

	// Create multiple orders
	for i := 0; i < 3; i++ {
		body := map[string]interface{}{
			"symbol":   "BTCUSDT",
			"side":     "BUY",
			"type":     "LIMIT",
			"quantity": "0.01",
			"price":    fmt.Sprintf("%d", 45000+i*100),
			"leverage": 10,
		}
		resp := makeRequest(t, "POST", "/orders", body, user.Token)
		resp.Body.Close()
	}

	// Get orders
	resp := makeRequest(t, "GET", "/orders", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orders []OrderResponse
	json.NewDecoder(resp.Body).Decode(&orders)

	assert.Len(t, orders, 3)
}
