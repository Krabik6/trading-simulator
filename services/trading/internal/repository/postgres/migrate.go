package postgres

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"trading/internal/logger"
)

func RunMigrations(db *sql.DB, fs embed.FS) error {
	source, err := iofs.New(fs, ".")
	if err != nil {
		return fmt.Errorf("create iofs source: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	// Fix dirty state before running migrations
	version, dirty, err := m.Version()
	if err == nil && dirty {
		logger.Warn("dirty database detected, forcing version", "version", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("force dirty version: %w", err)
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, _ = m.Version()
	logger.Info("migrations applied", "version", version, "dirty", dirty)

	return nil
}
