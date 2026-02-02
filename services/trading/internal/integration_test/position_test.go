package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	CreatedAt        string  `json:"created_at"`
}

func TestGetPositions_Empty(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_empty"), "password123")

	resp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var positions []PositionResponse
	err := json.NewDecoder(resp.Body).Decode(&positions)
	require.NoError(t, err)

	assert.Len(t, positions, 0)
}

func TestGetPositions_WithOpen(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_with_open"), "password123")

	// Open a position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get positions
	resp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var positions []PositionResponse
	err := json.NewDecoder(resp.Body).Decode(&positions)
	require.NoError(t, err)

	require.Len(t, positions, 1)
	pos := positions[0]

	assert.Equal(t, "BTCUSDT", pos.Symbol)
	assert.Equal(t, "LONG", pos.Side)
	assert.Equal(t, "OPEN", pos.Status)
	assert.Equal(t, "0.1", pos.Quantity)
	assert.Equal(t, 10, pos.Leverage)

	// Verify entry price is the ask price (buy orders execute at ask)
	entryPrice, _ := decimal.NewFromString(pos.EntryPrice)
	assert.True(t, entryPrice.Equal(decimal.NewFromFloat(50010)), "entry price should be ask price")

	// Verify initial margin = (qty * price) / leverage
	// (0.1 * 50010) / 10 = 500.1
	initialMargin, _ := decimal.NewFromString(pos.InitialMargin)
	expectedMargin := decimal.NewFromFloat(0.1 * 50010 / 10)
	assert.True(t, initialMargin.Equal(expectedMargin), "initial margin calculation")
}

func TestClosePosition_Success(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_close"), "password123")

	// Open a position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position ID
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	require.Len(t, positions, 1)
	positionID := positions[0].ID

	// Close position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", positionID), nil, user.Token)
	defer closeResp.Body.Close()

	assert.Equal(t, http.StatusOK, closeResp.StatusCode)

	var closeResult struct {
		Status      string `json:"status"`
		RealizedPnL string `json:"realized_pnl"`
	}
	json.NewDecoder(closeResp.Body).Decode(&closeResult)

	assert.Equal(t, "closed", closeResult.Status)
	// PnL should be negative due to spread (bought at 50010, sold at 50000)
	// PnL = 0.1 * (50000 - 50010) = -1
	pnl, _ := decimal.NewFromString(closeResult.RealizedPnL)
	assert.True(t, pnl.IsNegative(), "PnL should be negative due to spread")

	// Verify position is no longer in open positions
	posResp2 := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp2.Body.Close()

	var positions2 []PositionResponse
	json.NewDecoder(posResp2.Body).Decode(&positions2)
	assert.Len(t, positions2, 0)
}

func TestClosePosition_ShortPnL(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_short_pnl"), "password123")

	// Open a SHORT position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	require.Len(t, positions, 1)

	assert.Equal(t, "SHORT", positions[0].Side)
	// Entry price for short is bid price = 50000
	entryPrice, _ := decimal.NewFromString(positions[0].EntryPrice)
	assert.True(t, entryPrice.Equal(decimal.NewFromFloat(50000)))

	// Close position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", positions[0].ID), nil, user.Token)
	defer closeResp.Body.Close()

	var closeResult struct {
		RealizedPnL string `json:"realized_pnl"`
	}
	json.NewDecoder(closeResp.Body).Decode(&closeResult)

	// Short PnL = qty * (entry - close)
	// Close price for short = ask = 50010
	// PnL = 0.1 * (50000 - 50010) = -1
	pnl, _ := decimal.NewFromString(closeResult.RealizedPnL)
	assert.True(t, pnl.IsNegative())
}

