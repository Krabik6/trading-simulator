package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"

	"trading/internal/domain"
	"trading/internal/logger"
)

type TradeProducer struct {
	writer    *kafka.Writer
	topic     string
	brokers   []string
	inFlight  atomic.Int64
	closeOnce sync.Once
	closed    atomic.Bool
}

func NewTradeProducer(brokers []string, topic string) *TradeProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		BatchSize:              100,
		BatchTimeout:           100 * time.Millisecond,
		RequiredAcks:           kafka.RequireOne,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}

	logger.Info("trade producer initialized",
		"brokers", brokers,
		"topic", topic,
	)

	return &TradeProducer{
		writer:  writer,
		topic:   topic,
		brokers: brokers,
	}
}

func (p *TradeProducer) Connect(ctx context.Context, retries int, retryInterval time.Duration) error {
	var lastErr error

	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.Health(); err == nil {
			logger.Info("trade producer connected successfully", "attempt", i+1)
			return nil
		} else {
			lastErr = err
		}

		if i < retries-1 {
			logger.Warn("kafka connection failed, retrying",
				"attempt", i+1,
				"max_retries", retries,
				"error", lastErr,
			)
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("failed to connect to kafka after %d attempts: %w", retries, lastErr)
}

func (p *TradeProducer) PublishTrade(ctx context.Context, trade *domain.Trade) error {
	if p.closed.Load() {
		return fmt.Errorf("producer is closed")
	}

	p.inFlight.Add(1)
	defer p.inFlight.Add(-1)

	event := trade.ToEvent()
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal trade event: %w", err)
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(trade.Symbol),
			Value: data,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{Key: "trade_type", Value: []byte(trade.Type)},
				{Key: "symbol", Value: []byte(trade.Symbol)},
			},
		},
	)

	if err != nil {
		logger.Error("failed to publish trade",
			"trade_id", trade.ID,
			"error", err,
		)
		return fmt.Errorf("write to kafka: %w", err)
	}

	logger.Debug("trade published",
		"trade_id", trade.ID,
		"symbol", trade.Symbol,
		"type", trade.Type,
	)

	return nil
}

func (p *TradeProducer) WaitForCompletion(timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if p.inFlight.Load() == 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	remaining := p.inFlight.Load()
	if remaining > 0 {
		logger.Warn("shutdown timeout, messages may be lost", "in_flight", remaining)
	}
}

func (p *TradeProducer) Close() error {
	var err error
	p.closeOnce.Do(func() {
		p.closed.Store(true)
		logger.Info("closing trade producer")
		err = p.writer.Close()
	})
	return err
}

func (p *TradeProducer) Health() error {
	conn, err := kafka.Dial("tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("kafka connection failed: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get brokers: %w", err)
	}

	return nil
}
