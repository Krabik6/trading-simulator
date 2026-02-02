package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Krabik6/trading-simulator/market-data/config"
	"github.com/Krabik6/trading-simulator/market-data/internal/client"
	"github.com/Krabik6/trading-simulator/market-data/internal/domain"
	"github.com/Krabik6/trading-simulator/market-data/internal/kafka"
	"github.com/Krabik6/trading-simulator/market-data/internal/logger"
	"github.com/Krabik6/trading-simulator/market-data/internal/metrics"
)

type App struct {
	config   *config.Config
	client   client.PriceClient
	producer *kafka.KafkaProducer
	server   *http.Server

	processed atomic.Int64
	errors    atomic.Int64

	wg       sync.WaitGroup
	shutdown atomic.Bool
}

func New(cfg *config.Config) (*App, error) {
	symbols := domain.SymbolsFromStrings(cfg.Client.Symbols)

	priceClient, err := client.NewClient(
		client.ClientType(cfg.Client.Type),
		symbols,
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	producer := kafka.NewKafkaProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topic,
		cfg.Kafka.BatchSize,
		time.Duration(cfg.Kafka.BatchTimeout)*time.Millisecond,
	)

	return &App{
		config:   cfg,
		client:   priceClient,
		producer: producer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	logger.Info("starting market-data service",
		"client", a.config.Client.Type,
		"symbols", a.config.Client.Symbols,
		"kafka_brokers", a.config.Kafka.Brokers,
		"kafka_topic", a.config.Kafka.Topic,
	)

	// Connect to Kafka with retries
	connectCtx, cancel := context.WithTimeout(ctx, time.Duration(a.config.Kafka.ConnectRetries*a.config.Kafka.RetryInterval)*time.Second)
	defer cancel()

	if err := a.producer.Connect(
		connectCtx,
		a.config.Kafka.ConnectRetries,
		time.Duration(a.config.Kafka.RetryInterval)*time.Second,
	); err != nil {
		return fmt.Errorf("kafka connection: %w", err)
	}

	// Start HTTP server
	go a.startHTTPServer()

	// Start price stream
	priceCh, err := a.client.StreamPrices(ctx)
	if err != nil {
		return fmt.Errorf("start stream: %w", err)
	}

	metrics.SetClientConnected(true)
	logger.Info("price stream started")

	return a.processPrices(ctx, priceCh)
}

func (a *App) processPrices(ctx context.Context, priceCh <-chan domain.PriceWithError) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.shutdown.Store(true)
			logger.Info("shutting down, waiting for in-flight messages",
				"processed", a.processed.Load(),
				"errors", a.errors.Load(),
			)
			// Wait for in-flight goroutines
			a.wg.Wait()
			return nil

		case <-ticker.C:
			logger.Info("processing stats",
				"processed", a.processed.Load(),
				"errors", a.errors.Load(),
			)

		case priceWithErr, ok := <-priceCh:
			if !ok {
				logger.Warn("price channel closed")
				a.wg.Wait()
				return nil
			}

			if priceWithErr.Error != nil {
				a.errors.Add(1)
				metrics.RecordError("unknown", "client")
				logger.Error("client error", "error", priceWithErr.Error)
				continue
			}

			a.handlePrice(ctx, priceWithErr.Price)
		}
	}
}

func (a *App) handlePrice(ctx context.Context, price domain.Price) {
	if a.shutdown.Load() {
		return
	}

	a.processed.Add(1)
	metrics.RecordPrice(string(price.Symbol), price.Source)
	metrics.SetCurrentPrice(string(price.Symbol), price.Bid, price.Ask)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		sendCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := a.producer.Send(sendCtx, price); err != nil {
			a.errors.Add(1)
			logger.Error("failed to send price",
				"symbol", price.Symbol,
				"error", err,
			)
		}
	}()

	logger.Debug("price received",
		"symbol", price.Symbol,
		"bid", price.Bid,
		"ask", price.Ask,
		"spread", price.Spread(),
	)
}

func (a *App) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.healthHandler)
	mux.HandleFunc("/ready", a.readyHandler)
	mux.Handle("/metrics", promhttp.Handler())

	a.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.config.Service.HTTPPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("http server starting", "port", a.config.Service.HTTPPort)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http server error", "error", err)
	}
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (a *App) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := a.producer.Health(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf(`{"status":"not ready","error":"%s"}`, err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (a *App) Close() error {
	logger.Info("closing application")

	var errs []error

	// Stop accepting new prices
	a.shutdown.Store(true)

	// Wait for in-flight messages (with timeout)
	a.producer.WaitForCompletion(5 * time.Second)

	if a.client != nil {
		metrics.SetClientConnected(false)
		if err := a.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close client: %w", err))
		}
	}

	if a.producer != nil {
		if err := a.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close producer: %w", err))
		}
	}

	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown server: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing app: %v", errs)
	}

	logger.Info("application closed successfully")
	return nil
}
