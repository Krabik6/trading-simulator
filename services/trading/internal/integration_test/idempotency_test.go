package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaceOrder_IdempotentReplay(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("idem_replay"), "password123")
	idempotencyKey := "order-replay-key-1"

	body := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	resp1 := makeRequestWithHeaders(t, "POST", "/orders", body, user.Token, map[string]string{
		"Idempotency-Key": idempotencyKey,
	})
	defer resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	var first OrderResponse
	require.NoError(t, json.NewDecoder(resp1.Body).Decode(&first))

	resp2 := makeRequestWithHeaders(t, "POST", "/orders", body, user.Token, map[string]string{
		"Idempotency-Key": idempotencyKey,
	})
	defer resp2.Body.Close()
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Equal(t, "true", resp2.Header.Get("X-Idempotent-Replay"))

	var second OrderResponse
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&second))

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, first.Status, second.Status)

	positionsResp := makeRequest(t, "GET", "/positions", nil, user.Token)
	defer positionsResp.Body.Close()

	var positions []PositionResponse
	require.NoError(t, json.NewDecoder(positionsResp.Body).Decode(&positions))
	require.Len(t, positions, 1)
	assert.Equal(t, "0.1", positions[0].Quantity)
}

func TestPlaceOrder_IdempotencyConflict(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("idem_conflict"), "password123")
	idempotencyKey := "order-conflict-key-1"

	body1 := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}
	body2 := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.2",
		"leverage": 10,
	}

	resp1 := makeRequestWithHeaders(t, "POST", "/orders", body1, user.Token, map[string]string{
		"Idempotency-Key": idempotencyKey,
	})
	resp1.Body.Close()
	require.Equal(t, http.StatusCreated, resp1.StatusCode)

	resp2 := makeRequestWithHeaders(t, "POST", "/orders", body2, user.Token, map[string]string{
		"Idempotency-Key": idempotencyKey,
	})
	defer resp2.Body.Close()
	require.Equal(t, http.StatusConflict, resp2.StatusCode)

	errMsg := parseErrorResponse(t, resp2)
	assert.Contains(t, errMsg, "different request")
}

func TestPlaceOrder_IdempotencyKeyRequired(t *testing.T) {
	cleanupDatabase(t)

	user := registerUser(t, uniqueEmail("idem_required"), "password123")

	body := map[string]interface{}{
		"symbol":   "BTCUSDT",
		"side":     "BUY",
		"type":     "MARKET",
		"quantity": "0.1",
		"leverage": 10,
	}

	resp := makeRequestWithHeaders(t, "POST", "/orders", body, user.Token, map[string]string{
		"Idempotency-Key": "",
	})
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	errMsg := parseErrorResponse(t, resp)
	assert.Contains(t, errMsg, "idempotency")
}
