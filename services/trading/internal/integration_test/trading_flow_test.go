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

func TestFullTradingCycle(t *testing.T) {
	cleanupDatabase(t)

	// Step 1: Register user and get token
	email := uniqueEmail("full_cycle")
	user := registerUser(t, email, "password123")
	require.NotEmpty(t, user.Token)
	t.Logf("Step 1: User registered with ID %d", user.UserID)

	// Step 2: Verify initial account balance = 10000
	accountResp := makeRequest(t, "GET", "/account", nil, user.Token)
	var accountInfo AccountInfo
	json.NewDecoder(accountResp.Body).Decode(&accountInfo)
	accountResp.Body.Close()

	assert.Equal(t, "10000.00", accountInfo.Balance)
	assert.Equal(t, "10000.00", accountInfo.Equity)
	assert.Equal(t, "0.00", accountInfo.UsedMargin)
	t.Log("Step 2: Initial balance verified = 10000 USDT")

	// Step 3: Place BUY order (BTCUSDT, qty=0.1, leverage=10)
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	require.Equal(t, http.StatusCreated, orderResp.StatusCode)

	var order OrderResponse
	json.NewDecoder(orderResp.Body).Decode(&order)
	orderResp.Body.Close()

	assert.Equal(t, "FILLED", order.Status)
	t.Logf("Step 3: Order placed and filled, ID=%d", order.ID)

	// Step 4: Verify position (LONG, entry_price, liquidation_price)
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()

	require.Len(t, positions, 1)
	position := positions[0]

	assert.Equal(t, "BTCUSDT", position.Symbol)
	assert.Equal(t, "LONG", position.Side)
	assert.Equal(t, "OPEN", position.Status)
	assert.Equal(t, "0.1", position.Quantity)
	assert.Equal(t, 10, position.Leverage)

	// Entry price should be ask = 50010
	entryPrice, _ := decimal.NewFromString(position.EntryPrice)
	expectedEntry := decimal.NewFromFloat(50010)
	assert.True(t, entryPrice.Equal(expectedEntry), "entry price = ask price")

	// Initial margin = (0.1 * 50010) / 10 = 500.1
	initialMargin, _ := decimal.NewFromString(position.InitialMargin)
	expectedMargin := decimal.NewFromFloat(500.1)
	assert.True(t, initialMargin.Equal(expectedMargin), "initial margin calculation")

	// Liquidation price should be calculated
	liqPrice, _ := decimal.NewFromString(position.LiquidationPrice)
	assert.True(t, liqPrice.LessThan(entryPrice), "liquidation price < entry for LONG")

	t.Logf("Step 4: Position verified - entry=%s, margin=%s, liq=%s",
		position.EntryPrice, position.InitialMargin, position.LiquidationPrice)

	// Step 5: Update TP/SL
	updateBody := map[string]interface{}{
		"stop_loss":   "48000",
		"take_profit": "55000",
	}

	updateResp := makeRequest(t, "PATCH", fmt.Sprintf("/positions/%d", position.ID), updateBody, user.Token)
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updatedPos PositionResponse
	json.NewDecoder(updateResp.Body).Decode(&updatedPos)
	updateResp.Body.Close()

	require.NotNil(t, updatedPos.StopLoss)
	require.NotNil(t, updatedPos.TakeProfit)
	assert.Equal(t, "48000", *updatedPos.StopLoss)
	assert.Equal(t, "55000", *updatedPos.TakeProfit)
	t.Log("Step 5: TP/SL updated - SL=48000, TP=55000")

	// Step 6: Close position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", position.ID), nil, user.Token)
	require.Equal(t, http.StatusOK, closeResp.StatusCode)

	var closeResult struct {
		Status      string `json:"status"`
		RealizedPnL string `json:"realized_pnl"`
	}
	json.NewDecoder(closeResp.Body).Decode(&closeResult)
	closeResp.Body.Close()

	assert.Equal(t, "closed", closeResult.Status)

	// PnL calculation:
	// Entry (LONG BUY) = 50010 (ask)
	// Close (LONG SELL) = 50000 (bid)
	// PnL = 0.1 * (50000 - 50010) = -1
	realizedPnL, _ := decimal.NewFromString(closeResult.RealizedPnL)
	expectedPnL := decimal.NewFromFloat(-1)
	assert.True(t, realizedPnL.Equal(expectedPnL), "PnL = qty * (close - entry)")
	t.Logf("Step 6: Position closed - PnL=%s", closeResult.RealizedPnL)

	// Step 7: Check account (balance changed by PnL only; margin is virtual)
	accountResp2 := makeRequest(t, "GET", "/account", nil, user.Token)
	var accountInfo2 AccountInfo
	json.NewDecoder(accountResp2.Body).Decode(&accountInfo2)
	accountResp2.Body.Close()

	// Final balance = initial + PnL = 10000 + (-1) = 9999
	// Margin is virtual — never deducted on open, not credited on close
	finalBalance, _ := decimal.NewFromString(accountInfo2.Balance)
	expectedBalance := decimal.NewFromFloat(9999)
	assert.True(t, finalBalance.Equal(expectedBalance), "balance = initial + PnL")

	assert.Equal(t, "0.00", accountInfo2.UsedMargin, "margin released after close")
	t.Logf("Step 7: Account verified - balance=%s, margin=%s", accountInfo2.Balance, accountInfo2.UsedMargin)

	// Step 8: Check trades (OPEN + CLOSE records)
	tradesResp := makeRequest(t, "GET", "/trades", nil, user.Token)
	require.Equal(t, http.StatusOK, tradesResp.StatusCode)

	var trades []TradeResponse
	json.NewDecoder(tradesResp.Body).Decode(&trades)
	tradesResp.Body.Close()

	require.Len(t, trades, 2, "should have OPEN and CLOSE trades")

	// Find OPEN and CLOSE trades
	var openTrade, closeTrade *TradeResponse
	for i := range trades {
		if trades[i].Type == "OPEN" {
			openTrade = &trades[i]
		} else if trades[i].Type == "CLOSE" {
			closeTrade = &trades[i]
		}
	}

	require.NotNil(t, openTrade, "OPEN trade should exist")
	require.NotNil(t, closeTrade, "CLOSE trade should exist")

	assert.Equal(t, "BTCUSDT", openTrade.Symbol)
	assert.Equal(t, "LONG", openTrade.Side)
	assert.Equal(t, "0.1", openTrade.Quantity)
	assert.Equal(t, "0", openTrade.PnL, "OPEN trade has no PnL")

	assert.Equal(t, "BTCUSDT", closeTrade.Symbol)
	assert.Equal(t, "LONG", closeTrade.Side)
	assert.Equal(t, closeResult.RealizedPnL, closeTrade.PnL)

	t.Log("Step 8: Trades verified - OPEN and CLOSE records exist")
	t.Log("=== Full Trading Cycle Test PASSED ===")
}

