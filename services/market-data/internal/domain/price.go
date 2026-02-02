package domain

import (
	"encoding/json"
	"time"
)

type Symbol string

const (
	BTCUSDT Symbol = "BTCUSDT"
	ETHUSDT Symbol = "ETHUSDT"
	SOLUSDT Symbol = "SOLUSDT"
)

type Price struct {
	Symbol    Symbol    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type PriceWithError struct {
	Price Price
	Error error
}

func (p *Price) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func PriceFromJSON(data []byte) (*Price, error) {
	var price Price
	err := json.Unmarshal(data, &price)
	return &price, err
}

func (p *Price) Spread() float64 {
	return p.Ask - p.Bid
}

func SymbolsFromStrings(symbols []string) []Symbol {
	result := make([]Symbol, len(symbols))
	for i, s := range symbols {
		result[i] = Symbol(s)
	}
	return result
}
