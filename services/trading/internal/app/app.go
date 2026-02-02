package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"trading/config"
	"trading/internal/auth"
	httpdelivery "trading/internal/delivery/http"
	"trading/internal/delivery/http/handler"
	"trading/internal/delivery/http/middleware"
	"trading/internal/engine"
	"trading/internal/kafka"
	"trading/internal/logger"
	"trading/internal/repository/postgres"
	accountuc "trading/internal/usecase/account"
	authuc "trading/internal/usecase/auth"
	orderuc "trading/internal/usecase/order"
	positionuc "trading/internal/usecase/position"
	priceuc "trading/internal/usecase/price"
)

type App struct {
	config *config.Config
	db     *postgres.DB
	server *http.Server

	priceConsumer  *kafka.PriceConsumer
	tradeProducer  *kafka.TradeProducer
	priceProcessor *priceuc.Processor
}

func New(cfg *config.Config) (*App, error) {
	// Initialize database
	db, err := postgres.NewDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("create database: %w", err)
	}

	return &App{
		config: cfg,
		db:     db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	logger.Info("starting trading service",
		"http_port", a.config.Service.HTTPPort,
		"kafka_brokers", a.config.Kafka.Brokers,
	)

	// Connect to database
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.db.Connect(connectCtx, 10, 3*time.Second); err != nil {
		return fmt.Errorf("database connection: %w", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(a.db)
	accountRepo := postgres.NewAccountRepository(a.db)
	orderRepo := postgres.NewOrderRepository(a.db)
	positionRepo := postgres.NewPositionRepository(a.db)
	tradeRepo := postgres.NewTradeRepository(a.db)
	priceCache := postgres.NewPriceCache()

	// Initialize engine
	eng := engine.NewEngine(a.config.Trading.MaxLeverage, a.config.Trading.MaintenanceRate)

	// Initialize JWT service
	jwtService := auth.NewJWTService(a.config.JWT.Secret, a.config.JWT.ExpiryHours)

	// Initialize Kafka producer
	a.tradeProducer = kafka.NewTradeProducer(
		a.config.Kafka.Brokers,
		a.config.Kafka.TradesTopic,
	)

	if err := a.tradeProducer.Connect(
		connectCtx,
		a.config.Kafka.ConnectRetries,
		time.Duration(a.config.Kafka.RetryInterval)*time.Second,
	); err != nil {
		return fmt.Errorf("kafka producer connection: %w", err)
	}

	// Initialize Kafka consumer
	a.priceConsumer = kafka.NewPriceConsumer(
		a.config.Kafka.Brokers,
		a.config.Kafka.PricesTopic,
		a.config.Kafka.ConsumerGroup,
	)

	// Initialize use cases
	authUC := authuc.NewUseCase(
		userRepo,
		accountRepo,
		jwtService,
		a.config.Trading.InitialBalance,
	)

	positionUC := positionuc.NewUseCase(
		positionRepo,
		accountRepo,
		tradeRepo,
		orderRepo,
		priceCache,
		eng,
	)

	orderUC := orderuc.NewUseCase(
		orderRepo,
		positionRepo,
		accountRepo,
		tradeRepo,
		priceCache,
		eng,
		a.config.Trading.SupportedSymbols,
	)

	accountUC := accountuc.NewUseCase(accountRepo, positionRepo)

	// Initialize price processor
	a.priceProcessor = priceuc.NewProcessor(
		positionRepo,
		priceCache,
		eng,
		a.tradeProducer,
		positionUC,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)
	accountHandler := handler.NewAccountHandler(accountUC)
	orderHandler := handler.NewOrderHandler(orderUC)
	positionHandler := handler.NewPositionHandler(positionUC)
	tradeHandler := handler.NewTradeHandler(tradeRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Create router
	router := httpdelivery.NewRouter(httpdelivery.RouterDeps{
		AuthMiddleware:  authMiddleware,
		AuthHandler:     authHandler,
		AccountHandler:  accountHandler,
		OrderHandler:    orderHandler,
		PositionHandler: positionHandler,
		TradeHandler:    tradeHandler,
		HealthChecker:   a.healthCheck,
	})

	// Start HTTP server
	a.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.config.Service.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "port", a.config.Service.HTTPPort)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", "error", err)
		}
	}()

	// Start price consumer
	if err := a.priceConsumer.Start(ctx); err != nil {
		return fmt.Errorf("start price consumer: %w", err)
	}

	// Start price processor
	go a.priceProcessor.Start(ctx, a.priceConsumer.Prices())

	logger.Info("trading service started successfully")

	// Wait for shutdown signal
	<-ctx.Done()

	logger.Info("shutting down trading service")
	return nil
}

func (a *App) healthCheck() error {
	if err := a.db.Health(); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := a.priceConsumer.Health(); err != nil {
		return fmt.Errorf("kafka consumer: %w", err)
	}
	return nil
}

func (a *App) Close() error {
	logger.Info("closing application")

	var errs []error

	// Stop price producer
	if a.tradeProducer != nil {
		a.tradeProducer.WaitForCompletion(5 * time.Second)
		if err := a.tradeProducer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close trade producer: %w", err))
		}
	}

	// Stop price consumer
	if a.priceConsumer != nil {
		if err := a.priceConsumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close price consumer: %w", err))
		}
	}

	// Stop HTTP server
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown server: %w", err))
		}
	}

	// Close database
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close database: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing app: %v", errs)
	}

	logger.Info("application closed successfully")
	return nil
}