func TestTradingCycle_ShortPosition(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("short_cycle"), "password123")

	// Open SHORT position
	orderBody := map[string]interface{}{
		"symbol":   "ETHUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "1.0",
		"leverage": 5,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	require.Equal(t, http.StatusCreated, orderResp.StatusCode)
	orderResp.Body.Close()

	// Verify SHORT position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()

	require.Len(t, positions, 1)
	position := positions[0]

	assert.Equal(t, "ETHUSDT", position.Symbol)
	assert.Equal(t, "SHORT", position.Side)

	// Entry price for SHORT = bid = 3000
	entryPrice, _ := decimal.NewFromString(position.EntryPrice)
	assert.True(t, entryPrice.Equal(decimal.NewFromFloat(3000)))

	// Initial margin = (1.0 * 3000) / 5 = 600
	initialMargin, _ := decimal.NewFromString(position.InitialMargin)
	assert.True(t, initialMargin.Equal(decimal.NewFromFloat(600)))

	// Liquidation price for SHORT should be above entry
	liqPrice, _ := decimal.NewFromString(position.LiquidationPrice)
	assert.True(t, liqPrice.GreaterThan(entryPrice), "liquidation price > entry for SHORT")

	// Close position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", position.ID), nil, user.Token)
	var closeResult struct {
		RealizedPnL string `json:"realized_pnl"`
	}
	json.NewDecoder(closeResp.Body).Decode(&closeResult)
	closeResp.Body.Close()

	// Short PnL = qty * (entry - close)
	// Entry = 3000 (bid), Close = 3002 (ask)
	// PnL = 1.0 * (3000 - 3002) = -2
	realizedPnL, _ := decimal.NewFromString(closeResult.RealizedPnL)
	assert.True(t, realizedPnL.Equal(decimal.NewFromFloat(-2)))

	// Verify final balance
	// Final balance = initial + PnL = 10000 + (-2) = 9998
	// Margin is virtual — never deducted on open, not credited on close
	accountResp := makeRequest(t, "GET", "/account", nil, user.Token)
	var accountInfo AccountInfo
	json.NewDecoder(accountResp.Body).Decode(&accountInfo)
	accountResp.Body.Close()

	finalBalance, _ := decimal.NewFromString(accountInfo.Balance)
	assert.True(t, finalBalance.Equal(decimal.NewFromFloat(9998)))
}

