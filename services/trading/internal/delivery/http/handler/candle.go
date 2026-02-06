package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

var allowedIntervals = map[string]bool{
	"1s": true, "1m": true, "5m": true, "15m": true,
	"1h": true, "4h": true, "1d": true, "1w": true,
}

type Candle struct {
	Time  int64   `json:"time"`
	Open  float64 `json:"open"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Close float64 `json:"close"`
}

type CandleHandler struct {
	client *http.Client
}

func NewCandleHandler() *CandleHandler {
	return &CandleHandler{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *CandleHandler) GetCandles(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeError(w, "symbol is required", http.StatusBadRequest)
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1m"
	}
	if !allowedIntervals[interval] {
		writeError(w, fmt.Sprintf("unsupported interval: %s", interval), http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 300
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 || parsed > 1000 {
			writeError(w, "limit must be between 1 and 1000", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		symbol, interval, limit)

	resp, err := h.client.Get(url)
	if err != nil {
		writeError(w, "failed to fetch candles from Binance", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		writeError(w, fmt.Sprintf("Binance API error: %s", string(body)), resp.StatusCode)
		return
	}

	var raw [][]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		writeError(w, "failed to parse Binance response", http.StatusInternalServerError)
		return
	}

	candles := make([]Candle, 0, len(raw))
	for _, kline := range raw {
		if len(kline) < 5 {
			continue
		}

		var openTimeMs int64
		if err := json.Unmarshal(kline[0], &openTimeMs); err != nil {
			continue
		}

		var openStr, highStr, lowStr, closeStr string
		if err := json.Unmarshal(kline[1], &openStr); err != nil {
			continue
		}
		if err := json.Unmarshal(kline[2], &highStr); err != nil {
			continue
		}
		if err := json.Unmarshal(kline[3], &lowStr); err != nil {
			continue
		}
		if err := json.Unmarshal(kline[4], &closeStr); err != nil {
			continue
		}

		open, _ := strconv.ParseFloat(openStr, 64)
		high, _ := strconv.ParseFloat(highStr, 64)
		low, _ := strconv.ParseFloat(lowStr, 64)
		close_, _ := strconv.ParseFloat(closeStr, 64)

		candles = append(candles, Candle{
			Time:  openTimeMs / 1000,
			Open:  open,
			High:  high,
			Low:   low,
			Close: close_,
		})
	}

	writeJSON(w, candles, http.StatusOK)
}
