package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"trading/config"
	"trading/internal/app"
	"trading/internal/logger"
)

func main() {
	// Load and validate config
	cfg, err := config.Load()
	if err != nil {
		println("FATAL: " + err.Error())
		os.Exit(1)
	}

	// Initialize structured logger
	logger.Init(cfg.Service.LogLevel)

	logger.Info("config loaded successfully",
		"service", cfg.Service.Name,
		"http_port", cfg.Service.HTTPPort,
		"db_host", cfg.Database.Host,
		"kafka_brokers", cfg.Kafka.Brokers,
	)

	// Create application
	tradingApp, err := app.New(cfg)
	if err != nil {
		logger.Error("failed to create app", "error", err)
		os.Exit(1)
	}
	defer tradingApp.Close()

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run application
	if err := tradingApp.Run(ctx); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}

	logger.Info("application stopped gracefully")
}