func TestTradingCycle_MultiplePositions(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("multi_pos"), "password123")

	// Open BTC LONG
	btcOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.05",
		"leverage": 20,
	}
	resp1 := makeRequest(t, "POST", "/orders", btcOrder, user.Token)
	resp1.Body.Close()

	// Open ETH SHORT
	ethOrder := map[string]interface{}{
		"symbol":   "ETHUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "0.5",
		"leverage": 10,
	}
	resp2 := makeRequest(t, "POST", "/orders", ethOrder, user.Token)
	resp2.Body.Close()

	// Open SOL LONG
	solOrder := map[string]interface{}{
		"symbol":   "SOLUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "10",
		"leverage": 5,
	}
	resp3 := makeRequest(t, "POST", "/orders", solOrder, user.Token)
	resp3.Body.Close()

	// Verify 3 positions
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()

	assert.Len(t, positions, 3)

	// Calculate total margin used
	// BTC: (0.05 * 50010) / 20 = 125.025
	// ETH: (0.5 * 3000) / 10 = 150
	// SOL: (10 * 100.1) / 5 = 200.2
	// Total = 475.225

	accountResp := makeRequest(t, "GET", "/account", nil, user.Token)
	var accountInfo AccountInfo
	json.NewDecoder(accountResp.Body).Decode(&accountInfo)
	accountResp.Body.Close()

	usedMargin, _ := decimal.NewFromString(accountInfo.UsedMargin)
	assert.True(t, usedMargin.GreaterThan(decimal.NewFromFloat(400)))

	// Close all positions
	for _, pos := range positions {
		closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", pos.ID), nil, user.Token)
		closeResp.Body.Close()
	}

	// Verify all margins released
	accountResp2 := makeRequest(t, "GET", "/account", nil, user.Token)
	var accountInfo2 AccountInfo
	json.NewDecoder(accountResp2.Body).Decode(&accountInfo2)
	accountResp2.Body.Close()

	assert.Equal(t, "0.00", accountInfo2.UsedMargin)
}

func TestTradingCycle_OppositeOrderClosesPosition(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("opposite_close"), "password123")

	// Open LONG position
	buyOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	resp1 := makeRequest(t, "POST", "/orders", buyOrder, user.Token)
	resp1.Body.Close()

	// Verify position exists
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	require.Len(t, positions, 1)

	// Place SELL order to close position
	sellOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	resp2 := makeRequest(t, "POST", "/orders", sellOrder, user.Token)
	resp2.Body.Close()

	// Verify position is closed (no open positions)
	posResp2 := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions2 []PositionResponse
	json.NewDecoder(posResp2.Body).Decode(&positions2)
	posResp2.Body.Close()

	assert.Len(t, positions2, 0, "opposite order should close position")
}

func TestTradingCycle_PartialClose(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("partial_close"), "password123")

	// Open LONG position with quantity 0.2
	buyOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.2",
		"leverage": 10,
	}
	resp1 := makeRequest(t, "POST", "/orders", buyOrder, user.Token)
	resp1.Body.Close()

	// Verify position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()
	require.Len(t, positions, 1)
	assert.Equal(t, "0.2", positions[0].Quantity)

	// Place partial SELL order (0.1 of 0.2)
	sellOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "SELL",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	resp2 := makeRequest(t, "POST", "/orders", sellOrder, user.Token)
	resp2.Body.Close()

	// Verify position is reduced, not closed
	posResp2 := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions2 []PositionResponse
	json.NewDecoder(posResp2.Body).Decode(&positions2)
	posResp2.Body.Close()

	require.Len(t, positions2, 1, "position should still exist after partial close")
	assert.Equal(t, "0.1", positions2[0].Quantity, "position quantity should be reduced")
}

func TestGetTrades(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("get_trades"), "password123")

	// Initially no trades
	tradesResp := makeRequest(t, "GET", "/trades", nil, user.Token)
	var trades []TradeResponse
	json.NewDecoder(tradesResp.Body).Decode(&trades)
	tradesResp.Body.Close()
	assert.Len(t, trades, 0)

	// Open and close a position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()

	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	var positions []PositionResponse
	json.NewDecoder(posResp.Body).Decode(&positions)
	posResp.Body.Close()

	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", positions[0].ID), nil, user.Token)
	closeResp.Body.Close()

	// Now should have 2 trades
	tradesResp2 := makeRequest(t, "GET", "/trades", nil, user.Token)
	var trades2 []TradeResponse
	json.NewDecoder(tradesResp2.Body).Decode(&trades2)
	tradesResp2.Body.Close()

	assert.Len(t, trades2, 2)

	// Verify trade types
	tradeTypes := make(map[string]bool)
	for _, tr := range trades2 {
		tradeTypes[tr.Type] = true
	}
	assert.True(t, tradeTypes["OPEN"])
	assert.True(t, tradeTypes["CLOSE"])
}
