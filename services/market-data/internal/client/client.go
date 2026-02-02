package client

import (
	"context"

	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
)

type PriceClient interface {
	StreamPrices(ctx context.Context) (<-chan domain.PriceWithError, error)
	Close() error
	Name() string
}
