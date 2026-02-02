package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
	"github.com/Krabik6/trading-simulator/market-data/internal/logger"
	"github.com/Krabik6/trading-simulator/market-data/internal/metrics"
	"github.com/segmentio/kafka-go"
)

type PriceProducer interface {
	Send(ctx context.Context, price domain.Price) error
	SendBatch(ctx context.Context, prices []domain.Price) error
	Close() error
	Health() error
	WaitForCompletion(timeout time.Duration)
}

type KafkaProducer struct {
	writer    *kafka.Writer
	topic     string
	brokers   []string
	inFlight  atomic.Int64
	closeOnce sync.Once
	closed    atomic.Bool
}

func NewKafkaProducer(brokers []string, topic string, batchSize int, batchTimeout time.Duration) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		BatchSize:              batchSize,
		BatchTimeout:           batchTimeout,
		RequiredAcks:           kafka.RequireOne,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}

	logger.Info("kafka producer initialized",
		"brokers", brokers,
		"topic", topic,
		"batch_size", batchSize,
	)

	return &KafkaProducer{
		writer:  writer,
		topic:   topic,
		brokers: brokers,
	}
}

// Connect attempts to connect to Kafka with retries
func (p *KafkaProducer) Connect(ctx context.Context, retries int, retryInterval time.Duration) error {
	var lastErr error

	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.Health(); err == nil {
			logger.Info("kafka connected successfully", "attempt", i+1)
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

func (p *KafkaProducer) Send(ctx context.Context, price domain.Price) error {
	if p.closed.Load() {
		return fmt.Errorf("producer is closed")
	}

	p.inFlight.Add(1)
	defer p.inFlight.Add(-1)

	start := time.Now()

	data, err := json.Marshal(price)
	if err != nil {
		metrics.RecordError(string(price.Symbol), "marshal")
		return fmt.Errorf("marshal price: %w", err)
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(price.Symbol),
			Value: data,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{Key: "source", Value: []byte(price.Source)},
				{Key: "symbol", Value: []byte(price.Symbol)},
			},
		},
	)

	duration := time.Since(start).Seconds()
	metrics.RecordKafkaSend(string(price.Symbol), duration)

	if err != nil {
		metrics.RecordKafkaError(string(price.Symbol))
		logger.Error("failed to send to kafka",
			"symbol", price.Symbol,
			"error", err,
		)
		return fmt.Errorf("write to kafka: %w", err)
	}

	logger.Debug("price sent to kafka",
		"symbol", price.Symbol,
		"bid", price.Bid,
		"ask", price.Ask,
	)

	return nil
}

func (p *KafkaProducer) SendBatch(ctx context.Context, prices []domain.Price) error {
	if p.closed.Load() {
		return fmt.Errorf("producer is closed")
	}

	p.inFlight.Add(1)
	defer p.inFlight.Add(-1)

	messages := make([]kafka.Message, len(prices))

	for i, price := range prices {
		data, err := json.Marshal(price)
		if err != nil {
			return fmt.Errorf("marshal price %d: %w", i, err)
		}

		messages[i] = kafka.Message{
			Key:   []byte(price.Symbol),
			Value: data,
			Time:  time.Now(),
		}
	}

	return p.writer.WriteMessages(ctx, messages...)
}

// WaitForCompletion waits for all in-flight messages to complete
func (p *KafkaProducer) WaitForCompletion(timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if p.inFlight.Load() == 0 {
			logger.Info("all in-flight messages completed")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	remaining := p.inFlight.Load()
	if remaining > 0 {
		logger.Warn("shutdown timeout, messages may be lost", "in_flight", remaining)
	}
}

func (p *KafkaProducer) Close() error {
	var err error
	p.closeOnce.Do(func() {
		p.closed.Store(true)
		logger.Info("closing kafka producer")
		err = p.writer.Close()
	})
	return err
}

func (p *KafkaProducer) Health() error {
	conn, err := kafka.Dial("tcp", p.brokers[0])
	if err != nil {
		return fmt.Errorf("kafka connection failed: %w", err)
	}
	defer conn.Close()

	// Just check connection, don't require topic to exist yet
	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get brokers: %w", err)
	}

	return nil
}
