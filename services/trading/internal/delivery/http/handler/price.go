package handler

import (
	"net/http"

	"trading/internal/domain"
)

// PriceHandler handles price-related requests
type PriceHandler struct {
	priceCache domain.PriceCache
	symbols    []string
}

// NewPriceHandler creates a new PriceHandler
func NewPriceHandler(priceCache domain.PriceCache, symbols []string) *PriceHandler {
	return &PriceHandler{
		priceCache: priceCache,
		symbols:    symbols,
	}
}

// PriceResponse represents a single price
type PriceResponse struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Mid       float64 `json:"mid"`
	Spread    float64 `json:"spread"`
	Timestamp string  `json:"timestamp"`
}

// GetPrices returns all current prices
// GET /prices
func (h *PriceHandler) GetPrices(w http.ResponseWriter, r *http.Request) {
	prices := h.priceCache.GetAll()

	response := make([]PriceResponse, 0, len(prices))
	for _, p := range prices {
		response = append(response, PriceResponse{
			Symbol:    p.Symbol,
			Bid:       p.Bid,
			Ask:       p.Ask,
			Mid:       p.Mid(),
			Spread:    p.Spread(),
			Timestamp: p.Timestamp.Format("2006-01-02T15:04:05Z"),
		})
	}

	writeJSON(w, response, http.StatusOK)
}

// SymbolInfo represents trading pair information
type SymbolInfo struct {
	Symbol          string `json:"symbol"`
	BaseCurrency    string `json:"base_currency"`
	QuoteCurrency   string `json:"quote_currency"`
	MinQuantity     string `json:"min_quantity"`
	MaxQuantity     string `json:"max_quantity"`
	QuantityStep    string `json:"quantity_step"`
	MinLeverage     int    `json:"min_leverage"`
	MaxLeverage     int    `json:"max_leverage"`
	MaintenanceRate string `json:"maintenance_rate"`
}

// GetSymbols returns supported trading symbols
// GET /symbols
func (h *PriceHandler) GetSymbols(w http.ResponseWriter, r *http.Request) {
	symbols := make([]SymbolInfo, 0, len(h.symbols))

	for _, s := range h.symbols {
		// Parse symbol to get base and quote currencies
		base := s[:len(s)-4]  // e.g., "BTC" from "BTCUSDT"
		quote := s[len(s)-4:] // e.g., "USDT"

		symbols = append(symbols, SymbolInfo{
			Symbol:          s,
			BaseCurrency:    base,
			QuoteCurrency:   quote,
			MinQuantity:     "0.001",
			MaxQuantity:     "1000",
			QuantityStep:    "0.001",
			MinLeverage:     1,
			MaxLeverage:     100,
			MaintenanceRate: "0.005",
		})
	}

	writeJSON(w, symbols, http.StatusOK)
}
