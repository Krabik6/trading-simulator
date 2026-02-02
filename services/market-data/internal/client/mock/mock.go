package mock

import (
	"context"
	"math/rand"
	"time"

	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
	"github.com/Krabik6/trading-simulator/market-data/internal/logger"
)

type MockClient struct {
	symbols []domain.Symbol
	delay   time.Duration
	stop    chan struct{}
}

func NewMockClient(symbols []domain.Symbol) *MockClient {
	logger.Info("mock client created", "symbols", symbols)
	return &MockClient{
		symbols: symbols,
		delay:   1 * time.Second,
		stop:    make(chan struct{}),
	}
}

func (c *MockClient) StreamPrices(ctx context.Context) (<-chan domain.PriceWithError, error) {
	priceCh := make(chan domain.PriceWithError, 100)

	go func() {
		defer close(priceCh)
		logger.Info("mock price stream started")

		for {
			select {
			case <-ctx.Done():
				logger.Info("mock stream stopped by context")
				return
			case <-c.stop:
				logger.Info("mock stream stopped")
				return
			default:
				for _, symbol := range c.symbols {
					price := c.generatePrice(symbol)
					priceCh <- domain.PriceWithError{Price: price}
				}
				time.Sleep(c.delay)
			}
		}
	}()

	return priceCh, nil
}

func (c *MockClient) generatePrice(symbol domain.Symbol) domain.Price {
	basePrice := map[domain.Symbol]float64{
		domain.BTCUSDT: 60000,
		domain.ETHUSDT: 3000,
		domain.SOLUSDT: 150,
	}

	base, ok := basePrice[symbol]
	if !ok {
		base = 100
	}

	deviation := (rand.Float64()*0.04 - 0.02)
	currentPrice := base * (1 + deviation)
	spread := currentPrice * 0.001

	return domain.Price{
		Symbol:    symbol,
		Bid:       currentPrice - spread/2,
		Ask:       currentPrice + spread/2,
		Timestamp: time.Now(),
		Source:    "mock",
	}
}

func (c *MockClient) Close() error {
	close(c.stop)
	logger.Info("mock client closed")
	return nil
}

func (c *MockClient) Name() string {
	return "mock"
}

func (c *MockClient) SetDelay(delay time.Duration) {
	c.delay = delay
}
