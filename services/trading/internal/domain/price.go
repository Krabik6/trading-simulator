package domain

import (
	"encoding/json"
	"time"
)

// Price represents a price update from market-data service
type Price struct {
	Symbol    string    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// Mid returns the mid price
func (p *Price) Mid() float64 {
	return (p.Bid + p.Ask) / 2
}

// Spread returns the bid-ask spread
func (p *Price) Spread() float64 {
	return p.Ask - p.Bid
}

// FromJSON parses a Price from JSON
func PriceFromJSON(data []byte) (*Price, error) {
	var price Price
	err := json.Unmarshal(data, &price)
	return &price, err
}
