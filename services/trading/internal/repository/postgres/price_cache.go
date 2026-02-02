package postgres

import (
	"sync"

	"trading/internal/domain"
)

// PriceCache is an in-memory cache for latest prices
type PriceCache struct {
	mu     sync.RWMutex
	prices map[string]*domain.Price
}

func NewPriceCache() *PriceCache {
	return &PriceCache{
		prices: make(map[string]*domain.Price),
	}
}

func (c *PriceCache) Get(symbol string) (*domain.Price, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	price, ok := c.prices[symbol]
	return price, ok
}

func (c *PriceCache) Set(symbol string, price *domain.Price) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.prices[symbol] = price
}

func (c *PriceCache) GetAll() map[string]*domain.Price {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string]*domain.Price, len(c.prices))
	for k, v := range c.prices {
		result[k] = v
	}
	return result
}