func TestClosePosition_NotFound(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_not_found"), "password123")

	resp := makeRequest(t, "POST", "/positions/999999/close", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestClosePosition_WrongUser(t *testing.T) {
	cleanupDatabase(t)

	user1 := registerUser(t, uniqueEmail("pos_user1"), "password123")
	user2 := registerUser(t, uniqueEmail("pos_user2"), "password123")

	// User1 opens a position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user1.Token)
	orderResp.Body.Close()

	// Get position ID
	posResp := makeRequest(t, "GET", "/positions", nil, user1.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()

	// User2 tries to close user1's position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", positions[0].ID), nil, user2.Token)
	defer closeResp.Body.Close()

	assert.Equal(t, http.StatusNotFound, closeResp.StatusCode)
}

func TestUpdateTPSL_Valid(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_update_tpsl"), "password123")

	// Open a LONG position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	require.Len(t, positions, 1)
	positionID := positions[0].ID

	// Update TP/SL
	// For LONG: SL < entry < TP
	// Entry is 50010, so SL = 48000, TP = 55000
	updateBody := map[string]interface{}{
		"stop_loss":   "48000",
		"take_profit": "55000",
	}

	updateResp := makeRequest(t, "PATCH", fmt.Sprintf("/positions/%d", positionID), updateBody, user.Token)
	defer updateResp.Body.Close()

	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updatedPos PositionResponse
	json.NewDecoder(updateResp.Body).Decode(&updatedPos)

	require.NotNil(t, updatedPos.StopLoss)
	require.NotNil(t, updatedPos.TakeProfit)
	assert.Equal(t, "48000", *updatedPos.StopLoss)
	assert.Equal(t, "55000", *updatedPos.TakeProfit)
}

func TestUpdateTPSL_InvalidSL_LongAboveEntry(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_invalid_sl"), "password123")

	// Open a LONG position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	positionID := positions[0].ID

	// Try to set SL above entry (invalid for LONG)
	// Entry is 50010, SL = 52000 is invalid
	updateBody := map[string]interface{}{
		"stop_loss": "52000",
	}

	updateResp := makeRequest(t, "PATCH", fmt.Sprintf("/positions/%d", positionID), updateBody, user.Token)
	defer updateResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, updateResp.StatusCode)
}

func TestUpdateTPSL_InvalidTP_LongBelowEntry(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_invalid_tp"), "password123")

	// Open a LONG position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	positionID := positions[0].ID

	// Try to set TP below entry (invalid for LONG)
	// Entry is 50010, TP = 48000 is invalid
	updateBody := map[string]interface{}{
		"take_profit": "48000",
	}

	updateResp := makeRequest(t, "PATCH", fmt.Sprintf("/positions/%d", positionID), updateBody, user.Token)
	defer updateResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, updateResp.StatusCode)
}

func TestUpdateTPSL_Short(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_short_tpsl"), "password123")

	// Open a SHORT position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	positionID := positions[0].ID

	// For SHORT: SL > entry > TP
	// Entry is 50000, so SL = 52000, TP = 45000
	updateBody := map[string]interface{}{
		"stop_loss":   "52000",
		"take_profit": "45000",
	}

	updateResp := makeRequest(t, "PATCH", fmt.Sprintf("/positions/%d", positionID), updateBody, user.Token)
	defer updateResp.Body.Close()

	assert.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updatedPos PositionResponse
	json.NewDecoder(updateResp.Body).Decode(&updatedPos)

	require.NotNil(t, updatedPos.StopLoss)
	require.NotNil(t, updatedPos.TakeProfit)
}

func TestPosition_LiquidationPrice(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("pos_liq_price"), "password123")

	// Open a LONG position with leverage 10
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp.Body.Close()

	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	require.Len(t, positions, 1)

	// Liquidation price formula for LONG:
	// LiqPrice = EntryPrice * (1 - 1/Leverage + MaintenanceRate)
	// With leverage 10 and maintenance rate 0.005:
	// LiqPrice = 50010 * (1 - 0.1 + 0.005) = 50010 * 0.905 = 45259.05
	liqPrice, _ := decimal.NewFromString(positions[0].LiquidationPrice)
	entryPrice, _ := decimal.NewFromString(positions[0].EntryPrice)

	// Verify liquidation price is below entry price for LONG
	assert.True(t, liqPrice.LessThan(entryPrice), "liquidation price should be below entry for LONG")
}
