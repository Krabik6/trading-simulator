package client

import (
	"github.com/Krabik6/trading-simulator/market-data/internal/client/binance"
	"github.com/Krabik6/trading-simulator/market-data/internal/client/mock"
	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
)

// ClientType определяет тип клиента
type ClientType string

const (
	ClientTypeMock    ClientType = "mock"
	ClientTypeBinance ClientType = "binance"
)

// NewClient создает клиент по типу
func NewClient(clientType ClientType, symbols []domain.Symbol) (PriceClient, error) {
	switch clientType {
	case ClientTypeBinance:
		return binance.NewBinanceClient(symbols), nil
	case ClientTypeMock:
		return mock.NewMockClient(symbols), nil
	default:
		return mock.NewMockClient(symbols), nil
	}
}
