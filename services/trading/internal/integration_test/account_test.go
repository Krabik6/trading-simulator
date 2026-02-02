package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AccountInfo struct {
	Balance         string `json:"balance"`
	Equity          string `json:"equity"`
	UsedMargin      string `json:"used_margin"`
	AvailableMargin string `json:"available_margin"`
	UnrealizedPnL   string `json:"unrealized_pnl"`
	MarginRatio     string `json:"margin_ratio"`
}

func TestGetAccount_InitialBalance(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("account_initial"), "password123")

	resp := makeRequest(t, "GET", "/account", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info AccountInfo
	err := json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)

	assert.Equal(t, "10000.00", info.Balance)
	assert.Equal(t, "10000.00", info.Equity)
	assert.Equal(t, "0.00", info.UsedMargin)
	assert.Equal(t, "10000.00", info.AvailableMargin)
	assert.Equal(t, "0.00", info.UnrealizedPnL)
	assert.Equal(t, "0.0000", info.MarginRatio)
}

func TestGetAccount_WithOpenPosition(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("account_with_pos"), "password123")

	// Place a market order to open position
	orderBody := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	orderResp := makeRequest(t, "POST", "/orders", orderBody, user.Token)
	orderResp.Body.Close()
	require.Equal(t, http.StatusCreated, orderResp.StatusCode)

	// Check account - margin should be used
	resp := makeRequest(t, "GET", "/account", nil, user.Token)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info AccountInfo
	err := json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)

	// Initial margin = (0.1 * 50010) / 10 = 500.1
	// Balance should be 10000 (margin is locked, not deducted yet in this model)
	assert.Equal(t, "10000.00", info.Balance)

	// Used margin should be approximately 500.1 (0.1 * 50010 / 10)
	assert.NotEqual(t, "0.00", info.UsedMargin)

	// Available margin = Balance - UsedMargin (approximately)
	assert.NotEqual(t, info.Balance, info.AvailableMargin)
}

func TestGetAccount_AfterClosingPosition(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("account_close"), "password123")

	// Place a market order to open position
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

	var orderResult struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(orderResp.Body).Decode(&orderResult)

	// Get position
	posResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer posResp.Body.Close()
	require.Equal(t, http.StatusOK, posResp.StatusCode)

	var positions []struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(posResp.Body).Decode(&positions)
	require.Len(t, positions, 1)

	// Close position
	closeResp := makeRequest(t, "POST", fmt.Sprintf("/positions/%d/close", positions[0].ID), nil, user.Token)
	defer closeResp.Body.Close()
	require.Equal(t, http.StatusOK, closeResp.StatusCode)

	// Check account after close
	accountResp := makeRequest(t, "GET", "/account", nil, user.Token)
	defer accountResp.Body.Close()

	var info AccountInfo
	err := json.NewDecoder(accountResp.Body).Decode(&info)
	require.NoError(t, err)

	// Margin should be released
	assert.Equal(t, "0.00", info.UsedMargin)
	// Balance will change by PnL (spread loss in this case)
}

func TestGetAccount_MultiplePositions(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("account_multi"), "password123")

	// Open BTC position
	btcOrder := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.05",
		"leverage": 20,
	}
	resp1 := makeRequest(t, "POST", "/orders", btcOrder, user.Token)
	resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Open ETH position
	ethOrder := map[string]interface{}{
		"symbol":   "ETHUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "1.0",
		"leverage": 10,
	}
	resp2 := makeRequest(t, "POST", "/orders", ethOrder, user.Token)
	resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)

	// Check account
	accountResp := makeRequest(t, "GET", "/account", nil, user.Token)
	defer accountResp.Body.Close()

	var info AccountInfo
	err := json.NewDecoder(accountResp.Body).Decode(&info)
	require.NoError(t, err)

	// Used margin should be sum of both positions' margins
	// BTC: (0.05 * 50010) / 20 = 125.025
	// ETH: (1.0 * 3002) / 10 = 300.2
	// Total ~= 425.225
	assert.NotEqual(t, "0.00", info.UsedMargin)
}

