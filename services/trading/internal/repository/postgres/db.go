package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"trading/config"
	"trading/internal/logger"
)

type DB struct {
	*sql.DB
}

func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return &DB{db}, nil
}

func (db *DB) Connect(ctx context.Context, retries int, interval time.Duration) error {
	var lastErr error

	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := db.PingContext(ctx); err == nil {
			logger.Info("database connected successfully", "attempt", i+1)
			return nil
		} else {
			lastErr = err
		}

		if i < retries-1 {
			logger.Warn("database connection failed, retrying",
				"attempt", i+1,
				"max_retries", retries,
				"error", lastErr,
			)
			time.Sleep(interval)
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", retries, lastErr)
}

func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}
