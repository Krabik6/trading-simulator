package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"

	"trading/internal/domain"
	"trading/internal/logger"
)

type PriceConsumer struct {
	reader    *kafka.Reader
	pricesCh  chan *domain.Price
	closeOnce sync.Once
	closed    chan struct{}
}

func NewPriceConsumer(brokers []string, topic, groupID string) *PriceConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		StartOffset:    kafka.LastOffset,
		CommitInterval: time.Second,
	})

	return &PriceConsumer{
		reader:   reader,
		pricesCh: make(chan *domain.Price, 1000),
		closed:   make(chan struct{}),
	}
}

func (c *PriceConsumer) Start(ctx context.Context) error {
	logger.Info("starting price consumer")

	go c.consume(ctx)
	return nil
}

func (c *PriceConsumer) consume(ctx context.Context) {
	defer close(c.pricesCh)

	for {
		select {
		case <-ctx.Done():
			logger.Info("price consumer stopping due to context cancellation")
			return
		case <-c.closed:
			logger.Info("price consumer closed")
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error("failed to fetch message", "error", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		price, err := domain.PriceFromJSON(msg.Value)
		if err != nil {
			logger.Error("failed to parse price", "error", err)
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		select {
		case c.pricesCh <- price:
		default:
			// Channel full, log and continue
			logger.Warn("price channel full, dropping message", "symbol", price.Symbol)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			logger.Error("failed to commit message", "error", err)
		}
	}
}

func (c *PriceConsumer) Prices() <-chan *domain.Price {
	return c.pricesCh
}

func (c *PriceConsumer) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.closed)
		err = c.reader.Close()
		logger.Info("price consumer closed")
	})
	return err
}

func (c *PriceConsumer) Health() error {
	if c.reader == nil {
		return fmt.Errorf("reader not initialized")
	}
	return nil
}
