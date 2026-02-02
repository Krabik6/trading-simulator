package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Krabik6/trading-simulator/market-data/config"
	"github.com/Krabik6/trading-simulator/market-data/internal/app"
	"github.com/Krabik6/trading-simulator/market-data/internal/logger"
)

func main() {
	// Load and validate config
	cfg, err := config.Load()
	if err != nil {
		// Use basic logging before logger is initialized
		println("FATAL: " + err.Error())
		os.Exit(1)
	}

	// Initialize structured logger
	logger.Init(cfg.Service.LogLevel)

	logger.Info("config loaded successfully",
		"service", cfg.Service.Name,
		"client_type", cfg.Client.Type,
		"symbols", cfg.Client.Symbols,
	)

	// Create application
	marketApp, err := app.New(cfg)
	if err != nil {
		logger.Error("failed to create app", "error", err)
		os.Exit(1)
	}
	defer marketApp.Close()

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run application
	if err := marketApp.Run(ctx); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}

	logger.Info("application stopped gracefully")
}
