package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Ticker24h struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"priceChange"`
	PriceChangePercent float64 `json:"priceChangePercent"`
	LastPrice          float64 `json:"lastPrice"`
	HighPrice          float64 `json:"highPrice"`
	LowPrice           float64 `json:"lowPrice"`
	Volume             float64 `json:"volume"`
}

type TickerHandler struct {
	client           *http.Client
	supportedSymbols []string
}

func NewTickerHandler(supportedSymbols []string) *TickerHandler {
	return &TickerHandler{
		client:           &http.Client{Timeout: 10 * time.Second},
		supportedSymbols: supportedSymbols,
	}
}

func (h *TickerHandler) GetTicker24h(w http.ResponseWriter, r *http.Request) {
	symbolsJSON := "[\"" + strings.Join(h.supportedSymbols, "\",\"") + "\"]"

	url := "https://api.binance.com/api/v3/ticker/24hr?symbols=" + symbolsJSON

	resp, err := h.client.Get(url)
	if err != nil {
		writeError(w, "failed to fetch ticker from Binance", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeError(w, "Binance ticker API error", resp.StatusCode)
		return
	}

	var raw []struct {
		Symbol             string `json:"symbol"`
		PriceChange        string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		LastPrice          string `json:"lastPrice"`
		HighPrice          string `json:"highPrice"`
		LowPrice           string `json:"lowPrice"`
		Volume             string `json:"volume"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		writeError(w, "failed to parse Binance ticker response", http.StatusInternalServerError)
		return
	}

	tickers := make([]Ticker24h, 0, len(raw))
	for _, t := range raw {
		priceChange, _ := strconv.ParseFloat(t.PriceChange, 64)
		priceChangePercent, _ := strconv.ParseFloat(t.PriceChangePercent, 64)
		lastPrice, _ := strconv.ParseFloat(t.LastPrice, 64)
		highPrice, _ := strconv.ParseFloat(t.HighPrice, 64)
		lowPrice, _ := strconv.ParseFloat(t.LowPrice, 64)
		volume, _ := strconv.ParseFloat(t.Volume, 64)

		tickers = append(tickers, Ticker24h{
			Symbol:             t.Symbol,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			LastPrice:          lastPrice,
			HighPrice:          highPrice,
			LowPrice:           lowPrice,
			Volume:             volume,
		})
	}

	writeJSON(w, tickers, http.StatusOK)
}
